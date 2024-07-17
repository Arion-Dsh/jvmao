package jvmao

import (
	ctx "context"
	"crypto/tls"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
)

// New return a instance of Jvmao.
func New() *Jvmao {
	jm := &Jvmao{
		hs:    new(http.Server),
		tlsHs: new(http.Server),
		AutoTLSManager: autocert.Manager{
			// Cache:  autocert.DirCache("secret-dir"),
			Prompt: autocert.AcceptTOS,
			// HostPolicy: autocert.HostWhitelist("example.org", "www.example.org"),
		},
		tcpAlivePeriod: time.Minute * 3,

		grpc: NewGrpcHandler(),

		renderer: new(DefaultRenderer),
	}
	jm.Logger = DefaultLogger()
	jm.mux = newMux()
	jm.mux.httpErrHandler = DefaultHttpErrorHandler
	jm.mux.notFoundHandler = DefaultNotFoundHandler
	return jm
}

// Jumao top-level instance.
type Jvmao struct {
	mu sync.Mutex

	mux *mux

	grpc *GrpcHandler

	hs             *http.Server
	tlsHs          *http.Server
	AutoTLSManager autocert.Manager

	tcpAlivePeriod time.Duration

	middleware []MiddlewareFunc
	renderer   Renderer
	Logger     *Logger

	debug bool
}

func (jm *Jvmao) SetNotFoundHandler(h HandlerFunc) {
	jm.mux.notFoundHandler = h
}

func (jm *Jvmao) SetHTTPErrorHandler(h HTTPErrorHandler) {
	jm.mux.httpErrHandler = h
}

func (jm *Jvmao) RegisterGrpcServer(s *grpc.Server) {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	jm.grpc.RegisterGrpcServer(s)
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

func (jm *Jvmao) Static(prefix string, dir string) {
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	dir = path.Dir(dir)
	jm.mux.Static(prefix, dir)

}

func (jm *Jvmao) FileFS(file string, fsys fs.FS) {
	jm.GET(file, file, func(c Context) error {
		return c.FileFS(file, fsys)
	})
}

func (jm *Jvmao) File(file, dir string) {
	jm.GET(file, file, func(c Context) error {
		return c.File(file, http.Dir(filepath.Dir(dir)))
	})
}

func (jm *Jvmao) CONNECT(pattern, name string, handler HandlerFunc) {
	jm.handle(name, http.MethodConnect, pattern, handler)
}
func (jm *Jvmao) HEAD(pattern, name string, handler HandlerFunc) {
	jm.handle(name, http.MethodHead, pattern, handler)
}
func (jm *Jvmao) OPTIONS(pattern, name string, handler HandlerFunc) {
	jm.handle(name, http.MethodOptions, pattern, handler)
}
func (jm *Jvmao) PATCH(pattern, name string, handler HandlerFunc) {
	jm.handle(name, http.MethodPatch, pattern, handler)
}
func (jm *Jvmao) GET(pattern, name string, handler HandlerFunc) {
	jm.handle(name, http.MethodGet, pattern, handler)
}
func (jm *Jvmao) POST(pattern, name string, handler HandlerFunc) {
	jm.handle(name, http.MethodPost, pattern, handler)
}
func (jm *Jvmao) PUT(pattern, name string, handler HandlerFunc) {
	jm.handle(name, http.MethodPut, pattern, handler)
}
func (jm *Jvmao) DELETE(pattern, name string, handler HandlerFunc) {
	jm.handle(name, http.MethodDelete, pattern, handler)
}

func (jm *Jvmao) TRACE(pattern, name string, handler HandlerFunc) {
	jm.handle(name, http.MethodTrace, pattern, handler)
}

func (jm *Jvmao) handle(_, method, pattern string, h HandlerFunc) {

	if method == "" {
		method = http.MethodGet
	}
	// p := fmt.Sprintf("%s %s", method, pattern)
	h = applyMiddleware(h, jm.middleware...)
	jm.mux.Handle(pattern, method, h)
}

// Debug show debug is open or not.
func (jm *Jvmao) Debug() bool {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	return jm.debug
}

// Opendebug open debug.
func (jm *Jvmao) OpenDebug() {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	jm.debug = true
	// jm.Logger.SetPriority(LOG_PRINT)
}

func (jm *Jvmao) SetTCPAlivePeriod(d time.Duration) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	jm.tcpAlivePeriod = d

}

// Start start an http/2 server with h2c.
func (jm *Jvmao) Start(addr string) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	h2s := new(http2.Server)
	jm.hs.Addr = addr
	jm.hs.Handler = h2c.NewHandler(jm, h2s)
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
	jm.tlsHs.Handler = jm

	_ = http2.ConfigureServer(jm.tlsHs, &http2.Server{
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

// StartAutoTLS start an HTTPS server using certificates automatically from https://letsencrypt.org.
// you can change certificate with jm.AutoTLSManager before start
// server, such as HostPolicy.
// more [autocert.Manager]https://pkg.go.dev/golang.org/x/crypto/acme/autocert#Manager
func (jm *Jvmao) StartAutoTLS(addr string) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	jm.tlsHs.Addr = addr
	jm.tlsHs.Handler = jm

	jm.tlsHs.TLSConfig = &tls.Config{
		GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			cert, err := jm.AutoTLSManager.GetCertificate(hello)
			if err != nil {
				jm.Logger.Error(fmt.Sprintf("Jvmao: GetCertificate: %v", err))
			}
			return cert, err
		},
	}

	jm.tlsHs.TLSConfig.NextProtos = append(jm.tlsHs.TLSConfig.NextProtos, acme.ALPNProto)

	_ = http2.ConfigureServer(jm.tlsHs, &http2.Server{
		NewWriteScheduler: func() http2.WriteScheduler {
			return http2.NewPriorityWriteScheduler(nil)
		},
	})

	ln, err := jm.newListener(addr)
	if err != nil {
		return err
	}

	return jm.tlsHs.Serve(tls.NewListener(ln, jm.tlsHs.TLSConfig))
}

// Shutdown stops the server gracefully.
// calls `http.Server#Shutdown()`.
//
// ie. look like:
//
//	quit := make(chan os.Signal, 1)
//	signal.Notify(quit, os.Interrupt)
//	<-quit
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//	if err := jm.Shutdown(ctx); err != nil {
//		jm.Logger.Fatal(err)
//	}
func (jm *Jvmao) Shutdown(ctx ctx.Context) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	if err := jm.tlsHs.Shutdown(ctx); err != nil {
		return err
	}
	return jm.hs.Shutdown(ctx)
}

// ServeHTTP implements http.Handler
// more see [http.Handler]https://pkg.go.dev/net/http#Handler
func (jm *Jvmao) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if IsGrpc(r) {
		jm.grpc.ServeHTTP(w, r)
		return
	}
	jm.mux.ServeHTTP(w, r)
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
	_ = tc.SetKeepAlive(true)
	_ = tc.SetKeepAlivePeriod(ln.alivePeriod)
	return tc, nil
}
