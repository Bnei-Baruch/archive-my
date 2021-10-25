package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/Bnei-Baruch/archive-my/databases/mydb/models"
	"github.com/Bnei-Baruch/archive-my/pkg/utils"
)

func (s *ApiTestSuite) TestReactions_getReactions() {
	user := s.CreateUser()

	// no reactions whatsoever
	req, _ := http.NewRequest(http.MethodGet, "/rest/reactions", nil)
	s.apiAuthUser(req, user)
	var resp ReactionsResponse
	s.request200json(req, &resp)

	s.EqualValues(0, resp.Total, "total")
	s.Empty(resp.Items, "items empty")

	// with reactions
	reactions := make([]*models.Reaction, 5)
	for i := range reactions {
		reactions[i] = s.CreateReaction(user, "", "", "")
	}

	req, _ = http.NewRequest(http.MethodGet, "/rest/reactions?order_by=id", nil)
	s.apiAuthUser(req, user)
	s.request200json(req, &resp)

	s.EqualValues(len(reactions), resp.Total, "total")
	s.Require().Len(resp.Items, len(reactions), "items length")
	for i, x := range resp.Items {
		s.assertReactions(reactions[i], x, i)
	}

	// other users see no reactions
	s.assertTokenVerifier()
	otherUser := s.CreateUser()
	req, _ = http.NewRequest(http.MethodGet, "/rest/reactions", nil)
	s.apiAuthUser(req, otherUser)
	s.request200json(req, &resp)
	s.EqualValues(0, resp.Total, "total")
	s.Empty(resp.Items, "items empty")
}

func (s *ApiTestSuite) TestReactions_getReactions_pagination() {
	user := s.CreateUser()
	reactions := make([]*models.Reaction, 10)
	for i := range reactions {
		reactions[i] = s.CreateReaction(user, "", "", "")
	}

	for i := 0; i < 2; i++ {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/rest/reactions?page_no=%d&page_size=%d&order_by=id", i+1, 5), nil)
		s.apiAuthUser(req, user)
		var resp ReactionsResponse
		s.request200json(req, &resp)

		s.EqualValues(len(reactions), resp.Total, "total")
		s.Require().Len(resp.Items, 5, "items length")
		for j, x := range resp.Items {
			s.assertReactions(reactions[(i*5)+j], x, j)
		}
	}
}

func (s *ApiTestSuite) TestReactions_deleteReactions_notFound() {
	user := s.CreateUser()

	kind := utils.GenerateName(rand.Intn(15) + 1)
	sType := utils.GenerateName(rand.Intn(15) + 1)
	sUID := utils.GenerateUID(8)
	payload, err := json.Marshal(map[string]interface{}{
		"kind":         kind,
		"subject_type": sType,
		"subject_uid":  sUID,
	})
	s.NoError(err)

	req, _ := http.NewRequest(http.MethodDelete, "/rest/reactions", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)

	_ = s.CreateReaction(user, kind, sType, sUID)

	//try to remove with other user
	s.assertTokenVerifier()
	otherUser := s.CreateUser()
	s.apiAuthUser(req, otherUser)
	req, _ = http.NewRequest(http.MethodDelete, "/rest/reactions", bytes.NewReader(payload))
	resp = s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)
}

func (s *ApiTestSuite) TestReactions_deleteReactions() {
	user := s.CreateUser()

	reaction := s.CreateReaction(user, "", "", "")

	payload, err := json.Marshal(map[string]interface{}{
		"kind":         reaction.Kind,
		"subject_type": reaction.SubjectType,
		"subject_uid":  reaction.SubjectUID,
	})
	s.NoError(err)

	req, _ := http.NewRequest(http.MethodDelete, "/rest/reactions", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Equal(http.StatusOK, resp.Code)
}

func (s *ApiTestSuite) TestReactions_deleteReactions_fail() {
	user := s.CreateUser()

	reaction := s.CreateReaction(user, "", "", "")

	//remove without kind
	payload, err := json.Marshal(map[string]interface{}{
		"subject_type": reaction.SubjectType,
		"subject_uid":  reaction.SubjectUID,
	})
	s.NoError(err)
	req, _ := http.NewRequest(http.MethodDelete, "/rest/reactions", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Equal(http.StatusBadRequest, resp.Code)

	//remove without subject type
	payload, err = json.Marshal(map[string]interface{}{
		"kind":        reaction.Kind,
		"subject_uid": reaction.SubjectUID,
	})
	s.NoError(err)
	req, _ = http.NewRequest(http.MethodDelete, "/rest/reactions", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Equal(http.StatusBadRequest, resp.Code)

	//remove without subject UID
	payload, err = json.Marshal(map[string]interface{}{
		"kind":         reaction.Kind,
		"subject_type": reaction.SubjectType,
	})
	s.NoError(err)
	req, _ = http.NewRequest(http.MethodDelete, "/rest/reactions", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Equal(http.StatusBadRequest, resp.Code)
}

//help functions

func (s *ApiTestSuite) assertReactions(expected *models.Reaction, actual *Reaction, idx int) {
	s.Equal(expected.SubjectUID, actual.SubjectUID, "SubjectUID [%d]", idx)
	s.Equal(expected.SubjectType, actual.SubjectType, "SubjectType [%d]", idx)
}
