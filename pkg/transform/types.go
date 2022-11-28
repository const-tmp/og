package transform

import (
	"fmt"
	"strings"
)

func Go2ProtobufType(s string) string {
	var prefix string

	s = strings.Replace(s, "*", "", 1)

	if strings.Contains(s, "[]") {
		s = strings.Replace(s, "[]", "", 1)
		prefix = "repeated "
	}

	switch s {
	case "error":
		return prefix + "string"
	case "int":
		return prefix + "int32"
	case "int32":
		return prefix + "int32"
	case "uint":
		return prefix + "uint32"
	case "uint32":
		return prefix + "uint32"
	case "uint64":
		return prefix + "uint64"
	case "string":
		return prefix + "string"
	case "float32":
		return prefix + "float32"
	case "float64":
		return prefix + "float64"
	default:
		if strings.Contains(s, ".") {
			fmt.Println(s)
			return prefix + strings.Split(s, ".")[1]
		}
		return s
		//return "", fmt.Errorf("unknown type: %s", s)
	}
}
