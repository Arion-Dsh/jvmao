package jvmao

import (
	"net/http"
	"strings"
	"sync"

	"google.golang.org/grpc"
)

// NewGrpcHandler instance of GrpcHandler
func NewGrpcHandler() *GrpcHandler {
	return &GrpcHandler{
		pool: sync.Pool{
			New: func() interface{} {
				return &grpcResponse{}
			},
		},
	}
}

// GrpcHandler carries *grpc.Server.
// it is can handle grpc-web request.
type GrpcHandler struct {
	mux sync.Mutex

	srv *grpc.Server

	pool sync.Pool
}

// RegisterGrpcServer registerGrpcServer to handle.
func (g *GrpcHandler) RegisterGrpcServer(s *grpc.Server) {
	g.mux.Lock()
	g.srv = s
	g.mux.Unlock()
}

//IsGrpc test *http.Request content type has prefix application/grpc or not.
func IsGrpc(r *http.Request) bool {
	return r.Method == http.MethodPost && strings.HasPrefix(r.Header.Get(HeaderContentType), MIMEApplicationGrpc)
}

func (g *GrpcHandler) isGrpcWeb(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get(HeaderContentType), MIMEApplicationGrpcWeb)
}

// ServeHTTP implements http.Handler
// it will fake out grpc-web that carries standard gRPC.
func (g *GrpcHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if g.srv == nil {
		return
	}
	if g.isGrpcWeb(r) {
		g.handleGrpcWeb(w, r)
		return
	}
	g.srv.ServeHTTP(w, r)
}

func (g *GrpcHandler) handleGrpcWeb(w http.ResponseWriter, r *http.Request) {

	rw := g.pool.Get().(*grpcResponse)
	rw.reset(w, r)

	g.srv.ServeHTTP(rw, r)
	rw.finish()
	g.pool.Put(rw)
}
