package chronicles

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	pkgerr "github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/Bnei-Baruch/archive-my/common"
	"github.com/Bnei-Baruch/archive-my/databases/mdb"
	"github.com/Bnei-Baruch/archive-my/databases/mydb/models"
	"github.com/Bnei-Baruch/archive-my/pkg/sqlutil"
	"github.com/Bnei-Baruch/archive-my/pkg/utils"
	"github.com/Bnei-Baruch/archive-my/version"
)

const (
	SCAN_SIZE     = 100
	MAX_INTERVAL  = time.Minute
	MIN_INTERVAL  = 100 * time.Millisecond
	WAIT_FOR_SAVE = 5 * time.Minute
)

type Chronicles struct {
	ticker   *time.Ticker
	interval time.Duration

	lastReadId  string
	prevReadId  string
	nextRefresh time.Time

	MyDB             *sql.DB
	MDB              *sql.DB
	chroniclesClient *http.Client
}

type ScanRequest struct {
	Id    string `json:"id,omitempty"`
	Limit int    `json:"limit"`

	// Filters.
	EventTypes []string `json:"event_types,omitempty"`
	UserIds    []string `json:"user_ids,omitempty"`
	Namespaces []string `json:"namespaces,omitempty"`
	Keycloak   bool     `json:"keycloak"`
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

func (c *Chronicles) Init() {
	mydbConn, err := sql.Open("postgres", common.Config.MyDBUrl)
	utils.Must(err)

	mdbConn, err := sql.Open("postgres", common.Config.MDBUrl)
	utils.Must(err)

	c.InitWithDeps(mydbConn, mdbConn)
}

func (c *Chronicles) InitWithDeps(mydbConn, mdbConn *sql.DB) {
	c.MyDB = mydbConn

	c.MDB = mdbConn
	utils.Must(mdb.InitCT(c.MDB))

	c.chroniclesClient = &http.Client{
		Timeout: 100 * time.Second,
	}
}

func (c *Chronicles) Run() {
	c.interval = MIN_INTERVAL
	c.ticker = time.NewTicker(MIN_INTERVAL)

	var err error
	c.lastReadId, err = c.lastChroniclesId()
	utils.Must(err)

	go func() {
		refresh := func() {
			if err := c.refresh(); err != nil {
				log.Error().Err(err).Msg("refresh error")
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

func (c *Chronicles) Shutdown() {
	c.MyDB.Close()
	c.MDB.Close()
	c.chroniclesClient.CloseIdleConnections()
}

func (c *Chronicles) lastChroniclesId() (string, error) {
	var lastID null.String
	err := models.NewQuery(
		qm.Select(fmt.Sprintf("MAX('%s')", models.HistoryColumns.ChroniclesID)),
		qm.From(models.TableNames.History),
	).QueryRow(c.MyDB).Scan(&lastID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return lastID.String, err
}

func (c *Chronicles) refresh() error {
	n, err := c.scanEvents()
	if err != nil {
		return err
	}

	if n == SCAN_SIZE {
		c.interval = utils.MaxDuration(c.interval/2, MIN_INTERVAL)
	} else {
		c.interval = utils.MinDuration(c.interval*2, MAX_INTERVAL)
	}
	c.ticker.Reset(c.interval)

	return nil
}

func (c *Chronicles) scanEvents() (int, error) {
	resp, err := c.fetchEvents()
	if err != nil {
		return 0, pkgerr.Wrap(err, "fetch events from chronicles")
	}

	err = sqlutil.InTx(context.TODO(), c.MyDB, func(tx *sql.Tx) error {
		return c.saveEvents(tx, resp.Entries)
	})
	if err != nil {
		return 0, pkgerr.Wrap(err, "save chronicles events")
	}

	if len(resp.Entries) > 0 {
		c.lastReadId = resp.Entries[len(resp.Entries)-1].ID
	}

	return len(resp.Entries), nil
}

func (c *Chronicles) fetchEvents() (*ScanResponse, error) {
	log.Info().Msgf("fetching chronicles entries, last successful [%s]", c.lastReadId)

	payload := ScanRequest{
		Id:         c.lastReadId,
		Limit:      SCAN_SIZE,
		EventTypes: []string{"player-play", "player-stop"},
		Namespaces: []string{"archive"},
		Keycloak:   true,
	}
	payloadBytes, err := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, common.Config.ChroniclesUrl, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, pkgerr.Wrap(err, "http.NewRequest")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("archive-my/%s", version.Version))

	resp, err := c.chroniclesClient.Do(req)
	if err != nil {
		return nil, pkgerr.WithMessage(err, "http.Post")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, pkgerr.Errorf("response code %d for scan: %s.", resp.StatusCode, resp.Status)
	}

	defer func() {
		if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
			log.Warn().Err(err).Msg("error draining response body")
		}
		_ = resp.Body.Close()
	}()
	var scanResponse ScanResponse
	if err := json.NewDecoder(resp.Body).Decode(&scanResponse); err != nil {
		return nil, pkgerr.WithMessage(err, "json.Decode")
	}

	return &scanResponse, nil
}

func (c *Chronicles) saveEvents(tx *sql.Tx, events []*ChronicleEvent) error {
	usersLRU := make(map[string]*models.User)
	uniqEvents := make(map[string]*ChronicleEvent, 0)
	for _, x := range events {
		// skip too early events
		// TODO: is this comparison correct ? what about timezones ?
		// FIXME why data.current_time ? shouldn't it be timestamp ?
		if x.Data.CurrentTime.Float64 <= WAIT_FOR_SAVE.Minutes() {
			continue
		}

		// TODO: further optimize user lookup with some longer scope lru cache ?
		// for first time login users, we might be able to associate some of their anonymous history
		// with another dedicated scan of chronicles. (future milestone)
		user, ok := usersLRU[x.AccountId]
		if !ok {
			var err error
			user, err = models.Users(models.UserWhere.AccountsID.EQ(x.AccountId)).One(tx)
			if err != nil {
				if err == sql.ErrNoRows {
					continue // skip anonymous users
				}
				return pkgerr.Wrap(err, "lookup user in db")
			}
			usersLRU[x.AccountId] = user
		}

		k := fmt.Sprintf("%s_%s", x.AccountId, x.Data.UnitUID)
		uniqEvents[k] = x
	}

	for _, x := range uniqEvents {
		user := usersLRU[x.AccountId]
		if err := c.insertEvent(tx, x, user); err != nil {
			return err
		}
		if err := c.updateSubscriptions(tx, x, user); err != nil {
			return err
		}
	}

	return nil
}

func (c *Chronicles) insertEvent(tx *sql.Tx, ev *ChronicleEvent, user *models.User) error {
	nParams := make(map[string]interface{})
	nParams["current_time"] = ev.Data.CurrentTime
	data, err := json.Marshal(nParams)
	if err != nil {
		return pkgerr.Wrap(err, "json.Marshal data")
	}

	year, month, day := ev.CreatedAt.Date()
	timeZone, err := time.LoadLocation(ev.Data.TimeZone)
	if err != nil {
		return pkgerr.Wrapf(err, "time.LoadLocation [%s]", ev.Data.TimeZone)
	}

	// FIXME: given chronicles is immutable. Why not skip already processed events ?!

	sDay := time.Date(year, month, day, 0, 0, 0, 0, timeZone)
	eDay := sDay.Add(24 * time.Hour)
	history, err := models.Histories(
		models.HistoryWhere.UserID.EQ(user.ID),
		models.HistoryWhere.ContentUnitUID.EQ(null.StringFrom(ev.Data.UnitUID)),
		models.HistoryWhere.ChroniclesTimestamp.GT(sDay),
		models.HistoryWhere.ChroniclesTimestamp.LT(eDay),
	).One(tx)
	if err != nil {
		if err != sql.ErrNoRows {
			return pkgerr.Wrap(err, "lookup existing history record in db")
		}

		history = &models.History{
			UserID:              user.ID,
			ChroniclesID:        ev.ID,
			ChroniclesTimestamp: ev.CreatedAt,
			ContentUnitUID:      null.StringFrom(ev.Data.UnitUID),
			Data:                null.JSONFrom(data),
		}
		if err := history.Insert(tx, boil.Infer()); err != nil {
			return pkgerr.Wrap(err, "insert new history record to db")
		}
	}

	history.ChroniclesID = ev.ID
	history.ChroniclesTimestamp = ev.CreatedAt
	params, err := mergeData(history.Data, nParams)
	if err != nil {
		return pkgerr.Wrap(err, "merge chronicles data params")
	}
	history.Data = *params
	if _, err := history.Update(tx, boil.Infer()); err != nil {
		return pkgerr.Wrap(err, "update existing history record in db")
	}

	return nil
}

func (c *Chronicles) updateSubscriptions(tx boil.Executor, ev *ChronicleEvent, user *models.User) error {
	subs, err := models.Subscriptions(models.SubscriptionWhere.UserID.EQ(user.ID)).All(tx)
	if err != nil {
		return pkgerr.Wrap(err, "fetch subscriptions from db")
	}

	if len(subs) == 0 {
		return nil
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

	rows, err := queries.Raw(query).Query(c.MDB)
	if err != nil {
		return pkgerr.Wrap(err, "lookup CCUs in mdb")
	}
	defer rows.Close()

	forUpdate := models.SubscriptionSlice{}
	for rows.Next() {
		var coUID string
		var typeID int

		err = rows.Scan(&coUID, &typeID)
		if err != nil {
			return pkgerr.Wrap(err, "rows.Scan")
		}

		name := mdb.ContentTypesByID[typeID]
		for _, s := range byTypes {
			if s.ContentType.String == name {
				forUpdate = append(forUpdate, s)
			}
		}
		for _, s := range byCOs {
			if s.CollectionUID.String == coUID {
				forUpdate = append(forUpdate, s)
			}
		}
	}
	if err := rows.Err(); err != nil {
		return pkgerr.Wrap(err, "rows.Err")
	}

	if len(forUpdate) == 0 {
		return nil
	}

	_, err = forUpdate.UpdateAll(tx, map[string]interface{}{
		models.SubscriptionColumns.UpdatedAt: ev.CreatedAt,
	})
	return pkgerr.Wrap(err, "bulk update subscriptions in db")
}

func mergeData(data null.JSON, nd map[string]interface{}) (*null.JSON, error) {
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
