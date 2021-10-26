package testutil

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/lib/pq"
	pkgerr "github.com/pkg/errors"
	_ "github.com/stretchr/testify"
	"gopkg.in/khaiql/dbcleaner.v2"
	"gopkg.in/khaiql/dbcleaner.v2/engine"

	"github.com/Bnei-Baruch/archive-my/common"
	mdb_migrations "github.com/Bnei-Baruch/archive-my/databases/mdb/migrations"
	"github.com/Bnei-Baruch/archive-my/databases/mydb/models"
	"github.com/Bnei-Baruch/archive-my/pkg/utils"
)

type TestMyDBManager struct {
	DB        *sql.DB
	Name      string
	DSN       string
	DBCleaner dbcleaner.DbCleaner
}

func (m *TestMyDBManager) Init() error {
	//boil.DebugMode = true

	m.Name = fmt.Sprintf("test_%s", strings.ToLower(utils.GenerateName(5)))
	fmt.Println("Initializing test MyDB: ", m.Name)

	// Open connection
	db, err := sql.Open("postgres", common.Config.MyDBUrl)
	if err != nil {
		return pkgerr.Wrap(err, "sql.Open")
	}

	// Create a new temporary test database
	if _, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", m.Name)); err != nil {
		return pkgerr.Wrap(err, "create database")
	}

	// Close first connection
	if err := db.Close(); err != nil {
		return pkgerr.Wrap(err, "db.Close")
	}

	// Connect to temp database and run migrations
	m.DSN, err = replaceDBName(m.Name, common.Config.MyDBUrl)
	if err != nil {
		return pkgerr.Wrap(err, "replaceDBName")
	}

	if err := runMigrate(m.DSN); err != nil {
		return pkgerr.Wrap(err, "run migrations")
	}

	m.DB, err = sql.Open("postgres", m.DSN)
	if err != nil {
		return pkgerr.Wrap(err, "sql.Open temporary DB")
	}

	tmpDir, err := ioutil.TempDir("", "dbcleaner_mydb")
	if err != nil {
		return pkgerr.Wrap(err, "tmp dir for dbcleaner")
	}
	m.DBCleaner = dbcleaner.New(
		dbcleaner.SetLockFileDir(tmpDir),
	)
	m.DBCleaner.SetEngine(engine.NewPostgresEngine(m.DSN))

	return nil
}

func (m *TestMyDBManager) Destroy() error {
	fmt.Println("Destroying test MyDB: ", m.Name)

	// Close DB cleaner
	if err := m.DBCleaner.Close(); err != nil {
		return pkgerr.Wrap(err, "DBCleaner.Close")
	}

	// Close temp DB
	if err := m.DB.Close(); err != nil {
		return pkgerr.Wrap(err, "DB.Close")
	}

	// Connect to main dev DB
	db, err := sql.Open("postgres", common.Config.MyDBUrl)
	if err != nil {
		return pkgerr.Wrap(err, "sql.Open dev DB")
	}

	// Drop test DB
	if _, err = db.Exec(fmt.Sprintf("DROP DATABASE %s", m.Name)); err != nil {
		return pkgerr.Wrap(err, "drop database")
	}

	return nil
}

func (m *TestMyDBManager) AllTables() []string {
	v := reflect.ValueOf(models.TableNames)
	t := v.Type()
	tables := make([]string, 0)
	for i := 0; i < t.NumField(); i++ {
		name := t.Field(i).Name
		value := v.FieldByName(name).Interface()
		if value.(string) != models.TableNames.SchemaMigrations {
			tables = append(tables, value.(string))
		}
	}
	return tables
}

type TestMDBManager struct {
	DB        *sql.DB
	Name      string
	DSN       string
	DBCleaner dbcleaner.DbCleaner
}

