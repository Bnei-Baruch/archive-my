package utils

import (
	"database/sql"
)

var ContentTypesByName map[string]int
var ContentTypesByID map[int]string

func InitCT(db *sql.DB) error {
	ContentTypesByName = make(map[string]int, 0)
	ContentTypesByID = make(map[int]string, 0)

	rows, err := db.Query("SELECT id, name FROM content_types ct")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			name string
			id   int
		)
		err = rows.Scan(&id, &name)
		if err != nil {
			return err
		}
		ContentTypesByName[name] = id
		ContentTypesByID[id] = name
	}
	return nil
}
