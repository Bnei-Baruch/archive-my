package api

import (
	"database/sql"
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"archive-my/models"
	"archive-my/pkg/testutil"
	"archive-my/pkg/utils"
)

type RestTestSuite struct {
	suite.Suite
	testutil.TestDBManager
	tx   *sql.Tx
	ctx  *gin.Context
	kcId string
	app  *App
}

func (s *RestTestSuite) SetupSuite() {
	s.Require().Nil(utils.InitConfig("", "../"))
	s.app = new(App)
	s.Require().Nil(s.InitTestDB())
	s.app.SetDB(s.DB)

	verifier := testutil.OIDCTokenVerifier{}
	s.app.SetVerifier(&verifier)
	s.kcId = testutil.KEYCKLOAK_ID

	s.ctx = &gin.Context{}
	s.ctx.Set("KC_ID", s.kcId)
}

func (s *RestTestSuite) TearDownSuite() {
	s.DestroyTestDB()
	//s.Require().Nil(s.DestroyTestDB())
}

func (s *RestTestSuite) SetupTest() {
	var err error
	s.tx, err = s.DB.Begin()
	s.Require().Nil(err)
}

func (s *RestTestSuite) TearDownTest() {
	err := s.tx.Rollback()
	s.Require().Nil(err)
}

/* TESTS */
func (s *RestTestSuite) TestLikes() {
	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
	s.Require().Nil(err)

	s.app.handleGetLikes(c)
	var resp likesResponse
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(0, resp.Total, "empty total")
	s.Empty(resp.Likes, "empty data")

	items := s.createDummyLike(10)
	s.app.handleGetLikes(c)
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(10, resp.Total, "total")
	for i, x := range resp.Likes {
		s.assertEqualLikes(items[i], x, i)
	}

	items[1].AccountID = "new_account_id"
	_, err = items[1].Update(s.tx, boil.Whitelist("account_id"))
	s.Nil(err)
	s.app.handleGetLikes(c)
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(9, resp.Total, "total")

	cPart, wPart, err := testutil.PrepareContext(ListRequest{PageNumber: 2, PageSize: 4})
	s.Require().Nil(err)
	s.app.handleGetLikes(cPart)
	s.Nil(json.Unmarshal(wPart.Body.Bytes(), &resp))
	s.EqualValues(4, resp.Total, "total")
	for i, x := range resp.Likes {
		s.assertEqualLikes(items[i+5], x, i+5)
	}

	like := &models.Like{
		ID:             11,
		AccountID:      s.kcId,
		ContentUnitUID: utils.GenerateUID(8),
	}
	items = append(items, like)
	cAdd, wAdd, err := testutil.PrepareContext(like)
	s.Require().Nil(err)
	s.app.handleAddLikes(cAdd)

	var respAdd gin.H
	s.Nil(json.Unmarshal(wAdd.Body.Bytes(), &respAdd))
	s.Len(resp, 1)
	s.assertEqualLikes(like, respAdd["0"].(*models.Like), 11)

	s.app.handleRemoveLikes(c)
	var respRem gin.H
	s.Nil(json.Unmarshal(wAdd.Body.Bytes(), &respRem))

	s.Len(respRem, 1)
	s.assertEqualLikes(like, respRem["0"].(*models.Like), 11)

	s.app.handleGetLikes(c)
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(9, resp.Total, "total")

}

func (s *RestTestSuite) TestPlaylist() {
	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
	s.Require().Nil(err)

	s.app.handleGetPlaylists(c)
	var resp playlistsResponse
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(0, resp.Total, "empty total")
	s.Empty(resp.Playlist, "empty data")

	items := s.createDummyPlaylists(10, "It's play list")
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.app.handleGetPlaylists(c)
	s.EqualValues(10, resp.Total, "total")
	for i, x := range resp.Playlist {
		s.assertEqualPlaylist(items[i], x, i)
	}

	items[1].AccountID = "new_account_id"
	_, err = items[1].Update(s.tx, boil.Whitelist("account_id"))
	s.Nil(err)
	s.app.handleGetPlaylists(c)
	s.EqualValues(9, resp.Total, "total")

	cPart, wPart, err := testutil.PrepareContext(ListRequest{PageNumber: 2, PageSize: 4})
	s.Nil(err)
	s.app.handleGetPlaylists(cPart)
	s.Nil(json.Unmarshal(wPart.Body.Bytes(), &resp))
	s.EqualValues(4, resp.Total, "total")
	for i, x := range resp.Playlist {
		s.assertEqualPlaylist(items[i+5], x, i+5)
	}

	pl := &models.Playlist{
		ID:        int64(11),
		AccountID: s.kcId,
		Name:      null.String{String: "playlist 11", Valid: false},
		Public:    null.Bool{Bool: false, Valid: true},
	}
	cAdd, wAdd, err := testutil.PrepareContext(pl)
	s.Nil(err)
	s.app.handleAddToPlaylist(cAdd)
	var respAdd []*models.Playlist
	s.Nil(json.Unmarshal(wAdd.Body.Bytes(), &respAdd))
	s.Len(resp, 1)
	s.assertEqualPlaylist(items[11], respAdd[0], 11)

	cDel, wDel, err := testutil.PrepareContext([]int64{respAdd[0].ID})
	s.Nil(err)
	s.app.handleDeletePlaylist(cDel)
	var respDel []*models.Playlist
	s.Nil(json.Unmarshal(wDel.Body.Bytes(), &respDel))
	s.Len(respDel, 1)
	s.assertEqualPlaylist(respAdd[0], respDel[0], 11)

	s.app.handleGetPlaylists(c)
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(9, resp.Total, "total")
}

