package templates

import (
	"errors"
	"fmt"
	"github.com/nullc4t/og/pkg/extract"
	"github.com/nullc4t/og/pkg/names"
	"github.com/vetcher/go-astra/types"
	"strings"
	"text/template"
)

var (
	FuncMap = template.FuncMap{
		"argNames":                  argNames,
		"args":                      argsSting,
		"funcArgs":                  renderArgs,
		"join":                      strings.Join,
		"appendFormatter":           appendFormatter,
		"lower1":                    lower1,
		"receiver":                  receiver,
		"dict":                      dict,
		"MapDot":                    MapDot,
		"struct_return_args":        ReturnAllStructFields,
		"struct_return_types":       ReturnAllStructFieldTypes,
		"struct_constructor_args":   StructConstructorArgs,
		"struct_constructor_return": StructConstructorReturn,
		"exported":                  names.GetExportedName,
		"unexported":                names.GetUnexportedName,
		"mapslice2slice":            MapSlice2Slice,
	}
)

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

func ReturnAllStructFields(args extract.Args) string {
	var s []string
	for _, arg := range args {
		s = append(s, fmt.Sprintf("r.%s", arg.Name))
	}
	return strings.Join(s, ", ")
}

func ReturnAllStructFieldTypes(args extract.Args) string {
	var s []string
	for _, arg := range args {
		s = append(s, arg.Type.String())
	}
	return strings.Join(s, ", ")
}

func StructConstructorArgs(args extract.Args) string {
	var s []string
	for _, arg := range args {
		s = append(s, fmt.Sprintf("%s %s", names.GetUnexportedName(arg.Name), arg.Type.String()))
	}
	return strings.Join(s, ", ")
}

func StructConstructorReturn(args extract.Args) string {
	var s []string
	for _, arg := range args {
		s = append(s, names.GetUnexportedName(arg.Name))
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
