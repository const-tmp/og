package generator

import (
	"bytes"
	"github.com/nullc4t/gensta/pkg/source"
	"go/ast"
	"go/format"
	astparser "go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"text/template"
)

type (
	Dot        any
	DotGetter  func() Dot
	SourceCode = *bytes.Buffer
	CodeEditor func(SourceCode) (SourceCode, error)
	FileWriter func(path string, data SourceCode) error
)

type (
	Unit struct {
		src        *source.File
		template   *template.Template
		dot        Dot
		editAfter  []CodeEditor
		dstPath    string
		fileWriter FileWriter
	}
)

func NewUnit(src *source.File, template *template.Template, dot Dot, editAfter []CodeEditor, dstPath string, fileWriter FileWriter) *Unit {
	return &Unit{
		src:        src,
		template:   template,
		dot:        dot,
		editAfter:  editAfter,
		dstPath:    dstPath,
		fileWriter: fileWriter,
	}
}

// New returns new codegen Unit that can be Unit.Generate()'ed and written to FileWriter
func New(src *source.File, tmpl *template.Template, dot Dot, fw FileWriter, dstPath string) Unit {
	u := Unit{
		src:        src,
		template:   tmpl,
		dot:        dot,
		dstPath:    dstPath,
		fileWriter: fw,
	}
	u.editAfter = append(u.editAfter, u.AddSourcePackageToImports)
	u.editAfter = append(u.editAfter, Formatter)
	return u
}

func (u Unit) AddSourcePackageToImports(code SourceCode) (SourceCode, error) {
	fset := token.NewFileSet()

	file, err := astparser.ParseFile(fset, "", code, astparser.ParseComments)
	if err != nil {
		return nil, err
	}

	ok := astutil.AddImport(fset, file, u.src.ImportPath())
	if !ok {
		return nil, err
	}

	ast.SortImports(fset, file)

	tmp := new(bytes.Buffer)

	err = printer.Fprint(tmp, fset, file)
	if err != nil {
		return nil, err
	}

	return tmp, nil
}

func (u Unit) Generate() error {
	tmp := new(bytes.Buffer)

	err := u.template.Execute(tmp, u.dot)
	if err != nil {
		return err
	}

	for _, editor := range u.editAfter {
		tmp, err = editor(tmp)
		if err != nil {
			return err
		}
	}

	formatted, err := format.Source(tmp.Bytes())
	if err != nil {
		return err
	}

	return u.fileWriter(u.dstPath, bytes.NewBuffer(formatted))
}

func Formatter(code SourceCode) (SourceCode, error) {
	res, err := format.Source(code.Bytes())
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(res), nil
}
