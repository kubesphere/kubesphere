// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

const (
	none int64 = 1
	kb         = 1000
	ki         = 1024
	mb         = kb * 1000
	mi         = ki * 1024
	gb         = mb * 1000
	gi         = mi * 1024
	tb         = gb * 1000
	ti         = gi * 1024
)

// The rune values for 0..9 as well as the period symbol (for parsing floats)
var numRunes = []rune("0123456789.")

func parseNumBytesError(msg string) error {
	return fmt.Errorf("%s error: %s", ast.UnitsParseBytes.Name, msg)
}

func errUnitNotRecognized(unit string) error {
	return parseNumBytesError(fmt.Sprintf("byte unit %s not recognized", unit))
}

var (
	errNoAmount       = parseNumBytesError("no byte amount provided")
	errIntConv        = parseNumBytesError("could not parse byte amount to integer")
	errIncludesSpaces = parseNumBytesError("spaces not allowed in resource strings")
)

func builtinNumBytes(a ast.Value) (ast.Value, error) {
	var m int64

	raw, err := builtins.StringOperand(a, 1)
	if err != nil {
		return nil, err
	}

	s := formatString(raw)

	if strings.Contains(s, " ") {
		return nil, errIncludesSpaces
	}

	numStr, unitStr := extractNumAndUnit(s)

	if numStr == "" {
		return nil, errNoAmount
	}

	switch unitStr {
	case "":
		m = none
	case "kb", "k":
		m = kb
	case "kib", "ki":
		m = ki
	case "mb", "m":
		m = mb
	case "mib", "mi":
		m = mi
	case "gb", "g":
		m = gb
	case "gib", "gi":
		m = gi
	case "tb", "t":
		m = tb
	case "tib", "ti":
		m = ti
	default:
		return nil, errUnitNotRecognized(unitStr)
	}

	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return nil, errIntConv
	}

	total := num * m

	return builtins.IntToNumber(big.NewInt(total)), nil
}

// Makes the string lower case and removes spaces and quotation marks
func formatString(s ast.String) string {
	str := string(s)
	lower := strings.ToLower(str)
	return strings.Replace(lower, "\"", "", -1)
}

// Splits the string into a number string à la "10" or "10.2" and a unit string à la "gb" or "MiB" or "foo". Either
// can be an empty string (error handling is provided elsewhere).
func extractNumAndUnit(s string) (string, string) {
	isNum := func(r rune) (isNum bool) {
		for _, nr := range numRunes {
			if nr == r {
				return true
			}
		}

		return false
	}

	// Returns the index of the first rune that's not a number (or 0 if there are only numbers)
	getFirstNonNumIdx := func(s string) int {
		for idx, r := range s {
			if !isNum(r) {
				return idx
			}
		}

		return 0
	}

	firstRuneIsNum := func(s string) bool {
		return isNum(rune(s[0]))
	}

	firstNonNumIdx := getFirstNonNumIdx(s)

	// The string contains only a number
	numOnly := firstNonNumIdx == 0 && firstRuneIsNum(s)

	// The string contains only a unit
	unitOnly := firstNonNumIdx == 0 && !firstRuneIsNum(s)

	if numOnly {
		return s, ""
	} else if unitOnly {
		return "", s
	} else {
		return s[0:firstNonNumIdx], s[firstNonNumIdx:]
	}
}

func init() {
	RegisterFunctionalBuiltin1(ast.UnitsParseBytes.Name, builtinNumBytes)
}
