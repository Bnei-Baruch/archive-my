package utils

/*
import (
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
)


func InitTypeRegistries(exec boil.Executor) error {
	ContentTypeRegistry.Init(exec)

	return nil
}

type TypeRegistry interface {
	Init(exec boil.Executor) error
}

type ContentTypeRegistry struct {
	ByName map[string]*models.ContentType
	ByID   map[int64]*models.ContentType
}

func (r *ContentTypeRegistry) Init(exec boil.Executor) error {
	types, err := models.ContentTypes(exec).All()
	if err != nil {
		return errors.Wrap(err, "Load content_types from DB")
	}

	r.ByName = make(map[string]*models.ContentType)
	r.ByID = make(map[int64]*models.ContentType)
	for _, t := range types {
		r.ByName[t.Name] = t
		r.ByID[t.ID] = t
	}

	return nil
}

type ContentRoleTypeRegistry struct {
	ByName map[string]*models.ContentRoleType
}

func (r *ContentRoleTypeRegistry) Init(exec boil.Executor) error {
	types, err := models.ContentRoleTypes(exec).All()
	if err != nil {
		return errors.Wrap(err, "Load content_role_types from DB")
	}

	r.ByName = make(map[string]*models.ContentRoleType)
	for _, t := range types {
		r.ByName[t.Name] = t
	}

	return nil
}

type PersonsRegistry struct {
	ByPattern map[string]*models.Person
}

func (r *PersonsRegistry) Init(exec boil.Executor) error {
	types, err := models.Persons(exec, qm.Where("pattern is not null")).All()
	if err != nil {
		return errors.Wrap(err, "Load persons from DB")
	}

	r.ByPattern = make(map[string]*models.Person)
	for _, t := range types {
		r.ByPattern[t.Pattern.String] = t
	}

	return nil
}

type AuthorRegistry struct {
	ByCode map[string]*models.Author
}

func (r *AuthorRegistry) Init(exec boil.Executor) error {
	authors, err := models.Authors(exec).All()
	if err != nil {
		return errors.Wrap(err, "Load authors from DB")
	}

	r.ByCode = make(map[string]*models.Author)
	for _, a := range authors {
		r.ByCode[a.Code] = a
	}

	return nil
}

type SourceTypeRegistry struct {
	ByName map[string]*models.SourceType
	ByID   map[int64]*models.SourceType
}

func (r *SourceTypeRegistry) Init(exec boil.Executor) error {
	types, err := models.SourceTypes(exec).All()
	if err != nil {
		return errors.Wrap(err, "Load source_types from DB")
	}

	r.ByName = make(map[string]*models.SourceType)
	r.ByID = make(map[int64]*models.SourceType)
	for _, t := range types {
		r.ByName[t.Name] = t
		r.ByID[t.ID] = t
	}

	return nil
}

type TwitterUsersRegistry struct {
	ByUsername map[string]*models.TwitterUser
	ByID       map[int64]*models.TwitterUser
}

func (r *TwitterUsersRegistry) Init(exec boil.Executor) error {
	users, err := models.TwitterUsers(exec).All()
	if err != nil {
		return errors.Wrap(err, "Load twitter users from DB")
	}

	r.ByUsername = make(map[string]*models.TwitterUser)
	r.ByID = make(map[int64]*models.TwitterUser)
	for _, t := range users {
		r.ByUsername[t.Username] = t
		r.ByID[t.ID] = t
	}

	return nil
}

type BlogsRegistry struct {
	ByName map[string]*models.Blog
	ByID   map[int64]*models.Blog
}

func (r *BlogsRegistry) Init(exec boil.Executor) error {
	blogs, err := models.Blogs(exec).All()
	if err != nil {
		return errors.Wrap(err, "Load blogs from DB")
	}

	r.ByName = make(map[string]*models.Blog)
	r.ByID = make(map[int64]*models.Blog)
	for _, b := range blogs {
		r.ByName[b.Name] = b
		r.ByID[b.ID] = b
	}

	return nil
}
*/