func (s *RestTestSuite) TestHistory() {
	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
	s.Require().Nil(err)

	s.app.handleGetHistory(c)
	var resp historyResponse
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(0, resp.Total, "empty total")
	s.Empty(resp.History, "empty data")

	items := s.createDummyHistory(10)
	s.app.handleGetHistory(c)
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(10, resp.Total, "total")
	for i, x := range resp.History {
		s.assertEqualHistory(items[i], x, i)
	}

	items[1].AccountID = "new_account_id"
	_, err = items[1].Update(s.tx, boil.Whitelist("account_id"))
	s.Nil(err)
	s.app.handleGetHistory(c)
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(9, resp.Total, "total")

	cPart, wPart, err := testutil.PrepareContext(ListRequest{PageNumber: 2, PageSize: 4})
	s.Require().Nil(err)
	s.app.handleGetHistory(cPart)
	s.Nil(json.Unmarshal(wPart.Body.Bytes(), &resp))
	s.EqualValues(4, resp.Total, "total part")
	for i, x := range resp.History {
		s.assertEqualHistory(items[i+5], x, i+5)
	}

	cDel, wDel, err := testutil.PrepareContext([]int64{items[0].ID})
	s.app.handleDeleteHistory(cDel)
	var respDel gin.H
	s.Nil(json.Unmarshal(wDel.Body.Bytes(), &respDel))
	s.Len(respDel, 1)
	s.Equal(respDel["0"].(int64), items[0].ID, "deleted right history")

	s.app.handleGetHistory(c)
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(9, resp.Total, "total deleted")
}

/*
func (s *RestTestSuite) TestSubscriptions() {

	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10, StartIndex: 1, StopIndex: 10})
	s.Require().Nil(err)

	s.app.handleGetSubscriptions(c)
	s.app.handleGetLikes(c)
	var resp playlistsResponse
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(0, resp.Total, "empty total")
	s.Empty(resp.Playlist, "empty data")

	playlists := s.createDummyPlaylists(10, "playlist name")
	s.app.handleGetPlaylists(c)
	s.EqualValues(10, resp.Total, "total")
	for i, x := range resp.Playlist {
		s.assertEqualPlaylist(playlists[i], x, i)
	}

	playlists[1].AccountID = "new_account_id"
	s.NotNil(playlists[1].Insert(s.tx, boil.Infer()))
	s.app.handleGetSubscriptions(c)
	s.EqualValues(9, resp.Total, "total")

	cPart, wPart, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10, StartIndex: 6, StopIndex: 10})
	s.Require().Nil(err)
	s.app.handleGetPlaylists(cPart)
	s.Nil(json.Unmarshal(wPart.Body.Bytes(), &resp))
	s.EqualValues(10, resp.Total, "total")
	for i, x := range resp.Playlist {
		s.assertEqualPlaylist(playlists[i+5], x, i+5)
	}

	pl := &models.Playlist{
		ID:        11,
		AccountID: s.kcId,
		Name:      null.String{String: "Additional playlist", Valid: true},
	}
	playlists = append(playlists, pl)
	cAdd, wAdd, err := testutil.PrepareContext(pl)
	s.Require().Nil(err)
	s.app.handleCreatePlaylist(cAdd)
	s.Nil(json.Unmarshal(wAdd.Body.Bytes(), &resp))
	s.Len(resp, 1)
	var respAdd *models.Playlist
	s.assertEqualPlaylist(pl, respAdd, 11)


	cDel, wDel, err := testutil.PrepareContext([]int64{pl.ID})
	s.Nil(err)
	s.app.handleDeletePlaylist(cDel)
	s.Nil(err)
	s.Len(dLikes, 1)
	s.assertEqualPlaylist(like, dLikes[0], 11)

	resp, err = api.handleGetLikes(s.tx, s.kcId, req)
	s.Require().Nil(err)
	s.EqualValues(9, resp.Total, "total")

}
*/
/* HELPERS */

func (s *RestTestSuite) createDummyLike(n int64) []*models.Like {
	likes := make([]*models.Like, n)
	for i, l := range likes {
		likes[i] = &models.Like{
			ID:             int64(i),
			AccountID:      s.kcId,
			ContentUnitUID: utils.GenerateUID(8),
		}
		s.Nil(l.Insert(s.tx, boil.Infer()))
	}
	return likes
}

