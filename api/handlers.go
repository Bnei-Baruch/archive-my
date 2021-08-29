package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"

	"github.com/friendsofgo/errors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"archive-my/models"
	"archive-my/pkg/utils"
)

const (
	DefaultPageSize = 50
	MaxPageSize     = 1000
)

//Playlist handlers
func (a *App) handleGetPlaylists(c *gin.Context) {
	kcId := c.MustGet("KC_ID").(string)

	mods := []qm.QueryMod{
		qm.Load("PlaylistItems"),
		qm.Where("account_id = ?", kcId),
	}

	var req ListRequest
	if c.Bind(&req) != nil {
		NewBadRequestError(nil).Abort(c)
	}
	if err := appendMyListMods(&mods, req); err != nil {
		NewBadRequestError(err).Abort(c)
	}
	pls, err := models.Playlists(mods...).All(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	total, err := models.Playlists(qm.Where("account_id = ?", kcId)).Count(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}
	plResp := make([]*playlistResponse, len(pls))
	for i, p := range pls {
		plResp[i] = &playlistResponse{Playlist: p}
		if p.R.PlaylistItems != nil {
			plResp[i].ItemsCount = int64(len(p.R.PlaylistItems))
			plResp[i].PlaylistItems = []*models.PlaylistItem{p.R.PlaylistItems[0]}
		}
	}

	resp := playlistsResponse{ListResponse: ListResponse{Total: total}, Playlists: plResp}
	concludeRequest(c, resp, nil)
}

func (a *App) handleCreatePlaylist(c *gin.Context) {
	kcId := c.MustGet("KC_ID").(string)

	var p models.Playlist
	if c.Bind(&p) != nil {
		NewBadRequestError(nil).Abort(c)
	}
	pl := models.Playlist{
		AccountID:  kcId,
		Name:       p.Name,
		Parameters: p.Parameters,
		Public:     p.Public,
	}

	err := pl.Insert(a.DB, boil.Infer())
	concludeRequest(c, []models.Playlist{pl}, NewInternalError(err))
}

func (a *App) handleUpdatePlaylist(c *gin.Context) {
	id, e := strconv.ParseInt(c.Param("id"), 10, 0)
	if e != nil {
		NewBadRequestError(e).Abort(c)
	}

	kcId := c.MustGet("KC_ID").(string)
	var np models.Playlist
	if c.Bind(&np) != nil {
		NewBadRequestError(nil).Abort(c)
	}

	p, err := models.Playlists(
		qm.Where("id = ?", id),
		qm.Load("PlaylistItems"),
	).One(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	if kcId != p.AccountID {
		NewHttpError(http.StatusNotAcceptable, nil, gin.ErrorTypePrivate).Abort(c)
	}
	if p.Name != np.Name {
		p.Name = np.Name
	}
	if p.LastPlayed != np.LastPlayed {
		p.LastPlayed = np.LastPlayed
	}
	if p.Public != np.Public {
		p.Public = np.Public
	}

	_, err = p.Update(a.DB, boil.Infer())
	if kcId != p.AccountID {
		NewInternalError(err).Abort(c)
	}
	params, err := buildNewParams(&np, p)
	if err != nil {
		NewInternalError(err).Abort(c)
	}
	if params != nil {
		p.Parameters = null.JSONFrom(params)
	}
	_, err = p.Update(a.DB, boil.Infer())
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	concludeRequest(c, p, nil)
}

func (a *App) handleDeletePlaylist(c *gin.Context) {
	var req IDsRequest
	if c.Bind(&req) != nil {
		NewBadRequestError(nil).Abort(c)
	}

	kcId := c.MustGet("KC_ID").(string)

	var errTx error
	tx, err := boil.Begin()
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	ps, err := models.Playlists(
		qm.Load("PlaylistItems"),
		qm.Where("account_id = ?", kcId),
		qm.WhereIn("id in ?", utils.ConvertArgsInt64(req.IDs)...),
	).All(tx)
	if err != nil {
		NewInternalError(err).Abort(c)
		tx.Rollback()
	}
	for _, p := range ps {
		_, errTx = p.R.PlaylistItems.DeleteAll(tx)
		if errTx != nil {
			NewInternalError(err).Abort(c)
			tx.Rollback()
		}
	}

	resp, errTx := ps.DeleteAll(tx)
	if errTx != nil {
		NewInternalError(err).Abort(c)
		tx.Rollback()
	}

	err = tx.Commit()
	concludeRequest(c, resp, NewInternalError(err))
}

func (a *App) handleGetPlaylist(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil {
		NewBadRequestError(err).Abort(c)
	}

	kcId := c.MustGet("KC_ID").(string)

	pl, err := models.Playlists(
		qm.Where("account_id = ? AND id = ?", kcId, id),
		qm.Load("PlaylistItems"),
	).One(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	sort.SliceStable(pl.R.PlaylistItems, func(i int, j int) bool {
		return pl.R.PlaylistItems[i].Position > pl.R.PlaylistItems[j].Position
	})

	concludeRequest(c, playlistResponse{Playlist: pl, PlaylistItems: pl.R.PlaylistItems}, nil)
}

func (a *App) handleGetPlaylistItems(c *gin.Context) {
	kcId := c.MustGet("KC_ID").(string)

	var req playListItemRequest
	if err := c.Bind(&req); err != nil {
		NewBadRequestError(err).Abort(c)
	}

	mods := []qm.QueryMod{
		qm.InnerJoin("playlist pl ON pl.id = playlist_id"),
		qm.Where("pl.account_id = ?", kcId),
		qm.OrderBy("position DESC"),
	}

	if req.PlayListIds != nil {
		mods = append(mods, qm.WhereIn("pl.id IN ?", req.PlayListIds))
	}

	if len(req.UIDs) > 0 {
		mods = append(mods, qm.WhereIn("content_unit_uid IN ?", utils.ConvertArgsString(req.UIDs)...))
	}

	plis, err := models.PlaylistItems(mods...).All(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	concludeRequest(c, playlistItemResponse{PlaylistItems: plis}, nil)
}

func (a *App) handleAddToPlaylist(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil {
		NewBadRequestError(err).Abort(c)
	}

	kcId := c.MustGet("KC_ID").(string)

	pl, err := models.Playlists(qm.Load("PlaylistItems"), qm.Where("id = ?", id)).One(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	if kcId != pl.AccountID {
		err := errors.New("not acceptable")
		NewHttpError(http.StatusNotAcceptable, err, gin.ErrorTypePrivate).Abort(c)
	}

	var req UIDsRequest
	if err := c.Bind(&req); err != nil {
		NewBadRequestError(err).Abort(c)
	}
	hasUnit := false
	for _, x := range pl.R.PlaylistItems {
		for _, nuid := range req.UIDs {
			if x.ContentUnitUID == nuid {
				hasUnit = true
				break
			}
		}
	}
	if hasUnit {
		NewBadRequestError(errors.New("has unit on playlist")).Abort(c)
	}

	for _, nuid := range req.UIDs {
		item := models.PlaylistItem{PlaylistID: id, ContentUnitUID: nuid}
		if err := item.Insert(a.DB, boil.Infer()); err != nil {
			NewInternalError(err).Abort(c)
		}
	}
	err = pl.R.PlaylistItems.ReloadAll(a.DB)
	concludeRequest(c, pl.R.PlaylistItems, NewInternalError(err))
}

func (a *App) handleUpdatePlaylistItems(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil {
		NewBadRequestError(err).Abort(c)
	}

	kcId := c.MustGet("KC_ID").(string)

	pi, err := models.PlaylistItems(
		qm.InnerJoin("playlist pl ON pl.id = playlist_id"),
		qm.Where("\"playlist_item\".id = ? AND pl.account_id = ?", id, kcId),
	).One(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	var req models.PlaylistItem
	if err := c.Bind(&req); err != nil {
		NewBadRequestError(err).Abort(c)
	}
	if pi.Position != req.Position {
		pi.Position = req.Position
		_, err = pi.Update(a.DB, boil.Whitelist("position"))
		if err != nil {
			NewInternalError(err).Abort(c)
		}
		err = pi.Reload(a.DB)
	}
	concludeRequest(c, pi, NewInternalError(err))
}

func (a *App) handleDeleteFromPlaylist(c *gin.Context) {
	kcId := c.MustGet("KC_ID").(string)

	mods := []qm.QueryMod{
		qm.InnerJoin("playlist pl ON  pl.id = \"playlist_item\".playlist_id"),
		qm.Where("pl.account_id = ?", kcId),
	}

	var req DeletePlaylistItemRequest
	if err := c.Bind(&req); err != nil {
		NewBadRequestError(err).Abort(c)
	}
	if len(req.UIDs) > 0 && req.PlaylistId > 0 {
		mods = append(mods, qm.WhereIn("\"playlist_item\".content_unit_uid IN ?", utils.ConvertArgsString(req.UIDs)...), qm.Where("pl.id = ?", req.PlaylistId))
	}

	if len(req.IDs) > 0 {
		mods = append(mods, qm.WhereIn("\"playlist_item\".id IN ?", utils.ConvertArgsInt64(req.IDs)...))
	}

	plis, err := models.PlaylistItems(mods...).All(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}
	resp, err := plis.DeleteAll(a.DB)
	concludeRequest(c, resp, NewInternalError(err))
}

//Likes handlers
func (a *App) handleGetLikes(c *gin.Context) {
	kcId := c.MustGet("KC_ID").(string)
	var list ListRequest
	if err := c.Bind(&list); err != nil {
		NewBadRequestError(err).Abort(c)
	}

	mods := []qm.QueryMod{qm.Where("account_id = ?", kcId)}
	appendByCUMods(&mods, c)

	total, err := models.Likes(mods...).Count(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	if err := appendMyListMods(&mods, list); err != nil {
		NewInternalError(err).Abort(c)
	}

	ls, err := models.Likes(mods...).All(a.DB)

	resp := likesResponse{
		Likes:        ls,
		ListResponse: ListResponse{Total: total},
	}
	concludeRequest(c, resp, NewInternalError(err))
}

func (a *App) handleAddLikes(c *gin.Context) {
	kcId := c.MustGet("KC_ID").(string)
	var req UIDsRequest
	if c.Bind(&req) != nil {
		NewBadRequestError(nil).Abort(c)
	}

	var likes []*models.Like
	tx, err := openTransaction(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}
	for _, uid := range req.UIDs {
		l := models.Like{
			AccountID:      kcId,
			ContentUnitUID: uid,
		}
		err = l.Insert(tx, boil.Infer())
		if err != nil {
			break
		}
		likes = append(likes, &l)
	}
	closeTransaction(tx, err)
	concludeRequest(c, likes, NewInternalError(err))
}

func (a *App) handleRemoveLikes(c *gin.Context) {
	kcId := c.MustGet("KC_ID").(string)

	var ids IDsRequest
	if c.Bind(&ids) != nil {
		NewBadRequestError(nil).Abort(c)
	}
	ls, err := models.Likes(
		qm.WhereIn("id in ?", utils.ConvertArgsInt64(ids.IDs)...),
		qm.Where("account_id = ?", kcId),
	).All(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	resp, err := ls.DeleteAll(a.DB)
	concludeRequest(c, resp, NewInternalError(err))
}

func (a *App) handleLikeCount(c *gin.Context) {
	var req UIDsRequest
	if c.Bind(&req) != nil {
		NewBadRequestError(nil).Abort(c)
	}

	var mods []qm.QueryMod
	if len(req.UIDs) > 0 {
		mods = append(mods, qm.WhereIn("content_unit_uid in ?", utils.ConvertArgsString(req.UIDs)...))
	}

	count, err := models.Likes(mods...).Count(a.DB)
	concludeRequest(c, count, NewInternalError(err))
}

//Subscriptions handlers
func (a *App) handleGetSubscriptions(c *gin.Context) {
	kcId := c.MustGet("KC_ID").(string)
	mods := []qm.QueryMod{qm.Where("account_id = ?", kcId)}

	total, err := models.Subscriptions(mods...).Count(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}
	var list ListRequest
	if c.Bind(&list) != nil {
		NewBadRequestError(nil).Abort(c)
	}

	if list.OrderBy == "" {
		list.OrderBy = "updated_at DESC"
	}
	if err := appendMyListMods(&mods, list); err != nil {
		NewInternalError(err).Abort(c)
	}
	subs, err := models.Subscriptions(mods...).All(a.DB)

	resp := subscriptionsResponse{
		Subscriptions: subs,
		ListResponse:  ListResponse{Total: total},
	}
	concludeRequest(c, resp, NewInternalError(err))
}

func (a *App) handleSubscribe(c *gin.Context) {
	kcId := c.MustGet("KC_ID").(string)
	var req subscribeRequest
	if c.Bind(&req) != nil {
		NewBadRequestError(nil).Abort(c)
	}
	var subs []*models.Subscription
	for _, uid := range req.Collections {
		if uid == "" {
			continue
		}
		s := models.Subscription{
			AccountID:      kcId,
			CollectionUID:  null.String{String: uid, Valid: true},
			ContentUnitUID: req.ContentUnitUID,
		}
		subs = append(subs, &s)
	}

	for _, t := range req.ContentTypes {
		if t == "" {
			continue
		}
		s := models.Subscription{
			AccountID:      kcId,
			ContentType:    null.String{String: t, Valid: true},
			ContentUnitUID: req.ContentUnitUID,
		}
		subs = append(subs, &s)
	}
	if len(subs) == 0 {
		NewBadRequestError(errors.New("must add or collection or CU type for subscribe")).Abort(c)
	}

	tx, err := openTransaction(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}
	for _, s := range subs {
		err = s.Insert(tx, boil.Infer())
		if err != nil {
			break
		}
	}
	closeTransaction(tx, err)
	concludeRequest(c, subs, NewInternalError(err))
}

func (a *App) handleUnsubscribe(c *gin.Context) {

	kcId := c.MustGet("KC_ID").(string)
	var req IDsRequest
	if c.Bind(&req) != nil {
		NewBadRequestError(nil).Abort(c)
	}
	subs, err := models.Subscriptions(
		qm.WhereIn("id in ?", utils.ConvertArgsInt64(req.IDs)...),
		qm.Where("account_id = ?", kcId),
	).All(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}
	resp, err := subs.DeleteAll(a.DB)
	concludeRequest(c, resp, NewInternalError(err))
}

//History handlers
func (a *App) handleGetHistory(c *gin.Context) {
	kcId := c.MustGet("KC_ID").(string)
	var ids []int64
	if c.Bind(&ids) != nil {
		NewBadRequestError(nil).Abort(c)
	}

	mods := []qm.QueryMod{qm.Where("account_id = ?", kcId)}

	total, err := models.Histories(mods...).Count(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	var list ListRequest
	if c.Bind(&list) != nil {
		NewBadRequestError(nil).Abort(c)
	}
	if err := appendMyListMods(&mods, list); err != nil {
		NewInternalError(err).Abort(c)
	}
	history, err := models.Histories(mods...).All(a.DB)

	res := historyResponse{
		History:      history,
		ListResponse: ListResponse{Total: total},
	}
	concludeRequest(c, res, NewInternalError(err))
}

func (a *App) handleDeleteHistory(c *gin.Context) {
	kcId := c.MustGet("KC_ID").(string)
	var ids IDsRequest
	if c.Bind(&ids) != nil {
		NewBadRequestError(nil).Abort(c)
	}

	subs, err := models.Histories(
		qm.WhereIn("id in ?", utils.ConvertArgsInt64(ids.IDs)...),
		qm.Where("account_id = ?", kcId),
	).All(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}
	resp, err := subs.DeleteAll(a.DB)
	concludeRequest(c, resp, NewInternalError(err))
}

/* HELPERS */

func openTransaction(db *sql.DB) (*sql.Tx, error) {
	log.Info("open transaction")
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func closeTransaction(tx *sql.Tx, err error) {
	log.Info("close transaction")
	if err == nil {
		tx.Commit()
	} else {
		tx.Rollback()
	}
}

func appendMyListMods(mods *[]qm.QueryMod, r ListRequest) error {
	// group to remove duplicates
	*mods = append(*mods, qm.GroupBy("id"))

	if r.OrderBy != "" {
		*mods = append(*mods, qm.OrderBy(r.OrderBy))
	} else {
		*mods = append(*mods, qm.OrderBy("created_at DESC"))
	}

	var limit, offset int

	// pagination style
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

	return nil
}

func appendByCUMods(mods *[]qm.QueryMod, c *gin.Context) {
	var req UIDsRequest
	if err := c.Bind(&req); err != nil {
		log.Info("request without content units uids")
		return
	}
	if len(req.UIDs) > 0 {
		*mods = append(*mods, qm.WhereIn("content_unit_uid in ?", utils.ConvertArgsString(req.UIDs)...))
	}
}

func buildNewParams(newp, oldp *models.Playlist) ([]byte, error) {
	if !newp.Parameters.Valid {
		return nil, nil
	}

	var nParams map[string]interface{}
	if err := newp.Parameters.Unmarshal(&nParams); err != nil {
		return nil, NewBadRequestError(err)
	}
	if len(nParams) == 0 {
		return nil, nil
	}

	var params map[string]interface{}
	if oldp.Parameters.Valid {
		if err := oldp.Parameters.Unmarshal(&params); err != nil {
			return nil, err
		}

		for k, v := range nParams {
			params[k] = v
		}
	} else {
		params = nParams
	}

	fpa, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	return fpa, nil
}

func concludeRequest(c *gin.Context, resp interface{}, err *HttpError) {
	if err == nil {
		c.JSON(http.StatusOK, resp)
	} else {
		err.Abort(c)
	}
}
