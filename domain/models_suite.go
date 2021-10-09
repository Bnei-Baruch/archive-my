package domain

import (
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

func (s *ModelsSuite) CreatePlaylist(user *models.User, name string) *models.Playlist {
	uid, err := GetFreeUID(s.MyDB.DB, new(PlaylistUIDChecker))
	s.Require().NoError(err, "get free UID")

	playlist := &models.Playlist{
		UserID: user.ID,
		UID:    uid,
		Name:   null.StringFrom(name),
	}
	s.Require().NoError(playlist.Insert(s.MyDB.DB, boil.Infer()))

	items := make([]*models.PlaylistItem, rand.Intn(20)+1)
	for i := range items {
		items[i] = &models.PlaylistItem{
			Position:       i,
			ContentUnitUID: utils.GenerateUID(8),
		}
	}
	s.NoError(playlist.AddPlaylistItems(s.MyDB.DB, true, items...))

	return playlist
}
