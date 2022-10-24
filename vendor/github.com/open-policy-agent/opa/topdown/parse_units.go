// Copyright 2022 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

// Binary Si unit constants are borrowed from topdown/parse_bytes
const milli float64 = 0.001
const (
	k uint64 = 1000
	m        = k * 1000
	g        = m * 1000
	t        = g * 1000
	p        = t * 1000
	e        = p * 1000
)

func parseUnitsError(msg string) error {
	return fmt.Errorf("%s error: %s", ast.UnitsParse.Name, msg)
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
	var x big.Float

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
		x.SetFloat64(milli)
	case "":
		x.SetUint64(none)
	case "k", "K":
		x.SetUint64(k)
	case "ki", "Ki":
		x.SetUint64(ki)
	case "M":
		x.SetUint64(m)
	case "mi", "Mi":
		x.SetUint64(mi)
	case "g", "G":
		x.SetUint64(g)
	case "gi", "Gi":
		x.SetUint64(gi)
	case "t", "T":
		x.SetUint64(t)
	case "ti", "Ti":
		x.SetUint64(ti)
	case "p", "P":
		x.SetUint64(p)
	case "pi", "Pi":
		x.SetUint64(pi)
	case "e", "E":
		x.SetUint64(e)
	case "ei", "Ei":
		x.SetUint64(ei)
	default:
		return errUnitNotRecognized(unit)
	}

	numFloat, ok := new(big.Float).SetString(num)
	if !ok {
		return errNumConv
	}

	numFloat.Mul(numFloat, &x)
	return iter(ast.NewTerm(builtins.FloatToNumber(numFloat)))
}

func init() {
	RegisterBuiltinFunc(ast.UnitsParse.Name, builtinUnits)
}
