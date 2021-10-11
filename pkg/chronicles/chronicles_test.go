package chronicles

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Bnei-Baruch/archive-my/pkg/testutil"
)

type ChroniclesTestSuite struct {
	suite.Suite
	testutil.TestDBManager
	tx             *sql.Tx
	chr            *Chronicles
	mockChronicles *httptest.Server
}

func (s *ChroniclesTestSuite) SetupSuite() {
	dbS, mdbS, err := s.InitTestDB()
	s.Require().NoError(err)
	s.Nil(s.DB.Close())
	s.Nil(s.MDB.Close())
	s.chr = new(Chronicles)
	s.mockChronicles = httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		//res.WriteHeader(scenario.expectedRespStatus)
		res.Write([]byte("body"))
	}))
	s.chr.Init(dbS, mdbS)

}

func (s *ChroniclesTestSuite) TearDownSuite() {
	s.Require().Nil(s.DestroyTestDB())
}

func (s *ChroniclesTestSuite) SetupTest() {
	s.chr.Run()
}

func (s *ChroniclesTestSuite) TearDownTest() {
	err := s.tx.Rollback()
	s.Require().NoError(err)
	s.chr.Stop()
}

//test functions
/*func (s *ChroniclesTestSuite) TestChronicles() {
	s.Nil(s.chr.interval)
}
*/
//help functions
func TestChroniclesTestSuite(t *testing.T) {
	suite.Run(t, new(ChroniclesTestSuite))
}
