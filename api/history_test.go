package api

import (
	"fmt"
	"net/http"

	"github.com/Bnei-Baruch/archive-my/databases/mydb/models"
)

func (s *ApiTestSuite) TestHistory_getHistory() {
	user := s.CreateUser()

	// no histories whatsoever
	req, _ := http.NewRequest(http.MethodGet, "/rest/history", nil)
	s.apiAuthUser(req, user)
	var resp HistoryResponse
	s.request200json(req, &resp)

	s.EqualValues(0, resp.Total, "total")
	s.Empty(resp.Items, "items empty")

	// with histories
	histories := make([]*models.History, 5)
	for i := range histories {
		histories[i] = s.CreateHistory(user)
	}

	req, _ = http.NewRequest(http.MethodGet, "/rest/history?order_by=id", nil)
	s.apiAuthUser(req, user)
	s.request200json(req, &resp)

	s.EqualValues(len(histories), resp.Total, "total")
	s.Require().Len(resp.Items, len(histories), "items length")
	for i, x := range resp.Items {
		s.assertHistory(histories[i], x, i)
	}

	// other users see no histories
	s.assertTokenVerifier()
	otherUser := s.CreateUser()
	req, _ = http.NewRequest(http.MethodGet, "/rest/history", nil)
	s.apiAuthUser(req, otherUser)
	s.request200json(req, &resp)
	s.EqualValues(0, resp.Total, "total")
	s.Empty(resp.Items, "items empty")
}

func (s *ApiTestSuite) TestHistory_getHistory_pagination() {
	user := s.CreateUser()
	histories := make([]*models.History, 10)
	for i := range histories {
		histories[i] = s.CreateHistory(user)
	}

	for i := 0; i < 2; i++ {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/rest/history?page_no=%d&page_size=%d&order_by=id", i+1, 5), nil)
		s.apiAuthUser(req, user)
		var resp HistoryResponse
		s.request200json(req, &resp)

		s.EqualValues(len(histories), resp.Total, "total")
		s.Require().Len(resp.Items, 5, "items length")
		for j, x := range resp.Items {
			s.assertHistory(histories[(i*5)+j], x, j)
		}
	}
}

func (s *ApiTestSuite) TestHistory_deleteHistory() {
	user := s.CreateUser()
	history := s.CreateHistory(user)

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/rest/history/%d", history.ID), nil)
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Equal(http.StatusOK, resp.Code)

	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/rest/history/%d", history.ID), nil)
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)
}

//help functions

func (s *ApiTestSuite) assertHistory(expected *models.History, actual *History, idx int) {
	s.Equal(expected.ID, actual.ID, "ID [%d]", idx)
	s.Equal(expected.ContentUnitUID, actual.ContentUnitUID, "ContentUnitUID [%d]", idx)
}