func (s *RestTestSuite) assertEqualLikes(l *models.Like, x *models.Like, idx int) {
	s.Equal(l.ID, x.ID, "like.ID [%d]", idx)
	s.Equal(l.AccountID, x.AccountID, "like.AccountID [%d]", idx)
	s.Equal(l.ContentUnitUID, x.ContentUnitUID, "like.ContentUnitUID [%d]", idx)
}

func (s *RestTestSuite) createDummyPlaylists(n int64, name string) []*models.Playlist {
	playlists := make([]*models.Playlist, n)

	for i, pl := range playlists {
		pl = &models.Playlist{
			ID:         int64(i),
			AccountID:  s.kcId,
			Name:       null.String{String: name, Valid: false},
			Parameters: null.JSON{},
			Public:     null.Bool{Bool: false, Valid: true},
		}
		s.Nil(pl.Insert(s.tx, boil.Infer()))
		units := make([]*models.PlaylistItem, rand.Int())
		for i, u := range units {
			units[i] = &models.PlaylistItem{
				PlaylistID:     pl.ID,
				Position:       null.Int{Int: i, Valid: true},
				ContentUnitUID: utils.GenerateUID(8),
			}
			s.Nil(u.Insert(s.tx, boil.Infer()))
		}
	}
	return playlists
}

func (s *RestTestSuite) assertEqualPlaylist(l *models.Playlist, x *models.Playlist, idx int) {
	s.Equal(l.ID, x.ID, "playlist.ID [%d]", idx)
	s.Equal(l.AccountID, x.AccountID, "playlist.AccountID [%d]", idx)
	s.Equal(l.Name, x.Name, "playlist.Name [%d]", idx)
	s.Equal(l.Public, x.Public, "playlist.Public [%d]", idx)

	s.Len(x.R.PlaylistItems, len(l.R.PlaylistItems), "playlist.PlaylistItems len [%d]", idx)
	for i, u := range l.R.PlaylistItems {
		s.Equal(u, x.R.PlaylistItems[i], "playlist.PlaylistItem (item index %d) [%d]", idx, i)
	}
}

func (s *RestTestSuite) createDummyHistory(n int64) []*models.History {
	items := make([]*models.History, n)
	for i, l := range items {
		items[i] = &models.History{
			ID:             int64(i),
			AccountID:      s.kcId,
			ChronicleID:    utils.GenerateUID(36),
			ContentUnitUID: null.String{String: utils.GenerateUID(8), Valid: true},
		}
		s.Nil(l.Insert(s.tx, boil.Infer()))
	}
	return items
}

func (s *RestTestSuite) assertEqualHistory(l *models.History, x *models.History, idx int) {
	s.Equal(l.ID, x.ID, "like.ID [%d]", idx)
	s.Equal(l.AccountID, x.AccountID, "like.AccountID [%d]", idx)
	s.Equal(l.ContentUnitUID, x.ContentUnitUID, "like.UnitUID [%d]", idx)
}

/*
func (s *RestTestSuite) createDummySubscriptions(n int64, name string) []*models.Subscription {
	subs := make([]*models.Subscription, n)

	for i, s := range subs {
		s = &models.Subscription{
			AccountID: testutil.KEYCKLOAK_ID,
		}
		if i%2 == 0 {
			s.CollectionID = null.String{String: utils.GenerateUID(8), Valid: true}
		} else {
			t := collections_CT[rand.Int()%len(collections_CT)]
			s.ContentUnitType = null.Int64{Int64: mdb.CONTENT_TYPE_REGISTRY.ByName[t].ID, Valid: true}
		}
		s.Nil(s.Insert(s.tx))
		units := make([]*models.PlaylistItem, rand.Int())
		for i, u := range units {
			u = &models.PlaylistItem{
				PlaylistID:     s.ID,
				Position:       null.Int{Int: i, Valid: true},
				ContentUnitUID: utils.GenerateUID(8),
			}
			s.Nil(u.Insert(s.tx))
		}
	}
	return subs
}

func (s *RestTestSuite) assertEqualSubscriptions(l *models.Playlist, x *models.Playlist, idx int) {
	s.Equal(l.ID, x.ID, "playlist.ID [%d]", idx)
	s.Equal(l.AccountID, x.AccountID, "playlist.AccountID [%d]", idx)
	s.Equal(l.Name, x.Name, "playlist.Name [%d]", idx)
	s.Equal(l.Public, x.Public, "playlist.Public [%d]", idx)

	s.Len(x.R.PlaylistItems, len(l.R.PlaylistItems), "playlist.PlaylistItems len [%d]", idx)
	for i, u := range l.R.PlaylistItems {
		s.Equal(u, x.R.PlaylistItems[i], "playlist.PlaylistItem (item index %d) [%d]", idx, i)
	}
}
*/
func TestApiTestSuite(t *testing.T) {
	suite.Run(t, new(RestTestSuite))
}
