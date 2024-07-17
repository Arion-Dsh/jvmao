package jvmao

import (
	"net/http"
	"sync"
)

type mux struct {
	serverMux       *http.ServeMux
	mu              sync.RWMutex
	pool            sync.Pool
	route           *routeChache
	notFoundHandler HandlerFunc
	httpErrHandler  HTTPErrorHandler
}

// newMux returns a new Mux object.
func newMux() *mux {
	mux := &mux{
		serverMux: http.NewServeMux(),
		mu:        sync.RWMutex{},
		route:     newRouteChache(),
	}

	mux.pool = sync.Pool{New: func() interface{} { return &context{w: NewResponse(nil), route: mux.route} }}
	return mux
}

func (m *mux) SetRoute(name, pattern string) {
	m.route.SetRoute(name, pattern)
}

func (m *mux) Static(pattern, dir string) {
	m.serverMux.Handle(pattern, http.StripPrefix(pattern, http.FileServer(http.Dir(dir))))
}

// Handle registers the handler for the given pattern.
func (m *mux) Handle(pattern, method string, handlerFunc HandlerFunc) {

	m.serverMux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		ctx := m.pool.Get().(*context)
		defer m.pool.Put(ctx)
		ctx.reset(w, r)
		err := handlerFunc(ctx)
		if err != nil {
			m.httpErrHandler(err, ctx)
		}
	})
}

func (m *mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.serverMux.ServeHTTP(w, r)
}
