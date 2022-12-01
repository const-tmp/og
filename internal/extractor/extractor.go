package extractor

import (
	"fmt"
	"github.com/nullc4t/og/internal/types"
	"github.com/nullc4t/og/pkg/extract"
	"github.com/nullc4t/og/pkg/utils"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"path/filepath"
	"strings"
)

type (
	Extractable interface {
		types.Interface | types.Struct
	}

	Extractor struct {
		ModuleMap ModuleMap
		fset      *token.FileSet
		TypeMap   TypeMap
		//File      *ast.File
	}

	TypeData struct {
		Type      types.Type
		Struct    *types.Struct
		Interface *types.Interface
	}

	TypeMap   map[string]TypeData
	ModuleMap map[string]*types.Module
)

func NewExtractor() *Extractor {
	return &Extractor{ModuleMap: make(ModuleMap), TypeMap: make(TypeMap), fset: token.NewFileSet()}
}

func (m TypeMap) Contains(s string) bool {
	_, ok := m[s]
	return ok
}

func (m TypeMap) Add(t types.Type) {
	m[t.Package()+t.Name()] = TypeData{Type: t}
}

func (m ModuleMap) Add(f *types.GoFile) error {
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
		m[f.Module] = &types.Module{
			Name:     f.Module,
			Path:     f.ModulePath,
			Packages: map[string]*types.Package{f.Package: NewPackageFromGoFile(f)},
		}
	}
	return nil
}

func (e Extractor) ParseFile(path string) error {
	fmt.Println("parsing file ", path)
	f, err := GoFile(path)
	if err != nil {
		return err
	}

	err = e.ModuleMap.Add(f)
	if err != nil {
		return err
	}

	ifaces, structs, err := e.TypeDefs(f)
	if err != nil {
		fmt.Println(path, " parse file error:", err)
		return err
	}

	for _, iface := range ifaces {
		e.ModuleMap[f.Module].Packages[f.Package].Interfaces = append(e.ModuleMap[f.Module].Packages[f.Package].Interfaces, iface)
	}
	for _, str := range structs {
		e.ModuleMap[f.Module].Packages[f.Package].Structs = append(e.ModuleMap[f.Module].Packages[f.Package].Structs, str)
	}

	fmt.Println(path, "parsed")
	return nil
}

func NewPackageFromGoFile(f *types.GoFile) *types.Package {
	return &types.Package{
		Name:       f.Package,
		ImportPath: f.ImportPath(),
		Path:       f.FilePath,
		Files:      map[string]*types.GoFile{f.FilePath: f},
	}
}

func (e Extractor) TypeDefs(file *types.GoFile) ([]*types.Interface, []*types.Struct, error) {
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
				ifaces = append(ifaces, i)
			}

			if s := e.StructFromTypeSpec(file.AST, typeSpec); s != nil {
				structs = append(structs, s)
			}
		}
	}

	// map parsed types
TypeLoop:
	for _, data := range e.TypeMap {
		for _, iface := range ifaces {
			if iface.Name == data.Type.Name() && strings.Contains(data.Type.ImportPath(), file.Module) {
				data.Interface = iface
				continue TypeLoop
			}
		}
		for _, str := range structs {
			if str.Name == data.Type.Name() && strings.Contains(data.Type.ImportPath(), file.Module) {
				data.Struct = str
				continue TypeLoop
			}
		}
	}

	for _, str := range structs {
		for _, field := range str.Fields {
			if field.Type.IsImported() {
				fmt.Println("imported type:", field.Type)
				if e.checkIsTypeParsed(field.Type) {
					continue
				}

				err := e.recursiveParsePackage(file, field.Type.Package())
				if err != nil {
					fmt.Printf("recursive parse package %s error: %s", field.Type.Package(), err)
					//return nil, nil, err
				}
			}
		}
	}

	for _, iface := range ifaces {
		for _, method := range iface.Methods {
			for _, arg := range method.Args {
				if arg.Type == nil {
					utils.BugPanic(fmt.Sprint(method.Name, arg.Name, "null Type"))
				}
				if arg.Type.IsImported() {
					fmt.Println("imported type:", arg.Type)
					if e.checkIsTypeParsed(arg.Type) {
						continue
					}

					err := e.recursiveParsePackage(file, arg.Type.Package())
					if err != nil {
						fmt.Printf("recursive parse package %s error: %s", arg.Type.Package(), err)
						//return nil, nil, err
					}
				}
			}
			for _, arg := range method.Results.Args {
				if arg.Type.IsImported() {
					fmt.Println("imported type:", arg.Type)
					if e.checkIsTypeParsed(arg.Type) {
						continue
					}

					err := e.recursiveParsePackage(file, arg.Type.Package())
					if err != nil {
						fmt.Printf("recursive parse package %s error: %s", arg.Type.Package(), err)
						//return nil, nil, err
					}
				}
			}
		}

	}

	fmt.Println("getting type defs in ", file.FilePath, " done")
	return ifaces, structs, nil
}

