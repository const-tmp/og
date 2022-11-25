package test

import "github.com/nullc4t/og/pkg/extract"

type Simple interface {
	Get(i int) (err error)
	Get2(int) error
	Get3(i int) error
	Get4(int) (err error)
	Get5([]int) (err error)
	Method(method extract.Method) error
	MethodP(method *extract.Method) error
	MethodAP(method []*extract.Method) error
}

type s struct {
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

func (s s) Method(method extract.Method) error {
	//TODO implement me
	panic("implement me")
}

func (s s) MethodP(method *extract.Method) error {
	//TODO implement me
	panic("implement me")
}

func (s s) MethodAP(method []*extract.Method) error {
	//TODO implement me
	panic("implement me")
}

func NewSimple() Simple {
	return s{}
}
