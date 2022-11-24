package extract

import (
	"fmt"
	"go/ast"
	"log"
)

func GetArgs(file *ast.File, fields *ast.FieldList) Args {
	var args []*Arg

	for i, arg := range fields.List {
		var name string
		switch len(arg.Names) {
		case 0:
			//fmt.Println("[ THIS IS A BUG ]\targ", i, "has", len(arg.Names), "names")
		case 1:
			name = arg.Names[0].Name
		default:
			fmt.Println("[ THIS IS A BUG ]\targ", i, "has", len(arg.Names), "names")
		}
		a := Arg{Name: name}

		var t Type

		switch v := arg.Type.(type) {
		case *ast.Ident:
			t = Type{Name: v.Name}

		case *ast.SelectorExpr:
			var p string

			switch pIdent := v.X.(type) {
			case *ast.Ident:
				p = pIdent.Name
			default:
				log.Fatal("arg", i, v.Sel.Name, "type; package is not *ast.Ident")
			}

			t = Type{Name: v.Sel.Name, Package: p, ImportPath: ImportByPackage(file, p)}

		default:
			log.Fatal("arg", i, arg.Names[0], "type; package is not *ast.Ident")
		}

		a.Type = t
		args = append(args, &a)
	}

	return args
}
