package jvmao

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func New() *Jvmao {
	// logger := log.New(os.Stdout, " jvmao: ", 0)
	jm := &Jvmao{
		hs: new(http.Server),

		renderer:        new(DefaultRenderer),
		NotFoundHandler: defaultNotFoundHandler,
		HTTPErrHandler:  defaultHttpErrorHandler,
	}
	jm.Logger = DefaultLogger()
	jm.pool = sync.Pool{New: func() interface{} {
		return &Context{jm: jm, w: NewResponse(jm, nil)}
	}}
	jm.mux = newMux(jm)

	return jm
}

// Jumao ...
type Jvmao struct {
	mu sync.Mutex

	mux *mux
	hs  *http.Server
	h2s *http2.Server

	pool sync.Pool

	middleware      []MiddlewareFunc
	renderer        Renderer
	NotFoundHandler HandlerFunc
	HTTPErrHandler  HTTPErrorHandler
	Logger          Logger

	debug bool
}

func (jm *Jvmao) SetRenderer(r Renderer) {
	jm.renderer = r
}

func (jm *Jvmao) Use(middleware ...MiddlewareFunc) {
	jm.middleware = append(jm.middleware, middleware...)
}

func (jm *Jvmao) Group(prefix string) *Group {
	prefix = strings.Trim(prefix, "/")
	return &Group{prefix: prefix, jm: jm}
}

func (jm *Jvmao) Static(root, prefix string) {
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	jm.mux.setStatic(root, prefix)
}

func (jm *Jvmao) CONNECT(name, pattern string, handler HandlerFunc) {
	jm.handle(name, http.MethodConnect, pattern, handler)
}
func (jm *Jvmao) HEAD(name, pattern string, handler HandlerFunc) {
	jm.handle(name, http.MethodHead, pattern, handler)
}
func (jm *Jvmao) OPTIONS(name, pattern string, handler HandlerFunc) {
	jm.handle(name, http.MethodOptions, pattern, handler)
}
func (jm *Jvmao) PATCH(name, pattern string, handler HandlerFunc) {
	jm.handle(name, http.MethodPatch, pattern, handler)
}
func (jm *Jvmao) GET(name, pattern string, handler HandlerFunc) {
	jm.handle(name, http.MethodGet, pattern, handler)
}
func (jm *Jvmao) POST(name, pattern string, handler HandlerFunc) {
	jm.handle(name, http.MethodPost, pattern, handler)
}
func (jm *Jvmao) PUT(name, pattern string, handler HandlerFunc) {
	jm.handle(name, http.MethodPut, pattern, handler)
}
func (jm *Jvmao) DELETE(name, pattern string, handler HandlerFunc) {
	jm.handle(name, http.MethodDelete, pattern, handler)
}

func (jm *Jvmao) TRACE(name, pattern string, handler HandlerFunc) {
	jm.handle(name, http.MethodTrace, pattern, handler)
}

func (jm *Jvmao) handle(name, method, pattern string, h HandlerFunc) {

	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}

	jm.mux.handle(name, method, pattern, h)
}

func (jm *Jvmao) Reverse(name string, params ...interface{}) string {
	uri := new(bytes.Buffer)

	l := len(params)
	if p, ok := jm.mux.routes[name]; ok {
		c := strings.Count(p, ":")
		if c != l {
			panic(fmt.Sprintf("Reverse error: need %d params but get %d", c, l))
		}
		for i := 0; i < l; i++ {
			ps := strings.Split(p, ":")
			uri.WriteString(ps[0])
			uri.WriteString(fmt.Sprintf("%v", params[i]))
			p = p[strings.Index(p, "/"):]
		}
	}

	return uri.String()
}

func (jm *Jvmao) Debug() bool {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	return jm.debug
}

func (jm *Jvmao) OpenDebug() {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	jm.debug = true
	jm.Logger.SetPriority(LOG_PRINT)

}
func (jm *Jvmao) Start(addr string) error {

	h2s := new(http2.Server)
	// jm.hs.ErrorLog =
	jm.hs.Addr = addr
	jm.hs.Handler = h2c.NewHandler(jm.mux, h2s)

	return jm.hs.ListenAndServe()

}
