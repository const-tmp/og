package filters

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"os"
	"testing"
)

func TestAST(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "/Users/hightime/code/kk/core/internal/types/core.go", nil, parser.ParseComments)
	require.NoError(t, err)

	//for i, decl := range file.Decls {
	//	t.Log(i, decl)
	//}
	ast.Inspect(file, func(node ast.Node) bool {
		//t.Logf("%T\t%v", node, node)
		switch typeSpec := node.(type) {
		case *ast.TypeSpec:
			//t.Logf("%T\t%v", typeSpec, typeSpec)
			if typeSpec.Doc != nil {
				t.Log(typeSpec.Doc.List)
			}
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

func TestASTEdit(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "/Users/hightime/code/kk/core/internal/types/core.go", nil, parser.ParseComments)
	require.NoError(t, err)

	cmap := ast.NewCommentMap(fset, file, file.Comments)
	t.Log(cmap.String())
	for node, groups := range cmap {
		t.Log(node, groups)
	}
	//for s, object := range file.Scope.Objects {
	//	t.Log(s, object, fmt.Sprintf("%T", object))
	//}

	edited := astutil.Apply(file, func(c *astutil.Cursor) bool {
		typeSpec, ok := c.Node().(*ast.TypeSpec)
		if ok && typeSpec.Doc == nil {
			typeSpec.Doc = &ast.CommentGroup{List: []*ast.Comment{
				{Text: fmt.Sprintf("%s test", typeSpec.Name.Name), Slash: typeSpec.Pos()},
			}}
			c.Replace(typeSpec)
		}
		//parent, ok := c.Parent().(*ast.TypeSpec)
		//if ok && c.Name() == "Doc" && c.Node() == nil {
		//	t.Logf("%T\t%v", c.Node(), c.Node())
		//	c.Replace()
		//	cmap[c.Parent()] = append(cmap[c.Parent()], &ast.CommentGroup{List: []*ast.Comment{
		//		{Text: fmt.Sprintf("%s test", parent.Name.Name)},
		//	}})
		//}
		return true
	}, nil)
	require.NoError(t, printer.Fprint(os.Stdout, fset, edited))
}

//ast.Inspect(file, func(node ast.Node) bool {
//	//t.Logf("%T\t%v", node, node)
//	switch typeSpec := node.(type) {
//	case *ast.TypeSpec:
//		//t.Logf("%T\t%v", typeSpec, typeSpec)
//		if typeSpec.Doc != nil {
//			t.Log(typeSpec.Doc.List)
//		}
//		v, ok := typeSpec.Type.(*ast.StructType)
//		if !ok {
//			return false
//		}
//
//		//t.Logf("%T\t%v", v, v)
//		//t.Log(v.Fields.List)
//
//		if len(v.Fields.List) == 0 {
//			return false
//		}
//		if v.Fields.List[0].Names != nil {
//			return false
//		}
//		sel, ok := v.Fields.List[0].Type.(*ast.SelectorExpr)
//		if !ok {
//			return false
//		}
//		ident, ok := sel.X.(*ast.Ident)
//		if !ok {
//			return false
//		}
//		if ident.Name == "crud" && sel.Sel.Name == "Model" {
//			t.Log(typeSpec.Name.Name)
//			return false
//		}
//
//		//case *ast.TypeSpec:
//		//	t.Logf("%T\t%v", v, v)
//	}
//	//gd, ok := node.(*ast.GenDecl)
//	//if !ok {
//	//	return true
//	//}
//	//ast.Print(fset, gd)
//	return true
//})
