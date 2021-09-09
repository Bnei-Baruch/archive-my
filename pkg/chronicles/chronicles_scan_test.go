package chronicles

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"archive-my/pkg/testutil"
	"archive-my/pkg/utils"
)

type ChroniclesScanTestSuite struct {
	suite.Suite
	testutil.TestDBManager
	tx             *sql.Tx
	chr            *Chronicles
	mockChronicles *httptest.Server
}

func (s *ChroniclesScanTestSuite) SetupSuite() {
	s.NoError(utils.InitConfig("", "../../"))
	dbS, mdbS, err := s.InitTestDB()
	s.Require().Nil(err)
	s.Nil(s.DB.Close())
	s.Nil(s.MDB.Close())
	s.chr = new(Chronicles)
	s.mockChronicles = httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		//res.WriteHeader(scenario.expectedRespStatus)
		res.Write([]byte("body"))
	}))
	s.chr.Init(dbS, mdbS, s.mockChronicles.Client())

}

func (s *ChroniclesScanTestSuite) TearDownSuite() {
	s.DestroyTestDB()
	//s.Require().Nil(s.DestroyTestDB())
}

func (s *ChroniclesScanTestSuite) SetupTest() {
	s.chr.Run()
}

func (s *ChroniclesScanTestSuite) TearDownTest() {
	err := s.tx.Rollback()
	s.Require().Nil(err)
	s.chr.Stop()
}

//test functions
func (s *ChroniclesScanTestSuite) TestChronicles() {
	s.Nil(s.chr.interval)
}

//help functions
func TestChroniclesScanTestSuite(t *testing.T) {
	suite.Run(t, new(ChroniclesScanTestSuite))
}
