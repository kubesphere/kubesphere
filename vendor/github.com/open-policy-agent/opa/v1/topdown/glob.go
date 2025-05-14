package topdown

import (
	"strings"
	"sync"

	"github.com/gobwas/glob"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/topdown/builtins"
)

const globCacheMaxSize = 100
const globInterQueryValueCacheHits = "rego_builtin_glob_interquery_value_cache_hits"

var noDelimiters = []rune{}
var dotDelimiters = []rune{'.'}
var globCacheLock = sync.RWMutex{}
var globCache = map[string]glob.Glob{}

func builtinGlobMatch(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	pattern, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	var delimiters []rune
	switch operands[1].Value.(type) {
	case ast.Null:
		delimiters = noDelimiters
	case *ast.Array:
		delimiters, err = builtins.RuneSliceOperand(operands[1].Value, 2)
		if err != nil {
			return err
		}
		if len(delimiters) == 0 {
			delimiters = dotDelimiters
		}
	default:
		return builtins.NewOperandTypeErr(2, operands[1].Value, "array", "null")
	}

	match, err := builtins.StringOperand(operands[2].Value, 3)
	if err != nil {
		return err
	}

	builder := strings.Builder{}
	builder.WriteString(string(pattern))
	builder.WriteRune('-')
	for _, v := range delimiters {
		builder.WriteRune(v)
	}
	id := builder.String()

	m, err := globCompileAndMatch(bctx, id, string(pattern), string(match), delimiters)
	if err != nil {
		return err
	}
	return iter(ast.InternedBooleanTerm(m))
}

func globCompileAndMatch(bctx BuiltinContext, id, pattern, match string, delimiters []rune) (bool, error) {

	if bctx.InterQueryBuiltinValueCache != nil {
		// TODO: Use named cache
		val, ok := bctx.InterQueryBuiltinValueCache.Get(ast.String(id))
		if ok {
			pat, valid := val.(glob.Glob)
			if !valid {
				// The cache key may exist for a different value type (eg. regex).
				// In this case, we calculate the glob and return the result w/o updating the cache.
				var err error
				if pat, err = glob.Compile(pattern, delimiters...); err != nil {
					return false, err
				}
				return pat.Match(match), nil
			}
			bctx.Metrics.Counter(globInterQueryValueCacheHits).Incr()
			out := pat.Match(match)
			return out, nil
		}

		res, err := glob.Compile(pattern, delimiters...)
		if err != nil {
			return false, err
		}
		bctx.InterQueryBuiltinValueCache.Insert(ast.String(id), res)
		return res.Match(match), nil
	}

	globCacheLock.RLock()
	p, ok := globCache[id]
	globCacheLock.RUnlock()
	if !ok {
		var err error
		if p, err = glob.Compile(pattern, delimiters...); err != nil {
			return false, err
		}
		globCacheLock.Lock()
		if len(globCache) >= globCacheMaxSize {
			// Delete a (semi-)random key to make room for the new one.
			for k := range globCache {
				delete(globCache, k)
				break
			}
		}
		globCache[id] = p
		globCacheLock.Unlock()
	}

	return p.Match(match), nil
}

func builtinGlobQuoteMeta(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	pattern, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	return iter(ast.StringTerm(glob.QuoteMeta(string(pattern))))
}

func init() {
	RegisterBuiltinFunc(ast.GlobMatch.Name, builtinGlobMatch)
	RegisterBuiltinFunc(ast.GlobQuoteMeta.Name, builtinGlobQuoteMeta)
}
