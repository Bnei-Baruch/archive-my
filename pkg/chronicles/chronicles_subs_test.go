package chronicles

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"archive-my/consts"
	"archive-my/models"
	"archive-my/pkg/testutil"
	"archive-my/pkg/utils"
)

type ChroniclesSubsTestSuite struct {
	suite.Suite
	testutil.TestDBManager
	tx  *sql.Tx
	chr *Chronicles

	coUID string
	cuUID string
}

func (s *ChroniclesSubsTestSuite) SetupSuite() {
	boil.DebugMode = true
	s.NoError(utils.InitConfig("", "../../"))
	dbS, mdbS, err := s.InitTestDB()
	s.Require().Nil(err)
	s.chr = new(Chronicles)
	s.chr.Init(dbS, mdbS)
	s.coUID, s.cuUID, err = initMDB(s.MDB)
	s.NoError(err)
	s.NoError(s.MDB.Close())
}

func (s *ChroniclesSubsTestSuite) TearDownSuite() {
	s.Require().Nil(s.DestroyTestDB())
}

func (s *ChroniclesSubsTestSuite) SetupTest() {
	var err error
	s.tx, err = s.DB.Begin()
	s.NoError(err)
	s.NoError(s.MDB.Close())
}

func (s *ChroniclesSubsTestSuite) TearDownTest() {
	s.NoError(s.tx.Rollback())
}

func (s *ChroniclesSubsTestSuite) TestChronicles_byCO() {
	userId := "client:local:01EXTPWJFV1SBZTSCE566SJ"
	newUpdate := time.Now().UTC().Format(time.RFC3339)
	_, filename, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(filename), ".", "test_data", "subs_test.json")
	data, err := ioutil.ReadFile(path)
	resp := fmt.Sprintf(string(data), s.cuUID, userId, newUpdate)
	s.NoError(err)
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		_, err = res.Write([]byte(resp))
		s.NoError(err)
	}))

	sub := models.Subscription{
		AccountID: userId,
		CollectionUID: null.String{
			String: s.coUID,
			Valid:  true,
		},
		ContentUnitUID: s.cuUID,
	}
	s.NoError(sub.Insert(s.tx, boil.Infer()))

	_, err = s.chr.scanEventsOnTx(s.tx, server.URL)
	s.NoError(err)

	subs, err := models.Subscriptions(qm.Where(fmt.Sprintf("account_id = '%s'", userId))).All(s.tx)
	s.Equal(1, len(subs))
	s.Equal(newUpdate, subs[0].UpdatedAt.UTC().Format(time.RFC3339))
}

func (s *ChroniclesSubsTestSuite) TestChronicles_byType() {
	userId := "client:local:01EXTPWJFV1SBZTSCE566SJ"
	newUpdate := time.Now().UTC().Format(time.RFC3339)
	_, filename, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(filename), ".", "test_data", "subs_test.json")
	data, err := ioutil.ReadFile(path)
	resp := fmt.Sprintf(string(data), s.cuUID, userId, newUpdate)
	s.NoError(err)
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		_, err = res.Write([]byte(resp))
		s.NoError(err)
	}))

	sub := models.Subscription{
		AccountID:      userId,
		ContentType:    null.String{String: consts.CT_CLIPS, Valid: true},
		ContentUnitUID: s.cuUID,
	}
	s.NoError(sub.Insert(s.tx, boil.Infer()))

	_, err = s.chr.scanEventsOnTx(s.tx, server.URL)
	s.NoError(err)

	subs, err := models.Subscriptions(qm.Where(fmt.Sprintf("account_id = '%s'", userId))).All(s.tx)
	s.Equal(1, len(subs))
	s.Equal(newUpdate, subs[0].UpdatedAt.UTC().Format(time.RFC3339))
}

func TestChroniclesSubsTestSuite(t *testing.T) {
	suite.Run(t, new(ChroniclesSubsTestSuite))
}

//help functions

func initMDB(db *sql.DB) (string, string, error) {
	coUID, _, err := addMDBCO(db)
	if err != nil {
		return "", "", err
	}
	cuUID, _, err := addMDBCU(db)
	if err != nil {
		return "", "", err
	}
	err = getherCUAndCO(db, coUID, cuUID)
	return coUID, cuUID, err

}

func addMDBCO(db *sql.DB) (string, int, error) {
	uid := utils.GenerateUID(8)
	typeID := utils.ContentTypesByName[consts.CT_CLIPS]
	q := fmt.Sprintf("INSERT INTO collections (uid, type_id, published) VALUES ('%s', %d, %t)", uid, typeID, true)
	_, err := db.Exec(q)
	return uid, typeID, err
}

func addMDBCU(db *sql.DB) (string, int, error) {
	uid := utils.GenerateUID(8)
	typeID := utils.ContentTypesByName[consts.CT_CLIP]
	q := fmt.Sprintf("INSERT INTO content_units (uid, type_id, published) VALUES ('%s', %d, %t)", uid, typeID, true)
	_, err := db.Exec(q)
	return uid, typeID, err
}

func getherCUAndCO(db *sql.DB, coUID, cuUID string) error {
	var coID int
	coq := fmt.Sprintf("SELECT id FROM collections WHERE uid = '%s'", coUID)
	db.QueryRow(coq).Scan(&coID)
	var cuID int
	cuq := fmt.Sprintf("SELECT id FROM content_units WHERE uid = '%s'", cuUID)
	db.QueryRow(cuq).Scan(&cuID)
	q := fmt.Sprintf("INSERT INTO collections_content_units (collection_id, content_unit_id, name) VALUES (%d, %d, 'test')", coID, cuID)
	_, err := db.Exec(q)
	return err
}
