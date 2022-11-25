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
		t = NewTypeFromArrayType(file, v)
	case *ast.StarExpr:
		t = NewTypeFromStarExpr(file, v)
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

	return Type{Name: se.Sel.Name, Package: p, ImportPath: ImportByPackage(file, p)}
}

func NewTypeFromStarExpr(file *ast.File, se *ast.StarExpr) Type {
	var t Type
	switch x := se.X.(type) {
	case *ast.Ident:
		t = NewTypeFromIdent(x)
	case *ast.SelectorExpr:
		t = NewTypeFromASTSelectorExpr(file, x)
	case *ast.ArrayType:
		t = NewTypeFromArrayType(file, x)
	default:
		log.Fatalf("[ BUG ] unknown ast.StarExpr.X: %T", x)
	}
	t.IsPointer = true
	return t
}

func NewTypeFromArrayType(file *ast.File, at *ast.ArrayType) Type {
	var t Type
	switch elt := at.Elt.(type) {
	case *ast.Ident:
		t = NewTypeFromIdent(elt)
	case *ast.SelectorExpr:
		t = NewTypeFromASTSelectorExpr(file, elt)
	case *ast.StarExpr:
		t = NewTypeFromStarExpr(file, elt)
	default:
		log.Fatalf("[ BUG ] unknown ast.ArrayType.Elt: %T", elt)
	}
	t.IsArray = true
	return t
}
