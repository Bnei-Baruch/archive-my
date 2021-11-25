package api

import (
	"bytes"
	"encoding/json"
	"fmt"
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
		s.assertReaction(reactions[i], x, i)
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
			s.assertReaction(reactions[(i*5)+j], x, j)
		}
	}
}

func (s *ApiTestSuite) TestReactions_getReactions_filtered() {
	user := s.CreateUser()
	rClips := make([]*models.Reaction, 2)
	rClips[0] = s.CreateReaction(user, "like", "CLIPS", utils.GenerateUID(8))
	rClips[1] = s.CreateReaction(user, "super", "CLIPS", utils.GenerateUID(8))
	rClip := s.CreateReaction(user, "like", "CLIP", utils.GenerateUID(8))

	//cant filter by UID without subject type
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/rest/reactions?uids=%s", rClip.SubjectUID), nil)
	s.apiAuthUser(req, user)
	s.Equal(http.StatusBadRequest, s.request(req).Code)

	var resp ReactionsResponse
	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/rest/reactions?subject_type=%s&order_by=id", rClips[0].SubjectType), nil)
	s.apiAuthUser(req, user)
	s.request200json(req, &resp)
	s.Equal(int64(2), resp.Total, "total by SubjectType")
	for i, x := range resp.Items {
		s.assertReaction(rClips[i], x, i)
	}

	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/rest/reactions?uids=%s&subject_type=%s&order_by=id", rClips[0].SubjectUID, rClip.SubjectType), nil)
	s.apiAuthUser(req, user)
	s.request200json(req, &resp)
	s.Equal(int64(0), resp.Total, "total by SubjectType and wrong UID")

	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/rest/reactions?uids=%s&subject_type=%s", rClips[0].SubjectUID, rClips[0].SubjectType), nil)
	s.apiAuthUser(req, user)
	s.request200json(req, &resp)
	s.Equal(int64(1), resp.Total, "total by SubjectType and UID")
	s.assertReaction(rClips[0], resp.Items[0], 1)

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

func (s *ApiTestSuite) assertReaction(expected *models.Reaction, actual *Reaction, idx int) {
	s.Equal(expected.SubjectUID, actual.SubjectUID, "SubjectUID [%d]", idx)
	s.Equal(expected.SubjectType, actual.SubjectType, "SubjectType [%d]", idx)
}
