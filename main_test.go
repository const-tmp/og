package main

import (
	"fmt"
	"github.com/nullc4t/og/pkg/extract"
	"github.com/stretchr/testify/require"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestName(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	err = filepath.Walk(wd, func(path string, info fs.FileInfo, err error) error {
		t.Log(path, info, err)
		return nil
	})
	require.NoError(t, err)
	err = filepath.WalkDir(wd, func(path string, d fs.DirEntry, err error) error {
		t.Log(path, d, err)
		return nil
	})
	files, err := filepath.Glob("go.mod")
	require.NoError(t, err)
	for i, file := range files {
		t.Log(i, file)
	}
	info, err := os.ReadDir(wd)
	require.NoError(t, err)
	for i, entry := range info {
		t.Log(i, entry.Name(), entry.Type(), entry.IsDir())
		fi, err := entry.Info()
		require.NoError(t, err)
		t.Log(fi)
	}
}

func TestPath(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	t.Log(wd)
	t.Log(filepath.Join(wd, ".."))

	path, err := extract.SearchFileUp("go.mod", wd, 3)
	require.NoError(t, err)
	t.Log(path)

	path, err = extract.SearchFileDown("go.mod")
	require.NoError(t, err)
	t.Log(path)

	mod, err := extract.ModuleNameFromGoMod(path)
	require.NoError(t, err)
	t.Log(mod)
}

func TestComments(t *testing.T) {
	// parse file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "/Users/hightime/code/kk/core/internal/types/core.go", nil, parser.ParseComments)
	require.NoError(t, err)

	var comments []*ast.CommentGroup
	//ast.Inspect(node, func(n ast.Node) bool {
	//	// collect comments
	//	c, ok := n.(*ast.CommentGroup)
	//	if ok {
	//		comments = append(comments, c)
	//	}
	//	// handle function declarations without documentation
	//	fn, ok := n.(*ast.TypeSpec)
	//	if ok {
	//		if fn.Name.IsExported() && fn.Doc.Text() == "" {
	//			// print warning
	//			fmt.Printf("exported function declaration without documentation found on line %d: \n\t%s\n", fset.Position(fn.Pos()).Line, fn.Name.Name)
	//		}
	//	}
	//	return true
	//})
	ast.Inspect(node, func(n ast.Node) bool {
		// collect comments
		c, ok := n.(*ast.CommentGroup)
		if ok {
			comments = append(comments, c)
		}

		// handle function declarations without documentation
		fn, ok := n.(*ast.TypeSpec)
		if ok {
			if fn.Name.IsExported() && fn.Doc.Text() == "" {
				// print warning
				fmt.Printf("exported function declaration without documentation found on line %d: \n\t%s\n", fset.Position(fn.Pos()).Line, fn.Name.Name)
				// create todo-comment
				comment := &ast.Comment{
					Text:  "// TODO: document exported function",
					Slash: fn.Pos() - 1,
				}
				// create CommentGroup and set it to the function's documentation comment
				cg := &ast.CommentGroup{
					List: []*ast.Comment{comment},
				}
				fn.Doc = cg
				fmt.Println()
			}
		}
		return true
	})
	// set ast's comments to the collected comments
	node.Comments = comments
	// write new ast to file
	f, err := os.Create("new.go")
	defer f.Close()
	if err := printer.Fprint(f, fset, node); err != nil {
		log.Fatal(err)
	}

}
