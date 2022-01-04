package api

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/gin-gonic/gin"
	pkgerr "github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/Bnei-Baruch/archive-my/databases/mydb/models"
	"github.com/Bnei-Baruch/archive-my/pkg/errs"
	"github.com/Bnei-Baruch/archive-my/pkg/sqlutil"
)

func (a *App) handleGetAllLabels(c *gin.Context) {
	var r GetLabelsAdminRequest
	if c.Bind(&r) != nil {
		return
	}

	db := c.MustGet("MY_DB").(*sql.DB)
	mods := []qm.QueryMod{}

	if !r.GetAll {
		mods = append(mods, models.LabelWhere.Accepted.EQ(r.Accepted))
	}

	a.labelResponse(c, db, mods, r.GetLabelsRequest)
}

//Change Bookmark accept handlers
func (a *App) handleLabelModeration(c *gin.Context) {
	var r LabelModerationRequest
	if c.Bind(&r) != nil {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil {
		errs.NewBadRequestError(pkgerr.Wrap(err, "id expects int64")).Abort(c)
		return
	}

	db := c.MustGet("MY_DB").(*sql.DB)

	var resp *Label
	err = sqlutil.InTx(context.TODO(), db, func(tx *sql.Tx) error {
		b, err := models.FindLabel(db, id)
		if err != nil {
			return errs.NewNotFoundError(err)
		}

		b.Accepted = r.Accepted

		if _, err := b.Update(tx, boil.Whitelist(models.LabelColumns.Accepted)); err != nil {
			return errs.NewNotFoundError(err)
		}
		resp = makeLabelDTO(b)
		return nil
	})
	concludeRequest(c, resp, err)
}
