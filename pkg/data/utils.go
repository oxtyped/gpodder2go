package data

import "log"

// SubscriptionDiff takes in a slice of Subscription and returns a seperated
// string slice of podcasts that are to be added or removed.
// https://github.com/gpodder/mygpo/blob/e20f107009bd07e8baf226a48131fc1b1e0383ff/mygpo/subscriptions/__init__.py#L149-L167
func SubscriptionDiff(subs []Subscription) (Add []string, Remove []string) {

	subscriptions := make(map[string]int)

	for _, v := range subs {
		log.Printf("%#v", v.Podcast)
		switch v.Action {
		case "SUBSCRIBE":
			subscriptions[v.Podcast] += 1
		case "UNSUBSCRIBE":
			subscriptions[v.Podcast] -= 1
		default:
			break
		}
	}

	add := []string{}
	remove := []string{}
	for k, v := range subscriptions {
		if v > 0 {
			add = append(add, k)
		} else {
			remove = append(remove, k)
		}
	}

	return add, remove

}
