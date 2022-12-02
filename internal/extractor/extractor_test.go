package extractor

import (
	"github.com/nullc4t/og/internal/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExtractor(t *testing.T) {
	ex := NewExtractor()
	_, _, err := ex.ParseFile("extractor_test.go", "", 0)
	require.NoError(t, err)
	for _, module := range ex.ModuleMap {
		t.Log(module.Name, module.Path)
		for _, p := range module.Packages {
			t.Log("\t", p.Name, p.ImportPath, p.Path)
			for s, _ := range p.Files {
				t.Log("\t\t", s)
			}
			for _, s := range p.Structs {
				t.Log("\t\t", s.Name, s)
			}
		}
	}
}

type (
	SimpleStruct struct {
		A int
		B string
	}
	ImportedStruct struct {
		A types.Struct
		B types.Interface
	}
)
