package main

import (
	"context"
	"flag"
	"log"
	"time"

	pb "github.com/arion-dsh/jvmao/examples/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	defaultName = "world"
)

var (
	addr = flag.String("addr", "127.0.0.1:8000", "the address to connect to")
	name = flag.String("name", defaultName, "Name to echo")
)

func main() {
	flag.Parse()
	// Set up a connection to the server.
	creds, _ := credentials.NewClientTLSFromFile("../server.crt", "127.0.0.1")
	// conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewEchoClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.Hello(ctx, &pb.HelloRequest{Name: *name})
	if err != nil {
		log.Fatalf("could not echo: %v", err)
	}
	log.Printf("Echo: %s", r.GetMessage())
}
