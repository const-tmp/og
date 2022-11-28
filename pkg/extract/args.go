package extract

import (
	"fmt"
	"github.com/nullc4t/og/internal/types"
	"go/ast"
	"log"
)

func ArgsFromFields(file *ast.File, fields *ast.FieldList) types.Args {
	var args []*types.Arg

	for _, arg := range fields.List {
		var (
			names []string
			t     types.Type
		)

		for _, name := range arg.Names {
			names = append(names, name.Name)
		}

		t = TypeFromExpr(file, arg.Type)

		if len(arg.Names) == 0 {
			args = append(args, &types.Arg{Type: t})
		} else {
			for _, name := range names {
				args = append(args, &types.Arg{Name: name, Type: t})
			}
		}
	}

	return args
}

func TypeFromExpr(file *ast.File, field ast.Expr) types.Type {
	var t types.Type

	switch v := field.(type) {
	case *ast.Ident:
		t = TypeFromIdent(v)
	case *ast.SelectorExpr:
		t = TypeFromSelectorExpr(file, v)
	case *ast.ArrayType:
		t = TypeFromArrayType(file, v)
	case *ast.StarExpr:
		t = TypeFromStarExpr(file, v)
	case *ast.Ellipsis:
		t = TypeFromEllipsis(file, v)
	case *ast.MapType:
		t = TypeFromMapType(file, v)
	case *ast.IndexExpr:
		fmt.Println("ast.IndexExpr type is not implemented")
	default:
		log.Fatalf("[ BUG ] unknown ast.Expr: %T file: %s", v, file.Name.Name)
	}

	return t
}

func TypeFromIdent(id *ast.Ident) types.Type {
	return types.NewType(id.Name, "", "")
}

func TypeFromSelectorExpr(file *ast.File, se *ast.SelectorExpr) types.Type {
	var p string

	switch pIdent := se.X.(type) {
	case *ast.Ident:
		p = pIdent.Name
	default:
		log.Fatal("[ BUG ] unknown ast.SelectorExpr.X", pIdent)
	}

	return types.NewType(se.Sel.Name, p, ImportString(file, p))
}

func TypeFromStarExpr(file *ast.File, se *ast.StarExpr) types.Type {
	var t types.Type

	switch x := se.X.(type) {
	case *ast.Ident:
		t = TypeFromIdent(x)
	case *ast.SelectorExpr:
		t = TypeFromSelectorExpr(file, x)
	case *ast.ArrayType:
		t = TypeFromArrayType(file, x)
	default:
		log.Fatalf("[ TODO ] unknown ast.StarExpr.X: %T", x)
	}

	return types.Pointer{Type: t}
}

func TypeFromEllipsis(file *ast.File, el *ast.Ellipsis) types.Type {
	var t types.Type

	switch x := el.Elt.(type) {
	case *ast.Ident:
		t = TypeFromIdent(x)
	case *ast.SelectorExpr:
		t = TypeFromSelectorExpr(file, x)
	case *ast.ArrayType:
		t = TypeFromArrayType(file, x)
	default:
		log.Fatalf("[ TODO ] unknown ast.Ellipsis.Elt: %T", x)
	}

	return types.Pointer{Type: t}
}

func TypeFromArrayType(file *ast.File, at *ast.ArrayType) types.Type {
	var t types.Type

	switch elt := at.Elt.(type) {
	case *ast.Ident:
		t = TypeFromIdent(elt)
	case *ast.SelectorExpr:
		t = TypeFromSelectorExpr(file, elt)
	case *ast.StarExpr:
		t = TypeFromStarExpr(file, elt)
	default:
		log.Fatalf("[ TODO ] unknown ast.ArrayType.Elt: %T", elt)
	}

	return types.Slice{Type: t}
}

func TypeFromMapType(file *ast.File, mt *ast.MapType) types.Type {
	kType := TypeFromExpr(file, mt.Key)
	vType := TypeFromExpr(file, mt.Value)
	return types.NewMapType(kType, vType)
}
