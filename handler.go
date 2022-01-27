package jvmao

import (
	"net/http"
)

type HandlerFunc func(*Context) error

type HTTPErrorHandler func(err error, c *Context)

func defaultHttpErrorHandler(err error, c *Context) {
	code := http.StatusInternalServerError
	if h, ok := err.(*HTTPError); ok {
		code = h.Code
	}
	c.w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.w.Header().Set("X-Content-Type-Options", "nosniff")
	c.w.WriteHeader(code)
	c.w.Write([]byte(err.Error()))
	return
}

func defaultNotFoundHandler(c *Context) error {
	c.w.Header().Set("X-Content-Type-Options", "nosniff")
	return c.String(http.StatusNotFound, "not found")
}

func defaultMethodNotAllowedHandler(c *Context) error {
	c.w.Header().Set("X-Content-Type-Options", "nosniff")
	return c.String(http.StatusMethodNotAllowed, "method not allowed")
}
