package apis

import (
	"encoding/json"

	"github.com/oxtyped/gpodder2go/pkg/data"
	"github.com/oxtyped/gpodder2go/pkg/store"

	"github.com/augurysys/timestamp"
)

type DeviceAPI struct {
	Store store.Store
	Data  data.DataInterface
}

type SubscriptionAPI struct {
	Store store.Store
	Data  data.DataInterface
}

type EpisodeAPI struct {
	Store store.Store
	Data  data.DataInterface
}

type UserAPI struct {
	Data data.DataInterface
}

// Incoming Payload

type DeviceDataRequest struct {
	Caption string `json:"caption"`
	Type    string `json:"type"`
}

type SubscriptionChanges struct {
	Add       []string             `json:"add"`
	Remove    []string             `json:"remove"`
	Timestamp *timestamp.Timestamp `json:"timestamp"`
}

// Outgoing Payload

type SubscriptionChangeOutput struct {
	Timestamp  *timestamp.Timestamp `json:"timestamp"`
	UpdateUrls json.RawMessage      `json:"update_urls"`
}

type EpisodeActionOutput struct {
	Actions   []data.EpisodeAction `json:"actions"`
	Timestamp *timestamp.Timestamp `json:"timestamp"`
}
