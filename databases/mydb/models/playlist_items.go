// Code generated by SQLBoiler 4.8.6 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// PlaylistItem is an object representing the database table.
type PlaylistItem struct {
	ID             int64       `boil:"id" json:"id" toml:"id" yaml:"id"`
	PlaylistID     int64       `boil:"playlist_id" json:"playlist_id" toml:"playlist_id" yaml:"playlist_id"`
	Position       int         `boil:"position" json:"position" toml:"position" yaml:"position"`
	ContentUnitUID string      `boil:"content_unit_uid" json:"content_unit_uid" toml:"content_unit_uid" yaml:"content_unit_uid"`
	Name           null.String `boil:"name" json:"name,omitempty" toml:"name" yaml:"name,omitempty"`
	Properties     null.JSON   `boil:"properties" json:"properties,omitempty" toml:"properties" yaml:"properties,omitempty"`

	R *playlistItemR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L playlistItemL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var PlaylistItemColumns = struct {
	ID             string
	PlaylistID     string
	Position       string
	ContentUnitUID string
	Name           string
	Properties     string
}{
	ID:             "id",
	PlaylistID:     "playlist_id",
	Position:       "position",
	ContentUnitUID: "content_unit_uid",
	Name:           "name",
	Properties:     "properties",
}

var PlaylistItemTableColumns = struct {
	ID             string
	PlaylistID     string
	Position       string
	ContentUnitUID string
	Name           string
	Properties     string
}{
	ID:             "playlist_items.id",
	PlaylistID:     "playlist_items.playlist_id",
	Position:       "playlist_items.position",
	ContentUnitUID: "playlist_items.content_unit_uid",
	Name:           "playlist_items.name",
	Properties:     "playlist_items.properties",
}

// Generated where

type whereHelperint struct{ field string }

func (w whereHelperint) EQ(x int) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelperint) NEQ(x int) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelperint) LT(x int) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelperint) LTE(x int) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelperint) GT(x int) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelperint) GTE(x int) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }
func (w whereHelperint) IN(slice []int) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereIn(fmt.Sprintf("%s IN ?", w.field), values...)
}
func (w whereHelperint) NIN(slice []int) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereNotIn(fmt.Sprintf("%s NOT IN ?", w.field), values...)
}

var PlaylistItemWhere = struct {
	ID             whereHelperint64
	PlaylistID     whereHelperint64
	Position       whereHelperint
	ContentUnitUID whereHelperstring
	Name           whereHelpernull_String
	Properties     whereHelpernull_JSON
}{
	ID:             whereHelperint64{field: "\"playlist_items\".\"id\""},
	PlaylistID:     whereHelperint64{field: "\"playlist_items\".\"playlist_id\""},
	Position:       whereHelperint{field: "\"playlist_items\".\"position\""},
	ContentUnitUID: whereHelperstring{field: "\"playlist_items\".\"content_unit_uid\""},
	Name:           whereHelpernull_String{field: "\"playlist_items\".\"name\""},
	Properties:     whereHelpernull_JSON{field: "\"playlist_items\".\"properties\""},
}

// PlaylistItemRels is where relationship names are stored.
var PlaylistItemRels = struct {
	Playlist string
}{
	Playlist: "Playlist",
}

// playlistItemR is where relationships are stored.
type playlistItemR struct {
	Playlist *Playlist `boil:"Playlist" json:"Playlist" toml:"Playlist" yaml:"Playlist"`
}

// NewStruct creates a new relationship struct
func (*playlistItemR) NewStruct() *playlistItemR {
	return &playlistItemR{}
}

// playlistItemL is where Load methods for each relationship are stored.
type playlistItemL struct{}

var (
	playlistItemAllColumns            = []string{"id", "playlist_id", "position", "content_unit_uid", "name", "properties"}
	playlistItemColumnsWithoutDefault = []string{"playlist_id", "position", "content_unit_uid"}
	playlistItemColumnsWithDefault    = []string{"id", "name", "properties"}
	playlistItemPrimaryKeyColumns     = []string{"id"}
	playlistItemGeneratedColumns      = []string{}
)

