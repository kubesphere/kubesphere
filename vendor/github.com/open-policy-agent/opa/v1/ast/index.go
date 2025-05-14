// Copyright 2017 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/open-policy-agent/opa/v1/util"
)

// RuleIndex defines the interface for rule indices.
type RuleIndex interface {

	// Build tries to construct an index for the given rules. If the index was
	// constructed, it returns true, otherwise false.
	Build(rules []*Rule) bool

	// Lookup searches the index for rules that will match the provided
	// resolver. If the resolver returns an error, it is returned via err.
	Lookup(resolver ValueResolver) (*IndexResult, error)

	// AllRules traverses the index and returns all rules that will match
	// the provided resolver without any optimizations (effectively with
	// indexing disabled). If the resolver returns an error, it is returned
	// via err.
	AllRules(resolver ValueResolver) (*IndexResult, error)
}

// IndexResult contains the result of an index lookup.
type IndexResult struct {
	Rules          []*Rule
	Else           map[*Rule][]*Rule
	Default        *Rule
	Kind           RuleKind
	EarlyExit      bool
	OnlyGroundRefs bool
}

// NewIndexResult returns a new IndexResult object.
func NewIndexResult(kind RuleKind) *IndexResult {
	return &IndexResult{
		Kind: kind,
	}
}

// Empty returns true if there are no rules to evaluate.
func (ir *IndexResult) Empty() bool {
	return len(ir.Rules) == 0 && ir.Default == nil
}

type baseDocEqIndex struct {
	isVirtual      func(Ref) bool
	root           *trieNode
	defaultRule    *Rule
	kind           RuleKind
	onlyGroundRefs bool
}

var (
	equalityRef         = Equality.Ref()
	equalRef            = Equal.Ref()
	globMatchRef        = GlobMatch.Ref()
	internalPrintRef    = InternalPrint.Ref()
	internalTestCaseRef = InternalTestCase.Ref()

	skipIndexing = NewSet(NewTerm(internalPrintRef), NewTerm(internalTestCaseRef))
)

func newBaseDocEqIndex(isVirtual func(Ref) bool) *baseDocEqIndex {
	return &baseDocEqIndex{
		isVirtual:      isVirtual,
		root:           newTrieNodeImpl(),
		onlyGroundRefs: true,
	}
}

func (i *baseDocEqIndex) Build(rules []*Rule) bool {
	if len(rules) == 0 {
		return false
	}

	i.kind = rules[0].Head.RuleKind()
	indices := newrefindices(i.isVirtual)

	// build indices for each rule.
	for idx := range rules {
		WalkRules(rules[idx], func(rule *Rule) bool {
			if rule.Default {
				i.defaultRule = rule
				return false
			}
			if i.onlyGroundRefs {
				i.onlyGroundRefs = rule.Head.Reference.IsGround()
			}
			var skip bool
			for i := range rule.Body {
				if op := rule.Body[i].OperatorTerm(); op != nil && skipIndexing.Contains(op) {
					skip = true
					break
				}
			}
			if !skip {
				for i := range rule.Body {
					indices.Update(rule, rule.Body[i])
				}
			}
			return false
		})
	}

	// build trie out of indices.
	for idx := range rules {
		var prio int
		WalkRules(rules[idx], func(rule *Rule) bool {
			if rule.Default {
				return false
			}
			node := i.root
			if indices.Indexed(rule) {
				for _, ref := range indices.Sorted() {
					node = node.Insert(ref, indices.Value(rule, ref), indices.Mapper(rule, ref))
				}
			}
			// Insert rule into trie with (insertion order, priority order)
			// tuple. Retaining the insertion order allows us to return rules
			// in the order they were passed to this function.
			node.append([...]int{idx, prio}, rule)
			prio++
			return false
		})
	}
	return true
}

