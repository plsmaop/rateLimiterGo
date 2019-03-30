package ratelimiter

import (
	"fmt"
	"time"
)

// Context for a given key
type Context struct {
	CurrentCounter int64
	// UTC in sec
	ResetTime      int64
	IsReachedLimit bool
}

// Store for storing number of counts of a key
type Store interface {
	INCR(string, int64, int64) (int64, int64, error)
}

// RateLimiter limit the requests an key can make in a period of time
type RateLimiter interface {
}

type rateLimiter struct {
	config Config
	store  Store
	// req    *http.Request
}

func (r *rateLimiter) generateKey(key string) string {
	now := time.Now().Unix() / r.config.expiration
	return fmt.Sprintf("%s:%d", key, now)
}

func (r *rateLimiter) calculateResetTime(TTL int64) int64 {
	nextInterval := time.Now().Unix()/r.config.expiration + 1
	now := time.Now().Unix()
	resetTime := nextInterval*r.config.expiration - now
	/* if resetTime < 60 {
		unit = "second(s)"
	} else if resetTime < 60*60 {
		unit = "minute(s)"
	} else {
		unit = "hour(s)"
	} */
	return resetTime
}

func newRateLimiter(c Config) *rateLimiter {
	if err := c.ValidateAndNormalize(); err != nil {
		panic(err.Error())
	}

	return &rateLimiter{
		store:  c.Store,
		config: c,
	}
}

// NewRateLimiter for creating new rateLimiter instance
func NewRateLimiter(c Config) RateLimiter {
	return newRateLimiter(c)
}

// Get the Context of the given key
func (r *rateLimiter) Get(key string) (*Context, error) {
	keyWithTimestamp := r.generateKey(key)

	currentCounter, TTL, err := r.store.INCR(keyWithTimestamp, r.config.limit, r.config.expiration)

	if err != nil {
		return nil, err
	}

	c := &Context{
		IsReachedLimit: r.config.limit-currentCounter > 0,
		CurrentCounter: currentCounter,
		ResetTime:      TTL,
	}

	return c, nil
}
