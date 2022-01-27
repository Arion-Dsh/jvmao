package jvmao

type MiddlewareFunc func(HandlerFunc) HandlerFunc

func applyMiddleware(h HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
	for _, m := range middleware {
		h = m(h)
	}
	return h
}
