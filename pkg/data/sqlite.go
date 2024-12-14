package data

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/oxtyped/go-opml/opml"
	"github.com/pkg/errors"
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

func (s *SQLite) GetDB() *sql.DB {
	return s.db

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

// AddDevice creates a new Device and returns the id of the device and any error
func (s *SQLite) AddDevice(username string, deviceName string, caption string, deviceType string) (int, error) {

	var deviceId int

	db := s.db
	userId, err := s.GetUserIdFromName(username)
	if err != nil {
		log.Printf("error getting user id from name: %#v", err)
		return 0, err
	}

	err = db.QueryRow("INSERT INTO devices (user_id, name, type, caption) VALUES (?, ?, ?, ?) RETURNING id;", userId, deviceName, deviceType, caption).Scan(&deviceId)
	if err != nil {
		return 0, err
	}

	return deviceId, nil
}

func (s *SQLite) UpdateOrCreateDevice(username string, deviceName string, caption string, deviceType string) (int, error) {
	var deviceId int

	db := s.db
	userId, err := s.GetUserIdFromName(username)
	if err != nil {
		log.Printf("error getting user id from name: %#v", err)
		return 0, err
	}

	err = db.QueryRow("INSERT INTO devices (user_id, name, type, caption) VALUES (?,?,?,?) ON CONFLICT(user_id, name) DO UPDATE SET user_id=excluded.user_id,name=excluded.name,type=excluded.type,caption=excluded.caption  RETURNING id;", userId, deviceName, deviceType, caption).Scan(&deviceId)
	if err != nil {
		return 0, err
	}

	return deviceId, nil
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

	_, err := db.Exec("INSERT INTO episode_actions(device_id, podcast, episode, action, position, started, total, timestamp) VALUES (?,?,?,?,?,?,?,?)", e.Device, e.Podcast, e.Episode, e.Action, e.Position, e.Started, e.Total, e.Timestamp.Unix())
	if err != nil {
		return err
	}
	return nil
}

func (l *SQLite) RetrieveEpisodeActionHistory(username string, deviceId string, since time.Time) ([]EpisodeAction, error) {
	db := l.db

	actions := []EpisodeAction{}

	query := "SELECT a.podcast, a.episode, a.device_id, a.action, a.position, a.started, a.total, a.timestamp"
	query = query + " FROM episode_actions as a, devices as d, users as u"
	query = query + " WHERE a.device_id = d.name AND d.user_id = u.id AND u.username = ? ORDER BY a.id"
	args := []interface{}{username}
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
		if err != nil {
			log.Printf("error scanning: %#v", err)
			continue
		}
		timestamp := CustomTimestamp{}
		g, err := strconv.ParseInt(ts, 10, 64)
		if err != nil {
			log.Printf("error scanning: %#v", err)
			continue
		}
		timestamp.Time = time.Unix(g, 0)
		a.Timestamp = timestamp

		// For some reason, the timestamp for episode actions has been stored as
		// an integer, but in a field of type varchar(255). (that's why you see
		// the ParseInt call above). So we cannot use a DB query to sort or
		// filter by timestamp, because the DB would sort the integers
		// alphabetically. Someone should probably fix that by making a
		// migration to change the timestamp type in the DB, and then let the DB
		// handle the filtering.
		if !since.IsZero() {
			if a.Timestamp.Before(since) {
				continue
			}
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
	defer rows.Close()

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

// AddSubscriptionHistory adds and updates subscription
func (s *SQLite) AddSubscriptionHistory(sub Subscription) error {
	db := s.db

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	username := sub.User
	devices := sub.Devices

	userId, err := s.GetUserIdFromName(username)
	if err != nil {
		return errors.Wrap(err, "error getting user_id from name")
	}
	//deviceId, err := s.GetDeviceIdFromName(deviceName, username)
	//if err != nil {
	//	return errors.Wrap(err, "errors getting device_id from name")
	//}

	timestamp := strconv.FormatInt(sub.Timestamp.Unix(), 10)
	// Check  if a corresponding podcast exists
	for _, deviceId := range devices {
		// transaction! change to transaction
		_, err := tx.Exec("INSERT INTO subscriptions (user_id, device_id, podcast, action, timestamp) VALUES(?,?,?,?,?)", userId, deviceId, sub.Podcast, sub.Action, timestamp)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// RetrieveAllDeviceSubscriptionsSlice takes in a username and returns a slice
// containing the url of all the subscribed podcasts
func (s *SQLite) RetrieveAllDeviceSubscriptionsSlice(username string) ([]string, error) {

	// retrieve all devices's Add
	// subset it, if it exists anywhere its added

	allDevices := []string{}

	// get all devices
	devices, err := s.GetDevicesFromUsername(username)
	if err != nil {
		return nil, err
	}

	log.Printf("Retrieving %d devices from user %s", len(devices), username)

	for _, v := range devices {
		subs, err := s.RetrieveSubscriptionHistory(username, v, time.Time{})
		if err != nil {
			log.Printf("error retrieving subscription history: %#v", err)

			return nil, err
		}

		// calculate what's the diff
		add, _ := SubscriptionDiff(subs)
		allDevices = append(allDevices, add...)
	}

	return allDevices, nil

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

// RetrieveDeviceSubscriptionsSlice takes in a username and devicename and returns
// a slice of all the urls of the subscribed podcasts on the device
func (s *SQLite) RetrieveDeviceSubscriptionsSlice(username string, deviceName string) ([]string, error) {
	subs, err := s.RetrieveSubscriptionHistory(username, deviceName, time.Time{})
	if err != nil {
		log.Printf("error retrieving subscription history: %#v", err)

		return nil, err
	}

	// calculate what's the diff
	add, _ := SubscriptionDiff(subs)

	return add, nil
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
		go func() {
			err := o.AddRSSFromURL(v, 2*time.Second)
			if err != nil {
				log.Printf("error adding RSS feed from URL: %#v", err)
			}
		}()
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
	defer rows.Close()

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

// GetDevicesInSyncGroupFromDeviceId takes in a deviceId and returns a list of
// deviceIds that belongs to the same syncgroup including itself. If there's no
// sync group, it should return a nil to signal that the device has no existing
// sync_group, but not return an error as it is returning a valid position.
func (s *SQLite) GetDevicesInSyncGroupFromDeviceId(deviceId int) ([]int, error) {

	var deviceSyncGroupId sql.NullInt64
	var returnedDeviceId int
	var devicesFromSyncGroup []int
	db := s.db

	// get the sync group id
	err := db.QueryRow("select device_sync_group_id from devices where id =? LIMIT 1", deviceId).Scan(&deviceSyncGroupId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			log.Printf("error getting device_sync_group_id: %#v", err)
			return nil, err
		}
	}

	rows, err := db.Query("select id from devices WHERE device_sync_group_id = ?", deviceSyncGroupId)
	if err != nil {
		log.Printf("error getting devices from sync group: %#v", err)
		return nil, err
	}

	for rows.Next() {
		rows.Scan(&returnedDeviceId)
		devicesFromSyncGroup = append(devicesFromSyncGroup, returnedDeviceId)

	}

	return devicesFromSyncGroup, nil
}

// AddSyncGroup takes in a slice of device_ids and add them to a SyncGroup if
// they do not exist. Returns error if it already existed.
//
// Each device can only have one sync group at a time. Sync group's purpose is
// to link devices together so that any new podcast subscriptions will always be
// setup and installed on each device
func (s *SQLite) AddSyncGroup(device_names []string, username string) error {

	var device_ids []int
	var userId int

	db := s.db

	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	err = tx.QueryRow("select id from users where username = ?", username).Scan(&userId)
	if err != nil {
		log.Printf("error retrieving user info: %#v", err)
		return err
	}

	for _, v := range device_names {
		i, err := s.GetDeviceIdFromName(v, username)
		if err != nil {
			return err
		}

		device_ids = append(device_ids, i)
	}

	fmt.Printf("device_ids is %#v", device_ids)
	// get device_ids

	// do a check if device_ids all belong to the user. If it doesn't, send out an
	// error
	// TODO FIX THIS
	//	return errors.New("something screwed up")

	// getDeviceSyncGroup takes in a deviceId and returns the device_sync_group_id
	// of a device if it exists, if not it returns nil
	var getDeviceSyncGroup = func(deviceId int) *int {

		var deviceSyncGroupId *int

		err := tx.QueryRow("SELECT device_sync_group_id from devices WHERE id = ? AND user_id = ?", deviceId, userId).Scan(&deviceSyncGroupId)
		if err != nil {

			log.Printf("error selecting sync group id: %#v", err)
			return nil
		}

		return deviceSyncGroupId
	}

	// updateDeviceSyncGroup takes in a device_id and a device_sync_group_id to
	// update the device with the new sync_group_id
	var updateDeviceSyncGroup = func(deviceId *int, newSyncGroup *int) error {

		if deviceId == nil {
			return errors.New("no device provided")

		}

		result, err := tx.Exec("UPDATE devices SET device_sync_group_id = ? WHERE id = ?", newSyncGroup, deviceId)
		if err != nil {
			return err
		}

		if affected, _ := result.RowsAffected(); affected == 0 {

			return fmt.Errorf("expecting device to be updated with new device_sync_group but none is changed; is there a device_id %d", deviceId)
		}
		return nil

	}

	// createDeviceSyncGroup takes in 2 device_ids and creates a syncgroup linking
	// both of the devices together
	var createDeviceSyncGroup = func(firstDeviceId *int, secondDeviceId *int) error {

		var lastInsertId *int

		err = tx.QueryRow("insert into device_sync_groups (sync_status) VALUES ('pending') RETURNING id").Scan(&lastInsertId)
		if err != nil {
			return err
		}

		err = updateDeviceSyncGroup(firstDeviceId, lastInsertId)
		if err != nil {
			return err
		}
		err = updateDeviceSyncGroup(secondDeviceId, lastInsertId)
		if err != nil {
			return err
		}

		return nil
	}

	// take the first device as the first device
	firstDeviceId := device_ids[0]

	for _, currentDeviceId := range device_ids[1:] {

		currentDeviceSyncGroupId := getDeviceSyncGroup(currentDeviceId)
		firstDeviceSyncGroupId := getDeviceSyncGroup(firstDeviceId)

		// if both devices belong to a different sync group, assign the
		// sync group in the current_device to the other sync group
		if currentDeviceSyncGroupId != nil && firstDeviceSyncGroupId != nil {

			// update deviceSyncGroup id with firstDeviceSyncGroup id
			err = updateDeviceSyncGroup(&currentDeviceId, firstDeviceSyncGroupId)
			if err != nil {
				return errors.Wrapf(err, "error updating device %d sync group with %d", currentDeviceId, firstDeviceSyncGroupId)
			}

		} else if currentDeviceSyncGroupId == nil && firstDeviceSyncGroupId == nil {
			// if both devices have no sync groups, create one and assign it to both
			// of them

			log.Printf("no sync groups found, creating a new one")
			err = createDeviceSyncGroup(&firstDeviceId, &currentDeviceId)
			if err != nil {
				return errors.Wrapf(err, "error creating a new device sync group for device id %d and %d", firstDeviceId, currentDeviceId)
			}

		} else if currentDeviceSyncGroupId == nil && firstDeviceSyncGroupId != nil {
			// if current_device has no sync group but there's an existing sync group
			// on the other device, assign the sync group to current_device

			err = updateDeviceSyncGroup(&currentDeviceId, firstDeviceSyncGroupId)
			if err != nil {
				return err
			}

		} else if currentDeviceSyncGroupId != nil && firstDeviceSyncGroupId == nil {
			// if current_device has a sync group but the other device has no sync
			// group, assign the current_device's sync group to the other device

			err = updateDeviceSyncGroup(&firstDeviceId, currentDeviceSyncGroupId)
			if err != nil {
				return err
			}

		}

	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil

}

// StopDeviceSync takes in a device name and username to stop device sync
func (s *SQLite) StopDeviceSync(deviceName string, username string) error {
	db := s.db

	_, err := db.Exec("UPDATE devices SET device_sync_group_id = NULL WHERE name = ? AND user_id = (SELECT id from users WHERE username = ?)", deviceName, username)
	return err

}

// GetDeviceSyncGroupIds takes in a username and returns a SyncStatus and error
func (s *SQLite) GetDeviceSyncGroupIds(username string) ([]int, error) {
	var ids []int

	db := s.db

	rows, err := db.Query("select DISTINCT device_sync_group_id from devices WHERE user_id = (select id from users where username = ?);", username)
	if err != nil {
		return ids, err
	}

	defer rows.Close()

	for rows.Next() {
		var id *int
		if err := rows.Scan(&id); err != nil {
			return ids, err
		}

		if id != nil {
			ids = append(ids, *id)
		}
	}

	return ids, nil
}

// GetDeviceNameFromDeviceSyncGroupId takes in a device_sync_group_id and
// returns a string slice of device names that belongs to the group
func (s *SQLite) GetDeviceNameFromDeviceSyncGroupId(id int) ([]string, error) {
	var names []string
	db := s.db

	rows, err := db.Query("select name from devices where device_sync_group_id = ?;", id)
	if err != nil {
		return names, err
	}

	for rows.Next() {

		var name string
		if err := rows.Scan(&name); err != nil {

			return names, err
		}

		names = append(names, name)

	}

	return names, nil

}

func (s *SQLite) GetNotSyncedDevices(username string) ([]string, error) {
	var devices []string

	db := s.db

	userId, err := s.GetUserIdFromName(username)
	if err != nil {
		return nil, err
	}
	deviceGroupIds, err := s.GetDeviceSyncGroupIds(username)
	if err != nil {
		return nil, err
	}

	intValues := []int{}
	for _, v := range deviceGroupIds {
		intValues = append(intValues, v)
	}

	query, args, err := sqlx.In("select name from devices where user_id = ? AND id NOT IN (select id FROM devices WHERE device_sync_group_id IN (?));", userId, intValues)

	rows, err := db.Query(query, args...)
	if err != nil {
		return devices, err
	}

	for rows.Next() {

		var name string
		if err := rows.Scan(&name); err != nil {
			fmt.Printf("error scanning rows: %#v", err)
			return devices, err
		}

		devices = append(devices, name)

	}

	return devices, nil

}
