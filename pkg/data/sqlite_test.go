package data

import (
	"database/sql"
	"testing"
	"time"
)

func cleanup(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM users")
	if err != nil {
		t.Error(err)
	}
	_, err = db.Exec("DELETE FROM devices")
	if err != nil {
		t.Error(err)
	}

	_, err = db.Exec("DELETE FROM subscriptions")
	if err != nil {
		t.Error(err)
	}

	_, err = db.Exec("DELETE FROM episode_actions")
	if err != nil {
		t.Error(err)
	}
	_, err = db.Exec("DELETE FROM device_sync_groups")
	if err != nil {
		t.Error(err)
	}
	_, err = db.Exec("DELETE FROM device_sync_group_devices")
	if err != nil {
		t.Error(err)
	}
}

// Test
func TestAddEpisodeHistory(t *testing.T) {
	data := NewSQLite("testme.db")
	db := data.db

	cleanup(t, db)

	// setup
	_, err := db.Exec("INSERT INTO users (id, username, password, email, name) VALUES (1, 'test', 'test1024', 'test@test.com', 'somename')")
	if err != nil {
		t.Fatalf("error setting up user: %#v", err)
	}
	username := "test"

	_, err = db.Exec("INSERT INTO devices (id, user_id, name,  type) VALUES (1,1,'testdevice','laptop');")
	if err != nil {
		t.Fatalf("error setting up devices: %#v", err)
	}

	user := &User{Name: "test"}

	device := Device{Id: 1, Name: "testdevice", Type: "laptop", User: user}

	// doit
	timenow := CustomTimestamp{}
	timenow.Time = time.Now()

	e := EpisodeAction{
		Podcast:   "http://podcast.com/rss.xml",
		Episode:   "episode 232",
		Devices:   []int{device.Id},
		Action:    "PLAYING",
		Timestamp: timenow,
	}
	err = data.AddEpisodeActionHistory(username, e)
	if err != nil {
		t.Errorf("expecting no error adding episodeaction but got %#v", err)
	}
}

// Test all the various interface functions for adding of subscription!
func TestAddSubscriptionHistory(t *testing.T) {
	data := NewSQLite("testme.db")
	db := data.db

	cleanup(t, db)

	_, err := db.Exec("INSERT INTO users (id, username, password, email, name) VALUES (1, 'somename', 'test1024', 'test@test.com', 'somename')")
	if err != nil {
		t.Fatalf("error setting up user: %#v", err)
	}

	_, err = db.Exec("INSERT INTO devices (id, user_id, name,  type) VALUES (1,1,'testdevice','laptop');")
	if err != nil {
		t.Fatalf("error setting up devices: %#v", err)
	}

	s := Subscription{
		User:      "somename",
		Devices:   []int{1},
		Podcast:   "podcasturl",
		Action:    "SUBSCRIBE",
		Timestamp: CustomTimestamp{Time: time.Now()},
	}

	err = data.AddSubscriptionHistory(s)
	if err != nil {
		t.Fatalf("error adding subscription: %#v", err)
	}

	// setup run.
	// setup User table
	// setup Device table

	// Test that can pull the information
}

func TestUpdateOrCreateDevice(t *testing.T) {

	var deviceType string

	data := NewSQLite("testme.db")
	db := data.db

	cleanup(t, db)

	// setup user
	err := data.AddUser("username", "pass", "test@test.com", "name")
	if err != nil {
		t.Fatal(err)
	}

	deviceId, err := data.UpdateOrCreateDevice("username", "device1", "", "laptop")
	if err != nil {
		t.Fatalf("error updating or creating device: %#v", err)
	}

	updatedDeviceId, err := data.UpdateOrCreateDevice("username", "device1", "", "desktop")
	if err != nil {
		t.Fatalf("error updating or creating device: %#v", err)
	}

	if updatedDeviceId != deviceId {
		t.Fatalf("expecting deviceId to be %#v after upsert, but got %#v", updatedDeviceId, deviceId)
	}

	err = db.QueryRow("select type from devices WHERE id = ?;", updatedDeviceId).Scan(&deviceType)
	if err != nil {
		t.Fatalf("error selecting name: %#v", err)
	}

	if deviceType != "desktop" {
		t.Fatalf("expecting device type to be desktop but got %#v", deviceType)

	}

}

// Test the Sync Group
func TestAddSyncGroup(t *testing.T) {

	data := NewSQLite("testme.db")
	db := data.db

	cleanup(t, db)

	// setup user
	err := data.AddUser("username", "pass", "test@test.com", "name")
	if err != nil {
		t.Fatal(err)
	}

	_, err = data.AddDevice("username", "device1", "", "laptop")
	if err != nil {
		t.Fatal(err)
	}
	_, err = data.AddDevice("username", "device2", "", "laptop")
	if err != nil {
		t.Fatal(err)
	}
	_, err = data.AddDevice("username", "device3", "", "laptop")
	if err != nil {
		t.Fatal(err)
	}

	// setup device 1, 2 and 3

	err = data.AddSyncGroup([]string{"device1", "device2", "device3"}, "username")
	if err != nil {
		t.Fatal(err)
	}

	var id int
	// test that there's a new sync group and a new sync devices
	err = db.QueryRow("SELECT id from device_sync_groups LIMIT 1").Scan(&id)
	if err != nil {
		t.Fatalf("expecting to have a result but instead got sql error: %#v", err)
	}

	err = db.QueryRow("select COUNT(*) from devices WHERE device_sync_group_id = ?", id).Scan(&id)
	if err != nil {
		t.Fatalf("error querying row: %#v", err)
	}

	if id != 3 {
		t.Errorf("expecting id to be 3 but got %#v", id)
	}
}
