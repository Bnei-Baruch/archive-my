package api

import (
	"database/sql"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"

	"github.com/Bnei-Baruch/archive-my/pkg/testutil"
	"github.com/Bnei-Baruch/archive-my/pkg/utils"
)

type RestTestSuite struct {
	suite.Suite
	testutil.TestDBManager
	tx   *sql.Tx
	ctx  *gin.Context
	kcId string
	app  *App
}

func (s *RestTestSuite) SetupSuite() {
	s.NoError(utils.InitConfig("", "../"))
	s.app = new(App)
	_, _, err := s.InitTestDB()
	s.Require().Nil(err)
	s.app.SetDB(s.DB)

	verifier := testutil.OIDCTokenVerifier{}
	s.app.SetVerifier(&verifier)
	s.kcId = testutil.KEYCKLOAK_ID

	s.ctx = &gin.Context{}
	s.ctx.Set("KC_ID", s.kcId)
}

func (s *RestTestSuite) TearDownSuite() {
	s.DestroyTestDB()
	//s.Require().Nil(s.DestroyTestDB())
}

func (s *RestTestSuite) SetupTest() {
	var err error
	s.tx, err = s.DB.Begin()
	s.Require().Nil(err)
}

func (s *RestTestSuite) TearDownTest() {
	err := s.tx.Rollback()
	s.Require().Nil(err)
}

func TestApiBaseTestSuite(t *testing.T) {
	suite.Run(t, new(RestTestSuite))
}
