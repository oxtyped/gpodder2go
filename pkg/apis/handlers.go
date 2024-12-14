package apis

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/augurysys/timestamp"
	"github.com/go-chi/chi/v5"
	"k8s.io/utils/strings/slices"

	"github.com/oxtyped/gpodder2go/pkg/data"
)

type Pair struct {
	a, b interface{}
}

func (p Pair) String() string {
	output := fmt.Sprintf("[%q, %q]", p.a, p.b)
	return output
}

type PairArray struct {
	Pairs []Pair
}

func (p PairArray) String() string {
	astring := "["
	for idx, v := range p.Pairs {
		astring += v.String()
		if idx != (len(p.Pairs) - 1) {
			astring += ","
		}

	}
	astring += "]"

	return astring
}

// HandleLogin uses Basic Auth to check on a user's credentials and return a
// cookie session that will be used for subsequent calls
func (u *UserAPI) HandleLogin(w http.ResponseWriter, r *http.Request) {
	db := u.Data

	username, password, ok := r.BasicAuth()
	if !ok {
		w.WriteHeader(401)
		return
	}
	expire := time.Now().Add(2 * time.Minute)

	if !db.CheckUserPassword(username, password) {
		w.WriteHeader(401)
		return
	}

	mac := hmac.New(sha256.New, []byte(u.verifierSecretKey))
	mac.Write([]byte(username))
	sig := mac.Sum(nil)

	preEncoded := fmt.Sprintf("%s.%s", sig, username)

	hash := base64.StdEncoding.EncodeToString([]byte(preEncoded))

	cookie := http.Cookie{Name: "sessionid", Value: hash, Path: "/", SameSite: http.SameSiteLaxMode, Expires: expire}

	http.SetCookie(w, &cookie)
	w.WriteHeader(200)
}

