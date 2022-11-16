/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"fmt"
	"github.com/nullc4t/gensta/pkg/inspector"
	"github.com/nullc4t/gensta/pkg/names"
	parser2 "github.com/nullc4t/gensta/pkg/parser"
	"github.com/nullc4t/gensta/pkg/templates"
	"github.com/nullc4t/gensta/pkg/writer"
	"github.com/spf13/cobra"
	"go/ast"
	"go/format"
	astparser "go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"path/filepath"
)

// crudCmd represents the crud command
var crudCmd = &cobra.Command{
	Use:     "crud file-with-types output-dir",
	Aliases: []string{"c", "cr"},
	Short:   "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args:    cobra.ExactArgs(2),
	Example: "gensta gen crud types.go models/",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("crud called")
		srcFile, err := parser2.NewAstra(args[0])
		if err != nil {
			logger.Fatal(err)
		}

		tmpl, err := templates.NewCRUD()
		if err != nil {
			logger.Fatal(err)
		}

		ast.Inspect(srcFile.ASTFile, func(node ast.Node) bool {
			//t.Logf("%T\t%v", node, node)
			switch typeSpec := node.(type) {
			case *ast.TypeSpec:
				v, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					return false
				}
				if len(v.Fields.List) == 0 {
					return false
				}
				if v.Fields.List[0].Names != nil {
					return false
				}
				sel, ok := v.Fields.List[0].Type.(*ast.SelectorExpr)
				if !ok {
					return false
				}
				ident, ok := sel.X.(*ast.Ident)
				if !ok {
					return false
				}
				if ident.Name == "crud" && sel.Sel.Name == "Model" {
					// execute template
					tmp := new(bytes.Buffer)
					err = tmpl.Execute(tmp, map[string]any{
						"Package": names.PackageNameFromType(typeSpec.Name.Name),
						"Type":    names.TypeNameWithPackage(srcFile.Package, typeSpec.Name.Name),
					})
					if err != nil {
						logger.Fatal(err)
					}

					// add imports
					fset := token.NewFileSet()
					file, err := astparser.ParseFile(fset, "", tmp.Bytes(), 0)
					if err != nil {
						logger.Fatal(err)
					}
					//logger.Println("adding import:", srcFile.ImportPath())
					ok := astutil.AddImport(fset, file, srcFile.ImportPath())
					if !ok {
						logger.Fatal("not ok")
					}
					for t, _ := range inspector.GetImportedTypes(srcFile.Astra) {
						p := inspector.ExtractPackageFromType(t)
						if importPath := inspector.GetImportPathForPackage(p, srcFile.Astra); importPath != "" {
							//logger.Println("adding import:", importPath)
							astutil.AddImport(fset, file, importPath)
						}
					}
					ast.SortImports(fset, file)

					tmp = new(bytes.Buffer)
					err = printer.Fprint(tmp, fset, file)
					if err != nil {
						logger.Fatal(err)
					}

					// format code
					formatted, err := format.Source(tmp.Bytes())
					if err != nil {
						logger.Fatal(err)
					}

					err = writer.File(filepath.Join(args[1], names.PackageNameFromType(typeSpec.Name.Name)), "crud.gensta.go", formatted)
					if err != nil {
						logger.Fatal(filepath.Join(args[1], names.PackageNameFromType(typeSpec.Name.Name), "crud.gensta.go"), err)
					}
					logger.Println(filepath.Join(args[1], names.PackageNameFromType(typeSpec.Name.Name), "crud.gensta.go"), "Done")
					return false
				}
			}
			return true
		})
	},
}

func init() {
	genCmd.AddCommand(crudCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// crudCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// crudCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