type (
	// PlaylistItemSlice is an alias for a slice of pointers to PlaylistItem.
	// This should almost always be used instead of []PlaylistItem.
	PlaylistItemSlice []*PlaylistItem

	playlistItemQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	playlistItemType                 = reflect.TypeOf(&PlaylistItem{})
	playlistItemMapping              = queries.MakeStructMapping(playlistItemType)
	playlistItemPrimaryKeyMapping, _ = queries.BindMapping(playlistItemType, playlistItemMapping, playlistItemPrimaryKeyColumns)
	playlistItemInsertCacheMut       sync.RWMutex
	playlistItemInsertCache          = make(map[string]insertCache)
	playlistItemUpdateCacheMut       sync.RWMutex
	playlistItemUpdateCache          = make(map[string]updateCache)
	playlistItemUpsertCacheMut       sync.RWMutex
	playlistItemUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

// One returns a single playlistItem record from the query.
func (q playlistItemQuery) One(exec boil.Executor) (*PlaylistItem, error) {
	o := &PlaylistItem{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for playlist_items")
	}

	return o, nil
}

// All returns all PlaylistItem records from the query.
func (q playlistItemQuery) All(exec boil.Executor) (PlaylistItemSlice, error) {
	var o []*PlaylistItem

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to PlaylistItem slice")
	}

	return o, nil
}

// Count returns the count of all PlaylistItem records in the query.
func (q playlistItemQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count playlist_items rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q playlistItemQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if playlist_items exists")
	}

	return count > 0, nil
}

// Playlist pointed to by the foreign key.
func (o *PlaylistItem) Playlist(mods ...qm.QueryMod) playlistQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.PlaylistID),
	}

	queryMods = append(queryMods, mods...)

	query := Playlists(queryMods...)
	queries.SetFrom(query.Query, "\"playlists\"")

	return query
}

