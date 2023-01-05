package extract

import (
	"bytes"
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
	"text/template"
)

var (
	structSliceUtil = utils.NewSlice[*types.Struct](func(a, b *types.Struct) bool {
		return a.Name == b.Name && len(a.Fields) == len(b.Fields)
	})
	ifaceSliceUtil = utils.NewSlice[*types.Interface](func(a, b *types.Interface) bool {
		return a.Name == b.Name && len(a.Methods) == len(b.Methods)
	})
)

// TypesFromASTFile exported func TODO: edit
func TypesFromASTFile(file *types.GoFile) ([]*types.Interface, []*types.Struct) {
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

// Path4Package returns pkg source dir path;
func Path4Package(moduleName, modulePath, pkgImportPath, fileOrPackage string) (string, error) {
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

type Context struct {
	Interface map[string]*types.Interface
	Struct    map[string]*types.Struct
	File      map[string]*types.GoFile
	Type      map[string]types.Type
}

func (c Context) GetStruct(t types.Type) *types.Struct {
	return c.Struct[TypeIndex(t.ImportPath(), t.Name())]
}

func (c Context) GetInterface(t types.Type) *types.Interface {
	return c.Interface[TypeIndex(t.ImportPath(), t.Name())]
}

func (c Context) String() string {
	tmpl := template.Must(template.New("").Parse(`
~~~~~~~~~~~~~~~~~~~~~~~~~ Context ~~~~~~~~~~~~~~~~~~~~~~~~
Interfaces:
{{- range $pkg, $iface := .Interface }}
	{{ $pkg }}    {{ $iface.Name }}
{{- end }}

Structs:
{{- range $pkg, $str := .Struct }}
	{{ $pkg }}    {{ $str.Name }}
{{- end }}

Files:
{{- range $pkg, $file := .File }}
	{{ $pkg }}    {{ $file.ImportPath }}
{{- end }}
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
`))
	tmp := new(bytes.Buffer)
	template.Must(tmpl, tmpl.Execute(tmp, c))
	return tmp.String()
}

func NewContext() *Context {
	return &Context{
		Interface: map[string]*types.Interface{},
		Struct:    map[string]*types.Struct{},
		File:      map[string]*types.GoFile{},
		Type:      map[string]types.Type{},
	}
}

func TypeIndex(pkgImportStr, typeName string) string {
	return fmt.Sprintf("%s/%s", pkgImportStr, typeName)
}

// TypesRecursive  looks for non-builtin types in ifaces & structs.
// Try to find in sources. Return found.
func TypesRecursive(ctx *Context, file *types.GoFile, ifaces []*types.Interface, structs []*types.Struct, depth int) ([]*types.Interface, []*types.Struct, error) {
	fmt.Printf("types recursive: %s\tdepth: %d\n", file.FilePath, depth)
	if depth <= 0 {
		return nil, nil, nil
	}

	if ctx == nil {
		ctx = NewContext()
	}

	// find all NOT builtin types
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

		// do not search type if already done
		if ci, ok := ctx.Interface[TypeIndex(data.Type.ImportPath(), data.Type.Name())]; ok {
			foundIfaces = append(foundIfaces, ci)
			continue
		}
		if cs, ok := ctx.Struct[TypeIndex(data.Type.ImportPath(), data.Type.Name())]; ok {
			foundStructs = append(foundStructs, cs)
			continue
		}

		i, s, err := TypeFromPackage(ctx, file, data.Type.Package(), data.Type.Name(), depth)
		if err != nil {
			fmt.Printf("recursive find type %s error: %s\n", data.Type, err.Error())
		}

		if depth > 0 {
			depI, depS, err := TypesRecursive(ctx, file, i, s, depth-1)
			if err != nil {
				return foundIfaces, foundStructs, err
			}

			foundIfaces = ifaceSliceUtil.AppendIfNotExist(foundIfaces, depI...)
			foundStructs = structSliceUtil.AppendIfNotExist(foundStructs, depS...)
		}

		foundIfaces = ifaceSliceUtil.AppendIfNotExist(foundIfaces, i...)
		foundStructs = structSliceUtil.AppendIfNotExist(foundStructs, s...)
	}

	return foundIfaces, foundStructs, nil
}

// TypeFromPackage 1. get package import path.
// 2. Get source files in this package.
// 3. Look for name definition in package files
func TypeFromPackage(ctx *Context, file *types.GoFile, pkgName string, name string, depth int) ([]*types.Interface, []*types.Struct, error) {
	fmt.Printf("looking for type definition: %s\tcurrent file: %s\tpackage: %s\tdepth: %d\n", name, file.FilePath, pkgName, depth)
	if depth == 0 {
		return nil, nil, nil
	}

	var (
		packagePath string
		err         error
	)

	if pkgName == "" || pkgName == file.Package {
		packagePath, err = Path4Package(file.Module, file.ModulePath, file.ImportPath(), file.FilePath)
	} else {

		packagePath, err = Path4Package(file.Module, file.ModulePath, ImportStringForPackage(file, pkgName), file.FilePath)
	}
	if err != nil {
		return nil, nil, err
	}

	goFiles, err := GoSourceFilesFromPackage(packagePath)
	if err != nil {
		return nil, nil, err
	}

	for _, goFile := range goFiles {
		ifaces, structs, err := ParseFile(ctx, goFile, name, depth-1)
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
func ParseFile(ctx *Context, path string, query string, depth int) ([]*types.Interface, []*types.Struct, error) {
	//fmt.Println("parsing file ", path)
	f, err := GoFile(path)
	if err != nil {
		return nil, nil, err
	}

	ctx.File[f.FilePath] = f

	ifaces, structs, err := TypeDefs(ctx, f, query, depth)
	if err != nil {
		fmt.Println(path, " parse file error:", err)
		return nil, nil, err
	}

	return ifaces, structs, nil
}

// TypeDefs exported func TODO: edit
func TypeDefs(ctx *Context, file *types.GoFile, name string, depth int) ([]*types.Interface, []*types.Struct, error) {
	//fmt.Println("getting type defs in ", file.FilePath)
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

			if i := InterfaceFromTypeSpec(file, typeSpec); i != nil {
				if name != "" && name == i.Name {
					ifaces = append(ifaces, i)
					break
				}
				ifaces = append(ifaces, i)
			}

			if s := StructFromTypeSpec(file, typeSpec); s != nil {
				if name != "" && name == s.Name {
					structs = append(structs, s)
					break
				}
				structs = append(structs, s)
			}
		}
	}

	for _, iface := range ifaces {
		ctx.Interface[TypeIndex(file.ImportPath(), iface.Name)] = iface
	}
	for _, str := range structs {
		ctx.Struct[TypeIndex(file.ImportPath(), str.Name)] = str
	}

	i, s, err := TypesRecursive(ctx, file, ifaces, structs, depth)
	if err != nil {
		return nil, nil, err
	}

	ifaces = append(ifaces, i...)
	structs = append(structs, s...)

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