// HandleUserCreate takes in a username and password AND must only be able to be
// run on the same instance as the API Server
func (u *UserAPI) HandleUserCreate(w http.ResponseWriter, r *http.Request) {
	// Takes in a form data of username and password
	err := r.ParseForm()
	if err != nil {
		log.Printf("error parsing form: %#v", err)
		w.WriteHeader(400)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")
	name := r.FormValue("name")
	err = u.Data.AddUser(username, password, email, name)
	if err != nil {
		log.Printf("error adding user: %#v", err)
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(201)
}

// DeviceAPI
func (d *DeviceAPI) HandleUpdateDevice(w http.ResponseWriter, r *http.Request) {
	// username
	// deviceid

	username := chi.URLParam(r, "username")
	deviceName := chi.URLParam(r, "deviceid")

	log.Printf("username is %s, deviceName is %s", username, deviceName)

	ddr := &DeviceDataRequest{}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading body from payload: %#v", err)
		w.WriteHeader(400)
		return
	}

	// onboard the new directory

	err = json.Unmarshal(payload, ddr)
	if err != nil {
		log.Printf("error decoding json payload: %#v", err)
		w.WriteHeader(400)
		return
	}

	log.Printf("DDR is %#v and %#v %#v", ddr, username, deviceName)

	_, err = d.Data.UpdateOrCreateDevice(username, deviceName, ddr.Caption, ddr.Type)
	if err != nil {
		log.Printf("error adding device: %#v", err)
		w.WriteHeader(400)
		return
	}

	w.WriteHeader(200)
}

func (d *DeviceAPI) HandleGetDevices(w http.ResponseWriter, r *http.Request) {
	type GetDevicesOutput struct {
		Id            string `json:"id"` // This is not to be confused with the database "Id". This is the DeviceId that will be referenced throughout the various calls
		Caption       string `json:"caption"`
		Type          string `json:"type"`
		Subscriptions int    `json:"subscriptions"`
	}

	var deviceSlice []GetDevicesOutput

	username := chi.URLParam(r, "username")
	devices, err := d.Data.RetrieveDevices(username)
	if err != nil {
		log.Printf("error retrieving devices: %#v", err)
		w.WriteHeader(400)
		return
	}

	for _, v := range devices {
		subs, err := d.Data.RetrieveSubscriptionHistory(username, v.Name, time.Time{})
		if err != nil {
			log.Printf("error retrieving subscription history: %#v", err)
			continue
		}

		// calculate what's the diff
		add, _ := data.SubscriptionDiff(subs)
		device := GetDevicesOutput{
			Id:            v.Name,
			Caption:       v.Caption,
			Type:          v.Type,
			Subscriptions: len(add),
		}

		deviceSlice = append(deviceSlice, device)

	}

	var devicesOutput []byte

	if len(deviceSlice) == 0 {
		devicesOutput = []byte("[]")
	} else {

		devicesOutput, err = json.Marshal(deviceSlice)
		if err != nil {
			log.Printf("error marshalling devices: %#v", err)
			w.WriteHeader(400)
			return
		}
	}

	w.Write(devicesOutput)
}

// TODO: Handle Device Subscription Change
func (s *SubscriptionAPI) HandleDeviceSubscriptionChange(w http.ResponseWriter, r *http.Request) {
	// username
	// deviceid
	// format
	w.WriteHeader(404)
}

// API Endpoint: GET /api/2/subscriptions/{username}/{deviceid}.json
func (s *SubscriptionAPI) HandleGetDeviceSubscriptionChange(w http.ResponseWriter, r *http.Request) {
	// username
	// deviceid
	// format
	username := chi.URLParam(r, "username")
	deviceId := chi.URLParam(r, "deviceid")
	format := chi.URLParam(r, "format")
	add := []string{}
	remove := []string{}

	since := r.URL.Query().Get("since")
	if since == "" {
		log.Println("error with since query params - expecting it not to be empty but got \"\"")
		w.WriteHeader(400)
		return

	}

	if format != "json" {
		log.Printf("error uploading device subscription changes as format is expecting JSON but got %#v", format)
		w.WriteHeader(400)
		return
	}

	subscriptionChanges := &SubscriptionChanges{
		Add:       add,
		Remove:    remove,
		Timestamp: timestamp.Now(),
	}

	db := s.Data
	var tm time.Time

	if since == "0" {
		tm = time.Unix(0, 0)
	} else {
		i, err := strconv.ParseInt(since, 10, 64)
		if err != nil {
			log.Printf("error parsing strconv: %#v", err)
			w.WriteHeader(400)
			return
		}

		tm = time.Unix(i, 0)
	}

	subs, err := db.RetrieveSubscriptionHistory(username, deviceId, tm)
	if err != nil {
		log.Printf("error retrieving subscription history: %#v", err)

		w.WriteHeader(400)
		return
	}

	add, remove = data.SubscriptionDiff(subs)

	subscriptionChanges.Add = add
	subscriptionChanges.Remove = remove

	outputPayload, err := json.Marshal(subscriptionChanges)
	if err != nil {
		log.Printf("error marshalling subscription changes into JSON string: %#v", err)
		w.WriteHeader(400)
		return
	}

	w.WriteHeader(200)
	w.Write(outputPayload)
}

// API Endpoint: POST /api/2/subscriptions/{username}/{deviceid}.{format}
func (s *SubscriptionAPI) HandleUploadDeviceSubscriptionChange(w http.ResponseWriter, r *http.Request) {
	// username
	// deviceid
	// format
	// add (slice)
	// remove (slice)
	username := chi.URLParam(r, "username")
	deviceIdStr := chi.URLParam(r, "deviceid")
	format := chi.URLParam(r, "format")

	if format != "json" {
		log.Printf("error uploading device subscription changes as format is expecting JSON but got %#v", format)
		w.WriteHeader(400)
		return
	}

	deviceId, err := s.Data.GetDeviceIdFromName(deviceIdStr, username)
	if err != nil {
		log.Printf("error parsing device id: %s", err)
		w.WriteHeader(500)
		return
	}
	subscriptionChanges := &SubscriptionChanges{}
	err = json.NewDecoder(r.Body).Decode(&subscriptionChanges)
	if err != nil {

		log.Printf("error decoding json payload: %#v", err)
		w.WriteHeader(400)
		return
	}

	addSlice := subscriptionChanges.Add
	removeSlice := subscriptionChanges.Remove

	ts := data.CustomTimestamp{}
	ts.Time = time.Now()

	db := s.Data

	syncDevices, err := db.GetDevicesInSyncGroupFromDeviceId(deviceId)
	if err != nil {
		log.Printf("error trying to retrieve devices in sync_group: %s", err)
		w.WriteHeader(500)
		return
	}

	if syncDevices == nil {
		syncDevices = []int{deviceId}
	}

	pairz := []Pair{}
	for _, v := range addSlice {
		sub := data.Subscription{
			User:      username,
			Devices:   syncDevices,
			Podcast:   v,
			Timestamp: ts,
			Action:    "SUBSCRIBE",
		}
		pair := Pair{v, v}
		pairz = append(pairz, pair)
		err := db.AddSubscriptionHistory(sub)
		if err != nil {
			log.Printf("error adding subscription: %#v", err)
		}
	}

	for _, v := range removeSlice {
		sub := data.Subscription{
			User:      username,
			Devices:   syncDevices,
			Podcast:   v,
			Timestamp: ts,
			Action:    "UNSUBSCRIBE",
		}
		pair := Pair{v, v}
		pairz = append(pairz, pair)
		db.AddSubscriptionHistory(sub)
	}

	pp := PairArray{pairz}

	subscriptionChangeOutput := &SubscriptionChangeOutput{
		Timestamp:  timestamp.Time(ts.Time),
		UpdateUrls: json.RawMessage(pp.String()),
	}

	outputBytes, err := json.Marshal(subscriptionChangeOutput)
	if err != nil {
		log.Printf("error marshalling output: %#v", err)
		w.WriteHeader(400)
		return
	}
	w.WriteHeader(200)
	w.Write(outputBytes)
}

func (s *SubscriptionAPI) HandleGetSubscription(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	xml, err := s.Data.RetrieveAllDeviceSubscriptions(username)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	w.Write([]byte(xml))
}

func (s *SubscriptionAPI) HandleGetDeviceSubscription(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	deviceId := chi.URLParam(r, "deviceid")

	xml, err := s.Data.RetrieveDeviceSubscriptions(username, deviceId)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	w.Write([]byte(xml))
	w.WriteHeader(200)
}

// API Endpoint: POST and PUT /subscriptions/{username}/{deviceid}.{format}
func (s *SubscriptionAPI) HandleUploadDeviceSubscription(w http.ResponseWriter, r *http.Request) {

	var (
		toBeAdded   []string
		toBeRemoved []string
	)

	username := chi.URLParam(r, "username")
	deviceIdStr := chi.URLParam(r, "deviceid")
	deviceId, err := s.Data.GetDeviceIdFromName(deviceIdStr, username)
	if err != nil {
		log.Printf("error parsing device id: %s", err)
		w.WriteHeader(500)
		return
	}

	format := chi.URLParam(r, "format")
	ts := data.CustomTimestamp{}
	ts.Time = time.Now()

	switch r.Method {
	case "POST":
		log.Println("Receive a POST")
	case "PUT":
		// Upload entire subscriptions

		// TODO: need to handle all the different formats, json, xml, text etc

		log.Println("Receive a PUT")
		log.Printf("Saving subscription...")

		b, _ := io.ReadAll(r.Body)

		var arr []string
		err := json.Unmarshal(b, &arr)
		if err != nil {
			log.Printf("error unmarshalling payload to json: %s", err)
			w.WriteHeader(400)
			return
		}

		f, err := os.Create(fmt.Sprintf("%s-%d.%s", username, deviceId, format))
		if err != nil {
			log.Printf("error saving file: %#v", err)
			w.WriteHeader(400)
			return
		}
		defer f.Close()

		// get a list of currently subscribed podcasts
		// iterate through it and then check if it differs with the uploaded
		// content.

		subscribedPodcasts, err := s.Data.RetrieveDeviceSubscriptionsSlice(username, deviceIdStr)
		if err != nil {
			log.Printf("error getting subscriptions: %#v", err)
			w.WriteHeader(500)
			return
		}

		log.Printf("Current subscription on device on server count: %d", len(subscribedPodcasts))
		log.Printf("Subscription on local machine count: %d", len(arr))

		// there should be some room to optimize these 2 loops, will need to find a
		// better way
		for _, v := range arr {
			// if local subscription is not in subscribed podcast, add it in.
			if !slices.Contains(subscribedPodcasts, v) {
				log.Printf("To Be Added: %s\n", v)
				toBeAdded = append(toBeAdded, v)
			}
		}

		for _, v := range subscribedPodcasts {

			// if subscribed podcasts is not in the local subscriptions, remove it
			if !slices.Contains(arr, v) {
				log.Printf("To Be Removed: %s\n", v)
				toBeRemoved = append(toBeRemoved, v)

			}

		}

		// https://github.com/gpodder/mygpo/blob/80c41dc0c9a58dc0e85f6ef56662cdfd0d6e3b16/mygpo/api/simple.py#L213
		for _, v := range toBeAdded {
			sub := data.Subscription{
				User:      username,
				Devices:   []int{deviceId},
				Podcast:   v,
				Action:    "SUBSCRIBE",
				Timestamp: ts,
			}
			s.Data.AddSubscriptionHistory(sub)
		}
		for _, v := range toBeRemoved {
			sub := data.Subscription{
				User:      username,
				Devices:   []int{deviceId},
				Podcast:   v,
				Action:    "UNSUBSCRIBE",
				Timestamp: ts,
			}
			s.Data.AddSubscriptionHistory(sub)
		}

		w.WriteHeader(200)
		return
	default:
		w.WriteHeader(400)
		return

	}
}

// EpisodeAPI

func (e *EpisodeAPI) HandleEpisodeAction(w http.ResponseWriter, r *http.Request) {
	// username
	// format - defaulting to "json" as per spec
	// username := chi.URLParam(r, "username")

	// body:
	// podcast (string) optional
	// device (string) optional
	// since (int) optional also, if no actions, then release all
	// aggregated (bool)

	username := chi.URLParam(r, "username")
	device := r.URL.Query().Get("device")
	since := r.URL.Query().Get("since")

	var parsedTs time.Time
	if since != "" {
		sinceInt, err := strconv.Atoi(since)
		if err != nil {
			log.Printf("error parsing since: %#v", err)
			w.WriteHeader(400)
			return
		}
		parsedTs = time.Unix(int64(sinceInt), 0).UTC()
	}

	actions, err := e.Data.RetrieveEpisodeActionHistory(username, device, parsedTs)

	if err != nil {
		log.Printf("error retrieving episodes actions output: %#v", err)
		w.WriteHeader(400)
		return
	}

	episodeActionOutput := &EpisodeActionOutput{
		Actions:   actions,
		Timestamp: timestamp.Now(),
	}

	episodeActionOutputBytes, err := json.Marshal(episodeActionOutput)
	if err != nil {
		log.Printf("error marshalling episodes actions output: %#v", err)
		w.WriteHeader(400)
		return
	}
	w.WriteHeader(200)
	w.Write(episodeActionOutputBytes)
}

// POST /api/2/episodes/{username}.json
func (e *EpisodeAPI) HandleUploadEpisodeAction(w http.ResponseWriter, r *http.Request) {
	// username

	username := chi.URLParam(r, "username")
	ts := time.Now()

	b, _ := io.ReadAll(r.Body)

	var arr []data.EpisodeAction

	pairz := []Pair{}

	err := json.Unmarshal(b, &arr)
	if err != nil {
		log.Printf("error unmarshalling: %#v", err)
		w.WriteHeader(400)
		return
	}

	for _, data := range arr {
		err := e.Data.AddEpisodeActionHistory(username, data)
		if err != nil {
			log.Printf("error adding episode action into history: %#v", err)
			w.WriteHeader(400)
			return
		}
		pair := Pair{
			data.Episode, data.Episode,
		}
		pairz = append(pairz, pair)
	}

	// format

	pp := PairArray{pairz}
	subscriptionChangeOutput := &SubscriptionChangeOutput{
		Timestamp:  timestamp.Time(ts),
		UpdateUrls: json.RawMessage(pp.String()),
	}

	outputBytes, err := json.Marshal(subscriptionChangeOutput)
	if err != nil {
		log.Printf("error marshalling output: %#v", err)
		w.WriteHeader(400)
		return
	}
	w.WriteHeader(200)
	log.Printf("outputbytes is %#v", string(outputBytes))
	w.Write(outputBytes)
}

// GET /api/2/sync-devices/{username}.json
func (s *SyncAPI) HandleGetSync(w http.ResponseWriter, r *http.Request) {

	username := chi.URLParam(r, "username")

	syncStatus := &SyncDeviceStatus{}

	db := s.Data

	syncIds, err := db.GetDeviceSyncGroupIds(username)
	if err != nil {
		log.Printf("error getting sync devices for username (%s): %s", username, err)
		w.WriteHeader(500)
		return
	}
	notsyncDevices, err := db.GetNotSyncedDevices(username)
	if err != nil {
		log.Printf("error getting not_synced devices for username (%s): %s", username, err)
		w.WriteHeader(500)
		return
	}

	syncStatus.NotSynchronize = notsyncDevices

	for _, id := range syncIds {
		sync, err := db.GetDeviceNameFromDeviceSyncGroupId(id)
		if err != nil {
			log.Printf("error retrieving devices from sync group_id: (%d): %s", id, err)
		}

		syncStatus.Synchronized = append(syncStatus.Synchronized, sync)
	}

	jsonBytes, err := json.Marshal(syncStatus)
	if err != nil {
		log.Printf("error marshalling sync status: %#v", err)
		w.WriteHeader(500)
	}

	w.Write(jsonBytes)
}

// POST /api/2/sync-devices/{username}.json
// This endpoints takes in a SyncDeviceRequest to link up devices together
func (s *SyncAPI) HandlePostSync(w http.ResponseWriter, r *http.Request) {

	username := chi.URLParam(r, "username")

	syncReq := &SyncDeviceRequest{}
	syncResp := &SyncDeviceStatus{}

	respBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading request body: %#v", err)
		w.WriteHeader(400)
		return
	}

	err = json.Unmarshal(respBody, &syncReq)
	if err != nil {
		log.Printf("error unmarshalling sync device request: %#v", err)
		w.WriteHeader(400)
		return
	}

	for _, syncgroups := range syncReq.Synchronize {
		err := s.Data.AddSyncGroup(syncgroups, username)
		if err != nil {
			log.Printf("errors adding sync group: %#v", err)
			w.WriteHeader(500)
			return
		}

	}

	for _, device := range syncReq.StopSynchronize {
		s.Data.StopDeviceSync(device, username)
	}

	// start preparing the response back to the user

	// get all the device_sync group_id belonging to user
	ids, err := s.Data.GetDeviceSyncGroupIds(username)
	if err != nil {
		log.Printf("error getting devices sync groups id from username: %#v", err)
		w.WriteHeader(500)
		return
	}

	for _, deviceSyncGroupId := range ids {
		devices, err := s.Data.GetDeviceNameFromDeviceSyncGroupId(deviceSyncGroupId)
		if err != nil {
			log.Printf("error getting device names from device sync id: %#v", err)
			w.WriteHeader(500)
			return
		}

		syncResp.Synchronized = append(syncResp.Synchronized, devices)

	}

	notSyncedDevices, err := s.Data.GetNotSyncedDevices(username)
	if err != nil {
		log.Printf("error getting devices that are not synced: %#v", err)
		w.WriteHeader(500)
		return
	}

	syncResp.NotSynchronize = notSyncedDevices

	respBytes, err := json.Marshal(syncResp)
	if err != nil {
		log.Printf("error marshalling json for sync response: %#v", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
	w.Write(respBytes)
	return
}
