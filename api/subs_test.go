package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/Bnei-Baruch/archive-my/databases/mydb/models"
	"github.com/Bnei-Baruch/archive-my/pkg/utils"
)

var subscribeByType = []string{"CLIPS", "FRIENDS_GATHERINGS", "MEALS", "WOMEN_LESSONS", "CLIP", "LECTURE", "LESSON_PART", "MEAL"}

func (s *ApiTestSuite) TestSubscribe_getSubscriptions() {
	user := s.CreateUser()

	// no subscriptions whatsoever
	req, _ := http.NewRequest(http.MethodGet, "/rest/subscriptions", nil)
	s.apiAuthUser(req, user)
	var resp SubscriptionsResponse
	s.request200json(req, &resp)

	s.EqualValues(0, resp.Total, "total")
	s.Empty(resp.Items, "items empty")

	// with subscriptions
	subs := make([]*models.Subscription, 10)
	j := 0
	for i := range subs {
		var ct string
		if i%2 == 0 && j < len(subscribeByType) {
			ct = subscribeByType[j]
			j++
		}
		subs[i] = s.CreateSubscription(user, ct)
	}

	req, _ = http.NewRequest(http.MethodGet, "/rest/subscriptions?order_by=id", nil)
	s.apiAuthUser(req, user)
	s.request200json(req, &resp)

	s.EqualValues(len(subs), resp.Total, "total")
	s.Require().Len(resp.Items, len(subs), "items length")
	for i, x := range resp.Items {
		s.assertSubscriptions(subs[i], x)
	}

	// other users see no subs
	s.assertTokenVerifier()
	otherUser := s.CreateUser()
	req, _ = http.NewRequest(http.MethodGet, "/rest/subscriptions", nil)
	s.apiAuthUser(req, otherUser)
	s.request200json(req, &resp)
	s.EqualValues(0, resp.Total, "total")
	s.Empty(resp.Items, "items empty")
}

func (s *ApiTestSuite) TestSubscribe_getSubscriptions_filtered() {
	user := s.CreateUser()

	subByCT := s.CreateSubscription(user, "CLIPS")
	subByUID := s.CreateSubscription(user, "")

	var resp SubscriptionsResponse
	wrongUID := fmt.Sprintf("%sXX", subByUID.CollectionUID.String[:6])
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/rest/subscriptions?collection_uid=%s", wrongUID), nil)
	s.apiAuthUser(req, user)
	s.request200json(req, &resp)
	s.Equal(int64(0), resp.Total, "not subscribed by collection")

	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/rest/subscriptions?collection_uid=%s", subByUID.CollectionUID.String), nil)
	s.apiAuthUser(req, user)
	s.request200json(req, &resp)
	s.Equal(int64(1), resp.Total, "subscribed by collection")
	s.assertSubscriptions(subByUID, resp.Items[0])

	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/rest/subscriptions?content_type=%s", "PROGRAMMS"), nil)
	s.apiAuthUser(req, user)
	s.request200json(req, &resp)
	s.Equal(int64(0), resp.Total, "not subscribed by CT")

	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/rest/subscriptions?content_type=%s", subByCT.ContentType.String), nil)
	s.apiAuthUser(req, user)
	s.request200json(req, &resp)
	s.Equal(int64(1), resp.Total, "subscribed by CT")
	s.assertSubscriptions(subByCT, resp.Items[0])

}

func (s *ApiTestSuite) TestSubscribe_subscribe_fail() {
	user := s.CreateUser()

	//wrong collection uid
	payload, err := json.Marshal(map[string]interface{}{
		"collection_uid": utils.GenerateUID(rand.Intn(10) + 8),
	})
	s.NoError(err, "json.Marshal")
	req, _ := http.NewRequest(http.MethodPost, "/rest/subscriptions", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Equal(http.StatusInternalServerError, resp.Code, "wrong collection uid")

	payload, err = json.Marshal(map[string]interface{}{
		"content_type": utils.GenerateName(rand.Intn(10) + 33),
	})
	s.NoError(err, "json.Marshal")
	req, _ = http.NewRequest(http.MethodPost, "/rest/subscriptions", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Equal(http.StatusInternalServerError, resp.Code, "wrong content type")

	payload, err = json.Marshal(map[string]interface{}{
		"content_type":   "",
		"collection_uid": "",
	})
	s.NoError(err, "json.Marshal")
	req, _ = http.NewRequest(http.MethodPost, "/rest/subscriptions", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Equal(http.StatusBadRequest, resp.Code, "no content type and no collection")

}

func (s *ApiTestSuite) TestSubscribe_subscribe() {
	user := s.CreateUser()

	//by collection
	cUID := utils.GenerateUID(8)
	cuUID := utils.GenerateUID(8)
	payload, err := json.Marshal(map[string]interface{}{
		"collection_uid":   cUID,
		"content_unit_uid": cuUID,
	})
	s.NoError(err, "json.Marshal")
	req, _ := http.NewRequest(http.MethodPost, "/rest/subscriptions", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	var resp Subscription
	s.request200json(req, &resp)

	s.NotZero(resp.ID, "by collection: ID")
	s.Equal(resp.CollectionUID.String, cUID, "by collection: collection UID")
	s.Empty(resp.ContentType, "by collection: ContentType")
	s.Equal(resp.ContentUnitUID.String, cuUID, "by collection: ContentUnitUID")

	// by content type
	cType := utils.GenerateName(rand.Intn(31) + 1)
	payload, err = json.Marshal(map[string]interface{}{
		"content_type":     cType,
		"content_unit_uid": cuUID,
	})
	s.Require().NoError(err)
	req, _ = http.NewRequest(http.MethodPost, "/rest/subscriptions", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	s.request200json(req, &resp)

	s.NotZero(resp.ID, "by content type: ID")
	s.Empty(resp.CollectionUID, cUID, "by content type: collection UID")
	s.Equal(resp.ContentType.String, cType, "by content type: ContentType")
	s.Equal(resp.ContentUnitUID.String, cuUID, "by content type: ContentUnitUID")
}

func (s *ApiTestSuite) TestSubscribe_unsubscribe() {
	user := s.CreateUser()
	subscription := s.CreateSubscription(user, "")

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/rest/subscriptions/%d", subscription.ID), nil)
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Equal(http.StatusOK, resp.Code)

	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/rest/subscriptions/%d", subscription.ID), nil)
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)
}

//help functions

func (s *ApiTestSuite) assertSubscriptions(expected *models.Subscription, actual *Subscription) {
	s.Equal(expected.ContentType, actual.ContentType)
	s.Equal(expected.CollectionUID, actual.CollectionUID)
	s.Equal(expected.ContentUnitUID, actual.ContentUnitUID)
}
