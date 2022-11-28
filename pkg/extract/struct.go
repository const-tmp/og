package extract

import (
	"fmt"
	"github.com/nullc4t/og/internal/types"
	"github.com/nullc4t/og/pkg/utils"
	"go/ast"
)

func StructFromTypeSpec(file *ast.File, typeSpec *ast.TypeSpec) *types.Struct {
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
				Type: TypeFromExpr(file, field.Type),
				Tag:  tag,
			})
		case 0:
			s.Fields = append(s.Fields, types.Field{
				Type: TypeFromExpr(file, field.Type),
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
