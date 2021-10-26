package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	pkgerr "github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/net/context"

	"github.com/Bnei-Baruch/archive-my/databases/mydb/models"
	"github.com/Bnei-Baruch/archive-my/domain"
	"github.com/Bnei-Baruch/archive-my/pkg/errs"
	"github.com/Bnei-Baruch/archive-my/pkg/sqlutil"
	"github.com/Bnei-Baruch/archive-my/pkg/utils"
)

const (
	DefaultPageSize = 50
	MaxPageSize     = 1000
	MaxPlaylistSize = 1000
)

//Playlist handlers
func (a *App) handleGetPlaylists(c *gin.Context) {
	var r GetPlaylistsRequest
	if c.Bind(&r) != nil {
		return
	}

	user := c.MustGet("USER").(*models.User)
	if user.Disabled || user.RemovedAt.Valid {
		errs.NewForbiddenError(errors.New("inactive user")).Abort(c)
		return
	}

	mods := []qm.QueryMod{qm.Where("user_id = ?", user.ID)}

	db := c.MustGet("MY_DB").(*sql.DB)
	total, err := models.Playlists(mods...).Count(db)
	if err != nil {
		errs.NewInternalError(pkgerr.WithStack(err)).Abort(c)
		return
	}

	if total == 0 {
		concludeRequest(c, NewPlaylistsResponse(0, 0), nil)
		return
	}

	_, offset := appendListMods(&mods, r.ListRequest)
	if int64(offset) >= total {
		concludeRequest(c, NewPlaylistsResponse(0, 0), nil)
		return
	}

	items, err := models.Playlists(mods...).All(db)
	if err != nil {
		errs.NewInternalError(pkgerr.WithStack(err)).Abort(c)
		return
	}

	playlistIDs := make([]int64, len(items))
	playlistDTOByID := make(map[int64]*Playlist, len(items))
	resp := NewPlaylistsResponse(total, len(items))
	for i, x := range items {
		resp.Items[i] = makePlaylistDTO(x)
		playlistIDs[i] = x.ID
		playlistDTOByID[x.ID] = resp.Items[i]
	}

	// total_items histogram
	rows, err := models.NewQuery(
		qm.Select(models.PlaylistItemColumns.PlaylistID, "count(*)"),
		qm.From(models.TableNames.PlaylistItems),
		models.PlaylistItemWhere.PlaylistID.IN(playlistIDs),
		qm.GroupBy(models.PlaylistItemColumns.PlaylistID),
	).Query(db)
	if err != nil {
		errs.NewInternalError(pkgerr.WithStack(err)).Abort(c)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var playlistID int64
		var totalItems int
		if err := rows.Scan(&playlistID, &totalItems); err != nil {
			errs.NewInternalError(pkgerr.WithStack(err)).Abort(c)
			return
		}
		playlistDTOByID[playlistID].TotalItems = totalItems
	}

	if err := rows.Err(); err != nil {
		errs.NewInternalError(pkgerr.WithStack(err)).Abort(c)
		return
	}

	concludeRequest(c, resp, nil)
}

func (a *App) handleCreatePlaylist(c *gin.Context) {
	var r PlaylistRequest
	if c.BindJSON(&r) != nil {
		return
	}

	user := c.MustGet("USER").(*models.User)
	if user.Disabled || user.RemovedAt.Valid {
		errs.NewForbiddenError(errors.New("inactive user")).Abort(c)
		return
	}

	var resp interface{}
	db := c.MustGet("MY_DB").(*sql.DB)
	err := sqlutil.InTx(context.TODO(), db, func(tx *sql.Tx) error {
		uid, err := domain.GetFreeUID(tx, new(domain.PlaylistUIDChecker))
		if err != nil {
			return pkgerr.Wrap(err, "get free UID")
		}

		playlist := models.Playlist{
			UserID: user.ID,
			UID:    uid,
			Public: r.Public,
		}

		if r.Name != "" {
			playlist.Name = null.StringFrom(r.Name)
		}

		if r.Properties != nil {
			props, err := json.Marshal(r.Properties)
			if err != nil {
				return pkgerr.Wrap(err, "json.Marshal properties")
			}
			playlist.Properties = null.JSONFrom(props)
		}

		if err := playlist.Insert(tx, boil.Infer()); err != nil {
			return pkgerr.WithStack(err)
		}

		resp = makePlaylistDTO(&playlist)

		return nil
	})

	concludeRequest(c, resp, err)
}

