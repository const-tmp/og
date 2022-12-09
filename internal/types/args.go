package types

import (
	"fmt"
	"github.com/nullc4t/og/pkg/names"
	"github.com/nullc4t/og/pkg/utils"
	"strings"
)

type (
	Arg struct {
		Name         string
		Type         Type
		Dependencies []Import
	}

	Args    []*Arg
	Results struct{ Args }
)

func (a Args) String() string {
	var tmp []string
	for _, arg := range a {
		tmp = append(tmp, arg.String())
	}
	return fmt.Sprintf("(%s)", strings.Join(tmp, ", "))
}

var stringSliceUtil = utils.NewSlice[string](func(a, b string) bool {
	return a == b
})

func (a Args) UnexportedNames(exclude ...string) string {
	var s []string
	for _, arg := range a {
		if stringSliceUtil.Contains(exclude, arg.Name) || stringSliceUtil.Contains(exclude, names.Unexported(arg.Name)) {
			continue
		}
		s = append(s, names.Unexported(arg.Name))
	}
	return strings.Join(s, ", ")
}

func (a Args) HasError() bool {
	for _, arg := range a {
		if ArgIsError(*arg) {
			return true
		}
	}
	return false
}

func (a Args) HasContext() bool {
	for _, arg := range a {
		if ArgIsContext(*arg) {
			return true
		}
	}
	return false
}

func (r Results) String() string {
	if len(r.Args) == 0 {
		return ""
	}
	if len(r.Args) == 1 && r.Args[0].Name == "" {
		return r.Args[0].Type.String()
	}
	return r.Args.String()
}

func (a Arg) String() string {
	if a.Name == "" {
		return a.Type.String()
	}
	return fmt.Sprintf("%s %s", a.Name, a.Type)
}

func ArgIsContext(arg Arg) bool {
	return arg.Type.String() == "context.Context"
}

func ArgIsError(arg Arg) bool {
	return arg.Type.String() == "error"
}