// LoadPlaylist allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (playlistItemL) LoadPlaylist(e boil.Executor, singular bool, maybePlaylistItem interface{}, mods queries.Applicator) error {
	var slice []*PlaylistItem
	var object *PlaylistItem

	if singular {
		object = maybePlaylistItem.(*PlaylistItem)
	} else {
		slice = *maybePlaylistItem.(*[]*PlaylistItem)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &playlistItemR{}
		}
		args = append(args, object.PlaylistID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &playlistItemR{}
			}

			for _, a := range args {
				if a == obj.PlaylistID {
					continue Outer
				}
			}

			args = append(args, obj.PlaylistID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`playlists`),
		qm.WhereIn(`playlists.id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Playlist")
	}

	var resultSlice []*Playlist
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Playlist")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for playlists")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for playlists")
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.Playlist = foreign
		if foreign.R == nil {
			foreign.R = &playlistR{}
		}
		foreign.R.PlaylistItems = append(foreign.R.PlaylistItems, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.PlaylistID == foreign.ID {
				local.R.Playlist = foreign
				if foreign.R == nil {
					foreign.R = &playlistR{}
				}
				foreign.R.PlaylistItems = append(foreign.R.PlaylistItems, local)
				break
			}
		}
	}

	return nil
}

// SetPlaylist of the playlistItem to the related item.
// Sets o.R.Playlist to related.
// Adds o to related.R.PlaylistItems.
func (o *PlaylistItem) SetPlaylist(exec boil.Executor, insert bool, related *Playlist) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"playlist_items\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"playlist_id"}),
		strmangle.WhereClause("\"", "\"", 2, playlistItemPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.PlaylistID = related.ID
	if o.R == nil {
		o.R = &playlistItemR{
			Playlist: related,
		}
	} else {
		o.R.Playlist = related
	}

	if related.R == nil {
		related.R = &playlistR{
			PlaylistItems: PlaylistItemSlice{o},
		}
	} else {
		related.R.PlaylistItems = append(related.R.PlaylistItems, o)
	}

	return nil
}

// PlaylistItems retrieves all the records using an executor.
func PlaylistItems(mods ...qm.QueryMod) playlistItemQuery {
	mods = append(mods, qm.From("\"playlist_items\""))
	return playlistItemQuery{NewQuery(mods...)}
}

// FindPlaylistItem retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindPlaylistItem(exec boil.Executor, iD int64, selectCols ...string) (*PlaylistItem, error) {
	playlistItemObj := &PlaylistItem{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"playlist_items\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, playlistItemObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from playlist_items")
	}

	return playlistItemObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *PlaylistItem) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no playlist_items provided for insertion")
	}

	var err error

	nzDefaults := queries.NonZeroDefaultSet(playlistItemColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	playlistItemInsertCacheMut.RLock()
	cache, cached := playlistItemInsertCache[key]
	playlistItemInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			playlistItemAllColumns,
			playlistItemColumnsWithDefault,
			playlistItemColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(playlistItemType, playlistItemMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(playlistItemType, playlistItemMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"playlist_items\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"playlist_items\" %sDEFAULT VALUES%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			queryReturning = fmt.Sprintf(" RETURNING \"%s\"", strings.Join(returnColumns, "\",\""))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, vals)
	}

	if len(cache.retMapping) != 0 {
		err = exec.QueryRow(cache.query, vals...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	} else {
		_, err = exec.Exec(cache.query, vals...)
	}

	if err != nil {
		return errors.Wrap(err, "models: unable to insert into playlist_items")
	}

	if !cached {
		playlistItemInsertCacheMut.Lock()
		playlistItemInsertCache[key] = cache
		playlistItemInsertCacheMut.Unlock()
	}

	return nil
}

// Update uses an executor to update the PlaylistItem.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *PlaylistItem) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	key := makeCacheKey(columns, nil)
	playlistItemUpdateCacheMut.RLock()
	cache, cached := playlistItemUpdateCache[key]
	playlistItemUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			playlistItemAllColumns,
			playlistItemPrimaryKeyColumns,
		)
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update playlist_items, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"playlist_items\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, playlistItemPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(playlistItemType, playlistItemMapping, append(wl, playlistItemPrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	var result sql.Result
	result, err = exec.Exec(cache.query, values...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update playlist_items row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for playlist_items")
	}

	if !cached {
		playlistItemUpdateCacheMut.Lock()
		playlistItemUpdateCache[key] = cache
		playlistItemUpdateCacheMut.Unlock()
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values.
func (q playlistItemQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for playlist_items")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for playlist_items")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o PlaylistItemSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("models: update all requires at least one column argument")
	}

	colNames := make([]string, len(cols))
	args := make([]interface{}, len(cols))

	i := 0
	for name, value := range cols {
		colNames[i] = name
		args[i] = value
		i++
	}

	// Append all of the primary key values for each column
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), playlistItemPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"playlist_items\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, playlistItemPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in playlistItem slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all playlistItem")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *PlaylistItem) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no playlist_items provided for upsert")
	}

	nzDefaults := queries.NonZeroDefaultSet(playlistItemColumnsWithDefault, o)

	// Build cache key in-line uglily - mysql vs psql problems
	buf := strmangle.GetBuffer()
	if updateOnConflict {
		buf.WriteByte('t')
	} else {
		buf.WriteByte('f')
	}
	buf.WriteByte('.')
	for _, c := range conflictColumns {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(updateColumns.Kind))
	for _, c := range updateColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(insertColumns.Kind))
	for _, c := range insertColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzDefaults {
		buf.WriteString(c)
	}
	key := buf.String()
	strmangle.PutBuffer(buf)

	playlistItemUpsertCacheMut.RLock()
	cache, cached := playlistItemUpsertCache[key]
	playlistItemUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			playlistItemAllColumns,
			playlistItemColumnsWithDefault,
			playlistItemColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			playlistItemAllColumns,
			playlistItemPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert playlist_items, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(playlistItemPrimaryKeyColumns))
			copy(conflict, playlistItemPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"playlist_items\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(playlistItemType, playlistItemMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(playlistItemType, playlistItemMapping, ret)
			if err != nil {
				return err
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)
	var returns []interface{}
	if len(cache.retMapping) != 0 {
		returns = queries.PtrsFromMapping(value, cache.retMapping)
	}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, vals)
	}
	if len(cache.retMapping) != 0 {
		err = exec.QueryRow(cache.query, vals...).Scan(returns...)
		if err == sql.ErrNoRows {
			err = nil // Postgres doesn't return anything when there's no update
		}
	} else {
		_, err = exec.Exec(cache.query, vals...)
	}
	if err != nil {
		return errors.Wrap(err, "models: unable to upsert playlist_items")
	}

	if !cached {
		playlistItemUpsertCacheMut.Lock()
		playlistItemUpsertCache[key] = cache
		playlistItemUpsertCacheMut.Unlock()
	}

	return nil
}

// Delete deletes a single PlaylistItem record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *PlaylistItem) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no PlaylistItem provided for delete")
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), playlistItemPrimaryKeyMapping)
	sql := "DELETE FROM \"playlist_items\" WHERE \"id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from playlist_items")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for playlist_items")
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q playlistItemQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no playlistItemQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from playlist_items")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for playlist_items")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o PlaylistItemSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), playlistItemPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"playlist_items\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, playlistItemPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from playlistItem slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for playlist_items")
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *PlaylistItem) Reload(exec boil.Executor) error {
	ret, err := FindPlaylistItem(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *PlaylistItemSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := PlaylistItemSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), playlistItemPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"playlist_items\".* FROM \"playlist_items\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, playlistItemPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in PlaylistItemSlice")
	}

	*o = slice

	return nil
}

// PlaylistItemExists checks if the PlaylistItem row exists.
func PlaylistItemExists(exec boil.Executor, iD int64) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"playlist_items\" where \"id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if playlist_items exists")
	}

	return exists, nil
}