func (a *App) handleGetPlaylist(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil {
		errs.NewBadRequestError(pkgerr.Wrap(err, "id expects int64")).Abort(c)
		return
	}

	user := c.MustGet("USER").(*models.User)
	if user.Disabled || user.RemovedAt.Valid {
		errs.NewForbiddenError(errors.New("inactive user")).Abort(c)
		return
	}

	db := c.MustGet("MY_DB").(*sql.DB)

	playlist, err := models.FindPlaylist(db, id)
	if err != nil {
		if err != sql.ErrNoRows {
			errs.NewInternalError(pkgerr.WithStack(err)).Abort(c)
			return
		}
		errs.NewNotFoundError(err).Abort(c)
		return
	}
	if playlist.UserID != user.ID {
		errs.NewNotFoundError(errors.New("owner mismatch")).Abort(c)
		return
	}

	if err := playlist.L.LoadPlaylistItems(db, true, playlist, nil); err != nil {
		errs.NewInternalError(pkgerr.WithStack(err)).Abort(c)
		return
	}

	resp := makePlaylistDTO(playlist)

	concludeRequest(c, resp, nil)
}

func (a *App) handleUpdatePlaylist(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil {
		errs.NewBadRequestError(pkgerr.Wrap(err, "id expects int64")).Abort(c)
		return
	}

	var r PlaylistRequest
	if c.BindJSON(&r) != nil {
		return
	}

	user := c.MustGet("USER").(*models.User)
	if user.Disabled || user.RemovedAt.Valid {
		errs.NewForbiddenError(errors.New("inactive user")).Abort(c)
		return
	}

	var resp interface{}
	db := c.MustGet("MY_DB").(*sql.DB)
	err = sqlutil.InTx(context.TODO(), db, func(tx *sql.Tx) error {
		playlist, err := models.FindPlaylist(tx, id)
		if err != nil {
			if err != sql.ErrNoRows {
				return pkgerr.Wrap(err, "fetch playlist from db")
			}
			return errs.NewNotFoundError(err)
		}
		if playlist.UserID != user.ID {
			return errs.NewNotFoundError(errors.New("owner mismatch"))
		}

		playlist.Public = r.Public

		if r.Name == "" {
			playlist.Name = null.NewString("", false)
		} else {
			playlist.Name = null.StringFrom(r.Name)
		}

		if r.Properties != nil {
			props, err := json.Marshal(r.Properties)
			if err != nil {
				return pkgerr.Wrap(err, "json.Marshal properties")
			}
			playlist.Properties = null.JSONFrom(props)
		}

		_, err = playlist.Update(tx, boil.Whitelist(models.PlaylistColumns.Name,
			models.PlaylistColumns.Public,
			models.PlaylistColumns.Properties))
		if err != nil {
			return pkgerr.Wrap(err, "update DB")
		}

		resp = makePlaylistDTO(playlist)

		return nil
	})

	concludeRequest(c, resp, err)
}

func (a *App) handleDeletePlaylist(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil {
		errs.NewBadRequestError(pkgerr.Wrap(err, "id expects int64")).Abort(c)
		return
	}

	user := c.MustGet("USER").(*models.User)
	if user.Disabled || user.RemovedAt.Valid {
		errs.NewForbiddenError(errors.New("inactive user")).Abort(c)
		return
	}

	db := c.MustGet("MY_DB").(*sql.DB)
	err = sqlutil.InTx(context.TODO(), db, func(tx *sql.Tx) error {
		playlist, err := models.FindPlaylist(tx, id)
		if err != nil {
			if err != sql.ErrNoRows {
				return pkgerr.Wrap(err, "fetch playlist from db")
			}
			return errs.NewNotFoundError(err)
		}
		if playlist.UserID != user.ID {
			return errs.NewNotFoundError(errors.New("owner mismatch"))
		}

		if _, err := playlist.Delete(tx); err != nil {
			return pkgerr.Wrap(err, "delete playlist from db")
		}

		return nil
	})

	concludeRequest(c, nil, err)
}

