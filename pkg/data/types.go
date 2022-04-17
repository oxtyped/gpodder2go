package data

import (
	"strings"
	"time"
)

type DataInterface interface {
	AddUser(string, string, string, string) error
	AddSubscriptionHistory(Subscription) error
	RetrieveSubscriptionHistory(string, string, time.Time) ([]Subscription, error)
	AddEpisodeActionHistory(username string, e EpisodeAction) error
	RetrieveEpisodeActionHistory(username string, deviceId string, since time.Time) ([]EpisodeAction, error)
	//	RetrieveLoginToken(username string, password string) (string, error)
	RetrieveDevices(username string) ([]Device, error)
	AddDevice(username string, deviceName string, caption string, deviceType string) error

	RetrieveAllDeviceSubscriptions(username string) (string, error)
	RetrieveDeviceSubscriptions(username string, deviceNme string) (string, error)
}

type Subscription struct {
	User      string          `json:"user"`
	Device    string          `json:"device"`
	Podcast   string          `json:"podcast"`
	Action    string          `json:"action"`
	Timestamp CustomTimestamp `json:"timestamp"`
}

type Device struct {
	User    *User  `json:"user"`
	Id      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Caption string `json:"caption"`
}

type User struct {
	Name string `json:"name"`
}

type EpisodeAction struct {
	Podcast   string          `json:"podcast"`
	Episode   string          `json:"episode"`
	Device    string          `json:"device"`
	Action    string          `json:"action"`
	Position  int             `json:"position"`
	Started   int             `json:"started"`
	Total     int             `json:"total"`
	Timestamp CustomTimestamp `json:"timestamp"`
}

// CustomTimestamp is to handle ISO 8601 timestamp for unmarshalling
type CustomTimestamp struct {
	time.Time
}

func (c *CustomTimestamp) UnmarshalJSON(b []byte) error {
	value := strings.Trim(string(b), `"`) //get rid of "
	if value == "" || value == "null" {
		return nil
	}

	t, err := time.Parse("2006-01-02T15:04:05", value) //parse time
	if err != nil {
		return err
	}
	c.Time = t
	return nil
}

func (c CustomTimestamp) MarshalJSON() ([]byte, error) {
	return []byte(`"` + c.Time.Format("2006-01-02T15:04:05.999999999Z0700") + `"`), nil
}
