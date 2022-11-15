package parser

import (
	"github.com/nullc4t/gensta/pkg/inspector"
	"github.com/vetcher/go-astra"
	"github.com/vetcher/go-astra/types"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"path/filepath"
	"strings"
)

type SourceFile struct {
	FilePath   string
	Package    string
	Module     string
	ModulePath string
	Astra      *types.File
	ASTFile    *ast.File
	FSet       *token.FileSet
}

const SearchUpDirLimit = 5

func NewAstra(srcPath string) (*SourceFile, error) {
	f, err := astra.ParseFile(srcPath)
	if err != nil {
		log.Println("astra error:", err)
	}

	srcGoMod, err := inspector.SearchFileUp("go.mod", filepath.Dir(srcPath), SearchUpDirLimit)
	if err != nil {
		return nil, err
	}

	modulePath := filepath.Dir(srcGoMod)

	srcModule, err := inspector.GetModuleNameFromGoMod(srcGoMod)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, srcPath, nil, 0)
	if err != nil {
		log.Fatal(err)
	}

	return &SourceFile{
		FilePath:   srcPath,
		Astra:      f,
		Package:    file.Name.Name,
		Module:     srcModule,
		ModulePath: modulePath,
		ASTFile:    file,
		FSet:       fset,
	}, nil
}

func (f SourceFile) ImportPath() string {
	return filepath.Dir(strings.Replace(f.FilePath, f.ModulePath, f.Module, 1))
}
