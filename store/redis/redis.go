package redis

import (
	"errors"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

// Useful discussion
// https://stackoverflow.com/questions/37178897/some-questions-about-redigo-and-concurrency

// Config for connecting to redis server
type Config struct {
	Host        string
	Port        string
	MaxIdle     int
	MaxActive   int
	IdleTimeout time.Duration
}

func (c *Config) normalize() {
	if len(c.Host) == 0 {
		c.Host = "127.0.0.1"
	}
	if len(c.Port) == 0 {
		c.Port = "6379"
	}
	if c.MaxIdle == 0 {
		c.MaxIdle = 100
	}
	if c.MaxActive == 0 {
		c.MaxActive = 500
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = 240 * time.Second
	}
}

// Store for storing number of counts
type Store struct {
	pool *redis.Pool
}

// NewRedisStore creates a redis client
func NewRedisStore(c *Config) (*Store, error) {
	c.normalize()
	s := &Store{
		pool: &redis.Pool{
			MaxIdle:     c.MaxIdle,
			IdleTimeout: c.IdleTimeout,
			MaxActive:   c.MaxActive,
			Dial: func() (redis.Conn, error) {
				return redis.Dial(
					"tcp",
					fmt.Sprintf("%s:%s", c.Host, c.Port),
				)
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
			Wait: true,
		},
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

	return result[0], result[1], nil
}

func (s *Store) incr(key string) (int64, int64, error) {
	conn := s.pool.Get()
	defer conn.Close()

	// MULTI
	conn.Send("MULTI")
	conn.Send("INCR", key)
	conn.Send("TTL", key)
	replies, err := conn.Do("EXEC")

	return extractCountAndTTL(replies, err)
}

// for automatically deleting useless key
func (s *Store) setExpiration(key string, expiration int64) error {
	conn := s.pool.Get()
	defer conn.Close()

	_, err := conn.Do("EXPIRE", key, expiration)
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
		return limit, 0, err
	}

	// -1 means that the counter has no expiration
	if TTL == -1 {
		s.setExpiration(key, expiration)
		return currentCounter, expiration, nil
	}

	return currentCounter, TTL, nil
}
