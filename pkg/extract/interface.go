package extract

import (
	"fmt"
	"go/ast"
	"go/token"
)

type Interface struct {
	Name    string
	Methods []Method
}

type Method struct {
	Name    string
	Args    []Arg
	Results []Arg
}

type Arg struct {
	Name string
	Type string
}

func Interfaces(file *ast.File) []Interface {
	var ifaces []Interface

	for _, decl := range file.Decls {
		fmt.Println(decl)
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			iface, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}

			i := Interface{Name: typeSpec.Name.Name}

			for _, field := range iface.Methods.List {
				funcType, ok := field.Type.(*ast.FuncType)
				if !ok {
					continue
				}

				m := Method{Name: field.Names[0].Name}

				for _, arg := range funcType.Params.List {
					argType, ok := arg.Type.(*ast.Ident)
					if !ok {
						fmt.Println("interface:", i.Name, "method:", m.Name, arg, "is not type *ast.Ident")
						continue
					}
					a := Arg{Type: argType.Name, Name: arg.Names[0].Name}
					m.Args = append(m.Args, a)
				}

				for _, arg := range funcType.Results.List {
					argType, ok := arg.Type.(*ast.Ident)
					if !ok {
						fmt.Println("interface:", i.Name, "method:", m.Name, arg, "is not type *ast.Ident")
						continue
					}
					a := Arg{Type: argType.Name, Name: arg.Names[0].Name}
					m.Results = append(m.Results, a)
				}
				i.Methods = append(i.Methods, m)
			}
			ifaces = append(ifaces, i)
		}
	}
	return ifaces
}
