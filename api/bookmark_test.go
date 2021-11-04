package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Bnei-Baruch/archive-my/databases/mydb/models"
)

//Bookmark tests
func (s *ApiTestSuite) TestBookmark_getBookmarks() {
	user := s.CreateUser()

	// no Bookmarks whatsoever
	req, _ := http.NewRequest(http.MethodGet, "/rest/bookmarks", nil)
	s.apiAuthUser(req, user)
	var resp GetBookmarksResponse
	s.request200json(req, &resp)

	s.EqualValues(0, resp.Total, "total")
	s.Empty(resp.Items, "items empty")

	// with Bookmarks
	bookmarks := make([]*models.Bookmark, 5)
	for i := range bookmarks {
		bookmarks[i] = s.CreateBookmark(user, fmt.Sprintf("Bookmark-%d", i), "", nil)
	}

	req, _ = http.NewRequest(http.MethodGet, "/rest/bookmarks?order_by=id", nil)
	s.apiAuthUser(req, user)
	s.request200json(req, &resp)

	s.EqualValues(len(bookmarks), resp.Total, "total")
	s.Require().Len(resp.Items, len(bookmarks), "items length")
	for i, x := range resp.Items {
		s.assertBookmark(bookmarks[i], x, i)
	}

	// other users see no Bookmarks
	s.assertTokenVerifier()
	otherUser := s.CreateUser()
	req, _ = http.NewRequest(http.MethodGet, "/rest/bookmarks", nil)
	s.apiAuthUser(req, otherUser)
	s.request200json(req, &resp)
	s.EqualValues(0, resp.Total, "total")
	s.Empty(resp.Items, "items empty")
}