func (a *App) handleAddPlaylistItems(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil {
		errs.NewBadRequestError(pkgerr.Wrap(err, "id expects int64")).Abort(c)
		return
	}

	var r AddPlaylistItemsRequest
	if c.BindJSON(&r) != nil {
		return
	}

	user := c.MustGet("USER").(*models.User)
	if user.Disabled || user.RemovedAt.Valid {
		errs.NewForbiddenError(errors.New("inactive user")).Abort(c)
		return
	}

	var resp interface{}
	db := c.MustGet("MY_DB").(*sql.DB)
	err = sqlutil.InTx(context.TODO(), db, func(tx *sql.Tx) error {
		playlist, err := models.FindPlaylist(tx, id)
		if err != nil {
			if err != sql.ErrNoRows {
				return pkgerr.Wrap(err, "fetch playlist from db")
			}
			return errs.NewNotFoundError(err)
		}
		if playlist.UserID != user.ID {
			return errs.NewNotFoundError(errors.New("owner mismatch"))
		}

		items := make([]*models.PlaylistItem, len(r.Items))
		for i, item := range r.Items {
			items[i] = &models.PlaylistItem{
				Position:       item.Position,
				ContentUnitUID: item.ContentUnitUID,
			}
		}

		// enforce max playlist length
		total, err := models.PlaylistItems(models.PlaylistItemWhere.PlaylistID.EQ(id)).Count(tx)
		if err != nil {
			return pkgerr.Wrap(err, "count items in db")
		}
		if int(total)+len(items) > MaxPlaylistSize {
			return errs.NewBadRequestError(fmt.Errorf("max playlist size is %d", MaxPlaylistSize))
		}

		if err := playlist.AddPlaylistItems(tx, true, items...); err != nil {
			return pkgerr.Wrap(err, "insert items to db")
		}

		if err := playlist.L.LoadPlaylistItems(tx, true, playlist, nil); err != nil {
			return pkgerr.Wrap(err, "reload playlist items from db")
		}
		resp = makePlaylistDTO(playlist)

		return nil
	})

	concludeRequest(c, resp, err)
}

func (a *App) handleUpdatePlaylistItems(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil {
		errs.NewBadRequestError(pkgerr.Wrap(err, "id expects int64")).Abort(c)
		return
	}

	var r UpdatePlaylistItemsRequest
	if c.BindJSON(&r) != nil {
		return
	}

	user := c.MustGet("USER").(*models.User)
	if user.Disabled || user.RemovedAt.Valid {
		errs.NewForbiddenError(errors.New("inactive user")).Abort(c)
		return
	}

	var resp interface{}
	db := c.MustGet("MY_DB").(*sql.DB)
	err = sqlutil.InTx(context.TODO(), db, func(tx *sql.Tx) error {
		playlist, err := models.FindPlaylist(tx, id)
		if err != nil {
			if err != sql.ErrNoRows {
				return pkgerr.Wrap(err, "fetch playlist from db")
			}
			return errs.NewNotFoundError(err)
		}
		if playlist.UserID != user.ID {
			return errs.NewNotFoundError(errors.New("owner mismatch"))
		}

		for _, itemInfo := range r.Items {
			item, err := models.FindPlaylistItem(tx, itemInfo.ID)
			if err != nil {
				if err != sql.ErrNoRows {
					return pkgerr.Wrap(err, "fetch playlist item from db")
				}
				return errs.NewNotFoundError(err)
			}
			if playlist.ID != item.PlaylistID {
				return errs.NewNotFoundError(errors.New("parent playlist mismatch"))
			}

			item.Position = itemInfo.Position
			item.ContentUnitUID = itemInfo.ContentUnitUID
			if _, err := item.Update(tx, boil.Whitelist(models.PlaylistItemColumns.Position,
				models.PlaylistItemColumns.ContentUnitUID)); err != nil {
				return pkgerr.Wrap(err, "update playlist item in db")
			}
		}

		if err := playlist.L.LoadPlaylistItems(tx, true, playlist, nil); err != nil {
			return pkgerr.Wrap(err, "reload playlist items from db")
		}
		resp = makePlaylistDTO(playlist)

		return nil
	})

	concludeRequest(c, resp, err)
}

