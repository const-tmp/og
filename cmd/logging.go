package cmd

import (
	"github.com/nullc4t/og/internal/types"
	"github.com/nullc4t/og/pkg/editor"
	"github.com/nullc4t/og/pkg/extract"
	"github.com/nullc4t/og/pkg/generator"
	"github.com/nullc4t/og/pkg/templates"
	"github.com/nullc4t/og/pkg/transform"
	"github.com/nullc4t/og/pkg/writer"
	"github.com/spf13/cobra"
	"path/filepath"
	"text/template"
)

// loggingCmd represents the logging command
var loggingCmd = &cobra.Command{
	Use:     "logging from to",
	Aliases: []string{"log", "logs", "l"},
	Short:   "Logging middleware for intarface",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Example: "og gen logging service.go middleware/logging.go",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tmpl := template.Must(template.New("").Funcs(templates.FuncMap).Parse(templates.LoggingMiddleware))

		ctx := extract.NewContext()

		ifaces, _, err := extract.ParseFile(ctx, args[0], "", 0)
		if err != nil {
			logger.Fatal(err)
		}

		fp, err := filepath.Abs(args[0])
		if err != nil {
			logger.Fatal(err)
		}

		ifile := ctx.File[fp]

		for _, iface := range ifaces {
			transform.NameEmptyArgsInInterface(iface)
			if iface == nil {
				continue
			}
			epUnit := generator.NewUnit(
				ifile,
				tmpl,
				iface,
				nil,
				[]editor.ASTEditor{editor.ASTImportsFactory(append(iface.Dependencies, types.Import{Path: ifile.ImportPath()})...)},
				filepath.Join(filepath.Join(filepath.Dir(ifile.FilePath), "service"), "service_logging.gen.go"),
				writer.File,
			)
			err = epUnit.Generate()
			if err != nil {
				logger.Fatal(err)
			}
		}
	},
}

func init() {
	genCmd.AddCommand(loggingCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loggingCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loggingCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
