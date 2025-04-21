package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Bruno4copos/ratelimiter/config"
	"github.com/Bruno4copos/ratelimiter/limiter"
	"github.com/Bruno4copos/ratelimiter/persistence"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitMiddleware_Handler(t *testing.T) {

	cfg, err := config.LoadConfig("../.")
	assert.Nil(t, err)
	redisStrategy, err := persistence.NewRedisStrategy("localhost:6379", "")
	assert.Nil(t, err)

	rateLimiter := limiter.NewRateLimiter(cfg, redisStrategy)
	rateLimitMiddleware := NewRateLimitMiddleware(rateLimiter)

	handler := rateLimitMiddleware.Handler(http.HandlerFunc(handler))
	http.Handle("/hi", handler)
	go http.ListenAndServe(":8080", handler)

	token1HttpReq := httptest.NewRequest(http.MethodGet, "/hi", nil)
	token1HttpReq.RemoteAddr = "192.168.1.1:12345"
	token1HttpReq.Header.Set("API_KEY", "JOAO123")

	token2HttpReq := httptest.NewRequest(http.MethodGet, "/hi", nil)
	token2HttpReq.RemoteAddr = "192.168.1.1:12345"
	token2HttpReq.Header.Set("API_KEY", "CARLOS456")

	unknowTokenHttpReq := httptest.NewRequest(http.MethodGet, "/hi", nil)
	unknowTokenHttpReq.RemoteAddr = "192.168.1.1:12345"
	unknowTokenHttpReq.Header.Set("API_KEY", "unknowToken")

	ip1HttpReq := httptest.NewRequest(http.MethodGet, "/hi", nil)
	ip1HttpReq.RemoteAddr = "192.168.1.1:12345"

	ip2HttpReq := httptest.NewRequest(http.MethodGet, "/hi", nil)
	ip2HttpReq.RemoteAddr = "192.168.1.2:12345"

	simulateSimultaneousRequests(handler, token1HttpReq, 1)
	simulateSimultaneousRequests(handler, token2HttpReq, 1)
	simulateSimultaneousRequests(handler, unknowTokenHttpReq, 1)
	simulateSimultaneousRequests(handler, ip1HttpReq, 1)
	simulateSimultaneousRequests(handler, ip2HttpReq, 1)

	time.Sleep(1500 * time.Millisecond)

	countBlocked, countSucess := simulateSimultaneousRequests(handler, unknowTokenHttpReq, 30)
	assert.Equal(t, int32(30), countBlocked)
	assert.Equal(t, int32(0), countSucess)

	countBlocked, countSucess = simulateSimultaneousRequests(handler, token2HttpReq, 30)
	assert.Equal(t, int32(30), countBlocked)
	assert.Equal(t, int32(0), countSucess)

	countBlocked, countSucess = simulateSimultaneousRequests(handler, token1HttpReq, 20)
	assert.Equal(t, int32(0), countBlocked)
	assert.Equal(t, int32(20), countSucess)

	countBlocked, countSucess = simulateSimultaneousRequests(handler, ip1HttpReq, 20)
	assert.Equal(t, int32(9), countBlocked)
	assert.Equal(t, int32(11), countSucess)

	countBlocked, countSucess = simulateSimultaneousRequests(handler, ip2HttpReq, 20)
	assert.Equal(t, int32(7), countBlocked)
	assert.Equal(t, int32(13), countSucess)

	time.Sleep(1500 * time.Millisecond)

	countBlocked, countSucess = simulateSimultaneousRequests(handler, ip1HttpReq, 20)
	assert.Equal(t, int32(20), countBlocked)
	assert.Equal(t, int32(0), countSucess)

	// estes deveriam estar desbloqueados
	countBlocked, countSucess = simulateSimultaneousRequests(handler, token1HttpReq, 10)
	assert.Equal(t, int32(0), countBlocked)
	assert.Equal(t, int32(10), countSucess)

	countBlocked, countSucess = simulateSimultaneousRequests(handler, token2HttpReq, 20)
	assert.Equal(t, int32(20), countBlocked)
	assert.Equal(t, int32(0), countSucess)

	countBlocked, countSucess = simulateSimultaneousRequests(handler, unknowTokenHttpReq, 10)
	assert.Equal(t, int32(10), countBlocked)
	assert.Equal(t, int32(0), countSucess)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, it's me!")
}

func simulateSimultaneousRequests(handler http.Handler, req *http.Request, count int) (countBlocked int32, countSucess int32) {
	wg := sync.WaitGroup{}
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			if rr.Code == http.StatusTooManyRequests {
				atomic.AddInt32(&countBlocked, 1)
			} else {
				atomic.AddInt32(&countSucess, 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	return
}
