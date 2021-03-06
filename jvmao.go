package jvmao

import (
	"bytes"
	ctx "context"
	"crypto/tls"
	"fmt"
	"io/fs"
	"net"
	"net/http"
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

//New return a instance of Jvmao.
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

		renderer:        new(DefaultRenderer),
		NotFoundHandler: defaultNotFoundHandler,
		HTTPErrHandler:  defaultHttpErrorHandler,
	}
	jm.Logger = DefaultLogger()
	jm.pool = sync.Pool{New: func() interface{} {
		return &context{jm: jm, w: NewResponse(jm, nil)}
	}}
	jm.mux = newMux(jm)

	return jm
}

// Jumao top-level instance.
type Jvmao struct {
	mu   sync.Mutex
	pool sync.Pool

	mux *mux

	// listener       net.Listener
	// tlsListener    net.Listener

	grpc *GrpcHandler
	// srv  *grpc.Server

	hs             *http.Server
	tlsHs          *http.Server
	AutoTLSManager autocert.Manager

	tcpAlivePeriod time.Duration

	middleware      []MiddlewareFunc
	renderer        Renderer
	NotFoundHandler HandlerFunc
	HTTPErrHandler  HTTPErrorHandler
	Logger          Logger

	debug bool
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

func (jm *Jvmao) Static(root, prefix string) {
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	jm.mux.setStatic(root, prefix)
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

	if name == "" {
		name = fmt.Sprintf("%p%s", &h, pattern)
	}

	jm.mux.handle(name, method, pattern, h)
}

func (jm *Jvmao) Reverse(name string, params ...string) string {
	uri := new(bytes.Buffer)

	l := len(params)
	if p, ok := jm.mux.routes[name]; ok {
		c := strings.Count(p, ":")
		if c != l {
			panic(fmt.Sprintf("Reverse error: need %d params but get %d", c, l))
		}
		if l == 0 {
			return p
		}

		// /sf/: id /:sdf/d
		for i := 0; i < l; i++ {
			ps := strings.SplitN(p, ":", 2)
			uri.WriteString(ps[0])
			uri.WriteString(fmt.Sprintf("%v", params[i]))
			p = ps[1]
			slash := strings.Index(p, "/")
			if slash > 0 {
				p = p[slash:]
			}

		}
		if strings.HasPrefix(p, "/") {
			uri.WriteString(p)
		}
	}

	return uri.String()
}

// Debug show debug is open or not.
func (jm *Jvmao) Debug() bool {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	return jm.debug
}

//Opendebug open debug.
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

//StartAutoTLS start an HTTPS server using certificates automatically from https://letsencrypt.org.
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

	http2.ConfigureServer(jm.tlsHs, &http2.Server{
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
//calls `http.Server#Shutdown()`.
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
//
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
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(ln.alivePeriod)
	return tc, nil
}
