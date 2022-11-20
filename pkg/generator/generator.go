package generator

import (
	"bytes"
	"github.com/nullc4t/og/pkg/editor"
	"github.com/nullc4t/og/pkg/source"
	"go/format"
	"text/template"
)

type (
	Dot        any
	DotGetter  func() Dot
	SourceCode = *bytes.Buffer

	FileWriter func(path string, data SourceCode) error
)

type (
	Unit struct {
		src        *source.File
		template   *template.Template
		dot        Dot
		editAfter  []editor.CodeEditor
		dstPath    string
		fileWriter FileWriter
	}
)

func NewUnit(src *source.File, template *template.Template, dot Dot, editAfter []editor.CodeEditor, dstPath string, fileWriter FileWriter) *Unit {
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
	//u.editAfter = append(u.editAfter, u.AddSourcePackageToImports)
	u.editAfter = append(u.editAfter, editor.AddImportsFactory(src.ImportPath()))
	u.editAfter = append(u.editAfter, Formatter)
	return u
}

func (u Unit) Generate() error {
	tmp := new(bytes.Buffer)

	err := u.template.Execute(tmp, u.dot)
	if err != nil {
		return err
	}

	for _, codeEditor := range u.editAfter {
		tmp, err = codeEditor(tmp)
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
