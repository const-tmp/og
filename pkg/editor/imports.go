package editor

import (
	"bytes"
	"fmt"
	"github.com/nullc4t/og/internal/types"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
)

func AddImportsFactory(imports ...string) CodeEditor {
	return func(code *bytes.Buffer) (*bytes.Buffer, error) {
		fset := token.NewFileSet()

		file, err := parser.ParseFile(fset, "", code, parser.ParseComments)
		if err != nil {
			return nil, err
		}

		for _, s := range imports {
			ok := astutil.AddImport(fset, file, s)
			if !ok {
				return nil, fmt.Errorf("add import %s is not ok", s)
			}
		}

		ast.SortImports(fset, file)

		tmp := new(bytes.Buffer)

		err = printer.Fprint(tmp, fset, file)
		if err != nil {
			return nil, err
		}

		return tmp, nil

	}
}

func AddNamedImportsFactory(imports ...types.Import) CodeEditor {
	return func(code *bytes.Buffer) (*bytes.Buffer, error) {
		fset := token.NewFileSet()

		file, err := parser.ParseFile(fset, "", code, parser.ParseComments)
		if err != nil {
			return nil, err
		}

		for _, imp := range imports {
			if imp.IsAliasedImportRequired() {
				ok := astutil.AddNamedImport(fset, file, imp.Name, imp.Path)
				if !ok {
					return nil, fmt.Errorf("add import %s %s is not ok", imp.Name, imp.Path)
				}
			} else {
				ok := astutil.AddImport(fset, file, imp.Path)
				if !ok {
					return nil, fmt.Errorf("add import %s %s is not ok", imp.Name, imp.Path)
				}
			}
		}

		ast.SortImports(fset, file)

		tmp := new(bytes.Buffer)

		err = printer.Fprint(tmp, fset, file)
		if err != nil {
			return nil, err
		}

		return tmp, nil

	}
}

func ASTImportsFactory(imports ...types.Import) ASTEditor {
	return func(fset *token.FileSet, file *ast.File) (*ast.File, error) {
		for _, imp := range imports {
			if imp.IsAliasedImportRequired() {
				fmt.Println("adding named import", imp.Name, imp.Path)
				ok := astutil.AddNamedImport(fset, file, imp.Name, imp.Path)
				if !ok {
					fmt.Printf("add import %s %s is not ok\n", imp.Name, imp.Path)
					continue
					//return nil, fmt.Errorf("add import %s %s is not ok", imp.Name, imp.Path)
				}
			} else {
				fmt.Println("adding import", imp.Path)
				ok := astutil.AddImport(fset, file, imp.Path)
				if !ok {
					fmt.Printf("add import %s %s is not ok\n", imp.Name, imp.Path)
					continue
					//return nil, fmt.Errorf("add import %s %s is not ok", imp.Name, imp.Path)
				}
			}
		}

		ast.SortImports(fset, file)

		return file, nil

	}
}
