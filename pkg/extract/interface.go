package extract

import (
	"fmt"
	"github.com/nullc4t/og/internal/types"
	"github.com/nullc4t/og/pkg/utils"
	"go/ast"
	"go/token"
)

type Interfaces interface {
	Extract() []types.Interface
}

func InterfacesFromASTFile(file *ast.File) []types.Interface {
	var ifaces []types.Interface

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
		}
	}

	return ifaces
}

func InterfaceFromTypeSpec(file *ast.File, typeSpec *ast.TypeSpec) *types.Interface {
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
			Args:    ArgsFromFields(file, funcType.Params),
			Results: types.Results{Args: ArgsFromFields(file, funcType.Results)},
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
