package store

import (
	"fmt"
	"time"

	memcached "github.com/bradfitz/gomemcache/memcache"
)

type MemcachedStore struct {
	Cache  *memcached.Client
	Prefix string
}

func NewMemcachedStore(hosts string, prefix string) *MemcachedStore {

	cache := memcached.New(hosts)
	cache.Timeout = 3 * time.Second
	return &MemcachedStore{
		Cache:  cache,
		Prefix: prefix,
	}

}

func (m *MemcachedStore) Ping() error {
	return m.Cache.Ping()
}

func (m *MemcachedStore) Get(key string) (string, error) {
	if err := m.Ping(); err != nil {
		return "", err
	}

	item, err := m.Cache.Get(fmt.Sprintf("%s_%s", m.Prefix, key))
	if err != nil {
		return "", err
	}

	return string(item.Value), nil

}

func (m *MemcachedStore) Set(key string, value string) error {
	if err := m.Ping(); err != nil {
		return err
	}

	item := &memcached.Item{
		Key:   fmt.Sprintf("%s_%s", m.Prefix, key),
		Value: []byte(value),
	}

	return m.Cache.Set(item)
}
