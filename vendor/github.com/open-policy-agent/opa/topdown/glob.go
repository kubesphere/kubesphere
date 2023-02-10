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

func builtinGlobMatch(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	pattern, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	var delimiters []rune
	switch operands[1].Value.(type) {
	case ast.Null:
		delimiters = []rune{}
	case *ast.Array:
		delimiters, err = builtins.RuneSliceOperand(operands[1].Value, 2)
		if err != nil {
			return err
		}
		if len(delimiters) == 0 {
			delimiters = []rune{'.'}
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

	m, err := globCompileAndMatch(id, string(pattern), string(match), delimiters)
	if err != nil {
		return err
	}
	return iter(ast.BooleanTerm(m))
}

func globCompileAndMatch(id, pattern, match string, delimiters []rune) (bool, error) {
	globCacheLock.Lock()
	defer globCacheLock.Unlock()
	p, ok := globCache[id]
	if !ok {
		var err error
		if p, err = glob.Compile(pattern, delimiters...); err != nil {
			return false, err
		}
		globCache[id] = p
	}
	out := p.Match(match)
	return out, nil
}

func builtinGlobQuoteMeta(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	pattern, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	return iter(ast.StringTerm(glob.QuoteMeta(string(pattern))))
}

func init() {
	globCache = map[string]glob.Glob{}
	RegisterBuiltinFunc(ast.GlobMatch.Name, builtinGlobMatch)
	RegisterBuiltinFunc(ast.GlobQuoteMeta.Name, builtinGlobQuoteMeta)
}
