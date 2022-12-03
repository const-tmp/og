package cmd

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/spf13/cobra"
)

// astCmd represents the ast command
var astCmd = &cobra.Command{
	Use:   "ast [path to go file]",
	Short: "print file's AST",
	Long:  `print file's AST`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ast called")

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, args[0], nil, parser.ParseComments)
		if err != nil {
			logger.Fatal(err)
		}

		err = ast.Print(fset, file)
		if err != nil {
			logger.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(astCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// astCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// astCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
