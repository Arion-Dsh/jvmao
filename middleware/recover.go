package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/arion-dsh/jvmao"
)

// Recover from panics, logs the panic (and a backtrace),
// and returns a HTTP 500 (Internal Server Error) status if possible.
func Recover() jvmao.MiddlewareFunc {
	return func(next jvmao.HandlerFunc) jvmao.HandlerFunc {
		return func(c jvmao.Context) error {
			defer func() {
				if rvr := recover(); rvr != nil {

					if rvr == http.ErrAbortHandler {
						panic(rvr)
					}

					err, ok := rvr.(error)
					if !ok {
						err = fmt.Errorf("%v", rvr)
					}

					stack := debug.Stack()
					s := fmt.Sprintf("[PANIC RECOVER] %v %s\n", err, stack)

					c.Logger().Error(s)

					c.Error(err)

				}
			}()
			return next(c)
		}
	}
}
