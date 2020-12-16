// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package main

import (
	"Go-000/Week04/internal/biz"
	"Go-000/Week04/internal/data"
	"Go-000/Week04/internal/serivce"
)

// Injectors from wire.go:

func InitializeGreeterService() *serivce.GreeterService {
	helloNameRepo := data.NewHelloNameRepo()
	greeter := biz.NewGreeter(helloNameRepo)
	greeterService := serivce.NewGreeterService(greeter)
	return greeterService
}
