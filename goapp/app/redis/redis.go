package redis

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app/config"

	"github.com/hibiken/asynq"
	goredis "github.com/redis/go-redis/v9"
)

// Address returns the endpoint host on Redis's fixed port 6379.
func Address(ep config.RedisEndpoint) string {
	host := strings.TrimSpace(ep.Host)
	if host == "" {
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, "6379")
}

// AsynqConnOpt connects task queues to the persist Redis: queued tasks must not be evicted.
func AsynqConnOpt(cfg *config.Config) (asynq.RedisConnOpt, error) {
	return asynq.RedisClientOpt{Addr: Address(cfg.Redis.Persist)}, nil
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
	client := goredis.NewClient(&goredis.Options{Addr: Address(ep)})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
