package extract

import (
	"go/ast"
	"log"
)

func GetArgs(file *ast.File, fields *ast.FieldList) Args {
	var args []*Arg

	for _, arg := range fields.List {
		var (
			names []string
			t     Type
		)

		for _, name := range arg.Names {
			names = append(names, name.Name)
		}

		t = GetTypeFromASTExpr(file, arg.Type)

		if len(arg.Names) == 0 {
			args = append(args, &Arg{Type: t})
		} else {
			for _, name := range names {
				args = append(args, &Arg{Name: name, Type: t})
			}
		}
	}

	return args
}

func GetTypeFromASTExpr(file *ast.File, field ast.Expr) Type {
	var t Type

	switch v := field.(type) {
	case *ast.Ident:
		t = NewTypeFromIdent(v)
	case *ast.SelectorExpr:
		t = NewTypeFromASTSelectorExpr(file, v)
	case *ast.ArrayType:
		switch vv := v.Elt.(type) {
		case *ast.Ident:
			t = NewTypeFromIdent(vv)
			t.IsArray = true
		case *ast.SelectorExpr:
			t = NewTypeFromASTSelectorExpr(file, vv)
			t.IsArray = true
		default:
			log.Fatal("[ BUG ] unknown ast.ArrayType.Elt", vv)
		}
	case *ast.StarExpr:
		switch vv := v.X.(type) {
		case *ast.Ident:
			t = NewTypeFromIdent(vv)
			t.IsPointer = true
		case *ast.SelectorExpr:
			t = NewTypeFromASTSelectorExpr(file, vv)
			t.IsPointer = true
		default:
			log.Fatalf("[ BUG ] unknown ast.StarExpr.X: %T", vv)
		}
	default:
		log.Fatalf("[ BUG ] unknown ast.Expr: %T", v)
	}

	return t
}

func NewTypeFromIdent(id *ast.Ident) Type {
	return Type{Name: id.Name}
}

func NewTypeFromASTSelectorExpr(file *ast.File, se *ast.SelectorExpr) Type {
	var p string
	switch pIdent := se.X.(type) {
	case *ast.Ident:
		p = pIdent.Name
	default:
		log.Fatal("[ BUG ] unknown ast.SelectorExpr.X", pIdent)
	}

	//return Type{Name: se.Sel.Name, Package: p}
	return Type{Name: se.Sel.Name, Package: p, ImportPath: ImportByPackage(file, p)}
}
