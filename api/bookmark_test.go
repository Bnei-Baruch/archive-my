package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
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
		"name":         "test bookmark",
		"subject_uid":  "12345678",
		"subject_type": "TEST",
		"folder_ids":   "[1, 2]",
		"properties":   "malformed json {}",
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
		"name":         "test bookmark",
		"subject_type": "TEST",
	})
	s.Require().NoError(err)
	req, _ = http.NewRequest(http.MethodPost, "/rest/bookmarks", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Require().Equal(http.StatusBadRequest, resp.Code)

	// no required source content type
	payload, err = json.Marshal(map[string]interface{}{
		"name":        "test bookmark",
		"subject_uid": "12345678",
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
		"name":         bName,
		"subject_uid":  "12345678",
		"subject_type": "TEST",
		"folder_ids":   []int64{f1.ID, f2.ID},
		"properties": map[string]interface{}{
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
	s.Len(resp.Properties, 2, "props count")
	s.Equal(resp.Properties["key1"], "value1", "prop 1")
	s.Equal(resp.Properties["key2"], "value2", "prop 2")
	s.Len(resp.FolderIds, 2, "folders count")
	s.Equal(resp.FolderIds[0], f1.ID, "folder 1")
	s.Equal(resp.FolderIds[1], f2.ID, "folder 2")
}

func (s *ApiTestSuite) TestBookmark_updateBookmark() {
	user := s.CreateUser()
	properties := map[string]interface{}{
		"key1": "value1",
	}
	bookmark := s.CreateBookmark(user, "bookmark", "", properties)

	payload, err := json.Marshal(map[string]interface{}{
		"name": "edited bookmark",
		"properties": map[string]interface{}{
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

func (s *ApiTestSuite) TestBookmark_updateBookmarkFolders() {
	user := s.CreateUser()
	bookmark := s.CreateBookmark(user, "bookmark", "", nil)

	flen := 10 + rand.Intn(3)
	folders := make([]*models.Folder, flen)
	forUpdate := make([]int64, 0)
	addOnCreate := make([]int64, 0)
	notAddOnCreate := make([]int64, 0)
	for i, _ := range folders {
		folders[i] = s.CreateFolder(user, fmt.Sprintf("bookmark folder %d", i))
		if rand.Int()%2 == 0 {
			addOnCreate = append(addOnCreate, folders[i].ID)
			if rand.Intn(3)%3 == 0 {
				forUpdate = append(forUpdate, folders[i].ID)
			}
		} else {
			notAddOnCreate = append(notAddOnCreate, folders[i].ID)
			if rand.Intn(3)%3 == 0 {
				forUpdate = append(forUpdate, folders[i].ID)
			}
		}
	}

	bfs := make([]*models.BookmarkFolder, len(addOnCreate))
	folderIds := make([]int64, len(bfs))
	for i, id := range addOnCreate {
		bfs[i] = &models.BookmarkFolder{
			FolderID: id,
		}
		folderIds[i] = id
	}

	_ = bookmark.AddBookmarkFolders(s.MyDB.DB, true, bfs...)

	count, err := models.BookmarkFolders(
		models.BookmarkFolderWhere.BookmarkID.EQ(bookmark.ID),
	).Count(s.MyDB.DB)
	s.NoError(err)
	s.EqualValues(len(addOnCreate), count)

	payload, err := json.Marshal(map[string]interface{}{
		"folder_ids": addOnCreate,
	})
	s.Require().NoError(err)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/rest/bookmarks/%d", bookmark.ID), bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	var resp Bookmark
	s.request200json(req, &resp)
	s.EqualValues(addOnCreate, resp.FolderIds)

	payload, err = json.Marshal(map[string]interface{}{
		"folder_ids": forUpdate,
	})
	s.Require().NoError(err)
	req, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/rest/bookmarks/%d", bookmark.ID), bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	s.request200json(req, &resp)
	sort.Slice(resp.FolderIds, func(i, j int) bool {
		return resp.FolderIds[i] < resp.FolderIds[j]
	})
	s.EqualValues(forUpdate, resp.FolderIds)

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

func (s *ApiTestSuite) TestFolder_getFolder() {
	user := s.CreateUser()

	// no Bookmarks Folder whatsoever
	req, _ := http.NewRequest(http.MethodGet, "/rest/folders", nil)
	s.apiAuthUser(req, user)
	var resp GetFoldersResponse
	s.request200json(req, &resp)

	s.EqualValues(0, resp.Total, "total")
	s.Empty(resp.Items, "items empty")

	// with Bookmarks
	folders := make([]*models.Folder, 5)
	for i := range folders {
		folders[i] = s.CreateFolder(user, fmt.Sprintf("Folder-%d", i))
	}

	req, _ = http.NewRequest(http.MethodGet, "/rest/folders?order_by=id", nil)
	s.apiAuthUser(req, user)
	s.request200json(req, &resp)

	s.EqualValues(len(folders), resp.Total, "total")
	s.Require().Len(resp.Items, len(folders), "items length")
	for i, x := range resp.Items {
		s.assertFolders(folders[i], x, i)
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

func (s *ApiTestSuite) TestBookmark_createFolder_badRequest() {
	user := s.CreateUser()

	// too long name
	payload, err := json.Marshal(map[string]interface{}{
		"name": strings.Repeat("*", 257),
	})
	s.Require().NoError(err)
	req, _ := http.NewRequest(http.MethodPost, "/rest/bookmarks", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Require().Equal(http.StatusBadRequest, resp.Code)

}

func (s *ApiTestSuite) TestBookmark_createFolder() {
	user := s.CreateUser()
	name := "test bookmark folder"

	var resp Folder
	payload, err := json.Marshal(map[string]interface{}{"name": name})
	s.Require().NoError(err)
	req, _ := http.NewRequest(http.MethodPost, "/rest/folders", bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	s.request200json(req, &resp)

	s.NotZero(resp.ID, "ID")
	s.Equal(resp.Name, name, "Name")
}

func (s *ApiTestSuite) TestBookmark_updateFolder() {
	user := s.CreateUser()
	folder := s.CreateFolder(user, "bookmark folder")

	payload, err := json.Marshal(map[string]interface{}{"name": "edited bookmark folder"})
	s.Require().NoError(err)

	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/rest/folders/%d", folder.ID), bytes.NewReader(payload))
	s.apiAuthUser(req, user)
	var resp Folder
	s.request200json(req, &resp)

	s.Require().NoError(folder.Reload(s.MyDB.DB))
	s.assertFolders(folder, &resp, 0)
}

func (s *ApiTestSuite) TestBookmark_deleteFolder() {
	user := s.CreateUser()
	folder := s.CreateFolder(user, "bookmark folder")
	bbfs := make([]*models.BookmarkFolder, 5)
	for i, _ := range bbfs {
		b := s.CreateBookmark(user, fmt.Sprintf("test bookmark folder %d", i), "", nil)
		bbfs[i] = &models.BookmarkFolder{
			BookmarkID: b.ID,
		}
	}
	s.NoError(folder.AddBookmarkFolders(s.MyDB.DB, true, bbfs...))

	count, err := models.BookmarkFolders(models.BookmarkFolderWhere.FolderID.EQ(folder.ID)).Count(s.MyDB.DB)
	s.NoError(err)
	s.EqualValues(5, count)

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/rest/folders/%d", folder.ID), nil)
	s.apiAuthUser(req, user)
	resp := s.request(req)
	s.Equal(http.StatusOK, resp.Code)

	count, err = models.BookmarkFolders(models.BookmarkFolderWhere.FolderID.EQ(folder.ID)).Count(s.MyDB.DB)
	s.NoError(err)
	s.EqualValues(0, count)

	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/rest/folders/%d", folder.ID), nil)
	s.apiAuthUser(req, user)
	resp = s.request(req)
	s.Equal(http.StatusNotFound, resp.Code)
}

//help functions
func (s *ApiTestSuite) assertBookmark(expected *models.Bookmark, actual *Bookmark, idx int) {
	s.Equal(expected.ID, actual.ID, "ID [%d]", idx)
	s.Equal(expected.Name.String, actual.Name, "Name [%d]", idx)
	s.Equal(expected.SubjectType, actual.SubjectType, "SubjectType [%d]", idx)
	s.Equal(expected.SubjectUID, actual.SubjectUID, "SubjectUID [%d]", idx)

	if expected.Properties.Valid {
		var properties map[string]interface{}
		s.Require().NoError(expected.Properties.Unmarshal(&properties))
		s.Equal(properties, actual.Properties, "Properties [%d]", idx)
	}

	if expected.R != nil && expected.R.BookmarkFolders != nil {
		fIDs := make([]int64, len(expected.R.BookmarkFolders))
		for i, x := range expected.R.BookmarkFolders {
			fIDs[i] = x.FolderID
		}
		s.Equal(fIDs, actual.FolderIds, "FolderIds [%d]", idx)
	}
}

func (s *ApiTestSuite) assertFolders(expected *models.Folder, actual *Folder, idx int) {
	s.Equal(expected.ID, actual.ID, "ID [%d]", idx)
	s.Equal(expected.Name.String, actual.Name, "Name [%d]", idx)
}
