package limiter

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Bruno4copos/ratelimiter/config"
	"github.com/Bruno4copos/ratelimiter/persistence"
)

const apiKeyHeader = "API_KEY"
const blockedMessage = "you have reached the maximum number of requests or actions allowed within a certain time frame"

type RateLimiter struct {
	config      *config.Config
	strategy    persistence.RateLimiterStrategy
	ipPrefix    string
	tokenPrefix string
}

func NewRateLimiter(cfg *config.Config, strategy persistence.RateLimiterStrategy) *RateLimiter {
	return &RateLimiter{
		config:      cfg,
		strategy:    strategy,
		ipPrefix:    "rl:ip:",
		tokenPrefix: "rl:token:",
	}
}

func (rl *RateLimiter) Allow(r *http.Request) bool {
	token := rl.extractToken(r)
	if token != "" {
		return rl.allowByToken(token)
	}

	ipAddress := rl.extractIP(r)
	return rl.allowByIP(ipAddress)
}

func (rl *RateLimiter) extractToken(r *http.Request) string {
	authHeader := r.Header.Get(apiKeyHeader)
	if authHeader != "" {
		return authHeader
	}
	return ""
}

func (rl *RateLimiter) extractIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.RemoteAddr
		// Handle IPv6 addresses by stripping the port
		if strings.Contains(ip, ":") {
			ip = strings.SplitN(ip, ":", 2)[0]
		}
	}
	return ip
}

func (rl *RateLimiter) allowByIP(ipAddress string) bool {
	key := rl.ipPrefix + ipAddress
	now := time.Now()

	data, err := rl.strategy.Get(key)
	if err != nil {
		fmt.Printf("Error getting rate limit data for IP %s: %v\n", ipAddress, err)
		return true // Allow request on error to avoid blocking unnecessarily
	}

	if data != nil && data.BlockedUntil.After(now) {
		return false // IP is blocked
	}

	count, err := rl.strategy.Increment(key)
	if err != nil {
		fmt.Printf("Error incrementing counter for IP %s, Key: %v: %v\n", ipAddress, key, err)
		return true // Allow request on error
	}

	if count > rl.config.MaxRequestsPerSecondIP {
		blockUntil := now.Add(rl.config.BlockDurationIP)
		rl.strategy.Set(key, persistence.RateLimitData{Count: count, BlockedUntil: blockUntil}, rl.config.BlockDurationIP)
		persistence.Count = 0
		persistence.Key = ""
		return false // Exceeded limit, block IP
	}

	// Reset counter periodically (e.g., every second)
	if count == 1 {
		rl.strategy.Set(key, persistence.RateLimitData{Count: count, BlockedUntil: time.Time{}}, time.Second)
	}

	return true // Request allowed
}

func (rl *RateLimiter) allowByToken(token string) bool {
	key := rl.tokenPrefix + token
	now := time.Now()

	data, err := rl.strategy.Get(key)
	if err != nil {
		fmt.Printf("Error getting rate limit data for token %s: %v\n", token, err)
		return true // Allow request on error
	}

	if data != nil && data.BlockedUntil.After(now) {
		return false // Token is blocked
	}

	count, err := rl.strategy.Increment(key)
	if err != nil {
		fmt.Printf("Error incrementing counter for token %s, Key: %v: %v\n", token, key, err)
		return true // Allow request on error
	}

	if count > rl.config.MaxRequestsPerSecondToken {
		blockUntil := now.Add(rl.config.BlockDurationToken)
		rl.strategy.Set(key, persistence.RateLimitData{Count: count, BlockedUntil: blockUntil}, rl.config.BlockDurationToken)
		persistence.Count = 0
		persistence.Key = ""
		return false // Exceeded limit, block token
	}

	// Reset counter periodically
	if count == 1 {
		rl.strategy.Set(key, persistence.RateLimitData{Count: count, BlockedUntil: time.Time{}}, time.Second)
	}

	return true // Request allowed
}

func (rl *RateLimiter) TooManyRequestsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTooManyRequests)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"error": "%s"}`, blockedMessage)
}
