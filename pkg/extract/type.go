package extract

import (
	"fmt"
	"github.com/nullc4t/og/internal/types"
	"github.com/nullc4t/og/pkg/utils"
	"github.com/spf13/viper"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"path/filepath"
	"strings"
)

// TypesFromASTFile exported func TODO: edit
func TypesFromASTFile(file *ast.File) ([]*types.Interface, []*types.Struct) {
	var (
		ifaces  []*types.Interface
		structs []*types.Struct
	)

	for _, decl := range file.Decls {
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

			if i := InterfaceFromTypeSpec(file, typeSpec); i != nil {
				ifaces = append(ifaces, i)
			}

			if s := StructFromTypeSpec(file, typeSpec); s != nil {
				structs = append(structs, s)
			}
		}
	}

	return ifaces, structs
}

// DepPackagePathFromModule returns fs path for pkg, given moduleName and modulePath
func DepPackagePathFromModule(moduleName, modulePath, pkg string) string {
	return strings.Replace(pkg, moduleName, modulePath, 1)
}

// SourcePath4Package returns pkg source dir path;
func SourcePath4Package(moduleName, modulePath, pkgImportPath, fileOrPackage string) (string, error) {
	if strings.Contains(pkgImportPath, moduleName) {
		return DepPackagePathFromModule(moduleName, modulePath, pkgImportPath), nil

	}

	if !strings.Contains(pkgImportPath, "/") {
		return "", fmt.Errorf("%s is builtin package", pkgImportPath)
	}

	goModData, err := GoMod(filepath.Dir(fileOrPackage))
	if err != nil {
		return "", err
	}

	deps, err := DependenciesFromGoMod(string(goModData))
	if err != nil {
		return "", err
	}

	dep := DependencyForPackage(pkgImportPath, deps)
	if dep == nil {
		return "", fmt.Errorf("dependency for %s not found", pkgImportPath)
	}

	return DepPackagePathFromModule(dep.Module, dep.Path, pkgImportPath), nil
}

//// ImportedType find package with type t and return parsed struct/interface
//func ImportedType(file *ast.File, filePath, moduleName, modulePath string, t types.Type) (*types.ExchangeStruct, error) {
//	typeImportPath := ImportStringForPackage(file, t.Package())
//
//	packagePath, err := SourcePath4Package(moduleName, modulePath, typeImportPath, filePath)
//	if err != nil {
//		return nil, err
//	}
//
//	goFiles, err := GoSourceFilesFromPackage(packagePath)
//	if err != nil {
//		return nil, err
//	}
//
//	var exchangeStruct *types.ExchangeStruct
//
//	for _, goFile := range goFiles {
//		fset := token.NewFileSet()
//		f, err := parser.ParseFile(fset, goFile, nil, parser.ParseComments)
//		if err != nil {
//			return nil, err
//		}
//
//		ast.Inspect(f, func(node ast.Node) bool {
//			switch typeSpec := node.(type) {
//			case *ast.TypeSpec:
//				if !typeSpec.Name.IsExported() {
//					return false
//				}
//				if typeSpec.Name.Name != t.Name() {
//					return false
//				}
//
//				switch vv := typeSpec.Type.(type) {
//				case *ast.InterfaceType:
//					fmt.Println("interface found:", typeSpec.Name.Name)
//					exchangeStruct = &types.ExchangeStruct{
//						StructName:  typeSpec.Name.Name,
//						IsInterface: true,
//					}
//					return false
//				case *ast.StructType:
//					fmt.Println("struct found:", typeSpec.Name.Name)
//					exchangeStruct = &types.ExchangeStruct{
//						StructName: typeSpec.Name.Name,
//						Fields:     ArgsFromFields(f, vv.Fields),
//					}
//					return false
//				default:
//					fmt.Printf("unknown type %T\n", vv)
//				}
//			}
//			return true
//		})
//
//		if exchangeStruct != nil {
//			break
//		}
//	}
//	return exchangeStruct, nil
//}

// ImportedTypeFromPackage find package with type t and return parsed struct/interface
// packagePath must be *file path* to package source

//func ImportedTypeFromPackage(packagePath string, t types.Type) (*types.Interface, *types.Struct, error) {
//	goFiles, err := GoSourceFilesFromPackage(packagePath)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	for _, goFile := range goFiles {
//		fset := token.NewFileSet()
//		f, err := parser.ParseFile(fset, goFile, nil, parser.ParseComments)
//		if err != nil {
//			return nil, nil, err
//		}
//
//		ifaces, structs := TypesFromASTFile(f)
//
//		for _, iface := range ifaces {
//			if iface.Name == t.Name() {
//				return &iface, nil, nil
//			}
//		}
//
//		for _, str := range structs {
//			if str.Name == t.Name() {
//				return nil, &str, nil
//			}
//		}
//	}
//
//	return nil, nil, nil
//}

// DependencyForPackage return corresponding *types.Dependency for pkg
func DependencyForPackage(pkg string, deps []types.Dependency) *types.Dependency {
	for _, dep := range deps {
		if strings.Contains(pkg, dep.Module) {
			return &dep
		}
	}
	return nil
}

// GoSourceFilesFromPackage returns list of *.go src files in pkgPath package except doc.go, *_test.go
func GoSourceFilesFromPackage(pkgPath string) ([]string, error) {
	goFiles, err := filepath.Glob(filepath.Join(pkgPath, "*.go"))
	if err != nil {
		return nil, err
	}
	var files []string
	for _, goFile := range goFiles {
		if strings.HasSuffix(goFile, "doc.go") || strings.HasSuffix(goFile, "_test.go") {
			continue
		}
		files = append(files, goFile)
	}
	return files, nil
}

