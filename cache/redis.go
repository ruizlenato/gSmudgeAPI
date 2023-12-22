package cache

import (
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
)

var rdb *redis.Client

func init() {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:6379", redisHost),
		Password: "",
		DB:       0,
	})
}

func GetRedisClient() *redis.Client {
	return rdb
}
