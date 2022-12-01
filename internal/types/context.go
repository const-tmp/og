package types

import (
	"fmt"
)

type (
	TypeData struct {
		Type      Type
		Struct    *Struct
		Interface *Interface
	}

	TypeMap   map[string]TypeData
	ModuleMap map[string]*Module
)

func typeIndexString(t Type) string {
	return fmt.Sprintf("%s/%s", t.ImportPath(), t.Name())
}

func (m TypeMap) Contains(s string) bool {
	_, ok := m[s]
	return ok
}

func (m TypeMap) ContainsType(t Type) bool {
	_, ok := m[typeIndexString(t)]
	return ok
}

func (m TypeMap) Get(t Type) TypeData {
	return m[typeIndexString(t)]
}

func (m TypeMap) Add(t Type) {
	if !m.ContainsType(t) {
		m[typeIndexString(t)] = TypeData{Type: t}
	}
}

func (m TypeMap) Set(t Type, data TypeData) {
	m[typeIndexString(t)] = data
}

func (m ModuleMap) Add(f *GoFile) error {
	if mod, ok := m[f.Module]; ok {
		if p, ok := mod.Packages[f.Package]; ok {
			if _, ok := p.Files[f.FilePath]; ok {
				return fmt.Errorf("file %s already parsed", f.FilePath)
			} else {
				p.Files[f.FilePath] = f
			}
		} else {
			mod.Packages[f.Package] = NewPackageFromGoFile(f)
		}
	} else {
		m[f.Module] = &Module{
			Name:     f.Module,
			Path:     f.ModulePath,
			Packages: map[string]*Package{f.Package: NewPackageFromGoFile(f)},
		}
	}
	return nil
}

func NewPackageFromGoFile(f *GoFile) *Package {
	return &Package{
		Name:       f.Package,
		ImportPath: f.ImportPath(),
		Path:       f.FilePath,
		Files:      map[string]*GoFile{f.FilePath: f},
		Structs:    make(map[string]*Struct),
		Interfaces: make(map[string]*Interface),
	}
}
