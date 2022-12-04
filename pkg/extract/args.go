package extract

import (
	"fmt"
	"github.com/nullc4t/og/internal/types"
	"go/ast"
	"log"
)

// ArgsFromFields extract types.Args from *ast.FieldList
func ArgsFromFields(file *types.GoFile, fields *ast.FieldList) types.Args {
	if fields == nil || fields.List == nil {
		return nil
	}

	var args []*types.Arg

	for _, arg := range fields.List {
		var t types.Type

		t = TypeFromExpr(file, arg.Type)
		if t == nil {
			continue
		}

		if len(arg.Names) == 0 {
			args = append(args, &types.Arg{Type: t})
		} else {
			for _, name := range arg.Names {
				args = append(args, &types.Arg{Name: name.Name, Type: t})
			}
		}
	}

	return args
}

func TypeFromExpr(file *types.GoFile, field ast.Expr) types.Type {
	var t types.Type

	switch v := field.(type) {
	case *ast.Ident:
		t = TypeFromIdent(file, v)
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
	case *ast.InterfaceType:
		t = types.NewType("interface{}", "", "")
		t.SetIsInterface()
	case *ast.FuncType:
		fmt.Println("ast.FuncType cannot be used in transport")
		return nil
	case *ast.ChanType:
		fmt.Println("ast.ChanType cannot be used in transport")
		return nil

	default:
		log.Fatalf("[ BUG ] unknown ast.Expr: %T file: %s", v, file.FilePath)
	}

	return t
}

func TypeFromIdent(file *types.GoFile, id *ast.Ident) types.Type {
	if types.IsBuiltIn(id.Name) {
		return types.NewType(id.Name, "", "")
	}
	//return types.NewType(id.Name, file.Name.Name, ImportStringForPackage(file, file.Name.Name))
	return types.NewType(id.Name, "", file.ImportPath())
}

func TypeFromSelectorExpr(file *types.GoFile, se *ast.SelectorExpr) types.Type {
	var p string

	switch pIdent := se.X.(type) {
	case *ast.Ident:
		p = pIdent.Name
	default:
		log.Fatal("[ BUG ] unknown ast.SelectorExpr.X", pIdent)
	}

	return types.NewType(se.Sel.Name, p, ImportStringForPackage(file, p))
}

func TypeFromStarExpr(file *types.GoFile, se *ast.StarExpr) types.Type {
	var t types.Type

	switch x := se.X.(type) {
	case *ast.Ident:
		t = TypeFromIdent(file, x)
	case *ast.SelectorExpr:
		t = TypeFromSelectorExpr(file, x)
	case *ast.ArrayType:
		t = TypeFromArrayType(file, x)
	default:
		log.Fatalf("[ TODO ] unknown ast.StarExpr.X: %T", x)
	}

	return types.Pointer{Type: t}
}

func TypeFromEllipsis(file *types.GoFile, el *ast.Ellipsis) types.Type {
	var t types.Type

	switch x := el.Elt.(type) {
	case *ast.Ident:
		t = TypeFromIdent(file, x)
	case *ast.SelectorExpr:
		t = TypeFromSelectorExpr(file, x)
	case *ast.ArrayType:
		t = TypeFromArrayType(file, x)
	case *ast.InterfaceType:
		t = types.NewType("interface{}", "", "")
		t.SetIsInterface()
	default:
		log.Fatalf("[ TODO ] unknown ast.Ellipsis.Elt: %T", x)
	}

	return types.Pointer{Type: t}
}

func TypeFromArrayType(file *types.GoFile, at *ast.ArrayType) types.Type {
	var t types.Type

	switch elt := at.Elt.(type) {
	case *ast.Ident:
		t = TypeFromIdent(file, elt)
	case *ast.SelectorExpr:
		t = TypeFromSelectorExpr(file, elt)
	case *ast.StarExpr:
		t = TypeFromStarExpr(file, elt)
	case *ast.InterfaceType:
		t = types.NewType("interface{}", "", "")
		t.SetIsInterface()
	case *ast.FuncType:
		fmt.Println("ast.FuncType cannot be used in transport")
		return nil
	case *ast.ArrayType:
		return TypeFromArrayType(file, elt)
	default:
		log.Fatalf("[ TODO ] unknown ast.ArrayType.Elt: %T", elt)
	}

	return types.Slice{Type: t}
}

func TypeFromMapType(file *types.GoFile, mt *ast.MapType) types.Type {
	kType := TypeFromExpr(file, mt.Key)
	vType := TypeFromExpr(file, mt.Value)
	if kType == nil || vType == nil {
		return nil
	}
	return types.NewMapType(kType, vType)
}
