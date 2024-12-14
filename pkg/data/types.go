package data

import (
	"strings"
	"time"
)

type DataInterface interface {
	AddUser(string, string, string, string) error
	CheckUserPassword(string, string) bool
	AddSubscriptionHistory(Subscription) error
	RetrieveSubscriptionHistory(string, string, time.Time) ([]Subscription, error)
	AddEpisodeActionHistory(username string, e EpisodeAction) error
	RetrieveEpisodeActionHistory(username string, deviceId string, since time.Time) ([]EpisodeAction, error)

	// Devices
	RetrieveDevices(username string) ([]Device, error)
	AddDevice(username string, deviceName string, caption string, deviceType string) (int, error)
	UpdateOrCreateDevice(username string, deviceName string, caption string, deviceType string) (int, error)
	RetrieveAllDeviceSubscriptions(username string) (string, error)
	RetrieveAllDeviceSubscriptionsSlice(username string) ([]string, error)
	RetrieveDeviceSubscriptions(username string, deviceNme string) (string, error)
	RetrieveDeviceSubscriptionsSlice(username string, deviceNme string) ([]string, error)
	GetDeviceIdFromName(deviceName string, username string) (int, error)

	// sync
	AddSyncGroup(deviceIds []string, username string) error
	StopDeviceSync(deviceName string, username string) error
	GetDeviceSyncGroupIds(username string) ([]int, error)
	GetDevicesInSyncGroupFromDeviceId(deviceId int) ([]int, error)
	GetDeviceNameFromDeviceSyncGroupId(deviceId int) ([]string, error)
	GetNotSyncedDevices(username string) ([]string, error)
}

type Subscription struct {
	User      string          `json:"user"`
	Device    string          `json:"device"`
	Devices   []int           `json:"devices"`
	Podcast   string          `json:"podcast"`
	Action    string          `json:"action"`
	Timestamp CustomTimestamp `json:"timestamp"` // sqlite stores this as a varchar(255)
}

type Device struct {
	User    *User  `json:"user"`
	Id      int    `json:"id"`   // This is just a database ID and not to be confused with DeviceId
	Name    string `json:"name"` // Name is represents the actual DeviceId that is referenced in handlers
	Type    string `json:"type"`
	Caption string `json:"caption"` // To be deprecated
}

type User struct {
	Name string `json:"name"`
}

type EpisodeAction struct {
	Podcast   string          `json:"podcast"`
	Episode   string          `json:"episode"`
	Device    string          `json:"device"`
	Devices   []int           `json:"devices"`
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
	value := strings.Trim(string(b), `"`) // get rid of "
	if value == "" || value == "null" {
		return nil
	}

	t, err := time.Parse("2006-01-02T15:04:05", value) // parse time
	if err != nil {
		return err
	}
	c.Time = t
	return nil
}

func (c CustomTimestamp) MarshalJSON() ([]byte, error) {
	return []byte(`"` + c.Time.Format("2006-01-02T15:04:05.999999999Z0700") + `"`), nil
}