func (i *baseDocEqIndex) Lookup(resolver ValueResolver) (*IndexResult, error) {
	tr := ttrPool.Get().(*trieTraversalResult)

	defer func() {
		clear(tr.unordered)
		tr.ordering = tr.ordering[:0]
		tr.multiple = false
		tr.exist = nil

		ttrPool.Put(tr)
	}()

	err := i.root.Traverse(resolver, tr)
	if err != nil {
		return nil, err
	}

	result := IndexResultPool.Get()

	result.Kind = i.kind
	result.Default = i.defaultRule
	result.OnlyGroundRefs = i.onlyGroundRefs

	if result.Rules == nil {
		result.Rules = make([]*Rule, 0, len(tr.ordering))
	} else {
		result.Rules = result.Rules[:0]
	}

	clear(result.Else)

	for _, pos := range tr.ordering {
		slices.SortFunc(tr.unordered[pos], func(a, b *ruleNode) int {
			return a.prio[1] - b.prio[1]
		})
		nodes := tr.unordered[pos]
		root := nodes[0].rule

		result.Rules = append(result.Rules, root)
		if len(nodes) > 1 {
			if result.Else == nil {
				result.Else = map[*Rule][]*Rule{}
			}

			result.Else[root] = make([]*Rule, len(nodes)-1)
			for i := 1; i < len(nodes); i++ {
				result.Else[root][i-1] = nodes[i].rule
			}
		}
	}

	if !tr.multiple {
		// even when the indexer hasn't seen multiple values, the rule itself could be one
		// where early exit shouldn't be applied.
		var lastValue Value
		for i := range result.Rules {
			if result.Rules[i].Head.DocKind() != CompleteDoc {
				tr.multiple = true
				break
			}
			if result.Rules[i].Head.Value != nil {
				if lastValue != nil && !ValueEqual(lastValue, result.Rules[i].Head.Value.Value) {
					tr.multiple = true
					break
				}
				lastValue = result.Rules[i].Head.Value.Value
			}
		}
	}

	result.EarlyExit = !tr.multiple

	return result, nil
}

func (i *baseDocEqIndex) AllRules(_ ValueResolver) (*IndexResult, error) {
	tr := newTrieTraversalResult()

	// Walk over the rule trie and accumulate _all_ rules
	rw := &ruleWalker{result: tr}
	i.root.Do(rw)

	result := NewIndexResult(i.kind)
	result.Default = i.defaultRule
	result.OnlyGroundRefs = i.onlyGroundRefs
	result.Rules = make([]*Rule, 0, len(tr.ordering))

	for _, pos := range tr.ordering {
		slices.SortFunc(tr.unordered[pos], func(a, b *ruleNode) int {
			return a.prio[1] - b.prio[1]
		})
		nodes := tr.unordered[pos]
		root := nodes[0].rule
		result.Rules = append(result.Rules, root)
		if len(nodes) > 1 {
			if result.Else == nil {
				result.Else = map[*Rule][]*Rule{}
			}

			result.Else[root] = make([]*Rule, len(nodes)-1)
			for i := 1; i < len(nodes); i++ {
				result.Else[root][i-1] = nodes[i].rule
			}
		}
	}

	result.EarlyExit = !tr.multiple

	return result, nil
}

type ruleWalker struct {
	result *trieTraversalResult
}

func (r *ruleWalker) Do(x interface{}) trieWalker {
	tn := x.(*trieNode)
	r.result.Add(tn)
	return r
}

type valueMapper struct {
	Key      string
	MapValue func(Value) Value
}

type refindex struct {
	Ref    Ref
	Value  Value
	Mapper *valueMapper
}

type refindices struct {
	isVirtual func(Ref) bool
	rules     map[*Rule][]*refindex
	frequency *util.HasherMap[Ref, int]
	sorted    []Ref
}

func newrefindices(isVirtual func(Ref) bool) *refindices {
	return &refindices{
		isVirtual: isVirtual,
		rules:     map[*Rule][]*refindex{},
		frequency: util.NewHasherMap[Ref, int](RefEqual),
	}
}

