package chronicles

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Bnei-Baruch/archive-my/databases/mdb"
	"github.com/Bnei-Baruch/archive-my/domain"
	"github.com/Bnei-Baruch/archive-my/pkg/testutil"
)

type ChroniclesTestSuite struct {
	domain.ModelsSuite
	MDB *testutil.TestMDBManager
}

func (s *ChroniclesTestSuite) SetupSuite() {
	s.MyDB = new(testutil.TestMyDBManager)
	s.Require().NoError(s.MyDB.Init())
	s.MDB = new(testutil.TestMDBManager)
	s.Require().NoError(s.MDB.Init())
	s.Require().NoError(mdb.InitCT(s.MDB.DB))
}

func (s *ChroniclesTestSuite) TearDownSuite() {
	s.Require().NoError(s.MyDB.Destroy())
	s.Require().NoError(s.MDB.Destroy())
}

func (s *ChroniclesTestSuite) SetupTest() {
	s.MyDB.DBCleaner.Acquire(s.MyDB.AllTables()...)
	s.MDB.DBCleaner.Acquire(s.MDB.AllTables()...)
}

func (s *ChroniclesTestSuite) TearDownTest() {
	s.MyDB.DBCleaner.Clean(s.MyDB.AllTables()...)
	s.MDB.DBCleaner.Clean(s.MDB.AllTables()...)
}

func TestChroniclesTestSuite(t *testing.T) {
	suite.Run(t, new(ChroniclesTestSuite))
}

func (s *ChroniclesTestSuite) readTestResource(name string) []byte {
	_, filename, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(filename), ".", "test_data", name)
	data, err := os.ReadFile(path)
	s.Require().NoError(err)
	return data
}
