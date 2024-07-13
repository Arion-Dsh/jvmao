package jvmao

import (
	"net/http"
)

// HandlerFunc responds to an HTTP request.
type HandlerFunc func(Context) error

type HTTPErrorHandler func(err error, c Context)

func DefaultHttpErrorHandler(err error, c Context) {
	code := http.StatusInternalServerError
	if h, ok := err.(*HTTPError); ok {
		code = h.Code
	}
	c.Response().Header().Add(HeaderContentType, MIMETextPlainUTF8)
	c.Response().Header().Add(HeaderXContentTypeOptions, "nosniff")
	c.WriteHeader(code)
	_, _ = c.Response().Write([]byte(err.Error()))
}

func DefaultHttpJsonErrorHandler(err error, c Context) {
	code := http.StatusInternalServerError
	if h, ok := err.(*HTTPError); ok {
		code = h.Code
	} else {
		err = NewHTTPError(code, err.Error())
	}
	_ = c.Json(code, err)
}

func DefaultNotFoundHandler(c Context) error {
	c.Response().Header().Set(HeaderXContentTypeOptions, "nosniff")
	return c.String(http.StatusNotFound, "not found")
}

func DefaultMethodNotAllowedHandler(c Context) error {
	c.Response().Header().Add(HeaderXContentTypeOptions, "nosniff")
	return c.String(http.StatusMethodNotAllowed, "method not allowed")
}
