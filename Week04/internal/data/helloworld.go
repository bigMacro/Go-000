package data

import (
	"Go-000/Week04/internal/biz"
)

type HelloNameRepo struct {
}

func NewHelloNameRepo() biz.HelloNameRepo {
	return &HelloNameRepo{}
}

func (r *HelloNameRepo) IncreaseCounter(name string) {
	// increase counter for name
}
