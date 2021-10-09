package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Bnei-Baruch/archive-my/models"
)

func (s *ApiTestSuite) TestPlaylist_getPlaylists() {
	user := s.CreateUser()

	// no playlists whatsoever
	req, _ := http.NewRequest(http.MethodGet, "/rest/playlists", nil)
	s.apiAuthUser(req, user)
	var resp PlaylistsResponse
	s.request200json(req, &resp)

	s.EqualValues(0, resp.Total, "total")
	s.Empty(resp.Items, "items empty")

	// with playlists
	playlists := make([]*models.Playlist, 5)
	for i := range playlists {
		playlists[i] = s.CreatePlaylist(user, fmt.Sprintf("playlist-%d", i))
	}

	req, _ = http.NewRequest(http.MethodGet, "/rest/playlists?order_by=id", nil)
	s.apiAuthUser(req, user)
	s.request200json(req, &resp)

	s.EqualValues(len(playlists), resp.Total, "total")
	s.Require().Len(resp.Items, len(playlists), "items length")
	for i, x := range resp.Items {
		s.assertPlaylist(playlists[i], x, i)
	}

	// other users see no playlists
	s.assertTokenVerifier()
	otherUser := s.CreateUser()
	req, _ = http.NewRequest(http.MethodGet, "/rest/playlists", nil)
	s.apiAuthUser(req, otherUser)
	s.request200json(req, &resp)
	s.EqualValues(0, resp.Total, "total")
	s.Empty(resp.Items, "items empty")
}

func (s *ApiTestSuite) TestPlaylist_getPlaylists_pagination() {
	user := s.CreateUser()
	playlists := make([]*models.Playlist, 10)
	for i := range playlists {
		playlists[i] = s.CreatePlaylist(user, fmt.Sprintf("playlist-%d", i))
	}

	for i := 0; i < 2; i++ {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/rest/playlists?page_no=%d&page_size=%d&order_by=id", i+1, 5), nil)
		s.apiAuthUser(req, user)
		var resp PlaylistsResponse
		s.request200json(req, &resp)

		s.EqualValues(len(playlists), resp.Total, "total")
		s.Require().Len(resp.Items, 5, "items length")
		for j, x := range resp.Items {
			s.assertPlaylist(playlists[(i*5)+j], x, j)
		}
	}
}

