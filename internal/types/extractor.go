package types

type (
	Extractor interface {
		Types() ([]*Interface, []*Struct)
		Interfaces() []*Interface
		Structs() []*Struct
	}
)
