// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/tchap/go-patricia/v2/patricia"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

func builtinAnyPrefixMatch(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	a, b := operands[0].Value, operands[1].Value

	var strs []string
	switch a := a.(type) {
	case ast.String:
		strs = []string{string(a)}
	case *ast.Array, ast.Set:
		var err error
		strs, err = builtins.StringSliceOperand(a, 1)
		if err != nil {
			return err
		}
	default:
		return builtins.NewOperandTypeErr(1, a, "string", "set", "array")
	}

	var prefixes []string
	switch b := b.(type) {
	case ast.String:
		prefixes = []string{string(b)}
	case *ast.Array, ast.Set:
		var err error
		prefixes, err = builtins.StringSliceOperand(b, 2)
		if err != nil {
			return err
		}
	default:
		return builtins.NewOperandTypeErr(2, b, "string", "set", "array")
	}

	return iter(ast.BooleanTerm(anyStartsWithAny(strs, prefixes)))
}

func builtinAnySuffixMatch(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	a, b := operands[0].Value, operands[1].Value

	var strsReversed []string
	switch a := a.(type) {
	case ast.String:
		strsReversed = []string{reverseString(string(a))}
	case *ast.Array, ast.Set:
		strs, err := builtins.StringSliceOperand(a, 1)
		if err != nil {
			return err
		}
		strsReversed = make([]string, len(strs))
		for i := range strs {
			strsReversed[i] = reverseString(strs[i])
		}
	default:
		return builtins.NewOperandTypeErr(1, a, "string", "set", "array")
	}

	var suffixesReversed []string
	switch b := b.(type) {
	case ast.String:
		suffixesReversed = []string{reverseString(string(b))}
	case *ast.Array, ast.Set:
		suffixes, err := builtins.StringSliceOperand(b, 2)
		if err != nil {
			return err
		}
		suffixesReversed = make([]string, len(suffixes))
		for i := range suffixes {
			suffixesReversed[i] = reverseString(suffixes[i])
		}
	default:
		return builtins.NewOperandTypeErr(2, b, "string", "set", "array")
	}

	return iter(ast.BooleanTerm(anyStartsWithAny(strsReversed, suffixesReversed)))
}

func anyStartsWithAny(strs []string, prefixes []string) bool {
	if len(strs) == 0 || len(prefixes) == 0 {
		return false
	}
	if len(strs) == 1 && len(prefixes) == 1 {
		return strings.HasPrefix(strs[0], prefixes[0])
	}

	trie := patricia.NewTrie()
	for i := 0; i < len(strs); i++ {
		trie.Insert([]byte(strs[i]), true)
	}

	for i := 0; i < len(prefixes); i++ {
		if trie.MatchSubtree([]byte(prefixes[i])) {
			return true
		}
	}

	return false
}

