package types

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"
)

type GoFile struct {
	FilePath   string
	Module     string
	ModulePath string
	Package    string
	FSet       *token.FileSet
	AST        *ast.File
}

const SearchUpDirLimit = 5

func (f GoFile) ImportPath() string {
	return filepath.Dir(strings.Replace(f.FilePath, f.ModulePath, f.Module, 1))
}
