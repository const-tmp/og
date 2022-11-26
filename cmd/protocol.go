/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/nullc4t/og/pkg/editor"
	"github.com/nullc4t/og/pkg/extract"
	"github.com/nullc4t/og/pkg/generator"
	"github.com/nullc4t/og/pkg/names"
	"github.com/nullc4t/og/pkg/source"
	"github.com/nullc4t/og/pkg/templates"
	"github.com/nullc4t/og/pkg/writer"
	"github.com/spf13/viper"
	"go/parser"
	"go/token"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
)

type ProtocolStruct struct {
	StructName string
	Fields     extract.Args
}

// protocolCmd represents the protocol command
var protocolCmd = &cobra.Command{
	Use:   "protocol",
	Short: "Create request & response types",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("protocol called")
		logger.Println("files:", viper.GetStringSlice("files"))

		//tmpl := template.Must(template.New("struct").
		//	Funcs(templates.FuncMap).
		//	Parse(templates.StructTemplate))
		tmpl := template.Must(template.New("").Funcs(templates.FuncMap).Parse(templates.TransportExchanges))
		endpointTmpl := template.Must(tmpl.New("").Funcs(templates.FuncMap).Parse(templates.Endpoints))
		endpointSetTmpl := template.Must(tmpl.New("").Funcs(templates.FuncMap).Parse(templates.EndpointSet))

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, args[0], nil, parser.ParseComments)
		if err != nil {
			logger.Fatal(err)
		}

		ifaces := extract.Interfaces(file)

		// for each interface
		for _, iface := range ifaces {
			// we create structs
			var sds []ProtocolStruct
			for _, method := range iface.Methods {
				// for request
				requestStruct := ProtocolStruct{StructName: fmt.Sprintf("%sRequest", method.Name)}
				for _, arg := range method.Args {
					//if !ArgIsContext(*arg) {
					requestStruct.Fields = append(requestStruct.Fields, arg)
					//}
				}
				sds = append(sds, requestStruct)

				// for response
				responseStruct := ProtocolStruct{StructName: fmt.Sprintf("%sResponse", method.Name)}
				for _, arg := range method.Results.Args {
					responseStruct.Fields = append(responseStruct.Fields, arg)
					logger.Println(iface.Name, method.Name, arg)
				}
				sds = append(sds, responseStruct)
			}

			logger.Println(iface.Name, "used imports:")
			for _, imp := range iface.UsedImports {
				logger.Println(iface.Name, imp.Name, imp.Path)
			}

			// name/rename args
			for _, sd := range sds {
				for _, field := range sd.Fields {
					if field.Name == "" {
						switch field.Type.Name {
						case "error":
							field.Name = "Err"
						case "Context":
							field.Name = "Ctx"
						default:
							field.Name = names.GetExportedName(field.Type.Name)
						}
					} else {
						field.Name = names.GetExportedName(field.Name)
					}
				}
			}

			sf, err := source.NewFile(args[0])
			if err != nil {
				logger.Fatal(err)
			}

			// generate endpoints
			endpointSetUnit := generator.NewUnit(nil, endpointSetTmpl, map[string]any{
				"Package":        "transport",
				"Interface":      iface,
				"ServicePackage": sf.Package,
			}, nil,
				[]editor.ASTEditor{
					editor.ASTImportsFactory(extract.Import{Path: sf.ImportPath()}),
				}, filepath.Join(
					filepath.Dir(args[0]), "transport",
					names.FileNameWithSuffix(iface.Name, "endpoints"),
				), writer.File)
			err = endpointSetUnit.Generate()
			if err != nil {
				logger.Fatal("generate protocol error:", err)
			}

			// generate server endpoints
			endpointUnit := generator.NewUnit(nil, endpointTmpl, map[string]any{
				"Package":        "transport",
				"Interface":      iface,
				"ServicePackage": sf.Package,
			}, nil,
				[]editor.ASTEditor{
					editor.ASTImportsFactory(extract.Import{Path: sf.ImportPath()}),
				}, filepath.Join(
					filepath.Dir(args[0]), "transport",
					names.FileNameWithSuffix(iface.Name, "server"),
				), writer.File)
			err = endpointUnit.Generate()
			if err != nil {
				logger.Fatal("generate protocol error:", err)
			}

			// generate exchanges
			unit := generator.NewUnit(nil, tmpl, map[string]any{
				//"Package": file.Name.Name,
				"Package": "transport",
				"Structs": sds,
			}, []editor.CodeEditor{
				//editor.AddNamedImportsFactory(iface.UsedImports...),
			},
				[]editor.ASTEditor{
					editor.ASTImportsFactory(iface.UsedImports...),
				}, filepath.Join(
					filepath.Dir(args[0]), "transport",
					names.FileNameWithSuffix(iface.Name, "exchanges"),
				), writer.File)
			err = unit.Generate()
			if err != nil {
				logger.Fatal("generate protocol error:", err)
			}
		}
	},
}

func ArgIsContext(arg extract.Arg) bool {
	return arg.Type.String() == "context.Context"
}

func ArgIsError(arg extract.Arg) bool {
	return arg.Type.String() == "error"
}

func init() {
	rootCmd.AddCommand(protocolCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// protocolCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// protocolCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
