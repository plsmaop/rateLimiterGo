package memstore

import (
	"sync"
	"time"
)

type entry struct {
	counter int64
	// UTC in sec
	resetTime int64
}

type Store struct {
	// lock for table and ttl
	lock  sync.Mutex
	table map[string]entry
}

// NewMemStore for creating memory store instance
func NewMemStore() *Store {
	return &Store{
		lock:  sync.Mutex{},
		table: map[string]entry{},
	}
}

// for automatically deleting useless key
func (s *Store) setExpiration(key string, expiration int64) {
	time.AfterFunc(time.Duration(expiration)*time.Second, func() {
		s.lock.Lock()
		s.lock.Unlock()
		delete(s.table, key)
	})
}

// INCR increments counter of the key
// and return current counter and its TTL
func (s *Store) INCR(key string, limit int64, expiration int64) (int64, int64, error) {

	s.lock.Lock()
	defer s.lock.Unlock()
	e, ok := s.table[key]

	if !ok {
		e = entry{
			counter:   1,
			resetTime: time.Now().Unix() + expiration,
		}
		s.table[key] = e
		s.setExpiration(key, expiration)
	} else {
		e.counter++
		s.table[key] = e
	}

	ttl := e.resetTime - time.Now().Unix()

	return e.counter, ttl, nil
}
