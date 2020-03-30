// +build gofuzz

package ast

import (
	"regexp"
)

// nested { and [ tokens cause the parse time to explode.
// see: https://github.com/mna/pigeon/issues/75
var blacklistRegexp = regexp.MustCompile(`[{(\[]{5,}`)

func Fuzz(data []byte) int {

	if blacklistRegexp.Match(data) {
		return -1
	}

	str := string(data)
	_, _, err := ParseStatements("", str)

	if err == nil {
		CompileModules(map[string]string{"": str})
		return 1
	}

	return 0
}