func (s *ApiTestSuite) TestBookmark_createBookmark_badRequest() {
	user := s.CreateUser()

	// bad properties json
	payload, err := json.Marshal(map[string]interface{}{
		"name":        "test bookmark",
		"source_uid":  "12345678",
		"source_type": "TEST",
		"folder_ids":  "[1, 2]",
		"data":        "malformed json {}",
	})
	s.Require().NoError(err)
	req, _ := http.NewRequest(http.MethodPost, "/rest/bookmarks", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Require().Equal(http.StatusBadRequest, resp.Code)

	// too long name
	payload, err = json.Marshal(map[string]interface{}{
		"name": strings.Repeat("*", 257),
	})
	s.Require().NoError(err)
	req, _ = http.NewRequest(http.MethodPost, "/rest/bookmarks", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Require().Equal(http.StatusBadRequest, resp.Code)

	// no required source uid
	payload, err = json.Marshal(map[string]interface{}{
		"name":        "test bookmark",
		"source_type": "TEST",
	})
	s.Require().NoError(err)
	req, _ = http.NewRequest(http.MethodPost, "/rest/bookmarks", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Require().Equal(http.StatusBadRequest, resp.Code)

	// no required source content type
	payload, err = json.Marshal(map[string]interface{}{
		"name":       "test bookmark",
		"source_uid": "12345678",
	})
	s.Require().NoError(err)
	req, _ = http.NewRequest(http.MethodPost, "/rest/bookmarks", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Require().Equal(http.StatusBadRequest, resp.Code)
}

func (s *ApiTestSuite) TestBookmark_createBookmark() {
	user := s.CreateUser()
	f1 := s.CreateFolder(user, "test bookmark folder 1")
	f2 := s.CreateFolder(user, "test bookmark folder 2")
	bName := "test bookmark"

	var resp Bookmark
	payload, err := json.Marshal(map[string]interface{}{
		"name":        bName,
		"source_uid":  "12345678",
		"source_type": "TEST",
		"folder_ids":  []int64{f1.ID, f2.ID},
		"data": map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	})
	s.Require().NoError(err)
	req, _ := http.NewRequest(http.MethodPost, "/rest/bookmarks", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	s.request200json(req, &resp)

	s.NotZero(resp.ID, "ID")
	s.Equal(resp.Name, bName, "Name")
	s.Len(resp.Data, 2, "props count")
	s.Equal(resp.Data["key1"], "value1", "prop 1")
	s.Equal(resp.Data["key2"], "value2", "prop 2")
	s.Len(resp.FolderIds, 2, "folders count")
	s.Equal(resp.FolderIds[0], f1.ID, "folder 1")
	s.Equal(resp.FolderIds[1], f2.ID, "folder 2")
}

func (s *ApiTestSuite) TestBookmark_updateBookmark() {
	user := s.CreateUser()
	data := map[string]interface{}{
		"key1": "value1",
	}
	bookmark := s.CreateBookmark(user, "bookmark", "", data)

	payload, err := json.Marshal(map[string]interface{}{
		"name": "edited bookmark",
		"data": map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	})
	s.Require().NoError(err)

	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/rest/bookmarks/%d", bookmark.ID), bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	var resp Bookmark
	s.request200json(req, &resp)

	s.Require().NoError(bookmark.Reload(s.MyDB.DB))
	s.assertBookmark(bookmark, &resp, 0)
}

func (s *ApiTestSuite) TestBookmark_deleteBookmark() {
	user := s.CreateUser()
	bookmark := s.CreateBookmark(user, "bookmark", "", nil)
	bbfs := make([]*models.BookmarkFolder, 5)
	for i, _ := range bbfs {
		f := s.CreateFolder(user, fmt.Sprintf("test bookmark folder %d", i))
		bbfs[i] = &models.BookmarkFolder{
			FolderID: f.ID,
		}
	}
	s.NoError(bookmark.AddBookmarkFolders(s.MyDB.DB, true, bbfs...))

	count, err := models.BookmarkFolders(models.BookmarkFolderWhere.BookmarkID.EQ(bookmark.ID)).Count(s.MyDB.DB)
	s.NoError(err)
	s.EqualValues(5, count)

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/rest/bookmarks/%d", bookmark.ID), nil)
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Equal(http.StatusOK, resp.Code)

	count, err = models.BookmarkFolders(models.BookmarkFolderWhere.BookmarkID.EQ(bookmark.ID)).Count(s.MyDB.DB)
	s.NoError(err)
	s.EqualValues(0, count)

	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/rest/bookmarks/%d", bookmark.ID), nil)
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)
}

//Folder tests
/*
func (s *ApiTestSuite) TestFolder_getFolder() {
	user := s.CreateUser()

	// no Bookmarks Folder whatsoever
	req, _ := http.NewRequest(http.MethodGet, "/rest/folders", nil)
	s.apiAuthUser(req, user)
	var resp GetBookmarksResponse
	s.request200json(req, &resp)

	s.EqualValues(0, resp.Total, "total")
	s.Empty(resp.Items, "items empty")

	// with Bookmarks
	bookmarks := make([]*models.Bookmark, 5)
	for i := range bookmarks {
		bookmarks[i] = s.CreateBookmark(user, fmt.Sprintf("Bookmark-%d", i), "", nil)
	}

	req, _ = http.NewRequest(http.MethodGet, "/rest/bookmarks?order_by=id", nil)
	s.apiAuthUser(req, user)
	s.request200json(req, &resp)

	s.EqualValues(len(bookmarks), resp.Total, "total")
	s.Require().Len(resp.Items, len(bookmarks), "items length")
	for i, x := range resp.Items {
		s.assertBookmark(bookmarks[i], x, i)
	}

	// other users see no Bookmarks
	s.assertTokenVerifier()
	otherUser := s.CreateUser()
	req, _ = http.NewRequest(http.MethodGet, "/rest/bookmarks", nil)
	s.apiAuthUser(req, otherUser)
	s.request200json(req, &resp)
	s.EqualValues(0, resp.Total, "total")
	s.Empty(resp.Items, "items empty")
}
*/
/*
func (s *ApiTestSuite) TestBookmark_createBookmark_badRequest() {
	user := s.CreateUser()

	// bad properties json
	payload, err := json.Marshal(map[string]interface{}{
		"name":        "test bookmark",
		"source_uid":  "12345678",
		"source_type": "TEST",
		"folder_ids":  "[1, 2]",
		"data":        "malformed json {}",
	})
	s.Require().NoError(err)
	req, _ := http.NewRequest(http.MethodPost, "/rest/bookmarks", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Require().Equal(http.StatusBadRequest, resp.Code)

	// too long name
	payload, err = json.Marshal(map[string]interface{}{
		"name": strings.Repeat("*", 257),
	})
	s.Require().NoError(err)
	req, _ = http.NewRequest(http.MethodPost, "/rest/bookmarks", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Require().Equal(http.StatusBadRequest, resp.Code)

	// no required source uid
	payload, err = json.Marshal(map[string]interface{}{
		"name":        "test bookmark",
		"source_type": "TEST",
	})
	s.Require().NoError(err)
	req, _ = http.NewRequest(http.MethodPost, "/rest/bookmarks", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Require().Equal(http.StatusBadRequest, resp.Code)

	// no required source content type
	payload, err = json.Marshal(map[string]interface{}{
		"name":       "test bookmark",
		"source_uid": "12345678",
	})
	s.Require().NoError(err)
	req, _ = http.NewRequest(http.MethodPost, "/rest/bookmarks", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Require().Equal(http.StatusBadRequest, resp.Code)
}

func (s *ApiTestSuite) TestBookmark_createBookmark() {
	user := s.CreateUser()
	f1 := s.CreateFolder(user, "test bookmark folder 1")
	f2 := s.CreateFolder(user, "test bookmark folder 2")
	bName := "test bookmark"

	var resp Bookmark
	payload, err := json.Marshal(map[string]interface{}{
		"name":        bName,
		"source_uid":  "12345678",
		"source_type": "TEST",
		"folder_ids":  []int64{f1.ID, f2.ID},
		"data": map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	})
	s.Require().NoError(err)
	req, _ := http.NewRequest(http.MethodPost, "/rest/bookmarks", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	s.request200json(req, &resp)

	s.NotZero(resp.ID, "ID")
	s.Equal(resp.Name, bName, "Name")
	s.Len(resp.Data, 2, "props count")
	s.Equal(resp.Data["key1"], "value1", "prop 1")
	s.Equal(resp.Data["key2"], "value2", "prop 2")
	s.Len(resp.FolderIds, 2, "folders count")
	s.Equal(resp.FolderIds[0], f1.ID, "folder 1")
	s.Equal(resp.FolderIds[1], f2.ID, "folder 2")
}

func (s *ApiTestSuite) TestBookmark_updateBookmark() {
	user := s.CreateUser()
	data := map[string]interface{}{
		"key1": "value1",
	}
	bookmark := s.CreateBookmark(user, "bookmark", "", data)

	payload, err := json.Marshal(map[string]interface{}{
		"name": "edited bookmark",
		"data": map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	})
	s.Require().NoError(err)

	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/rest/bookmarks/%d", bookmark.ID), bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	var resp Bookmark
	s.request200json(req, &resp)

	s.Require().NoError(bookmark.Reload(s.MyDB.DB))
	s.assertBookmark(bookmark, &resp, 0)
}

func (s *ApiTestSuite) TestBookmark_deleteBookmark() {
	user := s.CreateUser()
	bookmark := s.CreateBookmark(user, "bookmark", "", nil)
	bbfs := make([]*models.BookmarkFolder, 5)
	for i, _ := range bbfs {
		f := s.CreateFolder(user, fmt.Sprintf("test bookmark folder %d", i))
		bbfs[i] = &models.BookmarkFolder{
			BookmarkFolderID: f.ID,
		}
	}
	s.NoError(bookmark.AddBookmarkFolders(s.MyDB.DB, true, bbfs...))

	count, err := models.BookmarkFolders(models.BookmarkFolderWhere.BookmarkID.EQ(bookmark.ID)).Count(s.MyDB.DB)
	s.NoError(err)
	s.EqualValues(5, count)

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/rest/bookmarks/%d", bookmark.ID), nil)
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Equal(http.StatusOK, resp.Code)

	count, err = models.BookmarkFolders(models.BookmarkFolderWhere.BookmarkID.EQ(bookmark.ID)).Count(s.MyDB.DB)
	s.NoError(err)
	s.EqualValues(0, count)

	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/rest/bookmarks/%d", bookmark.ID), nil)
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)
}
*/
func (s *ApiTestSuite) assertBookmark(expected *models.Bookmark, actual *Bookmark, idx int) {
	s.Equal(expected.ID, actual.ID, "ID [%d]", idx)
	s.Equal(expected.Name.String, actual.Name, "Name [%d]", idx)
	s.Equal(expected.SourceType, actual.SourceType, "SourceType [%d]", idx)
	s.Equal(expected.SourceUID, actual.SourceUID, "SourceUID [%d]", idx)

	if expected.Data.Valid {
		var data map[string]interface{}
		s.Require().NoError(expected.Data.Unmarshal(&data))
		s.Equal(data, actual.Data, "Data [%d]", idx)
	}
}
