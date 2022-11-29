package transform

import (
	"fmt"
	"regexp"
	"strings"
)

var mapRegex = regexp.MustCompile(`map\[(.+)](.+)`)

func Go2ProtobufType(s string) string {
	if strings.Contains(s, "map") {
		sm := mapRegex.FindStringSubmatch(s)
		return fmt.Sprintf("map<%s, %s>", Go2ProtobufType(sm[1]), Go2ProtobufType(sm[2]))
	}

	var prefix, protoType string

	s = strings.Replace(s, "*", "", 1)

	if strings.Contains(s, "[]") {
		s = strings.Replace(s, "[]", "", 1)
		prefix = "repeated "
	}

	switch s {
	case "error":
		protoType = "string"
	case "int":
		protoType = "int32"
	case "int32":
		protoType = "int32"
	case "uint":
		protoType = "uint32"
	case "uint32":
		protoType = "uint32"
	case "uint64":
		protoType = "uint64"
	case "string":
		protoType = "string"
	case "float32":
		protoType = "float32"
	case "float64":
		protoType = "float64"
	case "any":
		protoType = "google.protobuf.Any"
	case "interface{}":
		protoType = "google.protobuf.Any"
	default:
		if strings.Contains(s, ".") {
			fmt.Println(s)
			protoType = strings.Split(s, ".")[1]
		}
	}
	return prefix + protoType
}
