package common

import "net/http"

type Middleware func(http.Handler) http.Handler

type Components struct {
	AccessLogMiddleware Middleware
	RootHandler         http.Handler
}

func (c *Components) Wrap(handler http.Handler) http.Handler {
	if c.AccessLogMiddleware != nil {
		handler = c.AccessLogMiddleware(handler)
	}
	return handler
}