func (a *App) handleRemovePlaylistItems(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil {
		errs.NewBadRequestError(pkgerr.Wrap(err, "id expects int64")).Abort(c)
		return
	}

	var r RemovePlaylistItemsRequest
	if c.BindJSON(&r) != nil {
		return
	}

	user := c.MustGet("USER").(*models.User)
	if user.Disabled || user.RemovedAt.Valid {
		errs.NewForbiddenError(errors.New("inactive user")).Abort(c)
		return
	}

	var resp interface{}
	db := c.MustGet("MY_DB").(*sql.DB)
	err = sqlutil.InTx(context.TODO(), db, func(tx *sql.Tx) error {
		playlist, err := models.FindPlaylist(tx, id)
		if err != nil {
			if err != sql.ErrNoRows {
				return pkgerr.Wrap(err, "fetch playlist from db")
			}
			return errs.NewNotFoundError(err)
		}
		if playlist.UserID != user.ID {
			return errs.NewNotFoundError(errors.New("owner mismatch"))
		}

		for _, itemID := range r.IDs {
			item, err := models.FindPlaylistItem(tx, itemID)
			if err != nil {
				if err != sql.ErrNoRows {
					return pkgerr.Wrap(err, "fetch playlist item from db")
				}
				return errs.NewNotFoundError(err)
			}
			if playlist.ID != item.PlaylistID {
				return errs.NewNotFoundError(errors.New("parent playlist mismatch"))
			}

			if _, err := item.Delete(tx); err != nil {
				return pkgerr.Wrap(err, "delete playlist item from db")
			}
		}

		if err := playlist.L.LoadPlaylistItems(tx, true, playlist, nil); err != nil {
			return pkgerr.Wrap(err, "reload playlist items from db")
		}
		resp = makePlaylistDTO(playlist)

		return nil
	})

	concludeRequest(c, resp, err)
}

// Reaction handlers
func (a *App) handleGetReactions(c *gin.Context) {
	var r GetReactionsRequest
	if c.Bind(&r) != nil {
		return
	}

	user := c.MustGet("USER").(*models.User)
	if user.Disabled || user.RemovedAt.Valid {
		errs.NewForbiddenError(errors.New("inactive user")).Abort(c)
		return
	}

	mods := []qm.QueryMod{models.ReactionWhere.UserID.EQ(user.ID)}
	if len(r.UIDs) > 0 {
		if r.SubjectType == "" {
			errs.NewBadRequestError(errors.New("missing field subject_type")).Abort(c)
			return
		}
		mods = append(mods, models.ReactionWhere.SubjectUID.IN(r.UIDs))
	}

	if r.SubjectType != "" {
		mods = append(mods, models.ReactionWhere.SubjectType.EQ(r.SubjectType))
	}

	db := c.MustGet("MY_DB").(*sql.DB)
	total, err := models.Reactions(mods...).Count(db)
	if err != nil {
		errs.NewInternalError(pkgerr.WithStack(err)).Abort(c)
		return
	}
	if r.OrderBy == "" {
		r.OrderBy = models.ReactionColumns.ID
	}
	_, offset := appendListMods(&mods, r.ListRequest)
	if int64(offset) >= total {
		concludeRequest(c, new(ReactionsResponse), nil)
		return
	}

	items, err := models.Reactions(mods...).All(db)
	if err != nil {
		errs.NewInternalError(pkgerr.WithStack(err)).Abort(c)
		return
	}

	resp := &ReactionsResponse{
		Items:        make([]*Reaction, len(items)),
		ListResponse: ListResponse{Total: total},
	}
	for i, x := range items {
		resp.Items[i] = &Reaction{
			Kind:        x.Kind,
			SubjectType: x.SubjectType,
			SubjectUID:  x.SubjectUID,
		}
	}

	concludeRequest(c, resp, nil)
}

