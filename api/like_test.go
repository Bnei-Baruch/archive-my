package api

//
//import (
//	"encoding/json"
//	"math/rand"
//
//	"github.com/volatiletech/sqlboiler/v4/boil"
//
//	"github.com/Bnei-Baruch/archive-my/models"
//	"github.com/Bnei-Baruch/archive-my/pkg/testutil"
//	"github.com/Bnei-Baruch/archive-my/pkg/utils"
//)
//
//func (s *ApiTestSuite) TestLikes_noLikes() {
//	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
//	s.Require().NoError(err)
//
//	s.app.getLikes(c, s.mydb.DB)
//	var resp ReactionsResponse
//	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
//	s.EqualValues(0, resp.Total, "empty total")
//	s.Empty(resp.items, "empty data")
//}
//
//func (s *ApiTestSuite) TestLikes_simpleGet() {
//	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10, OrderBy: "id"})
//	s.Require().NoError(err)
//
//	var resp ReactionsResponse
//	items := s.createDummyLike(10)
//	s.app.getLikes(c, s.mydb.DB)
//	s.NoError(json.Unmarshal(w.Body.Bytes(), &resp))
//	s.EqualValues(10, resp.Total, "total")
//	for i, x := range resp.items {
//		s.assertEqualLikes(items[i], x, i)
//	}
//}
//
//func (s *ApiTestSuite) TestLikes_diffAccounts() {
//	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
//	s.Require().NoError(err)
//
//	items := s.createDummyLike(10)
//
//	var resp ReactionsResponse
//	items[1].AccountID = "new_account_id"
//	_, err = items[1].Update(s.mydb.DB, boil.Whitelist("account_id"))
//	s.Nil(err)
//	s.app.getLikes(c, s.mydb.DB)
//	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
//	s.EqualValues(9, resp.Total, "total")
//	for _, l := range resp.items {
//		s.Equal(s.kcId, l.AccountID)
//	}
//}
//
//func (s *ApiTestSuite) TestLikes_paginate() {
//	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 2, PageSize: 5, OrderBy: "id"})
//	s.Require().NoError(err)
//
//	items := s.createDummyLike(20)
//
//	var resp ReactionsResponse
//	s.Require().NoError(err)
//	s.app.getLikes(c, s.mydb.DB)
//	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
//	s.Equal(int64(20), resp.Total, "total")
//	s.Equal(5, len(resp.items))
//	for i, x := range resp.items {
//		s.assertEqualLikes(items[i+5], x, i+5)
//	}
//}
//
//func (s *ApiTestSuite) TestLikes_add() {
//	uids := UIDsFilter{UIDs: []string{utils.GenerateUID(8), utils.GenerateUID(8)}}
//	cAdd, _, err := testutil.PrepareContext(uids)
//	s.Require().NoError(err)
//	respAdd, err := s.app.addLikes(cAdd, s.mydb.DB)
//	//s.Require().IsType(&HttpError{}, err)
//	s.Nil(err)
//	s.Equal(len(uids.UIDs), len(respAdd))
//	for _, a := range respAdd {
//		s.Contains(uids.UIDs, a.ContentUnitUID)
//	}
//
//	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
//	s.Require().NoError(err)
//	var resp ReactionsResponse
//	s.Require().NoError(err)
//	s.app.getLikes(c, s.mydb.DB)
//	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
//	s.EqualValues(int64(2), resp.Total, "total")
//
//	s.Equal(len(uids.UIDs), len(resp.items))
//	for _, a := range resp.items {
//		s.Contains(uids.UIDs, a.ContentUnitUID)
//	}
//}
//
//func (s *ApiTestSuite) TestLikes_remove() {
//	items := s.createDummyLike(20)
//	itemR1 := items[rand.Intn(10-1)+1]
//	itemR2 := items[rand.Intn(20-11)+11]
//	ids := &IDsFilter{IDs: []int64{itemR1.ID, itemR2.ID}}
//	cDel, _, err := testutil.PrepareContext(ids)
//	s.Require().NoError(err)
//	s.app.removeLikes(cDel, s.mydb.DB)
//
//	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 20})
//	var resp ReactionsResponse
//	s.app.getLikes(c, s.mydb.DB)
//	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
//	s.EqualValues(18, resp.Total, "total")
//	s.NotContains(resp.items, itemR1)
//	s.NotContains(resp.items, itemR2)
//}
//
//func (s *ApiTestSuite) TestLikes_removeOtherAcc() {
//	items := s.createDummyLike(10)
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
//	s.app.removeLikes(cDel, s.mydb.DB)
//	count, err := models.Likes().Count(s.mydb.DB)
//	s.NoError(err)
//	s.Equal(int64(10), count)
//}
//
////help functions
//func (s *ApiTestSuite) createDummyLike(n int64) []*models.Like {
//	likes := make([]*models.Like, n)
//	for i, _ := range likes {
//		likes[i] = &models.Like{
//			ID:             int64(i + 1),
//			AccountID:      s.kcId,
//			ContentUnitUID: utils.GenerateUID(8),
//		}
//		s.NoError(likes[i].Insert(s.mydb.DB, boil.Infer()))
//	}
//	return likes
//}
//
//func (s *ApiTestSuite) assertEqualLikes(l *models.Like, x *models.Like, idx int) {
//	s.Equal(l.ID, x.ID, "like.ID [%d]", idx)
//	s.Equal(l.AccountID, x.AccountID, "like.AccountID [%d]", idx)
//	s.Equal(l.ContentUnitUID, x.ContentUnitUID, "like.ContentUnitUID [%d]", idx)
//}
