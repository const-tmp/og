package extract

import (
	"github.com/stretchr/testify/require"
	"go/parser"
	"go/token"
	"testing"
)

type testCase struct {
	srcPath  string
	expected []Interface
}

var testCases = []testCase{
	{
		"/Users/hightime/code/og/test/simple.go",
		[]Interface{
			{
				"Simple",
				[]Method{
					{
						Name: "Get",
						Args: []Arg{
							{
								Name: "i",
								Type: "int",
							},
						},
						Results: []Arg{
							{
								Name: "err",
								Type: "error",
							},
						},
					},
				}},
		},
	},
}

func TestExtractInterfaces(t *testing.T) {
	for _, tc := range testCases {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, tc.srcPath, nil, parser.ParseComments)
		require.NoError(t, err)

		ifaces := Interfaces(file)

		require.Equal(t, tc.expected, ifaces)
	}
}
