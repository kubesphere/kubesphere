// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"fmt"
	"regexp"
	"sync"

	gintersect "github.com/yashtewari/glob-intersection"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

var regexpCacheLock = sync.Mutex{}
var regexpCache map[string]*regexp.Regexp

func builtinRegexIsValid(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {

	s, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return iter(ast.BooleanTerm(false))
	}

	_, err = regexp.Compile(string(s))
	if err != nil {
		return iter(ast.BooleanTerm(false))
	}

	return iter(ast.BooleanTerm(true))
}

func builtinRegexMatch(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	s1, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}
	s2, err := builtins.StringOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}
	re, err := getRegexp(string(s1))
	if err != nil {
		return err
	}
	return iter(ast.BooleanTerm(re.Match([]byte(s2))))
}

func builtinRegexMatchTemplate(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	pattern, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}
	match, err := builtins.StringOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}
	start, err := builtins.StringOperand(operands[2].Value, 3)
	if err != nil {
		return err
	}
	end, err := builtins.StringOperand(operands[3].Value, 4)
	if err != nil {
		return err
	}
	if len(start) != 1 {
		return fmt.Errorf("start delimiter has to be exactly one character long but is %d long", len(start))
	}
	if len(end) != 1 {
		return fmt.Errorf("end delimiter has to be exactly one character long but is %d long", len(start))
	}
	re, err := getRegexpTemplate(string(pattern), string(start)[0], string(end)[0])
	if err != nil {
		return err
	}
	return iter(ast.BooleanTerm(re.MatchString(string(match))))
}

func builtinRegexSplit(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	s1, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}
	s2, err := builtins.StringOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}
	re, err := getRegexp(string(s1))
	if err != nil {
		return err
	}

	elems := re.Split(string(s2), -1)
	arr := make([]*ast.Term, len(elems))
	for i := range elems {
		arr[i] = ast.StringTerm(elems[i])
	}
	return iter(ast.NewTerm(ast.NewArray(arr...)))
}

func getRegexp(pat string) (*regexp.Regexp, error) {
	regexpCacheLock.Lock()
	defer regexpCacheLock.Unlock()
	re, ok := regexpCache[pat]
	if !ok {
		var err error
		re, err = regexp.Compile(pat)
		if err != nil {
			return nil, err
		}
		regexpCache[pat] = re
	}
	return re, nil
}

func getRegexpTemplate(pat string, delimStart, delimEnd byte) (*regexp.Regexp, error) {
	regexpCacheLock.Lock()
	defer regexpCacheLock.Unlock()
	re, ok := regexpCache[pat]
	if !ok {
		var err error
		re, err = compileRegexTemplate(pat, delimStart, delimEnd)
		if err != nil {
			return nil, err
		}
		regexpCache[pat] = re
	}
	return re, nil
}

func builtinGlobsMatch(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	s1, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}
	s2, err := builtins.StringOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}
	ne, err := gintersect.NonEmpty(string(s1), string(s2))
	if err != nil {
		return err
	}
	return iter(ast.BooleanTerm(ne))
}

func builtinRegexFind(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	s1, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}
	s2, err := builtins.StringOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}
	n, err := builtins.IntOperand(operands[2].Value, 3)
	if err != nil {
		return err
	}
	re, err := getRegexp(string(s1))
	if err != nil {
		return err
	}

	elems := re.FindAllString(string(s2), n)
	arr := make([]*ast.Term, len(elems))
	for i := range elems {
		arr[i] = ast.StringTerm(elems[i])
	}
	return iter(ast.NewTerm(ast.NewArray(arr...)))
}

func builtinRegexFindAllStringSubmatch(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	s1, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}
	s2, err := builtins.StringOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}
	n, err := builtins.IntOperand(operands[2].Value, 3)
	if err != nil {
		return err
	}

	re, err := getRegexp(string(s1))
	if err != nil {
		return err
	}
	matches := re.FindAllStringSubmatch(string(s2), n)

	outer := make([]*ast.Term, len(matches))
	for i := range matches {
		inner := make([]*ast.Term, len(matches[i]))
		for j := range matches[i] {
			inner[j] = ast.StringTerm(matches[i][j])
		}
		outer[i] = ast.NewTerm(ast.NewArray(inner...))
	}

	return iter(ast.NewTerm(ast.NewArray(outer...)))
}

func builtinRegexReplace(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	base, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	pattern, err := builtins.StringOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	value, err := builtins.StringOperand(operands[2].Value, 3)
	if err != nil {
		return err
	}

	re, err := getRegexp(string(pattern))
	if err != nil {
		return err
	}

	res := re.ReplaceAllString(string(base), string(value))

	return iter(ast.StringTerm(res))
}

func init() {
	regexpCache = map[string]*regexp.Regexp{}
	RegisterBuiltinFunc(ast.RegexIsValid.Name, builtinRegexIsValid)
	RegisterBuiltinFunc(ast.RegexMatch.Name, builtinRegexMatch)
	RegisterBuiltinFunc(ast.RegexMatchDeprecated.Name, builtinRegexMatch)
	RegisterBuiltinFunc(ast.RegexSplit.Name, builtinRegexSplit)
	RegisterBuiltinFunc(ast.GlobsMatch.Name, builtinGlobsMatch)
	RegisterBuiltinFunc(ast.RegexTemplateMatch.Name, builtinRegexMatchTemplate)
	RegisterBuiltinFunc(ast.RegexFind.Name, builtinRegexFind)
	RegisterBuiltinFunc(ast.RegexFindAllStringSubmatch.Name, builtinRegexFindAllStringSubmatch)
	RegisterBuiltinFunc(ast.RegexReplace.Name, builtinRegexReplace)
}
