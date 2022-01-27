package jvmao

import (
	"net/http"
	"strings"
)

type Group struct {
	jm         *Jvmao
	prefix     string
	middleware []MiddlewareFunc
}

func (g *Group) Use(middleware ...MiddlewareFunc) {
	g.middleware = append(g.middleware, middleware...)
}

func (g *Group) Group(prefix string) *Group {
	gp := &Group{prefix: g.prefix + "/" + prefix, jm: g.jm}
	gp.middleware = append(gp.middleware, g.middleware...)
	return gp
}

func (g *Group) CONNECT(name, pattern string, handler HandlerFunc) {
	g.handle(name, http.MethodConnect, pattern, handler)
}
func (g *Group) HEAD(name, pattern string, handler HandlerFunc) {
	g.handle(name, http.MethodHead, pattern, handler)
}
func (g *Group) OPTIONS(name, pattern string, handler HandlerFunc) {
	g.handle(name, http.MethodOptions, pattern, handler)
}
func (g *Group) PATCH(name, pattern string, handler HandlerFunc) {
	g.handle(name, http.MethodPatch, pattern, handler)
}
func (g *Group) GET(name, pattern string, handler HandlerFunc) {
	g.handle(name, http.MethodGet, pattern, handler)
}
func (g *Group) POST(name, pattern string, handler HandlerFunc) {
	g.handle(name, http.MethodPost, pattern, handler)
}
func (g *Group) PUT(name, pattern string, handler HandlerFunc) {
	g.handle(name, http.MethodPut, pattern, handler)
}
func (g *Group) DELETE(name, pattern string, handler HandlerFunc) {
	g.handle(name, http.MethodDelete, pattern, handler)
}
func (g *Group) TRACE(name, pattern string, handler HandlerFunc) {
	g.handle(name, http.MethodTrace, pattern, handler)
}

func (g *Group) handle(name, method, pattern string, h HandlerFunc) {

	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}

	h = applyMiddleware(h, g.middleware...)
	pattern = g.prefix + pattern
	g.jm.handle(name, method, pattern, h)
}
