package ratelimiter

import (
	"errors"
)

// Config for rateLimiter
type Config struct {
	Limit int64
	// in sec
	Expiration int64

	Store Store
}

// ValidateAndNormalize validates and sets default value for the Config
func (c *Config) ValidateAndNormalize() error {
	if c.Limit < 0 {
		return errors.New("negative Limit is not allowed")
	}
	if c.Expiration < 0 {
		return errors.New("negative Expiration is not allowed")
	}

	if c.Store == nil {
		// default store is memStore
		// c.Store = memstore.NewMemStore()
		return errors.New("empty store is not allowed")
	}
	if c.Limit == 0 {
		// default threshold of requests is 1000
		c.Limit = 1000
	}
	if c.Expiration == 0 {
		// default windowSize is an hour
		c.Expiration = 60 * 60
	}

	return nil
}