func builtinFormatInt(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {

	input, err := builtins.NumberOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	base, err := builtins.NumberOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	var format string
	switch base {
	case ast.Number("2"):
		format = "%b"
	case ast.Number("8"):
		format = "%o"
	case ast.Number("10"):
		format = "%d"
	case ast.Number("16"):
		format = "%x"
	default:
		return builtins.NewOperandEnumErr(2, "2", "8", "10", "16")
	}

	f := builtins.NumberToFloat(input)
	i, _ := f.Int(nil)

	return iter(ast.StringTerm(fmt.Sprintf(format, i)))
}

func builtinConcat(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {

	join, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	strs := []string{}

	switch b := operands[1].Value.(type) {
	case *ast.Array:
		err := b.Iter(func(x *ast.Term) error {
			s, ok := x.Value.(ast.String)
			if !ok {
				return builtins.NewOperandElementErr(2, operands[1].Value, x.Value, "string")
			}
			strs = append(strs, string(s))
			return nil
		})
		if err != nil {
			return err
		}
	case ast.Set:
		err := b.Iter(func(x *ast.Term) error {
			s, ok := x.Value.(ast.String)
			if !ok {
				return builtins.NewOperandElementErr(2, operands[1].Value, x.Value, "string")
			}
			strs = append(strs, string(s))
			return nil
		})
		if err != nil {
			return err
		}
	default:
		return builtins.NewOperandTypeErr(2, operands[1].Value, "set", "array")
	}

	return iter(ast.StringTerm(strings.Join(strs, string(join))))
}

func runesEqual(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func builtinIndexOf(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	base, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	search, err := builtins.StringOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}
	if len(string(search)) == 0 {
		return fmt.Errorf("empty search character")
	}

	baseRunes := []rune(string(base))
	searchRunes := []rune(string(search))
	searchLen := len(searchRunes)

	for i, r := range baseRunes {
		if len(baseRunes) >= i+searchLen {
			if r == searchRunes[0] && runesEqual(baseRunes[i:i+searchLen], searchRunes) {
				return iter(ast.IntNumberTerm(i))
			}
		} else {
			break
		}
	}

	return iter(ast.IntNumberTerm(-1))
}

func builtinIndexOfN(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	base, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	search, err := builtins.StringOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}
	if len(string(search)) == 0 {
		return fmt.Errorf("empty search character")
	}

	baseRunes := []rune(string(base))
	searchRunes := []rune(string(search))
	searchLen := len(searchRunes)

	var arr []*ast.Term
	for i, r := range baseRunes {
		if len(baseRunes) >= i+searchLen {
			if r == searchRunes[0] && runesEqual(baseRunes[i:i+searchLen], searchRunes) {
				arr = append(arr, ast.IntNumberTerm(i))
			}
		} else {
			break
		}
	}

	return iter(ast.ArrayTerm(arr...))
}

func builtinSubstring(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {

	base, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}
	runes := []rune(base)

	startIndex, err := builtins.IntOperand(operands[1].Value, 2)
	if err != nil {
		return err
	} else if startIndex >= len(runes) {
		return iter(ast.StringTerm(""))
	} else if startIndex < 0 {
		return fmt.Errorf("negative offset")
	}

	length, err := builtins.IntOperand(operands[2].Value, 3)
	if err != nil {
		return err
	}

	var s ast.String
	if length < 0 {
		s = ast.String(runes[startIndex:])
	} else {
		upto := startIndex + length
		if len(runes) < upto {
			upto = len(runes)
		}
		s = ast.String(runes[startIndex:upto])
	}

	return iter(ast.NewTerm(s))
}

func builtinContains(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	s, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	substr, err := builtins.StringOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	return iter(ast.BooleanTerm(strings.Contains(string(s), string(substr))))
}

func builtinStartsWith(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	s, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	prefix, err := builtins.StringOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	return iter(ast.BooleanTerm(strings.HasPrefix(string(s), string(prefix))))
}

func builtinEndsWith(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	s, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	suffix, err := builtins.StringOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	return iter(ast.BooleanTerm(strings.HasSuffix(string(s), string(suffix))))
}

func builtinLower(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	s, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	return iter(ast.StringTerm(strings.ToLower(string(s))))
}

func builtinUpper(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	s, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	return iter(ast.StringTerm(strings.ToUpper(string(s))))
}

func builtinSplit(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	s, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}
	d, err := builtins.StringOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}
	elems := strings.Split(string(s), string(d))
	arr := make([]*ast.Term, len(elems))
	for i := range elems {
		arr[i] = ast.StringTerm(elems[i])
	}
	return iter(ast.ArrayTerm(arr...))
}

func builtinReplace(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	s, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	old, err := builtins.StringOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	new, err := builtins.StringOperand(operands[2].Value, 3)
	if err != nil {
		return err
	}

	return iter(ast.StringTerm(strings.Replace(string(s), string(old), string(new), -1)))
}

func builtinReplaceN(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	patterns, err := builtins.ObjectOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}
	keys := patterns.Keys()
	sort.Slice(keys, func(i, j int) bool { return ast.Compare(keys[i].Value, keys[j].Value) < 0 })

	s, err := builtins.StringOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	oldnewArr := make([]string, 0, len(keys)*2)
	for _, k := range keys {
		keyVal, ok := k.Value.(ast.String)
		if !ok {
			return builtins.NewOperandErr(1, "non-string key found in pattern object")
		}
		val := patterns.Get(k) // cannot be nil
		strVal, ok := val.Value.(ast.String)
		if !ok {
			return builtins.NewOperandErr(1, "non-string value found in pattern object")
		}
		oldnewArr = append(oldnewArr, string(keyVal), string(strVal))
	}
	if err != nil {
		return err
	}

	r := strings.NewReplacer(oldnewArr...)
	replaced := r.Replace(string(s))

	return iter(ast.StringTerm(replaced))
}

