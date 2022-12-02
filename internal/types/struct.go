package types

type (
	// Struct exported type TODO: edit
	Struct struct {
		Name        string
		Fields      []Field
		UsedImports []Import
	}

	// Field exported type TODO: edit
	Field struct {
		Name string
		Type Type
		Tag  string
	}
)
