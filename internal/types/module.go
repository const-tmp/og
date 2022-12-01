package types

type (
	Module struct {
		Name     string
		Path     string
		Packages map[string]*Package
		Version  string
	}

	Package struct {
		Name       string
		ImportPath string
		Path       string
		Files      map[string]*GoFile
		Structs    []*Struct
		Interfaces []*Interface
	}
)
