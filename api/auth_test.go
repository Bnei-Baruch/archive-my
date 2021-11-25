package api

import (
	"math/rand"
	"net/http"

	"github.com/Bnei-Baruch/archive-my/databases/mydb/models"
	"github.com/Bnei-Baruch/archive-my/pkg/utils"
)

func (s *ApiTestSuite) TestAuth_createUser() {
	requests := make([]*http.Request, rand.Intn(100)+10)
	subjects := make([]string, 0)
	var sub string
	var resp HistoryResponse

	for i, _ := range requests {
		if i%3 == 0 {
			sub = utils.GenerateName(10)
			subjects = append(subjects, sub)
		}
		s.assertTokenVerifier()
		r, _ := http.NewRequest(http.MethodGet, "/rest/history", nil)
		s.apiAuthP(r, sub, nil)
		s.request200json(r, &resp)
	}

	count, err := models.Users(models.UserWhere.AccountsID.IN(subjects)).Count(s.MyDB.DB)
	s.NoError(err)
	s.EqualValues(len(subjects), count)
}
