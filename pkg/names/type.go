package names

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	regexIsLower = regexp.MustCompile("[a-z]")
	regexIsUpper = regexp.MustCompile("[A-Z]")
	//regexIsDigit = regexp.MustCompile("\\d")
)

func PackageNameFromType(typeName string) string {
	return strings.ToLower(typeName)
}

func TypeNameWithPackage(packageName, typeName string) string {
	return fmt.Sprintf("%s.%s", packageName, typeName)
}

func getRawAbbr(s string) string {
	wordBeg := 0
	var res string
	var words []string
	for i, c := range s {
		if i+1 == len(s) {
			break
		}

		if regexIsLower.MatchString(string(c)) && regexIsUpper.MatchString(string(s[i+1])) {
			words = append(words, s[wordBeg:i+1])
			wordBeg = i + 1
		}
		if i+2 < len(s) && regexIsUpper.MatchString(string(c)) &&
			regexIsUpper.MatchString(string(s[i+1])) &&
			regexIsLower.MatchString(string(s[i+2])) {
			words = append(words, s[wordBeg:i+1])
			wordBeg = i + 1
		}
	}
	words = append(words, s[wordBeg:])

	for _, word := range words {
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
