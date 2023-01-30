package api

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/Bnei-Baruch/archive-my/databases/mydb/models"
)

func (s *ApiTestSuite) TestNotes_getNotes() {
	user := s.CreateUser()

	// no notes whatsoever
	req, _ := http.NewRequest(http.MethodGet, "/rest/notes", nil)
	s.apiAuthUser(req, user)
	var resp NotesResponse
	s.request200json(req, &resp)

	s.Empty(resp.Items, "items empty")

	// with notes
	notes := make([]*models.Note, 5)
	for i := range notes {
		notes[i] = s.CreateNote(user, "en")
	}

	req, _ = http.NewRequest(http.MethodGet, "/rest/notes", nil)
	s.apiAuthUser(req, user)
	s.request200json(req, &resp)

	s.Require().Len(resp.Items, len(notes), "items length")
	for i, x := range resp.Items {
		s.assertNotes(notes[i], x, i)
	}

	// other users see no notes
	s.assertTokenVerifier()
	otherUser := s.CreateUser()
	req, _ = http.NewRequest(http.MethodGet, "/rest/notes", nil)
	s.apiAuthUser(req, otherUser)
	s.request200json(req, &resp)
	s.Empty(resp.Items, "items empty")
}

func (s *ApiTestSuite) TestNotes_deleteNote() {
	user := s.CreateUser()

	note := s.CreateNote(user, "en")

	payload, err := json.Marshal(map[string]interface{}{
		"language":    note.Language,
		"subject_uid": note.SubjectUID,
	})
	s.NoError(err)

	req, _ := http.NewRequest(http.MethodDelete, "/rest/notes", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Equal(http.StatusOK, resp.Code)
}

func (s *ApiTestSuite) TestNotes_deleteNote_fail() {
	user := s.CreateUser()

	note := s.CreateNote(user, "en")

	//remove without kind
	payload, err := json.Marshal(map[string]interface{}{
		"language":    note.Language,
		"subject_uid": note.SubjectUID,
	})
	s.NoError(err)
	req, _ := http.NewRequest(http.MethodDelete, "/rest/notes", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Equal(http.StatusBadRequest, resp.Code)
}

//help functions

func (s *ApiTestSuite) assertNotes(expected *models.Note, actual *Note, idx int) {
	s.Equal(expected.SubjectUID, actual.SubjectUID, "SubjectUID [%d]", idx)
	s.Equal(expected.Language, actual.Language, "Language [%d]", idx)
	s.Equal(expected.Content, actual.Content, "Content [%d]", idx)
}