func (a *App) handleAddReactions(c *gin.Context) {
	var r AddReactionsRequest
	if c.BindJSON(&r) != nil {
		return
	}

	user := c.MustGet("USER").(*models.User)
	if user.Disabled || user.RemovedAt.Valid {
		errs.NewForbiddenError(errors.New("inactive user")).Abort(c)
		return
	}

	var resp interface{}
	db := c.MustGet("MY_DB").(*sql.DB)
	err := sqlutil.InTx(context.TODO(), db, func(tx *sql.Tx) error {
		reaction := models.Reaction{
			UserID:      user.ID,
			Kind:        r.Kind,
			SubjectType: r.SubjectType,
			SubjectUID:  r.SubjectUID,
		}
		if err := reaction.Insert(tx, boil.Infer()); err != nil {
			return pkgerr.WithStack(err)
		}

		resp = Reaction{
			Kind:        reaction.Kind,
			SubjectType: reaction.SubjectType,
			SubjectUID:  reaction.SubjectUID,
		}

		return nil
	})

	concludeRequest(c, resp, err)
}

func (a *App) handleRemoveReactions(c *gin.Context) {
	var r RemoveReactionsRequest
	if c.BindJSON(&r) != nil {
		return
	}

	user := c.MustGet("USER").(*models.User)
	if user.Disabled || user.RemovedAt.Valid {
		errs.NewForbiddenError(errors.New("inactive user")).Abort(c)
		return
	}

	db := c.MustGet("MY_DB").(*sql.DB)
	err := sqlutil.InTx(context.TODO(), db, func(tx *sql.Tx) error {
		count, err := models.Reactions(models.ReactionWhere.UserID.EQ(user.ID),
			models.ReactionWhere.Kind.EQ(r.Kind),
			models.ReactionWhere.SubjectType.EQ(r.SubjectType),
			models.ReactionWhere.SubjectUID.EQ(r.SubjectUID)).DeleteAll(tx)
		if err != nil {
			return err
		}
		if count == 0 {
			return errs.NewNotFoundError(sql.ErrNoRows)
		}
		return nil
	})

	concludeRequest(c, nil, err)
}

func (a *App) handleReactionCount(c *gin.Context) {
	var req UIDsFilter
	if c.Bind(&req) != nil {
		return
	}

	mods := []qm.QueryMod{
		qm.Select(models.ReactionColumns.SubjectUID, models.ReactionColumns.SubjectType, models.ReactionColumns.Kind, "count(id)"),
		qm.From(models.TableNames.Reactions),
		qm.GroupBy(models.ReactionColumns.SubjectUID),
		qm.GroupBy(models.ReactionColumns.Kind),
	}

	db := c.MustGet("MY_DB").(*sql.DB)
	if len(req.UIDs) > 0 {
		mods = append(mods, models.ReactionWhere.SubjectUID.IN(req.UIDs))
	}

	rows, err := models.NewQuery(mods...).Query(db)
	if err != nil {
		errs.NewInternalError(pkgerr.WithStack(err)).Abort(c)
		return
	}
	defer rows.Close()

	resp := make([]*ReactionCount, 0)
	for rows.Next() {
		r := &ReactionCount{}
		err = rows.Scan(r.SubjectUID, r.SubjectType, r.Kind, r.Total)
		resp = append(resp, r)
	}

	concludeRequest(c, resp, nil)
}

