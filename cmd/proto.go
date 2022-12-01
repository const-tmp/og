/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/nullc4t/og/internal/extractor"
	"github.com/nullc4t/og/internal/types"
	"github.com/nullc4t/og/pkg/generator"
	"github.com/nullc4t/og/pkg/templates"
	"github.com/nullc4t/og/pkg/transform"
	"github.com/nullc4t/og/pkg/writer"
	"github.com/spf13/cobra"
	"path/filepath"
	"strings"
	"text/template"
)

// protoCmd represents the proto command
var protoCmd = &cobra.Command{
	Use:   "proto",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("proto called")

		sourceFile, err := extractor.GoFile(args[0])
		if err != nil {
			logger.Fatal(err)
		}

		ex := extractor.NewExtractor()
		err = ex.ParseFile(args[0], "", 2)
		if err != nil {
			logger.Fatal(err)
		}

		tmpl := template.Must(template.New("").Funcs(templates.FuncMap).Parse(templates.Proto2))

		protoPkg := "proto"
		var protoFile = types.ProtoFile{
			GoPackage:     protoPkg,
			GoPackagePath: fmt.Sprintf("%s/%s", sourceFile.ImportPath(), protoPkg),
			Package:       "service",
		}

		for _, iface := range ex.ModuleMap[sourceFile.Module].Packages[sourceFile.Package].Interfaces {
			protoService := transform.Interface2ProtoService(*iface)
			protoFile.Services = append(protoFile.Services, protoService)
			for _, rpc := range protoService.Fields {
				protoFile.Messages = append(protoFile.Messages, rpc.Request, rpc.Response)
			}
		}

		for _, module := range ex.ModuleMap {
			for _, p := range module.Packages {
				for _, s := range p.Structs {
					protoFile.Messages = append(protoFile.Messages, transform.Struct2ProtoMessage(*s))
					logger.Println(s.Name)
					protoFile.Messages = append(protoFile.Messages, types.ProtoMessage{
						Name:   s.Name,
						Fields: transform.Fields2ProtoFields(s.Fields),
					})
				}
			}
		}

		unit := generator.NewUnit(
			sourceFile, tmpl, protoFile, nil, nil,
			filepath.Join(
				filepath.Join(filepath.Dir(args[0]), "proto"),
				filepath.Base(strings.Replace(args[0], ".go", ".proto", 1)),
			), writer.File,
		)
		err = unit.Generate()
		if err != nil {
			logger.Fatal(err)
		}
	},
}

func stripSliceAndPointer(s string) string {
	return strings.Replace(strings.Replace(s, "[]", "", 1), "*", "", 1)
}

func init() {
	rootCmd.AddCommand(protoCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// protoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// protoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
