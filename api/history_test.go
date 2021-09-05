package api

import (
	"encoding/json"
	"math/rand"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"archive-my/models"
	"archive-my/pkg/testutil"
	"archive-my/pkg/utils"
)

func (s *RestTestSuite) TestHistory_noHistory() {
	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
	s.Require().Nil(err)

	s.app.getLikes(c, s.tx)
	var resp historyResponse
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(0, resp.Total, "empty total")
	s.Empty(resp.History, "empty data")
}

func (s *RestTestSuite) TestHistory_simpleGet() {
	items := s.createDummyHistory(10)
	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
	s.Require().Nil(err)
	var resp historyResponse
	s.app.getHistory(c, s.tx)
	s.NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(10, resp.Total, "total")
	for i, x := range resp.History {
		s.assertEqualHistory(items[i], x, i)
	}
}

func (s *RestTestSuite) TestHistory_diffAccounts() {
	items := s.createDummyHistory(10)
	items[1].AccountID = "new_account_id"
	_, err := items[1].Update(s.tx, boil.Whitelist("account_id"))
	s.NoError(err)
	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
	s.NoError(err)
	var resp historyResponse
	s.app.getHistory(c, s.tx)
	s.NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(9, resp.Total, "total")
	for _, l := range resp.History {
		s.Equal(s.kcId, l.AccountID)
	}
}

func (s *RestTestSuite) TestHistory_paginate() {
	items := s.createDummyHistory(20)
	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 2, PageSize: 5})
	s.Require().Nil(err)
	var resp historyResponse
	s.Require().Nil(err)
	s.app.getHistory(c, s.tx)
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(int64(20), resp.Total, "total")
	s.Equal(5, len(resp.History))
	for i, x := range resp.History {
		s.assertEqualHistory(items[i+5], x, i+5)
	}
}

func (s *RestTestSuite) TestHistory_remove() {
	items := s.createDummyHistory(20)
	itemR1 := items[rand.Intn(10-1)+1]
	itemR2 := items[rand.Intn(20-11)+11]
	ids := &IDsRequest{IDs: []int64{itemR1.ID, itemR2.ID}}
	cDel, _, err := testutil.PrepareContext(ids)
	s.Require().Nil(err)
	s.app.deleteHistory(cDel, s.tx)

	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 20})
	var resp historyResponse
	s.app.getHistory(c, s.tx)
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(18, resp.Total, "total")
	s.NotContains(resp.History, itemR1)
	s.NotContains(resp.History, itemR2)
}

func (s *RestTestSuite) TestHistory_removeOtherAcc() {
	items := s.createDummyHistory(10)

	item := items[rand.Intn(10-1)+1]
	item.AccountID = "new_account_id"
	c, err := item.Update(s.tx, boil.Infer())
	s.NoError(err)
	s.Equal(c, int64(1))

	ids := &IDsRequest{IDs: []int64{item.ID}}
	cDel, _, err := testutil.PrepareContext(ids)
	s.Require().Nil(err)
	s.app.deleteHistory(cDel, s.tx)
	count, err := models.Histories().Count(s.tx)
	s.NoError(err)
	s.Equal(int64(10), count)
}

//help functions
func (s *RestTestSuite) createDummyHistory(n int64) []*models.History {
	items := make([]*models.History, n)
	for i, _ := range items {
		items[i] = &models.History{
			ID:             int64(i + 1),
			AccountID:      s.kcId,
			ChronicleID:    utils.GenerateUID(36),
			ContentUnitUID: null.String{String: utils.GenerateUID(8), Valid: true},
		}
		s.Nil(items[i].Insert(s.tx, boil.Infer()))
	}
	return items
}

func (s *RestTestSuite) assertEqualHistory(l *models.History, x *models.History, idx int) {
	s.Equal(l.ID, x.ID, "like.ID [%d]", idx)
	s.Equal(l.AccountID, x.AccountID, "like.AccountID [%d]", idx)
	s.Equal(l.ContentUnitUID, x.ContentUnitUID, "like.UnitUID [%d]", idx)
}
