package names

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	regexIsLower = regexp.MustCompile("[a-z]")
	regexIsUpper = regexp.MustCompile("[A-Z]")
)

func PackageNameFromType(typeName string) string {
	return strings.ToLower(typeName)
}

func TypeNameWithPackage(packageName, typeName string) string {
	return fmt.Sprintf("%s.%s", packageName, typeName)
}

func getRawAbbr(s string) string {
	var res string
	for _, word := range SplitCamelCase(s) {
		if len(word) > 0 {
			res += string(word[0])
		}
	}
	return res
}

func GetLowerAbbr(s string) string {
	return strings.ToLower(getRawAbbr(s))
}

func GetUpperAbbr(s string) string {
	return strings.ToUpper(getRawAbbr(s))
}
