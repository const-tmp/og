/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/nullc4t/og/internal/extractor"
	"github.com/spf13/cobra"
)

// extractCmd represents the extract command
var extractCmd = &cobra.Command{
	Use:   "extract",
	Args:  cobra.ExactArgs(1),
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("extract called")
		ex := extractor.NewExtractor()
		_, _, err := ex.ParseFile(args[0], "", 2)
		if err != nil {
			logger.Fatal(err)
		}

		for _, module := range ex.ModuleMap {
			logger.Println(module.Name, module.Path)
			for _, p := range module.Packages {
				logger.Println("\t", p.Name, p.ImportPath, p.Path)
				for s := range p.Files {
					logger.Println("\t\t", s)
				}
				for _, s := range p.Structs {
					logger.Println("\t\t", s.Name, s)
				}
				for _, s := range p.Interfaces {
					logger.Println("\t\t", s.Name, s)
				}
			}
		}

		//for s, typeData := range ex.TypeMap {
		//	fmt.Println(s)
		//	fmt.Println(typeData.Type, typeData.Interface != nil, typeData.Struct != nil)
		//	fmt.Println()
		//	if typeData.Interface != nil || typeData.Struct != nil {
		//	}
		//}

	},
}

func init() {
	rootCmd.AddCommand(extractCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// extractCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// extractCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
