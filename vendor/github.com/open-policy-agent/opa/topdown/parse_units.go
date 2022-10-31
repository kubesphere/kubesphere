// Copyright 2022 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

// Binary Si unit constants are borrowed from topdown/parse_bytes
const siMilli = 0.001
const (
	siK uint64 = 1000
	siM        = siK * 1000
	siG        = siM * 1000
	siT        = siG * 1000
	siP        = siT * 1000
	siE        = siP * 1000
)

func parseUnitsError(msg string) error {
	return fmt.Errorf("%s: %s", ast.UnitsParse.Name, msg)
}

func errUnitNotRecognized(unit string) error {
	return parseUnitsError(fmt.Sprintf("unit %s not recognized", unit))
}

var (
	errNoAmount       = parseUnitsError("no amount provided")
	errNumConv        = parseUnitsError("could not parse amount to a number")
	errIncludesSpaces = parseUnitsError("spaces not allowed in resource strings")
)

// Accepts both normal SI and binary SI units.
func builtinUnits(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	var x big.Rat

	raw, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	// We remove escaped quotes from strings here to retain parity with units.parse_bytes.
	s := string(raw)
	s = strings.Replace(s, "\"", "", -1)

	if strings.Contains(s, " ") {
		return errIncludesSpaces
	}

	num, unit := extractNumAndUnit(s)
	if num == "" {
		return errNoAmount
	}

	// Unlike in units.parse_bytes, we only lowercase after the first letter,
	// so that we can distinguish between 'm' and 'M'.
	if len(unit) > 1 {
		lower := strings.ToLower(unit[1:])
		unit = unit[:1] + lower
	}

	switch unit {
	case "m":
		x.SetFloat64(siMilli)
	case "":
		x.SetUint64(none)
	case "k", "K":
		x.SetUint64(siK)
	case "ki", "Ki":
		x.SetUint64(ki)
	case "M":
		x.SetUint64(siM)
	case "mi", "Mi":
		x.SetUint64(mi)
	case "g", "G":
		x.SetUint64(siG)
	case "gi", "Gi":
		x.SetUint64(gi)
	case "t", "T":
		x.SetUint64(siT)
	case "ti", "Ti":
		x.SetUint64(ti)
	case "p", "P":
		x.SetUint64(siP)
	case "pi", "Pi":
		x.SetUint64(pi)
	case "e", "E":
		x.SetUint64(siE)
	case "ei", "Ei":
		x.SetUint64(ei)
	default:
		return errUnitNotRecognized(unit)
	}

	numRat, ok := new(big.Rat).SetString(num)
	if !ok {
		return errNumConv
	}

	numRat.Mul(numRat, &x)

	// Cleaner printout when we have a pure integer value.
	if numRat.IsInt() {
		return iter(ast.NumberTerm(json.Number(numRat.Num().String())))
	}

	// When using just big.Float, we had floating-point precision
	// issues because quantities like 0.001 are not exactly representable.
	// Rationals (such as big.Rat) do not suffer this problem, but are
	// more expensive to compute with in general.
	return iter(ast.NumberTerm(json.Number(numRat.FloatString(10))))
}

func init() {
	RegisterBuiltinFunc(ast.UnitsParse.Name, builtinUnits)
}
