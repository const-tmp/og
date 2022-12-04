/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/nullc4t/og/internal/types"
	"github.com/nullc4t/og/pkg/editor"
	"github.com/nullc4t/og/pkg/extract"
	"github.com/nullc4t/og/pkg/generator"
	"github.com/nullc4t/og/pkg/templates"
	"github.com/nullc4t/og/pkg/writer"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
)

// grpcConvertersCmd represents the grpcConverters command
var grpcConvertersCmd = &cobra.Command{
	Use:     "grpcConverters exchanges_file.go file.pb.go",
	Aliases: []string{"gc", "grpcconv"},
	Short:   "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("grpcConverters called")

		epTmpl := template.Must(template.New("").Funcs(templates.FuncMap).Parse(templates.GRPCEnpointConverters))
		tyTmpl := template.Must(template.New("").Funcs(templates.FuncMap).Parse(templates.GRPCTypeConverters))

		exchFile, err := extract.GoFile(args[0])
		if err != nil {
			logger.Fatal(err)
		}

		_, exchanges := extract.TypesFromASTFile(exchFile)

		pbFile, err := extract.GoFile(args[1])
		if err != nil {
			logger.Fatal(err)
		}

		_, pbTypes := extract.TypesFromASTFile(pbFile)

		//im := utils.NewSet[types.Import]()
		//
		//for i, exchange := range exchanges {
		//	logger.Println(i, exchange.Name)
		//	im.Add(exchange.UsedImports...)
		//}

		for i, pbType := range pbTypes {
			logger.Println(i, pbType.Name)
		}

		epUnit := generator.NewUnit(
			exchFile, epTmpl, map[string]any{
				"Package":   "transportgrpc",
				"Exchanges": exchanges,
			}, nil,
			//nil,
			[]editor.ASTEditor{editor.ASTImportsFactory(
				types.Import{Path: exchFile.ImportPath()},
				types.Import{Path: pbFile.ImportPath()}),
			},
			filepath.Join(filepath.Join(filepath.Dir(args[0]), "..", "transport", "grpc"), "converters.gen.go"), writer.File,
		)
		err = epUnit.Generate()
		if err != nil {
			logger.Fatal(err)
		}

		tyUnit := generator.NewUnit(
			exchFile, tyTmpl, map[string]any{
				"Package":   "transportgrpc",
				"Exchanges": exchanges,
			}, nil,
			//nil,
			[]editor.ASTEditor{editor.ASTImportsFactory(
				types.Import{Path: exchFile.ImportPath()},
				types.Import{Path: pbFile.ImportPath()}),
			},
			filepath.Join(filepath.Join(filepath.Dir(args[0]), "..", "transport", "grpc"), "type_converters.gen.go"), writer.File,
		)
		err = tyUnit.Generate()
		if err != nil {
			logger.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(grpcConvertersCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// grpcConvertersCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// grpcConvertersCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
