package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/zetaoss/zengine/goapp/app/config"
	appredis "github.com/zetaoss/zengine/goapp/app/redis"

	goredis "github.com/redis/go-redis/v9"
)

var (
	throttleRedisOnce sync.Once
	throttleRedis     *goredis.Client
)

var throttleHitScript = goredis.NewScript(`
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local ttlSec = tonumber(ARGV[2])
local n = redis.call("INCR", key)
if n == 1 then
  redis.call("EXPIRE", key, ttlSec)
end
local ttl = redis.call("TTL", key)
if ttl < 1 then
  ttl = ttlSec
end
if n > limit then
  return {0, ttl}
end
return {1, ttl}
`)

func Throttle(cfg *config.Config, limit int, window time.Duration, namespace string) Middleware {
	if limit < 1 {
		limit = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	ns := strings.TrimSpace(namespace)
	if ns == "" {
		ns = "default"
	}
	redisClient := throttleRedisClient(cfg)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := ns + ":" + clientIP(r)
			allowed, retryAfter := hitThrottle(redisClient, key, limit, window)
			if !allowed {
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				writeErrorJSON(w, http.StatusTooManyRequests, "Too Many Requests")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func hitThrottle(client *goredis.Client, key string, limit int, window time.Duration) (bool, int) {
	if client == nil {
		return true, 0
	}
	allowed, retryAfter, ok := hitRedisThrottle(client, key, limit, window)
	if !ok {
		return true, 0
	}
	return allowed, retryAfter
}

func hitRedisThrottle(client *goredis.Client, key string, limit int, window time.Duration) (bool, int, bool) {
	ttlSec := int(window.Seconds())
	if ttlSec < 1 {
		ttlSec = 1
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	res, err := throttleHitScript.Run(ctx, client, []string{fmt.Sprintf("zengine:throttle:v1:%s", key)}, limit, ttlSec).Result()
	if err != nil {
		return false, 0, false
	}
	arr, ok := res.([]any)
	if !ok || len(arr) < 2 {
		return false, 0, false
	}
	allowed := asInt64(arr[0]) == 1
	retryAfter := int(asInt64(arr[1]))
	if retryAfter < 1 {
		retryAfter = ttlSec
	}
	return allowed, retryAfter, true
}

func throttleRedisClient(cfg *config.Config) *goredis.Client {
	throttleRedisOnce.Do(func() {
		if cfg == nil {
			return
		}
		client, err := appredis.Open(cfg)
		if err != nil {
			return
		}
		throttleRedis = client
	})
	return throttleRedis
}

func asInt64(v any) int64 {
	switch x := v.(type) {
	case int64:
		return x
	case int:
		return int64(x)
	case uint64:
		return int64(x)
	case string:
		n, _ := strconv.ParseInt(strings.TrimSpace(x), 10, 64)
		return n
	default:
		return 0
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i != -1 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return host
}
