package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

type RedisClient struct {
	Client *redis.Client
}

func NewRedis() *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})
	return &RedisClient{Client: rdb}
}

func (r *RedisClient) Get(key string) (string, error) {
	return r.Client.Get(Ctx, key).Result()
}

func (r *RedisClient) Set(key, value string) error {
	return r.Client.Set(Ctx, key, value, 60*time.Second).Err()
}
