package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/arion-dsh/jvmao"
)

type CORSOptions struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	// ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSOptions are the default options for the CORS middleware
// allowing all origins, all methods, and the default headers
var DefaultCORSOptions = &CORSOptions{
	AllowedOrigins: []string{"*"},
	AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	// ExposedHeaders:   []string{},
	AllowCredentials: false,
	MaxAge:           172800,
	// AllowedHeaders: []string{"*"},
}

// CORSMiddleware adds the necessary headers for CORS requests
func CORSMiddleware(opt *CORSOptions) jvmao.MiddlewareFunc {

	return func(next jvmao.HandlerFunc) jvmao.HandlerFunc {
		return func(c jvmao.Context) error {
			origins := strings.Join(opt.AllowedOrigins, ", ")
			methods := strings.Join(opt.AllowedMethods, ", ")
			// headers := strings.Join(opt.AllowedHeaders, ", ")
			// Set the headers for CORS
			c.Response().Header().Set("Access-Control-Allow-Origin", origins)
			c.Response().Header().Set("Access-Control-Allow-Methods", methods)
			c.Response().Header().Set("Access-Control-Allow-Headers", "*")
			c.Response().Header().Set("Access-Control-Max-Age", strconv.Itoa(opt.MaxAge))
			c.Response().Header().Set("Access-Control-Allow-Credentials", strconv.FormatBool(opt.AllowCredentials))

			// If this is a preflight request, we want to return immediately
			if c.Request().Method == "OPTIONS" {
				return c.Json(http.StatusOK, map[string]string{"status": "ok"})
			}

			return next(c)
		}
	}
}
