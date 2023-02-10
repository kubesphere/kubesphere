// Copyright 2017 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

var tzCache map[string]*time.Location
var tzCacheMutex *sync.Mutex

// 1677-09-21T00:12:43.145224192-00:00
var minDateAllowedForNsConversion = time.Unix(0, math.MinInt64)

// 2262-04-11T23:47:16.854775807-00:00
var maxDateAllowedForNsConversion = time.Unix(0, math.MaxInt64)

func toSafeUnixNano(t time.Time, iter func(*ast.Term) error) error {
	if t.Before(minDateAllowedForNsConversion) || t.After(maxDateAllowedForNsConversion) {
		return fmt.Errorf("time outside of valid range")
	}

	return iter(ast.NewTerm(ast.Number(int64ToJSONNumber(t.UnixNano()))))
}

func builtinTimeNowNanos(bctx BuiltinContext, _ []*ast.Term, iter func(*ast.Term) error) error {
	return iter(bctx.Time)
}

func builtinTimeParseNanos(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	format, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	value, err := builtins.StringOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	result, err := time.Parse(string(format), string(value))
	if err != nil {
		return err
	}

	return toSafeUnixNano(result, iter)
}

func builtinTimeParseRFC3339Nanos(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	value, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	result, err := time.Parse(time.RFC3339, string(value))
	if err != nil {
		return err
	}

	return toSafeUnixNano(result, iter)
}
func builtinParseDurationNanos(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	duration, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}
	value, err := time.ParseDuration(string(duration))
	if err != nil {
		return err
	}
	return iter(ast.NumberTerm(int64ToJSONNumber(int64(value))))
}

func builtinFormat(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	t, layout, err := tzTime(operands[0].Value)
	if err != nil {
		return err
	}
	// Using RFC3339Nano time formatting as default
	if layout == "" {
		layout = time.RFC3339Nano
	}
	timestamp := t.Format(layout)
	return iter(ast.StringTerm(timestamp))
}

func builtinDate(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	t, _, err := tzTime(operands[0].Value)
	if err != nil {
		return err
	}
	year, month, day := t.Date()
	result := ast.NewArray(ast.IntNumberTerm(year), ast.IntNumberTerm(int(month)), ast.IntNumberTerm(day))
	return iter(ast.NewTerm(result))
}

func builtinClock(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	t, _, err := tzTime(operands[0].Value)
	if err != nil {
		return err
	}
	hour, minute, second := t.Clock()
	result := ast.NewArray(ast.IntNumberTerm(hour), ast.IntNumberTerm(minute), ast.IntNumberTerm(second))
	return iter(ast.NewTerm(result))
}

func builtinWeekday(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	t, _, err := tzTime(operands[0].Value)
	if err != nil {
		return err
	}
	weekday := t.Weekday().String()
	return iter(ast.StringTerm(weekday))
}

func builtinAddDate(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	t, _, err := tzTime(operands[0].Value)
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

	return toSafeUnixNano(result, iter)
}

func builtinDiff(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	t1, _, err := tzTime(operands[0].Value)
	if err != nil {
		return err
	}
	t2, _, err := tzTime(operands[1].Value)
	if err != nil {
		return err
	}

	// The following implementation of this function is taken
	// from https://github.com/icza/gox licensed under Apache 2.0.
	// The only modification made is to variable names.
	//
	// For details, see https://stackoverflow.com/a/36531443/1705598
	//
	// Copyright 2021 icza
	// BEGIN REDISTRIBUTION FROM APACHE 2.0 LICENSED PROJECT
	if t1.Location() != t2.Location() {
		t2 = t2.In(t1.Location())
	}
	if t1.After(t2) {
		t1, t2 = t2, t1
	}
	y1, M1, d1 := t1.Date()
	y2, M2, d2 := t2.Date()

	h1, m1, s1 := t1.Clock()
	h2, m2, s2 := t2.Clock()

	year := y2 - y1
	month := int(M2 - M1)
	day := d2 - d1
	hour := h2 - h1
	min := m2 - m1
	sec := s2 - s1

	// Normalize negative values
	if sec < 0 {
		sec += 60
		min--
	}
	if min < 0 {
		min += 60
		hour--
	}
	if hour < 0 {
		hour += 24
		day--
	}
	if day < 0 {
		// Days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}
	// END REDISTRIBUTION FROM APACHE 2.0 LICENSED PROJECT

	return iter(ast.ArrayTerm(ast.IntNumberTerm(year), ast.IntNumberTerm(month), ast.IntNumberTerm(day),
		ast.IntNumberTerm(hour), ast.IntNumberTerm(min), ast.IntNumberTerm(sec)))
}

func tzTime(a ast.Value) (t time.Time, lay string, err error) {
	var nVal ast.Value
	loc := time.UTC
	layout := ""
	switch va := a.(type) {
	case *ast.Array:
		if va.Len() == 0 {
			return time.Time{}, layout, builtins.NewOperandTypeErr(1, a, "either number (ns) or [number (ns), string (tz)]")
		}

		nVal, err = builtins.NumberOperand(va.Elem(0).Value, 1)
		if err != nil {
			return time.Time{}, layout, err
		}

		if va.Len() > 1 {
			tzVal, err := builtins.StringOperand(va.Elem(1).Value, 1)
			if err != nil {
				return time.Time{}, layout, err
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
						return time.Time{}, layout, err
					}
					tzCache[tzName] = loc
				}
				tzCacheMutex.Unlock()
			}
		}

		if va.Len() > 2 {
			lay, err := builtins.StringOperand(va.Elem(2).Value, 1)
			if err != nil {
				return time.Time{}, layout, err
			}
			layout = string(lay)
		}

	case ast.Number:
		nVal = a

	default:
		return time.Time{}, layout, builtins.NewOperandTypeErr(1, a, "either number (ns) or [number (ns), string (tz)]")
	}

	value, err := builtins.NumberOperand(nVal, 1)
	if err != nil {
		return time.Time{}, layout, err
	}

	f := builtins.NumberToFloat(value)
	i64, acc := f.Int64()
	if acc != big.Exact {
		return time.Time{}, layout, fmt.Errorf("timestamp too big")
	}

	t = time.Unix(0, i64).In(loc)

	return t, layout, nil
}

func int64ToJSONNumber(i int64) json.Number {
	return json.Number(strconv.FormatInt(i, 10))
}

func init() {
	RegisterBuiltinFunc(ast.NowNanos.Name, builtinTimeNowNanos)
	RegisterBuiltinFunc(ast.ParseRFC3339Nanos.Name, builtinTimeParseRFC3339Nanos)
	RegisterBuiltinFunc(ast.ParseNanos.Name, builtinTimeParseNanos)
	RegisterBuiltinFunc(ast.ParseDurationNanos.Name, builtinParseDurationNanos)
	RegisterBuiltinFunc(ast.Format.Name, builtinFormat)
	RegisterBuiltinFunc(ast.Date.Name, builtinDate)
	RegisterBuiltinFunc(ast.Clock.Name, builtinClock)
	RegisterBuiltinFunc(ast.Weekday.Name, builtinWeekday)
	RegisterBuiltinFunc(ast.AddDate.Name, builtinAddDate)
	RegisterBuiltinFunc(ast.Diff.Name, builtinDiff)
	tzCacheMutex = &sync.Mutex{}
	tzCache = make(map[string]*time.Location)
}
