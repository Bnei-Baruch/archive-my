package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
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

func (a *App) handleGetPlaylists(c *gin.Context) {
	tx, err := openTransaction(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

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
	pl, err := models.Playlists(mods...).All(tx)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	total, err := models.Playlists(qm.Where("account_id = ?", kcId)).Count(tx)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	resp := playlistsResponse{Playlist: pl, ListResponse: ListResponse{Total: total}}

	concludeRequest(c, resp, nil)
}

func (a *App) handleCreatePlaylist(c *gin.Context) {
	tx, err := openTransaction(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

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

	err = pl.Insert(tx, boil.Infer())
	concludeRequest(c, pl, NewInternalError(err))
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
	tx, err := openTransaction(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	p, err := models.Playlists(qm.Where("id = ?", id)).One(tx)
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

	_, err = p.Update(tx, boil.Infer())
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
	_, err = p.Update(tx, boil.Infer())
	if err != nil {
		NewInternalError(err).Abort(c)
	}
}

func (a *App) handleDeletePlaylist(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil {
		NewBadRequestError(err).Abort(c)
	}
	tx, err := openTransaction(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	kcId := c.MustGet("KC_ID").(string)

	p, err := models.Playlists(qm.Where("id = ?", id)).One(tx)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	if kcId != p.AccountID {
		err := errors.New("not acceptable")
		NewHttpError(http.StatusNotAcceptable, err, gin.ErrorTypePrivate).Abort(c)
	}
	_, err = p.Delete(tx)
	concludeRequest(c, p, NewInternalError(err))
}

func (a *App) handleAddToPlaylist(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil {
		NewBadRequestError(err).Abort(c)
	}
	tx, err := openTransaction(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	kcId := c.MustGet("KC_ID").(string)

	pl, err := models.Playlists(qm.Load("PlaylistItems"), qm.Where("id = ?", id)).One(tx)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	if kcId != pl.AccountID {
		err := errors.New("not acceptable")
		NewHttpError(http.StatusNotAcceptable, err, gin.ErrorTypePrivate).Abort(c)
	}

	var uids []string
	if c.Bind(&uids) != nil {
		NewBadRequestError(err).Abort(c)
	}
	hasUnit := false
	for _, x := range pl.R.PlaylistItems {
		for _, nuid := range uids {
			if x.ContentUnitUID == nuid {
				hasUnit = true
				break
			}
		}
	}
	if hasUnit {
		NewBadRequestError(errors.New("has unit on playlist")).Abort(c)
	}

	for _, nuid := range uids {
		item := models.PlaylistItem{PlaylistID: id, ContentUnitUID: nuid}
		if _, err := item.Update(tx, boil.Infer()); err != nil {
			NewInternalError(err).Abort(c)
		}
	}
	err = pl.R.PlaylistItems.ReloadAll(tx)
	concludeRequest(c, pl.R.PlaylistItems, NewInternalError(err))
}

func (a *App) handleDeleteFromPlaylist(c *gin.Context) {
	var ids []int64
	if c.Bind(&ids) != nil {
		NewBadRequestError(nil).Abort(c)
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 0)
	if err != nil {
		NewBadRequestError(err).Abort(c)
	}
	tx, err := openTransaction(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	kcId := c.MustGet("KC_ID").(string)
	plis, err := models.PlaylistItems(
		qm.From("playlist_item as pli"),
		qm.Load("PlaylistItems"),
		qm.InnerJoin("playlist pl ON  pl.id = pli.playlist_id"),
		qm.Where("pl.account_id = ? AND pl.id = ? AND pli.id IN (?)", kcId, id, utils.ConvertArgsInt64(ids)),
	).All(tx)
	if err != nil {
		NewInternalError(err).Abort(c)
	}
	_, err = plis.DeleteAll(tx)
	concludeRequest(c, plis, NewInternalError(err))
}

func (a *App) handleGetLikes(c *gin.Context) {
	kcId := c.MustGet("KC_ID").(string)
	var list ListRequest
	if err := c.Bind(&list); err != nil {
		NewBadRequestError(err).Abort(c)
	}
	tx, err := openTransaction(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	mods := []qm.QueryMod{qm.Where("account_id = ?", kcId)}

	total, err := models.Likes(mods...).Count(tx)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	if err := appendMyListMods(&mods, list); err != nil {
		NewInternalError(err).Abort(c)
	}
	ls, err := models.Likes(mods...).All(tx)

	resp := likesResponse{
		Likes:        ls,
		ListResponse: ListResponse{Total: total},
	}
	concludeRequest(c, resp, NewInternalError(err))
}

func (a *App) handleAddLikes(c *gin.Context) {
	kcId := c.MustGet("KC_ID").(string)
	var uids []string
	if c.Bind(&uids) != nil {
		NewBadRequestError(nil).Abort(c)
	}
	tx, err := openTransaction(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}
	var likes []*models.Like
	for _, uid := range uids {
		l := models.Like{
			AccountID:      kcId,
			ContentUnitUID: uid,
		}
		if err := l.Insert(tx, boil.Infer()); err != nil {
			NewInternalError(err).Abort(c)
		}
		likes = append(likes, &l)
	}
	c.JSON(http.StatusOK, likes)
}

func (a *App) handleRemoveLikes(c *gin.Context) {

	kcId := c.MustGet("KC_ID").(string)
	tx, err := openTransaction(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	var ids []int64
	if c.Bind(&ids) != nil {
		NewBadRequestError(nil).Abort(c)
	}
	ls, err := models.Likes(
		qm.WhereIn("id in (?)", utils.ConvertArgsInt64(ids)...),
		qm.Where("account_id = ?", kcId),
	).All(tx)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	_, err = ls.DeleteAll(tx)
	concludeRequest(c, ls, NewInternalError(err))
}

func (a *App) handleGetSubscriptions(c *gin.Context) {
	kcId := c.MustGet("KC_ID").(string)
	tx, err := openTransaction(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	mods := []qm.QueryMod{qm.Where("account_id = ?", kcId)}

	total, err := models.Subscriptions(mods...).Count(tx)
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
	subs, err := models.Subscriptions(mods...).All(tx)

	res := subscriptionsResponse{
		Subscriptions: subs,
		ListResponse:  ListResponse{Total: total},
	}
	concludeRequest(c, res, NewInternalError(err))
}

func (a *App) handleSubscribe(c *gin.Context) {

	kcId := c.MustGet("KC_ID").(string)
	var uids subscribeRequest
	if c.Bind(&uids) != nil {
		NewBadRequestError(nil).Abort(c)
	}
	tx, err := openTransaction(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	var subs []*models.Subscription
	for _, uid := range uids.Collections {
		s := models.Subscription{
			AccountID:    kcId,
			CollectionID: null.String{String: uid, Valid: true},
		}
		if err := s.Insert(tx, boil.Infer()); err != nil {
			NewInternalError(err).Abort(c)
		}
		subs = append(subs, &s)
	}

	for _, id := range uids.ContentTypes {
		s := models.Subscription{
			AccountID:       kcId,
			ContentUnitType: null.Int64{Int64: id, Valid: true},
		}
		if err := s.Insert(tx, boil.Infer()); err != nil {
			NewInternalError(err).Abort(c)
		}
		subs = append(subs, &s)
	}
	c.JSON(http.StatusOK, subs)
}

func (a *App) handleUnsubscribe(c *gin.Context) {

	kcId := c.MustGet("KC_ID").(string)
	var ids []int64
	if c.Bind(&ids) != nil {
		NewBadRequestError(nil).Abort(c)
	}
	tx, err := openTransaction(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	subs, err := models.Subscriptions(
		qm.WhereIn("id in (?)", utils.ConvertArgsInt64(ids)...),
		qm.Where("account_id = ?", kcId),
	).All(tx)
	if err != nil {
		NewInternalError(err).Abort(c)
	}
	_, err = subs.DeleteAll(tx)
	concludeRequest(c, subs, NewInternalError(err))
}

func (a *App) handleGetHistory(c *gin.Context) {
	kcId := c.MustGet("KC_ID").(string)
	var ids []int64
	if c.Bind(&ids) != nil {
		NewBadRequestError(nil).Abort(c)
	}
	tx, err := openTransaction(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	mods := []qm.QueryMod{qm.Where("account_id = ?", kcId)}

	total, err := models.Histories(mods...).Count(tx)
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
	history, err := models.Histories(mods...).All(tx)

	res := historyResponse{
		History:      history,
		ListResponse: ListResponse{Total: total},
	}
	concludeRequest(c, res, NewInternalError(err))
}

func (a *App) handleDeleteHistory(c *gin.Context) {
	kcId := c.MustGet("KC_ID").(string)
	var ids []int64
	if c.Bind(&ids) != nil {
		NewBadRequestError(nil).Abort(c)
	}
	tx, err := openTransaction(a.DB)
	if err != nil {
		NewInternalError(err).Abort(c)
	}

	subs, err := models.Subscriptions(
		qm.WhereIn("id in (?)", utils.ConvertArgsInt64(ids)...),
		qm.Where("account_id = ?", kcId),
	).All(tx)
	if err != nil {
		NewInternalError(err).Abort(c)
	}
	_, err = subs.DeleteAll(tx)
	concludeRequest(c, subs, NewInternalError(err))
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

func appendMyListMods(mods *[]qm.QueryMod, r ListRequest) error {
	// group to remove duplicates
	*mods = append(*mods, qm.GroupBy("id"))

	if r.OrderBy != "" {
		*mods = append(*mods, qm.OrderBy(r.OrderBy))
	}

	var limit, offset int

	if r.StartIndex == 0 {
		// pagination style
		if r.PageSize == 0 {
			limit = DefaultPageSize
		} else {
			limit = utils.Min(r.PageSize, MaxPageSize)
		}
		if r.PageNumber > 1 {
			offset = (r.PageNumber - 1) * limit
		}
	} else {
		// start & stop index style for "infinite" lists
		offset = r.StartIndex - 1
		if r.StopIndex == 0 {
			limit = MaxPageSize
		} else if r.StopIndex < r.StartIndex {
			return errors.New(fmt.Sprintf("Invalid range [%d-%d]", r.StartIndex, r.StopIndex))
		} else {
			limit = r.StopIndex - r.StartIndex + 1
		}
	}

	*mods = append(*mods, qm.Limit(limit))
	if offset != 0 {
		*mods = append(*mods, qm.Offset(offset))
	}

	return nil
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
