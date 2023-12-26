package data

import (
	"database/sql"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/oxtyped/go-opml/opml"
	"github.com/pkg/errors"
	"github.com/relvacode/iso8601"
	_ "modernc.org/sqlite"
)

type SQLite struct {
	db *sql.DB
}

func NewSQLite(file string) *SQLite {
	db, err := sql.Open("sqlite", file)
	if err != nil {
		panic("failed to connect database")
	}

	return &SQLite{db: db}
}

func (s *SQLite) GetUserIdFromName(username string) (int, error) {
	var userId int
	db := s.db

	err := db.QueryRow("SELECT id from users WHERE username = ?", username).Scan(&userId)
	if err != nil {
		return userId, err
	}

	return userId, nil

}

func (s *SQLite) CheckUserPassword(username, password string) bool {

	var count int
	db := s.db
	err := db.QueryRow("SELECT count(*) from users WHERE username = ? AND password = ?", username, password).Scan(&count)
	if err != nil {
		return false
	}

	if count == 1 {
		return true
	}

	// returning false here to explicitly ensure that only if user is checked that
	// it returns true
	return false
}

func (s *SQLite) AddUser(username, password, email, name string) error {

	db := s.db
	_, err := db.Exec("INSERT INTO users (username, password, email, name) VALUES ($1, $2, $3, $4)", username, password, email, name)
	if err != nil {
		return err
	}
	return nil

}
func (s *SQLite) AddDevice(username string, deviceName string, caption string, deviceType string) error {

	db := s.db
	userId, err := s.GetUserIdFromName(username)
	if err != nil {
		log.Printf("error getting user id from name: %#v", err)
		return err
	}

	_, err = db.Exec("INSERT INTO devices (user_id, name, type, caption) VALUES (?, ?, ?, ?);", userId, deviceName, deviceType, caption)
	if err != nil {
		return err
	}

	return nil

}

func (s *SQLite) RetrieveDevices(username string) ([]Device, error) {

	db := s.db
	data := []Device{}

	userId, err := s.GetUserIdFromName(username)
	if err != nil {
		return nil, errors.Wrap(err, "error getting user id from name")
	}

	rows, err := db.Query("SELECT name, type, caption from devices WHERE user_id = ?", userId)
	if err != nil {
		return nil, errors.Wrap(err, "error getting devices from user")
	}
	defer rows.Close()

	for rows.Next() {
		i := Device{}
		err := rows.Scan(&i.Name, &i.Type, &i.Caption)
		if err != nil {
			return nil, errors.Wrap(err, "error scanning devices from query")
		}

		data = append(data, i)
	}

	return data, nil

}

func (l *SQLite) AddEpisodeActionHistory(username string, e EpisodeAction) error {
	db := l.db

	deviceId, err := l.GetDeviceIdFromName(e.Device, username)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO episode_actions(device_id, podcast, episode, action, position, started, total, timestamp) VALUES (?,?,?,?,?,?,?,?)", deviceId, e.Podcast, e.Episode, e.Action, e.Position, e.Started, e.Total, e.Timestamp.Unix())
	if err != nil {
		return err
	}
	return nil
}

func (l *SQLite) RetrieveEpisodeActionHistory(username string, device string, since string) ([]EpisodeAction, error) {
	db := l.db

	actions := []EpisodeAction{}

	query := "SELECT a.podcast, a.episode, d.name, a.action, a.position, a.started, a.total, a.timestamp"
	query = query + " FROM episode_actions as a, devices as d, users as u"
	query = query + " WHERE a.device_id = d.id AND d.user_id = u.id AND u.name = ?"
	var args []interface{}
	args = append(args, username)
	if device != "" {
		query = query + " AND d.name = ?"
		args = append(args, device)
	}
	if since != "" {
		parsedTs, err := iso8601.ParseString(since)
		if err != nil {
			return nil, err
		}
		query = query + " AND a.timestamp > ?"
		args = append(args, parsedTs.Unix())
	}
	query = query + " ORDER BY a.id"
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		a := EpisodeAction{}
		var ts string
		err := rows.Scan(
			&a.Podcast,
			&a.Episode,
			&a.Device,
			&a.Action,
			&a.Position,
			&a.Started,
			&a.Total,
			&ts,
		)
		timestamp := CustomTimestamp{}
		g, _ := strconv.ParseInt(ts, 10, 64)
		timestamp.Time = time.Unix(g, 0)
		a.Timestamp = timestamp
		if err != nil {
			log.Printf("error scanning: %#v", err)
			continue
		}

		actions = append(actions, a)

	}

	return actions, nil
}

// GetDevicesFromUsername returns a list of device names that belongs to
// username
func (s *SQLite) GetDevicesFromUsername(username string) ([]string, error) {
	db := s.db

	devices := []string{}

	rows, err := db.Query("select devices.name from devices, users WHERE devices.user_id = users.id AND users.username = ?", username)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var s string
		err := rows.Scan(&s)
		if err != nil {
			log.Printf("error scanning: %#v", err)
			continue
		}

		devices = append(devices, s)

	}

	return devices, nil

}

