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

	"github.com/Bnei-Baruch/archive-my/models"
	"github.com/Bnei-Baruch/archive-my/pkg/utils"
)

const (
	SCAN_SIZE                 = 1000
	MAX_INTERVAL              = time.Duration(time.Minute)
	MIN_INTERVAL              = time.Duration(100 * time.Millisecond)
	CR_EVENT_TYPE_PLAYER_PLAY = "player-play"
	CR_EVENT_TYPE_PLAYER_STOP = "player-stop"
	WAIT_FOR_SAVE             = time.Duration(5 * time.Minute)
)

type Chronicles struct {
	ticker   *time.Ticker
	interval time.Duration
	evByAcc  map[string]*ChronicleEvent

	lastReadId  string
	prevReadId  string
	nextRefresh time.Time

	DBstr  string
	MDBstr string

	httpClient *http.Client
}

type ScanResponse struct {
	Entries []*ChronicleEvent `json:"entries"`
}

type ChronicleEvent struct {
	AccountId       string      `json:"user_id"`
	CreatedAt       time.Time   `json:"created_at"`
	IPAddr          string      `boil:"ip_addr" json:"ip_addr" toml:"ip_addr" yaml:"ip_addr"`
	ID              string      `json:"id"`
	UserAgent       string      `json:"user_agent"`
	Namespace       string      `json:"namespace"`
	ClientEventID   null.String `json:"client_event_id,omitempty"`
	ClientEventType string      `json:"client_event_type"`
	ClientFlowID    null.String `json:"client_flow_id,omitempty"`
	ClientFlowType  null.String `json:"client_flow_type,omitempty"`
	ClientSessionID null.String `toml:"client_session_id"`
	Data            null.JSON   `json:"data,omitempty"`
	FirstScanAt     time.Time   `json:"-"`
}

type ChronicleEventData struct {
	UnitUID     string     `json:"unit_uid"`
	TimeZone    string     `json:"time_zone,omitempty"`
	CurrentTime null.Int64 `json:"current_time,omitempty"`
}

func (c *Chronicles) Init(dbstr, mdbstr string, client *http.Client) {
	if dbstr == "" {
		dbstr = viper.GetString("app.mydb")
	}
	c.DBstr = dbstr

	if mdbstr == "" {
		mdbstr = viper.GetString("app.mdb")
	}
	c.MDBstr = mdbstr

	if client == nil {
		client = &http.Client{}
	}
	c.httpClient = client
}

func (c *Chronicles) Run() {
	c.interval = MIN_INTERVAL
	c.ticker = time.NewTicker(MIN_INTERVAL)
	c.evByAcc = make(map[string]*ChronicleEvent, 0)

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
	log.Infof("Scanning chronicles entries, last successfull [%s]", c.lastReadId)
	b := bytes.NewBuffer([]byte(fmt.Sprintf(`{"id":"%s","limit":%d, "event_types": ["player-play", "player-stop"], "namespaces": ["archive"]}`, c.lastReadId, SCAN_SIZE)))
	resp, err := c.httpClient.Post(viper.GetString("app.scan_url"), "application/json", b)
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
	if err := c.saveEvents(scanResponse.Entries); err != nil {
		return nil, err
	}
	return scanResponse.Entries, nil
}

func (c *Chronicles) saveEvents(events []*ChronicleEvent) error {
	db, err := sql.Open("postgres", c.DBstr)
	utils.Must(err)
	utils.Must(db.Ping())
	defer db.Close()

	mdb, err := sql.Open("postgres", c.MDBstr)
	utils.Must(err)
	utils.Must(mdb.Ping())
	defer mdb.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	for _, x := range events {
		if x.ClientEventType != CR_EVENT_TYPE_PLAYER_PLAY && x.ClientEventType != CR_EVENT_TYPE_PLAYER_STOP {
			continue
		}

		if prevE, ok := c.evByAcc[x.AccountId]; ok && prevE != nil {
			if x.ClientEventType == CR_EVENT_TYPE_PLAYER_STOP && x.FirstScanAt.After(c.nextRefresh.Add(WAIT_FOR_SAVE)) {
				c.evByAcc[x.AccountId] = nil
				continue
			}
			if x.ClientEventType != CR_EVENT_TYPE_PLAYER_PLAY {
				x.FirstScanAt = prevE.FirstScanAt
			}
		} else {
			x.FirstScanAt = c.nextRefresh
		}
		c.evByAcc[x.AccountId] = x
	}

	for _, x := range c.evByAcc {
		if x == nil {
			continue
		}
		if x.FirstScanAt.After(c.nextRefresh.Add(WAIT_FOR_SAVE)) {
			continue
		}
		c.evByAcc[x.AccountId] = nil
		err = c.insertEvent(tx, mdb, x)
		if err != nil {
			break
		}
	}

	var eTx error
	if err == nil {
		eTx = tx.Commit()
	} else {
		eTx = tx.Rollback()
	}
	if eTx != nil {
		return eTx
	}
	return nil
}

