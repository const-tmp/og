package cmd

import (
	"github.com/nullc4t/og/pkg/editor"
	"github.com/nullc4t/og/pkg/extract"
	"github.com/nullc4t/og/pkg/generator"
	"github.com/nullc4t/og/pkg/names"
	"github.com/nullc4t/og/pkg/templates"
	"github.com/nullc4t/og/pkg/utils"
	"github.com/nullc4t/og/pkg/writer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go/ast"
	"path/filepath"
)

// crudCmd represents the crud command
var crudCmd = &cobra.Command{
	Use:     "crud -f file.go [-f file.go]... output-dir",
	Aliases: []string{"c", "cr"},
	Short:   "Implements DB CRUD interface",
	Long: `Generates:
1. CRUD Interface and impl
2. Repo to be edited by user
3. General DB Repo with shorthands for every type Repo

Example:
cd internal
og gen types/crud -f types/typeA.go -f typeB.go .

generates:

internal
├── a
│	├── crud.og.go
│	└── repo.go
├── b
│	├── crud.og.go
│	└── repo.go
├── repo
│	└── repo.go
└── types
    ├── typeA.go
    └── typeB.go
`,
	//Args:    cobra.ExactArgs(1),
	Example: "og gen crud types.go models/",
	Run: func(cmd *cobra.Command, args []string) {
		crudTmpl, err := templates.NewCRUD()
		if err != nil {
			logger.Fatal(err)
		}
		repoTmpl, err := templates.NewRepo()
		if err != nil {
			logger.Fatal(err)
		}
		genRepoTmpl, err := templates.NewGeneralRepo()
		if err != nil {
			logger.Fatal(err)
		}

		var repos []map[string]any
		var imports []string

		logger.Println("files", viper.GetStringSlice("files"))

		for _, s := range viper.GetStringSlice("files") {
			logger.Println("file", s)
			src, err := extract.GoFile(s)
			if err != nil {
				logger.Fatal(err)
			}
			ast.Inspect(src.AST, func(node ast.Node) bool {
				switch typeSpec := node.(type) {
				case *ast.TypeSpec:
					v, ok := typeSpec.Type.(*ast.StructType)
					dir := filepath.Join(args[0], "repo", names.PackageNameFromType(typeSpec.Name.Name))
					if !ok {
						return false
					}
					if len(v.Fields.List) == 0 {
						return false
					}
					if v.Fields.List[0].Names != nil {
						return false
					}
					sel, ok := v.Fields.List[0].Type.(*ast.SelectorExpr)
					if !ok {
						return false
					}
					ident, ok := sel.X.(*ast.Ident)
					if !ok {
						return false
					}

					if ident.Name == "crud" && sel.Sel.Name == "Model" {
						dot := map[string]any{
							"Package": names.PackageNameFromType(typeSpec.Name.Name),
							"Type":    names.TypeNameWithPackage(src.Package, typeSpec.Name.Name),
						}

						logger.Println("creating CRUD for", typeSpec.Name.Name)

						crudPath := filepath.Join(dir, "crud.og.go")
						crudUnit := generator.New(src, crudTmpl, dot, writer.File, crudPath)
						err = crudUnit.Generate()
						if err != nil {
							logger.Fatal("generate crud error:", err)
						}

						repoPath := filepath.Join(dir, "repo.go")
						ok, err = utils.Exists(repoPath)
						if err != nil {
							logger.Fatal("check exists", repoPath, "error:", err)
						}
						if !ok || viper.GetBool("regen") {
							logger.Println("creating Repo for", typeSpec.Name.Name)

							repoUnit := generator.New(src, repoTmpl, dot, writer.File, repoPath)
							err = repoUnit.Generate()
							if err != nil {
								logger.Fatal("generate repo error:", err)
							}
						} else {
							logger.Println(repoPath, "already exists, skipping")
						}

						repos = append(repos, map[string]any{
							"Method":  typeSpec.Name.Name,
							"Package": dot["Package"],
							"Type":    "Repo",
						})

						crudFile, err := extract.GoFile(crudPath)
						if err != nil {
							logger.Fatal("crud file parse error:", err)
						}
						imports = append(imports, crudFile.ImportPath())
						return false
					}
				}
				return true
			})
		}

		genRepoPath := filepath.Join(args[0], "repo", "repo.go")
		ok, err := utils.Exists(genRepoPath)
		if err != nil {
			logger.Fatal("check exists", genRepoPath, "error:", err)
		}
		if !ok || viper.GetBool("regen") {
			logger.Println("creating gen Repo")

			genRepoUnit := generator.NewUnit(
				nil,
				genRepoTmpl,
				map[string]any{
					"PackageName": "repo",
					"Repos":       repos,
				}, []editor.CodeEditor{
					editor.AddImportsFactory(imports...),
					generator.Formatter,
				}, nil, genRepoPath,
				writer.File,
			)
			err = genRepoUnit.Generate()
			if err != nil {
				logger.Fatal("generate general repo error:", err)
			}
		} else {
			logger.Println(genRepoPath, "already exists, skipping")
		}
	},
}

func init() {
	genCmd.AddCommand(crudCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// crudCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// crudCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	//crudCmd.Flags().StringArrayP("file", "f", nil, "files to be parsed; might be provided multiple times; example: -f file.go")
	//_ = viper.BindPFlag("files", crudCmd.Flag("file"))

	crudCmd.Flags().BoolP("regen", "r", false, "regenerate existing files")
	_ = viper.BindPFlag("regen", crudCmd.Flag("regen"))
}
