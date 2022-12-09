/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/nullc4t/og/internal/types"
	"github.com/nullc4t/og/pkg/extract"
	"github.com/nullc4t/og/pkg/generator"
	"github.com/nullc4t/og/pkg/names"
	"github.com/nullc4t/og/pkg/templates"
	"github.com/nullc4t/og/pkg/transform"
	"github.com/nullc4t/og/pkg/writer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
	"text/template"
)

// protoCmd represents the proto command
var protoCmd = &cobra.Command{
	Use:   "proto -i interfaces.go -e exchanges.go",
	Short: "generate .proto file",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("proto called")

		tmpl := template.Must(template.New("").Funcs(templates.FuncMap).Parse(templates.Proto2))

		ifaceFile, err := extract.GoFile(viper.GetString("interfaces_file"))
		if err != nil {
			logger.Fatal(err)
		}

		services, _ := extract.TypesFromASTFile(ifaceFile)

		protoPkg := "proto"
		var protoFile = types.ProtoFile{
			GoPackage:     protoPkg,
			GoPackagePath: fmt.Sprintf("%s/%s", ifaceFile.ImportPath(), protoPkg),
			Package:       "service",
		}

		for _, iface := range services {
			protoFile.Services = append(protoFile.Services, transform.Interface2ProtoService(*iface))
		}
		for _, service := range protoFile.Services {
			for _, field := range service.Fields {
				protoFile.Messages = append(protoFile.Messages, field.Request, field.Response)
			}
		}

		//exchFile, err := extract.GoFile(viper.GetString("exchanges_file"))
		//if err != nil {
		//	logger.Fatal(err)
		//}

		//_, exchanges := extract.TypesFromASTFile(exchFile)

		ctx := extract.NewContext()
		ifaces, structs, err := extract.ParseFile(ctx, viper.GetString("interfaces_file"), "", 2)
		if err != nil {
			logger.Fatal(err)
		}

		//depIfaces, depStructs, err := extract.TypesRecursive(ctx, ifaceFile, nil, exchanges, 2)

		logger.Println(ctx)

		for _, iface := range ifaces {
			protoFile.Messages = append(protoFile.Messages, types.ProtoMessage{
				Name: iface.Name,
				Fields: []types.ProtoField{
					{
						Name:  names.Camel2Snake(iface.Name),
						OneOf: true,
					},
				},
			})
		}

		protoImports := make(map[string]struct{})

		for _, str := range structs {
			msg := transform.Struct2ProtoMessage(ctx, *str)
			protoFile.Messages = append(protoFile.Messages, msg)
			for _, field := range msg.Fields {
				switch field.Type {
				case "google.protobuf.Timestamp":
					protoImports["google/protobuf/timestamp.proto"] = struct{}{}
				case "google.protobuf.Any":
					protoImports["google/protobuf/any.proto"] = struct{}{}
				}
			}
		}

		for imp, _ := range protoImports {
			protoFile.Imports = append(protoFile.Imports, types.ProtoImport{Path: imp})
		}

		unit := generator.NewUnit(
			ifaceFile, tmpl, protoFile, nil, nil,
			filepath.Join(
				filepath.Join(filepath.Dir(viper.GetString("interfaces_file")), "proto"),
				filepath.Base(strings.Replace(viper.GetString("interfaces_file"), ".go", ".proto", 1)),
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

	protoCmd.Flags().StringP("exchanges_file", "e", "", "go file with *Request/Response")
	_ = protoCmd.MarkFlagRequired("exchanges_file")
	_ = viper.BindPFlag("exchanges_file", protoCmd.Flag("exchanges_file"))

	protoCmd.Flags().StringP("interfaces_file", "i", "", "go file with interface(s)")
	_ = protoCmd.MarkFlagRequired("interfaces_file")
	_ = viper.BindPFlag("interfaces_file", protoCmd.Flag("interfaces_file"))

	protoCmd.Flags().StringSliceP("exclude_types", "x", nil, "exclude types from parsing")
	_ = viper.BindPFlag("exclude_types", protoCmd.Flag("exclude_types"))

}
