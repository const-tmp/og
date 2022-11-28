package extract

import (
	"fmt"
	"github.com/nullc4t/og/internal/types"
	"os"
	"path/filepath"
	"regexp"
)

var dependencyRegex = regexp.MustCompile(`\s*(?P<module>.+)\s+(?P<version>v\S+)`)

func DependenciesFromGoMod(data string) ([]types.Dependency, error) {
	var res []types.Dependency

	goPath := os.Getenv("GOPATH")

	for _, s := range dependencyRegex.FindAllStringSubmatch(data, -1) {
		module, version := s[1], s[2]
		res = append(res, types.Dependency{
			Module:  module,
			Version: version,
			Path:    filepath.Join(goPath, "pkg/mod", fmt.Sprintf("%s@%s", module, version)),
		})
	}

	return res, nil
}

func GoMod(path string) ([]byte, error) {
	goModPath, err := SearchGoModUp(path, 5)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(goModPath)
}
