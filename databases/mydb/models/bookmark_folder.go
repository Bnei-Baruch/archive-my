// Code generated by SQLBoiler 4.7.1 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
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

// BookmarkFolder is an object representing the database table.
type BookmarkFolder struct {
	BookmarkID int64    `boil:"bookmark_id" json:"bookmark_id" toml:"bookmark_id" yaml:"bookmark_id"`
	FolderID   int64    `boil:"folder_id" json:"folder_id" toml:"folder_id" yaml:"folder_id"`
	Position   null.Int `boil:"position" json:"position,omitempty" toml:"position" yaml:"position,omitempty"`

	R *bookmarkFolderR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L bookmarkFolderL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var BookmarkFolderColumns = struct {
	BookmarkID string
	FolderID   string
	Position   string
}{
	BookmarkID: "bookmark_id",
	FolderID:   "folder_id",
	Position:   "position",
}

var BookmarkFolderTableColumns = struct {
	BookmarkID string
	FolderID   string
	Position   string
}{
	BookmarkID: "bookmark_folder.bookmark_id",
	FolderID:   "bookmark_folder.folder_id",
	Position:   "bookmark_folder.position",
}

// Generated where

type whereHelperint64 struct{ field string }

func (w whereHelperint64) EQ(x int64) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelperint64) NEQ(x int64) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelperint64) LT(x int64) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelperint64) LTE(x int64) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelperint64) GT(x int64) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelperint64) GTE(x int64) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }
func (w whereHelperint64) IN(slice []int64) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereIn(fmt.Sprintf("%s IN ?", w.field), values...)
}
func (w whereHelperint64) NIN(slice []int64) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereNotIn(fmt.Sprintf("%s NOT IN ?", w.field), values...)
}

type whereHelpernull_Int struct{ field string }

func (w whereHelpernull_Int) EQ(x null.Int) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, false, x)
}
func (w whereHelpernull_Int) NEQ(x null.Int) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, true, x)
}
func (w whereHelpernull_Int) LT(x null.Int) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpernull_Int) LTE(x null.Int) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpernull_Int) GT(x null.Int) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpernull_Int) GTE(x null.Int) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

func (w whereHelpernull_Int) IsNull() qm.QueryMod    { return qmhelper.WhereIsNull(w.field) }
func (w whereHelpernull_Int) IsNotNull() qm.QueryMod { return qmhelper.WhereIsNotNull(w.field) }

var BookmarkFolderWhere = struct {
	BookmarkID whereHelperint64
	FolderID   whereHelperint64
	Position   whereHelpernull_Int
}{
	BookmarkID: whereHelperint64{field: "\"bookmark_folder\".\"bookmark_id\""},
	FolderID:   whereHelperint64{field: "\"bookmark_folder\".\"folder_id\""},
	Position:   whereHelpernull_Int{field: "\"bookmark_folder\".\"position\""},
}

// BookmarkFolderRels is where relationship names are stored.
var BookmarkFolderRels = struct {
	Bookmark string
	Folder   string
}{
	Bookmark: "Bookmark",
	Folder:   "Folder",
}

// bookmarkFolderR is where relationships are stored.
type bookmarkFolderR struct {
	Bookmark *Bookmark `boil:"Bookmark" json:"Bookmark" toml:"Bookmark" yaml:"Bookmark"`
	Folder   *Folder   `boil:"Folder" json:"Folder" toml:"Folder" yaml:"Folder"`
}

// NewStruct creates a new relationship struct
func (*bookmarkFolderR) NewStruct() *bookmarkFolderR {
	return &bookmarkFolderR{}
}

// bookmarkFolderL is where Load methods for each relationship are stored.
type bookmarkFolderL struct{}

var (
	bookmarkFolderAllColumns            = []string{"bookmark_id", "folder_id", "position"}
	bookmarkFolderColumnsWithoutDefault = []string{"bookmark_id", "folder_id", "position"}
	bookmarkFolderColumnsWithDefault    = []string{}
	bookmarkFolderPrimaryKeyColumns     = []string{"folder_id", "bookmark_id"}
)

