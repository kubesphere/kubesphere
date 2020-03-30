package planner

import (
	"sort"

	"github.com/open-policy-agent/opa/ast"
)

// funcstack implements a simple map structure used to keep track of virtual
// document => planned function names. The structure supports Push and Pop
// operations so that the planner can shadow planned functions when 'with'
// statements are found.
type funcstack struct {
	stack []map[string]string
	gen   int
}

func newFuncstack() *funcstack {
	return &funcstack{
		stack: []map[string]string{
			map[string]string{},
		},
		gen: 0,
	}
}

func (p funcstack) Add(key, value string) {
	p.stack[len(p.stack)-1][key] = value
}

func (p funcstack) Get(key string) (string, bool) {
	value, ok := p.stack[len(p.stack)-1][key]
	return value, ok
}

func (p *funcstack) Push(funcs map[string]string) {
	p.stack = append(p.stack, funcs)
	p.gen++
}

func (p *funcstack) Pop() map[string]string {
	last := p.stack[len(p.stack)-1]
	p.stack = p.stack[:len(p.stack)-1]
	p.gen++
	return last
}

// ruletrie implements a simple trie structure for organizing rules that may be
// planned. The trie nodes are keyed by the rule path. The ruletrie supports
// Push and Pop operations that allow the planner to shadow subtrees when 'with'
// statements are found.
type ruletrie struct {
	children map[ast.Value][]*ruletrie
	rules    []*ast.Rule
}

func newRuletrie() *ruletrie {
	return &ruletrie{
		children: map[ast.Value][]*ruletrie{},
	}
}

func (t *ruletrie) Arity() int {
	rules := t.Rules()
	if len(rules) > 0 {
		return len(rules[0].Head.Args)
	}
	return 0
}

func (t *ruletrie) Rules() []*ast.Rule {
	if t != nil {
		return t.rules
	}
	return nil
}

func (t *ruletrie) Push(key ast.Ref) {
	node := t
	for i := 0; i < len(key)-1; i++ {
		node = node.Get(key[i].Value)
		if node == nil {
			return
		}
	}
	elem := key[len(key)-1]
	node.children[elem.Value] = append(node.children[elem.Value], nil)
}

func (t *ruletrie) Pop(key ast.Ref) {
	node := t
	for i := 0; i < len(key)-1; i++ {
		node = node.Get(key[i].Value)
		if node == nil {
			return
		}
	}
	elem := key[len(key)-1]
	sl := node.children[elem.Value]
	node.children[elem.Value] = sl[:len(sl)-1]
}

func (t *ruletrie) Insert(key ast.Ref) *ruletrie {
	node := t
	for _, elem := range key {
		child := node.Get(elem.Value)
		if child == nil {
			child = newRuletrie()
			node.children[elem.Value] = append(node.children[elem.Value], child)
		}
		node = child
	}
	return node
}

func (t *ruletrie) Lookup(key ast.Ref) *ruletrie {
	node := t
	for _, elem := range key {
		node = node.Get(elem.Value)
		if node == nil {
			return nil
		}
	}
	return node
}

func (t *ruletrie) LookupOrInsert(key ast.Ref) *ruletrie {
	if val := t.Lookup(key); val != nil {
		return val
	}
	return t.Insert(key)
}

func (t *ruletrie) Children() []ast.Value {
	sorted := make([]ast.Value, 0, len(t.children))
	for key := range t.children {
		if t.Get(key) != nil {
			sorted = append(sorted, key)
		}
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Compare(sorted[j]) < 0
	})
	return sorted
}

func (t *ruletrie) Get(k ast.Value) *ruletrie {
	if t == nil {
		return nil
	}
	nodes := t.children[k]
	if len(nodes) == 0 {
		return nil
	}
	return nodes[len(nodes)-1]
}
