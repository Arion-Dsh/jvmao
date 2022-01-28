package jvmao

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func New() *Jvmao {
	// logger := log.New(os.Stdout, " jvmao: ", 0)
	jm := &Jvmao{
		hs:             new(http.Server),
		tlsHs:          new(http.Server),
		tcpAlivePeriod: time.Minute * 3,

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
	mu   sync.Mutex
	pool sync.Pool

	mux *mux

	// listener       net.Listener
	// tlsListener    net.Listener

	hs    *http.Server
	tlsHs *http.Server

	tcpAlivePeriod time.Duration

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

func (jm *Jvmao) SetTCPAlivePeriod(d time.Duration) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	jm.tcpAlivePeriod = d

}

//Start start an http/2 server with h2c.
func (jm *Jvmao) Start(addr string) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	h2s := new(http2.Server)
	jm.hs.Addr = addr
	jm.hs.Handler = h2c.NewHandler(jm.mux, h2s)
	ln, err := jm.newListener(addr)
	if err != nil {
		return err
	}
	return jm.hs.Serve(ln)

}

// StartTLS start an HTTPS server.
func (jm *Jvmao) StartTLS(addr string, certFile, keyFile string) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	jm.tlsHs.Addr = addr
	jm.tlsHs.Handler = jm.mux

	http2.ConfigureServer(jm.tlsHs, &http2.Server{
		NewWriteScheduler: func() http2.WriteScheduler {
			return http2.NewPriorityWriteScheduler(nil)
		},
	})

	ln, err := jm.newListener(addr)
	if err != nil {
		return err
	}

	return jm.tlsHs.ServeTLS(ln, certFile, keyFile)
}

func (jm *Jvmao) newListener(addr string) (*tcpKeepAliveListener, error) {

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &tcpKeepAliveListener{ln.(*net.TCPListener), jm.tcpAlivePeriod}, nil

}

type tcpKeepAliveListener struct {
	*net.TCPListener
	alivePeriod time.Duration
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(ln.alivePeriod)
	return tc, nil
}