type (
	// BookmarkFolderSlice is an alias for a slice of pointers to BookmarkFolder.
	// This should almost always be used instead of []BookmarkFolder.
	BookmarkFolderSlice []*BookmarkFolder

	bookmarkFolderQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	bookmarkFolderType                 = reflect.TypeOf(&BookmarkFolder{})
	bookmarkFolderMapping              = queries.MakeStructMapping(bookmarkFolderType)
	bookmarkFolderPrimaryKeyMapping, _ = queries.BindMapping(bookmarkFolderType, bookmarkFolderMapping, bookmarkFolderPrimaryKeyColumns)
	bookmarkFolderInsertCacheMut       sync.RWMutex
	bookmarkFolderInsertCache          = make(map[string]insertCache)
	bookmarkFolderUpdateCacheMut       sync.RWMutex
	bookmarkFolderUpdateCache          = make(map[string]updateCache)
	bookmarkFolderUpsertCacheMut       sync.RWMutex
	bookmarkFolderUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

// One returns a single bookmarkFolder record from the query.
func (q bookmarkFolderQuery) One(exec boil.Executor) (*BookmarkFolder, error) {
	o := &BookmarkFolder{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for bookmark_folder")
	}

	return o, nil
}

// All returns all BookmarkFolder records from the query.
func (q bookmarkFolderQuery) All(exec boil.Executor) (BookmarkFolderSlice, error) {
	var o []*BookmarkFolder

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to BookmarkFolder slice")
	}

	return o, nil
}

// Count returns the count of all BookmarkFolder records in the query.
func (q bookmarkFolderQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count bookmark_folder rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q bookmarkFolderQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if bookmark_folder exists")
	}

	return count > 0, nil
}

// Bookmark pointed to by the foreign key.
func (o *BookmarkFolder) Bookmark(mods ...qm.QueryMod) bookmarkQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.BookmarkID),
	}

	queryMods = append(queryMods, mods...)

	query := Bookmarks(queryMods...)
	queries.SetFrom(query.Query, "\"bookmarks\"")

	return query
}

// Folder pointed to by the foreign key.
func (o *BookmarkFolder) Folder(mods ...qm.QueryMod) folderQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.FolderID),
	}

	queryMods = append(queryMods, mods...)

	query := Folders(queryMods...)
	queries.SetFrom(query.Query, "\"folders\"")

	return query
}

// LoadBookmark allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (bookmarkFolderL) LoadBookmark(e boil.Executor, singular bool, maybeBookmarkFolder interface{}, mods queries.Applicator) error {
	var slice []*BookmarkFolder
	var object *BookmarkFolder

	if singular {
		object = maybeBookmarkFolder.(*BookmarkFolder)
	} else {
		slice = *maybeBookmarkFolder.(*[]*BookmarkFolder)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &bookmarkFolderR{}
		}
		args = append(args, object.BookmarkID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &bookmarkFolderR{}
			}

			for _, a := range args {
				if a == obj.BookmarkID {
					continue Outer
				}
			}

			args = append(args, obj.BookmarkID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`bookmarks`),
		qm.WhereIn(`bookmarks.id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Bookmark")
	}

	var resultSlice []*Bookmark
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Bookmark")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for bookmarks")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for bookmarks")
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.Bookmark = foreign
		if foreign.R == nil {
			foreign.R = &bookmarkR{}
		}
		foreign.R.BookmarkFolders = append(foreign.R.BookmarkFolders, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.BookmarkID == foreign.ID {
				local.R.Bookmark = foreign
				if foreign.R == nil {
					foreign.R = &bookmarkR{}
				}
				foreign.R.BookmarkFolders = append(foreign.R.BookmarkFolders, local)
				break
			}
		}
	}

	return nil
}

// LoadFolder allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (bookmarkFolderL) LoadFolder(e boil.Executor, singular bool, maybeBookmarkFolder interface{}, mods queries.Applicator) error {
	var slice []*BookmarkFolder
	var object *BookmarkFolder

	if singular {
		object = maybeBookmarkFolder.(*BookmarkFolder)
	} else {
		slice = *maybeBookmarkFolder.(*[]*BookmarkFolder)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &bookmarkFolderR{}
		}
		args = append(args, object.FolderID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &bookmarkFolderR{}
			}

			for _, a := range args {
				if a == obj.FolderID {
					continue Outer
				}
			}

			args = append(args, obj.FolderID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`folders`),
		qm.WhereIn(`folders.id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Folder")
	}

	var resultSlice []*Folder
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Folder")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for folders")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for folders")
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.Folder = foreign
		if foreign.R == nil {
			foreign.R = &folderR{}
		}
		foreign.R.BookmarkFolders = append(foreign.R.BookmarkFolders, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.FolderID == foreign.ID {
				local.R.Folder = foreign
				if foreign.R == nil {
					foreign.R = &folderR{}
				}
				foreign.R.BookmarkFolders = append(foreign.R.BookmarkFolders, local)
				break
			}
		}
	}

	return nil
}

