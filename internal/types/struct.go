package types

type (
	Struct struct {
		Name        string
		Fields      []Field
		UsedImports []Import
	}

	Field struct {
		Name string
		Type Type
		Tag  string
	}
)