//History handlers
func (a *App) handleGetHistory(c *gin.Context) {
	var r GetHistoryRequest
	if c.Bind(&r) != nil {
		return
	}

	user := c.MustGet("USER").(*models.User)
	if user.Disabled || user.RemovedAt.Valid {
		errs.NewForbiddenError(errors.New("inactive user")).Abort(c)
		return
	}

	mods := []qm.QueryMod{qm.Where("user_id = ?", user.ID)}

	db := c.MustGet("MY_DB").(*sql.DB)
	total, err := models.Histories(mods...).Count(db)
	if err != nil {
		errs.NewInternalError(pkgerr.WithStack(err)).Abort(c)
		return
	}

	_, offset := appendListMods(&mods, r.ListRequest)
	if int64(offset) >= total {
		concludeRequest(c, new(ReactionsResponse), nil)
		return
	}

	items, err := models.Histories(mods...).All(db)
	if err != nil {
		errs.NewInternalError(pkgerr.WithStack(err)).Abort(c)
		return
	}

	// TODO: make sure items are compacted on null values (JSON marshal)
	resp := &HistoryResponse{
		Items:        make([]*History, len(items)),
		ListResponse: ListResponse{Total: total},
	}
	for i, x := range items {
		resp.Items[i] = &History{
			ID:             x.ID,
			ContentUnitUID: x.ContentUnitUID,
			Data:           x.Data,
			Timestamp:      x.ChroniclesTimestamp,
			CreatedAt:      x.CreatedAt,
		}
	}

	concludeRequest(c, resp, nil)
}

func (a *App) handleDeleteHistory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil {
		errs.NewBadRequestError(pkgerr.Wrap(err, "id expects int64")).Abort(c)
		return
	}

	user := c.MustGet("USER").(*models.User)
	if user.Disabled || user.RemovedAt.Valid {
		errs.NewForbiddenError(errors.New("inactive user")).Abort(c)
		return
	}

	db := c.MustGet("MY_DB").(*sql.DB)
	err = sqlutil.InTx(context.TODO(), db, func(tx *sql.Tx) error {
		count, err := models.Histories(models.HistoryWhere.UserID.EQ(user.ID), models.HistoryWhere.ID.EQ(id)).DeleteAll(tx)
		if err != nil {
			return err
		}
		if count == 0 {
			return errs.NewNotFoundError(sql.ErrNoRows)
		}
		return nil
	})

	concludeRequest(c, nil, err)
}

//Subscription handlers

func (a *App) handleGetSubscriptions(c *gin.Context) {
	var r GetSubscriptionsRequest
	if c.Bind(&r) != nil {
		return
	}

	user := c.MustGet("USER").(*models.User)
	if user.Disabled || user.RemovedAt.Valid {
		errs.NewForbiddenError(errors.New("inactive user")).Abort(c)
		return
	}

	mods := []qm.QueryMod{qm.Where("user_id = ?", user.ID)}

	db := c.MustGet("MY_DB").(*sql.DB)
	total, err := models.Subscriptions(mods...).Count(db)
	if err != nil {
		errs.NewInternalError(pkgerr.WithStack(err)).Abort(c)
		return
	}

	if r.OrderBy == "" {
		r.OrderBy = "updated_at DESC"
	}
	_, offset := appendListMods(&mods, r.ListRequest)
	if int64(offset) >= total {
		concludeRequest(c, new(ReactionsResponse), nil)
		return
	}

	items, err := models.Subscriptions(mods...).All(db)
	if err != nil {
		errs.NewInternalError(pkgerr.WithStack(err)).Abort(c)
		return
	}

	resp := &SubscriptionsResponse{
		Items:        make([]*Subscription, len(items)),
		ListResponse: ListResponse{Total: total},
	}
	for i, x := range items {
		resp.Items[i] = &Subscription{
			ID:             x.ID,
			CollectionUID:  x.CollectionUID,
			ContentType:    x.ContentType,
			ContentUnitUID: x.ContentUnitUID,
			CreatedAt:      x.CreatedAt,
			UpdatedAt:      x.UpdatedAt,
		}
	}

	concludeRequest(c, resp, nil)
}

