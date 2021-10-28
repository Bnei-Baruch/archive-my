package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/Bnei-Baruch/archive-my/databases/mydb/models"
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
		playlists[i] = s.CreatePlaylist(user, fmt.Sprintf("playlist-%d", i), nil)
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

func (s *ApiTestSuite) TestPlaylist_getPlaylists_with_exist() {
	user := s.CreateUser()

	playlist := s.CreatePlaylist(user, "playlist with exist", nil)
	_ = s.CreatePlaylist(user, "playlist with no exist", nil)
	cuUID := playlist.R.PlaylistItems[0].ContentUnitUID

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/rest/playlists?exist_cu=%s&order_by=id", cuUID), nil)
	s.apiAuthUser(req, user)

	var resp PlaylistsResponse
	s.request200json(req, &resp)

	s.EqualValues(2, resp.Total, "total")
	s.Equal(playlist.ID, resp.Items[0].ID)
	s.EqualValues(1, len(resp.Items[0].Items), "number items")
	s.EqualValues(cuUID, resp.Items[0].Items[0].ContentUnitUID, "wright unit")
}

func (s *ApiTestSuite) TestPlaylist_getPlaylists_pagination() {
	user := s.CreateUser()
	playlists := make([]*models.Playlist, 10)
	for i := range playlists {
		playlists[i] = s.CreatePlaylist(user, fmt.Sprintf("playlist-%d", i), nil)
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
	playlist := s.CreatePlaylist(user, "playlist", nil)
	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/rest/playlists/%d", playlist.ID), nil)
	s.apiAuthUser(req, otherUser)
	resp = s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)
}

func (s *ApiTestSuite) TestPlaylist_getPlaylist() {
	user := s.CreateUser()
	playlist := s.CreatePlaylist(user, "playlist", nil)
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/rest/playlists/%d", playlist.ID), nil)
	s.apiAuthUser(req, user)
	var resp Playlist
	s.request200json(req, &resp)
	s.assertPlaylist(playlist, &resp, 0)
}

func (s *ApiTestSuite) TestPlaylist_updatePlaylist_notFound() {
	user := s.CreateUser()

	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/rest/playlists/%d", 1), bytes.NewReader([]byte("{}")))
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)

	s.assertTokenVerifier()
	otherUser := s.CreateUser()
	playlist := s.CreatePlaylist(user, "playlist", nil)
	req, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/rest/playlists/%d", playlist.ID), bytes.NewReader([]byte("{}")))
	s.apiAuthUser(req, otherUser)
	resp = s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)
}

func (s *ApiTestSuite) TestPlaylist_updatePlaylist() {
	user := s.CreateUser()
	playlist := s.CreatePlaylist(user, "playlist", map[string]interface{}{
		"key1": "value1",
	})

	payload, err := json.Marshal(map[string]interface{}{
		"name": "edited playlist",
		"properties": map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	})
	s.Require().NoError(err)

	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/rest/playlists/%d", playlist.ID), bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	var resp Playlist
	s.request200json(req, &resp)

	s.Require().NoError(playlist.Reload(s.MyDB.DB))
	s.assertPlaylistInfo(playlist, &resp, 0)
}

func (s *ApiTestSuite) TestPlaylist_deletePlaylist_notFound() {
	user := s.CreateUser()

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/rest/playlists/%d", 1), nil)
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)

	s.assertTokenVerifier()
	otherUser := s.CreateUser()
	playlist := s.CreatePlaylist(user, "playlist", nil)
	req, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/rest/playlists/%d", playlist.ID), nil)
	s.apiAuthUser(req, otherUser)
	resp = s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)
}

func (s *ApiTestSuite) TestPlaylist_deletePlaylist() {
	user := s.CreateUser()
	playlist := s.CreatePlaylist(user, "playlist", nil)

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/rest/playlists/%d", playlist.ID), nil)
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Equal(http.StatusOK, resp.Code)

	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/rest/playlists/%d", playlist.ID), nil)
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)
}

func (s *ApiTestSuite) TestPlaylist_addItems_notFound() {
	user := s.CreateUser()

	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/rest/playlists/%d/add_items", 1), bytes.NewReader([]byte("{\"items\":[]}")))
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)

	s.assertTokenVerifier()
	otherUser := s.CreateUser()
	playlist := s.CreatePlaylist(user, "playlist", nil)
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/rest/playlists/%d/add_items", playlist.ID), bytes.NewReader([]byte("{\"items\":[]}")))
	s.apiAuthUser(req, otherUser)
	resp = s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)
}

