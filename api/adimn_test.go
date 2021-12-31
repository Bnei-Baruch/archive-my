package api

/*
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

	req, _ := http.NewRequest(http.MethodGet, "/admin/bookmarks", nil)
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Require().Equal(http.StatusForbidden, resp.Code)

	s.assertTokenVerifier()
	s.apiAuthP(req, user.AccountsID, []string{"kmedia_moderator"})
	resp = s.request(req)
	s.Require().Equal(http.StatusOK, resp.Code)
}

func (s *ApiTestSuite) TestAdmin_getBookmarks() {
	admin := s.CreateUser()
	user1 := s.CreateUser()
	user2 := s.CreateUser()

	// no Bookmarks whatsoever
	req, _ := http.NewRequest(http.MethodGet, "/admin/bookmarks", nil)
	s.apiAuthP(req, admin.AccountsID, []string{"kmedia_moderator"})
	var resp GetBookmarksResponse
	s.request200json(req, &resp)

	s.EqualValues(0, resp.Total, "total")
	s.Empty(resp.Items, "items empty")

	s.assertTokenVerifier()
	bookmarks := make([]*models.Bookmark, 6)
	for i := range bookmarks {
		odd := i%2 == 0
		if odd {
			bookmarks[i] = s.CreateBookmark(user1, fmt.Sprintf("Bookmark-%d", i), "", nil, true)
			_ = s.CreateBookmark(user2, fmt.Sprintf("Bookmark-%d", i), "", nil, false)
		} else {
			_ = s.CreateBookmark(user1, fmt.Sprintf("Bookmark-%d", i), "", nil, false)
			bookmarks[i] = s.CreateBookmark(user2, fmt.Sprintf("Bookmark-%d", i), "", nil, true)
		}
	}

	req, _ = http.NewRequest(http.MethodGet, "/admin/labels?order_by=id", nil)
	s.apiAuthP(req, admin.AccountsID, []string{"kmedia_moderator"})
	s.request200json(req, &resp)

	s.EqualValues(len(bookmarks), resp.Total, "total")
	s.Require().Len(resp.Items, len(bookmarks), "items length")
	for i, x := range resp.Items {
		s.assertBookmark(bookmarks[i], x, i)
	}
}

func (s *ApiTestSuite) TestAdmin_setAccept() {
	admin := s.CreateUser()
	user := s.CreateUser()
	b := s.CreateBookmark(user, "Bookmark", "", nil, true)

	var resp GetLabelsResponse
	req, _ := http.NewRequest(http.MethodGet, "/admin/bookmarks", nil)
	s.apiAuthP(req, admin.AccountsID, []string{"kmedia_moderator"})
	s.request200json(req, &resp)
	s.EqualValues(1, resp.Total, "total")
	//s.assertBookmark(b, resp.Items[0], 1)

	s.assertTokenVerifier()
	payload, err := json.Marshal(LabelModerationRequest{Accepted: null.BoolFrom(false)})
	s.NoError(err)
	req, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/admin/bookmarks/%d/accept", b.ID), bytes.NewReader(payload))
	s.apiAuthP(req, admin.AccountsID, []string{"kmedia_moderator"})

	var respChanged Label
	s.request200json(req, &respChanged)
	s.EqualValues(false, respChanged.Accepted)

}
*/
