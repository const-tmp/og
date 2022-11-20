/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/nullc4t/gensta/pkg/generator"
	"github.com/nullc4t/gensta/pkg/names"
	"github.com/nullc4t/gensta/pkg/source"
	"github.com/nullc4t/gensta/pkg/templates"
	"github.com/nullc4t/gensta/pkg/writer"
	"github.com/spf13/cobra"
	"go/ast"
	"io/fs"
	"os"
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
		//srcFile, err := parser2.NewAstra(args[0])
		//if err != nil {
		//	logger.Fatal(err)
		//}

		crudTmpl, err := templates.NewCRUD()
		if err != nil {
			logger.Fatal(err)
		}
		repoTmpl, err := templates.NewRepo()
		if err != nil {
			logger.Fatal(err)
		}

		src, err := source.NewFile(args[0])
		if err != nil {
			logger.Fatal(err)
		}

		ast.Inspect(src.AST, func(node ast.Node) bool {
			switch typeSpec := node.(type) {
			case *ast.TypeSpec:
				v, ok := typeSpec.Type.(*ast.StructType)
				dir := filepath.Join(args[1], names.PackageNameFromType(typeSpec.Name.Name))
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
					dot := map[string]any{
						"Package": names.PackageNameFromType(typeSpec.Name.Name),
						"Type":    names.TypeNameWithPackage(src.Package, typeSpec.Name.Name),
					}

					crudPath := filepath.Join(dir, "crud.gensta.go")
					crudUnit := generator.New(src, crudTmpl, dot, writer.File, crudPath)
					err = crudUnit.Generate()
					if err != nil {
						logger.Fatal("generate crud error:", err)
					}

					repoPath := filepath.Join(dir, "repo.go")
					f, err := os.Open(repoPath)
					if err != nil {
						if !errors.Is(err, fs.ErrNotExist) {
							logger.Fatal("open file", repoPath, "error:", err)
						}
						repoUnit := generator.New(src, repoTmpl, dot, writer.File, repoPath)
						err = repoUnit.Generate()
						if err != nil {
							logger.Fatal("generate repo error:", err)
						}
					}
					f.Close()
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
