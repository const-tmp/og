/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/nullc4t/og/internal/extractor"
	"github.com/nullc4t/og/internal/types"
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
	Use:   "proto",
	Short: "generate .proto file",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("proto called")

		ifaceFile, err := extractor.GoFile(viper.GetString("interfaces_file"))
		if err != nil {
			logger.Fatal(err)
		}

		//exchFile, err := extractor.GoFile(viper.GetString("exchanges_file"))
		//if err != nil {
		//	logger.Fatal(err)
		//}

		ex := extractor.NewExtractor()
		ii, is, err := ex.ParseFile(viper.GetString("interfaces_file"), "", 0)
		if err != nil {
			logger.Fatal(err)
		}
		ei, es, err := ex.ParseFile(viper.GetString("exchanges_file"), "", 1)
		if err != nil {
			logger.Fatal(err)
		}

		tmpl := template.Must(template.New("").Funcs(templates.FuncMap).Parse(templates.Proto2))

		protoPkg := "proto"
		var protoFile = types.ProtoFile{
			GoPackage:     protoPkg,
			GoPackagePath: fmt.Sprintf("%s/%s", ifaceFile.ImportPath(), protoPkg),
			Package:       "service",
		}

		protoImports := make(map[string]struct{})

		iMap := make(map[string]*types.Interface)
		for _, iface := range ii {
			if v, ok := iMap[iface.Name]; !ok {
				iMap[iface.Name] = iface
			} else {
				logger.Println("name conflict:", iface.Name, v.Name)
			}
		}

		sMap := make(map[string]*types.Struct)
		for _, s := range append(is, es...) {
			if _, ok := sMap[s.Name]; !ok {
				sMap[s.Name] = s
			}
			//} else {
			//	logger.Println("name conflict:", s.Name, v.Name)
			//}
		}

		em := make(map[string]struct{})
		for _, iface := range ei {
			if _, ok := em[iface.Name]; ok {
				continue
			}
			protoFile.Messages = append(protoFile.Messages, types.ProtoMessage{
				Name: iface.Name,
				Fields: []types.ProtoField{
					{
						Name:  names.Camel2Snake(iface.Name),
						OneOf: true,
					},
				},
			})
			em[iface.Name] = struct{}{}
		}

		for _, iface := range iMap {
			protoService := transform.Interface2ProtoService(*iface)
			protoFile.Services = append(protoFile.Services, protoService)
		}

		for _, str := range sMap {
			msg := transform.Struct2ProtoMessage(*str)
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
