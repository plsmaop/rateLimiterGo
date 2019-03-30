package ratelimiter

import (
	"fmt"
	"time"
)

// Context for a given key
type Context struct {
	CurrentCounter   int64
	RemainingCounter int64
	// in sec
	TTL int64
	// UTC in sec
	ResetTime      int64
	IsReachedLimit bool
}

// Store for storing number of counts of a key
type Store interface {
	// currentCounter, TTL, err
	INCR(string, int64, int64) (int64, int64, error)
}

// RateLimiter limit the requests an key can make in a period of time
type RateLimiter interface {
	Get(key string) (*Context, error)
}

type rateLimiter struct {
	config *Config
	store  Store
}

func (r *rateLimiter) generateKey(key string) string {
	now := time.Now().Unix() / r.config.Expiration
	return fmt.Sprintf("%s:%d", key, now)
}

func newRateLimiter(c *Config) *rateLimiter {
	if err := c.ValidateAndNormalize(); err != nil {
		panic(err.Error())
	}

	return &rateLimiter{
		store:  c.Store,
		config: c,
	}
}

// NewRateLimiter for creating new rateLimiter instance
func NewRateLimiter(c *Config) RateLimiter {
	return newRateLimiter(c)
}

// Get the Context of the given key
func (r *rateLimiter) Get(key string) (*Context, error) {
	keyWithTimestamp := r.generateKey(key)

	currentCounter, TTL, err := r.store.INCR(keyWithTimestamp, r.config.Limit, r.config.Expiration)

	if err != nil {
		return nil, err
	}

	c := &Context{
		IsReachedLimit:   currentCounter-r.config.Limit > 0,
		CurrentCounter:   currentCounter,
		RemainingCounter: 0,
		TTL:              TTL,
		ResetTime:        time.Now().Unix() + TTL,
	}

	if !c.IsReachedLimit {
		c.RemainingCounter = r.config.Limit - currentCounter
	}

	return c, nil
}
