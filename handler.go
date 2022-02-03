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
	c.Response().Header().Add(HeaderContentType, MIMETextPlainUTF8)
	c.Response().Header().Add(HeaderXContentTypeOptions, "nosniff")
	c.WriteHeader(code)
	c.Response().Write([]byte(err.Error()))
	return
}

func defaultNotFoundHandler(c Context) error {
	c.Response().Header().Set(HeaderXContentTypeOptions, "nosniff")
	return c.String(http.StatusNotFound, "not found")
}

func defaultMethodNotAllowedHandler(c Context) error {
	c.Response().Header().Add(HeaderXContentTypeOptions, "nosniff")
	return c.String(http.StatusMethodNotAllowed, "method not allowed")
}
