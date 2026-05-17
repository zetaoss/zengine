package redis

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app/config"

	goredis "github.com/redis/go-redis/v9"
)

func Open(cfg *config.Config) (*goredis.Client, error) {
	addr := "127.0.0.1:6379"
	if cfg != nil {
		if cfg.Redis.Host != "" {
			host := strings.TrimSpace(cfg.Redis.Host)
			port := cfg.Redis.Port
			if port <= 0 {
				port = 6379
			}
			if host != "" {
				if strings.HasPrefix(host, "redis://") || strings.HasPrefix(host, "rediss://") || strings.Contains(host, ":") {
					addr = host
				} else {
					addr = net.JoinHostPort(host, fmt.Sprintf("%d", port))
				}
			}
		}
	}

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
