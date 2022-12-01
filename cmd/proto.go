/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/nullc4t/og/internal/extractor"
	"github.com/nullc4t/og/internal/types"
	"github.com/nullc4t/og/pkg/extract"
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

		tmpl := template.Must(template.New("").Funcs(templates.FuncMap).Parse(templates.Proto2))

		ifaces := extract.InterfacesFromASTFile(sourceFile.AST)

		protoPkg := "proto"
		var protoFile = types.ProtoFile{
			GoPackage:     protoPkg,
			GoPackagePath: fmt.Sprintf("%s/%s", sourceFile.ImportPath(), protoPkg),
			Package:       "service",
		}

		for _, iface := range ifaces {
			transform.RenameArgsInInterface(iface)

			logger.Println(iface.Name, "used imports:")
			for _, imp := range iface.Dependencies {
				logger.Println(iface.Name, imp.Name, imp.Path)
			}

			var importedTypes = map[string]types.Type{}

			for _, method := range iface.Methods {
				for _, arg := range method.Args {
					if arg.Type.IsImported() {
						ts := stripSliceAndPointer(arg.Type.String())
						a, ok := importedTypes[ts]
						if ok {
							// doing a "singleton" type
							arg.Type = a
						} else {
							importedTypes[ts] = arg.Type
						}
					}
				}
				for _, arg := range method.Results.Args {
					if arg.Type.IsImported() {
						ts := stripSliceAndPointer(arg.Type.String())
						a, ok := importedTypes[ts]
						if ok {
							// doing a "singleton" type
							arg.Type = a
						}
						importedTypes[ts] = arg.Type
					}
				}
			}

			// for each imported type, find package, file, decl to add them to proto file
			for _, ty := range importedTypes {
				if ty.Name() == "Context" {
					continue
				}
				// get import string for ty
				importString := extract.ImportStringForPackage(sourceFile.AST, ty.Package())

				// get fs path for package
				packagePath, err := extract.SourcePath4Package(sourceFile.Module, sourceFile.ModulePath, importString, sourceFile.FilePath)
				if err != nil {
					logger.Println(err)
					continue
				}

				iface, str, err := extract.ImportedTypeFromPackage(packagePath, ty)
				if err != nil {
					logger.Fatal(err)
				}

				if iface != nil {
					ty.SetIsInterface()
					protoFile.Messages = append(protoFile.Messages, types.ProtoMessage{Name: iface.Name})
				}
				if str != nil {
					protoFile.Messages = append(protoFile.Messages, transform.Struct2ProtoMessage(*str))
				}

				logger.Println(ty.String(), ty.IsInterface())
			}

			protoService := transform.Interface2ProtoService(iface)
			protoFile.Services = append(protoFile.Services, protoService)
		}

		for _, service := range protoFile.Services {
			for _, rpc := range service.Fields {
				protoFile.Messages = append(protoFile.Messages, rpc.Request, rpc.Response)
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
