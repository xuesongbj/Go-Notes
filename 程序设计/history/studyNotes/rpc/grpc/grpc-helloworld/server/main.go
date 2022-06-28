package main

import (
	"log"
	"net"

	pb "grpc/grpc-helloworld/helloworld"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":8888"
)

type server struct{}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func main() {
	// Listen 监听端口
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("faield to listen: %v", err)
	}

	// 创建一个gRPC服务器.
	// 1. 未注册服务。
	// 2. 尚未开始接受请求。
	s := grpc.NewServer()

	// RegisterGreeterServer将server注册到gRPC服务器。
	// 必须在调用Serve之前进行注册。
	pb.RegisterGreeterServer(s, &server{})

	// 注册反射服务.
	// 实现服务器的反射服务.
	// example: https://github.com/grpc/grpc/blob/master/src/proto/grpc/reflection/v1alpha/reflection.proto
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("faild to serve: %v", err)
	}
}
