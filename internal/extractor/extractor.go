package extractor

import (
	"fmt"
	"github.com/nullc4t/og/internal/types"
	"github.com/nullc4t/og/pkg/extract"
	"github.com/nullc4t/og/pkg/transform"
	"github.com/nullc4t/og/pkg/utils"
	"github.com/spf13/viper"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"path/filepath"
)

type (
	Extractable interface {
		types.Interface | types.Struct
	}

	Extractor struct {
		ModuleMap types.ModuleMap
		fset      *token.FileSet
		TypeMap   types.TypeMap
	}
)

func NewExtractor() *Extractor {
	return &Extractor{ModuleMap: make(types.ModuleMap), TypeMap: make(types.TypeMap), fset: token.NewFileSet()}
}

func (e Extractor) ParseFile(path string, query string, depth int) ([]*types.Interface, []*types.Struct, error) {
	fmt.Println("parsing file ", path)
	f, err := GoFile(path)
	if err != nil {
		return nil, nil, err
	}

	_ = e.ModuleMap.Add(f)

	ifaces, structs, err := e.TypeDefs(f, query, depth)
	if err != nil {
		fmt.Println(path, " parse file error:", err)
		return nil, nil, err
	}

	if query != "" {
		e.findAndAddTypeByName(f, ifaces, structs, query)
	} else {
		e.addTypeDefsToMapping(f, ifaces, structs)
	}

	fmt.Println(path, "parsed")
	return ifaces, structs, nil
}

func (e Extractor) addTypeDefsToMapping(f *types.GoFile, ifaces []*types.Interface, structs []*types.Struct) {
	for _, iface := range ifaces {
		e.ModuleMap[f.Module].Packages[f.Package].Interfaces[iface.Name] = iface
	}
	for _, str := range structs {
		e.ModuleMap[f.Module].Packages[f.Package].Structs[str.Name] = str
	}
}

func (e Extractor) findAndAddTypeByName(f *types.GoFile, ifaces []*types.Interface, structs []*types.Struct, name string) {
	for _, iface := range ifaces {
		if iface.Name == name {
			e.ModuleMap[f.Module].Packages[f.Package].Interfaces[iface.Name] = iface
			return
		}
	}
	for _, str := range structs {
		if str.Name == name {
			e.ModuleMap[f.Module].Packages[f.Package].Structs[str.Name] = str
			return
		}
	}
}

func (e Extractor) TypeDefs(file *types.GoFile, name string, depth int) ([]*types.Interface, []*types.Struct, error) {
	fmt.Println("getting type defs in ", file.FilePath)
	var (
		ifaces  []*types.Interface
		structs []*types.Struct
	)

	for _, decl := range file.AST.Decls {
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

			if i := e.InterfaceFromTypeSpec(file.AST, typeSpec); i != nil {
				if name != "" && name == i.Name {
					ifaces = append(ifaces, i)
					break
				}
				ifaces = append(ifaces, i)
			}

			if s := e.StructFromTypeSpec(file.AST, typeSpec); s != nil {
				if name != "" && name == s.Name {
					structs = append(structs, s)
					break
				}
				structs = append(structs, s)
			}
		}
	}
	i, s, err := e.findAndParseImportedTypes(file, ifaces, structs, depth)
	if err != nil {
		return nil, nil, err
	}

	ifaces = append(ifaces, i...)
	structs = append(structs, s...)

	fmt.Println("type defs ", file.FilePath, " done")
	return ifaces, structs, nil
}

func (e Extractor) findAndParseImportedTypes(file *types.GoFile, ifaces []*types.Interface, structs []*types.Struct, depth int) ([]*types.Interface, []*types.Struct, error) {
	if depth <= 0 {
		return nil, nil, nil
	}

	types2Find := make(types.TypeMap)
	for _, iface := range ifaces {
		for _, method := range iface.Methods {
			for _, arg := range method.Args {
				if !arg.Type.IsBuiltin() {
					types2Find.Add(arg.Type)
				}
			}
			for _, arg := range method.Results.Args {
				if !arg.Type.IsBuiltin() {
					types2Find.Add(arg.Type)
				}
			}
		}
	}
	for _, str := range structs {
		for _, field := range str.Fields {
			if !field.Type.IsBuiltin() {
				types2Find.Add(field.Type)
			}
		}
	}

	for _, data := range types2Find {
		for _, s := range viper.GetStringSlice("exclude_types") {
			if data.Type.String() == s {
				continue
			}
		}

		i, s, err := e.RecursiveParsePackage(file, data.Type.Package(), data.Type.Name(), depth)
		if err != nil {
			fmt.Printf("recursive find type %s error: %s\n", data.Type, err.Error())
		}
		ifaces = append(ifaces, i...)
		structs = append(structs, s...)
	}

	return ifaces, structs, nil
}