func (e Extractor) checkIsTypeParsed(ty types.Type) bool {
	t := e.TypeMap[ty.Package()+ty.Name()]
	if t.Interface != nil || t.Struct != nil {
		fmt.Println(ty, "has interface=", t.Interface != nil, "has struct=", t.Struct != nil)
		return true
	}
	return false
}

func (e Extractor) recursiveParsePackage(file *types.GoFile, pkgName string) error {
	fmt.Println("recursiveParsePackage:", pkgName)

	packagePath, err := extract.SourcePath4Package(file.Module, file.ModulePath, extract.ImportStringForPackage(file.AST, pkgName), file.FilePath)
	if err != nil {
		return err
	}

	goFiles, err := extract.GoSourceFilesFromPackage(packagePath)
	if err != nil {
		return err
	}

	fmt.Println(packagePath, "go files:", file.FilePath)

	for _, goFile := range goFiles {
		err = e.ParseFile(goFile)
		if err != nil {
			return err
		}
	}

	return nil
}

//func (e Extractor) SourcePath4Package(f *types.GoFile) (string, error) {
//	return extract.SourcePath4Package(f.Module, f.ModulePath, f.Package, f.FilePath)
//}

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
	var args []*types.Arg

	for _, arg := range fields.List {
		var (
			names []string
			t     types.Type
		)

		for _, name := range arg.Names {
			names = append(names, name.Name)
		}

		t = e.TypeFromExpr(file, arg.Type)

		if e.TypeMap.Contains(t.Package() + t.Name()) {
			t = e.TypeMap[t.Package()+t.Name()].Type
		} else {
			e.TypeMap.Add(t)
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

		switch len(field.Names) {
		case 1:
			s.Fields = append(s.Fields, types.Field{
				Name: field.Names[0].Name,
				Type: e.TypeFromExpr(file, field.Type),
				Tag:  tag,
			})
		case 0:
			s.Fields = append(s.Fields, types.Field{
				Type: e.TypeFromExpr(file, field.Type),
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
		t = e.TypeFromIdent(v)
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
	default:
		log.Fatalf("[ BUG ] unknown ast.Expr: %T file: %s", v, file.Name.Name)
	}

	return t
}

func (e Extractor) TypeFromIdent(id *ast.Ident) types.Type {
	return types.NewType(id.Name, "", "")
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
		t = e.TypeFromIdent(x)
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
		t = e.TypeFromIdent(x)
	case *ast.SelectorExpr:
		t = e.TypeFromSelectorExpr(file, x)
	case *ast.ArrayType:
		t = e.TypeFromArrayType(file, x)
	default:
		log.Fatalf("[ TODO ] unknown ast.Ellipsis.Elt: %T", x)
	}

	return types.Pointer{Type: t}
}

func (e Extractor) TypeFromArrayType(file *ast.File, at *ast.ArrayType) types.Type {
	var t types.Type

	switch elt := at.Elt.(type) {
	case *ast.Ident:
		t = e.TypeFromIdent(elt)
	case *ast.SelectorExpr:
		t = e.TypeFromSelectorExpr(file, elt)
	case *ast.StarExpr:
		t = e.TypeFromStarExpr(file, elt)
	default:
		log.Fatalf("[ TODO ] unknown ast.ArrayType.Elt: %T", elt)
	}

	return types.Slice{Type: t}
}

func (e Extractor) TypeFromMapType(file *ast.File, mt *ast.MapType) types.Type {
	kType := e.TypeFromExpr(file, mt.Key)
	vType := e.TypeFromExpr(file, mt.Value)
	return types.NewMapType(kType, vType)
}
