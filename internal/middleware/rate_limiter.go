package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Lua скрипт для атомарного sliding window rate limiter
const rateLimitLua = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local window_key = key .. ":" .. math.floor(now / window)

redis.call("EXPIRE", window_key, window * 2)
local count = tonumber(redis.call("GET", window_key) or "0")

if count + 1 > limit then
	return 0
end

redis.call("INCR", window_key)
return 1
`

type RateLimiterConfig struct {
	Limit          int
	WindowSeconds  int
	BlockThreshold int
	BlockDuration  time.Duration
}

type RateLimiter struct {
	rdb    *redis.Client
	cfg    RateLimiterConfig
	sha    string
}

func NewRateLimiter(rdb *redis.Client, cfg RateLimiterConfig) (*RateLimiter, error) {
	sha, err := rdb.ScriptLoad(context.Background(), rateLimitLua).Result()
	if err != nil {
		return nil, fmt.Errorf("load rate limit script: %w", err)
	}
	return &RateLimiter{rdb: rdb, cfg: cfg, sha: sha}, nil
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		key := fmt.Sprintf("rl:ip:%s", ip)

		// Проверка блокировки
		if blocked, _ := rl.rdb.Exists(r.Context(), key+":blocked").Result(); blocked > 0 {
			w.Header().Set("Retry-After", strconv.Itoa(int(rl.cfg.BlockDuration.Seconds())))
			http.Error(w, "Rate limit exceeded. IP temporarily blocked.", http.StatusTooManyRequests)
			return
		}

		now := time.Now().Unix()
		res, err := rl.rdb.EvalSha(r.Context(), rl.sha, []string{key}, rl.cfg.Limit, rl.cfg.WindowSeconds, now).Int()
		if err != nil || res == 0 {
			// Превышен лимит -> блокируем если превышен порог
			rl.handleViolation(r.Context(), key)
			w.Header().Set("Retry-After", strconv.Itoa(rl.cfg.WindowSeconds))
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) handleViolation(ctx context.Context, key string) {
	// Экспоненциальный backoff: увеличиваем счетчик нарушений
	failKey := key + ":fails"
	fails, _ := rl.rdb.Incr(ctx, failKey).Result()
	rl.rdb.Expire(ctx, failKey, rl.cfg.BlockDuration)

	if int(fails) >= rl.cfg.BlockThreshold {
		rl.rdb.Set(ctx, key+":blocked", "1", rl.cfg.BlockDuration)
	}
}