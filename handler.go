package jvmao

import (
	"net/http"
)

// HandlerFunc responds to an HTTP request.
type HandlerFunc func(Context) error

type HTTPErrorHandler func(err error, c Context)

func defaultHttpErrorHandler(err error, c Context) {
	code := http.StatusInternalServerError
	if h, ok := err.(*HTTPError); ok {
		code = h.Code
	}
	c.SetHeader("Content-Type", "text/plain; charset=utf-8")
	c.SetHeader("X-Content-Type-Options", "nosniff")
	c.WriteHeader(code)
	c.Response().Write([]byte(err.Error()))
	return
}

func defaultNotFoundHandler(c Context) error {
	c.Response().Header().Set("X-Content-Type-Options", "nosniff")
	return c.String(http.StatusNotFound, "not found")
}

func defaultMethodNotAllowedHandler(c Context) error {
	c.Response().Header().Set("X-Content-Type-Options", "nosniff")
	return c.String(http.StatusMethodNotAllowed, "method not allowed")
}
