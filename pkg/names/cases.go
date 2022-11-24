package names

import (
	"fmt"
	"strings"
)

func GetExportedName(name string) string {
	if len(name) == 0 {
		return ""
	}
	return fmt.Sprintf("%s%s", strings.ToUpper(name[:1]), name[1:])
}

func GetUnexportedName(name string) string {
	if len(name) == 0 {
		return ""
	}
	return fmt.Sprintf("%s%s", strings.ToLower(name[:1]), name[1:])
}

//func GetAbbr(name string) string {
//	//var abbr string
//	for i, c := range name {
//
//	}
//}
