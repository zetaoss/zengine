package providers

import (
	"github.com/zetaoss/zengine/goapp/server/handlers/auth/social/providers/facebook"
	"github.com/zetaoss/zengine/goapp/server/handlers/auth/social/providers/github"
	"github.com/zetaoss/zengine/goapp/server/handlers/auth/social/providers/google"
)

var registry = map[string]Provider{
	"facebook": facebook.Provider{},
	"github":   github.Provider{},
	"google":   google.Provider{},
}

func Get(name string) (Provider, bool) {
	p, ok := registry[name]
	return p, ok
}
