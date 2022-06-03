// Copyright 2017 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

var tzCache map[string]*time.Location
var tzCacheMutex *sync.Mutex

func builtinTimeNowNanos(bctx BuiltinContext, _ []*ast.Term, iter func(*ast.Term) error) error {
	return iter(bctx.Time)
}

func builtinTimeParseNanos(a, b ast.Value) (ast.Value, error) {

	format, err := builtins.StringOperand(a, 1)
	if err != nil {
		return nil, err
	}

	value, err := builtins.StringOperand(b, 2)
	if err != nil {
		return nil, err
	}

	result, err := time.Parse(string(format), string(value))
	if err != nil {
		return nil, err
	}

	return ast.Number(int64ToJSONNumber(result.UnixNano())), nil
}

func builtinTimeParseRFC3339Nanos(a ast.Value) (ast.Value, error) {

	value, err := builtins.StringOperand(a, 1)
	if err != nil {
		return nil, err
	}

	result, err := time.Parse(time.RFC3339, string(value))
	if err != nil {
		return nil, err
	}

	return ast.Number(int64ToJSONNumber(result.UnixNano())), nil
}
func builtinParseDurationNanos(a ast.Value) (ast.Value, error) {

	duration, err := builtins.StringOperand(a, 1)
	if err != nil {
		return nil, err
	}
	value, err := time.ParseDuration(string(duration))
	if err != nil {
		return nil, err
	}
	return ast.Number(int64ToJSONNumber(int64(value))), nil
}

func builtinDate(a ast.Value) (ast.Value, error) {
	t, err := tzTime(a)
	if err != nil {
		return nil, err
	}
	year, month, day := t.Date()
	result := ast.NewArray(ast.IntNumberTerm(year), ast.IntNumberTerm(int(month)), ast.IntNumberTerm(day))
	return result, nil
}

func builtinClock(a ast.Value) (ast.Value, error) {
	t, err := tzTime(a)
	if err != nil {
		return nil, err
	}
	hour, minute, second := t.Clock()
	result := ast.NewArray(ast.IntNumberTerm(hour), ast.IntNumberTerm(minute), ast.IntNumberTerm(second))
	return result, nil
}

func builtinWeekday(a ast.Value) (ast.Value, error) {
	t, err := tzTime(a)
	if err != nil {
		return nil, err
	}
	weekday := t.Weekday().String()
	return ast.String(weekday), nil
}

func builtinAddDate(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	t, err := tzTime(operands[0].Value)
	if err != nil {
		return err
	}

	years, err := builtins.IntOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	months, err := builtins.IntOperand(operands[2].Value, 3)
	if err != nil {
		return err
	}

	days, err := builtins.IntOperand(operands[3].Value, 4)
	if err != nil {
		return err
	}

	result := t.AddDate(years, months, days)
	return iter(ast.NewTerm(ast.Number(int64ToJSONNumber(result.UnixNano()))))
}

func tzTime(a ast.Value) (t time.Time, err error) {
	var nVal ast.Value
	loc := time.UTC

	switch va := a.(type) {
	case *ast.Array:
		if va.Len() == 0 {
			return time.Time{}, builtins.NewOperandTypeErr(1, a, "either number (ns) or [number (ns), string (tz)]")
		}

		nVal, err = builtins.NumberOperand(va.Elem(0).Value, 1)
		if err != nil {
			return time.Time{}, err
		}

		if va.Len() > 1 {
			tzVal, err := builtins.StringOperand(va.Elem(1).Value, 1)
			if err != nil {
				return time.Time{}, err
			}

			tzName := string(tzVal)

			switch tzName {
			case "", "UTC":
				// loc is already UTC

			case "Local":
				loc = time.Local

			default:
				var ok bool

				tzCacheMutex.Lock()
				loc, ok = tzCache[tzName]

				if !ok {
					loc, err = time.LoadLocation(tzName)
					if err != nil {
						tzCacheMutex.Unlock()
						return time.Time{}, err
					}
					tzCache[tzName] = loc
				}
				tzCacheMutex.Unlock()
			}
		}

	case ast.Number:
		nVal = a

	default:
		return time.Time{}, builtins.NewOperandTypeErr(1, a, "either number (ns) or [number (ns), string (tz)]")
	}

	value, err := builtins.NumberOperand(nVal, 1)
	if err != nil {
		return time.Time{}, err
	}

	f := builtins.NumberToFloat(value)
	i64, acc := f.Int64()
	if acc != big.Exact {
		return time.Time{}, fmt.Errorf("timestamp too big")
	}

	t = time.Unix(0, i64).In(loc)

	return t, nil
}

func int64ToJSONNumber(i int64) json.Number {
	return json.Number(strconv.FormatInt(i, 10))
}

func init() {
	RegisterBuiltinFunc(ast.NowNanos.Name, builtinTimeNowNanos)
	RegisterFunctionalBuiltin1(ast.ParseRFC3339Nanos.Name, builtinTimeParseRFC3339Nanos)
	RegisterFunctionalBuiltin2(ast.ParseNanos.Name, builtinTimeParseNanos)
	RegisterFunctionalBuiltin1(ast.ParseDurationNanos.Name, builtinParseDurationNanos)
	RegisterFunctionalBuiltin1(ast.Date.Name, builtinDate)
	RegisterFunctionalBuiltin1(ast.Clock.Name, builtinClock)
	RegisterFunctionalBuiltin1(ast.Weekday.Name, builtinWeekday)
	RegisterBuiltinFunc(ast.AddDate.Name, builtinAddDate)
	tzCacheMutex = &sync.Mutex{}
	tzCache = make(map[string]*time.Location)
}
