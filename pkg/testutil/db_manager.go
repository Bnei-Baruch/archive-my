package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/lib/pq"
	"github.com/spf13/viper"
	_ "github.com/stretchr/testify"
	"gopkg.in/khaiql/dbcleaner.v2"
	"gopkg.in/khaiql/dbcleaner.v2/engine"

	migrations "archive-my/mdb_migrations"
	"archive-my/pkg/utils"
)

type TestDBManager struct {
	DB        *sql.DB
	MDB       *sql.DB
	testDB    string
	testMDB   string
	DBCleaner dbcleaner.DbCleaner
}

func (m *TestDBManager) InitTestDB() (string, string, error) {
	//boil.DebugMode = true

	m.DBCleaner = dbcleaner.New()
	var mdbDs string
	var dbDs string
	if db, dsn, name, err := m.initDB(false); err != nil {
		return "", "", err
	} else {
		m.DB = db
		m.testDB = name
		dbDs = dsn
	}

	if db, dsn, name, err := m.initDB(true); err != nil {
		return "", "", err
	} else {
		m.MDB = db
		m.testMDB = name
		mdbDs = dsn
	}

	return dbDs, mdbDs, nil
}

func (m *TestDBManager) initDB(isMDB bool) (*sql.DB, string, string, error) {
	//boil.DebugMode = true
	prefix := ""
	if isMDB {
		prefix = "_mdb"
	}
	name := fmt.Sprintf("test%s_%s", prefix, strings.ToLower(utils.GenerateName(5)))
	fmt.Println("Initializing test DB: ", name)
	// Open connection
	db, err := sql.Open("postgres", viper.GetString("app.mydb"))
	if err != nil {
		return nil, "", "", err
	}

	// Create a new temporary test database
	if _, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", name)); err != nil {
		return nil, "", "", err
	}

	// Close first connection
	if err := db.Close(); err != nil {
		return nil, "", "", err
	}

	// Connect to temp database and run migrations
	ds := viper.GetString("app.mydb")
	if isMDB {
		ds = viper.GetString("app.mdb")
	}
	dsn, err := m.replaceDBName(name, ds)
	if err != nil {
		return nil, "", "", err
	}

	if !isMDB {
		if err := m.runMigrations(dsn); err != nil {
			return nil, "", "", err
		}
	}

	db, err = sql.Open("postgres", dsn)
	if err != nil {
		return nil, "", "", err
	}

	if isMDB {
		if err := m.runMDBMigrations(db); err != nil {
			return nil, "", "", err
		}
	}
	m.DBCleaner.SetEngine(engine.NewPostgresEngine(dsn))

	return db, dsn, name, nil
}

func (m *TestDBManager) DestroyTestDB() error {
	fmt.Println("Destroying testDB: ", m.testDB)

	// Close DB cleaner
	if err := m.DBCleaner.Close(); err != nil {
		return err
	}

	if err := m.destroyDB(m.DB, m.testDB); err != nil {
		return err
	}
	return m.destroyDB(m.MDB, m.testMDB)
}

func (m *TestDBManager) destroyDB(db *sql.DB, name string) error {
	fmt.Println("Destroying testDB: ", name)

	// Close temp DB
	if err := db.Close(); err != nil {
		return err
	}

	// Connect to main dev DB
	db, err := sql.Open("postgres", viper.GetString("app.mydb"))
	if err != nil {
		return err
	}

	// Drop test DB
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE %s", name))
	if err != nil {
		return err
	}

	return nil
}

func (m *TestDBManager) runMigrations(dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	_, filename, _, _ := runtime.Caller(0)
	rel := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")
	migrator, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", rel), "postgres", driver)
	if err != nil {
		return err
	}

	if err := migrator.Up(); err != nil {
		return err
	}

	srcErr, dbErr := migrator.Close()
	if srcErr != nil {
		return err
	}
	if dbErr != nil {
		return err
	}

	return nil
}

func (m *TestDBManager) replaceDBName(tempName, dbStr string) (string, error) {
	paramsStr, err := pq.ParseURL(dbStr)
	if err != nil {
		return "", err
	}
	params := strings.Split(paramsStr, " ")
	found := false
	for i := range params {
		if strings.HasPrefix(params[i], "dbname") {
			params[i] = fmt.Sprintf("dbname=%s", tempName)
			found = true
			break
		}
	}
	if !found {
		params = append(params, fmt.Sprintf("dbname=%s", tempName))
	}
	return strings.Join(params, " "), nil
}

func (m *TestDBManager) runMDBMigrations(db *sql.DB) error {
	var visit = func(path string, f os.FileInfo, err error) error {
		match, _ := regexp.MatchString(".*\\.sql$", path)
		if !match {
			return nil
		}

		//fmt.Printf("Applying migration %s\n", path)
		m, err := migrations.NewMigration(path)
		if err != nil {
			fmt.Printf("Error migrating %s, %s", path, err.Error())
			return err
		}

		for _, statement := range m.Up() {
			if _, err := db.Exec(statement); err != nil {
				return fmt.Errorf("Unable to apply migration %s: %s\nStatement: %s\n", m.Name, err, statement)
			}
		}

		return nil
	}

	_, filename, _, _ := runtime.Caller(0)
	rel := filepath.Join(filepath.Dir(filename), "..", "..", "mdb_migrations")
	return filepath.Walk(rel, visit)
}
