package extract

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//go:generate og e doc file.go -e -F -t -f file.go

var (
	moduleRegex = regexp.MustCompile("module (.+)")
)

// ModuleNameFromGoMod read go.mod from path, returns module name or error
func ModuleNameFromGoMod(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	stringSubmatch := moduleRegex.FindStringSubmatch(string(b))
	if len(stringSubmatch) == 0 {
		return "", fmt.Errorf("no module pattern matched if %s", path)
	}
	return stringSubmatch[1], nil
}

// SearchFileUp looks for file in heightLimit directories up
func SearchFileUp(file, dir string, heightLimit int) (string, error) {
	if heightLimit == 0 {
		return "", fmt.Errorf("file %s not found", file)
	}
	info, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, entry := range info {
		if entry.IsDir() {
			continue
		}
		if file == entry.Name() {
			return filepath.Join(dir, file), nil
		}
	}
	return SearchFileUp(file, filepath.Join(dir, ".."), heightLimit-1)
}

// SearchGoModUp is SearchFileUp for go.mod
func SearchGoModUp(dir string, heightLimit int) (string, error) {
	return SearchFileUp("go.mod", dir, heightLimit)
}

// SearchFile looks for file in current directory
func SearchFile(file string) (string, error) {
	files, err := filepath.Glob("go.mod")
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", fmt.Errorf("file %s not found", file)
	}
	if len(files) > 1 {
		return "", fmt.Errorf("multiple files found:\n%s", strings.Join(files, "\n"))
	}
	return files[0], nil
}

//// ImportedTypes exported func TODO: edit
//func ImportedTypes(f *types.File) map[types.Type]struct{} {
//	imported := make(map[types.Type]struct{})
//	if f == nil {
//		return imported
//	}
//	for _, i := range f.Interfaces {
//		for _, method := range i.Methods {
//			for j, arg := range method.Args {
//				log.Println(j, arg.Name, arg.Type)
//				if strings.Contains(arg.Type.String(), ".") {
//					imported[arg.Type] = struct{}{}
//				}
//			}
//			for j, arg := range method.Results {
//				log.Println(j, arg.Name, arg.Type)
//				if strings.Contains(arg.Type.String(), ".") {
//					imported[arg.Type] = struct{}{}
//				}
//			}
//		}
//	}
//	return imported
//}

//// ImportPathForPackage exported func TODO: edit
//func ImportPathForPackage(p string, f *types.File) string {
//	for _, i := range f.Imports {
//		if i.Name == p {
//			return i.Package
//		}
//		s := strings.Split(i.Package, "/")
//		if s[len(s)-1] == p {
//			return i.Package
//		}
//	}
//	return ""
//}
