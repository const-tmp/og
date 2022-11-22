package editor

import (
	"bytes"
	"go/ast"
)

type (
	CodeEditor func(*bytes.Buffer) (*bytes.Buffer, error)
	ASTEditor  func(file *ast.File) (*ast.File, error)
)
