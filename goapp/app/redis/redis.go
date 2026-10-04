package redis

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app/config"

	"github.com/hibiken/asynq"
	goredis "github.com/redis/go-redis/v9"
)

// Address returns host:port (or a redis:// URI) for an endpoint, defaulting to 127.0.0.1:6379.
func Address(ep config.RedisEndpoint) string {
	addr := "127.0.0.1:6379"
	host := strings.TrimSpace(ep.Host)
	if host == "" {
		return addr
	}
	port := ep.Port
	if port <= 0 {
		port = 6379
	}
	if strings.HasPrefix(host, "redis://") || strings.HasPrefix(host, "rediss://") || strings.Contains(host, ":") {
		return host
	}
	return net.JoinHostPort(host, fmt.Sprintf("%d", port))
}

// AsynqConnOpt connects task queues to the persist Redis: queued tasks must not be evicted.
func AsynqConnOpt(cfg *config.Config) (asynq.RedisConnOpt, error) {
	addr := Address(cfg.Redis.Persist)
	if strings.HasPrefix(addr, "redis://") || strings.HasPrefix(addr, "rediss://") {
		return asynq.ParseRedisURI(addr)
	}
	return asynq.RedisClientOpt{Addr: addr}, nil
}

// OpenPersist opens the Redis for data that must not be lost (tokens, rate-limit counters).
func OpenPersist(cfg *config.Config) (*goredis.Client, error) {
	return open(cfg.Redis.Persist)
}

// OpenVolatile opens the Redis for data that may be evicted (caches).
func OpenVolatile(cfg *config.Config) (*goredis.Client, error) {
	return open(cfg.Redis.Volatile)
}

func open(ep config.RedisEndpoint) (*goredis.Client, error) {
	addr := Address(ep)

	var opts *goredis.Options
	var err error
	if strings.HasPrefix(addr, "redis://") || strings.HasPrefix(addr, "rediss://") {
		opts, err = goredis.ParseURL(addr)
		if err != nil {
			return nil, err
		}
	} else {
		opts = &goredis.Options{Addr: addr}
	}

	client := goredis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
