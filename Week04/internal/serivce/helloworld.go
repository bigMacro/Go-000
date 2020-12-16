package serivce

import (
	v1 "Go-000/Week04/api/helloworld/v1"
	"Go-000/Week04/internal/biz"
	"context"
)

type GreeterService struct {
	greeter *biz.Greeter
	*v1.UnimplementedGreeterServer
}

func NewGreeterService(greeter *biz.Greeter) *GreeterService {
	return &GreeterService{
		greeter: greeter,
	}
}

func (s *GreeterService) SayHello(ctx context.Context, r *v1.HelloRequest) (*v1.HelloReply, error) {
	// deep copy: dto --> do
	o := biz.NewHelloName(r.Name)
	o.Name = r.GetName()

	// call biz function to complete service
	msg := s.greeter.GenerateHelloMessage(o)
	return &v1.HelloReply{Message: msg}, nil
}
