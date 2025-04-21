package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisStrategy struct {
	client *redis.Client
}

var (
	Count int
	Key   string
	t     = time.Now()
)

func NewRedisStrategy(address, password string) (*RedisStrategy, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       0, // Use default DB
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStrategy{client: client}, nil
}

func (r *RedisStrategy) Increment(key string) (int, error) {
	taux := time.Now()
	if taux.Sub(t) >= time.Second {
		t = taux
		Count = 0
	}
	if Key == "" {
		Key = key
		Count = 0
	}
	if key == Key {
		Count++
	} else {
		Count = 0
		Key = key
	}
	return Count, nil
}

func (r *RedisStrategy) Get(key string) (*RateLimitData, error) {
	val, err := r.client.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var data RateLimitData
	err = json.Unmarshal([]byte(val), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal rate limit data: %w", err)
	}

	return &data, nil
}

func (r *RedisStrategy) Set(key string, data RateLimitData, expiration time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal rate limit data: %w", err)
	}

	err = r.client.Set(context.Background(), key, jsonData, expiration).Err()
	return err
}
