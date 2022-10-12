package chronicles

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/Bnei-Baruch/archive-my/common"
)

func (s *ChroniclesTestSuite) TestChronicles_brokenJson() {
	user := s.CreateUser()

	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		data := s.readTestResource("broken_json_test.json")
		_, err := res.Write([]byte(fmt.Sprintf(string(data), user.AccountsID)))
		s.Require().NoError(err)
	}))
	defer server.Close()
	common.Config.ChroniclesUrl = server.URL

	chr := new(Chronicles)
	chr.InitWithDeps(s.MyDB.DB, s.MDB.DB)

	_, err := chr.scanEvents()
	s.Require().Error(err)
}
