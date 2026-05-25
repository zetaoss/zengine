package router

import (
	"net/http"

	"github.com/zetaoss/zengine/goapp/server/serverctx"
)

type HandlerFunc func(c *serverctx.Context)

type Route struct {
	Method string
	Path   string
}

type Router struct {
	mux       *http.ServeMux
	serverCtx *serverctx.Context
	routes    []Route
}

func New(mux *http.ServeMux, serverCtx *serverctx.Context) *Router {
	return &Router{mux: mux, serverCtx: serverCtx}
}

func (r *Router) GET(path string, h HandlerFunc, m ...Middleware) {
	r.handle(http.MethodGet, path, h, m...)
}

func (r *Router) POST(path string, h HandlerFunc, m ...Middleware) {
	r.handle(http.MethodPost, path, h, m...)
}

func (r *Router) DELETE(path string, h HandlerFunc, m ...Middleware) {
	r.handle(http.MethodDelete, path, h, m...)
}

func (r *Router) PUT(path string, h HandlerFunc, m ...Middleware) {
	r.handle(http.MethodPut, path, h, m...)
}

func (r *Router) handle(method string, path string, h HandlerFunc, m ...Middleware) {
	r.routes = append(r.routes, Route{Method: method, Path: path})
	pattern := method + " " + path
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		c := *r.serverCtx
		c.W = w
		c.R = req
		h(&c)
	})
	r.mux.Handle(pattern, Chain(handler, m...))
}

func (r *Router) Routes() []Route {
	return r.routes
}
