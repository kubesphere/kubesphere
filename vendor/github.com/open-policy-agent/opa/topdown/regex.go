// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/yashtewari/glob-intersection"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

var regexpCacheLock = sync.Mutex{}
var regexpCache map[string]*regexp.Regexp

func builtinRegexMatch(a, b ast.Value) (ast.Value, error) {
	s1, err := builtins.StringOperand(a, 1)
	if err != nil {
		return nil, err
	}
	s2, err := builtins.StringOperand(b, 2)
	if err != nil {
		return nil, err
	}
	re, err := getRegexp(string(s1))
	if err != nil {
		return nil, err
	}
	return ast.Boolean(re.Match([]byte(s2))), nil
}

func builtinRegexMatchTemplate(a, b, c, d ast.Value) (ast.Value, error) {
	pattern, err := builtins.StringOperand(a, 1)
	if err != nil {
		return nil, err
	}
	match, err := builtins.StringOperand(b, 2)
	if err != nil {
		return nil, err
	}
	start, err := builtins.StringOperand(c, 3)
	if err != nil {
		return nil, err
	}
	end, err := builtins.StringOperand(d, 4)
	if err != nil {
		return nil, err
	}
	if len(start) != 1 {
		return nil, fmt.Errorf("start delimiter has to be exactly one character long but is %d long", len(start))
	}
	if len(end) != 1 {
		return nil, fmt.Errorf("end delimiter has to be exactly one character long but is %d long", len(start))
	}
	re, err := getRegexpTemplate(string(pattern), string(start)[0], string(end)[0])
	if err != nil {
		return nil, err
	}
	return ast.Boolean(re.MatchString(string(match))), nil
}

func builtinRegexSplit(a, b ast.Value) (ast.Value, error) {
	s1, err := builtins.StringOperand(a, 1)
	if err != nil {
		return nil, err
	}
	s2, err := builtins.StringOperand(b, 2)
	if err != nil {
		return nil, err
	}
	re, err := getRegexp(string(s1))
	if err != nil {
		return nil, err
	}

	elems := re.Split(string(s2), -1)
	arr := make(ast.Array, len(elems))
	for i := range arr {
		arr[i] = ast.StringTerm(elems[i])
	}
	return arr, nil
}

func getRegexp(pat string) (*regexp.Regexp, error) {
	regexpCacheLock.Lock()
	defer regexpCacheLock.Unlock()
	re, ok := regexpCache[pat]
	if !ok {
		var err error
		re, err = regexp.Compile(string(pat))
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
		re, err = compileRegexTemplate(string(pat), delimStart, delimEnd)
		if err != nil {
			return nil, err
		}
		regexpCache[pat] = re
	}
	return re, nil
}

func builtinGlobsMatch(a, b ast.Value) (ast.Value, error) {
	s1, err := builtins.StringOperand(a, 1)
	if err != nil {
		return nil, err
	}
	s2, err := builtins.StringOperand(b, 2)
	if err != nil {
		return nil, err
	}
	ne, err := gintersect.NonEmpty(string(s1), string(s2))
	if err != nil {
		return nil, err
	}
	return ast.Boolean(ne), nil
}

func builtinRegexFind(a, b, c ast.Value) (ast.Value, error) {
	s1, err := builtins.StringOperand(a, 1)
	if err != nil {
		return nil, err
	}
	s2, err := builtins.StringOperand(b, 2)
	if err != nil {
		return nil, err
	}
	n, err := builtins.IntOperand(c, 3)
	if err != nil {
		return nil, err
	}
	re, err := getRegexp(string(s1))
	if err != nil {
		return nil, err
	}

	elems := re.FindAllString(string(s2), n)
	arr := make(ast.Array, len(elems))
	for i := range arr {
		arr[i] = ast.StringTerm(elems[i])
	}
	return arr, nil
}

func builtinRegexFindAllStringSubmatch(a, b, c ast.Value) (ast.Value, error) {
	s1, err := builtins.StringOperand(a, 1)
	if err != nil {
		return nil, err
	}
	s2, err := builtins.StringOperand(b, 2)
	if err != nil {
		return nil, err
	}
	n, err := builtins.IntOperand(c, 3)
	if err != nil {
		return nil, err
	}

	re, err := getRegexp(string(s1))
	if err != nil {
		return nil, err
	}
	matches := re.FindAllStringSubmatch(string(s2), n)

	outer := make(ast.Array, len(matches))
	for i := range outer {
		inner := make(ast.Array, len(matches[i]))
		for j := range inner {
			inner[j] = ast.StringTerm(matches[i][j])
		}
		outer[i] = ast.ArrayTerm(inner...)
	}

	return outer, nil
}

func init() {
	regexpCache = map[string]*regexp.Regexp{}
	RegisterFunctionalBuiltin2(ast.RegexMatch.Name, builtinRegexMatch)
	RegisterFunctionalBuiltin2(ast.RegexSplit.Name, builtinRegexSplit)
	RegisterFunctionalBuiltin2(ast.GlobsMatch.Name, builtinGlobsMatch)
	RegisterFunctionalBuiltin4(ast.RegexTemplateMatch.Name, builtinRegexMatchTemplate)
	RegisterFunctionalBuiltin3(ast.RegexFind.Name, builtinRegexFind)
	RegisterFunctionalBuiltin3(ast.RegexFindAllStringSubmatch.Name, builtinRegexFindAllStringSubmatch)
}
