package api

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/Bnei-Baruch/archive-my/models"
	"github.com/Bnei-Baruch/archive-my/pkg/testutil"
	"github.com/Bnei-Baruch/archive-my/pkg/utils"
)

func (s *RestTestSuite) TestPlaylist_noPlaylist() {
	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
	s.Require().Nil(err)
	s.app.getPlaylists(c, s.tx)
	var resp playlistsResponse
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(int64(0), resp.Total, "empty total")
	s.Empty(resp.Playlists, "empty data")
}

func (s *RestTestSuite) TestPlaylist_simpleGet() {
	items := s.createDummyPlaylists(10, "first playlist")

	var resp playlistsResponse
	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
	s.Require().NoError(err)
	s.app.getPlaylists(c, s.tx)
	s.NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(10, resp.Total, "total")
	for i, x := range resp.Playlists {
		s.assertEqualPlaylist(items[i], x.Playlist, i)
	}
}

func (s *RestTestSuite) TestPlaylist_diffAccounts() {
	items := s.createDummyLike(10)

	items[1].AccountID = "new_account_id"
	_, err := items[1].Update(s.tx, boil.Whitelist("account_id"))
	s.NoError(err)

	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
	s.Require().Nil(err)
	s.app.getLikes(c, s.tx)
	var resp playlistsResponse
	s.NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(9, resp.Total, "total")
	for _, l := range resp.Playlists {
		s.Equal(s.kcId, l.AccountID)
	}
}

func (s *RestTestSuite) TestPlaylist_paginate() {
	items := s.createDummyPlaylists(10, "my playlist")

	var resp playlistsResponse
	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 2, PageSize: 5})
	s.Require().NoError(err)
	s.app.getPlaylists(c, s.tx)
	s.NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(10, resp.Total)
	s.Equal(5, len(resp.Playlists))
	for i, x := range resp.Playlists {
		s.assertEqualPlaylist(items[i+5], x.Playlist, i+5)
	}
}

func (s *RestTestSuite) TestPlaylist_add() {
	newPl := models.Playlist{ID: 1, Name: null.String{String: "new playlist", Valid: true}}
	cAdd, _, err := testutil.PrepareContext(newPl)
	s.NoError(err)
	respAdd, err := s.app.createPlaylist(cAdd, s.tx)
	s.NoError(err)
	s.Equal(newPl.ID, respAdd[0].ID)

	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
	s.NoError(err)
	var resp playlistsResponse
	s.app.getPlaylists(c, s.tx)
	s.NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(1, resp.Total, "total")
	s.Equal(respAdd[0].ID, resp.Playlists[0].ID)
	s.Equal(respAdd[0].Name, resp.Playlists[0].Name)
	s.Equal(s.kcId, resp.Playlists[0].AccountID)
}

func (s *RestTestSuite) TestPlaylist_remove() {
	items := s.createDummyPlaylists(10, "playlist for remove")
	itemR1 := items[rand.Intn(10-1)+1]
	ids := &IDsRequest{IDs: []int64{itemR1.ID}}
	cDel, _, err := testutil.PrepareContext(ids)
	s.NoError(err)
	s.Nil(s.app.deletePlaylist(cDel, s.tx))

	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 20})
	var resp playlistsResponse
	s.app.getPlaylists(c, s.tx)
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(9, resp.Total, "total")
	s.NotContains(resp.Playlists, itemR1)
}

func (s *RestTestSuite) TestPlaylist_removeOtherAcc() {
	items := s.createDummyPlaylists(10, "playlists for remove")

	item := items[rand.Intn(10-1)+1]
	item.AccountID = "new_account_id"
	c, err := item.Update(s.tx, boil.Infer())
	s.NoError(err)
	s.Equal(c, int64(1))

	ids := &IDsRequest{IDs: []int64{item.ID}}
	cDel, _, err := testutil.PrepareContext(ids)
	s.Require().NoError(err)
	s.Nil(s.app.deletePlaylist(cDel, s.tx))
	count, err := models.Playlists().Count(s.tx)
	s.NoError(err)
	s.Equal(int64(10), count)
}

//help functions
func (s *RestTestSuite) createDummyPlaylists(n int64, name string) []*models.Playlist {
	playlists := make([]*models.Playlist, n)

	for i, _ := range playlists {
		playlists[i] = &models.Playlist{
			AccountID: s.kcId,
			Name:      null.String{String: fmt.Sprintf("%s - %d", name, i), Valid: true},
		}
		s.NoError(playlists[i].Insert(s.tx, boil.Infer()))
		units := make([]*models.PlaylistItem, rand.Intn(20)+1)
		for j, _ := range units {
			units[j] = &models.PlaylistItem{
				PlaylistID:     playlists[i].ID,
				Position:       j,
				ContentUnitUID: utils.GenerateUID(8),
			}
			s.NoError(units[j].Insert(s.tx, boil.Infer()))
		}
	}
	return playlists
}

func (s *RestTestSuite) assertEqualPlaylist(l *models.Playlist, x *models.Playlist, idx int) {
	s.Equal(l.ID, x.ID, "playlist.ID [%d]", idx)
	s.Equal(l.AccountID, x.AccountID, "playlist.AccountID [%d]", idx)
	s.Equal(l.Name, x.Name, "playlist.Name [%d]", idx)
	s.Equal(l.Public, x.Public, "playlist.Public [%d]", idx)

	if x.R == nil && l.R == nil {
		return
	}

	s.Len(x.R.PlaylistItems, len(l.R.PlaylistItems), "playlist.PlaylistItems len [%d]", idx)
	for i, u := range l.R.PlaylistItems {
		s.Equal(u, x.R.PlaylistItems[i], "playlist.PlaylistItem (item index %d) [%d]", idx, i)
	}
}
