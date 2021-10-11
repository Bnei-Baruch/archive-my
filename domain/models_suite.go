package domain

import (
	"encoding/json"
	"math/rand"

	"github.com/stretchr/testify/suite"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/Bnei-Baruch/archive-my/models"
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
