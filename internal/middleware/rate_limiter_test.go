package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/diasoft/diplomaverify/internal/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter_AllowAndBlock(t *testing.T) {
	mr, err := miniredis.RunT(t)
	require.NoError(t, err)
	
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	rl, err := middleware.NewRateLimiter(rdb, middleware.RateLimiterConfig{
		Limit: 2, WindowSeconds: 10, BlockThreshold: 2, BlockDuration: time.Minute,
	})
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mw := rl.Middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.2.3.4:1234"

	// 1 & 2 запросы проходят
	assert.Equal(t, http.StatusOK, doRequest(mw, req))
	assert.Equal(t, http.StatusOK, doRequest(mw, req))

	// 3 запрос блокируется
	assert.Equal(t, http.StatusTooManyRequests, doRequest(mw, req))
	
	// Проверка блокировки IP
	blockedKey := "rl:ip:1.2.3.4:1234:blocked"
	exists, _ := rdb.Exists(context.Background(), blockedKey).Result()
	assert.Equal(t, int64(1), exists)
}

func doRequest(h http.Handler, r *http.Request) int {
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, r)
	return rr.Code
}