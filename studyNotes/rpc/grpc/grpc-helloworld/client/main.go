package main

import (
	"context"
	"log"
	"os"

	pb "grpc/grpc-helloworld/helloworld"

	"google.golang.org/grpc"
)

const (
	address     = "127.0.0.1:8888"
	defaultName = "world"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	name := defaultName
	c := pb.NewGreeterClient(conn)
	if len(os.Args) > 1 {
		name = os.Args[1]
	}

	r, err := c.SayHello(context.Background(), &pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.Message)
}
