package types

import (
	"fmt"
	"github.com/nullc4t/og/pkg/utils"
)

type (
	Type interface {
		Name() string
		Package() string
		String() string
		ImportPath() string
		IsImported() bool
		IsInterface() bool
		IsBuiltin() bool
		SetIsInterface()
	}

	BaseType struct {
		name        string
		pkg         string
		importPath  string
		isInterface bool
	}

	MapType struct {
		key   Type
		value Type
	}
	GenericType struct {
		name  Type
		index Type
	}

	Pointer  struct{ Type }
	Slice    struct{ Type }
	Ellipsis struct{ Type }
)

func NewType(name, pkg, importPath string) Type {
	return &BaseType{
		name:       name,
		pkg:        pkg,
		importPath: importPath,
	}
}

func (t *BaseType) String() string {
	if t.Package() == "" {
		return t.Name()
	}
	return fmt.Sprintf("%s.%s", t.Package(), t.Name())
}

func (t *BaseType) Name() string {
	return t.name
}

func (t *BaseType) ImportPath() string {
	return t.importPath
}

func (t *BaseType) Package() string {
	return t.pkg
}

func (t *BaseType) IsImported() bool {
	return len(t.ImportPath()) > 0
}

func (t *BaseType) IsInterface() bool {
	return t.isInterface
}

func (t *BaseType) SetIsInterface() {
	t.isInterface = true
}

func (t *BaseType) IsBuiltin() bool {
	if t.Package() != "" {
		switch t.ImportPath() {
		case "fmt", "errors", "strings", "time", "net", "net/http", "context", "math", "math/big":
			return true
		default:
			return false
		}
	}
	return IsBuiltIn(t.Name())
}

func (p Pointer) String() string {
	return "*" + p.Type.String()
}

func (p Slice) String() string {
	return "[]" + p.Type.String()
}

func (e Ellipsis) String() string {
	return "..." + e.Type.String()
}

func NewMapType(k, v Type) Type {
	return &MapType{key: k, value: v}
}

func (t *MapType) String() string {
	return fmt.Sprintf("map[%s]%s", t.key.String(), t.value.String())
}

func (t *MapType) IsImported() bool {
	return t.key.IsImported() || t.value.IsImported()
}

func (t *MapType) ImportPath() string {
	if t.key.IsImported() && t.value.IsImported() {
		utils.BugPanic("MapType: both key and value are imported. " +
			"Sorry this is bug by design and should be reworked :(")
	}

	if t.key.IsImported() {
		return t.key.ImportPath()
	}

	if t.value.IsImported() {
		return t.value.ImportPath()
	}
	return ""
}

func (t *MapType) Name() string {
	return t.String()
}

func (t *MapType) Package() string {
	return ""
}

func (t *MapType) IsInterface() bool {
	return false
}

func (t *MapType) SetIsInterface() {
	panic("cannot set interface on map")
}

func (t *MapType) IsBuiltin() bool {
	return t.key.IsBuiltin() || t.value.IsBuiltin()
}

func NewGenericType(name, index Type) Type {
	return &GenericType{name: name, index: index}
}

func (t *GenericType) String() string {
	return fmt.Sprintf("%s[%s]", t.name.String(), t.index.String())
}

func (t *GenericType) IsImported() bool {
	return t.name.IsImported() || t.index.IsImported()
}

func (t *GenericType) ImportPath() string {
	if t.name.IsImported() && t.index.IsImported() {
		utils.BugPanic("GenericType: both name and index are imported. " +
			"Sorry this is bug by design and should be reworked :(")
	}

	if t.name.IsImported() {
		return t.name.ImportPath()
	}

	if t.index.IsImported() {
		return t.index.ImportPath()
	}
	return ""
}

func (t *GenericType) Name() string {
	return t.String()
}

func (t *GenericType) Package() string {
	return ""
}

func (t *GenericType) IsInterface() bool {
	return false
}

func (t *GenericType) SetIsInterface() {
	panic("cannot set interface on generic type")
}

func (t *GenericType) IsBuiltin() bool {
	return t.name.IsBuiltin() || t.index.IsBuiltin()
}

func IsBuiltIn(s string) bool {
	switch s {
	case "error":
		return true
	case "int":
		return true
	case "int8":
		return true
	case "int16":
		return true
	case "int32":
		return true
	case "int64":
		return true
	case "uint":
		return true
	case "uint8":
		return true
	case "uint16":
		return true
	case "uint32":
		return true
	case "uint64":
		return true
	case "string":
		return true
	case "float32":
		return true
	case "float64":
		return true
	case "interface{}":
		return true
	case "any":
		return true
	case "bool":
		return true
	default:
		return false
	}
}
