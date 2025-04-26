package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisAddr = os.Getenv("REDIS_ADDR")
var RedisPassword = os.Getenv("REDIS_PASSWORD") // set address and password through docker

func InitRedis() *redis.Client {
	if RedisAddr == "" {
		panic("REDIS_ADDR not set in environment")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     RedisAddr,
		Password: RedisPassword,
		DB:       0, // use default DB
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}

	fmt.Println("Redis successfully connected")
	return rdb
}

var RedisClient *redis.Client = InitRedis()
