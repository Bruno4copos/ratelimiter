package persistence

import "time"

type RateLimitData struct {
	Count        int
	BlockedUntil time.Time
}

type RateLimiterStrategy interface {
	Increment(key string) (int, error)
	Get(key string) (*RateLimitData, error)
	Set(key string, data RateLimitData, expiration time.Duration) error
}
