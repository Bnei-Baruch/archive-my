package chronicles

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/Bnei-Baruch/archive-my/common"
	"github.com/Bnei-Baruch/archive-my/databases/mydb/models"
	"github.com/Bnei-Baruch/archive-my/pkg/utils"
)

func (s *ChroniclesTestSuite) TestChronicles_simple() {
	user := s.CreateUser()

	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// TODO: assert request
		data := s.readTestResource("simple_scan_test.json")
		_, err := res.Write([]byte(fmt.Sprintf(string(data), user.AccountsID)))
		s.Require().NoError(err)
	}))
	defer server.Close()
	common.Config.ChroniclesUrl = server.URL

	chr := new(Chronicles)
	chr.InitWithDeps(s.MyDB.DB, s.MDB.DB)

	_, err := chr.scanEvents()
	s.Require().NoError(err)

	count, err := models.Histories().Count(s.MyDB.DB)
	s.Require().NoError(err)
	s.EqualValues(1, count)
	count, err = models.Histories(models.HistoryWhere.UserID.EQ(user.ID)).Count(s.MyDB.DB)
	s.Require().NoError(err)
	s.EqualValues(1, count)
}

func (s *ChroniclesTestSuite) TestChronicles_multiAcc() {
	user1 := s.CreateUser()
	user2 := s.CreateUser()

	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// TODO: assert request
		data := s.readTestResource("multiacc_scan_test.json")
		_, err := res.Write([]byte(fmt.Sprintf(string(data), user1.AccountsID, user2.AccountsID)))
		s.Require().NoError(err)
	}))
	defer server.Close()
	common.Config.ChroniclesUrl = server.URL

	chr := new(Chronicles)
	chr.InitWithDeps(s.MyDB.DB, s.MDB.DB)

	_, err := chr.scanEvents()
	s.Require().NoError(err)

	count, err := models.Histories().Count(s.MyDB.DB)
	s.Require().NoError(err)
	s.EqualValues(2, count)
	count, err = models.Histories(models.HistoryWhere.UserID.EQ(user1.ID)).Count(s.MyDB.DB)
	s.Require().NoError(err)
	s.EqualValues(1, count)
	count, err = models.Histories(models.HistoryWhere.UserID.EQ(user2.ID)).Count(s.MyDB.DB)
	s.Require().NoError(err)
	s.EqualValues(1, count)
}

func (s *ChroniclesTestSuite) TestChronicles_multiEv() {
	user1 := s.CreateUser()
	user2 := s.CreateUser()
	cuID := utils.GenerateUID(8)
	updated := 2000.00

	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// TODO: assert request
		data := s.readTestResource("multiev1_scan_test.json")
		_, err := res.Write([]byte(fmt.Sprintf(string(data), user1.AccountsID, user2.AccountsID, updated, cuID)))
		s.Require().NoError(err)
	}))
	defer server.Close()
	common.Config.ChroniclesUrl = server.URL

	chr := new(Chronicles)
	chr.InitWithDeps(s.MyDB.DB, s.MDB.DB)

	_, err := chr.scanEvents()
	s.Require().NoError(err)

	count, err := models.Histories().Count(s.MyDB.DB)
	s.Require().NoError(err)
	s.EqualValues(2, count)

	history, err := models.Histories(models.HistoryWhere.UserID.EQ(user1.ID)).All(s.MyDB.DB)
	s.Require().NoError(err)
	s.Len(history, 1)
	var data ChronicleEventData
	s.NoError(json.Unmarshal(history[0].Data.JSON, &data))
	s.Equal(updated, data.CurrentTime.Float64)

	count, err = models.Histories(models.HistoryWhere.UserID.EQ(user2.ID)).Count(s.MyDB.DB)
	s.Require().NoError(err)
	s.EqualValues(1, count)

	//second call
	updated = 5000.01
	server.Close()
	server = httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// TODO: assert request
		data := s.readTestResource("multiev2_scan_test.json")
		_, err := res.Write([]byte(fmt.Sprintf(string(data), user1.AccountsID, user2.AccountsID, updated, cuID)))
		s.Require().NoError(err)
	}))
	defer server.Close()
	common.Config.ChroniclesUrl = server.URL

	_, err = chr.scanEvents()
	s.Require().NoError(err)

	history, err = models.Histories(models.HistoryWhere.UserID.EQ(user1.ID)).All(s.MyDB.DB)
	s.Require().NoError(err)
	s.Len(history, 1)
	s.NoError(json.Unmarshal(history[0].Data.JSON, &data))
	s.Equal(updated, data.CurrentTime.Float64)

	count, err = models.Histories(models.HistoryWhere.UserID.EQ(user2.ID)).Count(s.MyDB.DB)
	s.Require().NoError(err)
	s.EqualValues(2, count)
}
