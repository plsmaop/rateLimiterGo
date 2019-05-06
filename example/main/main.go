package main

import (
	"fmt"
	"time"

	ratelimiter "github.com/plsmaop/rateLimiterGo"
	memstore "github.com/plsmaop/rateLimiterGo/store/memStore"
	redis "github.com/plsmaop/rateLimiterGo/store/redis"
)

var key = "KEY"

func main() {
	// memStore
	memStore := memstore.NewMemStore()
	rateLimiter := ratelimiter.NewRateLimiter(&ratelimiter.Config{
		Limit: 1000,
		// in sec
		Expiration: 60 * 60,

		Store: memStore,
	})

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

	rateLimiter = ratelimiter.NewRateLimiter(&ratelimiter.Config{
		Limit: 1000,
		// in sec
		Expiration: 60 * 60,

		Store: redisStore,
	})

	// Get will increment the counter of the key
	// and return the Context of the key
	c, err := rateLimiter.Get(key)
	if err != nil {
		panic(err)
	}

	if !c.IsReachedLimit {
		fmt.Printf("The counter of the key is %d\n", c.CurrentCounter)
		fmt.Printf("The limit will be reset in %d second(s)\n", c.TTL)
	}
}
