package cmd

import (
	"bytes"
	"fmt"
	"github.com/nullc4t/og/pkg/editor"
	"github.com/nullc4t/og/pkg/writer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
)

// docsCmd represents the docs command
var docsCmd = &cobra.Command{
	Use:     "docs -f file.go [flags]",
	Aliases: []string{"doc", "d"},
	Short:   "Generates docstrings",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Println("docs called")
		logger.Println(viper.GetStringSlice("files"))
		for _, s := range viper.GetStringSlice("files") {
			logger.Println("file:", s)

			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, s, nil, parser.ParseComments)
			if err != nil {
				logger.Fatal(err)
			}

			var comments []*ast.CommentGroup

			// TODO move to editor
			// TODO astutil.Apply
			ast.Inspect(file, func(n ast.Node) bool {
				c, isType := n.(*ast.CommentGroup)
				if isType {
					comments = append(comments, c)
					return true
				}

				var name string

				valueSpec, isVar := n.(*ast.ValueSpec)
				typeSpec, isType := n.(*ast.TypeSpec)
				funcDecl, isFunc := n.(*ast.FuncDecl)

				if isType && typeSpec.Doc == nil && (viper.GetBool("types") || viper.GetBool("all")) {
					name = typeSpec.Name.Name
					if typeSpec.Name.IsExported() && (viper.GetBool("exported") || viper.GetBool("all")) {
						typeSpec.Doc = editor.Comment4Node(n, fmt.Sprintf("// %s exported type TODO: edit", name))
					}
					if !typeSpec.Name.IsExported() && (viper.GetBool("unexported") || viper.GetBool("all")) {
						typeSpec.Doc = editor.Comment4Node(n, fmt.Sprintf("// %s unexported type TODO: edit", name))
					}
				}
				if isVar && valueSpec.Doc == nil && (viper.GetBool("vars") || viper.GetBool("all")) {
					name = valueSpec.Names[0].Name
					if valueSpec.Names[0].IsExported() && (viper.GetBool("exported") || viper.GetBool("all")) {
						valueSpec.Doc = editor.Comment4Node(n, fmt.Sprintf("// %s exported var TODO: edit", name))
					}
					if !valueSpec.Names[0].IsExported() && (viper.GetBool("unexported") || viper.GetBool("all")) {
						valueSpec.Doc = editor.Comment4Node(n, fmt.Sprintf("// %s unexported var TODO: edit", name))
					}
				}
				if isFunc && funcDecl.Doc == nil && (viper.GetBool("funcs") || viper.GetBool("all")) {
					name = funcDecl.Name.Name
					if funcDecl.Name.IsExported() && (viper.GetBool("exported") || viper.GetBool("all")) {
						funcDecl.Doc = editor.Comment4Node(n, fmt.Sprintf("// %s exported func TODO: edit", name))
					}
					if !funcDecl.Name.IsExported() && (viper.GetBool("unexported") || viper.GetBool("all")) {
						funcDecl.Doc = editor.Comment4Node(n, fmt.Sprintf("// %s unexported func TODO: edit", name))
					}
				}
				return true
			})
			file.Comments = comments

			if viper.GetBool("dry") {
				if err := printer.Fprint(os.Stdout, fset, file); err != nil {
					logger.Fatal(err)
				}
				return
			}

			var name string
			if viper.GetBool("new") {
				name = fmt.Sprintf("%sdocs.%s", s[:len(s)-2], s[len(s)-2:])
			} else {
				name = s
			}

			buf := new(bytes.Buffer)
			if err := printer.Fprint(buf, fset, file); err != nil {
				logger.Fatal(err)
			}

			err = writer.File(name, buf)
			if err != nil {
				logger.Fatal(err)
			}
		}
	},
}

func init() {
	editCmd.AddCommand(docsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// docsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// docsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	//docsCmd.Flags().StringArrayP("file", "f", nil, "files to be edited; might be provided multiple times; example: -f file.go")
	//_ = docsCmd.MarkFlagRequired("file")
	//_ = viper.BindPFlag("files", docsCmd.Flag("file"))

	docsCmd.Flags().BoolP("types", "t", true, "Add comments to types")
	//_ = docsCmd.MarkFlagRequired("types")
	_ = viper.BindPFlag("types", docsCmd.Flag("types"))

	docsCmd.Flags().BoolP("funcs", "F", true, "Add comments to funcs")
	//_ = docsCmd.MarkFlagRequired("funcs")
	_ = viper.BindPFlag("funcs", docsCmd.Flag("funcs"))

	docsCmd.Flags().BoolP("exported", "e", true, "Add comments to exported")
	//_ = docsCmd.MarkFlagRequired("exported")
	_ = viper.BindPFlag("exported", docsCmd.Flag("exported"))

	docsCmd.Flags().BoolP("unexported", "u", false, "Add comments to unexported")
	//_ = docsCmd.MarkFlagRequired("unexported")
	_ = viper.BindPFlag("unexported", docsCmd.Flag("unexported"))

	docsCmd.Flags().BoolP("all", "a", false, "Add comments to all")
	//_ = docsCmd.MarkFlagRequired("all")
	_ = viper.BindPFlag("all", docsCmd.Flag("all"))

	docsCmd.Flags().BoolP("dry", "d", false, "print to stdout")
	_ = viper.BindPFlag("dry", docsCmd.Flag("dry"))

	docsCmd.Flags().BoolP("new", "n", false, "generate new file *.docs.go instead of rewriting original file")
	_ = viper.BindPFlag("new", docsCmd.Flag("new"))
}
