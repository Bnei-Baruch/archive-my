package chronicles

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"archive-my/models"
	"archive-my/pkg/utils"
)

const (
	SCAN_SIZE     = 30
	MAX_INTERVAL  = time.Duration(time.Minute)
	MIN_INTERVAL  = time.Duration(100 * time.Millisecond)
	WAIT_FOR_SAVE = time.Duration(1 * time.Minute)
)

type Chronicles struct {
	ticker   *time.Ticker
	interval time.Duration

	lastReadId  string
	prevReadId  string
	nextRefresh time.Time

	DBstr  string
	MDBstr string
}

type ScanResponse struct {
	Entries []*ChronicleEvent `json:"entries"`
}

type ChronicleEvent struct {
	AccountId       string             `json:"user_id"`
	CreatedAt       time.Time          `json:"created_at"`
	IPAddr          string             `boil:"ip_addr" json:"ip_addr" toml:"ip_addr" yaml:"ip_addr"`
	ID              string             `json:"id"`
	UserAgent       string             `json:"user_agent"`
	Namespace       string             `json:"namespace"`
	ClientEventID   null.String        `json:"client_event_id,omitempty"`
	ClientEventType string             `json:"client_event_type"`
	ClientFlowID    null.String        `json:"client_flow_id,omitempty"`
	ClientFlowType  null.String        `json:"client_flow_type,omitempty"`
	ClientSessionID null.String        `toml:"client_session_id"`
	Data            ChronicleEventData `json:"data,omitempty"`
}

type ChronicleEventData struct {
	UnitUID     string       `json:"unit_uid,omitempty"`
	TimeZone    string       `json:"time_zone,omitempty"`
	CurrentTime null.Float64 `json:"current_time,omitempty"`
}

func (c *Chronicles) Init(dbstr, mdbstr string) {
	if dbstr == "" {
		dbstr = viper.GetString("app.mydb")
	}
	c.DBstr = dbstr

	if mdbstr == "" {
		mdbstr = viper.GetString("app.mdb")
	}
	c.MDBstr = mdbstr
}

func (c *Chronicles) Run() {
	c.interval = MIN_INTERVAL
	c.ticker = time.NewTicker(MIN_INTERVAL)

	mdb, err := sql.Open("postgres", c.MDBstr)
	utils.Must(err)
	utils.Must(mdb.Ping())
	defer mdb.Close()
	utils.Must(utils.InitCT(mdb))

	c.lastReadId, err = c.lastChroniclesId()
	utils.Must(err)

	go func() {
		refresh := func() {
			if err := c.refresh(); err != nil {
				log.Errorf("Error Refresh: %+v", err)
				c.interval = c.interval * 2
				_ = c.refresh()
			}
		}
		refresh()

		for range c.ticker.C {
			refresh()
		}
	}()
}

func (c *Chronicles) Stop() {
	c.ticker.Stop()
}

func (c *Chronicles) lastChroniclesId() (string, error) {
	db, err := sql.Open("postgres", c.DBstr)
	if err != nil {
		return "", err
	}
	if err := db.Ping(); err != nil {
		return "", err
	}
	defer db.Close()

	h, err := models.Histories(qm.OrderBy("chronicle_id")).One(db)
	if err == sql.ErrNoRows {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return h.ChronicleID, nil
}

func (c *Chronicles) refresh() error {
	entries, err := c.scanEvents()
	if err != nil {
		return err
	}
	if len(entries) == SCAN_SIZE {
		c.interval = maxDuration(c.interval/2, MIN_INTERVAL)
	} else {
		c.interval = minDuration(c.interval*2, MAX_INTERVAL)
	}
	c.ticker.Reset(c.interval)
	return nil
}

func (c *Chronicles) scanEvents() ([]*ChronicleEvent, error) {
	db, err := sql.Open("postgres", c.DBstr)
	utils.Must(err)
	utils.Must(db.Ping())
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	events, err := c.scanEventsOnTx(tx, viper.GetString("app.scan_url"))
	var eTx error
	if err == nil {
		eTx = tx.Commit()
	} else {
		eTx = tx.Rollback()
	}
	if eTx != nil {
		return nil, eTx
	}
	return events, nil
}

func (c *Chronicles) scanEventsOnTx(tx *sql.Tx, scanUrl string) ([]*ChronicleEvent, error) {
	log.Infof("Scanning chronicles entries, last successfull [%s]", c.lastReadId)
	args := fmt.Sprintf(`{"id":"%s","limit":%d, "event_types": ["player-play", "player-stop"], "namespaces": ["archive"], "keycloak": true}`, c.lastReadId, SCAN_SIZE)
	log.Infof("Scan chronicles with arguments [%s]", args)
	b := bytes.NewBuffer([]byte(args))
	resp, err := http.Post(scanUrl, "application/json", b)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("Response code %d for scan: %s.", resp.StatusCode, resp.Status))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var scanResponse ScanResponse
	if err = json.Unmarshal(body, &scanResponse); err != nil {
		return nil, err
	}

	if len(scanResponse.Entries) > 0 {
		c.lastReadId = scanResponse.Entries[len(scanResponse.Entries)-1].ID
	}
	if err := c.saveEvents(tx, scanResponse.Entries); err != nil {
		return nil, err
	}
	return scanResponse.Entries, nil
}