// Update attempts to update the refindices for the given expression in the
// given rule. If the expression cannot be indexed the update does not affect
// the indices.
func (i *refindices) Update(rule *Rule, expr *Expr) {

	if expr.Negated {
		return
	}

	if len(expr.With) > 0 {
		// NOTE(tsandall): In the future, we may need to consider expressions
		// that have with statements applied to them.
		return
	}

	op := expr.Operator()

	switch {
	case op.Equal(equalityRef):
		i.updateEq(rule, expr)

	case op.Equal(equalRef) && len(expr.Operands()) == 2:
		// NOTE(tsandall): if equal() is called with more than two arguments the
		// output value is being captured in which case the indexer cannot
		// exclude the rule if the equal() call would return false (because the
		// false value must still be produced.)
		i.updateEq(rule, expr)

	case op.Equal(globMatchRef) && len(expr.Operands()) == 3:
		// NOTE(sr): Same as with equal() above -- 4 operands means the output
		// of `glob.match` is captured and the rule can thus not be excluded.
		i.updateGlobMatch(rule, expr)
	}
}

// Sorted returns a sorted list of references that the indices were built from.
// References that appear more frequently in the indexed rules are ordered
// before less frequently appearing references.
func (i *refindices) Sorted() []Ref {

	if i.sorted == nil {
		counts := make([]int, 0, i.frequency.Len())
		i.sorted = make([]Ref, 0, i.frequency.Len())

		i.frequency.Iter(func(k Ref, v int) bool {
			counts = append(counts, v)
			i.sorted = append(i.sorted, k)
			return false
		})

		sort.Slice(i.sorted, func(a, b int) bool {
			if counts[a] > counts[b] {
				return true
			} else if counts[b] > counts[a] {
				return false
			}
			return i.sorted[a][0].Loc().Compare(i.sorted[b][0].Loc()) < 0
		})
	}

	return i.sorted
}

func (i *refindices) Indexed(rule *Rule) bool {
	return len(i.rules[rule]) > 0
}

func (i *refindices) Value(rule *Rule, ref Ref) Value {
	if index := i.index(rule, ref); index != nil {
		return index.Value
	}
	return nil
}

func (i *refindices) Mapper(rule *Rule, ref Ref) *valueMapper {
	if index := i.index(rule, ref); index != nil {
		return index.Mapper
	}
	return nil
}

func (i *refindices) updateEq(rule *Rule, expr *Expr) {
	a, b := expr.Operand(0), expr.Operand(1)
	args := rule.Head.Args
	if idx, ok := eqOperandsToRefAndValue(i.isVirtual, args, a, b); ok {
		i.insert(rule, idx)
		return
	}
	if idx, ok := eqOperandsToRefAndValue(i.isVirtual, args, b, a); ok {
		i.insert(rule, idx)
		return
	}
}

func (i *refindices) updateGlobMatch(rule *Rule, expr *Expr) {
	args := rule.Head.Args

	delim, ok := globDelimiterToString(expr.Operand(1))
	if !ok {
		return
	}

	if arr := globPatternToArray(expr.Operand(0), delim); arr != nil {
		// The 3rd operand of glob.match is the value to match. We assume the
		// 3rd operand was a reference that has been rewritten and bound to a
		// variable earlier in the query OR a function argument variable.
		match := expr.Operand(2)
		if _, ok := match.Value.(Var); ok {
			var ref Ref
			for _, other := range i.rules[rule] {
				if _, ok := other.Value.(Var); ok && other.Value.Compare(match.Value) == 0 {
					ref = other.Ref
				}
			}
			if ref == nil {
				for j, arg := range args {
					if arg.Equal(match) {
						ref = Ref{FunctionArgRootDocument, InternedIntNumberTerm(j)}
					}
				}
			}
			if ref != nil {
				i.insert(rule, &refindex{
					Ref:   ref,
					Value: arr.Value,
					Mapper: &valueMapper{
						Key: delim,
						MapValue: func(v Value) Value {
							if s, ok := v.(String); ok {
								return stringSliceToArray(splitStringEscaped(string(s), delim))
							}
							return v
						},
					},
				})
			}
		}
	}
}

func (i *refindices) insert(rule *Rule, index *refindex) {

	count, ok := i.frequency.Get(index.Ref)
	if !ok {
		count = 0
	}

	i.frequency.Put(index.Ref, count+1)

	for pos, other := range i.rules[rule] {
		if other.Ref.Equal(index.Ref) {
			i.rules[rule][pos] = index
			return
		}
	}

	i.rules[rule] = append(i.rules[rule], index)
}

