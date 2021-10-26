package api

//
//import (
//	"encoding/json"
//	"math/rand"
//
//	"github.com/volatiletech/null/v8"
//	"github.com/volatiletech/sqlboiler/v4/boil"
//
//	"github.com/Bnei-Baruch/archive-my/models"
//	"github.com/Bnei-Baruch/archive-my/pkg/testutil"
//	"github.com/Bnei-Baruch/archive-my/pkg/utils"
//)
//
//func (s *ApiTestSuite) TestHistory_noHistory() {
//	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
//	s.Require().NoError(err)
//
//	s.app.getLikes(c, s.mydb.DB)
//	var resp HistoryResponse
//	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
//	s.EqualValues(0, resp.Total, "empty total")
//	s.Empty(resp.Items, "empty data")
//}
//
//func (s *ApiTestSuite) TestHistory_simpleGet() {
//	items := s.createDummyHistory(10)
//	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10, OrderBy: "id"})
//	s.Require().NoError(err)
//	var resp HistoryResponse
//	s.app.getHistory(c, s.mydb.DB)
//	s.NoError(json.Unmarshal(w.Body.Bytes(), &resp))
//	s.EqualValues(10, resp.Total, "total")
//	for i, x := range resp.Items {
//		s.assertEqualHistory(items[i], x, i)
//	}
//}
//
//func (s *ApiTestSuite) TestHistory_diffAccounts() {
//	items := s.createDummyHistory(10)
//	items[1].AccountID = "new_account_id"
//	_, err := items[1].Update(s.mydb.DB, boil.Whitelist("account_id"))
//	s.NoError(err)
//	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
//	s.NoError(err)
//	var resp HistoryResponse
//	s.app.getHistory(c, s.mydb.DB)
//	s.NoError(json.Unmarshal(w.Body.Bytes(), &resp))
//	s.EqualValues(9, resp.Total, "total")
//	for _, l := range resp.Items {
//		s.Equal(s.kcId, l.AccountID)
//	}
//}
//
//func (s *ApiTestSuite) TestHistory_paginate() {
//	items := s.createDummyHistory(20)
//	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 2, PageSize: 5, OrderBy: "id"})
//	s.Require().NoError(err)
//	var resp HistoryResponse
//	s.Require().NoError(err)
//	s.app.getHistory(c, s.mydb.DB)
//	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
//	s.Equal(int64(20), resp.Total, "total")
//	s.Equal(5, len(resp.Items))
//	for i, x := range resp.Items {
//		s.assertEqualHistory(items[i+5], x, i+5)
//	}
//}
//
//func (s *ApiTestSuite) TestHistory_remove() {
//	items := s.createDummyHistory(20)
//	itemR1 := items[rand.Intn(10-1)+1]
//	itemR2 := items[rand.Intn(20-11)+11]
//	ids := &IDsFilter{IDs: []int64{itemR1.ID, itemR2.ID}}
//	cDel, _, err := testutil.PrepareContext(ids)
//	s.Require().NoError(err)
//	s.app.deleteHistory(cDel, s.mydb.DB)
//
//	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 20})
//	var resp HistoryResponse
//	s.app.getHistory(c, s.mydb.DB)
//	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
//	s.EqualValues(18, resp.Total, "total")
//	s.NotContains(resp.Items, itemR1)
//	s.NotContains(resp.Items, itemR2)
//}
//
//func (s *ApiTestSuite) TestHistory_removeOtherAcc() {
//	items := s.createDummyHistory(10)
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
//	s.app.deleteHistory(cDel, s.mydb.DB)
//	count, err := models.Histories().Count(s.mydb.DB)
//	s.NoError(err)
//	s.Equal(int64(10), count)
//}
//
////help functions
//func (s *ApiTestSuite) createDummyHistory(n int64) []*models.History {
//	items := make([]*models.History, n)
//	for i, _ := range items {
//		items[i] = &models.History{
//			ID:             int64(i + 1),
//			AccountID:      s.kcId,
//			ChronicleID:    utils.GenerateUID(27),
//			ContentUnitUID: null.String{String: utils.GenerateUID(8), Valid: true},
//		}
//		s.NoError(items[i].Insert(s.mydb.DB, boil.Infer()))
//	}
//	return items
//}
//
//func (s *ApiTestSuite) assertEqualHistory(l *models.History, x *models.History, idx int) {
//	s.Equal(l.ID, x.ID, "history.ID [%d]", idx)
//	s.Equal(l.AccountID, x.AccountID, "history.AccountID [%d]", idx)
//	s.Equal(l.ContentUnitUID, x.ContentUnitUID, "history.UnitUID [%d]", idx)
//}
