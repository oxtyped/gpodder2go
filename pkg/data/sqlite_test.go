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

	device := Device{Name: "testdevice", Type: "laptop", User: user}

	// doit
	timenow := CustomTimestamp{}
	timenow.Time = time.Now()

	e := EpisodeAction{
		Podcast:   "http://podcast.com/rss.xml",
		Episode:   "episode 232",
		Device:    device.Name,
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
		Device:    "testdevice",
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
