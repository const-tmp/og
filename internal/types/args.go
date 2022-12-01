package types

import (
	"fmt"
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
