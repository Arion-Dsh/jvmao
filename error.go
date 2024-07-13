package jvmao

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
)

// HTTPError represents an error.
type HTTPError struct {
	Code    int         `json:"code"`
	Message interface{} `json:"msg"`
}

// Error fit for error interface
func (he *HTTPError) Error() string {
	return fmt.Sprintf("code=%d, message=%v", he.Code, he.Message)
}

func NewHTTPError(statusCode int, msg string) error {
	return &HTTPError{statusCode, msg}

}

func NewHTTPErrorWithError(err error) error {

	if _, ok := err.(*HTTPError); ok {
		return err
	}

	var msg string
	var code int
	switch {
	case errors.Is(err, fs.ErrNotExist):
		msg, code = "404 page not found", http.StatusNotFound
	case errors.Is(err, fs.ErrPermission):
		msg, code = "403 Forbidden", http.StatusForbidden
	default:
		msg, code = "500 Internal Server Error", http.StatusInternalServerError
	}
	return NewHTTPError(code, msg)
}
