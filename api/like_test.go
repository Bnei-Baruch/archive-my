package api

import (
	"encoding/json"
	"math/rand"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"archive-my/models"
	"archive-my/pkg/testutil"
	"archive-my/pkg/utils"
)

func (s *RestTestSuite) TestLikes_noLikes() {
	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
	s.Require().Nil(err)

	s.app.getLikes(c, s.tx)
	var resp likesResponse
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(0, resp.Total, "empty total")
	s.Empty(resp.Likes, "empty data")
}

func (s *RestTestSuite) TestLikes_simpleGet() {
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

func (s *RestTestSuite) TestLikes_diffAccounts() {
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

func (s *RestTestSuite) TestLikes_paginate() {
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

func (s *RestTestSuite) TestLikes_add() {
	uids := UIDsRequest{UIDs: []string{utils.GenerateUID(8), utils.GenerateUID(8)}}
	cAdd, _, err := testutil.PrepareContext(uids)
	s.Require().Nil(err)
	respAdd, err := s.app.addLikes(cAdd, s.tx)
	s.NoError(err)
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

func (s *RestTestSuite) TestLikes_remove() {
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

func (s *RestTestSuite) TestLikes_removeOtherAcc() {
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
func (s *RestTestSuite) createDummyLike(n int64) []*models.Like {
	likes := make([]*models.Like, n)
	for i, _ := range likes {
		likes[i] = &models.Like{
			ID:             int64(i + 1),
			AccountID:      s.kcId,
			ContentUnitUID: utils.GenerateUID(8),
		}
		s.NoError(likes[i].Insert(s.tx, boil.Infer()))
	}
	return likes
}

func (s *RestTestSuite) assertEqualLikes(l *models.Like, x *models.Like, idx int) {
	s.Equal(l.ID, x.ID, "like.ID [%d]", idx)
	s.Equal(l.AccountID, x.AccountID, "like.AccountID [%d]", idx)
	s.Equal(l.ContentUnitUID, x.ContentUnitUID, "like.ContentUnitUID [%d]", idx)
}