func (m *TestMDBManager) Init() error {
	m.Name = fmt.Sprintf("test_mdb_%s", strings.ToLower(utils.GenerateName(5)))
	fmt.Println("Initializing test DB [mdb]: ", m.Name)

	// Open connection
	db, err := sql.Open("postgres", common.Config.MDBUrl)
	if err != nil {
		return pkgerr.Wrap(err, "sql.Open")
	}

	// Create a new temporary test database
	if _, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", m.Name)); err != nil {
		return pkgerr.Wrap(err, "create database")
	}

	// Close first connection
	if err := db.Close(); err != nil {
		return pkgerr.Wrap(err, "db.Close")
	}

	// Connect to temp database and run migrations
	m.DSN, err = replaceDBName(m.Name, common.Config.MDBUrl)
	if err != nil {
		return pkgerr.Wrap(err, "replaceDBName")
	}

	m.DB, err = sql.Open("postgres", m.DSN)
	if err != nil {
		return pkgerr.Wrap(err, "sql.Open temporary DB")
	}

	if err := runRambler(m.DB); err != nil {
		return pkgerr.Wrap(err, "run migrations")
	}

	tmpDir, err := ioutil.TempDir("", "dbcleaner_mdb")
	if err != nil {
		return pkgerr.Wrap(err, "tmp dir for dbcleaner")
	}
	m.DBCleaner = dbcleaner.New(
		dbcleaner.SetLockFileDir(tmpDir),
	)
	m.DBCleaner.SetEngine(engine.NewPostgresEngine(m.DSN))

	return nil
}

func (m *TestMDBManager) Destroy() error {
	fmt.Println("Destroying test MDB: ", m.Name)

	// Close DB cleaner
	if err := m.DBCleaner.Close(); err != nil {
		return pkgerr.Wrap(err, "DBCleaner.Close")
	}

	// Close temp DB
	if err := m.DB.Close(); err != nil {
		return pkgerr.Wrap(err, "DB.Close")
	}

	// Connect to main dev DB
	db, err := sql.Open("postgres", common.Config.MDBUrl)
	if err != nil {
		return pkgerr.Wrap(err, "sql.Open dev DB")
	}

	// Drop test DB
	if _, err = db.Exec(fmt.Sprintf("DROP DATABASE %s", m.Name)); err != nil {
		return pkgerr.Wrap(err, "drop database")
	}

	return nil
}

func (m *TestMDBManager) AllTables() []string {
	// hard coded since currently we don't reuse mdb golang models
	// see https://github.com/Bnei-Baruch/mdb/blob/master/models/boil_table_names.go
	return []string{
		"author_i18n",
		"authors",
		"authors_sources",
		"blog_posts",
		"blogs",
		"collection_i18n",
		"collections",
		"collections_content_units",
		"content_role_types",
		//"content_types",  // holds data from migrations (don't clean each time)
		"content_unit_derivations",
		"content_unit_i18n",
		"content_units",
		"content_units_persons",
		"content_units_publishers",
		"content_units_sources",
		"content_units_tags",
		"files",
		"files_operations",
		"files_storages",
		//"operation_types",  // holds data from migrations (don't clean each time)
		"operations",
		"person_i18n",
		"persons",
		"publisher_i18n",
		"publishers",
		"source_i18n",
		"source_types",
		"sources",
		"storages",
		"tag_i18n",
		"tags",
		"twitter_tweets",
		"twitter_users",
		"users",
	}
}

func runMigrate(dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	_, filename, _, _ := runtime.Caller(0)
	rel := filepath.Join(filepath.Dir(filename), "..", "..", "databases", "mydb", "migrations")
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

func runRambler(db *sql.DB) error {
	var visit = func(path string, f os.FileInfo, err error) error {
		match, _ := regexp.MatchString(".*\\.sql$", path)
		if !match {
			return nil
		}

		//fmt.Printf("Applying migration %s\n", path)
		m, err := mdb_migrations.NewMigration(path)
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
	rel := filepath.Join(filepath.Dir(filename), "..", "..", "databases", "mdb", "migrations")
	return filepath.Walk(rel, visit)
}

func replaceDBName(tempName, dbStr string) (string, error) {
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
