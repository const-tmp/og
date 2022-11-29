package extract

import (
	"fmt"
	"github.com/nullc4t/og/internal/types"
	"github.com/nullc4t/og/pkg/utils"
	"go/ast"
	"go/token"
	"log"
)

type (
	Extractable interface {
		types.Interface | types.Struct
	}

	TypeData struct {
		Type      types.Type
		Struct    types.Struct
		Interface types.Interface
	}

	TypeMapping map[string]TypeData

	Extractor struct {
		Mapping TypeMapping
		Fset    *token.FileSet
		File    *ast.File
	}
)

func (e Extractor) Types() ([]types.Interface, []types.Struct) {
	var (
		ifaces  []types.Interface
		structs []types.Struct
	)

	for _, decl := range e.File.Decls {
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

			if i := e.InterfaceFromTypeSpec(typeSpec); i != nil {
				ifaces = append(ifaces, *i)
			}

			if s := e.StructFromTypeSpec(typeSpec); s != nil {
				structs = append(structs, *s)
			}
		}
	}

	return ifaces, structs
}

func (e Extractor) InterfaceFromTypeSpec(typeSpec *ast.TypeSpec) *types.Interface {
	iface, ok := typeSpec.Type.(*ast.InterfaceType)
	if !ok {
		return nil
	}

	i := types.Interface{Name: typeSpec.Name.Name}

	importSet := utils.NewSet[types.Import]()

	for _, field := range iface.Methods.List {
		funcType, ok := field.Type.(*ast.FuncType)
		if !ok {
			return nil
		}

		i.Methods = append(i.Methods, types.Method{
			Name:    field.Names[0].Name,
			Args:    e.ArgsFromFields(funcType.Params),
			Results: types.Results{Args: e.ArgsFromFields(funcType.Results)},
		})
	}

	for _, method := range i.Methods {
		for _, arg := range method.Args {
			if arg.Type == nil {
				utils.BugPanic(fmt.Sprint(method.Name, arg.Name, "null Type"))
			}
			if arg.Type.IsImported() {
				importSet.Add(types.Import{Name: arg.Type.Package(), Path: arg.Type.ImportPath()})
			}
		}
		for _, arg := range method.Results.Args {
			if arg.Type.IsImported() {
				importSet.Add(types.Import{Name: arg.Type.Package(), Path: arg.Type.ImportPath()})
			}
		}
	}

	i.UsedImports = importSet.All()

	return &i
}

func (e Extractor) ArgsFromFields(fields *ast.FieldList) types.Args {
	var args []*types.Arg

	for _, arg := range fields.List {
		var (
			names []string
			t     types.Type
		)

		for _, name := range arg.Names {
			names = append(names, name.Name)
		}

		t = e.TypeFromExpr(arg.Type)
		if cached, ok := e.Mapping[t.Package()+t.Name()]; ok {
			if cached.Type != nil {
				t = cached.Type
			} else {
				cached.Type = t
			}
		} else {
			e.Mapping[t.Package()+t.Name()] = TypeData{Type: t}
		}

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

func (e Extractor) TypeFromExpr(field ast.Expr) types.Type {
	var t types.Type

	switch v := field.(type) {
	case *ast.Ident:
		t = e.TypeFromIdent(v)
	case *ast.SelectorExpr:
		t = e.TypeFromSelectorExpr(v)
	case *ast.ArrayType:
		t = e.TypeFromArrayType(v)
	case *ast.StarExpr:
		t = e.TypeFromStarExpr(v)
	case *ast.Ellipsis:
		t = e.TypeFromEllipsis(v)
	case *ast.MapType:
		t = e.TypeFromMapType(v)
	case *ast.IndexExpr:
		fmt.Println("ast.IndexExpr type is not implemented")
	default:
		log.Fatalf("[ BUG ] unknown ast.Expr: %T file: %s", v, e.File.Name.Name)
	}

	return t
}

func (e Extractor) TypeFromIdent(id *ast.Ident) types.Type {
	return types.NewType(id.Name, "", "")
}

func (e Extractor) TypeFromSelectorExpr(se *ast.SelectorExpr) types.Type {
	var p string

	switch pIdent := se.X.(type) {
	case *ast.Ident:
		p = pIdent.Name
	default:
		log.Fatal("[ BUG ] unknown ast.SelectorExpr.X", pIdent)
	}

	return types.NewType(se.Sel.Name, p, ImportString(e.File, p))
}

func (e Extractor) TypeFromStarExpr(se *ast.StarExpr) types.Type {
	var t types.Type

	switch x := se.X.(type) {
	case *ast.Ident:
		t = e.TypeFromIdent(x)
	case *ast.SelectorExpr:
		t = e.TypeFromSelectorExpr(x)
	case *ast.ArrayType:
		t = e.TypeFromArrayType(x)
	default:
		log.Fatalf("[ TODO ] unknown ast.StarExpr.X: %T", x)
	}

	return types.Pointer{Type: t}
}

func (e Extractor) TypeFromEllipsis(el *ast.Ellipsis) types.Type {
	var t types.Type

	switch x := el.Elt.(type) {
	case *ast.Ident:
		t = e.TypeFromIdent(x)
	case *ast.SelectorExpr:
		t = e.TypeFromSelectorExpr(x)
	case *ast.ArrayType:
		t = e.TypeFromArrayType(x)
	default:
		log.Fatalf("[ TODO ] unknown ast.Ellipsis.Elt: %T", x)
	}

	return types.Pointer{Type: t}
}

func (e Extractor) TypeFromArrayType(at *ast.ArrayType) types.Type {
	var t types.Type

	switch elt := at.Elt.(type) {
	case *ast.Ident:
		t = e.TypeFromIdent(elt)
	case *ast.SelectorExpr:
		t = e.TypeFromSelectorExpr(elt)
	case *ast.StarExpr:
		t = e.TypeFromStarExpr(elt)
	default:
		log.Fatalf("[ TODO ] unknown ast.ArrayType.Elt: %T", elt)
	}

	return types.Slice{Type: t}
}

func (e Extractor) TypeFromMapType(mt *ast.MapType) types.Type {
	kType := e.TypeFromExpr(mt.Key)
	vType := e.TypeFromExpr(mt.Value)
	return types.NewMapType(kType, vType)
}

func (e Extractor) StructFromTypeSpec(typeSpec *ast.TypeSpec) *types.Struct {
	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return nil
	}

	s := types.Struct{Name: typeSpec.Name.Name}

	importSet := utils.NewSet[types.Import]()

	for _, field := range structType.Fields.List {
		var tag string
		if field.Tag != nil {
			tag = field.Tag.Value
		}

		switch len(field.Names) {
		case 1:
			s.Fields = append(s.Fields, types.Field{
				Name: field.Names[0].Name,
				Type: e.TypeFromExpr(field.Type),
				Tag:  tag,
			})
		case 0:
			s.Fields = append(s.Fields, types.Field{
				Type: e.TypeFromExpr(field.Type),
				Tag:  tag,
			})
		default:
			panic(fmt.Sprintf("[ THIS IS A BUG ] unexpected len(field.Names) == %d", len(field.Names)))
		}
	}

	for _, field := range s.Fields {
		if field.Type != nil && field.Type.IsImported() {
			importSet.Add(types.Import{Name: field.Type.Package(), Path: field.Type.ImportPath()})
		}
	}

	s.UsedImports = importSet.All()

	return &s
}