func (i *refindices) index(rule *Rule, ref Ref) *refindex {
	for _, index := range i.rules[rule] {
		if index.Ref.Equal(ref) {
			return index
		}
	}
	return nil
}

type trieWalker interface {
	Do(x interface{}) trieWalker
}

type trieTraversalResult struct {
	unordered map[int][]*ruleNode
	ordering  []int
	exist     *Term
	multiple  bool
}

var ttrPool = sync.Pool{
	New: func() any {
		return newTrieTraversalResult()
	},
}

func newTrieTraversalResult() *trieTraversalResult {
	return &trieTraversalResult{
		unordered: map[int][]*ruleNode{},
	}
}

func (tr *trieTraversalResult) Add(t *trieNode) {
	for _, node := range t.rules {
		root := node.prio[0]
		nodes, ok := tr.unordered[root]
		if !ok {
			tr.ordering = append(tr.ordering, root)
		}
		tr.unordered[root] = append(nodes, node)
	}
	if t.multiple {
		tr.multiple = true
	}
	if tr.multiple || t.value == nil {
		return
	}
	if t.value.IsGround() && tr.exist == nil || tr.exist.Equal(t.value) {
		tr.exist = t.value
		return
	}
	tr.multiple = true
}

type trieNode struct {
	ref       Ref
	mappers   []*valueMapper
	next      *trieNode
	any       *trieNode
	undefined *trieNode
	scalars   *util.HasherMap[Value, *trieNode]
	array     *trieNode
	rules     []*ruleNode
	value     *Term
	multiple  bool
}

func (node *trieNode) String() string {
	var flags []string
	flags = append(flags, fmt.Sprintf("self:%p", node))
	if len(node.ref) > 0 {
		flags = append(flags, node.ref.String())
	}
	if node.next != nil {
		flags = append(flags, fmt.Sprintf("next:%p", node.next))
	}
	if node.any != nil {
		flags = append(flags, fmt.Sprintf("any:%p", node.any))
	}
	if node.undefined != nil {
		flags = append(flags, fmt.Sprintf("undefined:%p", node.undefined))
	}
	if node.array != nil {
		flags = append(flags, fmt.Sprintf("array:%p", node.array))
	}
	if node.scalars.Len() > 0 {
		buf := make([]string, 0, node.scalars.Len())
		node.scalars.Iter(func(key Value, val *trieNode) bool {
			buf = append(buf, fmt.Sprintf("scalar(%v):%p", key, val))
			return false
		})
		sort.Strings(buf)
		flags = append(flags, strings.Join(buf, " "))
	}
	if len(node.rules) > 0 {
		flags = append(flags, fmt.Sprintf("%d rule(s)", len(node.rules)))
	}
	if len(node.mappers) > 0 {
		flags = append(flags, fmt.Sprintf("%d mapper(s)", len(node.mappers)))
	}
	if node.value != nil {
		flags = append(flags, "value exists")
	}
	return strings.Join(flags, " ")
}

func (node *trieNode) append(prio [2]int, rule *Rule) {
	node.rules = append(node.rules, &ruleNode{prio, rule})

	if node.value != nil && rule.Head.Value != nil && !node.value.Equal(rule.Head.Value) {
		node.multiple = true
	}

	if node.value == nil && rule.Head.DocKind() == CompleteDoc {
		node.value = rule.Head.Value
	}
}

type ruleNode struct {
	prio [2]int
	rule *Rule
}

func newTrieNodeImpl() *trieNode {
	return &trieNode{
		scalars: util.NewHasherMap[Value, *trieNode](ValueEqual),
	}
}

func (node *trieNode) Do(walker trieWalker) {
	next := walker.Do(node)
	if next == nil {
		return
	}
	if node.any != nil {
		node.any.Do(next)
	}
	if node.undefined != nil {
		node.undefined.Do(next)
	}

	node.scalars.Iter(func(_ Value, child *trieNode) bool {
		child.Do(next)
		return false
	})

	if node.array != nil {
		node.array.Do(next)
	}
	if node.next != nil {
		node.next.Do(next)
	}
}

