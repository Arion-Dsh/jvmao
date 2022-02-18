package jvmao

import (
	"net/http"
	"strings"
	"sync"

	"google.golang.org/grpc"
)

func newGrpcServer() *grpcServer {
	return &grpcServer{
		pool: sync.Pool{
			New: func() interface{} {
				return &grpcResponse{}
			},
		},
	}
}

type grpcServer struct {
	srv *grpc.Server

	pool sync.Pool
}

func (g *grpcServer) register(s *grpc.Server) {
	g.srv = s
}

func (g *grpcServer) isGrpc(r *http.Request) bool {
	return r.Method == http.MethodPost && strings.HasPrefix(r.Header.Get(HeaderContentType), MIMEApplicationGrpc)

}

func (g *grpcServer) isGrpcWeb(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get(HeaderContentType), MIMEApplicationGrpcWeb)
}

func (g *grpcServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if g.srv == nil {
		return
	}
	if g.isGrpcWeb(r) {
		// g.handleGrpcWeb(w, r)
		http.HandlerFunc(g.handleGrpcWeb).ServeHTTP(w, r)
		return
	}
	g.srv.ServeHTTP(w, r)
}

func (g *grpcServer) handleGrpcWeb(w http.ResponseWriter, r *http.Request) {

	rw := g.pool.Get().(*grpcResponse)
	rw.reset(w, r)

	g.srv.ServeHTTP(rw, r)
	rw.finish()
	g.pool.Put(rw)
}