func (s *SQLite) GetDeviceIdFromName(deviceName string, username string) (int, error) {
	var deviceId int

	userId, err := s.GetUserIdFromName(username)
	if err != nil {
		return deviceId, err
	}

	db := s.db
	err = db.QueryRow("SELECT id from devices WHERE name = ? AND user_id = ?", deviceName, userId).Scan(&deviceId)
	if err != nil {
		return deviceId, err
	}

	return deviceId, nil

}

func (s *SQLite) AddSubscriptionHistory(sub Subscription) error {

	db := s.db

	username := sub.User
	deviceName := sub.Device

	userId, err := s.GetUserIdFromName(username)
	if err != nil {
		return errors.Wrap(err, "error getting user_id from name")
	}
	deviceId, err := s.GetDeviceIdFromName(deviceName, username)
	if err != nil {
		return errors.Wrap(err, "errors getting device_id from name")
	}

	timestamp := strconv.FormatInt(sub.Timestamp.Unix(), 10)
	// Check  if a corresponding podcast exists
	_, err = db.Exec("INSERT INTO subscriptions (user_id, device_id, podcast, action, timestamp) VALUES (?,?,?,?,?)", userId, deviceId, sub.Podcast, sub.Action, timestamp)
	if err != nil {
		return err
	}

	return nil
}

// RetrieveAllDeviceSubscriptions takes in a username and returns an OPML file
// of all the RSS feeds that the user was subscribed to on the platform
func (s *SQLite) RetrieveAllDeviceSubscriptions(username string) (string, error) {

	// retrieve all devices's Add
	// subset it, if it exists anywhere its added

	allDevices := []string{}

	// get all devices
	devices, err := s.GetDevicesFromUsername(username)
	if err != nil {
		return "", err
	}

	log.Printf("Retrieving %d devices from user %s", len(devices), username)

	for _, v := range devices {
		subs, err := s.RetrieveSubscriptionHistory(username, v, time.Time{})
		if err != nil {
			log.Printf("error retrieving subscription history: %#v", err)

			return "", err
		}

		// calculate what's the diff
		add, _ := SubscriptionDiff(subs)
		allDevices = append(allDevices, add...)
	}
	allDevices = unique(allDevices)

	var wg sync.WaitGroup

	o := opml.NewOPMLFromBlank("tmpfile")
	for _, v := range allDevices {
		v := v

		wg.Add(1)
		go func() {
			defer wg.Done()

			o.AddRSSFromURL(v, 2*time.Second)
		}()
	}

	wg.Wait()

	return o.XML()

}

// RetrieveDeviceSubscriptions takes in a username and devicename and returns
// the OPML of its subscriptions
func (s *SQLite) RetrieveDeviceSubscriptions(username string, deviceName string) (string, error) {

	subs, err := s.RetrieveSubscriptionHistory(username, deviceName, time.Time{})
	if err != nil {
		log.Printf("error retrieving subscription history: %#v", err)

		return "", err
	}

	// calculate what's the diff
	add, _ := SubscriptionDiff(subs)

	o := opml.NewOPMLFromBlank("tmpfile")
	for _, v := range add {
		err := o.AddRSSFromURL(v, 2*time.Second)
		if err != nil {
			log.Printf("error adding RSS feed from URL: %#v", err)
		}
	}

	return o.XML()
}

func (s *SQLite) RetrieveSubscriptionHistory(username string, deviceName string, since time.Time) ([]Subscription, error) {

	db := s.db
	userId, err := s.GetUserIdFromName(username)
	if err != nil {
		log.Printf("unable to find user id from username: %#v", err)
		return nil, err
	}
	deviceId, err := s.GetDeviceIdFromName(deviceName, username)
	if err != nil {
		log.Printf("unable to find device id from device name: %#v", err)
		return nil, err
	}
	subscriptions := []Subscription{}

	rows, err := db.Query("select podcast, action, timestamp from subscriptions where user_id = ? AND device_id = ? AND timestamp > ? ", userId, deviceId, strconv.FormatInt(since.Unix(), 10))
	if err != nil {
		log.Printf("error selecting rows: %#v", err)
		return nil, err
	}

	for rows.Next() {
		sub := Subscription{}
		var ts string
		err := rows.Scan(&sub.Podcast, &sub.Action, &ts)
		if err != nil {
			log.Printf("error scanning rows into struct: %#v", err)
			continue
		}

		timestampTime, err := strconv.ParseInt(ts, 10, 64)
		if err != nil {
			log.Printf("error parsing timestamp into struct: %#v", err)
			continue
		}
		sub.Timestamp.Time = time.Unix(timestampTime, 0)

		subscriptions = append(subscriptions, sub)

	}

	return subscriptions, nil

}

func unique(stringSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
