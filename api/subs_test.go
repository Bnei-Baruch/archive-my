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

func (s *RestTestSuite) TestSubscribe_noSubscriptions() {
	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
	s.Require().Nil(err)

	s.app.getSubscriptions(c, s.tx)
	var resp subscriptionsResponse
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(0, resp.Total, "empty total")
	s.Empty(resp.Subscriptions, "empty data")
}

func (s *RestTestSuite) TestSubscribe_simpleGet() {
	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
	s.Require().Nil(err)

	var resp subscriptionsResponse
	items := s.createDummySubscriptions(10)
	s.app.getSubscriptions(c, s.tx)
	s.NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(10, resp.Total, "total")
	for i, x := range resp.Subscriptions {
		s.assertEqualSubscriptions(items[i], x)
	}
}

func (s *RestTestSuite) TestSubscribe_diffAccounts() {
	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
	s.Require().Nil(err)

	items := s.createDummySubscriptions(10)

	var resp subscriptionsResponse
	items[1].AccountID = "new_account_id"
	_, err = items[1].Update(s.tx, boil.Whitelist("account_id"))
	s.Nil(err)
	s.app.getSubscriptions(c, s.tx)
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(9, resp.Total, "total")
	for _, l := range resp.Subscriptions {
		s.Equal(s.kcId, l.AccountID)
	}
}

func (s *RestTestSuite) TestSubscribe_paginate() {
	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 2, PageSize: 5})
	s.Require().Nil(err)

	items := s.createDummySubscriptions(20)

	var resp subscriptionsResponse
	s.Require().Nil(err)
	s.app.getSubscriptions(c, s.tx)
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(int64(20), resp.Total, "total")
	s.Equal(5, len(resp.Subscriptions))
	for i, x := range resp.Subscriptions {
		s.assertEqualSubscriptions(items[i+5], x)
	}
}

func (s *RestTestSuite) TestSubscribe_add() {
	ctSub := consts.CT_SUBSCRIBE_BY_TYPE[rand.Int()%len(consts.CT_SUBSCRIBE_BY_TYPE)]
	coSub := utils.GenerateUID(8)
	subs := subscribeRequest{
		Collections:    []string{coSub},
		ContentTypes:   []string{ctSub},
		ContentUnitUID: "",
	}
	cAdd, _, err := testutil.PrepareContext(subs)
	s.Require().Nil(err)
	respAdd, err := s.app.subscribe(cAdd, s.tx)
	s.Nil(err)
	s.Equal(2, len(respAdd))

	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 10})
	s.Require().Nil(err)
	var resp subscriptionsResponse
	s.Require().Nil(err)
	s.app.getSubscriptions(c, s.tx)
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(int64(2), resp.Total, "total")

	for _, sub := range resp.Subscriptions {
		if sub.CollectionUID.Valid {
			s.Equal(sub.CollectionUID.String, coSub)
			s.False(sub.ContentType.Valid)
		}
		if sub.ContentType.Valid {
			s.Equal(sub.ContentType.String, ctSub)
			s.False(sub.CollectionUID.Valid)
		}
	}
}

func (s *RestTestSuite) TestSubscribe_remove() {
	items := s.createDummySubscriptions(20)
	itemR1 := items[rand.Intn(10-1)+1]
	itemR2 := items[rand.Intn(20-11)+11]
	ids := &IDsRequest{IDs: []int64{itemR1.ID, itemR2.ID}}
	cDel, _, err := testutil.PrepareContext(ids)
	s.Require().Nil(err)
	err = s.app.unsubscribe(cDel, s.tx)
	s.Nil(err)

	c, w, err := testutil.PrepareContext(ListRequest{PageNumber: 1, PageSize: 20})
	var resp subscriptionsResponse
	s.app.getSubscriptions(c, s.tx)
	s.Nil(json.Unmarshal(w.Body.Bytes(), &resp))
	s.EqualValues(18, resp.Total, "total")
	s.NotContains(resp.Subscriptions, itemR1)
	s.NotContains(resp.Subscriptions, itemR2)
}

func (s *RestTestSuite) TestSubscribe_removeOtherAcc() {
	items := s.createDummySubscriptions(10)

	item := items[rand.Intn(10-1)+1]
	item.AccountID = "new_account_id"
	c, err := item.Update(s.tx, boil.Infer())
	s.NoError(err)
	s.Equal(c, int64(1))

	ids := &IDsRequest{IDs: []int64{item.ID}}
	cDel, _, err := testutil.PrepareContext(ids)
	s.Require().Nil(err)
	err = s.app.unsubscribe(cDel, s.tx)
	s.Require().Nil(err)
	count, err := models.Subscriptions().Count(s.tx)
	s.NoError(err)
	s.Equal(int64(10), count)
}

//help functions

func (s *RestTestSuite) createDummySubscriptions(n int) []*models.Subscription {
	subs := make([]*models.Subscription, n)

	for i, _ := range subs {
		subs[i] = &models.Subscription{
			AccountID: testutil.KEYCKLOAK_ID,
		}
		if i%2 == 0 {
			subs[i].CollectionUID = null.String{String: utils.GenerateUID(8), Valid: true}
		} else {
			ct := consts.CT_SUBSCRIBE_BY_TYPE[rand.Int()%len(consts.CT_SUBSCRIBE_BY_TYPE)]
			subs[i].ContentType = null.String{String: ct, Valid: true}
		}
		s.Nil(subs[i].Insert(s.tx, boil.Infer()))
	}
	return subs
}

func (s *RestTestSuite) assertEqualSubscriptions(l *models.Subscription, x *models.Subscription) {
	s.Equal(l.AccountID, x.AccountID)
	s.Equal(l.ContentType, x.ContentType)
	s.Equal(l.CollectionUID, x.CollectionUID)
	s.Equal(l.ContentUnitUID, x.ContentUnitUID)
}
