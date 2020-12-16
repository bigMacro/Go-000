package biz

import "fmt"

// Greeter
type Greeter struct {
	repo HelloNameRepo
}

func NewGreeter(repo HelloNameRepo) *Greeter {
	return &Greeter{repo: repo}
}

func (g *Greeter) GenerateHelloMessage(msg *HelloName) string {
	g.repo.IncreaseCounter(msg.Name)
	return fmt.Sprintf("hello %v", msg.Name)
}

// do
type HelloName struct {
	Name string
}

func NewHelloName(name string) *HelloName {
	return &HelloName{Name: name}
}

// repo
type HelloNameRepo interface {
	IncreaseCounter(string)
}
