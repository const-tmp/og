package generator

import (
	"bytes"
	"fmt"
	"github.com/nullc4t/og/internal/types"
	"github.com/nullc4t/og/pkg/editor"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"text/template"
)

type (
	Dot        any
	DotGetter  func() Dot
	SourceCode = *bytes.Buffer

	FileWriter func(path string, data SourceCode) error

	Unit struct {
		src           *types.GoFile
		template      *template.Template
		dot           Dot
		editCodeAfter []editor.CodeEditor
		editASTAfter  []editor.ASTEditor
		dstPath       string
		fileWriter    FileWriter
	}
)

func NewUnit(src *types.GoFile, template *template.Template, dot Dot, editAfter []editor.CodeEditor, editASTAfter []editor.ASTEditor, dstPath string, fileWriter FileWriter) *Unit {
	return &Unit{
		src:           src,
		template:      template,
		dot:           dot,
		editCodeAfter: editAfter,
		editASTAfter:  editASTAfter,
		dstPath:       dstPath,
		fileWriter:    fileWriter,
	}
}

// New returns new codegen Unit that can be Unit.Generate()'ed and written to FileWriter
func New(src *types.GoFile, tmpl *template.Template, dot Dot, fw FileWriter, dstPath string) Unit {
	u := Unit{
		src:        src,
		template:   tmpl,
		dot:        dot,
		dstPath:    dstPath,
		fileWriter: fw,
	}
	//u.editCodeAfter = append(u.editCodeAfter, u.AddSourcePackageToImports)
	u.editCodeAfter = append(u.editCodeAfter, editor.AddImportsFactory(src.ImportPath()))
	u.editCodeAfter = append(u.editCodeAfter, Formatter)
	return u
}

func (u Unit) Generate() error {
	tmp := new(bytes.Buffer)

	fmt.Println("executing template", u.dstPath)

	err := u.template.Execute(tmp, u.dot)
	if err != nil {
		return err
	}

	if u.editCodeAfter != nil {
		fmt.Println("code editors", u.dstPath)

		for _, codeEditor := range u.editCodeAfter {
			tmp, err = codeEditor(tmp)
			if err != nil {
				return err
			}
		}
	}
	fmt.Println(tmp.String())
	if u.editASTAfter != nil {
		fmt.Println("parsing AST", u.dstPath)

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, u.dstPath, tmp, parser.ParseComments)
		if err != nil {
			return err
		}

		fmt.Println("AST editors", u.dstPath)
		for _, astEditor := range u.editASTAfter {
			file, err = astEditor(fset, file)
			if err != nil {
				return err
			}
		}

		fmt.Println("printing", u.dstPath)
		tmp = new(bytes.Buffer)
		err = printer.Fprint(tmp, fset, file)
		if err != nil {
			return err
		}
	}

	if u.editCodeAfter != nil || u.editASTAfter != nil {
		formatted, err := format.Source(tmp.Bytes())
		if err != nil {
			return err
		}
		return u.fileWriter(u.dstPath, bytes.NewBuffer(formatted))
	}

	return u.fileWriter(u.dstPath, tmp)
}

func Formatter(code SourceCode) (SourceCode, error) {
	res, err := format.Source(code.Bytes())
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(res), nil
}
