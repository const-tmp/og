package filters

import (
	"github.com/stretchr/testify/require"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestAST(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "/Users/hightime/code/kk/auth/internal/types/types.go", nil, 0)
	require.NoError(t, err)

	//for i, decl := range file.Decls {
	//	t.Log(i, decl)
	//}
	ast.Inspect(file, func(node ast.Node) bool {
		//t.Logf("%T\t%v", node, node)
		switch typeSpec := node.(type) {
		case *ast.TypeSpec:
			//t.Logf("%T\t%v", typeSpec, typeSpec)

			v, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				return false
			}

			//t.Logf("%T\t%v", v, v)
			//t.Log(v.Fields.List)

			if len(v.Fields.List) == 0 {
				return false
			}
			if v.Fields.List[0].Names != nil {
				return false
			}
			sel, ok := v.Fields.List[0].Type.(*ast.SelectorExpr)
			if !ok {
				return false
			}
			ident, ok := sel.X.(*ast.Ident)
			if !ok {
				return false
			}
			if ident.Name == "crud" && sel.Sel.Name == "Model" {
				t.Log(typeSpec.Name.Name)
				return false
			}

			//case *ast.TypeSpec:
			//	t.Logf("%T\t%v", v, v)
		}
		//gd, ok := node.(*ast.GenDecl)
		//if !ok {
		//	return true
		//}
		//ast.Print(fset, gd)
		return true
	})
}
