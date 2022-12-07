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

func Unexported(name string) string {
	if len(name) == 0 {
		return ""
	}
	return fmt.Sprintf("%s%s", strings.ToLower(name[:1]), name[1:])
}

func SplitCamelCase(s string) []string {
	wordBeg := 0
	var words []string

	for i, c := range s {
		if i+1 == len(s) {
			break
		}

		if regexIsLower.MatchString(string(c)) && regexIsUpper.MatchString(string(s[i+1])) {
			words = append(words, s[wordBeg:i+1])
			wordBeg = i + 1
			continue
		}

		if i+2 < len(s) && regexIsUpper.MatchString(string(c)) &&
			regexIsUpper.MatchString(string(s[i+1])) &&
			regexIsLower.MatchString(string(s[i+2])) {
			words = append(words, s[wordBeg:i+1])
			wordBeg = i + 1
		}
	}

	words = append(words, s[wordBeg:])

	return words
}

func Camel2Snake(s string) string {
	if s == "Err" {
		return "error"
	}
	if s == "err" {
		return "error"
	}
	var tmp []string
	for _, word := range SplitCamelCase(s) {
		tmp = append(tmp, strings.ToLower(word))
	}
	return strings.Join(tmp, "_")
}
