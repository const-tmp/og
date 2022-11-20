package editor

import (
	"bytes"
	"fmt"
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