func (e Extractor) checkIsTypeParsed(ty types.Type) bool {
	t := e.TypeMap.Get(ty)
	if t.Interface != nil || t.Struct != nil {
		fmt.Println(ty, "has interface=", t.Interface != nil, "has struct=", t.Struct != nil)
		return true
	}
	return false
}

func (e Extractor) checkAndParseType(file *types.GoFile, t types.Type, name string, depth int) ([]*types.Interface, []*types.Struct, error) {
	if t.IsBuiltin() {
		return nil, nil, nil
	}
	if e.checkIsTypeParsed(t) {
		return nil, nil, nil
	}
	return e.RecursiveParsePackage(file, t.Package(), name, depth-1)
}

func (e Extractor) RecursiveParsePackage(file *types.GoFile, pkgName string, name string, depth int) ([]*types.Interface, []*types.Struct, error) {
	if depth == 0 {
		return nil, nil, nil
	}
	var (
		packagePath string
		err         error
	)
	if pkgName == "" {
		packagePath, err = extract.SourcePath4Package(file.Module, file.ModulePath, file.ImportPath(), file.FilePath)
	} else {
		packagePath, err = extract.SourcePath4Package(file.Module, file.ModulePath, extract.ImportStringForPackage(file.AST, pkgName), file.FilePath)
	}
	if err != nil {
		return nil, nil, err
	}

	goFiles, err := extract.GoSourceFilesFromPackage(packagePath)
	if err != nil {
		return nil, nil, err
	}

	fmt.Println(packagePath, "go files:", file.FilePath)

	for _, goFile := range goFiles {
		ifaces, structs, err := e.ParseFile(goFile, name, depth)
		if err != nil {
			return nil, nil, err
		}
		for _, iface := range ifaces {
			if iface.Name == name {
				return ifaces, structs, nil
			}
		}
		for _, str := range structs {
			if str.Name == name {
				return ifaces, structs, nil
			}
		}
	}

	return nil, nil, nil
}

func (e Extractor) InterfaceFromTypeSpec(file *ast.File, typeSpec *ast.TypeSpec) *types.Interface {
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
			Args:    e.ArgsFromFields(file, funcType.Params),
			Results: types.Results{Args: e.ArgsFromFields(file, funcType.Results)},
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

	i.Dependencies = importSet.All()

	return &i
}

