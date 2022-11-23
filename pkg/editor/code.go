package editor

import (
	"bytes"
	"go/ast"
	"go/token"
)

type (
	CodeEditor func(*bytes.Buffer) (*bytes.Buffer, error)
	ASTEditor  func(fset *token.FileSet, file *ast.File) (*ast.File, error)
)
