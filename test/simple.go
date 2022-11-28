package test

import (
	"github.com/go-kit/kit/endpoint"
	"github.com/nullc4t/og/internal/types"
)

type Simple interface {
	Get(i int) (err error)
	Get2(int) error
	Get3(i int) error
	Get4(int) (err error)
	Get5([]int) (err error)
	Method(method types.Method) error
	MethodP(method *types.Method) error
	MethodAP(method []*types.Method) error
	EP(end endpoint.Endpoint) error
}

type s struct {
	A string
	B int
}

func (s s) EP(end endpoint.Endpoint) error {
	//TODO implement me
	panic("implement me")
}

func (s s) Get(i int) (err error) {
	//TODO implement me
	panic("implement me")
}

func (s s) Get2(i int) error {
	//TODO implement me
	panic("implement me")
}

func (s s) Get3(i int) error {
	//TODO implement me
	panic("implement me")
}

func (s s) Get4(i int) (err error) {
	//TODO implement me
	panic("implement me")
}

func (s s) Get5(ints []int) (err error) {
	//TODO implement me
	panic("implement me")
}

func (s s) Method(method types.Method) error {
	//TODO implement me
	panic("implement me")
}

func (s s) MethodP(method *types.Method) error {
	//TODO implement me
	panic("implement me")
}

func (s s) MethodAP(method []*types.Method) error {
	//TODO implement me
	panic("implement me")
}

func NewSimple() Simple {
	return s{}
}
