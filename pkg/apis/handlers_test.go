package apis

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/augurysys/timestamp"
	"github.com/go-chi/chi/v5"
	"github.com/oxtyped/gpodder2go/pkg/data"
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

// TestHandleUpdateSubscription tests for the update subscription endpoint to
// ensure that when a subscription is added on one device, it is also added onto
// the other sync devices
func TestHandleUpdateSubscription(t *testing.T) {

	var (
		subUrl string = "https://rubbishurl.com"
	)
	dataInterface := data.NewSQLite("testme.db")
	db := dataInterface.GetDB()
	cleanup(t, db)

	username := "username"
	err := dataInterface.AddUser(username, "pass", "test@test.com", "name")
	if err != nil {
		t.Fatal(err)
	}

	_, err = dataInterface.AddDevice(username, "device1", "", "laptop")
	if err != nil {
		t.Fatal(err)
	}
	_, err = dataInterface.AddDevice(username, "device2", "", "laptop")
	if err != nil {
		t.Fatal(err)
	}

	err = dataInterface.AddSyncGroup([]string{"device1", "device2"}, username)
	if err != nil {
		t.Fatalf("error adding sync group: %#v", err)
	}

	syncChanges := &SubscriptionChanges{
		Add:       []string{subUrl},
		Timestamp: timestamp.Now(),
	}

	body, err := json.Marshal(syncChanges)
	if err != nil {
		t.Fatal(err)
	}

	subscriptionAPI := SubscriptionAPI{Data: dataInterface}
	path := fmt.Sprintf("/api/2/subscriptions/%s/%s.json", username, "device1")
	m := chi.NewRouter()
	m.Post("/api/2/subscriptions/{username}/{deviceid}.{format}", subscriptionAPI.HandleUploadDeviceSubscriptionChange)
	ts := httptest.NewServer(m)

	reqBody := bytes.NewBuffer(body)
	req, err := http.NewRequest("POST", ts.URL+path, reqBody)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	// Get Subscription from
	sub1, err := dataInterface.RetrieveSubscriptionHistory("username", "device1", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	sub2, err := dataInterface.RetrieveSubscriptionHistory("username", "device2", time.Time{})
	if err != nil {
		t.Fatal(err)
	}

	if sub1[0].Podcast != subUrl {
		t.Fatalf("sub1 podcast url should be the one defined in the test but instead is %#v", sub1[0].Podcast)
	}

	if !reflect.DeepEqual(sub1, sub2) {
		t.Fatal("sub1 and sub2 is not equal")

	}

	if status := resp.StatusCode; status != http.StatusOK {
		t.Errorf("expecting handler to be ok but instead got: %#v", status)
	}

}

func TestHandlePostSync(t *testing.T) {

	// test happy path
	dataInterface := data.NewSQLite("testme.db")
	db := dataInterface.GetDB()
	username := "username"

	cleanup(t, db)

	err := dataInterface.AddUser(username, "pass", "test@test.com", "name")
	if err != nil {
		t.Fatal(err)
	}

	_, err = dataInterface.AddDevice(username, "device1", "", "laptop")
	if err != nil {
		t.Fatal(err)
	}
	_, err = dataInterface.AddDevice(username, "device2", "", "laptop")
	if err != nil {
		t.Fatal(err)
	}
	_, err = dataInterface.AddDevice(username, "device3", "", "laptop")
	if err != nil {
		t.Fatal(err)
	}

	verifierSecretKey := "itsatest"
	syncAPI := NewSyncAPI(dataInterface, verifierSecretKey)

	var synchronizedDevices [][]string

	dd := []string{"device1", "device2", "device3"}
	synchronizedDevices = append(synchronizedDevices, dd)

	syncReq := &SyncDeviceRequest{}
	syncReq.Synchronize = synchronizedDevices

	body, err := json.Marshal(syncReq)
	if err != nil {
		t.Fatal(err)
	}

	//handler := http.Handler(http.HandlerFunc(syncAPI.HandlePostSync))

	path := fmt.Sprintf("/api/2/sync-devices/%s.json", username)
	m := chi.NewRouter()
	m.Post("/api/2/sync-devices/{username}.json", syncAPI.HandlePostSync)
	ts := httptest.NewServer(m)

	reqBody := bytes.NewBuffer(body)
	req, err := http.NewRequest("POST", ts.URL+path, reqBody)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%#v", string(resBody))

	if status := resp.StatusCode; status != http.StatusOK {
		t.Errorf("expecting handler to be ok but instead got: %#v", status)
	}

}
