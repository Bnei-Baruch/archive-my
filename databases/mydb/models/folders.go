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

// Folder is an object representing the database table.
type Folder struct {
	ID        int64       `boil:"id" json:"id" toml:"id" yaml:"id"`
	Name      null.String `boil:"name" json:"name,omitempty" toml:"name" yaml:"name,omitempty"`
	UserID    int64       `boil:"user_id" json:"user_id" toml:"user_id" yaml:"user_id"`
	CreatedAt time.Time   `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`

	R *folderR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L folderL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var FolderColumns = struct {
	ID        string
	Name      string
	UserID    string
	CreatedAt string
}{
	ID:        "id",
	Name:      "name",
	UserID:    "user_id",
	CreatedAt: "created_at",
}

var FolderTableColumns = struct {
	ID        string
	Name      string
	UserID    string
	CreatedAt string
}{
	ID:        "folders.id",
	Name:      "folders.name",
	UserID:    "folders.user_id",
	CreatedAt: "folders.created_at",
}

// Generated where

var FolderWhere = struct {
	ID        whereHelperint64
	Name      whereHelpernull_String
	UserID    whereHelperint64
	CreatedAt whereHelpertime_Time
}{
	ID:        whereHelperint64{field: "\"folders\".\"id\""},
	Name:      whereHelpernull_String{field: "\"folders\".\"name\""},
	UserID:    whereHelperint64{field: "\"folders\".\"user_id\""},
	CreatedAt: whereHelpertime_Time{field: "\"folders\".\"created_at\""},
}

// FolderRels is where relationship names are stored.
var FolderRels = struct {
	User            string
	BookmarkFolders string
}{
	User:            "User",
	BookmarkFolders: "BookmarkFolders",
}

// folderR is where relationships are stored.
type folderR struct {
	User            *User               `boil:"User" json:"User" toml:"User" yaml:"User"`
	BookmarkFolders BookmarkFolderSlice `boil:"BookmarkFolders" json:"BookmarkFolders" toml:"BookmarkFolders" yaml:"BookmarkFolders"`
}

// NewStruct creates a new relationship struct
func (*folderR) NewStruct() *folderR {
	return &folderR{}
}

// folderL is where Load methods for each relationship are stored.
type folderL struct{}

var (
	folderAllColumns            = []string{"id", "name", "user_id", "created_at"}
	folderColumnsWithoutDefault = []string{"user_id"}
	folderColumnsWithDefault    = []string{"id", "name", "created_at"}
	folderPrimaryKeyColumns     = []string{"id"}
	folderGeneratedColumns      = []string{}
)

type (
	// FolderSlice is an alias for a slice of pointers to Folder.
	// This should almost always be used instead of []Folder.
	FolderSlice []*Folder

	folderQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	folderType                 = reflect.TypeOf(&Folder{})
	folderMapping              = queries.MakeStructMapping(folderType)
	folderPrimaryKeyMapping, _ = queries.BindMapping(folderType, folderMapping, folderPrimaryKeyColumns)
	folderInsertCacheMut       sync.RWMutex
	folderInsertCache          = make(map[string]insertCache)
	folderUpdateCacheMut       sync.RWMutex
	folderUpdateCache          = make(map[string]updateCache)
	folderUpsertCacheMut       sync.RWMutex
	folderUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

// One returns a single folder record from the query.
func (q folderQuery) One(exec boil.Executor) (*Folder, error) {
	o := &Folder{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for folders")
	}

	return o, nil
}

// All returns all Folder records from the query.
func (q folderQuery) All(exec boil.Executor) (FolderSlice, error) {
	var o []*Folder

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to Folder slice")
	}

	return o, nil
}

// Count returns the count of all Folder records in the query.
func (q folderQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count folders rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q folderQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if folders exists")
	}

	return count > 0, nil
}

