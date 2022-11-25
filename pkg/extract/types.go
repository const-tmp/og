package extract

import (
	"fmt"
	"strings"
)

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

	Arg struct {
		Name string
		Type Type
	}
	Args    []*Arg
	Results struct{ Args }

	Type struct {
		Name       string
		Package    string
		ImportPath string
		IsArray    bool
		IsPointer  bool
	}

	Import struct {
		Name string
		Path string
	}
)

func (m Method) String() string {
	return fmt.Sprintf("%s%s %s", m.Name, m.Args.String(), m.Results.String())
}

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

func (t Type) String() string {
	var prefix string
	if t.IsArray {
		prefix = "[]"
	}
	if t.IsPointer {
		prefix += "*"
	}
	if t.Package == "" {
		return prefix + t.Name
	}
	return fmt.Sprintf("%s.%s", prefix+t.Package, t.Name)
}

func (t Type) IsImported() bool {
	return t.Package != ""
}

func (i Import) IsAliasedImportRequired() bool {
	path := strings.Split(i.Path, "/")
	return path[len(path)-1] != i.Name
}

func (t Type) GetName() string {
	return t.Name
}
