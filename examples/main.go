package main

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
	// resp = &pb.HelloReply{
	// Message:    req.GetName(),
	// MessageOne: "msgOne",
	// }
	err = status.Error(codes.PermissionDenied, "error")
	return
}

func (e *echo) RepeatHello(req *pb.RepeatHelloRequest, stream pb.Echo_RepeatHelloServer) error {

	for i := 0; i < int(req.GetCount()); i++ {
		stream.Send(&pb.HelloReply{
			Message:    fmt.Sprintf("message: name: %s, count: %d \n", req.GetName(), i),
			MessageOne: "msgOne",
		})
	}

	return nil
}

func (e *echo) StreamHello(stream pb.Echo_StreamHelloServer) error {

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.Send(&pb.HelloReply{Message: req.GetName()})
		}

		if err != nil {
			return err
		}
	}
}

func main() {
	opts := []grpc.ServerOption{}
	serv := grpc.NewServer(opts...)
	pb.RegisterEchoServer(serv, new(echo))

	j := jvmao.New()
	j.RegisterGrpcServer(serv)

	h := func(c jvmao.Context) error {
		return c.String(http.StatusOK, "123")
	}
	// j.Use(middleware.Logger())
	// j.Use(middleware.Recover())
	j.Use(tM)

	j.GET("home", "", h)

	g := j.Group("/group")
	g.GET("g-home", "", h)

	j.Static("static/", "/static/")

	// j.Start(":8000")
	j.StartTLS(":8000", "./server.crt", "./server.key")

}
