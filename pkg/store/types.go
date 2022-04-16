package store

import (
	"errors"
	"time"

	"github.com/patrickmn/go-cache"
)

type Store interface {
	Get(key string) (string, error)
	Set(key string, value string) error
}

type LocalCacheStore struct {
	Cache *cache.Cache
}

func NewCacheStore() *LocalCacheStore {
	localCacheStore := &LocalCacheStore{
		Cache: cache.New(cache.NoExpiration, 10*time.Minute),
	}

	return localCacheStore

}

func (s *LocalCacheStore) Get(key string) (string, error) {

	c := s.Cache

	k, ok := c.Get(key)
	if ok {
		return k.(string), nil
	}

	return "", errors.New("error retrieving value")
}

func (s *LocalCacheStore) Set(key string, value string) error {
	c := s.Cache
	c.Set(key, value, cache.NoExpiration)
	return nil
}
