// Package tester is a package to test gin rate limiter middleware
// with different packages of stores, including redisSotre and memStore.
// TODO: unit test, pure limiter test
package tester

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	rateLimiter "github.com/plsmaop/rateLimiterGo"
	ginMiddleware "github.com/plsmaop/rateLimiterGo/middleware/gin"
)

type storeFactory func(*testing.T) rateLimiter.Store

const (
	MEM_STORE = iota
	REDIS_STORE
)

func init() {
	gin.SetMode(gin.TestMode)
}

var (
	reqNum     = int64(150)
	expiration = int64(60 * 60)
)

func ResponseHeader(t *testing.T, newStore storeFactory) {
	r := gin.Default()
	r.Use(ginMiddleware.NewRateLimiterMiddleware(&ginMiddleware.Config{
		Limit:      10,
		Expiration: expiration,
		Store:      newStore(t),
	}))
	r.GET("/", func(c *gin.Context) {
		c.String(200, "Hello World")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-Forwarded-For", "1.2.3.1")
	r.ServeHTTP(res, req)
	XRatelimitRemaining, ok := res.HeaderMap["X-Ratelimit-Remaining"]
	if !ok {
		t.Error("X-Ratelimit-Remaining Header should have been set")
	}
	if XRatelimitRemaining[0] != "9" {
		t.Error("X-Ratelimit-Remaining Error\nWant: 9, Got: ", XRatelimitRemaining[0])
	}
	XRatelimitReset, ok := res.HeaderMap["X-Ratelimit-Reset"]
	if !ok {
		t.Error("X-Ratelimit-Reset Header should have been set")
	}
	now := time.Now().Unix()
	loc := time.FixedZone("", 8*60*60)
	timeString := time.Unix(now+expiration, 0).In(loc).Format("2006-01-02 15:04")
	if XRatelimitReset[0] != timeString {
		t.Error("X-Ratelimit-Reset Error\nWant: ", timeString, ", Got: ", XRatelimitReset[0])
	}
}

func ResponseStatus(t *testing.T, newStore storeFactory) {
	r := gin.Default()
	r.Use(ginMiddleware.NewRateLimiterMiddleware(&ginMiddleware.Config{
		Limit:      1,
		Expiration: 60,
		Store:      newStore(t),
	}))
	r.GET("/", func(c *gin.Context) {
		c.String(200, "Hello World")
	})

	// 1th res
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-Forwarded-For", "1.2.3.2")
	r.ServeHTTP(res, req)
	if res.Code != 200 {
		t.Errorf("Status Code Error:\nWant 200, Got: %d", res.Code)
	}

	// 2nd res
	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Add("X-Forwarded-For", "1.2.3.2")
	r.ServeHTTP(res, req)
	if res.Code != 429 {
		t.Errorf("Status Code Error:\nWant 429, Got: %d", res.Code)
	}
}

func TooManyRequests(t *testing.T, newStore storeFactory) {
	r := gin.Default()
	r.Use(ginMiddleware.NewRateLimiterMiddleware(&ginMiddleware.Config{
		Limit:      reqNum,
		Expiration: expiration,
		Store:      newStore(t),
	}))
	r.GET("/", func(c *gin.Context) {
		c.String(200, "Hello World")
	})

	TooManyRequestsCounter := int32(0)
	wg := sync.WaitGroup{}

	for i := int64(0); i < reqNum+10; i++ {
		wg.Add(1)
		go func(i int64) {
			res := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/", nil)
			req.Header.Add("X-Forwarded-For", "1.2.3.3")
			r.ServeHTTP(res, req)
			if res.Code == 429 {
				atomic.AddInt32(&TooManyRequestsCounter, 1)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	if TooManyRequestsCounter != 10 {
		t.Errorf("TooManyRequestsCounter Error\nWant: 10, Got: %d", TooManyRequestsCounter)
	}
}

func TimeReset(t *testing.T, newStore storeFactory) {
	r := gin.Default()
	r.Use(ginMiddleware.NewRateLimiterMiddleware(&ginMiddleware.Config{
		Limit:      10,
		Expiration: 2,
		Store:      newStore(t),
	}))
	r.GET("/", func(c *gin.Context) {
		c.String(200, "Hello World")
	})

	wg := sync.WaitGroup{}

	for i := int64(0); i < 20; i++ {
		wg.Add(1)
		go func(i int64) {
			res := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/", nil)
			req.Header.Add("X-Forwarded-For", "1.2.3.4")
			r.ServeHTTP(res, req)
			wg.Done()
		}(i)
	}
	wg.Wait()

	time.Sleep(3 * time.Second)
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-Forwarded-For", "1.2.3.4")
	r.ServeHTTP(res, req)

	if res.Code != 200 {
		t.Errorf("TimeReset Error\nWant: 200, Got: %d", res.Code)
	}
}
