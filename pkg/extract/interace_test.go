package extract

import (
	"github.com/stretchr/testify/require"
	"go/parser"
	"go/token"
	"testing"
)

var testFiles = []string{
	"/Users/hightime/code/og/test/simple.go",
	"/Users/hightime/code/kk/core/internal/ledger/repo/repo.go",
}

func TestExtractInterfaces(t *testing.T) {
	for _, s := range testFiles {
		//t.Log(i, s)
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, s, nil, parser.ParseComments)
		require.NoError(t, err)
		ifaces := Interfaces(file)
		for j, iface := range ifaces {
			t.Log(j, "interface:", iface.Name)
			t.Log("\tmethods:")
			for k, method := range iface.Methods {
				t.Log(k, method)
			}
		}
		t.Log()
		for _, iface := range ifaces {
			for _, method := range iface.Methods {
				for _, arg := range method.Args {
					t.Log(arg.Type.Name, arg.Type.Package, arg.Type.IsImported(), arg.Type.ImportPath)
				}
			}
		}
	}
}

func TestFindImport(t *testing.T) {
	for _, s := range testFiles {
		//t.Log(i, s)
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, s, nil, parser.ParseComments)
		require.NoError(t, err)
		require.NotEmpty(t, ImportByPackage(file, "extract"))
	}
}

//type testCase struct {
//	srcPath  string
//	expected []Interface
//}
//
//var testCases = []testCase{
//	{
//		"/Users/hightime/code/og/test/simple.go",
//		[]Interface{
//			{"Simple", []Method{
//				{Name: "Get", Args: []Arg{{Name: "i", Type: "int"}}, Results: []Arg{{Name: "err", Type: "error"}}},
//			}},
//		},
//	},
//	{
//		"/Users/hightime/code/kk/core/internal/ledger/repo/repo.go",
//		[]Interface{
//			{"Repo", []Method{
//				{Name: "Balance", Args: nil, Results: []Arg{{Type: "balance.Repo"}}},
//				{Name: "Transaction", Args: nil, Results: []Arg{{Type: "transaction.Repo"}}},
//				{Name: "Crypto", Args: nil, Results: []Arg{{Type: "crypto.Repo"}}},
//				{Name: "Card", Args: nil, Results: []Arg{{Type: "card.Repo"}}},
//			}},
//		},
//	},
//}
//
//func TestExtractInterfaces(t *testing.T) {
//	for _, tc := range testCases {
//		fset := token.NewFileSet()
//		file, err := parser.ParseFile(fset, tc.srcPath, nil, parser.ParseComments)
//		require.NoError(t, err)
//
//		ifaces := Interfaces(file)
//
//		require.Equal(t, tc.expected, ifaces)
//	}
//}