// User pointed to by the foreign key.
func (o *Folder) User(mods ...qm.QueryMod) userQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.UserID),
	}

	queryMods = append(queryMods, mods...)

	query := Users(queryMods...)
	queries.SetFrom(query.Query, "\"users\"")

	return query
}

// BookmarkFolders retrieves all the bookmark_folder's BookmarkFolders with an executor.
func (o *Folder) BookmarkFolders(mods ...qm.QueryMod) bookmarkFolderQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"bookmark_folder\".\"folder_id\"=?", o.ID),
	)

	query := BookmarkFolders(queryMods...)
	queries.SetFrom(query.Query, "\"bookmark_folder\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"bookmark_folder\".*"})
	}

	return query
}

// LoadUser allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (folderL) LoadUser(e boil.Executor, singular bool, maybeFolder interface{}, mods queries.Applicator) error {
	var slice []*Folder
	var object *Folder

	if singular {
		object = maybeFolder.(*Folder)
	} else {
		slice = *maybeFolder.(*[]*Folder)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &folderR{}
		}
		args = append(args, object.UserID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &folderR{}
			}

			for _, a := range args {
				if a == obj.UserID {
					continue Outer
				}
			}

			args = append(args, obj.UserID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`users`),
		qm.WhereIn(`users.id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load User")
	}

	var resultSlice []*User
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice User")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for users")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for users")
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.User = foreign
		if foreign.R == nil {
			foreign.R = &userR{}
		}
		foreign.R.Folders = append(foreign.R.Folders, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.UserID == foreign.ID {
				local.R.User = foreign
				if foreign.R == nil {
					foreign.R = &userR{}
				}
				foreign.R.Folders = append(foreign.R.Folders, local)
				break
			}
		}
	}

	return nil
}

// LoadBookmarkFolders allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (folderL) LoadBookmarkFolders(e boil.Executor, singular bool, maybeFolder interface{}, mods queries.Applicator) error {
	var slice []*Folder
	var object *Folder

	if singular {
		object = maybeFolder.(*Folder)
	} else {
		slice = *maybeFolder.(*[]*Folder)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &folderR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &folderR{}
			}

			for _, a := range args {
				if a == obj.ID {
					continue Outer
				}
			}

			args = append(args, obj.ID)
		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`bookmark_folder`),
		qm.WhereIn(`bookmark_folder.folder_id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load bookmark_folder")
	}

	var resultSlice []*BookmarkFolder
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice bookmark_folder")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on bookmark_folder")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for bookmark_folder")
	}

	if singular {
		object.R.BookmarkFolders = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &bookmarkFolderR{}
			}
			foreign.R.Folder = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ID == foreign.FolderID {
				local.R.BookmarkFolders = append(local.R.BookmarkFolders, foreign)
				if foreign.R == nil {
					foreign.R = &bookmarkFolderR{}
				}
				foreign.R.Folder = local
				break
			}
		}
	}

	return nil
}

// SetUser of the folder to the related item.
// Sets o.R.User to related.
// Adds o to related.R.Folders.
func (o *Folder) SetUser(exec boil.Executor, insert bool, related *User) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"folders\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"user_id"}),
		strmangle.WhereClause("\"", "\"", 2, folderPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.UserID = related.ID
	if o.R == nil {
		o.R = &folderR{
			User: related,
		}
	} else {
		o.R.User = related
	}

	if related.R == nil {
		related.R = &userR{
			Folders: FolderSlice{o},
		}
	} else {
		related.R.Folders = append(related.R.Folders, o)
	}

	return nil
}

