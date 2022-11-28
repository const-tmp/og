package types

import "fmt"

type (
	Interface struct {
		Name        string
		Methods     []Method
		UsedImports []Import
	}

	Method struct {
		Name    string
		Args    Args
		Results Results
	}
)

func (m Method) String() string {
	return fmt.Sprintf("%s%s %s", m.Name, m.Args.String(), m.Results.String())
}