func (node *trieNode) Insert(ref Ref, value Value, mapper *valueMapper) *trieNode {

	if node.next == nil {
		node.next = newTrieNodeImpl()
		node.next.ref = ref
	}

	if mapper != nil {
		node.next.addMapper(mapper)
	}

	return node.next.insertValue(value)
}

func (node *trieNode) Traverse(resolver ValueResolver, tr *trieTraversalResult) error {

	if node == nil {
		return nil
	}

	tr.Add(node)

	return node.next.traverse(resolver, tr)
}

func (node *trieNode) addMapper(mapper *valueMapper) {
	for i := range node.mappers {
		if node.mappers[i].Key == mapper.Key {
			return
		}
	}
	node.mappers = append(node.mappers, mapper)
}

func (node *trieNode) insertValue(value Value) *trieNode {

	switch value := value.(type) {
	case nil:
		if node.undefined == nil {
			node.undefined = newTrieNodeImpl()
		}
		return node.undefined
	case Var:
		if node.any == nil {
			node.any = newTrieNodeImpl()
		}
		return node.any
	case Null, Boolean, Number, String:
		child, ok := node.scalars.Get(value)
		if !ok {
			child = newTrieNodeImpl()
			node.scalars.Put(value, child)
		}
		return child
	case *Array:
		if node.array == nil {
			node.array = newTrieNodeImpl()
		}
		return node.array.insertArray(value)
	}

	panic("illegal value")
}

func (node *trieNode) insertArray(arr *Array) *trieNode {

	if arr.Len() == 0 {
		return node
	}

	switch head := arr.Elem(0).Value.(type) {
	case Var:
		if node.any == nil {
			node.any = newTrieNodeImpl()
		}
		return node.any.insertArray(arr.Slice(1, -1))
	case Null, Boolean, Number, String:
		child, ok := node.scalars.Get(head)
		if !ok {
			child = newTrieNodeImpl()
			node.scalars.Put(head, child)
		}
		return child.insertArray(arr.Slice(1, -1))
	}

	panic("illegal value")
}

func (node *trieNode) traverse(resolver ValueResolver, tr *trieTraversalResult) error {

	if node == nil {
		return nil
	}

	v, err := resolver.Resolve(node.ref)
	if err != nil {
		if IsUnknownValueErr(err) {
			return node.traverseUnknown(resolver, tr)
		}
		return err
	}

	if node.undefined != nil {
		err = node.undefined.Traverse(resolver, tr)
		if err != nil {
			return err
		}
	}

	if v == nil {
		return nil
	}

	if node.any != nil {
		err = node.any.Traverse(resolver, tr)
		if err != nil {
			return err
		}
	}

	if err := node.traverseValue(resolver, tr, v); err != nil {
		return err
	}

	for i := range node.mappers {
		if err := node.traverseValue(resolver, tr, node.mappers[i].MapValue(v)); err != nil {
			return err
		}
	}

	return nil
}

func (node *trieNode) traverseValue(resolver ValueResolver, tr *trieTraversalResult, value Value) error {

	switch value := value.(type) {
	case *Array:
		if node.array == nil {
			return nil
		}
		return node.array.traverseArray(resolver, tr, value)

	case Null, Boolean, Number, String:
		child, ok := node.scalars.Get(value)
		if !ok {
			return nil
		}
		return child.Traverse(resolver, tr)
	}

	return nil
}

func (node *trieNode) traverseArray(resolver ValueResolver, tr *trieTraversalResult, arr *Array) error {

	if arr.Len() == 0 {
		return node.Traverse(resolver, tr)
	}

	if node.any != nil {
		err := node.any.traverseArray(resolver, tr, arr.Slice(1, -1))
		if err != nil {
			return err
		}
	}

	head := arr.Elem(0).Value

	if !IsScalar(head) {
		return nil
	}

	switch head := head.(type) {
	case Null, Boolean, Number, String:
		child, ok := node.scalars.Get(head)
		if !ok {
			return nil
		}
		return child.traverseArray(resolver, tr, arr.Slice(1, -1))
	}

	panic("illegal value")
}

