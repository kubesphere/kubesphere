package topdown

import (
	"strings"
	"sync"

	"github.com/gobwas/glob"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

var globCacheLock = sync.Mutex{}
var globCache map[string]glob.Glob

func builtinGlobMatch(a, b, c ast.Value) (ast.Value, error) {
	pattern, err := builtins.StringOperand(a, 1)
	if err != nil {
		return nil, err
	}
	var delimiters []rune
	switch b.(type) {
	case ast.Null:
		delimiters = []rune{}
	case *ast.Array:
		delimiters, err = builtins.RuneSliceOperand(b, 2)
		if err != nil {
			return nil, err
		}

		if len(delimiters) == 0 {
			delimiters = []rune{'.'}
		}
	default:
		return nil, builtins.NewOperandTypeErr(2, b, "array", "null")
	}
	match, err := builtins.StringOperand(c, 3)
	if err != nil {
		return nil, err
	}

	builder := strings.Builder{}
	builder.WriteString(string(pattern))
	builder.WriteRune('-')
	for _, v := range delimiters {
		builder.WriteRune(v)
	}
	id := builder.String()

	globCacheLock.Lock()
	defer globCacheLock.Unlock()
	p, ok := globCache[id]
	if !ok {
		var err error
		if p, err = glob.Compile(string(pattern), delimiters...); err != nil {
			return nil, err
		}
		globCache[id] = p
	}

	m := p.Match(string(match))
	return ast.Boolean(m), nil
}

func builtinGlobQuoteMeta(a ast.Value) (ast.Value, error) {
	pattern, err := builtins.StringOperand(a, 1)
	if err != nil {
		return nil, err
	}

	return ast.String(glob.QuoteMeta(string(pattern))), nil
}

func init() {
	globCache = map[string]glob.Glob{}
	RegisterFunctionalBuiltin3(ast.GlobMatch.Name, builtinGlobMatch)
	RegisterFunctionalBuiltin1(ast.GlobQuoteMeta.Name, builtinGlobQuoteMeta)
}