func (a *App) handleSubscribe(c *gin.Context) {
	var r SubscribeRequest
	if c.BindJSON(&r) != nil {
		return
	}

	if r.CollectionUID == "" && r.ContentType == "" {
		errs.NewBadRequestError(errors.New("collection_uid and content_type are mutually exclusive")).Abort(c)
		return
	}

	user := c.MustGet("USER").(*models.User)
	if user.Disabled || user.RemovedAt.Valid {
		errs.NewForbiddenError(errors.New("inactive user")).Abort(c)
		return
	}

	var resp interface{}
	db := c.MustGet("MY_DB").(*sql.DB)
	err := sqlutil.InTx(context.TODO(), db, func(tx *sql.Tx) error {
		subscription := models.Subscription{
			UserID:         user.ID,
			CollectionUID:  null.NewString(r.CollectionUID, r.CollectionUID != ""),
			ContentType:    null.NewString(r.ContentType, r.ContentType != ""),
			ContentUnitUID: null.NewString(r.ContentUnitUID, r.ContentUnitUID != ""),
		}
		if err := subscription.Insert(tx, boil.Infer()); err != nil {
			return pkgerr.WithStack(err)
		}

		resp = Subscription{
			ID:             subscription.ID,
			CollectionUID:  subscription.CollectionUID,
			ContentType:    subscription.ContentType,
			ContentUnitUID: subscription.ContentUnitUID,
			CreatedAt:      subscription.CreatedAt,
		}

		return nil
	})

	concludeRequest(c, resp, err)
}

func (a *App) handleUnsubscribe(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil {
		errs.NewBadRequestError(pkgerr.Wrap(err, "id expects int64")).Abort(c)
		return
	}

	user := c.MustGet("USER").(*models.User)
	if user.Disabled || user.RemovedAt.Valid {
		errs.NewForbiddenError(errors.New("inactive user")).Abort(c)
		return
	}

	db := c.MustGet("MY_DB").(*sql.DB)
	err = sqlutil.InTx(context.TODO(), db, func(tx *sql.Tx) error {
		count, err := models.Subscriptions(models.SubscriptionWhere.UserID.EQ(user.ID), models.SubscriptionWhere.ID.EQ(id)).DeleteAll(tx)
		if err != nil {
			return err
		}
		if count == 0 {
			return errs.NewNotFoundError(sql.ErrNoRows)
		}
		return nil
	})

	concludeRequest(c, nil, err)
}

//help functions

func appendListMods(mods *[]qm.QueryMod, r ListRequest) (int, int) {
	if r.OrderBy != "" {
		*mods = append(*mods, qm.OrderBy(r.OrderBy))
	} else {
		*mods = append(*mods, qm.OrderBy("created_at DESC"))
	}

	var limit, offset int

	if r.PageSize == 0 {
		limit = DefaultPageSize
	} else {
		limit = utils.Min(r.PageSize, MaxPageSize)
	}
	if r.PageNumber > 1 {
		offset = (r.PageNumber - 1) * limit
	}

	*mods = append(*mods, qm.Limit(limit))
	if offset != 0 {
		*mods = append(*mods, qm.Offset(offset))
	}

	return limit, offset
}

func concludeRequest(c *gin.Context, resp interface{}, err error) {
	if err != nil {
		var hErr *errs.HttpError
		if errors.As(err, &hErr) {
			hErr.Abort(c)
		} else {
			errs.NewInternalError(err).Abort(c)
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

func makePlaylistDTO(playlist *models.Playlist) *Playlist {
	resp := Playlist{
		ID:        playlist.ID,
		UID:       playlist.UID,
		UserID:    playlist.UserID,
		Public:    playlist.Public,
		CreatedAt: playlist.CreatedAt,
	}
	if playlist.Name.Valid {
		resp.Name = playlist.Name.String
	}
	if playlist.Properties.Valid {
		utils.Must(playlist.Properties.Unmarshal(&resp.Properties))
	}

	if playlist.R != nil && len(playlist.R.PlaylistItems) > 0 {
		resp.TotalItems = len(playlist.R.PlaylistItems)

		resp.Items = make([]*PlaylistItem, resp.TotalItems)
		sort.SliceStable(playlist.R.PlaylistItems, func(i int, j int) bool {
			return playlist.R.PlaylistItems[i].Position < playlist.R.PlaylistItems[j].Position
		})

		for i, item := range playlist.R.PlaylistItems {
			resp.Items[i] = &PlaylistItem{
				ID:             item.ID,
				Position:       item.Position,
				ContentUnitUID: item.ContentUnitUID,
			}
		}
	}

	return &resp
}
