package redis

import (
	"testing"

	"github.com/zetaoss/zengine/goapp/app/config"
)

func TestAddressUsesFixedRedisPort(t *testing.T) {
	tests := []struct {
		name string
		host string
		want string
	}{
		{name: "configured host", host: "redis-persist", want: "redis-persist:6379"},
		{name: "empty host defaults to localhost", want: "127.0.0.1:6379"},
		{name: "IPv6 host", host: "2001:db8::1", want: "[2001:db8::1]:6379"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Address(config.RedisEndpoint{Host: tt.host}); got != tt.want {
				t.Errorf("Address() = %q, want %q", got, tt.want)
			}
		})
	}
}
