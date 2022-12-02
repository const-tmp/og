package extract

import (
	"fmt"
	"go/ast"
	"strings"
)

func ImportStringForPackage(file *ast.File, pkg string) string {
	for _, spec := range file.Imports {
		pathString := strings.Replace(spec.Path.Value, "\"", "", -1)

		if spec.Name != nil {
			if spec.Name.Name == pkg {
				return pathString
			}
		} else {
			path := strings.Split(pathString, "/")
			switch {
			case len(path) == 1 && pathString == pkg:
				return pathString
			case len(path) > 1 && path[len(path)-1] == pkg:
				return pathString
			case len(path) > 1 && path[len(path)-2] == pkg && strings.HasPrefix(path[len(path)-1], "v"):
				return pathString
			}
		}
	}

	fmt.Println("import string not found; file:", file.Name.Name, "package:", pkg)

	return ""
}
