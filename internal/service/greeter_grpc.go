package service

import (
	"context"

	pb "demo/api/helloworld/greeter/v1"
)

var (
	_ pb.GreeterServer = (*GreeterService)(nil)
)

type GreeterService struct {
	pb.UnimplementedGreeterServer
}

func NewGreeterService() *GreeterService {
	return &GreeterService{}
}

func (s *GreeterService) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + req.Name}, nil
}