func (s *ApiTestSuite) TestPlaylist_addItems_badRequest() {
	user := s.CreateUser()
	playlist := s.CreatePlaylist(user, "playlist", nil)

	// malformed UID
	payload, err := json.Marshal(AddPlaylistItemsRequest{Items: []PlaylistItemAddInfo{
		{Position: 1, ContentUnitUID: "malformed UID"},
		{Position: 2, ContentUnitUID: "12345678"},
	}})
	s.Require().NoError(err)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/rest/playlists/%d/add_items", playlist.ID), bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Equal(http.StatusBadRequest, resp.Code)

	// max playlist size
	s.AddPlaylistItems(playlist, MaxPlaylistSize-len(playlist.R.PlaylistItems))
	payload, err = json.Marshal(AddPlaylistItemsRequest{Items: []PlaylistItemAddInfo{
		{Position: 2, ContentUnitUID: "12345678"},
	}})
	s.Require().NoError(err)
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/rest/playlists/%d/add_items", playlist.ID), bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Equal(http.StatusBadRequest, resp.Code)
}

func (s *ApiTestSuite) TestPlaylist_addItems() {
	user := s.CreateUser()
	playlist := s.CreatePlaylist(user, "playlist", nil)

	payload, err := json.Marshal(AddPlaylistItemsRequest{Items: []PlaylistItemAddInfo{
		{Position: len(playlist.R.PlaylistItems) + 1, ContentUnitUID: "12345678"},
		{Position: len(playlist.R.PlaylistItems) + 2, ContentUnitUID: "87654321"},
	}})
	s.Require().NoError(err)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/rest/playlists/%d/add_items", playlist.ID), bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	var resp Playlist
	s.request200json(req, &resp)

	s.Require().NoError(playlist.L.LoadPlaylistItems(s.MyDB.DB, true, playlist, qm.OrderBy("position")))
	s.assertPlaylist(playlist, &resp, 0)
}

func (s *ApiTestSuite) TestPlaylist_updateItems_notFound() {
	user := s.CreateUser()

	// non existing playlist
	updateRequest := UpdatePlaylistItemsRequest{
		Items: []PlaylistItemUpdateInfo{
			{
				ID: 1,
				PlaylistItemAddInfo: PlaylistItemAddInfo{
					Position:       100,
					ContentUnitUID: "12345678",
				},
			},
		},
	}

	payload, err := json.Marshal(updateRequest)
	s.Require().NoError(err)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/rest/playlists/%d/update_items", 1), bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)

	// owner mismatch
	playlist := s.CreatePlaylist(user, "playlist", nil)
	updateRequest.Items[0].ID = playlist.R.PlaylistItems[0].ID
	payload, err = json.Marshal(updateRequest)
	s.Require().NoError(err)
	otherUser := s.CreateUser()
	req, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/rest/playlists/%d/update_items", playlist.ID), bytes.NewReader(payload))
	s.assertTokenVerifier()
	s.apiAuthUser(req, otherUser)
	resp = s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)

	// non existing item
	updateRequest.Items[0].ID = 8888
	payload, err = json.Marshal(updateRequest)
	s.Require().NoError(err)
	req, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/rest/playlists/%d/update_items", playlist.ID), bytes.NewReader(payload))
	s.assertTokenVerifier()
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)

	// parent playlist mismatch
	updateRequest.Items[0].ID = playlist.R.PlaylistItems[0].ID
	payload, err = json.Marshal(updateRequest)
	s.Require().NoError(err)
	playlist2 := s.CreatePlaylist(otherUser, "playlist", nil)
	req, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/rest/playlists/%d/update_items", playlist2.ID), bytes.NewReader(payload))
	s.apiAuthUser(req, otherUser)
	resp = s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)
}

func (s *ApiTestSuite) TestPlaylist_updateItems_badRequest() {
	user := s.CreateUser()
	playlist := s.CreatePlaylist(user, "playlist", nil)

	// malformed UID
	payload, err := json.Marshal(UpdatePlaylistItemsRequest{
		Items: []PlaylistItemUpdateInfo{
			{
				ID: playlist.R.PlaylistItems[0].ID,
				PlaylistItemAddInfo: PlaylistItemAddInfo{
					Position:       100,
					ContentUnitUID: "malformed UID",
				},
			},
		},
	})
	s.Require().NoError(err)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/rest/playlists/%d/update_items", playlist.ID), bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Equal(http.StatusBadRequest, resp.Code)
}

