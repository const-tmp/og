package types

import "fmt"

type (
	Interface struct {
		Name         string
		Methods      []Method
		Dependencies []Import
	}

	Method struct {
		Name         string
		Args         Args
		Results      Results
		Dependencies []Import
	}
)

func (m Method) String() string {
	return fmt.Sprintf("%s%s %s", m.Name, m.Args.String(), m.Results.String())
}
