package storage

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	Redis *redis.Client
}

func (c *Redis) Set(key, value string) error {
	if key == "" {
		return errors.New("key is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	return c.Redis.Set(ctx, key, value, time.Minute*5).Err()
}

func (c *Redis) Get(key string) (string, error) {
	if key == "" {
		return "", errors.New("key is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return c.Redis.Get(ctx, key).Result()
}
