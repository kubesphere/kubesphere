package planner

import (
	"fmt"
	"sort"

	"github.com/open-policy-agent/opa/ast"
)

// funcstack implements a simple map structure used to keep track of virtual
// document => planned function names. The structure supports Push and Pop
// operations so that the planner can shadow planned functions when 'with'
// statements are found.
// The "gen" numbers indicate the "generations"; whenever a 'with' statement
// is planned (a new map is `Push()`ed), it will jump to a previously unused
// number.
type funcstack struct {
	stack []taggedPairs
	next  int
}

type taggedPairs struct {
	pairs map[string]string
	gen   int
}

func newFuncstack() *funcstack {
	return &funcstack{
		stack: []taggedPairs{{pairs: map[string]string{}, gen: 0}},
		next:  1}
}

func (p funcstack) last() taggedPairs {
	return p.stack[len(p.stack)-1]
}

func (p funcstack) Add(key, value string) {
	p.last().pairs[key] = value
}

func (p funcstack) Get(key string) (string, bool) {
	value, ok := p.last().pairs[key]
	return value, ok
}

func (p *funcstack) Push(funcs map[string]string) {
	p.stack = append(p.stack, taggedPairs{pairs: funcs, gen: p.next})
	p.next++
}

func (p *funcstack) Pop() map[string]string {
	last := p.last()
	p.stack = p.stack[:len(p.stack)-1]
	return last.pairs
}

func (p funcstack) gen() int {
	return p.last().gen
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
		if t.rules == nil {
			return nil
		}
		rules := make([]*ast.Rule, len(t.rules), len(t.rules)+len(t.children)) // could be too little
		copy(rules, t.rules)

		// NOTE(sr): We pull in one layer of children: the compiler ensures
		// that these are the only possible, relevant rule sources for a given
		// ref: If the trie is what we get for
		//
		//     a.b.c  = 1 { ... }
		//     a.b[x] = 2 { ... }
		//
		// and we're retrieving a.b, we want Rules() to include the rule body
		// of a.b.c.
		for _, rs := range t.children {
			if r := rs[len(rs)-1].rules; r != nil {
				rules = append(rules, r...)
			}
		}
		return rules
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
	if t == nil {
		return nil
	}
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

func (t *ruletrie) DepthFirst(f func(*ruletrie) bool) {
	if f(t) {
		return
	}
	for _, rules := range t.children {
		for i := range rules {
			rules[i].DepthFirst(f)
		}
	}
}

func (t *ruletrie) Depth() int {
	if len(t.Children()) == 0 {
		return 0
	}
	c := make([]int, 0, len(t.Children()))
	for _, nodes := range t.children {
		c = append(c, nodes[len(nodes)-1].Depth())
	}
	max := 0
	for i := range c {
		if max < c[i] {
			max = c[i]
		}
	}
	return max + 1
}

func (t *ruletrie) String() string {
	return fmt.Sprintf("<ruletrie rules:%v children:%v>", t.rules, t.children)
}

type functionMocksStack struct {
	stack []*functionMocksElem
}

type functionMocksElem []frame

type frame map[string]*ast.Term

func newFunctionMocksStack() *functionMocksStack {
	stack := &functionMocksStack{}
	stack.Push()
	return stack
}

func newFunctionMocksElem() *functionMocksElem {
	return &functionMocksElem{}
}

func (s *functionMocksStack) Push() {
	s.stack = append(s.stack, newFunctionMocksElem())
}

func (s *functionMocksStack) Pop() {
	s.stack = s.stack[:len(s.stack)-1]
}

func (s *functionMocksStack) PushFrame(f frame) {
	current := s.stack[len(s.stack)-1]
	*current = append(*current, f)
}

func (s *functionMocksStack) PopFrame() {
	current := s.stack[len(s.stack)-1]
	*current = (*current)[:len(*current)-1]
}

func (s *functionMocksStack) Lookup(f string) *ast.Term {
	current := *s.stack[len(s.stack)-1]
	for i := len(current) - 1; i >= 0; i-- {
		if t, ok := current[i][f]; ok {
			return t
		}
	}
	return nil
}
