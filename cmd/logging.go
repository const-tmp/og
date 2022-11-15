/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/nullc4t/gensta/pkg/inspector"
	"github.com/nullc4t/gensta/pkg/parser"
	"github.com/nullc4t/gensta/pkg/templates"
	"github.com/spf13/cobra"
	"github.com/vetcher/go-astra/types"
	"go/ast"
	"go/format"
	astparser "go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"os"
	"path/filepath"
	"strings"
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
	Example: "gensta gen logging service.go middleware/logging.go",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		srcFile, err := parser.NewAstra(args[0])
		if err != nil {
			logger.Fatal(err)
		}

		//tmpl := templates.NewRoot()
		//tmpl, err = tmpl.ParseGlob("templates/*.tmpl")
		tmpl, err := templates.NewRoot()
		if err != nil {
			logger.Fatal(err)
		}

		tmpl, err = tmpl.ParseFiles("templates/logging_middleware.tmpl")
		if err != nil {
			logger.Fatal(err)
		}

		tmp := new(bytes.Buffer)
		//err = tmpl.ExecuteTemplate(tmp, "logmw.tmpl", srcFile)
		err = tmpl.ExecuteTemplate(tmp, "logging_middleware.tmpl", srcFile)
		if err != nil {
			logger.Fatal(err)
		}

		fmt.Println(string(tmp.Bytes()))

		fset := token.NewFileSet()
		file, err := astparser.ParseFile(fset, "", tmp.Bytes(), 0)
		if err != nil {
			logger.Fatal(err)
		}

		ok := astutil.AddImport(fset, file, srcFile.ImportPath())
		if !ok {
			logger.Fatal("not ok")
		}

		for t, _ := range inspector.GetImportedTypes(srcFile.Astra) {
			p := inspector.ExtractPackageFromType(t)
			if importPath := inspector.GetImportPathForPackage(p, srcFile.Astra); importPath != "" {
				astutil.AddImport(fset, file, importPath)
			}
		}

		ast.SortImports(fset, file)

		tmp = new(bytes.Buffer)
		err = printer.Fprint(tmp, fset, file)
		if err != nil {
			logger.Fatal(err)
		}

		formatted, err := format.Source(tmp.Bytes())
		if err != nil {
			logger.Fatal(err)
		}

		f, err := os.OpenFile(args[1], os.O_WRONLY|os.O_CREATE, 0644)
		if os.IsNotExist(err) {
			err = os.MkdirAll(filepath.Dir(args[1]), 0755)
			if err != nil {
				logger.Fatal(err)
			}

			f, err = os.Create(args[1])
			if err != nil {
				logger.Fatal(err)
			}

		}
		if err != nil {
			logger.Fatal(err)
		}
		defer f.Close()

		_, err = f.Write(formatted)
		if err != nil {
			logger.Fatal(err)
		}

		fmt.Println("Done")
	},
}

func errorHandler(err error) {
	if err != nil {
		logger.Fatal(err)
	}
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

func lower1(s string) string { return strings.ToLower(s[:1]) + s[1:] }

func receiver(s string) string { return fmt.Sprintf("%s %s", s[:1], s) }

func dict(args ...interface{}) (map[string]interface{}, error) {
	if len(args)%2 != 0 {
		return nil, errors.New("dict: must be an even number of arguments")
	}
	m := make(map[string]interface{})
	for i := 0; i < len(args); i += 2 {
		s, ok := args[i].(string)
		if !ok {
			return nil, fmt.Errorf("%v key must be string but got %T", args[i], args[i])
		}
		m[s] = args[i+1]
	}
	return m, nil
}

func renderArgs(args []types.Variable) string {
	var s []string
	for _, a := range args {
		s = append(s, fmt.Sprintf("%s %s", a.Name, a.Type))
	}
	return strings.Join(s, ", ")
}

func argNames(args []types.Variable) []string {
	var res []string
	for _, arg := range args {
		res = append(res, arg.Name)
	}
	return res
}

func argsSting(args []types.Variable) []string {
	var res []string
	for _, arg := range args {
		res = append(res, fmt.Sprintf("%s %s", arg.Name, arg.Type))
	}
	return res
}

func appendFormatter(ss []string) []string {
	for i, s := range ss {
		ss[i] = fmt.Sprintf("%s:\t%%v", s)
	}
	return ss
}