func builtinTrim(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	s, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	c, err := builtins.StringOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	return iter(ast.StringTerm(strings.Trim(string(s), string(c))))
}

func builtinTrimLeft(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	s, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	c, err := builtins.StringOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	return iter(ast.StringTerm(strings.TrimLeft(string(s), string(c))))
}

func builtinTrimPrefix(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	s, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	pre, err := builtins.StringOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	return iter(ast.StringTerm(strings.TrimPrefix(string(s), string(pre))))
}

func builtinTrimRight(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	s, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	c, err := builtins.StringOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	return iter(ast.StringTerm(strings.TrimRight(string(s), string(c))))
}

func builtinTrimSuffix(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	s, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	suf, err := builtins.StringOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	return iter(ast.StringTerm(strings.TrimSuffix(string(s), string(suf))))
}

func builtinTrimSpace(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	s, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	return iter(ast.StringTerm(strings.TrimSpace(string(s))))
}

func builtinSprintf(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	s, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	astArr, ok := operands[1].Value.(*ast.Array)
	if !ok {
		return builtins.NewOperandTypeErr(2, operands[1].Value, "array")
	}

	args := make([]interface{}, astArr.Len())

	for i := range args {
		switch v := astArr.Elem(i).Value.(type) {
		case ast.Number:
			if n, ok := v.Int(); ok {
				args[i] = n
			} else if b, ok := new(big.Int).SetString(v.String(), 10); ok {
				args[i] = b
			} else if f, ok := v.Float64(); ok {
				args[i] = f
			} else {
				args[i] = v.String()
			}
		case ast.String:
			args[i] = string(v)
		default:
			args[i] = astArr.Elem(i).String()
		}
	}

	return iter(ast.StringTerm(fmt.Sprintf(string(s), args...)))
}

func builtinReverse(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	s, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	return iter(ast.StringTerm(reverseString(string(s))))
}

func reverseString(str string) string {
	sRunes := []rune(str)
	length := len(sRunes)
	reversedRunes := make([]rune, length)

	for index, r := range sRunes {
		reversedRunes[length-index-1] = r
	}

	return string(reversedRunes)
}

func init() {
	RegisterBuiltinFunc(ast.FormatInt.Name, builtinFormatInt)
	RegisterBuiltinFunc(ast.Concat.Name, builtinConcat)
	RegisterBuiltinFunc(ast.IndexOf.Name, builtinIndexOf)
	RegisterBuiltinFunc(ast.IndexOfN.Name, builtinIndexOfN)
	RegisterBuiltinFunc(ast.Substring.Name, builtinSubstring)
	RegisterBuiltinFunc(ast.Contains.Name, builtinContains)
	RegisterBuiltinFunc(ast.StartsWith.Name, builtinStartsWith)
	RegisterBuiltinFunc(ast.EndsWith.Name, builtinEndsWith)
	RegisterBuiltinFunc(ast.Upper.Name, builtinUpper)
	RegisterBuiltinFunc(ast.Lower.Name, builtinLower)
	RegisterBuiltinFunc(ast.Split.Name, builtinSplit)
	RegisterBuiltinFunc(ast.Replace.Name, builtinReplace)
	RegisterBuiltinFunc(ast.ReplaceN.Name, builtinReplaceN)
	RegisterBuiltinFunc(ast.Trim.Name, builtinTrim)
	RegisterBuiltinFunc(ast.TrimLeft.Name, builtinTrimLeft)
	RegisterBuiltinFunc(ast.TrimPrefix.Name, builtinTrimPrefix)
	RegisterBuiltinFunc(ast.TrimRight.Name, builtinTrimRight)
	RegisterBuiltinFunc(ast.TrimSuffix.Name, builtinTrimSuffix)
	RegisterBuiltinFunc(ast.TrimSpace.Name, builtinTrimSpace)
	RegisterBuiltinFunc(ast.Sprintf.Name, builtinSprintf)
	RegisterBuiltinFunc(ast.AnyPrefixMatch.Name, builtinAnyPrefixMatch)
	RegisterBuiltinFunc(ast.AnySuffixMatch.Name, builtinAnySuffixMatch)
	RegisterBuiltinFunc(ast.StringReverse.Name, builtinReverse)
}
