package names

import (
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func TestGetExportedName(t *testing.T) {
	var testCases = []struct{ name, expected string }{
		{"testName", "TestName"},
		{"t", "T"},
		{"", ""},
	}
	for _, tc := range testCases {
		require.Equal(t, tc.expected, GetExportedName(tc.name))
	}
}

func TestToSnakeCase(t *testing.T) {
	var testCases = []struct{ name, expected string }{
		{"testName", "test_name"},
		{"str2Bytes", "str2bytes"},
		{"VeryVeryLongVarName", "very_very_long_var_name"},
		{"t", "t"},
		{"T", "t"},
		{"", ""},
	}
	regex := regexp.MustCompile("(?:\\b|[a-z]|\\d)[A-Z]")
	for _, tc := range testCases {
		t.Log("test case:", tc.name)
		t.Log("FindStringSubmatchIndex", regex.FindAllStringIndex(tc.name, -1))
		for i, s := range regex.Split(tc.name, -1) {
			t.Log(i, s)
		}
		t.Log()
	}
}

func TestAbbr(t *testing.T) {
	var testCases = []struct{ name, expected string }{
		{"testACLName", "tAN"},
		{"testName", "tN"},
		{"testName", "tN"},
		{"t", "t"},
		{"", ""},
	}
	for _, tc := range testCases {
		require.Equal(t, tc.expected, getRawAbbr(tc.name))
	}
}
