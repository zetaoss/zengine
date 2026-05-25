package throttle

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/zetaoss/zengine/goapp/app/config"
	appredis "github.com/zetaoss/zengine/goapp/app/redis"
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

func Allow(r *http.Request, cfg *config.Config, limit int, window time.Duration, namespace string) (bool, int) {
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
	client := throttleRedisClient(cfg)
	if client == nil {
		return true, 0
	}

	key := ns + ":" + clientIP(r)
	ttlSec := int(window.Seconds())
	if ttlSec < 1 {
		ttlSec = 1
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	res, err := throttleHitScript.Run(ctx, client, []string{fmt.Sprintf("zengine:throttle:v1:%s", key)}, limit, ttlSec).Result()
	if err != nil {
		return true, 0
	}
	arr, ok := res.([]any)
	if !ok || len(arr) < 2 {
		return true, 0
	}
	allowed := asInt64(arr[0]) == 1
	retryAfter := int(asInt64(arr[1]))
	if retryAfter < 1 {
		retryAfter = ttlSec
	}
	return allowed, retryAfter
}

func throttleRedisClient(cfg *config.Config) *goredis.Client {
	throttleRedisOnce.Do(func() {
		client, _ := appredis.Open(cfg)
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
