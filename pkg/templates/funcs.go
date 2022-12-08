package templates

import (
	"errors"
	"fmt"
	"github.com/nullc4t/og/internal/types"
	"github.com/nullc4t/og/pkg/names"
	"github.com/nullc4t/og/pkg/transform"
	"strings"
	"text/template"
)

var (
	FuncMap = template.FuncMap{
		"join":                      strings.Join,
		"appendFormatter":           appendFormatter,
		"lower1":                    lower1,
		"receiver":                  receiver,
		"dict":                      dict,
		"MapDot":                    MapDot,
		"struct_args":               StructFields,
		"struct_types":              StructFieldTypes,
		"struct_constructor_args":   StructConstructorArgs,
		"struct_constructor_return": StructConstructorReturn,
		"exported":                  names.GetExportedName,
		"unexported":                names.Unexported,
		"mapslice2slice":            MapSlice2Slice,
		"plus":                      Plus,
		"camel2snake":               names.Camel2Snake,
		"pbtype":                    transform.Go2ProtobufType,
		"jsonTag":                   JSONTag,
		"callArgs":                  CallArgs,
	}
)

func CallArgs(a types.Args) string {
	var s []string
	for _, arg := range a {
		s = append(s, arg.Name)
	}
	return strings.Join(s, ", ")
}

func MapDot(args ...interface{}) (map[string]interface{}, error) {
	if len(args)%2 != 0 {
		return nil, errors.New("MapDot: must be an even number of arguments")
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

func appendFormatter(ss []string) []string {
	for i, s := range ss {
		ss[i] = fmt.Sprintf("%s:\t%%v", s)
	}
	return ss
}

func StructFields(args types.Args) string {
	var s []string
	for _, arg := range args {
		s = append(s, fmt.Sprintf("r.%s", arg.Name))
	}
	return strings.Join(s, ", ")
}

func StructFieldTypes(args types.Args) string {
	var s []string
	for _, arg := range args {
		s = append(s, arg.Type.String())
	}
	return strings.Join(s, ", ")
}

func StructConstructorArgs(args types.Args) string {
	var s []string
	for _, arg := range args {
		s = append(s, fmt.Sprintf("%s %s", names.Unexported(arg.Name), arg.Type.String()))
	}
	return strings.Join(s, ", ")
}

func StructConstructorReturn(args types.Args) string {
	var s []string
	for _, arg := range args {
		s = append(s, names.Unexported(arg.Name))
	}
	return strings.Join(s, ", ")
}

func MapSlice2Slice(sm []map[string]any, key string) []any {
	var res []any
	for _, m := range sm {
		res = append(res, m[key])
	}
	return res
}

func Plus(i, n int) int {
	return i + n
}

func JSONTag(s string) string {
	return fmt.Sprintf("`json:\"%s,omitempty\"`", s)
}
