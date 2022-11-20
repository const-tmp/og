package source

import (
	"fmt"
	"github.com/nullc4t/gensta/pkg/inspector"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"path/filepath"
	"strings"
)

type File struct {
	FilePath   string
	Package    string
	Module     string
	ModulePath string
	FSet       *token.FileSet
	AST        *ast.File
}

const SearchUpDirLimit = 5

func NewFile(path string) (*File, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("file abs path error: %w", err)
	}

	goMod, err := inspector.SearchFileUp("go.mod", filepath.Dir(absPath), SearchUpDirLimit)
	if err != nil {
		return nil, err
	}

	absModulePath, err := filepath.Abs(filepath.Dir(goMod))
	if err != nil {
		return nil, fmt.Errorf("file abs path error: %w", err)
	}

	module, err := inspector.GetModuleNameFromGoMod(goMod)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, absPath, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	return &File{
		FilePath:   absPath,
		Package:    file.Name.Name,
		Module:     module,
		ModulePath: absModulePath,
		FSet:       fset,
		AST:        file,
	}, nil
}

func (f File) ImportPath() string {
	return filepath.Dir(strings.Replace(f.FilePath, f.ModulePath, f.Module, 1))
}
