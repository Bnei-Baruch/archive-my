package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Bnei-Baruch/archive-my/pkg/utils"
	"math/rand"
	"net/http"
	"strings"

	"github.com/Bnei-Baruch/archive-my/databases/mydb/models"
)

//Label tests
func (s *ApiTestSuite) TestLabel_getPublicLabels() {

	// no Labels whatsoever
	req, _ := http.NewRequest(http.MethodGet, "/labels", nil)
	var resp GetLabelsResponse
	s.request200json(req, &resp)

	s.EqualValues(0, resp.Total, "total")
	s.Empty(resp.Items, "items empty")

	// with Labels

	labels := make([]*models.Label, rand.Intn(10))
	for i, _ := range labels {
		labels[i] = s.CreateLabel(s.CreateUser(), fmt.Sprintf("Label num %d", i), "SOURCE", "en", nil)
	}

	req, _ = http.NewRequest(http.MethodGet, "/labels?order_by=id", nil)
	s.request200json(req, &resp)

	s.EqualValues(len(labels), resp.Total, "total")
	s.Require().Len(resp.Items, len(labels), "items length")
	for i, x := range resp.Items {
		s.assertLabel(labels[i], x, i)
	}
}

func (s *ApiTestSuite) TestLabel_createLabel_badRequest() {
	user := s.CreateUser()

	// bad properties json
	payload, err := json.Marshal(map[string]interface{}{
		"name":         "test label",
		"subject_uid":  "12345678",
		"subject_type": "TEST",
		"language":     "en",
		"tag_uids":     "[1, 2]",
		"properties":   "malformed json {}",
	})
	s.Require().NoError(err)
	req, _ := http.NewRequest(http.MethodPost, "/rest/labels", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Require().Equal(http.StatusBadRequest, resp.Code)

	// too long name
	payload, err = json.Marshal(map[string]interface{}{
		"name":         strings.Repeat("*", 257),
		"subject_uid":  "12345678",
		"subject_type": "TEST",
		"language":     "en",
	})
	s.Require().NoError(err)
	req, _ = http.NewRequest(http.MethodPost, "/rest/labels", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Require().Equal(http.StatusBadRequest, resp.Code)

	// no required source uid
	payload, err = json.Marshal(map[string]interface{}{
		"name":         "test label",
		"subject_type": "TEST",
		"language":     "en",
	})
	s.Require().NoError(err)
	req, _ = http.NewRequest(http.MethodPost, "/rest/labels", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Require().Equal(http.StatusBadRequest, resp.Code)

	// no required source content type
	payload, err = json.Marshal(map[string]interface{}{
		"name":        "test bookmark",
		"subject_uid": "12345678",
		"language":    "en",
	})
	s.Require().NoError(err)
	req, _ = http.NewRequest(http.MethodPost, "/rest/labels", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Require().Equal(http.StatusBadRequest, resp.Code)

	// no required language
	payload, err = json.Marshal(map[string]interface{}{
		"name":         "test label",
		"subject_uid":  "12345678",
		"subject_type": "TEST",
	})
	s.Require().NoError(err)
	req, _ = http.NewRequest(http.MethodPost, "/rest/labels", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Require().Equal(http.StatusBadRequest, resp.Code)
}

func (s *ApiTestSuite) TestLabel_createLabel() {
	user := s.CreateUser()
	bName := "test label"
	tuids := []string{utils.GenerateUID(8), utils.GenerateUID(8)}

	var resp Label
	payload, err := json.Marshal(map[string]interface{}{
		"name":         bName,
		"subject_uid":  "12345678",
		"subject_type": "TEST",
		"language":     "en",
		"tag_uids":     tuids,
		"properties": map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	})
	s.Require().NoError(err)
	req, _ := http.NewRequest(http.MethodPost, "/rest/labels", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	s.request200json(req, &resp)

	s.NotZero(resp.ID, "ID")
	s.NotZero(resp.UID, "UID")
	s.Equal(resp.Name, bName, "Name")
	s.Len(resp.Properties, 2, "props count")
	s.Equal(resp.Properties["key1"], "value1", "prop 1")
	s.Equal(resp.Properties["key2"], "value2", "prop 2")
	s.Len(resp.TagUIds, 2, "tags count")
	s.Equal(resp.TagUIds[0], tuids[0], "tag 1")
	s.Equal(resp.TagUIds[1], tuids[1], "tag 2")
}

//help functions
func (s *ApiTestSuite) assertLabel(expected *models.Label, actual *Label, idx int) {
	s.Equal(expected.ID, actual.ID, "ID [%d]", idx)
	s.Equal(expected.UID, actual.UID, "UID [%d]", idx)
	s.Equal(expected.Name.String, actual.Name, "Name [%d]", idx)
	s.Equal(expected.SubjectType, actual.SubjectType, "SubjectType [%d]", idx)
	s.Equal(expected.SubjectUID, actual.SubjectUID, "SubjectUID [%d]", idx)
	s.Equal(expected.Accepted, actual.Accepted, "Accepted [%d]", idx)
	s.Equal(expected.Language, actual.Language, "Language [%d]", idx)

	if expected.Properties.Valid {
		var properties map[string]interface{}
		s.Require().NoError(expected.Properties.Unmarshal(&properties))
		s.Equal(properties, actual.Properties, "Properties [%d]", idx)
	}

	if expected.R != nil && expected.R.LabelTags != nil {
		tUIDs := make([]string, len(expected.R.LabelTags))
		for i, x := range expected.R.LabelTags {
			tUIDs[i] = x.TagUID
		}
		s.Equal(tUIDs, actual.TagUIds, "FolderIds [%d]", idx)
	}
}
