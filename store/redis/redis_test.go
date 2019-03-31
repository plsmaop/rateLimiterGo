package redis

import (
	"testing"

	rateLimiter "github.com/plsmaop/rateLimiterGo"
	"github.com/plsmaop/rateLimiterGo/tester"
)

var newStore = func(_ *testing.T) rateLimiter.Store {
	r, err := NewRedisStore(&Config{})
	if err != nil {
		panic(err)
	}
	return r
}

func Test_ResponseHeader(t *testing.T) {
	tester.ResponseHeader(t, newStore)
}

func Test_ResponseStatus(t *testing.T) {
	tester.ResponseStatus(t, newStore)
}

func Test_TooManyRequests(t *testing.T) {
	tester.TooManyRequests(t, newStore)
}

func Test_TimeReset(t *testing.T) {
	tester.TimeReset(t, newStore)
}