func (node *trieNode) traverseUnknown(resolver ValueResolver, tr *trieTraversalResult) error {

	if node == nil {
		return nil
	}

	if err := node.Traverse(resolver, tr); err != nil {
		return err
	}

	if err := node.undefined.traverseUnknown(resolver, tr); err != nil {
		return err
	}

	if err := node.any.traverseUnknown(resolver, tr); err != nil {
		return err
	}

	if err := node.array.traverseUnknown(resolver, tr); err != nil {
		return err
	}

	var iterErr error
	node.scalars.Iter(func(_ Value, child *trieNode) bool {
		return child.traverseUnknown(resolver, tr) != nil
	})

	return iterErr
}

// If term `a` is one of the function's operands, we store a Ref: `args[0]`
// for the argument number. So for `f(x, y) { x = 10; y = 12 }`, we'll
// bind `args[0]` and `args[1]` to this rule when called for (x=10) and
// (y=12) respectively.
func eqOperandsToRefAndValue(isVirtual func(Ref) bool, args []*Term, a, b *Term) (*refindex, bool) {
	switch v := a.Value.(type) {
	case Var:
		for i, arg := range args {
			if arg.Value.Compare(a.Value) == 0 {
				if bval, ok := indexValue(b); ok {
					return &refindex{Ref: Ref{FunctionArgRootDocument, InternedIntNumberTerm(i)}, Value: bval}, true
				}
			}
		}
	case Ref:
		if !RootDocumentNames.Contains(v[0]) {
			return nil, false
		}
		if isVirtual(v) {
			return nil, false
		}
		if v.IsNested() || !v.IsGround() {
			return nil, false
		}
		if bval, ok := indexValue(b); ok {
			return &refindex{Ref: v, Value: bval}, true
		}
	}
	return nil, false
}

func indexValue(b *Term) (Value, bool) {
	switch b := b.Value.(type) {
	case Null, Boolean, Number, String, Var:
		return b, true
	case *Array:
		stop := false
		first := true
		vis := NewGenericVisitor(func(x interface{}) bool {
			if first {
				first = false
				return false
			}
			switch x.(type) {
			// No nested structures or values that require evaluation (other than var).
			case *Array, Object, Set, *ArrayComprehension, *ObjectComprehension, *SetComprehension, Ref:
				stop = true
			}
			return stop
		})
		vis.Walk(b)
		if !stop {
			return b, true
		}
	}

	return nil, false
}

func globDelimiterToString(delim *Term) (string, bool) {

	arr, ok := delim.Value.(*Array)
	if !ok {
		return "", false
	}

	var result string

	if arr.Len() == 0 {
		result = "."
	} else {
		for i := range arr.Len() {
			term := arr.Elem(i)
			s, ok := term.Value.(String)
			if !ok {
				return "", false
			}
			result += string(s)
		}
	}

	return result, true
}

var globwildcard = VarTerm("$globwildcard")

func globPatternToArray(pattern *Term, delim string) *Term {

	s, ok := pattern.Value.(String)
	if !ok {
		return nil
	}

	parts := splitStringEscaped(string(s), delim)
	arr := make([]*Term, len(parts))

	for i := range parts {
		if parts[i] == "*" {
			arr[i] = globwildcard
		} else {
			var escaped bool
			for _, c := range parts[i] {
				if c == '\\' {
					escaped = !escaped
					continue
				}
				if !escaped {
					switch c {
					case '[', '?', '{', '*':
						// TODO(tsandall): super glob and character pattern
						// matching not supported yet.
						return nil
					}
				}
				escaped = false
			}
			arr[i] = StringTerm(parts[i])
		}
	}

	return NewTerm(NewArray(arr...))
}

// splits s on characters in delim except if delim characters have been escaped
// with reverse solidus.
func splitStringEscaped(s string, delim string) []string {

	var last, curr int
	var escaped bool
	var result []string

	for ; curr < len(s); curr++ {
		if s[curr] == '\\' || escaped {
			escaped = !escaped
			continue
		}
		if strings.ContainsRune(delim, rune(s[curr])) {
			result = append(result, s[last:curr])
			last = curr + 1
		}
	}

	result = append(result, s[last:])

	return result
}

func stringSliceToArray(s []string) *Array {
	arr := make([]*Term, len(s))
	for i, v := range s {
		arr[i] = StringTerm(v)
	}
	return NewArray(arr...)
}