// AddBookmarkFolders adds the given related objects to the existing relationships
// of the folder, optionally inserting them as new records.
// Appends related to o.R.BookmarkFolders.
// Sets related.R.Folder appropriately.
func (o *Folder) AddBookmarkFolders(exec boil.Executor, insert bool, related ...*BookmarkFolder) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.FolderID = o.ID
			if err = rel.Insert(exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"bookmark_folder\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"folder_id"}),
				strmangle.WhereClause("\"", "\"", 2, bookmarkFolderPrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.FolderID, rel.BookmarkID}

			if boil.DebugMode {
				fmt.Fprintln(boil.DebugWriter, updateQuery)
				fmt.Fprintln(boil.DebugWriter, values)
			}
			if _, err = exec.Exec(updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.FolderID = o.ID
		}
	}

	if o.R == nil {
		o.R = &folderR{
			BookmarkFolders: related,
		}
	} else {
		o.R.BookmarkFolders = append(o.R.BookmarkFolders, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &bookmarkFolderR{
				Folder: o,
			}
		} else {
			rel.R.Folder = o
		}
	}
	return nil
}

// Folders retrieves all the records using an executor.
func Folders(mods ...qm.QueryMod) folderQuery {
	mods = append(mods, qm.From("\"folders\""))
	return folderQuery{NewQuery(mods...)}
}

// FindFolder retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindFolder(exec boil.Executor, iD int64, selectCols ...string) (*Folder, error) {
	folderObj := &Folder{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"folders\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, folderObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from folders")
	}

	return folderObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Folder) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no folders provided for insertion")
	}

	var err error

	nzDefaults := queries.NonZeroDefaultSet(folderColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	folderInsertCacheMut.RLock()
	cache, cached := folderInsertCache[key]
	folderInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			folderAllColumns,
			folderColumnsWithDefault,
			folderColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(folderType, folderMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(folderType, folderMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"folders\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"folders\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "models: unable to insert into folders")
	}

	if !cached {
		folderInsertCacheMut.Lock()
		folderInsertCache[key] = cache
		folderInsertCacheMut.Unlock()
	}

	return nil
}

// Update uses an executor to update the Folder.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Folder) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	key := makeCacheKey(columns, nil)
	folderUpdateCacheMut.RLock()
	cache, cached := folderUpdateCache[key]
	folderUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			folderAllColumns,
			folderPrimaryKeyColumns,
		)
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update folders, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"folders\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, folderPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(folderType, folderMapping, append(wl, folderPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update folders row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for folders")
	}

	if !cached {
		folderUpdateCacheMut.Lock()
		folderUpdateCache[key] = cache
		folderUpdateCacheMut.Unlock()
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values.
func (q folderQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for folders")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for folders")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o FolderSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), folderPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"folders\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, folderPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in folder slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all folder")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Folder) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no folders provided for upsert")
	}

	nzDefaults := queries.NonZeroDefaultSet(folderColumnsWithDefault, o)

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

	folderUpsertCacheMut.RLock()
	cache, cached := folderUpsertCache[key]
	folderUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			folderAllColumns,
			folderColumnsWithDefault,
			folderColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			folderAllColumns,
			folderPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert folders, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(folderPrimaryKeyColumns))
			copy(conflict, folderPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"folders\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(folderType, folderMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(folderType, folderMapping, ret)
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
		return errors.Wrap(err, "models: unable to upsert folders")
	}

	if !cached {
		folderUpsertCacheMut.Lock()
		folderUpsertCache[key] = cache
		folderUpsertCacheMut.Unlock()
	}

	return nil
}

// Delete deletes a single Folder record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Folder) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no Folder provided for delete")
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), folderPrimaryKeyMapping)
	sql := "DELETE FROM \"folders\" WHERE \"id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from folders")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for folders")
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q folderQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no folderQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from folders")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for folders")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o FolderSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), folderPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"folders\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, folderPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from folder slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for folders")
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *Folder) Reload(exec boil.Executor) error {
	ret, err := FindFolder(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *FolderSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := FolderSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), folderPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"folders\".* FROM \"folders\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, folderPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in FolderSlice")
	}

	*o = slice

	return nil
}

// FolderExists checks if the Folder row exists.
func FolderExists(exec boil.Executor, iD int64) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"folders\" where \"id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if folders exists")
	}

	return exists, nil
}