func (s *ApiTestSuite) TestPlaylist_updateItems() {
	user := s.CreateUser()
	playlist := s.CreatePlaylist(user, "playlist", nil)
	s.AddPlaylistItems(playlist, 1) // ensure at least 2 items

	updateRequest := UpdatePlaylistItemsRequest{
		Items: []PlaylistItemUpdateInfo{
			{
				ID: playlist.R.PlaylistItems[0].ID,
				PlaylistItemAddInfo: PlaylistItemAddInfo{
					Position:       100,
					ContentUnitUID: playlist.R.PlaylistItems[0].ContentUnitUID,
				},
			},
			{
				ID: playlist.R.PlaylistItems[1].ID,
				PlaylistItemAddInfo: PlaylistItemAddInfo{
					Position:       playlist.R.PlaylistItems[1].Position,
					ContentUnitUID: "87654321",
				},
			},
		},
	}
	payload, err := json.Marshal(updateRequest)
	s.Require().NoError(err)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/rest/playlists/%d/update_items", playlist.ID), bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	var resp Playlist
	s.request200json(req, &resp)

	s.Require().NoError(playlist.L.LoadPlaylistItems(s.MyDB.DB, true, playlist, qm.OrderBy("position")))
	s.assertPlaylist(playlist, &resp, 0)
}

func (s *ApiTestSuite) TestPlaylist_removeItems_notFound() {
	user := s.CreateUser()

	// non existing playlist
	removeRequest := RemovePlaylistItemsRequest{
		IDs: []int64{},
	}
	payload, err := json.Marshal(removeRequest)
	s.Require().NoError(err)
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/rest/playlists/%d/remove_items", 1), bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)

	// owner mismatch
	playlist := s.CreatePlaylist(user, "playlist", nil)
	removeRequest.IDs = []int64{playlist.R.PlaylistItems[0].ID}
	payload, err = json.Marshal(removeRequest)
	s.Require().NoError(err)
	otherUser := s.CreateUser()
	req, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/rest/playlists/%d/remove_items", playlist.ID), bytes.NewReader(payload))
	s.assertTokenVerifier()
	s.apiAuthUser(req, otherUser)
	resp = s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)

	// non existing item
	removeRequest.IDs = []int64{8888}
	payload, err = json.Marshal(removeRequest)
	s.Require().NoError(err)
	req, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/rest/playlists/%d/remove_items", playlist.ID), bytes.NewReader(payload))
	s.assertTokenVerifier()
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)

	// parent playlist mismatch
	removeRequest.IDs = []int64{playlist.R.PlaylistItems[0].ID}
	payload, err = json.Marshal(removeRequest)
	s.Require().NoError(err)
	playlist2 := s.CreatePlaylist(otherUser, "playlist", nil)
	req, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/rest/playlists/%d/remove_items", playlist2.ID), bytes.NewReader(payload))
	s.apiAuthUser(req, otherUser)
	resp = s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)
}

func (s *ApiTestSuite) TestPlaylist_removeItems() {
	user := s.CreateUser()
	playlist := s.CreatePlaylist(user, "playlist", nil)
	s.AddPlaylistItems(playlist, 1) // ensure at least 2 items

	removeRequest := RemovePlaylistItemsRequest{
		IDs: []int64{playlist.R.PlaylistItems[0].ID},
	}
	payload, err := json.Marshal(removeRequest)
	s.Require().NoError(err)
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/rest/playlists/%d/remove_items", playlist.ID), bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	var resp Playlist
	s.request200json(req, &resp)

	s.Require().NoError(playlist.L.LoadPlaylistItems(s.MyDB.DB, true, playlist, qm.OrderBy("position")))
	s.assertPlaylist(playlist, &resp, 0)
}

func (s *ApiTestSuite) assertPlaylist(expected *models.Playlist, actual *Playlist, idx int) {
	s.assertPlaylistInfo(expected, actual, idx)
	s.assertPlaylistItems(expected, actual, idx)
}

func (s *ApiTestSuite) assertPlaylistInfo(expected *models.Playlist, actual *Playlist, idx int) {
	s.Equal(expected.ID, actual.ID, "ID [%d]", idx)
	s.Equal(expected.UserID, actual.UserID, "UserID [%d]", idx)
	s.Equal(expected.UID, actual.UID, "UID [%d]", idx)
	s.Equal(expected.Name.String, actual.Name, "Name [%d]", idx)
	s.Equal(expected.Public, actual.Public, "Public [%d]", idx)

	if expected.Properties.Valid {
		var props map[string]interface{}
		s.Require().NoError(expected.Properties.Unmarshal(&props))
		s.Equal(props, actual.Properties, "Properties [%d]", idx)
	}
}

func (s *ApiTestSuite) assertPlaylistItems(expected *models.Playlist, actual *Playlist, idx int) {
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