// SetBookmark of the bookmarkFolder to the related item.
// Sets o.R.Bookmark to related.
// Adds o to related.R.BookmarkFolders.
func (o *BookmarkFolder) SetBookmark(exec boil.Executor, insert bool, related *Bookmark) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"bookmark_folder\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"bookmark_id"}),
		strmangle.WhereClause("\"", "\"", 2, bookmarkFolderPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.FolderID, o.BookmarkID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.BookmarkID = related.ID
	if o.R == nil {
		o.R = &bookmarkFolderR{
			Bookmark: related,
		}
	} else {
		o.R.Bookmark = related
	}

	if related.R == nil {
		related.R = &bookmarkR{
			BookmarkFolders: BookmarkFolderSlice{o},
		}
	} else {
		related.R.BookmarkFolders = append(related.R.BookmarkFolders, o)
	}

	return nil
}

// SetFolder of the bookmarkFolder to the related item.
// Sets o.R.Folder to related.
// Adds o to related.R.BookmarkFolders.
func (o *BookmarkFolder) SetFolder(exec boil.Executor, insert bool, related *Folder) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"bookmark_folder\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"folder_id"}),
		strmangle.WhereClause("\"", "\"", 2, bookmarkFolderPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.FolderID, o.BookmarkID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.FolderID = related.ID
	if o.R == nil {
		o.R = &bookmarkFolderR{
			Folder: related,
		}
	} else {
		o.R.Folder = related
	}

	if related.R == nil {
		related.R = &folderR{
			BookmarkFolders: BookmarkFolderSlice{o},
		}
	} else {
		related.R.BookmarkFolders = append(related.R.BookmarkFolders, o)
	}

	return nil
}

// BookmarkFolders retrieves all the records using an executor.
func BookmarkFolders(mods ...qm.QueryMod) bookmarkFolderQuery {
	mods = append(mods, qm.From("\"bookmark_folder\""))
	return bookmarkFolderQuery{NewQuery(mods...)}
}

// FindBookmarkFolder retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindBookmarkFolder(exec boil.Executor, folderID int64, bookmarkID int64, selectCols ...string) (*BookmarkFolder, error) {
	bookmarkFolderObj := &BookmarkFolder{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"bookmark_folder\" where \"folder_id\"=$1 AND \"bookmark_id\"=$2", sel,
	)

	q := queries.Raw(query, folderID, bookmarkID)

	err := q.Bind(nil, exec, bookmarkFolderObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from bookmark_folder")
	}

	return bookmarkFolderObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *BookmarkFolder) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no bookmark_folder provided for insertion")
	}

	var err error

	nzDefaults := queries.NonZeroDefaultSet(bookmarkFolderColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	bookmarkFolderInsertCacheMut.RLock()
	cache, cached := bookmarkFolderInsertCache[key]
	bookmarkFolderInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			bookmarkFolderAllColumns,
			bookmarkFolderColumnsWithDefault,
			bookmarkFolderColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(bookmarkFolderType, bookmarkFolderMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(bookmarkFolderType, bookmarkFolderMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"bookmark_folder\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"bookmark_folder\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "models: unable to insert into bookmark_folder")
	}

	if !cached {
		bookmarkFolderInsertCacheMut.Lock()
		bookmarkFolderInsertCache[key] = cache
		bookmarkFolderInsertCacheMut.Unlock()
	}

	return nil
}

