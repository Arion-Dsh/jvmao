package main

import (
	"context"
	"net/http"

	"github.com/arion-dsh/jvmao/middleware"
	"google.golang.org/grpc"

	"github.com/arion-dsh/jvmao"
	pb "github.com/arion-dsh/jvmao/examples/proto"
)

func tM(next jvmao.HandlerFunc) jvmao.HandlerFunc {
	return func(c jvmao.Context) error {
		return next(c)
	}
}

type echo struct {
	pb.UnimplementedEchoServer
}

func (e *echo) Hello(ctx context.Context, req *pb.HelloRequest) (resp *pb.HelloReply, err error) {
	resp = &pb.HelloReply{
		Message: req.GetName(),
	}
	return
}

func main() {
	// cert, err := tls.LoadX509KeyPair("./server.crt", "./server.key")
	// if err != nil {
	// panic(err)
	// }
	opts := []grpc.ServerOption{
		// Enable TLS for all incoming connections.
		// grpc.Creds(credentials.NewServerTLSFromCert(&cert)),
	}
	serv := grpc.NewServer(opts...)
	pb.RegisterEchoServer(serv, new(echo))

	j := jvmao.New()
	j.RegisterGrpcServer(serv)

	h := func(c jvmao.Context) error {
		return c.String(http.StatusOK, "123")
	}

	j.Use(middleware.Logger())
	j.Use(middleware.Recover())
	j.Use(tM)

	j.GET("home", "", h)

	g := j.Group("/group")
	g.GET("g-home", "", h)

	j.Static("static/", "/static/")

	j.StartTLS(":8000", "./server.crt", "./server.key")

}
