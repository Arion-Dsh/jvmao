package jvmao

import (
	"net/http"
	"strings"

	"google.golang.org/grpc"
)

type grpcServer struct {
	srv *grpc.Server
}

func (rpc *grpcServer) isGrpc(r *http.Request) bool {
	return r.Method == http.MethodPost && strings.HasPrefix(r.Header.Get("content-type"), MIMEApplicationGrpc)

}

func (rpc *grpcServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rpc.srv.ServeHTTP(w, r)
}
