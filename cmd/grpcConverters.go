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
	"github.com/nullc4t/og/pkg/transform"
	"github.com/nullc4t/og/pkg/utils"
	"github.com/nullc4t/og/pkg/writer"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
)

// grpcConvertersCmd represents the grpcConverters command
var grpcConvertersCmd = &cobra.Command{
	Use:     "grpcConverters exchanges_file.go file.pb.go service_interface.go",
	Aliases: []string{"gc", "grpcconv"},
	Short:   "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("grpcConverters called")

		srvTmpl := template.Must(template.New("").Funcs(templates.FuncMap).Parse(templates.GRPCServer))
		epTmpl := template.Must(template.New("").Funcs(templates.FuncMap).Parse(templates.GRPCEnpointConverters))
		tyTmpl := template.Must(template.New("").Funcs(templates.FuncMap).Parse(templates.GRPCEncoder))

		ctx := extract.NewContext()

		_, exchanges, err := extract.ParseFile(ctx, args[0], "", 2)
		if err != nil {
			logger.Fatal(err)
		}

		_, pbTypes, err := extract.ParseFile(ctx, args[1], "", 0)
		if err != nil {
			logger.Fatal(err)
		}

		ifaces, _, err := extract.ParseFile(ctx, args[2], "", 0)
		if err != nil {
			logger.Fatal(err)
		}

		exchPath, err := filepath.Abs(args[0])
		if err != nil {
			logger.Fatal(err)
		}
		pbPath, err := filepath.Abs(args[1])
		if err != nil {
			logger.Fatal(err)
		}
		exchFile := ctx.File[exchPath]
		pbFile := ctx.File[pbPath]
		//ifaceFile := ctx.File[args[2]]

		logger.Println(ctx)

		var encoders, decoders []transform.Converter

		encoderSliceUtil := utils.NewSlice[transform.Converter](func(a, b transform.Converter) bool {
			return a.StructName == b.StructName && a.IsSlice == b.IsSlice && a.IsPointer == b.IsPointer
		})
		structSliceUtil := utils.NewSlice[*types.Struct](func(t *types.Struct, pb *types.Struct) bool {
			return t.Name == pb.Name
		})

		for _, pbType := range pbTypes {
			if idx := structSliceUtil.Index(exchanges, pbType); idx >= 0 {
				exType := exchanges[idx]
				newEnc := transform.Structs2ProtoEncoder(ctx, exType, pbType)
				encoders = encoderSliceUtil.AppendIfNotExist(encoders, newEnc)
			}
		}

		for _, encoder := range encoders {
			for _, dependency := range encoder.Deps {
				if dependency.IsSlice {
					encoders = encoderSliceUtil.AppendIfNotExist(encoders, transform.Converter{
						StructName: dependency.Type.Name,
						Type:       dependency.Type,
						Proto:      dependency.Proto,
						IsSlice:    dependency.IsSlice,
						IsPointer:  dependency.IsPointer,
					})
				} else {
					encoders = encoderSliceUtil.AppendIfNotExist(encoders, transform.Structs2ProtoEncoder(ctx, &dependency.Type, &dependency.Proto))
				}
			}
		}

		for _, exType := range exchanges {
			if idx := structSliceUtil.Index(pbTypes, exType); idx >= 0 {
				pbType := pbTypes[idx]
				newDec := transform.Structs2ProtoDecoder(ctx, exType, pbType)
				decoders = encoderSliceUtil.AppendIfNotExist(decoders, newDec)
			}
		}

		for _, decoder := range decoders {
			for _, dependency := range decoder.Deps {
				if dependency.IsSlice {
					decoders = encoderSliceUtil.AppendIfNotExist(decoders, transform.Converter{
						StructName: dependency.Type.Name,
						Type:       dependency.Type,
						Proto:      dependency.Proto,
						IsSlice:    dependency.IsSlice,
						IsPointer:  dependency.IsPointer,
					})
				} else {
					decoders = encoderSliceUtil.AppendIfNotExist(decoders, transform.Structs2ProtoDecoder(ctx, &dependency.Type, &dependency.Proto))
				}
			}
		}

		im := utils.NewSet[types.Import]()
		icm := map[struct{ t, p string }]transform.InterfaceConverter{}

		for _, encoder := range encoders {
			im.Add(encoder.Imports.All()...)
			for _, converter := range encoder.InterfaceConverters {
				icm[struct{ t, p string }{t: converter.Type.Name, p: converter.Proto.Name}] = converter
			}
		}
		for _, decoder := range decoders {
			im.Add(decoder.Imports.All()...)
			for _, converter := range decoder.InterfaceConverters {
				icm[struct{ t, p string }{t: converter.Type.Name, p: converter.Proto.Name}] = converter
			}
		}

		epUnit := generator.NewUnit(
			exchFile, epTmpl, map[string]any{
				"Package": "transportgrpc",
				"Exchanges": utils.Filter[*types.Struct](exchanges, func(s *types.Struct) bool {
					return s.ImportPath == exchFile.ImportPath()
				}),
			}, nil,
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
				"Package":             "transportgrpc",
				"Encoders":            encoders,
				"Decoders":            decoders,
				"InterfaceConverters": icm,
			}, nil,
			[]editor.ASTEditor{editor.ASTImportsFactory(append(
				im.All(),
				types.Import{Path: exchFile.ImportPath()},
				types.Import{Path: pbFile.ImportPath()},
			)...)},
			filepath.Join(filepath.Join(filepath.Dir(args[0]), "..", "transport", "grpc"), "type_converters.gen.go"), writer.File,
		)
		err = tyUnit.Generate()
		if err != nil {
			logger.Fatal(err)
		}

		srvUnit := generator.NewUnit(
			exchFile, srvTmpl, map[string]any{
				"Package":     "transportgrpc",
				"ServiceName": ifaces[0].Name,
				"Endpoints":   ifaces[0].Methods,
			}, nil,
			[]editor.ASTEditor{editor.ASTImportsFactory(
				types.Import{Path: exchFile.ImportPath()},
				types.Import{Path: pbFile.ImportPath()},
			)},
			filepath.Join(filepath.Join(filepath.Dir(args[0]), "..", "transport", "grpc"), "server.gen.go"), writer.File,
		)
		err = srvUnit.Generate()
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
