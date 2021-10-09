package domain

import (
	pkgerr "github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/Bnei-Baruch/archive-my/models"
	"github.com/Bnei-Baruch/archive-my/pkg/utils"
)

type UIDChecker interface {
	Check(exec boil.Executor, uid string) (exists bool, err error)
}

type PlaylistUIDChecker struct{}

func (c *PlaylistUIDChecker) Check(exec boil.Executor, uid string) (exists bool, err error) {
	return models.Playlists(models.PlaylistWhere.UID.EQ(uid)).Exists(exec)
}

func GetFreeUID(exec boil.Executor, checker UIDChecker) (uid string, err error) {
	for {
		uid = utils.GenerateUID(8)
		exists, ex := checker.Check(exec, uid)
		if ex != nil {
			err = pkgerr.Wrap(ex, "Check UID exists")
			break
		}
		if !exists {
			break
		}
	}

	return
}
