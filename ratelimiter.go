package ratelimiter

import (
	"fmt"
	"net/http"

	"github.com/Bruno4copos/ratelimiter/config"
	"github.com/Bruno4copos/ratelimiter/limiter"
	"github.com/Bruno4copos/ratelimiter/middleware"
	"github.com/Bruno4copos/ratelimiter/persistence"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	redisStrategy, err := persistence.NewRedisStrategy(cfg.RedisAddress, cfg.RedisPassword)
	if err != nil {
		fmt.Printf("Error creating Redis strategy: %v\n", err)
		return
	}

	rateLimiter := limiter.NewRateLimiter(cfg, redisStrategy)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(rateLimiter)

	http.Handle("/", rateLimitMiddleware.Handler(http.HandlerFunc(handler)))

	fmt.Println("Server listening on port 8080...")
	http.ListenAndServe(":8080", nil)
}