// ImportedTypesRecursive  looks for imported types in ifaces & structs.
// Try to find in sources. Return found.
func ImportedTypesRecursive(file *types.GoFile, ifaces []*types.Interface, structs []*types.Struct, depth int) ([]*types.Interface, []*types.Struct, error) {
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

	var foundIfaces []*types.Interface
	var foundStructs []*types.Struct
typeLoop:
	for _, data := range types2Find {
		for _, s := range viper.GetStringSlice("exclude_types") {
			if data.Type.String() == s {
				continue typeLoop
			}
		}

		i, s, err := TypeFromPackage(file, data.Type.Package(), data.Type.Name(), depth)
		if err != nil {
			fmt.Printf("recursive find type %s error: %s\n", data.Type, err.Error())
		}

		if depth > 0 {
			depI, depS, err := ImportedTypesRecursive(file, i, s, depth-1)
			if err != nil {
				return foundIfaces, foundStructs, err
			}

			foundIfaces = utils.AddIfNotContains[*types.Interface](foundIfaces, func(a, b *types.Interface) bool {
				return a.Name == b.Name && len(a.Methods) == len(b.Methods)
			}, depI...)

			foundStructs = utils.AddIfNotContains[*types.Struct](foundStructs, func(a, b *types.Struct) bool {
				return a.Name == b.Name && len(a.Fields) == len(b.Fields)
			}, depS...)
		}

		foundIfaces = utils.AddIfNotContains[*types.Interface](foundIfaces, func(a, b *types.Interface) bool {
			return a.Name == b.Name && len(a.Methods) == len(b.Methods)
		}, i...)

		foundStructs = utils.AddIfNotContains[*types.Struct](foundStructs, func(a, b *types.Struct) bool {
			return a.Name == b.Name && len(a.Fields) == len(b.Fields)
		}, s...)
	}

	return foundIfaces, foundStructs, nil
}

// TypeFromPackage 1. get package import path.
// 2. Get source files in this package.
// 3. Look for name definition in package files
func TypeFromPackage(file *types.GoFile, pkgName string, name string, depth int) ([]*types.Interface, []*types.Struct, error) {
	if depth == 0 {
		return nil, nil, nil
	}
	fmt.Println("recursive find types for ", file.FilePath)
	var (
		packagePath string
		err         error
	)
	if pkgName == "" || pkgName == file.Package {
		packagePath, err = SourcePath4Package(file.Module, file.ModulePath, file.ImportPath(), file.FilePath)
	} else {

		packagePath, err = SourcePath4Package(file.Module, file.ModulePath, ImportStringForPackage(file.AST, pkgName), file.FilePath)
	}
	if err != nil {
		return nil, nil, err
	}

	goFiles, err := GoSourceFilesFromPackage(packagePath)
	if err != nil {
		return nil, nil, err
	}

	fmt.Println(packagePath, "go files:", file.FilePath)

	for _, goFile := range goFiles {
		ifaces, structs, err := ParseFile(goFile, name, depth-1)
		if err != nil {
			return nil, nil, err
		}

		for _, iface := range ifaces {
			if iface.Name == name {
				return []*types.Interface{iface}, nil, nil
			}
		}

		for _, str := range structs {
			if str.Name == name {
				return nil, []*types.Struct{str}, nil
			}
		}
	}

	return nil, nil, nil
}

// ParseFile exported func TODO: edit
func ParseFile(path string, query string, depth int) ([]*types.Interface, []*types.Struct, error) {
	fmt.Println("parsing file ", path)
	f, err := GoFile(path)
	if err != nil {
		return nil, nil, err
	}

	ifaces, structs, err := TypeDefs(f, query, depth)
	if err != nil {
		fmt.Println(path, " parse file error:", err)
		return nil, nil, err
	}

	fmt.Println(path, "parsed")
	return ifaces, structs, nil
}

// TypeDefs exported func TODO: edit
func TypeDefs(file *types.GoFile, name string, depth int) ([]*types.Interface, []*types.Struct, error) {
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

			if i := InterfaceFromTypeSpec(file.AST, typeSpec); i != nil {
				if name != "" && name == i.Name {
					ifaces = append(ifaces, i)
					break
				}
				ifaces = append(ifaces, i)
			}

			if s := StructFromTypeSpec(file.AST, typeSpec); s != nil {
				if name != "" && name == s.Name {
					structs = append(structs, s)
					break
				}
				structs = append(structs, s)
			}
		}
	}
	i, s, err := ImportedTypesRecursive(file, ifaces, structs, depth)
	if err != nil {
		return nil, nil, err
	}

	ifaces = append(ifaces, i...)
	structs = append(structs, s...)

	fmt.Println("type defs ", file.FilePath, " done")
	return ifaces, structs, nil
}

// GoFile exported func TODO: edit
func GoFile(path string) (*types.GoFile, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("file abs path error: %w", err)
	}

	goMod, err := SearchFileUp("go.mod", filepath.Dir(absPath), types.SearchUpDirLimit)
	if err != nil {
		return nil, err
	}

	absModulePath, err := filepath.Abs(filepath.Dir(goMod))
	if err != nil {
		return nil, fmt.Errorf("file abs path error: %w", err)
	}

	module, err := ModuleNameFromGoMod(goMod)
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
