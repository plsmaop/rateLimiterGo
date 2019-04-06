package gin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	ratelimiter "github.com/plsmaop/rateLimiterGo"
)

// Config for rate limiter middleware
type Config struct {
	Header     bool
	Limit      int64
	Expiration int64
	Store      ratelimiter.Store

	KeyGetter           KeyGetter
	ErrorHandler        ErrorHandler
	LimitReachedHandler LimitReachedHandler
}

func (c *Config) validateAndNormalize() {
	if c.KeyGetter == nil {
		c.KeyGetter = func(c *gin.Context) string {
			return c.ClientIP()
		}
	}
	if c.ErrorHandler == nil {
		c.ErrorHandler = func(c *gin.Context, err error) {
			c.Status(http.StatusServiceUnavailable)
		}
	}
	if c.LimitReachedHandler == nil {
		c.LimitReachedHandler = func(c *gin.Context) {
			c.String(http.StatusTooManyRequests, "Too Many Requests")
		}
	}
}

// KeyGetter gets key from gin Context
type KeyGetter func(c *gin.Context) string

// ErrorHandler is triggered when an error happens
type ErrorHandler func(c *gin.Context, err error)

// LimitReachedHandler is triggered when the limit is exceed
type LimitReachedHandler func(c *gin.Context)

type middleware struct {
	rateLimiter    ratelimiter.RateLimiter
	onError        ErrorHandler
	onLimitReached LimitReachedHandler
	keyGetter      KeyGetter
	header         bool
}

// NewRateLimiterMiddleware creats an instance of gin rate limiter middleware
func NewRateLimiterMiddleware(c *Config) gin.HandlerFunc {
	c.validateAndNormalize()
	r := ratelimiter.NewRateLimiter(&ratelimiter.Config{
		Limit:      c.Limit,
		Expiration: c.Expiration,
		Store:      c.Store,
	})

	m := &middleware{
		rateLimiter:    r,
		onError:        c.ErrorHandler,
		onLimitReached: c.LimitReachedHandler,
		keyGetter:      c.KeyGetter,
		header:         c.Header,
	}

	return func(ctx *gin.Context) {
		m.handle(ctx)
	}
}

func (m *middleware) handle(ctx *gin.Context) {
	key := m.keyGetter(ctx)
	keyContext, err := m.rateLimiter.Get(key)
	if err != nil {
		m.onError(ctx, err)
		ctx.Abort()
		return
	}

	if m.header {
		ctx.Header("X-RateLimit-Remaining", strconv.FormatInt(keyContext.RemainingCounter, 10))
		// UTC in sec
		ctx.Header("X-RateLimit-Reset", fmt.Sprintf("%d", keyContext.ResetTime))
	}

	if keyContext.IsReachedLimit {
		m.onLimitReached(ctx)
		ctx.Abort()
		return
	}

	ctx.Next()
}
