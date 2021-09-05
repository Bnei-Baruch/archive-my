package utils

import (
	"github.com/volatiletech/sqlboiler/v4/boil"
)

var ContentTypesByName map[string]*contentType
var ContentTypesByID map[int]*contentType

type contentType struct {
	Name string
	ID   int
}

func InitCT(db boil.Executor) error {
	q, err := db.Query("SELECT id, name FROM content_types ct")
	if err != nil {
		return err
	}
	var cts []*contentType
	err = q.Scan(&cts)
	for _, ct := range cts {
		ContentTypesByName[ct.Name] = ct
		ContentTypesByID[ct.ID] = ct
	}
	return nil
}
