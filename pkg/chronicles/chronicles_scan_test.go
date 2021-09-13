package chronicles

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"archive-my/models"
	"archive-my/pkg/testutil"
	"archive-my/pkg/utils"
)

type ChroniclesScanTestSuite struct {
	suite.Suite
	testutil.TestDBManager
	tx  *sql.Tx
	chr *Chronicles
}

func (s *ChroniclesScanTestSuite) SetupSuite() {
	s.NoError(utils.InitConfig("", "../../"))
	dbS, mdbS, err := s.InitTestDB()
	s.Require().Nil(err)
	s.Nil(s.MDB.Close())
	s.chr = new(Chronicles)
	s.chr.Init(dbS, mdbS)

}

func (s *ChroniclesScanTestSuite) TearDownSuite() {
	s.Require().Nil(s.DestroyTestDB())
}

func (s *ChroniclesScanTestSuite) SetupTest() {
	var err error
	s.tx, err = s.DB.Begin()
	s.NoError(err)
}

func (s *ChroniclesScanTestSuite) TearDownTest() {
	s.NoError(s.tx.Rollback())
}

//test functions
func (s *ChroniclesScanTestSuite) TestChronicles_simple() {
	userId := "client:local:01EXTPWJFV1SBZTSCE566SJ"
	_, filename, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(filename), ".", "test_data", "simple_scan_test.json")
	data, err := ioutil.ReadFile(path)
	s.NoError(err)
	resp := fmt.Sprintf(string(data), userId)
	s.NoError(err)
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		_, err = res.Write([]byte(resp))
		s.NoError(err)
	}))

	_, err = s.chr.scanEventsOnTx(s.tx, server.URL)
	s.NoError(err)
	countAll, err := models.Histories().Count(s.tx)
	s.NoError(err)
	s.Equal(int64(1), countAll)
	count, err := models.Histories(qm.Where("account_id = ?", userId)).Count(s.tx)
	s.NoError(err)
	s.Equal(int64(1), count)
}

func (s *ChroniclesScanTestSuite) TestChronicles_multiAcc() {
	userId1 := "client:local:01EXTPWJFV1SBZTSCE566SJ"
	userId2 := "client:local:01EXR0T3BCGH5P4V1Y1Q"
	_, filename, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(filename), ".", "test_data", "multiacc_scan_test.json")
	data, err := ioutil.ReadFile(path)
	s.NoError(err)
	resp := fmt.Sprintf(string(data), userId1, userId2)
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		_, err = res.Write([]byte(resp))
		s.NoError(err)
	}))

	_, err = s.chr.scanEventsOnTx(s.tx, server.URL)
	s.NoError(err)
	countAll, err := models.Histories().Count(s.tx)
	s.NoError(err)
	s.Equal(int64(2), countAll)
	count1, err := models.Histories(qm.Where("account_id = ?", userId1)).Count(s.tx)
	s.NoError(err)
	s.Equal(int64(1), count1)
	count2, err := models.Histories(qm.Where("account_id = ?", userId2)).Count(s.tx)
	s.NoError(err)
	s.Equal(int64(1), count2)
}

func (s *ChroniclesScanTestSuite) TestChronicles_multiEv() {
	userId1 := "client:local:01EXTPWJFV1SBZTSCE566SJ"
	userId2 := "client:local:01EXR0T3BCGH5P4V1Y1Q"
	updated1 := 2000.00
	cuID := utils.GenerateUID(8)
	_, filename, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(filename), ".", "test_data", "multiev1_scan_test.json")
	d, err := ioutil.ReadFile(path)
	s.NoError(err)
	resp := fmt.Sprintf(string(d), userId1, userId2, updated1, cuID)
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		_, err = res.Write([]byte(resp))
		s.NoError(err)
	}))

	_, err = s.chr.scanEventsOnTx(s.tx, server.URL)
	s.NoError(err)

	countAll, err := models.Histories().Count(s.tx)
	s.NoError(err)
	s.Equal(int64(2), countAll)

	hs, err := models.Histories(qm.Where("account_id = ?", userId1)).All(s.tx)
	s.NoError(err)
	s.Equal(1, len(hs))
	var data ChronicleEventData
	s.NoError(json.Unmarshal(hs[0].Data.JSON, &data))
	s.Equal(updated1, data.CurrentTime.Float64)

	count, err := models.Histories(qm.Where("account_id = ?", userId2)).Count(s.tx)
	s.NoError(err)
	s.Equal(int64(1), count)

	//second call
	updated1 = 5000.01
	path = filepath.Join(filepath.Dir(filename), ".", "test_data", "multiev2_scan_test.json")
	d, err = ioutil.ReadFile(path)
	s.NoError(err)
	resp = fmt.Sprintf(string(d), userId1, userId2, updated1, cuID)
	server = httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		_, err = res.Write([]byte(resp))
		s.NoError(err)
	}))

	_, err = s.chr.scanEventsOnTx(s.tx, server.URL)
	hs2, err := models.Histories(qm.Where("account_id = 'client:local:01EXTPWJFV1SBZTSCE566SJ'")).All(s.tx)
	s.NoError(err)
	s.Equal(1, len(hs2))
	var data2 ChronicleEventData
	s.NoError(json.Unmarshal(hs2[0].Data.JSON, &data2))
	s.Equal(updated1, data2.CurrentTime.Float64)

	count, err = models.Histories(qm.Where("account_id = 'client:local:01EXR0T3BCGH5P4V1Y1Q'")).Count(s.tx)
	s.NoError(err)
	s.Equal(int64(2), count)
}

func TestChroniclesScanTestSuite(t *testing.T) {
	suite.Run(t, new(ChroniclesScanTestSuite))
}
