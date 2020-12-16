//+build wireinject

package main

import (
	"Go-000/Week04/internal/biz"
	"Go-000/Week04/internal/data"
	"Go-000/Week04/internal/serivce"

	"github.com/google/wire"
)

func InitializeGreeterService() *serivce.GreeterService {
	wire.Build(serivce.NewGreeterService, biz.NewGreeter, data.NewHelloNameRepo)
	return nil
}
