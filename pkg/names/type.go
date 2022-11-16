package names

import (
	"fmt"
	"strings"
)

func PackageNameFromType(typeName string) string {
	return strings.ToLower(typeName)
}

func TypeNameWithPackage(packageName, typeName string) string {
	return fmt.Sprintf("%s.%s", packageName, typeName)
}
