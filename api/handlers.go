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

	user := &models.User{}
	if err := a.validateUser(c, user); err != nil {
		err.Abort(c)
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
	//add playlist item to response if this item of specific unit (need for mark it on client)
	if r.ExistUnit != "" {
		mods = append(mods, qm.Load("PlaylistItems", models.PlaylistItemWhere.ContentUnitUID.EQ(r.ExistUnit)))
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

	user := &models.User{}
	if err := a.validateUser(c, user); err != nil {
		err.Abort(c)
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

	user := &models.User{}
	if err := a.validateUser(c, user); err != nil {
		err.Abort(c)
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

	user := &models.User{}
	if err := a.validateUser(c, user); err != nil {
		err.Abort(c)
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

	user := &models.User{}
	if err := a.validateUser(c, user); err != nil {
		err.Abort(c)
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

// PlaylistItem handlers
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

	user := &models.User{}
	if err := a.validateUser(c, user); err != nil {
		err.Abort(c)
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

		var maxPosition null.Int
		err = models.NewQuery(
			qm.Select(fmt.Sprintf("MAX(%s)", models.PlaylistItemColumns.Position)),
			qm.From(models.TableNames.PlaylistItems),
			models.PlaylistItemWhere.PlaylistID.EQ(id),
		).QueryRow(tx).Scan(&maxPosition)
		if err != nil && err != sql.ErrNoRows {
			return pkgerr.Wrap(err, "find max position of playlist items from db")
		}

		items := make([]*models.PlaylistItem, len(r.Items))
		var maxPos = 1
		if ok := maxPosition.Valid; ok {
			maxPos = maxPosition.Int
		}
		for i, item := range r.Items {
			p := item.Position
			if item.Position < 0 {
				maxPos++
				p = maxPos
			}
			items[i] = &models.PlaylistItem{
				Position:       p,
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

		if !playlist.PosterUnitUID.Valid {
			playlist.PosterUnitUID = null.StringFrom(items[0].ContentUnitUID)
			if _, err := playlist.Update(tx, boil.Whitelist(models.PlaylistColumns.PosterUnitUID)); err != nil {
				return pkgerr.Wrap(err, "insert poster uid to db")
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

	user := &models.User{}
	if err := a.validateUser(c, user); err != nil {
		err.Abort(c)
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

	user := &models.User{}
	if err := a.validateUser(c, user); err != nil {
		err.Abort(c)
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

	user := &models.User{}
	if err := a.validateUser(c, user); err != nil {
		err.Abort(c)
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
		r.OrderBy = fmt.Sprintf("%s DESC", models.ReactionColumns.ID)
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

	user := &models.User{}
	if err := a.validateUser(c, user); err != nil {
		err.Abort(c)
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

	user := &models.User{}
	if err := a.validateUser(c, user); err != nil {
		err.Abort(c)
		return
	}

	db := c.MustGet("MY_DB").(*sql.DB)
	err := sqlutil.InTx(context.TODO(), db, func(tx *sql.Tx) error {
		_, err := models.Reactions(models.ReactionWhere.UserID.EQ(user.ID),
			models.ReactionWhere.Kind.EQ(r.Kind),
			models.ReactionWhere.SubjectType.EQ(r.SubjectType),
			models.ReactionWhere.SubjectUID.EQ(r.SubjectUID)).DeleteAll(tx)
		return err
	})

	concludeRequest(c, nil, err)
}

func (a *App) handleReactionCount(c *gin.Context) {
	var req ReactionCountRequest
	if c.Bind(&req) != nil {
		return
	}

	mods := []qm.QueryMod{
		qm.Select(models.ReactionColumns.SubjectUID, models.ReactionColumns.Kind, "count(id)"),
		qm.From(models.TableNames.Reactions),
		models.ReactionWhere.SubjectType.EQ(req.SubjectType),
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
		r.SubjectType = req.SubjectType
		err = rows.Scan(&r.SubjectUID, &r.Kind, &r.Total)
		resp = append(resp, r)
	}

	if err := rows.Err(); err != nil {
		errs.NewInternalError(pkgerr.WithStack(err)).Abort(c)
		return
	}

	concludeRequest(c, resp, nil)
}

//History handlers
func (a *App) handleGetHistory(c *gin.Context) {
	var r GetHistoryRequest
	if c.Bind(&r) != nil {
		return
	}

	user := &models.User{}
	if err := a.validateUser(c, user); err != nil {
		err.Abort(c)
		return
	}

	mods := []qm.QueryMod{qm.Where("user_id = ?", user.ID)}

	db := c.MustGet("MY_DB").(*sql.DB)
	total, err := models.Histories(mods...).Count(db)
	if err != nil {
		errs.NewInternalError(pkgerr.WithStack(err)).Abort(c)
		return
	}
	if r.OrderBy == "" {
		r.OrderBy = fmt.Sprintf("%s DESC", models.HistoryColumns.ChroniclesTimestamp)
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

	user := &models.User{}
	if err := a.validateUser(c, user); err != nil {
		err.Abort(c)
		return
	}

	db := c.MustGet("MY_DB").(*sql.DB)
	err = sqlutil.InTx(context.TODO(), db, func(tx *sql.Tx) error {
		_, err := models.Histories(models.HistoryWhere.UserID.EQ(user.ID), models.HistoryWhere.ID.EQ(id)).DeleteAll(tx)
		return err
	})

	concludeRequest(c, nil, err)
}

//Subscription handlers
func (a *App) handleGetSubscriptions(c *gin.Context) {
	var r GetSubscriptionsRequest
	if c.Bind(&r) != nil {
		return
	}

	user := &models.User{}
	if err := a.validateUser(c, user); err != nil {
		err.Abort(c)
		return
	}

	mods := []qm.QueryMod{models.SubscriptionWhere.UserID.EQ(user.ID)}
	if r.ContentType != "" {
		mods = append(mods, models.SubscriptionWhere.ContentType.EQ(null.StringFrom(r.ContentType)))
	}
	if r.CollectionUID != "" {
		mods = append(mods, models.SubscriptionWhere.CollectionUID.EQ(null.StringFrom(r.CollectionUID)))
	}

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

	user := &models.User{}
	if err := a.validateUser(c, user); err != nil {
		err.Abort(c)
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

	user := &models.User{}
	if err := a.validateUser(c, user); err != nil {
		err.Abort(c)
		return
	}

	db := c.MustGet("MY_DB").(*sql.DB)
	err = sqlutil.InTx(context.TODO(), db, func(tx *sql.Tx) error {
		_, err := models.Subscriptions(models.SubscriptionWhere.UserID.EQ(user.ID), models.SubscriptionWhere.ID.EQ(id)).DeleteAll(tx)
		return err
	})

	concludeRequest(c, nil, err)
}

//Bookmark handlers
func (a *App) handleGetBookmarks(c *gin.Context) {
	var r GetBookmarksRequest
	if c.Bind(&r) != nil {
		return
	}

	user := &models.User{}
	if err := a.validateUser(c, user); err != nil {
		err.Abort(c)
		return
	}

	db := c.MustGet("MY_DB").(*sql.DB)
	mods := []qm.QueryMod{models.BookmarkWhere.UserID.EQ(user.ID)}

	a.respBookmarks(c, db, mods, r)
}

func (a *App) handleCreateBookmark(c *gin.Context) {
	var r AddBookmarksRequest
	if c.BindJSON(&r) != nil {
		return
	}

	user := &models.User{}
	if err := a.validateUser(c, user); err != nil {
		err.Abort(c)
		return
	}

	db := c.MustGet("MY_DB").(*sql.DB)

	bookmark := &models.Bookmark{
		UserID:     user.ID,
		SourceUID:  r.SourceUID,
		SourceType: r.SourceType,
	}
	if r.Name != "" {
		bookmark.Name = null.StringFrom(r.Name)
	}
	if r.Data != nil {
		data, err := json.Marshal(r.Data)
		bookmark.Data = null.JSONFrom(data)
		if err != nil {
			errs.NewBadRequestError(err).Abort(c)
			return
		}
	}
	if r.Public {
		bookmark.Public = r.Public
	}
	err := sqlutil.InTx(context.TODO(), db, func(tx *sql.Tx) error {
		uid, err := domain.GetFreeUID(tx, new(domain.PlaylistUIDChecker))
		if err != nil {
			return pkgerr.Wrap(err, "get free UID")
		}
		bookmark.UID = uid

		err = bookmark.Insert(tx, boil.Infer())
		if err != nil {
			return pkgerr.WithStack(err)
		}
		if r.FolderIDs != nil {
			bbfs := make([]*models.BookmarkFolder, len(r.FolderIDs))
			for i, id := range r.FolderIDs {
				bbfs[i] = &models.BookmarkFolder{
					FolderID: id,
				}
			}
			if err := bookmark.AddBookmarkFolders(tx, true, bbfs...); err != nil {
				return pkgerr.WithStack(err)
			}
		}

		if r.TagsUIDs != nil {
			bts := make([]*models.BookmarkTag, len(r.TagsUIDs))
			for i, uid := range r.TagsUIDs {
				bts[i] = &models.BookmarkTag{
					TagUID: uid,
				}
			}
			if err := bookmark.AddBookmarkTags(tx, true, bts...); err != nil {
				return pkgerr.WithStack(err)
			}
		}
		return nil
	})
	if err != nil {
		errs.NewInternalError(pkgerr.WithStack(err)).Abort(c)
		return
	}
	resp := makeBookmarkDTO(bookmark)
	resp.FolderIds = r.FolderIDs
	concludeRequest(c, resp, err)
}

func (a *App) handleUpdateBookmark(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil {
		errs.NewBadRequestError(pkgerr.Wrap(err, "id expects int64")).Abort(c)
		return
	}

	var r UpdateBookmarkRequest
	if c.BindJSON(&r) != nil {
		return
	}

	user := &models.User{}
	if err := a.validateUser(c, user); err != nil {
		err.Abort(c)
		return
	}

	resp := &Bookmark{}
	db := c.MustGet("MY_DB").(*sql.DB)
	err = sqlutil.InTx(context.TODO(), db, func(tx *sql.Tx) error {
		b, err := models.Bookmarks(models.BookmarkWhere.ID.EQ(id), qm.Load(models.BookmarkRels.BookmarkFolders)).One(tx)
		if err != nil {
			return pkgerr.WithStack(err)
		}
		if r.Name != "" {
			b.Name = null.StringFrom(r.Name)
		}
		if r.FolderIDs != nil {
			if len(r.FolderIDs) == 0 {
				_, err = b.R.BookmarkFolders.DeleteAll(tx)
				if err != nil {
					return err
				}
			} else {
				forAdd, forDel := diffBFs(r.FolderIDs, b.R.BookmarkFolders)

				bfs := make([]*models.BookmarkFolder, len(forAdd))
				for i, id := range forAdd {
					bfs[i] = &models.BookmarkFolder{
						FolderID: id,
					}
				}

				if len(forAdd) > 0 {
					if err = b.AddBookmarkFolders(tx, true, bfs...); err != nil {
						return err
					}
				}

				if len(forDel) > 0 {
					_, err := models.BookmarkFolders(
						models.BookmarkFolderWhere.FolderID.IN(forDel),
						models.BookmarkFolderWhere.BookmarkID.EQ(b.ID),
					).DeleteAll(tx)
					if err != nil {
						return err
					}
				}
			}
		}

		if r.TagsUIDs != nil {
			bts := make([]*models.BookmarkTag, len(r.TagsUIDs))
			for i, uid := range r.TagsUIDs {
				bts[i] = &models.BookmarkTag{
					TagUID: uid,
				}
			}
			if err := b.AddBookmarkTags(tx, true, bts...); err != nil {
				return pkgerr.WithStack(err)
			}
		}

		if _, err := b.Update(tx, boil.Infer()); err != nil {
			return err
		}

		reloaded, err := models.Bookmarks(models.BookmarkWhere.ID.EQ(id), qm.Load(models.BookmarkRels.BookmarkFolders)).One(tx)
		if err != nil {
			return err
		}
		*resp = *makeBookmarkDTO(reloaded)
		return nil
	})

	concludeRequest(c, resp, err)
}

func (a *App) handleDeleteBookmark(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil {
		errs.NewBadRequestError(pkgerr.Wrap(err, "id expects int64")).Abort(c)
		return
	}

	user := &models.User{}
	if err := a.validateUser(c, user); err != nil {
		err.Abort(c)
		return
	}

	db := c.MustGet("MY_DB").(*sql.DB)
	err = sqlutil.InTx(context.TODO(), db, func(tx *sql.Tx) error {
		b, err := models.FindBookmark(tx, id)

		if err == sql.ErrNoRows {
			return nil
		}

		if err != nil {
			return pkgerr.Wrap(err, "fetch bookmark from db")
		}
		_, err = b.Delete(tx)
		return err
	})

	concludeRequest(c, nil, err)
}

func (a *App) handleGetPublicBookmarks(c *gin.Context) {
	var r GetBookmarksRequest
	if c.Bind(&r) != nil {
		return
	}

	db := c.MustGet("MY_DB").(*sql.DB)
	mods := []qm.QueryMod{models.BookmarkWhere.Public.EQ(true)}

	a.respBookmarks(c, db, mods, r)
}

func (a *App) respBookmarks(c *gin.Context, db *sql.DB, mods []qm.QueryMod, r GetBookmarksRequest) {

	total, err := models.Bookmarks(mods...).Count(db)
	mods = append(mods,
		qm.Load(models.BookmarkRels.BookmarkFolders),
		qm.Load(models.BookmarkRels.BookmarkTags),
	)
	_, offset := appendListMods(&mods, r.ListRequest)
	if int64(offset) >= total {
		concludeRequest(c, new(GetBookmarksResponse), nil)
		return
	}
	appendQueryFilter(&mods, r.QueryFilter, "name")
	if len(r.FolderIDsFilter) > 0 {
		mods = append(mods,
			qm.InnerJoin("bookmark_folder bf ON id = bf.bookmark_id"),
			qm.WhereIn("bf.folder_id in ?", utils.ConvertArgsInt64(r.FolderIDsFilter)...),
		)
	}

	bookmarks, err := models.Bookmarks(mods...).All(db)
	if err != nil {
		errs.NewInternalError(pkgerr.WithStack(err)).Abort(c)
		return
	}

	items := make([]*Bookmark, len(bookmarks))
	for i, b := range bookmarks {
		items[i] = makeBookmarkDTO(b)
	}

	resp := GetBookmarksResponse{
		ListResponse: ListResponse{Total: total},
		Items:        items,
	}
	concludeRequest(c, resp, err)
}

//Folder handlers
func (a *App) handleGetFolders(c *gin.Context) {
	var r GetFoldersRequest
	if c.Bind(&r) != nil {
		return
	}

	user := &models.User{}
	if err := a.validateUser(c, user); err != nil {
		err.Abort(c)
		return
	}

	db := c.MustGet("MY_DB").(*sql.DB)

	mods := []qm.QueryMod{models.FolderWhere.UserID.EQ(user.ID)}

	total, err := models.Folders(mods...).Count(db)

	_, offset := appendListMods(&mods, r.ListRequest)
	if int64(offset) >= total {
		concludeRequest(c, new(GetFoldersResponse), nil)
		return
	}

	if r.BookmarkIdFilter != 0 {
		mods = append(mods, qm.Load("BookmarkFolders", models.BookmarkFolderWhere.BookmarkID.EQ(r.BookmarkIdFilter)))
	}
	appendQueryFilter(&mods, r.QueryFilter, "name")
	folders, err := models.Folders(mods...).All(db)
	if err != nil {
		errs.NewInternalError(pkgerr.WithStack(err)).Abort(c)
		return
	}

	items := make([]*Folder, len(folders))
	for i, f := range folders {
		items[i] = makeFoldersDTO(f)
	}

	resp := GetFoldersResponse{
		ListResponse: ListResponse{Total: total},
		Items:        items,
	}
	concludeRequest(c, resp, err)
}

func (a *App) handleCreateFolder(c *gin.Context) {
	var r AddFolderRequest
	if c.BindJSON(&r) != nil {
		return
	}

	user := &models.User{}
	if err := a.validateUser(c, user); err != nil {
		err.Abort(c)
		return
	}

	db := c.MustGet("MY_DB").(*sql.DB)
	folder := &models.Folder{UserID: user.ID}
	if r.Name != "" {
		folder.Name = null.StringFrom(r.Name)
	}

	err := sqlutil.InTx(context.TODO(), db, func(tx *sql.Tx) error {
		err := folder.Insert(tx, boil.Infer())
		if err != nil {
			return pkgerr.WithStack(err)
		}

		return nil
	})
	if err != nil {
		errs.NewInternalError(pkgerr.WithStack(err)).Abort(c)
		return
	}
	resp := makeFoldersDTO(folder)
	concludeRequest(c, resp, err)
}

func (a *App) handleUpdateFolder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil {
		errs.NewBadRequestError(pkgerr.Wrap(err, "id expects int64")).Abort(c)
		return
	}

	var r UpdateFolderRequest
	if c.BindJSON(&r) != nil {
		return
	}

	user := &models.User{}
	if err := a.validateUser(c, user); err != nil {
		err.Abort(c)
		return
	}

	var resp *Folder
	db := c.MustGet("MY_DB").(*sql.DB)
	err = sqlutil.InTx(context.TODO(), db, func(tx *sql.Tx) error {
		f, err := models.FindFolder(tx, id)
		if err != nil {
			return pkgerr.WithStack(err)
		}
		if r.Name != "" {
			f.Name = null.StringFrom(r.Name)
		}
		_, err = f.Update(tx, boil.Infer())
		resp = makeFoldersDTO(f)
		return err
	})

	concludeRequest(c, resp, err)
}

func (a *App) handleDeleteFolder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil {
		errs.NewBadRequestError(pkgerr.Wrap(err, "id expects int64")).Abort(c)
		return
	}

	user := &models.User{}
	if err := a.validateUser(c, user); err != nil {
		err.Abort(c)
		return
	}

	db := c.MustGet("MY_DB").(*sql.DB)
	err = sqlutil.InTx(context.TODO(), db, func(tx *sql.Tx) error {
		b, err := models.FindFolder(tx, id)

		if err == sql.ErrNoRows {
			return nil
		}

		if err != nil {
			return pkgerr.Wrap(err, "fetch folder from db")
		}
		_, err = b.Delete(tx)
		return err
	})

	concludeRequest(c, nil, err)
}

//help functions

func (a *App) validateUser(c *gin.Context, user *models.User) *errs.HttpError {
	*user = *c.MustGet("USER").(*models.User)
	if user.Disabled || user.RemovedAt.Valid {
		return errs.NewForbiddenError(errors.New("inactive user"))
	}
	return nil

}

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

func appendQueryFilter(mods *[]qm.QueryMod, r QueryFilter, column string) {
	if r.Query == "" {
		return
	}
	q := fmt.Sprintf("%s ILIKE '%%%s%%'", column, r.Query)
	*mods = append(*mods, qm.Where(q))
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
		Public:    playlist.Public,
		CreatedAt: playlist.CreatedAt,
	}
	if playlist.Name.Valid {
		resp.Name = playlist.Name.String
	}
	if playlist.Properties.Valid {
		utils.Must(playlist.Properties.Unmarshal(&resp.Properties))
	}

	if playlist.PosterUnitUID.Valid {
		resp.PosterUnitUID = playlist.PosterUnitUID.String
	}

	if playlist.R != nil && len(playlist.R.PlaylistItems) > 0 {
		resp.TotalItems = len(playlist.R.PlaylistItems)

		resp.Items = make([]*PlaylistItem, resp.TotalItems)
		sort.SliceStable(playlist.R.PlaylistItems, func(i int, j int) bool {
			return playlist.R.PlaylistItems[i].Position > playlist.R.PlaylistItems[j].Position
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

func makeBookmarkDTO(bookmark *models.Bookmark) *Bookmark {
	resp := Bookmark{
		ID:         bookmark.ID,
		UID:        bookmark.UID,
		SourceUID:  bookmark.SourceUID,
		SourceType: bookmark.SourceType,
		Public:     bookmark.Public,
		Accepted:   bookmark.Accepted,
	}

	if bookmark.Name.Valid {
		resp.Name = bookmark.Name.String
	}

	if bookmark.Data.Valid {
		utils.Must(bookmark.Data.Unmarshal(&resp.Data))
	}

	if bookmark.R != nil && bookmark.R.BookmarkFolders != nil {
		resp.FolderIds = make([]int64, len(bookmark.R.BookmarkFolders))
		for i, bbf := range bookmark.R.BookmarkFolders {
			resp.FolderIds[i] = bbf.FolderID
		}
	}

	if bookmark.R != nil && bookmark.R.BookmarkTags != nil {
		resp.TagUIds = make([]string, len(bookmark.R.BookmarkTags))
		for i, bt := range bookmark.R.BookmarkTags {
			resp.TagUIds[i] = bt.TagUID
		}
	}

	return &resp
}

func makeFoldersDTO(folder *models.Folder) *Folder {
	resp := Folder{
		ID: folder.ID,
	}

	if folder.Name.Valid {
		resp.Name = folder.Name.String
	}

	if folder.R != nil && folder.R.BookmarkFolders != nil {
		resp.BookmarkIds = make([]int64, len(folder.R.BookmarkFolders))
		for i, bbf := range folder.R.BookmarkFolders {
			resp.BookmarkIds[i] = bbf.BookmarkID
		}
	}

	return &resp
}

func diffBFs(reqIDs []int64, bfFromDB []*models.BookmarkFolder) ([]int64, []int64) {
	if len(bfFromDB) == 0 {
		return reqIDs, nil
	}

	forDel := make([]int64, 0)
	forAdd := make([]int64, 0)
	mDB := make(map[int64]bool, len(bfFromDB))
	for _, x := range bfFromDB {
		mDB[x.FolderID] = false
	}

	for _, x := range reqIDs {
		if _, ok := mDB[x]; !ok {
			forAdd = append(forAdd, x)
			continue
		}
		mDB[x] = true
	}

	for x, ok := range mDB {
		if !ok {
			forDel = append(forDel, x)
		}
	}
	return forAdd, forDel
}
