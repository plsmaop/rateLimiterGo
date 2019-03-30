package redis

import (
	"errors"

	"github.com/gomodule/redigo/redis"
)

// Config for connecting to redis server
type Config struct {
	host string
	port string
}

func (c *Config) normalize() {
	if len(c.host) == 0 {
		c.host = "127.0.0.1"
	}
	if len(c.port) == 0 {
		c.port = "6379"
	}
}

// Store for storing number of counts
type Store struct {
	c redis.Conn
}

// NewRedisStore creates a redis client
func NewRedisStore(c Config) (*Store, error) {
	c.normalize()
	conn, err := redis.Dial(c.host, c.port)
	if err != nil {
		return nil, err
	}

	defer conn.Close()
	s := &Store{
		c: conn,
	}
	return s, nil
}

func extractCountAndTTL(replies interface{}, e error) (int64, int64, error) {
	result, err := redis.Int64s(replies, e)
	if err != nil {
		return 0, 0, err
	}

	if len(result) != 2 {
		return 0, 0, errors.New("INCRCount return number Error")
	}

	return result[0], result[1], err
}

func (s *Store) incr(key string) (int64, int64, error) {
	// MULTI
	s.c.Send("MULTI")
	s.c.Send("INCR", key)
	s.c.Send("TTL", key)
	replies, err := s.c.Do("EXEC")

	return extractCountAndTTL(replies, err)
}

// for automatically deleting useless key
func (s *Store) setExpiration(key string, expiration int64) error {
	_, err := s.c.Do("EXPIRE", key, expiration)
	if err != nil {
		return err
	}

	return nil
}

// INCR increments counter of the key
// and return current counter and its TTL
func (s *Store) INCR(key string, limit int64, expiration int64) (int64, int64, error) {
	currentCounter, TTL, err := s.incr(key)
	if err != nil {
		return 0, 0, err
	}

	if currentCounter > limit {
		return 0, limit, nil
	}

	if TTL == -1 {
		s.setExpiration(key, expiration)
		return currentCounter, expiration, nil
	}
	return currentCounter, TTL, nil
}
