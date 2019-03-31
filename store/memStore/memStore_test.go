package memstore

import (
	"testing"

	ratelimiter "github.com/plsmaop/rateLimiterGo"
	"github.com/plsmaop/rateLimiterGo/tester"
)

var newStore = func(_ *testing.T) ratelimiter.Store {
	return NewMemStore()
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
