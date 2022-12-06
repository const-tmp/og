package types

type (
	// Struct exported type TODO: edit
	Struct struct {
		Name        string
		Fields      []Field
		UsedImports []Import
		ImportPath  string
		Package     string
	}

	// Field exported type TODO: edit
	Field struct {
		Name string
		Type Type
		Tag  string
	}
)