func (s *ApiTestSuite) TestPlaylist_createPlaylist_badRequest() {
	user := s.CreateUser()

	// bad properties json
	payload, err := json.Marshal(map[string]interface{}{
		"name":       "test playlist",
		"properties": "malformed json {}",
	})
	s.Require().NoError(err)
	req, _ := http.NewRequest(http.MethodPost, "/rest/playlists", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Require().Equal(http.StatusBadRequest, resp.Code)

	// too long name
	payload, err = json.Marshal(map[string]interface{}{
		"name": strings.Repeat("*", 257),
	})
	s.Require().NoError(err)
	req, _ = http.NewRequest(http.MethodPost, "/rest/playlists", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Require().Equal(http.StatusBadRequest, resp.Code)
}

func (s *ApiTestSuite) TestPlaylist_createPlaylist() {
	user := s.CreateUser()

	// all defaults
	req, _ := http.NewRequest(http.MethodPost, "/rest/playlists", bytes.NewReader([]byte("{}")))
	s.apiAuthUser(req, user)
	var resp Playlist
	s.request200json(req, &resp)

	s.NotZero(resp.ID, "ID")
	s.NotZero(resp.UID, "UID")
	s.Empty(resp.Name, "test playlist", "Name")
	s.False(resp.Public, "Public")
	s.Empty(resp.Properties, "props")

	// custom params
	payload, err := json.Marshal(map[string]interface{}{
		"name": "test playlist",
		"properties": map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	})
	s.Require().NoError(err)
	req, _ = http.NewRequest(http.MethodPost, "/rest/playlists", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	s.request200json(req, &resp)

	s.NotZero(resp.ID, "ID")
	s.NotZero(resp.UID, "UID")
	s.Equal(resp.Name, "test playlist", "Name")
	s.False(resp.Public, "Public")
	s.Len(resp.Properties, 2, "props count")
	s.Equal(resp.Properties["key1"], "value1", "prop 1")
	s.Equal(resp.Properties["key2"], "value2", "prop 2")
}

func (s *ApiTestSuite) TestPlaylist_getPlaylist_notFound() {
	user := s.CreateUser()

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/rest/playlists/%d", 1), nil)
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)

	s.assertTokenVerifier()
	otherUser := s.CreateUser()
	playlist := s.CreatePlaylist(user, "playlist")
	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/rest/playlists/%d", playlist.ID), nil)
	s.apiAuthUser(req, otherUser)
	resp = s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)
}

func (s *ApiTestSuite) TestPlaylist_getPlaylist() {
	user := s.CreateUser()
	playlist := s.CreatePlaylist(user, "playlist")
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/rest/playlists/%d", playlist.ID), nil)
	s.apiAuthUser(req, user)
	var resp Playlist
	s.request200json(req, &resp)
	s.assertPlaylist(playlist, &resp, 0)
}

//
//func (s *ApiTestSuite) TestPlaylist_add() {
//	newPl := models.Playlist{ID: 1, Name: null.String{String: "new playlist", Valid: true}}
//	cAdd, _, err := testutil.PrepareContext(newPl)
//	s.NoError(err)
//	respAdd, err := s.app.createPlaylist(cAdd, s.mydb.DB)
//	s.NoError(err)
//	s.Equal(newPl.ID, respAdd[0].ID)
//
//	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
//	s.NoError(err)
//	var resp playlistsResponse
//	s.app.getPlaylists(c, s.mydb.DB)
//	s.NoError(json.Unmarshal(w.Body.Bytes(), &resp))
//	s.EqualValues(1, resp.Total, "total")
//	s.Equal(respAdd[0].ID, resp.Playlists[0].ID)
//	s.Equal(respAdd[0].Name, resp.Playlists[0].Name)
//	s.Equal(s.kcId, resp.Playlists[0].AccountID)
//}
//
//func (s *ApiTestSuite) TestPlaylist_remove() {
//	items := s.createDummyPlaylists(10, "playlist for remove")
//	itemR1 := items[rand.Intn(10-1)+1]
//	ids := &IDsFilter{IDs: []int64{itemR1.ID}}
//	cDel, _, err := testutil.PrepareContext(ids)
//	s.NoError(err)
//	s.Nil(s.app.deletePlaylist(cDel, s.mydb.DB))
//
//	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 20})
//	var resp playlistsResponse
//	s.app.getPlaylists(c, s.mydb.DB)
//	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
//	s.EqualValues(9, resp.Total, "total")
//	s.NotContains(resp.Playlists, itemR1)
//}
//
//func (s *ApiTestSuite) TestPlaylist_removeOtherAcc() {
//	items := s.createDummyPlaylists(10, "playlists for remove")
//
//	item := items[rand.Intn(10-1)+1]
//	item.AccountID = "new_account_id"
//	c, err := item.Update(s.mydb.DB, boil.Infer())
//	s.NoError(err)
//	s.Equal(c, int64(1))
//
//	ids := &IDsFilter{IDs: []int64{item.ID}}
//	cDel, _, err := testutil.PrepareContext(ids)
//	s.Require().NoError(err)
//	s.Nil(s.app.deletePlaylist(cDel, s.mydb.DB))
//	count, err := models.Playlists().Count(s.mydb.DB)
//	s.NoError(err)
//	s.Equal(int64(10), count)
//}
//
//help functions

func (s *ApiTestSuite) assertPlaylist(expected *models.Playlist, actual *Playlist, idx int) {
	s.Equal(expected.ID, actual.ID, "ID [%d]", idx)
	s.Equal(expected.UserID, actual.UserID, "UserID [%d]", idx)
	s.Equal(expected.UID, actual.UID, "UID [%d]", idx)
	s.Equal(expected.Name.String, actual.Name, "Name [%d]", idx)
	s.Equal(expected.Public, actual.Public, "Public [%d]", idx)

	if expected.R.PlaylistItems != nil {
		s.Equal(len(expected.R.PlaylistItems), actual.TotalItems, "TotalItems [%d]", idx)
		if len(actual.Items) > 0 {
			s.Len(actual.Items, len(expected.R.PlaylistItems), "playlist.PlaylistItems len [%d]", idx)
			for i, u := range expected.R.PlaylistItems {
				s.assertPlaylistItem(u, actual.Items[i], i)
			}
		}
	}
}

func (s *ApiTestSuite) assertPlaylistItem(expected *models.PlaylistItem, actual *PlaylistItem, idx int) {
	s.Equal(expected.ID, actual.ID, "PlaylistItem ID [%d]", idx)
	s.Equal(expected.Position, actual.Position, "PlaylistItem Position [%d]", idx)
	s.Equal(expected.ContentUnitUID, actual.ContentUnitUID, "PlaylistItem ContentUnitUID [%d]", idx)
}
