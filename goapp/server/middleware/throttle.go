package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type throttleState struct {
	mu      sync.Mutex
	buckets map[string]*throttleBucket
}

type throttleBucket struct {
	count   int
	expires time.Time
}

var globalThrottleState = &throttleState{
	buckets: map[string]*throttleBucket{},
}

func Throttle(limit int, window time.Duration, namespace string) Middleware {
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

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := ns + ":" + clientIP(r)
			allowed, retryAfter := globalThrottleState.hit(key, limit, window)
			if !allowed {
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				writeErrorJSON(w, http.StatusTooManyRequests, "Too Many Requests")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (s *throttleState) hit(key string, limit int, window time.Duration) (bool, int) {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.buckets) > 1000 {
		for k, b := range s.buckets {
			if now.After(b.expires) {
				delete(s.buckets, k)
			}
		}
	}

	b, ok := s.buckets[key]
	if !ok || now.After(b.expires) {
		s.buckets[key] = &throttleBucket{
			count:   1,
			expires: now.Add(window),
		}
		return true, 0
	}
	if b.count >= limit {
		remain := int(time.Until(b.expires).Seconds())
		if remain < 1 {
			remain = 1
		}
		return false, remain
	}
	b.count++
	return true, 0
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
