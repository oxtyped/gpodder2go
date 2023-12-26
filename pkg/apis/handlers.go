package apis

import (
	"crypto/hmac"
	"crypto/sha256"
	"os"
	"strconv"
	"time"

	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/oxtyped/gpodder2go/pkg/data"

	"github.com/augurysys/timestamp"
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
	return

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
	return

}

// DeviceAPI
func (d *DeviceAPI) HandleUpdateDevice(w http.ResponseWriter, r *http.Request) {

	// username
	// deviceid

	username := chi.URLParam(r, "username")
	deviceName := chi.URLParam(r, "deviceid")

	log.Printf("username is %s, deviceName is %s", username, deviceName)

	ddr := &DeviceDataRequest{}

	payload, err := ioutil.ReadAll(r.Body)
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
	err = d.Data.AddDevice(username, deviceName, ddr.Caption, ddr.Type)
	if err != nil {
		log.Printf("error adding device: %#v", err)
		w.WriteHeader(400)
		return
	}

	// 200
	// 401
	// 404
	// 400
	w.WriteHeader(200)
	return

}

func (d *DeviceAPI) HandleGetDevices(w http.ResponseWriter, r *http.Request) {

	type GetDevicesOutput struct {
		Name          string `json:"id"`
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
			Name:          v.Name,
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
	return

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
	tm := time.Time{}

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
	return

}

// API Endpoint: POST /api/2/subscriptions/{username}/{deviceid}.{format}
func (s *SubscriptionAPI) HandleUploadDeviceSubscriptionChange(w http.ResponseWriter, r *http.Request) {

	// username
	// deviceid
	// format
	// add (slice)
	// remove (slice)
	username := chi.URLParam(r, "username")
	deviceId := chi.URLParam(r, "deviceid")
	format := chi.URLParam(r, "format")

	if format != "json" {
		log.Printf("error uploading device subscription changes as format is expecting JSON but got %#v", format)
		w.WriteHeader(400)
		return
	}
	subscriptionChanges := &SubscriptionChanges{}
	err := json.NewDecoder(r.Body).Decode(&subscriptionChanges)
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

	pairz := []Pair{}
	for _, v := range addSlice {
		sub := data.Subscription{
			User:      username,
			Device:    deviceId,
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
			Device:    deviceId,
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
	return

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

	username := chi.URLParam(r, "username")
	deviceId := chi.URLParam(r, "deviceid")
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

		b, _ := ioutil.ReadAll(r.Body)

		var arr []string
		err := json.Unmarshal(b, &arr)
		if err != nil {
			log.Fatal(err)
		}

		f, err := os.Create(fmt.Sprintf("%s-%s.%s", username, deviceId, format))
		if err != nil {
			log.Printf("error saving file: %#v", err)
			w.WriteHeader(400)
			return
		}
		defer f.Close()

		// start to write each line by line
		for _, v := range arr {

			sub := data.Subscription{
				User:      username,
				Device:    deviceId,
				Podcast:   v,
				Action:    "SUBSCRIBE",
				Timestamp: ts,
			}
			s.Data.AddSubscriptionHistory(sub)
			f.WriteString(v + "\n")
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

	//username
	//format - defaulting to "json" as per spec
	//username := chi.URLParam(r, "username")

	// body:
	// podcast (string) optional
	// device (string) optional
	// since (int) optional also, if no actions, then release all
	// aggregated (bool)

	username := chi.URLParam(r, "username")
	device := r.URL.Query().Get("device")
	since := r.URL.Query().Get("since")

	actions, err := e.Data.RetrieveEpisodeActionHistory(username, device, since)

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
	return

}

// POST /api/2/episodes/{username}.json
func (e *EpisodeAPI) HandleUploadEpisodeAction(w http.ResponseWriter, r *http.Request) {
	//username

	username := chi.URLParam(r, "username")
	ts := time.Now()

	b, _ := ioutil.ReadAll(r.Body)

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
