package api

import (
	"encoding/json"
	"math/rand"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"archive-my/consts"
	"archive-my/models"
	"archive-my/pkg/testutil"
	"archive-my/pkg/utils"
)

func (s *RestTestSuite) TestSubscribe_noLikes() {
	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
	s.Require().Nil(err)

	s.app.getLikes(c, s.tx)
	var resp likesResponse
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(0, resp.Total, "empty total")
	s.Empty(resp.Likes, "empty data")
}

func (s *RestTestSuite) TestSubscribe_simpleGet() {
	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
	s.Require().Nil(err)

	var resp likesResponse
	items := s.createDummyLike(10)
	s.app.getLikes(c, s.tx)
	s.NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(10, resp.Total, "total")
	for i, x := range resp.Likes {
		s.assertEqualLikes(items[i], x, i)
	}
}

func (s *RestTestSuite) TestSubscribe_diffAccounts() {
	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
	s.Require().Nil(err)

	items := s.createDummyLike(10)

	var resp likesResponse
	items[1].AccountID = "new_account_id"
	_, err = items[1].Update(s.tx, boil.Whitelist("account_id"))
	s.Nil(err)
	s.app.getLikes(c, s.tx)
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(9, resp.Total, "total")
	for _, l := range resp.Likes {
		s.Equal(s.kcId, l.AccountID)
	}
}

func (s *RestTestSuite) TestSubscribe_paginate() {
	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 2, PageSize: 5})
	s.Require().Nil(err)

	items := s.createDummyLike(20)

	var resp likesResponse
	s.Require().Nil(err)
	s.app.getLikes(c, s.tx)
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(int64(20), resp.Total, "total")
	s.Equal(5, len(resp.Likes))
	for i, x := range resp.Likes {
		s.assertEqualLikes(items[i+5], x, i+5)
	}
}

func (s *RestTestSuite) TestSubscribe_add() {
	uids := UIDsRequest{UIDs: []string{utils.GenerateUID(8), utils.GenerateUID(8)}}
	cAdd, _, err := testutil.PrepareContext(uids)
	s.Require().Nil(err)
	respAdd, err := s.app.addLikes(cAdd, s.tx)
	s.Nil(err)
	s.Equal(len(uids.UIDs), len(respAdd))
	for _, a := range respAdd {
		s.Contains(uids.UIDs, a.ContentUnitUID)
	}

	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
	s.Require().Nil(err)
	var resp likesResponse
	s.Require().Nil(err)
	s.app.getLikes(c, s.tx)
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(int64(2), resp.Total, "total")

	s.Equal(len(uids.UIDs), len(resp.Likes))
	for _, a := range resp.Likes {
		s.Contains(uids.UIDs, a.ContentUnitUID)
	}
}

func (s *RestTestSuite) TestSubscribe_remove() {
	items := s.createDummyLike(20)
	itemR1 := items[rand.Intn(10-1)+1]
	itemR2 := items[rand.Intn(20-11)+11]
	ids := &IDsRequest{IDs: []int64{itemR1.ID, itemR2.ID}}
	cDel, _, err := testutil.PrepareContext(ids)
	s.Require().Nil(err)
	s.app.removeLikes(cDel, s.tx)

	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 20})
	var resp likesResponse
	s.app.getLikes(c, s.tx)
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(18, resp.Total, "total")
	s.NotContains(resp.Likes, itemR1)
	s.NotContains(resp.Likes, itemR2)
}

func (s *RestTestSuite) TestSubscribe_removeOtherAcc() {
	items := s.createDummyLike(10)

	item := items[rand.Intn(10-1)+1]
	item.AccountID = "new_account_id"
	c, err := item.Update(s.tx, boil.Infer())
	s.NoError(err)
	s.Equal(c, int64(1))

	ids := &IDsRequest{IDs: []int64{item.ID}}
	cDel, _, err := testutil.PrepareContext(ids)
	s.Require().Nil(err)
	s.app.removeLikes(cDel, s.tx)
	count, err := models.Likes().Count(s.tx)
	s.NoError(err)
	s.Equal(int64(10), count)
}

//help functions

func (s *RestTestSuite) createDummySubscriptions(n int64, name string) []*models.Subscription {
	subs := make([]*models.Subscription, n)

	for i, sub := range subs {
		sub = &models.Subscription{
			AccountID: testutil.KEYCKLOAK_ID,
		}
		if i%2 == 0 {
			sub.CollectionUID = null.String{String: utils.GenerateUID(8), Valid: true}
		} else {
			ct := consts.CT_SUBSCRIBE_BY_TYPE[rand.Int()%len(consts.CT_SUBSCRIBE_BY_TYPE)]
			sub.ContentType = null.String{String: ct, Valid: true}
		}
		s.Nil(sub.Insert(s.tx, boil.Infer()))
		units := make([]*models.PlaylistItem, rand.Int())
		for i, u := range units {
			u = &models.PlaylistItem{
				PlaylistID:     sub.ID,
				Position:       i,
				ContentUnitUID: utils.GenerateUID(8),
			}
			s.Nil(u.Insert(s.tx, boil.Infer()))
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
