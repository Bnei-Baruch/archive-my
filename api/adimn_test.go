package api

import (
	"bytes"
	"fmt"
	"github.com/volatiletech/null/v8"
	"net/http"

	"gopkg.in/square/go-jose.v2/json"

	"github.com/Bnei-Baruch/archive-my/databases/mydb/models"
)

//Admin tests
func (s *ApiTestSuite) TestAdmin_permissions() {
	user := s.CreateUser()

	req, _ := http.NewRequest(http.MethodGet, "/admin/labels", nil)
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Require().Equal(http.StatusForbidden, resp.Code)

	s.assertTokenVerifier()
	s.apiAuthP(req, user.AccountsID, []string{"kmedia_moderator"})
	resp = s.request(req)
	s.Require().Equal(http.StatusOK, resp.Code)
}

func (s *ApiTestSuite) TestAdmin_getLabels() {
	admin := s.CreateUser()
	user1 := s.CreateUser()
	user2 := s.CreateUser()

	// no Labels whatsoever
	req, _ := http.NewRequest(http.MethodGet, "/admin/labels", nil)
	s.apiAuthP(req, admin.AccountsID, []string{"kmedia_moderator"})
	var resp GetLabelsResponse
	s.request200json(req, &resp)

	s.EqualValues(0, resp.Total, "total")
	s.Empty(resp.Items, "items empty")

	s.assertTokenVerifier()
	labels := make([]*models.Label, 6)
	for i := range labels {
		odd := i%2 == 0
		if odd {
			labels[i] = s.CreateLabel(user1, fmt.Sprintf("Label-%d", i), "", "en", nil)
		} else {
			labels[i] = s.CreateLabel(user2, fmt.Sprintf("Label-%d", i), "", "en", nil)
		}
	}

	req, _ = http.NewRequest(http.MethodGet, "/admin/labels?order_by=id", nil)
	s.apiAuthP(req, admin.AccountsID, []string{"kmedia_moderator"})
	s.request200json(req, &resp)

	s.EqualValues(len(labels), resp.Total, "total")
	s.Require().Len(resp.Items, len(labels), "items length")
	for i, x := range resp.Items {
		s.assertLabel(labels[i], x, i)
	}
}

func (s *ApiTestSuite) TestAdmin_setAccept() {
	admin := s.CreateUser()
	user := s.CreateUser()
	b := s.CreateLabel(user, "Label", "", "en", nil)

	var resp GetLabelsResponse
	req, _ := http.NewRequest(http.MethodGet, "/admin/labels", nil)
	s.apiAuthP(req, admin.AccountsID, []string{"kmedia_moderator"})
	s.request200json(req, &resp)
	s.EqualValues(1, resp.Total, "total")
	s.assertLabel(b, resp.Items[0], 1)

	s.assertTokenVerifier()
	payload, err := json.Marshal(LabelModerationRequest{Accepted: null.BoolFrom(false)})
	s.NoError(err)
	req, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/admin/labels/%d", b.ID), bytes.NewReader(payload))
	s.apiAuthP(req, admin.AccountsID, []string{"kmedia_moderator"})

	var respChanged Label
	s.request200json(req, &respChanged)
	s.EqualValues(null.Bool{Bool: false, Valid: true}, respChanged.Accepted)

}