func (e Extractor) ArgsFromFields(file *ast.File, fields *ast.FieldList) types.Args {
	if fields == nil || fields.List == nil {
		return nil
	}

	var args []*types.Arg

	for _, arg := range fields.List {
		var t types.Type

		t = e.TypeFromExpr(file, arg.Type)
		if t == nil {
			continue
		}

		e.TypeMap.Add(t)

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

func (e Extractor) StructFromTypeSpec(file *ast.File, typeSpec *ast.TypeSpec) *types.Struct {
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
		t := e.TypeFromExpr(file, field.Type)
		if t == nil {
			continue
		}

		if len(field.Names) == 0 {
			s.Fields = append(s.Fields, types.Field{
				Name: transform.RenameEmpty(t),
				Type: t,
				Tag:  tag,
			})
		} else {
			for _, name := range field.Names {
				s.Fields = append(s.Fields, types.Field{
					Name: name.Name,
					Type: t,
					Tag:  tag,
				})
			}
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

func GoFile(path string) (*types.GoFile, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("file abs path error: %w", err)
	}

	goMod, err := extract.SearchFileUp("go.mod", filepath.Dir(absPath), types.SearchUpDirLimit)
	if err != nil {
		return nil, err
	}

	absModulePath, err := filepath.Abs(filepath.Dir(goMod))
	if err != nil {
		return nil, fmt.Errorf("file abs path error: %w", err)
	}

	module, err := extract.ModuleNameFromGoMod(goMod)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, absPath, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	return &types.GoFile{
		FilePath:   absPath,
		Package:    file.Name.Name,
		Module:     module,
		ModulePath: absModulePath,
		FSet:       fset,
		AST:        file,
	}, nil
}

func (e Extractor) TypeFromExpr(file *ast.File, field ast.Expr) types.Type {
	var t types.Type

	switch v := field.(type) {
	case *ast.Ident:
		t = e.TypeFromIdent(file, v)
	case *ast.SelectorExpr:
		t = e.TypeFromSelectorExpr(file, v)
	case *ast.ArrayType:
		t = e.TypeFromArrayType(file, v)
	case *ast.StarExpr:
		t = e.TypeFromStarExpr(file, v)
	case *ast.Ellipsis:
		t = e.TypeFromEllipsis(file, v)
	case *ast.MapType:
		t = e.TypeFromMapType(file, v)
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
		log.Fatalf("[ BUG ] unknown ast.Expr: %T file: %s", v, file.Name.Name)
	}

	return t
}

func (e Extractor) TypeFromIdent(file *ast.File, id *ast.Ident) types.Type {
	if types.IsBuiltIn(id.Name) {
		return types.NewType(id.Name, "", "")
	}
	if extract.ImportStringForPackage(file, file.Name.Name) == "" {
		fmt.Println(id.Name, "empty import")
	}
	return types.NewType(id.Name, file.Name.Name, extract.ImportStringForPackage(file, file.Name.Name))
}

func (e Extractor) TypeFromSelectorExpr(file *ast.File, se *ast.SelectorExpr) types.Type {
	var p string

	switch pIdent := se.X.(type) {
	case *ast.Ident:
		p = pIdent.Name
	default:
		log.Fatal("[ BUG ] unknown ast.SelectorExpr.X", pIdent)
	}

	return types.NewType(se.Sel.Name, p, extract.ImportStringForPackage(file, p))
}

func (e Extractor) TypeFromStarExpr(file *ast.File, se *ast.StarExpr) types.Type {
	var t types.Type

	switch x := se.X.(type) {
	case *ast.Ident:
		t = e.TypeFromIdent(file, x)
	case *ast.SelectorExpr:
		t = e.TypeFromSelectorExpr(file, x)
	case *ast.ArrayType:
		t = e.TypeFromArrayType(file, x)
	default:
		log.Fatalf("[ TODO ] unknown ast.StarExpr.X: %T", x)
	}

	return types.Pointer{Type: t}
}

func (e Extractor) TypeFromEllipsis(file *ast.File, el *ast.Ellipsis) types.Type {
	var t types.Type

	switch x := el.Elt.(type) {
	case *ast.Ident:
		t = e.TypeFromIdent(file, x)
	case *ast.SelectorExpr:
		t = e.TypeFromSelectorExpr(file, x)
	case *ast.ArrayType:
		t = e.TypeFromArrayType(file, x)
	case *ast.InterfaceType:
		t = types.NewType("interface{}", "", "")
		t.SetIsInterface()
	default:
		log.Fatalf("[ TODO ] unknown ast.Ellipsis.Elt: %T", x)
	}

	return types.Pointer{Type: t}
}

func (e Extractor) TypeFromArrayType(file *ast.File, at *ast.ArrayType) types.Type {
	var t types.Type

	switch elt := at.Elt.(type) {
	case *ast.Ident:
		t = e.TypeFromIdent(file, elt)
	case *ast.SelectorExpr:
		t = e.TypeFromSelectorExpr(file, elt)
	case *ast.StarExpr:
		t = e.TypeFromStarExpr(file, elt)
	case *ast.InterfaceType:
		t = types.NewType("interface{}", "", "")
		t.SetIsInterface()
	case *ast.FuncType:
		fmt.Println("ast.FuncType cannot be used in transport")
		return nil
	case *ast.ArrayType:
		return e.TypeFromArrayType(file, elt)
	default:
		log.Fatalf("[ TODO ] unknown ast.ArrayType.Elt: %T", elt)
	}

	return types.Slice{Type: t}
}

func (e Extractor) TypeFromMapType(file *ast.File, mt *ast.MapType) types.Type {
	kType := e.TypeFromExpr(file, mt.Key)
	vType := e.TypeFromExpr(file, mt.Value)
	if kType == nil || vType == nil {
		return nil
	}
	return types.NewMapType(kType, vType)
}
