// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"fmt"
	"regexp"
	"sync"

	gintersect "github.com/yashtewari/glob-intersection"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/topdown/builtins"
)

const regexCacheMaxSize = 100
const regexInterQueryValueCacheHits = "rego_builtin_regex_interquery_value_cache_hits"

var regexpCacheLock = sync.Mutex{}
var regexpCache map[string]*regexp.Regexp

func builtinRegexIsValid(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {

	s, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return iter(ast.InternedBooleanTerm(false))
	}

	_, err = regexp.Compile(string(s))
	if err != nil {
		return iter(ast.InternedBooleanTerm(false))
	}

	return iter(ast.InternedBooleanTerm(true))
}

func builtinRegexMatch(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	s1, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}
	s2, err := builtins.StringOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}
	re, err := getRegexp(bctx, string(s1))
	if err != nil {
		return err
	}
	return iter(ast.InternedBooleanTerm(re.MatchString(string(s2))))
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
	return iter(ast.InternedBooleanTerm(re.MatchString(string(match))))
}

func builtinRegexSplit(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	s1, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}
	s2, err := builtins.StringOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}
	re, err := getRegexp(bctx, string(s1))
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

func getRegexp(bctx BuiltinContext, pat string) (*regexp.Regexp, error) {
	if bctx.InterQueryBuiltinValueCache != nil {
		// TODO: Use named cache
		val, ok := bctx.InterQueryBuiltinValueCache.Get(ast.String(pat))
		if ok {
			res, valid := val.(*regexp.Regexp)
			if !valid {
				// The cache key may exist for a different value type (eg. glob).
				// In this case, we calculate the regex and return the result w/o updating the cache.
				return regexp.Compile(pat)
			}

			bctx.Metrics.Counter(regexInterQueryValueCacheHits).Incr()
			return res, nil
		}

		re, err := regexp.Compile(pat)
		if err != nil {
			return nil, err
		}
		bctx.InterQueryBuiltinValueCache.Insert(ast.String(pat), re)
		return re, nil
	}

	regexpCacheLock.Lock()
	defer regexpCacheLock.Unlock()
	re, ok := regexpCache[pat]
	if !ok {
		var err error
		re, err = regexp.Compile(pat)
		if err != nil {
			return nil, err
		}
		if len(regexpCache) >= regexCacheMaxSize {
			// Delete a (semi-)random key to make room for the new one.
			for k := range regexpCache {
				delete(regexpCache, k)
				break
			}
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
	return iter(ast.InternedBooleanTerm(ne))
}

func builtinRegexFind(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
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
	re, err := getRegexp(bctx, string(s1))
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

func builtinRegexFindAllStringSubmatch(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
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

	re, err := getRegexp(bctx, string(s1))
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

func builtinRegexReplace(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
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

	re, err := getRegexp(bctx, string(pattern))
	if err != nil {
		return err
	}

	res := re.ReplaceAllString(string(base), string(value))
	if res == string(base) {
		return iter(operands[0])
	}

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
