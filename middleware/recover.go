package middleware

import (
	"fmt"
	"jvmao"
	"net/http"
	"runtime/debug"
)

func Recover() jvmao.MiddlewareFunc {
	return func(next jvmao.HandlerFunc) jvmao.HandlerFunc {
		return func(c *jvmao.Context) error {
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
