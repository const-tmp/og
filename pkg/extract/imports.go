package extract

import (
	"go/ast"
	"strings"
)

func ImportString(file *ast.File, pkg string) string {
	for _, spec := range file.Imports {
		pathString := strings.Replace(spec.Path.Value, "\"", "", -1)

		if spec.Name != nil {
			if spec.Name.Name == pkg {
				return pathString
			}
		} else {
			path := strings.Split(pathString, "/")
			if len(path) == 1 {
				if pathString == pkg {
					return pathString
				}
			} else {
				if path[len(path)-1] == pkg {
					return pathString
				}
			}
		}
	}

	return ""
}
