package templates

import (
	"errors"
	"fmt"
	types2 "github.com/nullc4t/og/internal/types"
	"github.com/nullc4t/og/pkg/names"
	"github.com/nullc4t/og/pkg/transform"
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
		"plus":                      Plus,
		"camel2snake":               names.Camel2Snake,
		"pbtype":                    transform.Go2ProtobufType,
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

func ReturnAllStructFields(args types2.Args) string {
	var s []string
	for _, arg := range args {
		s = append(s, fmt.Sprintf("r.%s", arg.Name))
	}
	return strings.Join(s, ", ")
}

func ReturnAllStructFieldTypes(args types2.Args) string {
	var s []string
	for _, arg := range args {
		s = append(s, arg.Type.String())
	}
	return strings.Join(s, ", ")
}

func StructConstructorArgs(args types2.Args) string {
	var s []string
	for _, arg := range args {
		s = append(s, fmt.Sprintf("%s %s", names.GetUnexportedName(arg.Name), arg.Type.String()))
	}
	return strings.Join(s, ", ")
}

func StructConstructorReturn(args types2.Args) string {
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

func Plus(i, n int) int {
	return i + n
}

//func ToProtobufType(s string) (string, error) {
//	var prefix string
//
//	s = strings.Replace(s, "*", "", 1)
//	if strings.Contains(s, "[]") {
//		s = strings.Replace(s, "[]", "", 1)
//		prefix = "repeated "
//	}
//
//	switch s {
//	case "error":
//		return prefix + "string", nil
//	case "int":
//		return prefix + "int32", nil
//	case "int32":
//		return prefix + "int32", nil
//	case "uint":
//		return prefix + "uint32", nil
//	case "uint32":
//		return prefix + "uint32", nil
//	case "uint64":
//		return prefix + "uint64", nil
//	case "string":
//		return prefix + "string", nil
//	case "float32":
//		return prefix + "float32", nil
//	case "float64":
//		return prefix + "float64", nil
//	default:
//		if strings.Contains(s, ".") {
//			fmt.Println(s)
//			return prefix + strings.Split(s, ".")[1], nil
//		}
//		return "", fmt.Errorf("unknown type: %s", s)
//	}
//}
