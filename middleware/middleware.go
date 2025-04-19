package middleware

import (
	"net/http"

	"github.com/Bruno4copos/ratelimiter/limiter"
)

type RateLimitMiddleware struct {
	limiter *limiter.RateLimiter
}

func NewRateLimitMiddleware(l *limiter.RateLimiter) *RateLimitMiddleware {
	return &RateLimitMiddleware{limiter: l}
}

func (rlm *RateLimitMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rlm.limiter.Allow(r) {
			rlm.limiter.TooManyRequestsHandler(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
