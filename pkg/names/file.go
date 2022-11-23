package names

import (
	"fmt"
	"strings"
)

func FileNameWithSuffix(name, suffix string) string {
	return fmt.Sprintf("%s.%s.go", strings.ToLower(name), suffix)
}
