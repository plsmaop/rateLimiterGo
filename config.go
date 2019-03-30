package ratelimiter

import (
	"errors"
	"net"
	"net/http"
)

// Config for rateLimiter
type Config struct {
	limit int64
	// UTC in sec
	expiration int64

	Store  Store
	getKey func(req *http.Request) string
}

// ValidateAndNormalize validates and sets default value for the Config
func (c *Config) ValidateAndNormalize() error {
	if c.limit < 0 {
		return errors.New("negative maxRequests is not allowed")
	}
	if c.expiration < 0 {
		return errors.New("negative windowSize is not allowed")
	}
	if c.Store == nil {
		return errors.New("empty Store is not allowed")
	}

	if c.limit == 0 {
		// default threshold of requests is 1000
		c.limit = 1000
	}
	if c.expiration == 0 {
		// default windowSize is an hour
		c.expiration = 60 * 60
	}
	if c.getKey == nil {
		c.getKey = func(req *http.Request) string {
			ip, _, _ := net.SplitHostPort(req.RemoteAddr)
			return ip
		}
	}
	return nil
}
