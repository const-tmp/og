package extract

import (
	"fmt"
	"github.com/nullc4t/og/internal/types"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

func TypesFromASTFile(file *ast.File) ([]types.Interface, []types.Struct) {
	var (
		ifaces  []types.Interface
		structs []types.Struct
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
				ifaces = append(ifaces, *i)
			}

			if s := StructFromTypeSpec(file, typeSpec); s != nil {
				structs = append(structs, *s)
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
func SourcePath4Package(moduleName, modulePath, pkg, fileOrPackage string) (string, error) {
	if strings.Contains(pkg, moduleName) {
		return DepPackagePathFromModule(moduleName, modulePath, pkg), nil
		//return strings.Replace(pkg, moduleName, modulePath, 1), nil
	}

	if !strings.Contains(pkg, "/") {
		return "", fmt.Errorf("%s is builtin package", pkg)
	}

	goModData, err := GoMod(filepath.Dir(fileOrPackage))
	if err != nil {
		return "", err
	}

	deps, err := DependenciesFromGoMod(string(goModData))
	if err != nil {
		return "", err
	}

	dep := DependencyForPackage(pkg, deps)
	if dep == nil {
		return "", fmt.Errorf("dependency for %s not found", pkg)
	}

	return DepPackagePathFromModule(dep.Module, dep.Path, pkg), nil
}

// ImportedType find package with type t and return parsed struct/interface
func ImportedType(file *ast.File, filePath, moduleName, modulePath string, t types.Type) (*types.ExchangeStruct, error) {
	typeImportPath := ImportStringForPackage(file, t.Package())

	packagePath, err := SourcePath4Package(moduleName, modulePath, typeImportPath, filePath)
	if err != nil {
		return nil, err
	}

	goFiles, err := GoSourceFilesFromPackage(packagePath)
	if err != nil {
		return nil, err
	}

	var exchangeStruct *types.ExchangeStruct

	for _, goFile := range goFiles {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, goFile, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}

		ast.Inspect(f, func(node ast.Node) bool {
			switch typeSpec := node.(type) {
			case *ast.TypeSpec:
				if !typeSpec.Name.IsExported() {
					return false
				}
				if typeSpec.Name.Name != t.Name() {
					return false
				}

				switch vv := typeSpec.Type.(type) {
				case *ast.InterfaceType:
					fmt.Println("interface found:", typeSpec.Name.Name)
					exchangeStruct = &types.ExchangeStruct{
						StructName:  typeSpec.Name.Name,
						IsInterface: true,
					}
					return false
				case *ast.StructType:
					fmt.Println("struct found:", typeSpec.Name.Name)
					exchangeStruct = &types.ExchangeStruct{
						StructName: typeSpec.Name.Name,
						Fields:     ArgsFromFields(f, vv.Fields),
					}
					return false
				default:
					fmt.Printf("unknown type %T\n", vv)
				}
			}
			return true
		})

		if exchangeStruct != nil {
			break
		}
	}
	return exchangeStruct, nil
}

// ImportedTypeFromPackage find package with type t and return parsed struct/interface
// packagePath must be *file path* to package source
func ImportedTypeFromPackage(packagePath string, t types.Type) (*types.Interface, *types.Struct, error) {
	goFiles, err := GoSourceFilesFromPackage(packagePath)
	if err != nil {
		return nil, nil, err
	}

	for _, goFile := range goFiles {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, goFile, nil, parser.ParseComments)
		if err != nil {
			return nil, nil, err
		}

		ifaces, structs := TypesFromASTFile(f)

		for _, iface := range ifaces {
			if iface.Name == t.Name() {
				return &iface, nil, nil
			}
		}

		for _, str := range structs {
			if str.Name == t.Name() {
				return nil, &str, nil
			}
		}
	}

	return nil, nil, nil
}

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
