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

// Set устанавливает ключ и значение в Redis
func (c *Redis) Set(key, value string, t ...int) error {
	if key == "" {
		return errors.New("key is nil")
	}

	// Устанавливаем значение по умолчанию для времени жизни ключа
	expiration := time.Minute * 5
	if len(t) > 0 {
		if t[0] <= 0 {
			return errors.New("expiration time must be positive")
		}
		expiration = time.Minute * time.Duration(t[0])
	}

	// Тайм-аут для контекста, равный времени жизни ключа
	ctx, cancel := context.WithTimeout(context.Background(), expiration)
	defer cancel()

	// Устанавливаем ключ в Redis
	return c.Redis.Set(ctx, key, value, expiration).Err()
}

// Get получает значение из Redis по ключу
func (c *Redis) Get(key string) (result string, err error) {
	if key == "" {
		return "", errors.New("key is nil")
	}

	// Тайм-аут для контекста
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Получаем значение из Redis
	result, err = c.Redis.Get(ctx, key).Result()
	if err != nil {
		return
	}

	// Проверяем, что значение не пустое
	if result == "" {
		return "", errors.New("value is nil")
	}

	return
}
