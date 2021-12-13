package api

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/gin-gonic/gin"
	pkgerr "github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/Bnei-Baruch/archive-my/databases/mydb/models"
	"github.com/Bnei-Baruch/archive-my/pkg/errs"
	"github.com/Bnei-Baruch/archive-my/pkg/sqlutil"
)

func (a *App) handleGetAllPublicBookmarks(c *gin.Context) {
	var r GetBookmarksRequest
	if c.Bind(&r) != nil {
		return
	}

	db := c.MustGet("MY_DB").(*sql.DB)
	mods := []qm.QueryMod{models.BookmarkWhere.Public.EQ(true)}

	if r.StatusFilter != "" {
		s := null.Bool{
			Bool:  false,
			Valid: false,
		}
		if r.StatusFilter == "accepted" {
			s.Bool = true
			s.Valid = true
		}
		if r.StatusFilter == "declined" {
			s.Bool = false
			s.Valid = true
		}
		mods = append(mods, models.BookmarkWhere.Accepted.EQ(s))
	}

	a.respBookmarks(c, db, mods, r)
}

//Change Bookmark accept handlers
func (a *App) handleBookmarkModeration(c *gin.Context) {
	var r BookmarkModerationRequest
	if c.Bind(&r) != nil {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil {
		errs.NewBadRequestError(pkgerr.Wrap(err, "id expects int64")).Abort(c)
		return
	}

	db := c.MustGet("MY_DB").(*sql.DB)

	var resp *Bookmark
	err = sqlutil.InTx(context.TODO(), db, func(tx *sql.Tx) error {
		b, err := models.FindBookmark(db, id)
		if err != nil {
			return errs.NewNotFoundError(err)
		}

		b.Accepted = r.Accepted

		if _, err := b.Update(tx, boil.Whitelist(models.BookmarkColumns.Accepted)); err != nil {
			return errs.NewNotFoundError(err)
		}
		resp = makeBookmarkDTO(b)
		return nil
	})
	concludeRequest(c, resp, err)
}
