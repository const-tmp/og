package types

import (
	"strings"
)

type (
	Import struct {
		Name string
		Path string
	}

	Dependency struct {
		Module  string
		Version string
		Path    string
	}
)

func (i Import) IsAliasedImportRequired() bool {
	path := strings.Split(i.Path, "/")
	return path[len(path)-1] != i.Name
}