func (c *Chronicles) insertEvent(tx *sql.Tx, mdb *sql.DB, ev *ChronicleEvent) error {
	var data map[string]interface{}
	if err := json.Unmarshal(ev.Data.JSON, &data); err != nil {
		return err
	}

	unitUID := fmt.Sprint(data["unit_uid"])
	if err := c.updateSubscriptions(tx, mdb, ev, unitUID); err != nil {
		return err
	}
	nParams := make(map[string]interface{})
	nParams["current_time"] = data["current_time"]

	j, err := json.Marshal(nParams)
	if err != nil {
		return err
	}

	year, month, day := ev.CreatedAt.Date()
	var tz string
	if v, ok := data["time_zone"]; ok {
		tz = fmt.Sprint(v)
	}
	timeZone, err := time.LoadLocation(tz)
	if err != nil {
		return err
	}
	sDay := time.Date(year, month, day, 0, 0, 0, 0, timeZone)
	eDay := sDay.Add(24 * time.Hour)
	log.Infof("%v, %v", sDay, eDay)
	h, errDB := models.Histories(
		qm.Where("account_id = ? AND content_unit_uid = ? AND created_at > ? AND created_at < ?", ev.AccountId, unitUID, sDay, eDay),
	).One(tx)
	if errDB == sql.ErrNoRows {
		h = &models.History{
			AccountID:      ev.AccountId[0:36],
			ChronicleID:    ev.ID,
			ContentUnitUID: null.String{String: unitUID, Valid: true},
			Data:           null.JSON{JSON: j, Valid: true},
			CreatedAt:      ev.CreatedAt,
		}
		return h.Insert(tx, boil.Infer())
	} else if errDB != nil {
		return err
	}

	params, err := margeData(h.Data, nParams)
	if err != nil {
		return err
	}
	h.Data = *params
	_, err = h.Update(tx, boil.Infer())
	return err
}

func (c *Chronicles) updateSubscriptions(tx *sql.Tx, mdb *sql.DB, ev *ChronicleEvent, uid string) error {
	if ev.ClientEventType != CR_EVENT_TYPE_PLAYER_PLAY {
		return nil
	}
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
		if s.CollectionUID.Valid {
			byCOs = append(byCOs, s)
		}
		if s.ContentType.Valid {
			byTypes = append(byTypes, s)
		}
	}

	query := `SELECT co.uid as co_uid, cu.type_id as type_id  FROM collections_content_units ccu
			INNER JOIN content_units cu ON ccu.content_unit_id = cu.id
			INNER JOIN collections co ON ccu.collection_id = co.id
			WHERE cu.uid = ?`

	rows, err := queries.Raw(query, uid).Query(mdb)
	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return err
	}
	defer rows.Close()
	forUpdate := models.SubscriptionSlice{}
	for rows.Next() {
		var (
			type_id int
			co_uid  string
		)
		err = rows.Scan(&type_id, &co_uid)
		if err != nil {
			return err
		}
		ct := utils.ContentTypesByID[type_id]
		for _, s := range byTypes {
			if s.ContentType.String == ct.Name {
				forUpdate = append(forUpdate, s)
			}
		}
		for _, s := range byCOs {
			if s.CollectionUID.String == co_uid {
				forUpdate = append(forUpdate, s)
			}
		}
	}
	if len(forUpdate) == 0 {
		return nil
	}

	col := make(map[string]interface{}, 0)
	col["updated_at"] = time.Now()
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
	return &null.JSON{JSON: dStr}, nil
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
