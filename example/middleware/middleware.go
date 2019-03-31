package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	ginMiddleware "github.com/plsmaop/rateLimiterGo/middleware/gin"
	redis "github.com/plsmaop/rateLimiterGo/store/redis"
)

func middleware() {
	// redisStore
	redisStore, err := redis.NewRedisStore(&redis.Config{
		Host: "127.0.0.1",
		Port: "6379",
		// https://godoc.org/github.com/garyburd/redigo/redis#Pool
		MaxIdle:     100,
		MaxActive:   500,
		IdleTimeout: 240 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	r := gin.Default()
	r.Use(ginMiddleware.NewRateLimiterMiddleware(&ginMiddleware.Config{
		Limit: 1000,
		// in sec
		Expiration: 60 * 60,
		Store:      redisStore,
	}))
}
