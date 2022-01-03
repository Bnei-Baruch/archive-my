package domain

import (
	"encoding/json"
	"github.com/stretchr/testify/suite"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"math/rand"

	"github.com/Bnei-Baruch/archive-my/databases/mydb/models"
	"github.com/Bnei-Baruch/archive-my/pkg/testutil"
	"github.com/Bnei-Baruch/archive-my/pkg/utils"
)

type ModelsSuite struct {
	suite.Suite
	MyDB *testutil.TestMyDBManager
}

func (s *ModelsSuite) CreateUser() *models.User {
	user := &models.User{
		AccountsID: utils.GenerateName(36),
		Email:      null.StringFrom("user@example.com"),
		FirstName:  null.StringFrom("first"),
		LastName:   null.StringFrom("last"),
	}
	s.Require().NoError(user.Insert(s.MyDB.DB, boil.Infer()))
	return user
}

func (s *ModelsSuite) CreatePlaylist(user *models.User, name string, props map[string]interface{}) *models.Playlist {
	uid, err := GetFreeUID(s.MyDB.DB, new(PlaylistUIDChecker))
	s.Require().NoError(err, "get free UID")

	playlist := &models.Playlist{
		UserID: user.ID,
		UID:    uid,
		Name:   null.StringFrom(name),
	}
	if len(props) > 0 {
		propsJson, err := json.Marshal(props)
		s.Require().NoError(err)
		playlist.Properties = null.JSONFrom(propsJson)
	}

	s.Require().NoError(playlist.Insert(s.MyDB.DB, boil.Infer()))

	s.AddPlaylistItems(playlist, rand.Intn(20)+1)

	return playlist
}

func (s *ModelsSuite) AddPlaylistItems(playlist *models.Playlist, n int) {
	items := make([]*models.PlaylistItem, n)
	for i := range items {
		items[i] = &models.PlaylistItem{
			Position:       i,
			ContentUnitUID: utils.GenerateUID(8),
		}
	}
	s.Require().NoError(playlist.AddPlaylistItems(s.MyDB.DB, true, items...))
}

func (s *ModelsSuite) CreateHistory(user *models.User) *models.History {
	history := &models.History{
		UserID:         user.ID,
		ContentUnitUID: null.String{String: utils.GenerateUID(8), Valid: true},
	}
	s.Require().NoError(history.Insert(s.MyDB.DB, boil.Infer()))
	return history
}

func (s *ModelsSuite) CreateSubscription(user *models.User, ct string) *models.Subscription {
	subscription := &models.Subscription{
		UserID: user.ID,
	}
	if ct == "" {
		subscription.CollectionUID = null.StringFrom(utils.GenerateUID(8))
	} else {
		subscription.ContentType = null.StringFrom(ct)
	}
	s.Require().NoError(subscription.Insert(s.MyDB.DB, boil.Infer()))
	return subscription
}

func (s *ModelsSuite) CreateReaction(user *models.User, kind, sType, sUID string) *models.Reaction {

	reaction := &models.Reaction{UserID: user.ID}

	if kind != "" {
		reaction.Kind = kind
	} else {
		reaction.Kind = utils.GenerateName(rand.Intn(15) + 1)
	}

	if sType != "" {
		reaction.SubjectType = sType
	} else {
		reaction.SubjectType = utils.GenerateName(rand.Intn(15) + 1)
	}

	if sUID != "" {
		reaction.SubjectUID = sUID
	} else {
		reaction.SubjectUID = utils.GenerateUID(8)
	}

	s.NoError(reaction.Insert(s.MyDB.DB, boil.Infer()))

	return reaction
}

func (s *ModelsSuite) CreateBookmark(user *models.User, name, sType string, properties map[string]interface{}) *models.Bookmark {
	bookmark := &models.Bookmark{
		Name:        null.StringFrom(name),
		UserID:      user.ID,
		SubjectUID:  utils.GenerateUID(8),
		SubjectType: sType,
	}
	if sType != "" {
		bookmark.SubjectType = "TEST_CONTENT_TYPE"
	}

	if properties != nil {
		dataJson, err := json.Marshal(properties)
		s.Require().NoError(err)
		bookmark.Properties = null.JSONFrom(dataJson)
	}

	s.Require().NoError(bookmark.Insert(s.MyDB.DB, boil.Infer()))

	return bookmark
}

func (s *ModelsSuite) CreateFolder(user *models.User, name string) *models.Folder {
	folder := &models.Folder{
		Name:   null.StringFrom(name),
		UserID: user.ID,
	}

	s.Require().NoError(folder.Insert(s.MyDB.DB, boil.Infer()))
	return folder
}

func (s *ModelsSuite) CreateLabel(user *models.User, name, sType, lang string, properties map[string]interface{}) *models.Label {
	label := &models.Label{
		ID:          0,
		UID:         utils.GenerateUID(8),
		Name:        null.StringFrom(name),
		UserID:      user.ID,
		SubjectUID:  utils.GenerateUID(8),
		SubjectType: sType,
		Language:    lang,
	}
	if sType != "" {
		label.SubjectType = "TEST_CONTENT_TYPE"
	}

	if properties != nil {
		dataJson, err := json.Marshal(properties)
		s.Require().NoError(err)
		label.Properties = null.JSONFrom(dataJson)
	}

	s.Require().NoError(label.Insert(s.MyDB.DB, boil.Infer()))

	lts := make([]*models.LabelTag, 2)
	for i, _ := range lts {
		lts[i] = &models.LabelTag{TagUID: utils.GenerateUID(8)}
	}
	s.Require().NoError(label.AddLabelTags(s.MyDB.DB, true, lts...))

	return label
}
