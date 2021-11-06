package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/coreos/go-oidc"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/Bnei-Baruch/archive-my/common"
	"github.com/Bnei-Baruch/archive-my/databases/mdb"
	"github.com/Bnei-Baruch/archive-my/databases/mydb/models"
	"github.com/Bnei-Baruch/archive-my/domain"
	"github.com/Bnei-Baruch/archive-my/middleware"
	"github.com/Bnei-Baruch/archive-my/pkg/testutil"
	"github.com/Bnei-Baruch/archive-my/pkg/testutil/mocks"
)

type ApiTestSuite struct {
	domain.ModelsSuite
	MDB           *testutil.TestMDBManager
	tokenVerifier *mocks.OIDCTokenVerifier
	app           *App
}

func (s *ApiTestSuite) SetupSuite() {
	common.Config.GinMode = "test"
	s.app = new(App)
	s.MyDB = new(testutil.TestMyDBManager)
	s.Require().NoError(s.MyDB.Init())
	s.MDB = new(testutil.TestMDBManager)
	s.Require().NoError(s.MDB.Init())
	s.Require().NoError(mdb.InitCT(s.MDB.DB))
	s.tokenVerifier = new(mocks.OIDCTokenVerifier)
	s.app.InitializeWithDeps(s.MyDB.DB, s.MDB.DB, s.tokenVerifier)
}

func (s *ApiTestSuite) TearDownSuite() {
	s.Require().Nil(s.MyDB.Destroy())
}

func (s *ApiTestSuite) SetupTest() {
	s.MyDB.DBCleaner.Acquire(s.MyDB.AllTables()...)
}

func (s *ApiTestSuite) TearDownTest() {
	s.assertTokenVerifier()
	s.MyDB.DBCleaner.Clean(s.MyDB.AllTables()...)
}

func TestApiBaseTestSuite(t *testing.T) {
	suite.Run(t, new(ApiTestSuite))
}

func (s *ApiTestSuite) assertTokenVerifier() {
	s.tokenVerifier.AssertExpectations(s.T())
	s.tokenVerifier.ExpectedCalls = nil
	s.tokenVerifier.Calls = nil
}

func (s *ApiTestSuite) request(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.app.Router.ServeHTTP(rr, req)
	return rr
}

func (s *ApiTestSuite) request200json(req *http.Request, body interface{}) {
	s.requestJson(req, http.StatusOK, body)
}

func (s *ApiTestSuite) request201json(req *http.Request, body interface{}) {
	s.requestJson(req, http.StatusCreated, body)
}

func (s *ApiTestSuite) requestJson(req *http.Request, statusCode int, body interface{}) {
	resp := s.request(req)
	s.Require().Equal(statusCode, resp.Code)
	s.Require().NoError(json.Unmarshal(resp.Body.Bytes(), &body))
}

func (s *ApiTestSuite) apiAuth(req *http.Request) {
	s.apiAuthP(req, "Subject", nil)
}

func (s *ApiTestSuite) apiAuthUser(req *http.Request, user *models.User) {
	s.apiAuthP(req, user.AccountsID, nil)
}

func (s *ApiTestSuite) apiAuthP(req *http.Request, subject string, roles []string) {
	req.Header.Set("Authorization", "Bearer token")

	oidcIDToken := &oidc.IDToken{
		Issuer:          "https://test.issuer",
		Audience:        []string{"Audience"},
		Subject:         subject,
		Expiry:          time.Now().Add(10 * time.Minute),
		IssuedAt:        time.Now(),
		Nonce:           "nonce",
		AccessTokenHash: "access_token_hash",
	}

	claims := middleware.IDTokenClaims{
		Aud: oidcIDToken.Audience,
		Exp: int(oidcIDToken.Expiry.Unix()),
		Iat: int(oidcIDToken.IssuedAt.Unix()),
		Iss: oidcIDToken.Issuer,
		RealmAccess: middleware.Roles{
			Roles: roles,
		},
		Sub: oidcIDToken.Subject,
	}

	b, err := json.Marshal(claims)
	s.Require().NoError(err, "json.Marshal(claims)")

	pointerVal := reflect.ValueOf(oidcIDToken)
	val := reflect.Indirect(pointerVal)
	member := val.FieldByName("claims")
	ptrToY := unsafe.Pointer(member.UnsafeAddr())
	realPtrToY := (*[]byte)(ptrToY)
	*realPtrToY = b

	s.tokenVerifier.On("Verify", mock.Anything, "token").Return(oidcIDToken, nil)
}
