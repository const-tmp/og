package cmd

import (
	"fmt"
	"github.com/nullc4t/og/internal/types"
	"github.com/nullc4t/og/pkg/editor"
	"github.com/nullc4t/og/pkg/extract"
	"github.com/nullc4t/og/pkg/generator"
	"github.com/nullc4t/og/pkg/names"
	"github.com/nullc4t/og/pkg/templates"
	"github.com/nullc4t/og/pkg/transform"
	"github.com/nullc4t/og/pkg/writer"
	"github.com/spf13/viper"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
)

// protocolCmd represents the protocol command
var protocolCmd = &cobra.Command{
	Use:   "protocol [go file with interface(s)]",
	Short: "Create request & response types",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("protocol called")
		logger.Println("files:", viper.GetStringSlice("files"))

		tmpl := template.Must(template.New("").Funcs(templates.FuncMap).Parse(templates.TransportExchanges))
		endpointTmpl := template.Must(tmpl.New("").Funcs(templates.FuncMap).Parse(templates.Endpoints))
		endpointSetTmpl := template.Must(tmpl.New("").Funcs(templates.FuncMap).Parse(templates.EndpointSet))

		//fset := token.NewFileSet()
		//file, err := parser.ParseFile(fset, args[0], nil, parser.ParseComments)
		file, err := extract.GoFile(args[0])
		if err != nil {
			logger.Fatal(err)
		}

		ifaces := extract.InterfacesFromASTFile(file)

		// for each interface
		for _, iface := range ifaces {
			transform.NameEmptyArgsInInterface(&iface)
			exchangeStructs := transform.Interface2ExchangeStructs(iface)

			logger.Println(iface.Name, "used imports:")
			for _, imp := range iface.Dependencies {
				logger.Println(iface.Name, imp.Name, imp.Path)
			}

			// name/rename args
			for _, exchangeStruct := range exchangeStructs {
				exchangeStruct = transform.RenameExchangeStruct(exchangeStruct)
			}

			sf, err := extract.GoFile(args[0])
			if err != nil {
				logger.Fatal(err)
			}

			// generate endpoints
			endpointSetUnit := generator.NewUnit(nil, endpointSetTmpl, map[string]any{
				"Package":        "endpoints",
				"Interface":      iface,
				"ServicePackage": sf.Package,
			}, nil,
				[]editor.ASTEditor{
					editor.ASTImportsFactory(types.Import{Path: sf.ImportPath()}),
				}, filepath.Join(
					filepath.Dir(args[0]), "endpoints",
					names.FileNameWithSuffix(iface.Name, "endpoints"),
				), writer.File)
			err = endpointSetUnit.Generate()
			if err != nil {
				logger.Fatal("generate protocol error:", err)
			}

			// generate server endpoints
			endpointUnit := generator.NewUnit(nil, endpointTmpl, map[string]any{
				"Package":        "endpoints",
				"Interface":      iface,
				"ServicePackage": sf.Package,
			}, nil,
				[]editor.ASTEditor{
					editor.ASTImportsFactory(types.Import{Path: sf.ImportPath()}),
				}, filepath.Join(
					filepath.Dir(args[0]), "endpoints",
					names.FileNameWithSuffix(iface.Name, "server"),
				), writer.File)
			err = endpointUnit.Generate()
			if err != nil {
				logger.Fatal("generate protocol error:", err)
			}

			// generate exchanges
			unit := generator.NewUnit(nil, tmpl, map[string]any{
				"Package": "endpoints",
				"Structs": exchangeStructs,
			}, []editor.CodeEditor{
				//editor.AddNamedImportsFactory(iface.Dependencies...),
			},
				[]editor.ASTEditor{
					editor.ASTImportsFactory(iface.Dependencies...),
				}, filepath.Join(
					filepath.Dir(args[0]), "endpoints",
					names.FileNameWithSuffix(iface.Name, "exchanges"),
				), writer.File)
			err = unit.Generate()
			if err != nil {
				logger.Fatal("generate protocol error:", err)
			}
		}
	},
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
