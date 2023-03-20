package utils

import (
	"regexp"
	"strings"
)

const (
	camelToUnderLineRegex = "([A-Z0-9])"
	underLineToCamelRegex = "_(.)"
)

var camelToUnderLineCompiler = regexp.MustCompile(camelToUnderLineRegex)

var underLineToCamelCompiler = regexp.MustCompile(underLineToCamelRegex)

func CamelToUnderLine(src string) string {
	dest := camelToUnderLineCompiler.ReplaceAllStringFunc(src, func(s string) string {
		return "_" + strings.ToLower(s)
	})
	return strings.TrimLeft(dest, "_")
}

func UnderLineToCamel(src string) string {

	res := underLineToCamelCompiler.ReplaceAllStringFunc(src, func(s string) string {
		return strings.ToUpper(string(s[1]))
	})

	return strings.ToUpper(string(res[0])) + res[1:]

}
