package inspector

import (
	"fmt"
	"github.com/vetcher/go-astra/types"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	moduleRegex       = regexp.MustCompile("module (.+)")
	importedTypeRegex = regexp.MustCompile("([\\w\\d]+)\\.([\\w\\d]+)")
)

func GetModuleNameFromGoMod(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	strings := moduleRegex.FindStringSubmatch(string(b))
	if len(strings) == 0 {
		return "", fmt.Errorf("no module pattern matched if %s", path)
	}
	return strings[1], nil
}

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

func SearchFileDown(file string) (string, error) {
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

func GetImportedTypes(f *types.File) map[types.Type]struct{} {
	imported := make(map[types.Type]struct{})
	if f == nil {
		return imported
	}
	for _, i := range f.Interfaces {
		for _, method := range i.Methods {
			for j, arg := range method.Args {
				log.Println(j, arg.Name, arg.Type)
				if strings.Contains(arg.Type.String(), ".") {
					imported[arg.Type] = struct{}{}
				}
			}
			for j, arg := range method.Results {
				log.Println(j, arg.Name, arg.Type)
				if strings.Contains(arg.Type.String(), ".") {
					imported[arg.Type] = struct{}{}
				}
			}
		}
	}
	return imported
}

func ExtractPackageFromType(p types.Type) string {
	sub := importedTypeRegex.FindStringSubmatch(p.String())
	if len(sub) != 3 {
		fmt.Println("matches")
		for i, s := range sub {
			fmt.Println(i, s)
		}
		panic(fmt.Sprintf("imported type %s not matched", p.String()))
	}
	return sub[1]
}

func GetImportPathForPackage(p string, f *types.File) string {
	for _, i := range f.Imports {
		if i.Name == p {
			return i.Package
		}
		s := strings.Split(i.Package, "/")
		if s[len(s)-1] == p {
			return i.Package
		}
	}
	return ""
}
