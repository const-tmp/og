package test

import "github.com/nullc4t/og/pkg/extract"

type Simple interface {
	Get(i int) (err error)
	Get2(int) error
	Get3(i int) error
	Get4(int) (err error)
	Get5([]int) (err error)
	Method(method extract.Method) error
}