// Update uses an executor to update the BookmarkFolder.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *BookmarkFolder) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	key := makeCacheKey(columns, nil)
	bookmarkFolderUpdateCacheMut.RLock()
	cache, cached := bookmarkFolderUpdateCache[key]
	bookmarkFolderUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			bookmarkFolderAllColumns,
			bookmarkFolderPrimaryKeyColumns,
		)

		if len(wl) == 0 {
			return 0, errors.New("models: unable to update bookmark_folder, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"bookmark_folder\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, bookmarkFolderPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(bookmarkFolderType, bookmarkFolderMapping, append(wl, bookmarkFolderPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update bookmark_folder row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for bookmark_folder")
	}

	if !cached {
		bookmarkFolderUpdateCacheMut.Lock()
		bookmarkFolderUpdateCache[key] = cache
		bookmarkFolderUpdateCacheMut.Unlock()
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values.
func (q bookmarkFolderQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for bookmark_folder")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for bookmark_folder")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o BookmarkFolderSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), bookmarkFolderPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"bookmark_folder\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, bookmarkFolderPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in bookmarkFolder slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all bookmarkFolder")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *BookmarkFolder) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no bookmark_folder provided for upsert")
	}

	nzDefaults := queries.NonZeroDefaultSet(bookmarkFolderColumnsWithDefault, o)

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

	bookmarkFolderUpsertCacheMut.RLock()
	cache, cached := bookmarkFolderUpsertCache[key]
	bookmarkFolderUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			bookmarkFolderAllColumns,
			bookmarkFolderColumnsWithDefault,
			bookmarkFolderColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			bookmarkFolderAllColumns,
			bookmarkFolderPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert bookmark_folder, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(bookmarkFolderPrimaryKeyColumns))
			copy(conflict, bookmarkFolderPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"bookmark_folder\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(bookmarkFolderType, bookmarkFolderMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(bookmarkFolderType, bookmarkFolderMapping, ret)
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
		return errors.Wrap(err, "models: unable to upsert bookmark_folder")
	}

	if !cached {
		bookmarkFolderUpsertCacheMut.Lock()
		bookmarkFolderUpsertCache[key] = cache
		bookmarkFolderUpsertCacheMut.Unlock()
	}

	return nil
}

// Delete deletes a single BookmarkFolder record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *BookmarkFolder) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no BookmarkFolder provided for delete")
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), bookmarkFolderPrimaryKeyMapping)
	sql := "DELETE FROM \"bookmark_folder\" WHERE \"folder_id\"=$1 AND \"bookmark_id\"=$2"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from bookmark_folder")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for bookmark_folder")
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q bookmarkFolderQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no bookmarkFolderQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from bookmark_folder")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for bookmark_folder")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o BookmarkFolderSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), bookmarkFolderPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"bookmark_folder\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, bookmarkFolderPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from bookmarkFolder slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for bookmark_folder")
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *BookmarkFolder) Reload(exec boil.Executor) error {
	ret, err := FindBookmarkFolder(exec, o.FolderID, o.BookmarkID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *BookmarkFolderSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := BookmarkFolderSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), bookmarkFolderPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"bookmark_folder\".* FROM \"bookmark_folder\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, bookmarkFolderPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in BookmarkFolderSlice")
	}

	*o = slice

	return nil
}

// BookmarkFolderExists checks if the BookmarkFolder row exists.
func BookmarkFolderExists(exec boil.Executor, folderID int64, bookmarkID int64) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"bookmark_folder\" where \"folder_id\"=$1 AND \"bookmark_id\"=$2 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, folderID, bookmarkID)
	}
	row := exec.QueryRow(sql, folderID, bookmarkID)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if bookmark_folder exists")
	}

	return exists, nil
}