func (c *Chronicles) saveEvents(tx *sql.Tx, events []*ChronicleEvent) error {
	mdb, err := sql.Open("postgres", c.MDBstr)
	utils.Must(err)
	utils.Must(mdb.Ping())
	defer mdb.Close()

	evByAcc := make(map[string]*ChronicleEvent, 0)
	for _, x := range events {
		k := fmt.Sprintf("%s_%s", x.AccountId, x.Data.UnitUID)
		if x.Data.CurrentTime.Float64 > WAIT_FOR_SAVE.Minutes() {
			evByAcc[k] = x
		}
	}

	for _, x := range evByAcc {
		if x == nil {
			continue
		}
		err = c.insertEvent(tx, mdb, x)
		if err != nil {
			break
		}
		err = c.updateSubscriptions(tx, mdb, x)
		if err != nil {
			break
		}
	}
	return err
}

func (c *Chronicles) insertEvent(tx *sql.Tx, mdb *sql.DB, ev *ChronicleEvent) error {
	nParams := make(map[string]interface{})
	nParams["current_time"] = ev.Data.CurrentTime

	j, err := json.Marshal(nParams)
	if err != nil {
		return err
	}

	year, month, day := ev.CreatedAt.Date()
	var tz string
	if ev.Data.TimeZone != "" {
		tz = fmt.Sprint(ev.Data.TimeZone)
	}
	timeZone, err := time.LoadLocation(tz)
	if err != nil {
		return err
	}
	sDay := time.Date(year, month, day, 0, 0, 0, 0, timeZone)
	eDay := sDay.Add(24 * time.Hour)
	log.Infof("%v, %v", sDay, eDay)
	h, errDB := models.Histories(
		qm.Where("account_id = ? AND content_unit_uid = ? AND created_at > ? AND created_at < ?", ev.AccountId, ev.Data.UnitUID, sDay, eDay),
	).One(tx)
	if errDB == sql.ErrNoRows {
		h = &models.History{
			AccountID:      ev.AccountId,
			ChronicleID:    ev.ID,
			ContentUnitUID: null.String{String: ev.Data.UnitUID, Valid: true},
			Data:           null.JSON{JSON: j, Valid: true},
			CreatedAt:      ev.CreatedAt,
		}
		return h.Insert(tx, boil.Infer())
	} else if errDB != nil {
		return err
	}
	h.ChronicleID = ev.ID
	h.CreatedAt = ev.CreatedAt
	params, err := margeData(h.Data, nParams)
	if err != nil {
		return err
	}
	h.Data = *params
	_, err = h.Update(tx, boil.Infer())
	return err
}

func (c *Chronicles) updateSubscriptions(tx boil.Executor, mdb *sql.DB, ev *ChronicleEvent) error {
	subs, err := models.Subscriptions(qm.Where("account_id = ?", ev.AccountId)).All(tx)
	if subs == nil {
		return nil
	} else if err != nil {
		return err
	}

	byCOs := make([]*models.Subscription, 0)
	byTypes := make([]*models.Subscription, 0)
	cuUIDs := make([]int64, len(subs))
	for i, s := range subs {
		cuUIDs[i] = s.ID
		if s.CollectionUID.Valid && s.CollectionUID.String != "" {
			byCOs = append(byCOs, s)
		} else if s.ContentType.Valid && s.ContentType.String != "" {
			byTypes = append(byTypes, s)
		}
	}

	query := fmt.Sprintf(`SELECT co.uid, co.type_id FROM collections_content_units ccu
			INNER JOIN content_units cu ON ccu.content_unit_id = cu.id
			INNER JOIN collections co ON ccu.collection_id = co.id
			WHERE cu.uid = '%s'`, ev.Data.UnitUID)

	rows, err := queries.Raw(query).Query(mdb)
	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return err
	}
	defer rows.Close()
	forUpdate := models.SubscriptionSlice{}
	for rows.Next() {
		var (
			coUid  string
			typeId int
		)
		err = rows.Scan(&coUid, &typeId)
		if err != nil {
			return err
		}
		name := utils.ContentTypesByID[typeId]
		for _, s := range byTypes {
			if s.ContentType.String == name {
				forUpdate = append(forUpdate, s)
			}
		}
		for _, s := range byCOs {
			if s.CollectionUID.String == coUid {
				forUpdate = append(forUpdate, s)
			}
		}
	}
	if len(forUpdate) == 0 {
		return nil
	}

	col := make(map[string]interface{}, 0)
	col["updated_at"] = ev.CreatedAt
	_, err = forUpdate.UpdateAll(tx, col)
	return err
}

func margeData(data null.JSON, nd map[string]interface{}) (*null.JSON, error) {
	var d map[string]interface{}
	if err := json.Unmarshal(data.JSON, &d); err != nil {
		return nil, err
	}

	for k, v := range nd {
		d[k] = v
	}
	dStr, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	return &null.JSON{JSON: dStr, Valid: true}, nil
}

func minDuration(x, y time.Duration) time.Duration {
	if x < y {
		return x
	}
	return y
}

func maxDuration(x, y time.Duration) time.Duration {
	if x >= y {
		return x
	}
	return y
}
