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
	"testing"

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
