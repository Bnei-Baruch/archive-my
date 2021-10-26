package chronicles

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/Bnei-Baruch/archive-my/common"
	"github.com/Bnei-Baruch/archive-my/databases/mdb"
	"github.com/Bnei-Baruch/archive-my/databases/mydb/models"
	"github.com/Bnei-Baruch/archive-my/pkg/utils"
)

const (
	CT_CLIPS = "CLIPS"
	CT_CLIP  = "CLIP"
)

func (s *ChroniclesTestSuite) TestChronicles_byCO() {
	user := s.CreateUser()
	newUpdate := time.Now().UTC().Format(time.RFC3339)
	coUID := s.CreateMDBCollection(s.MDB.DB, CT_CLIPS)
	cuUID := s.CreateMDBContentUnit(s.MDB.DB, CT_CLIP)
	s.AssociateCollectionAndUnit(s.MDB.DB, coUID, cuUID)

	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// TODO: assert request
		data := s.readTestResource("subs_test.json")
		_, err := res.Write([]byte(fmt.Sprintf(string(data), cuUID, user.AccountsID, newUpdate)))
		s.Require().NoError(err)
	}))
	defer server.Close()
	common.Config.ChroniclesUrl = server.URL

	subscription := models.Subscription{
		UserID:         user.ID,
		CollectionUID:  null.StringFrom(coUID),
		ContentUnitUID: null.StringFrom(cuUID),
	}
	s.Require().NoError(subscription.Insert(s.MyDB.DB, boil.Infer()))

	chr := new(Chronicles)
	chr.InitWithDeps(s.MyDB.DB, s.MDB.DB)

	_, err := chr.scanEvents()
	s.Require().NoError(err)

	subscriptions, err := models.Subscriptions(models.SubscriptionWhere.UserID.EQ(user.ID)).All(s.MyDB.DB)
	s.Len(subscriptions, 1)
	s.Equal(newUpdate, subscriptions[0].UpdatedAt.Time.UTC().Format(time.RFC3339))
}

func (s *ChroniclesTestSuite) TestChronicles_byType() {
	user := s.CreateUser()
	newUpdate := time.Now().UTC().Format(time.RFC3339)

	cuUID := s.CreateMDBContentUnit(s.MDB.DB, CT_CLIP)

	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// TODO: assert request
		data := s.readTestResource("subs_test.json")
		_, err := res.Write([]byte(fmt.Sprintf(string(data), cuUID, user.AccountsID, newUpdate)))
		s.Require().NoError(err)
	}))
	defer server.Close()
	common.Config.ChroniclesUrl = server.URL

	subscription := models.Subscription{
		UserID:         user.ID,
		ContentType:    null.StringFrom(CT_CLIPS),
		ContentUnitUID: null.StringFrom(cuUID),
	}
	s.Require().NoError(subscription.Insert(s.MyDB.DB, boil.Infer()))

	chr := new(Chronicles)
	chr.InitWithDeps(s.MyDB.DB, s.MDB.DB)

	_, err := chr.scanEvents()
	s.Require().NoError(err)

	subscriptions, err := models.Subscriptions(models.SubscriptionWhere.UserID.EQ(user.ID)).All(s.MyDB.DB)
	s.Len(subscriptions, 1)
	s.Equal(newUpdate, subscriptions[0].UpdatedAt.Time.UTC().Format(time.RFC3339))
}

func (s *ChroniclesTestSuite) CreateMDBCollection(db *sql.DB, contentType string) string {
	uid := utils.GenerateUID(8)
	typeID := mdb.ContentTypesByName[contentType]
	q := fmt.Sprintf("INSERT INTO collections (uid, type_id, published) VALUES ('%s', %d, %t)", uid, typeID, true)
	_, err := db.Exec(q)
	s.Require().NoError(err)
	return uid
}

func (s *ChroniclesTestSuite) CreateMDBContentUnit(db *sql.DB, contentType string) string {
	uid := utils.GenerateUID(8)
	typeID := mdb.ContentTypesByName[contentType]
	q := fmt.Sprintf("INSERT INTO content_units (uid, type_id, published) VALUES ('%s', %d, %t)", uid, typeID, true)
	_, err := db.Exec(q)
	s.Require().NoError(err)

	coUID := s.CreateMDBCollection(s.MDB.DB, CT_CLIPS)
	s.AssociateCollectionAndUnit(s.MDB.DB, coUID, uid)
	return uid
}

func (s *ChroniclesTestSuite) AssociateCollectionAndUnit(db *sql.DB, coUID, cuUID string) {
	var coID int
	err := db.QueryRow(fmt.Sprintf("SELECT id FROM collections WHERE uid = '%s'", coUID)).Scan(&coID)
	s.Require().NoError(err)

	var cuID int
	err = db.QueryRow(fmt.Sprintf("SELECT id FROM content_units WHERE uid = '%s'", cuUID)).Scan(&cuID)
	s.Require().NoError(err)

	_, err = db.Exec(fmt.Sprintf("INSERT INTO collections_content_units (collection_id, content_unit_id, name) VALUES (%d, %d, 'test')", coID, cuID))
	s.Require().NoError(err)
}
