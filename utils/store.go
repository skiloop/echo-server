package utils

import (
	"context"
	"sync"
	"time"
)

type KVStore struct {
	data sync.Map
	TTL  int64
}

// valueWithExpiry holds the value and its expiration time.
type valueWithExpiry struct {
	value      interface{}
	expiration int64 // Unix timestamp in nanoseconds
}

// Set adds a key-value pair with an optional expiration time (in seconds).
func (s *KVStore) Set(key string, value interface{}) {
	var expiration int64
	if s.TTL > 0 {
		expiration = time.Now().Add(time.Duration(s.TTL) * time.Second).UnixNano()
	}
	s.data.Store(key, valueWithExpiry{value: value, expiration: expiration})
}

func NewKVStore(ttl int64) *KVStore {
	return &KVStore{data: sync.Map{}, TTL: ttl}
}

// Get retrieves the value for a given key. If the key is expired or does not exist, it returns nil and false.
func (s *KVStore) Get(key string) (interface{}, bool) {
	item, ok := s.data.Load(key)
	if !ok {
		return nil, false
	}

	v := item.(valueWithExpiry)
	if v.expiration > 0 && time.Now().UnixNano() > v.expiration {
		s.data.Delete(key) // Remove expired key
		return nil, false
	}

	return v.value, true
}

// Delete removes a key-value pair from the store.
func (s *KVStore) Delete(key string) {
	s.data.Delete(key)
}

// Cleanup removes all expired keys from the store.
func (s *KVStore) Cleanup() {
	s.data.Range(func(key, value interface{}) bool {
		v := value.(valueWithExpiry)
		if v.expiration > 0 && time.Now().UnixNano() > v.expiration {
			s.data.Delete(key)
		}
		return true
	})
}

// StartCleanupRoutine starts a background Goroutine to periodically clean up expired keys.
// It accepts a context to allow graceful shutdown.
func (s *KVStore) StartCleanupRoutine(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// Exit the Goroutine when the context is canceled
				return
			case <-ticker.C:
				s.Cleanup()
			}
		}
	}()
}
