// Copyright 2016 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"strings"

	iStrs "github.com/open-policy-agent/opa/internal/strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
)

const (
	minLocationWidth      = 5 // len("query")
	maxIdealLocationWidth = 64
	columnPadding         = 4
	maxExprVarWidth       = 32
	maxPrettyExprVarWidth = 64
)

// Op defines the types of tracing events.
type Op string

const (
	// EnterOp is emitted when a new query is about to be evaluated.
	EnterOp Op = "Enter"

	// ExitOp is emitted when a query has evaluated to true.
	ExitOp Op = "Exit"

	// EvalOp is emitted when an expression is about to be evaluated.
	EvalOp Op = "Eval"

	// RedoOp is emitted when an expression, rule, or query is being re-evaluated.
	RedoOp Op = "Redo"

	// SaveOp is emitted when an expression is saved instead of evaluated
	// during partial evaluation.
	SaveOp Op = "Save"

	// FailOp is emitted when an expression evaluates to false.
	FailOp Op = "Fail"

	// DuplicateOp is emitted when a query has produced a duplicate value. The search
	// will stop at the point where the duplicate was emitted and backtrack.
	DuplicateOp Op = "Duplicate"

	// NoteOp is emitted when an expression invokes a tracing built-in function.
	NoteOp Op = "Note"

	// IndexOp is emitted during an expression evaluation to represent lookup
	// matches.
	IndexOp Op = "Index"

	// WasmOp is emitted when resolving a ref using an external
	// Resolver.
	WasmOp Op = "Wasm"

	// UnifyOp is emitted when two terms are unified.  Node will be set to an
	// equality expression with the two terms.  This Node will not have location
	// info.
	UnifyOp           Op = "Unify"
	FailedAssertionOp Op = "FailedAssertion"
)

// VarMetadata provides some user facing information about
// a variable in some policy.
type VarMetadata struct {
	Name     ast.Var       `json:"name"`
	Location *ast.Location `json:"location"`
}

// Event contains state associated with a tracing event.
type Event struct {
	Op            Op                      // Identifies type of event.
	Node          ast.Node                // Contains AST node relevant to the event.
	Location      *ast.Location           // The location of the Node this event relates to.
	QueryID       uint64                  // Identifies the query this event belongs to.
	ParentID      uint64                  // Identifies the parent query this event belongs to.
	Locals        *ast.ValueMap           // Contains local variable bindings from the query context. Nil if variables were not included in the trace event.
	LocalMetadata map[ast.Var]VarMetadata // Contains metadata for the local variable bindings. Nil if variables were not included in the trace event.
	Message       string                  // Contains message for Note events.
	Ref           *ast.Ref                // Identifies the subject ref for the event. Only applies to Index and Wasm operations.

	input                     *ast.Term
	bindings                  *bindings
	localVirtualCacheSnapshot *ast.ValueMap
}

func (evt *Event) WithInput(input *ast.Term) *Event {
	evt.input = input
	return evt
}

// HasRule returns true if the Event contains an ast.Rule.
func (evt *Event) HasRule() bool {
	_, ok := evt.Node.(*ast.Rule)
	return ok
}

// HasBody returns true if the Event contains an ast.Body.
func (evt *Event) HasBody() bool {
	_, ok := evt.Node.(ast.Body)
	return ok
}

// HasExpr returns true if the Event contains an ast.Expr.
func (evt *Event) HasExpr() bool {
	_, ok := evt.Node.(*ast.Expr)
	return ok
}

// Equal returns true if this event is equal to the other event.
func (evt *Event) Equal(other *Event) bool {
	if evt.Op != other.Op {
		return false
	}
	if evt.QueryID != other.QueryID {
		return false
	}
	if evt.ParentID != other.ParentID {
		return false
	}
	if !evt.equalNodes(other) {
		return false
	}
	return evt.Locals.Equal(other.Locals)
}

func (evt *Event) String() string {
	return fmt.Sprintf("%v %v %v (qid=%v, pqid=%v)", evt.Op, evt.Node, evt.Locals, evt.QueryID, evt.ParentID)
}

// Input returns the input object as it was at the event.
func (evt *Event) Input() *ast.Term {
	return evt.input
}

// Plug plugs event bindings into the provided ast.Term. Because bindings are mutable, this only makes sense to do when
// the event is emitted rather than on recorded trace events as the bindings are going to be different by then.
func (evt *Event) Plug(term *ast.Term) *ast.Term {
	return evt.bindings.Plug(term)
}

func (evt *Event) equalNodes(other *Event) bool {
	switch a := evt.Node.(type) {
	case ast.Body:
		if b, ok := other.Node.(ast.Body); ok {
			return a.Equal(b)
		}
	case *ast.Rule:
		if b, ok := other.Node.(*ast.Rule); ok {
			return a.Equal(b)
		}
	case *ast.Expr:
		if b, ok := other.Node.(*ast.Expr); ok {
			return a.Equal(b)
		}
	case nil:
		return other.Node == nil
	}
	return false
}

// Tracer defines the interface for tracing in the top-down evaluation engine.
// Deprecated: Use QueryTracer instead.
type Tracer interface {
	Enabled() bool
	Trace(*Event)
}

// QueryTracer defines the interface for tracing in the top-down evaluation engine.
// The implementation can provide additional configuration to modify the tracing
// behavior for query evaluations.
type QueryTracer interface {
	Enabled() bool
	TraceEvent(Event)
	Config() TraceConfig
}

// TraceConfig defines some common configuration for Tracer implementations
type TraceConfig struct {
	PlugLocalVars bool // Indicate whether to plug local variable bindings before calling into the tracer.
}

// legacyTracer Implements the QueryTracer interface by wrapping an older Tracer instance.
type legacyTracer struct {
	t Tracer
}

func (l *legacyTracer) Enabled() bool {
	return l.t.Enabled()
}

func (l *legacyTracer) Config() TraceConfig {
	return TraceConfig{
		PlugLocalVars: true, // For backwards compatibility old tracers will plug local variables
	}
}

func (l *legacyTracer) TraceEvent(evt Event) {
	l.t.Trace(&evt)
}

// WrapLegacyTracer will create a new QueryTracer which wraps an
// older Tracer instance.
func WrapLegacyTracer(tracer Tracer) QueryTracer {
	return &legacyTracer{t: tracer}
}

// BufferTracer implements the Tracer and QueryTracer interface by
// simply buffering all events received.
type BufferTracer []*Event

// NewBufferTracer returns a new BufferTracer.
func NewBufferTracer() *BufferTracer {
	return &BufferTracer{}
}

// Enabled always returns true if the BufferTracer is instantiated.
func (b *BufferTracer) Enabled() bool {
	return b != nil
}

// Trace adds the event to the buffer.
// Deprecated: Use TraceEvent instead.
func (b *BufferTracer) Trace(evt *Event) {
	*b = append(*b, evt)
}

// TraceEvent adds the event to the buffer.
func (b *BufferTracer) TraceEvent(evt Event) {
	*b = append(*b, &evt)
}

// Config returns the Tracers standard configuration
func (b *BufferTracer) Config() TraceConfig {
	return TraceConfig{PlugLocalVars: true}
}

// PrettyTrace pretty prints the trace to the writer.
func PrettyTrace(w io.Writer, trace []*Event) {
	PrettyTraceWithOpts(w, trace, PrettyTraceOptions{})
}

// PrettyTraceWithLocation prints the trace to the writer and includes location information
func PrettyTraceWithLocation(w io.Writer, trace []*Event) {
	PrettyTraceWithOpts(w, trace, PrettyTraceOptions{Locations: true})
}

type PrettyTraceOptions struct {
	Locations      bool // Include location information
	ExprVariables  bool // Include variables found in the expression
	LocalVariables bool // Include all local variables
}

type traceRow []string

func (r *traceRow) add(s string) {
	*r = append(*r, s)
}

type traceTable struct {
	rows      []traceRow
	maxWidths []int
}

func (t *traceTable) add(row traceRow) {
	t.rows = append(t.rows, row)
	for i := range row {
		if i >= len(t.maxWidths) {
			t.maxWidths = append(t.maxWidths, len(row[i]))
		} else if len(row[i]) > t.maxWidths[i] {
			t.maxWidths[i] = len(row[i])
		}
	}
}

func (t *traceTable) write(w io.Writer, padding int) {
	for _, row := range t.rows {
		for i, cell := range row {
			width := t.maxWidths[i] + padding
			if i < len(row)-1 {
				_, _ = fmt.Fprintf(w, "%-*s ", width, cell)
			} else {
				_, _ = fmt.Fprintf(w, "%s", cell)
			}
		}
		_, _ = fmt.Fprintln(w)
	}
}

func PrettyTraceWithOpts(w io.Writer, trace []*Event, opts PrettyTraceOptions) {
	depths := depths{}

	// FIXME: Can we shorten each location as we process each trace event instead of beforehand?
	filePathAliases, _ := getShortenedFileNames(trace)

	table := traceTable{}

	for _, event := range trace {
		depth := depths.GetOrSet(event.QueryID, event.ParentID)
		row := traceRow{}

		if opts.Locations {
			location := formatLocation(event, filePathAliases)
			row.add(location)
		}

		row.add(formatEvent(event, depth))

		if opts.ExprVariables {
			vars := exprLocalVars(event)
			keys := sortedKeys(vars)

			buf := new(bytes.Buffer)
			buf.WriteString("{")
			for i, k := range keys {
				if i > 0 {
					buf.WriteString(", ")
				}
				_, _ = fmt.Fprintf(buf, "%v: %s", k, iStrs.Truncate(vars.Get(k).String(), maxExprVarWidth))
			}
			buf.WriteString("}")
			row.add(buf.String())
		}

		if opts.LocalVariables {
			if locals := event.Locals; locals != nil {
				keys := sortedKeys(locals)

				buf := new(bytes.Buffer)
				buf.WriteString("{")
				for i, k := range keys {
					if i > 0 {
						buf.WriteString(", ")
					}
					_, _ = fmt.Fprintf(buf, "%v: %s", k, iStrs.Truncate(locals.Get(k).String(), maxExprVarWidth))
				}
				buf.WriteString("}")
				row.add(buf.String())
			} else {
				row.add("{}")
			}
		}

		table.add(row)
	}

	table.write(w, columnPadding)
}

func sortedKeys(vm *ast.ValueMap) []ast.Value {
	keys := make([]ast.Value, 0, vm.Len())
	vm.Iter(func(k, _ ast.Value) bool {
		keys = append(keys, k)
		return false
	})
	slices.SortFunc(keys, func(a, b ast.Value) int {
		return strings.Compare(a.String(), b.String())
	})
	return keys
}

func exprLocalVars(e *Event) *ast.ValueMap {
	vars := ast.NewValueMap()

	findVars := func(term *ast.Term) bool {
		//if r, ok := term.Value.(ast.Ref); ok {
		//	fmt.Printf("ref: %v\n", r)
		//	//return true
		//}
		if name, ok := term.Value.(ast.Var); ok {
			if meta, ok := e.LocalMetadata[name]; ok {
				if val := e.Locals.Get(name); val != nil {
					vars.Put(meta.Name, val)
				}
			}
		}
		return false
	}

	if r, ok := e.Node.(*ast.Rule); ok {
		// We're only interested in vars in the head, not the body
		ast.WalkTerms(r.Head, findVars)
		return vars
	}

	// The local cache snapshot only contains a snapshot for those refs present in the event node,
	// so they can all be added to the vars map.
	e.localVirtualCacheSnapshot.Iter(func(k, v ast.Value) bool {
		vars.Put(k, v)
		return false
	})

	ast.WalkTerms(e.Node, findVars)

	return vars
}

func formatEvent(event *Event, depth int) string {
	padding := formatEventPadding(event, depth)
	if event.Op == NoteOp {
		return fmt.Sprintf("%v%v %q", padding, event.Op, event.Message)
	}

	var details interface{}
	if node, ok := event.Node.(*ast.Rule); ok {
		details = node.Path()
	} else if event.Ref != nil {
		details = event.Ref
	} else {
		details = rewrite(event).Node
	}

	template := "%v%v %v"
	opts := []interface{}{padding, event.Op, details}

	if event.Message != "" {
		template += " %v"
		opts = append(opts, event.Message)
	}

	return fmt.Sprintf(template, opts...)
}

func formatEventPadding(event *Event, depth int) string {
	spaces := formatEventSpaces(event, depth)
	if spaces > 1 {
		return strings.Repeat("| ", spaces-1)
	}
	return ""
}

func formatEventSpaces(event *Event, depth int) int {
	switch event.Op {
	case EnterOp:
		return depth
	case RedoOp:
		if _, ok := event.Node.(*ast.Expr); !ok {
			return depth
		}
	}
	return depth + 1
}

// getShortenedFileNames will return a map of file paths to shortened aliases
// that were found in the trace. It also returns the longest location expected
func getShortenedFileNames(trace []*Event) (map[string]string, int) {
	// Get a deduplicated list of all file paths
	// and the longest file path size
	fpAliases := map[string]string{}
	var canShorten []string
	longestLocation := 0
	for _, event := range trace {
		if event.Location != nil {
			if event.Location.File != "" {
				// length of "<name>:<row>"
				curLen := len(event.Location.File) + numDigits10(event.Location.Row) + 1
				if curLen > longestLocation {
					longestLocation = curLen
				}

				if _, ok := fpAliases[event.Location.File]; ok {
					continue
				}

				canShorten = append(canShorten, event.Location.File)

				// Default to just alias their full path
				fpAliases[event.Location.File] = event.Location.File
			} else {
				// length of "<min width>:<row>"
				curLen := minLocationWidth + numDigits10(event.Location.Row) + 1
				if curLen > longestLocation {
					longestLocation = curLen
				}
			}
		}
	}

	if len(canShorten) > 0 && longestLocation > maxIdealLocationWidth {
		fpAliases, longestLocation = iStrs.TruncateFilePaths(maxIdealLocationWidth, longestLocation, canShorten...)
	}

	return fpAliases, longestLocation
}

func numDigits10(n int) int {
	if n < 10 {
		return 1
	}
	return numDigits10(n/10) + 1
}

func formatLocation(event *Event, fileAliases map[string]string) string {

	location := event.Location
	if location == nil {
		return ""
	}

	if location.File == "" {
		return fmt.Sprintf("query:%v", location.Row)
	}

	return fmt.Sprintf("%v:%v", fileAliases[location.File], location.Row)
}

// depths is a helper for computing the depth of an event. Events within the
// same query all have the same depth. The depth of query is
// depth(parent(query))+1.
type depths map[uint64]int

func (ds depths) GetOrSet(qid uint64, pqid uint64) int {
	depth := ds[qid]
	if depth == 0 {
		depth = ds[pqid]
		depth++
		ds[qid] = depth
	}
	return depth
}

func builtinTrace(bctx BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {

	str, err := builtins.StringOperand(operands[0].Value, 1)
	if err != nil {
		return handleBuiltinErr(ast.Trace.Name, bctx.Location, err)
	}

	if !bctx.TraceEnabled {
		return iter(ast.BooleanTerm(true))
	}

	evt := Event{
		Op:       NoteOp,
		Location: bctx.Location,
		QueryID:  bctx.QueryID,
		ParentID: bctx.ParentID,
		Message:  string(str),
	}

	for i := range bctx.QueryTracers {
		bctx.QueryTracers[i].TraceEvent(evt)
	}

	return iter(ast.BooleanTerm(true))
}

func rewrite(event *Event) *Event {

	cpy := *event

	var node ast.Node

	switch v := event.Node.(type) {
	case *ast.Expr:
		expr := v.Copy()

		// Hide generated local vars in 'key' position that have not been
		// rewritten.
		if ev, ok := v.Terms.(*ast.Every); ok {
			if kv, ok := ev.Key.Value.(ast.Var); ok {
				if rw, ok := cpy.LocalMetadata[kv]; !ok || rw.Name.IsGenerated() {
					expr.Terms.(*ast.Every).Key = nil
				}
			}
		}
		node = expr
	case ast.Body:
		node = v.Copy()
	case *ast.Rule:
		node = v.Copy()
	}

	_, _ = ast.TransformVars(node, func(v ast.Var) (ast.Value, error) {
		if meta, ok := cpy.LocalMetadata[v]; ok {
			return meta.Name, nil
		}
		return v, nil
	})

	cpy.Node = node

	return &cpy
}

type varInfo struct {
	VarMetadata
	val     ast.Value
	exprLoc *ast.Location
	col     int // 0-indexed column
}

func (v varInfo) Value() string {
	if v.val != nil {
		return v.val.String()
	}
	return "undefined"
}

func (v varInfo) Title() string {
	if v.exprLoc != nil && v.exprLoc.Text != nil {
		return string(v.exprLoc.Text)
	}
	return string(v.Name)
}

func padLocationText(loc *ast.Location) string {
	if loc == nil {
		return ""
	}

	text := string(loc.Text)

	if loc.Col == 0 {
		return text
	}

	buf := new(bytes.Buffer)
	j := 0
	for i := 1; i < loc.Col; i++ {
		if len(loc.Tabs) > 0 && j < len(loc.Tabs) && loc.Tabs[j] == i {
			buf.WriteString("\t")
			j++
		} else {
			buf.WriteString(" ")
		}
	}

	buf.WriteString(text)
	return buf.String()
}

type PrettyEventOpts struct {
	PrettyVars bool
}

func walkTestTerms(x interface{}, f func(*ast.Term) bool) {
	var vis *ast.GenericVisitor
	vis = ast.NewGenericVisitor(func(x interface{}) bool {
		switch x := x.(type) {
		case ast.Call:
			for _, t := range x[1:] {
				vis.Walk(t)
			}
			return true
		case *ast.Expr:
			if x.IsCall() {
				for _, o := range x.Operands() {
					vis.Walk(o)
				}
				for i := range x.With {
					vis.Walk(x.With[i])
				}
				return true
			}
		case *ast.Term:
			return f(x)
		case *ast.With:
			vis.Walk(x.Value)
			return true
		}
		return false
	})
	vis.Walk(x)
}

func PrettyEvent(w io.Writer, e *Event, opts PrettyEventOpts) error {
	if !opts.PrettyVars {
		_, _ = fmt.Fprintln(w, padLocationText(e.Location))
		return nil
	}

	buf := new(bytes.Buffer)
	exprVars := map[string]varInfo{}

	findVars := func(unknownAreUndefined bool) func(term *ast.Term) bool {
		return func(term *ast.Term) bool {
			if term.Location == nil {
				return false
			}

			switch v := term.Value.(type) {
			case *ast.ArrayComprehension, *ast.SetComprehension, *ast.ObjectComprehension:
				// we don't report on the internals of a comprehension, as it's already evaluated, and we won't have the local vars.
				return true
			case ast.Var:
				var info *varInfo
				if meta, ok := e.LocalMetadata[v]; ok {
					info = &varInfo{
						VarMetadata: meta,
						val:         e.Locals.Get(v),
						exprLoc:     term.Location,
					}
				} else if unknownAreUndefined {
					info = &varInfo{
						VarMetadata: VarMetadata{Name: v},
						exprLoc:     term.Location,
						col:         term.Location.Col,
					}
				}

				if info != nil {
					if v, exists := exprVars[info.Title()]; !exists || v.val == nil {
						if term.Location != nil {
							info.col = term.Location.Col
						}
						exprVars[info.Title()] = *info
					}
				}
			}
			return false
		}
	}

	expr, ok := e.Node.(*ast.Expr)
	if !ok || expr == nil {
		return nil
	}

	base := expr.BaseCogeneratedExpr()
	exprText := padLocationText(base.Location)
	buf.WriteString(exprText)

	e.localVirtualCacheSnapshot.Iter(func(k, v ast.Value) bool {
		var info *varInfo
		switch k := k.(type) {
		case ast.Ref:
			info = &varInfo{
				VarMetadata: VarMetadata{Name: ast.Var(k.String())},
				val:         v,
				exprLoc:     k[0].Location,
				col:         k[0].Location.Col,
			}
		case *ast.ArrayComprehension:
			info = &varInfo{
				VarMetadata: VarMetadata{Name: ast.Var(k.String())},
				val:         v,
				exprLoc:     k.Term.Location,
				col:         k.Term.Location.Col,
			}
		case *ast.SetComprehension:
			info = &varInfo{
				VarMetadata: VarMetadata{Name: ast.Var(k.String())},
				val:         v,
				exprLoc:     k.Term.Location,
				col:         k.Term.Location.Col,
			}
		case *ast.ObjectComprehension:
			info = &varInfo{
				VarMetadata: VarMetadata{Name: ast.Var(k.String())},
				val:         v,
				exprLoc:     k.Key.Location,
				col:         k.Key.Location.Col,
			}
		}

		if info != nil {
			exprVars[info.Title()] = *info
		}

		return false
	})

	// If the expression is negated, we can't confidently assert that vars with unknown values are 'undefined',
	// since the compiler might have opted out of the necessary rewrite.
	walkTestTerms(expr, findVars(!expr.Negated))
	coExprs := expr.CogeneratedExprs()
	for _, coExpr := range coExprs {
		// Only the current "co-expr" can have undefined vars, if we don't know the value for a var in any other co-expr,
		// it's unknown, not undefined. A var can be unknown if it hasn't been assigned a value yet, because the co-expr
		// hasn't been evaluated yet (the fail happened before it).
		walkTestTerms(coExpr, findVars(false))
	}

	printPrettyVars(buf, exprVars)
	_, _ = fmt.Fprint(w, buf.String())
	return nil
}

func printPrettyVars(w *bytes.Buffer, exprVars map[string]varInfo) {
	containsTabs := false
	varRows := make(map[int]interface{})
	for _, info := range exprVars {
		if len(info.exprLoc.Tabs) > 0 {
			containsTabs = true
		}
		varRows[info.exprLoc.Row] = nil
	}

	if containsTabs && len(varRows) > 1 {
		// We can't (currently) reliably point to var locations when they are on different rows that contain tabs.
		// So we'll just print them in alphabetical order instead.
		byName := make([]varInfo, 0, len(exprVars))
		for _, info := range exprVars {
			byName = append(byName, info)
		}
		slices.SortStableFunc(byName, func(a, b varInfo) int {
			return strings.Compare(a.Title(), b.Title())
		})

		w.WriteString("\n\nWhere:\n")
		for _, info := range byName {
			w.WriteString(fmt.Sprintf("\n%s: %s", info.Title(), iStrs.Truncate(info.Value(), maxPrettyExprVarWidth)))
		}

		return
	}

	byCol := make([]varInfo, 0, len(exprVars))
	for _, info := range exprVars {
		byCol = append(byCol, info)
	}
	slices.SortFunc(byCol, func(a, b varInfo) int {
		// sort first by column, then by reverse row (to present vars in the same order they appear in the expr)
		if a.col == b.col {
			if a.exprLoc.Row == b.exprLoc.Row {
				return strings.Compare(a.Title(), b.Title())
			}
			return b.exprLoc.Row - a.exprLoc.Row
		}
		return a.col - b.col
	})

	if len(byCol) == 0 {
		return
	}

	w.WriteString("\n")
	printArrows(w, byCol, -1)
	for i := len(byCol) - 1; i >= 0; i-- {
		w.WriteString("\n")
		printArrows(w, byCol, i)
	}
}

func printArrows(w *bytes.Buffer, l []varInfo, printValueAt int) {
	prevCol := 0
	var slice []varInfo
	if printValueAt >= 0 {
		slice = l[:printValueAt+1]
	} else {
		slice = l
	}
	isFirst := true
	for i, info := range slice {

		isLast := i >= len(slice)-1
		col := info.col

		if !isLast && col == l[i+1].col {
			// We're sharing the same column with another, subsequent var
			continue
		}

		spaces := col - 1
		if i > 0 && !isFirst {
			spaces = (col - prevCol) - 1
		}

		for j := 0; j < spaces; j++ {
			tab := false
			for _, t := range info.exprLoc.Tabs {
				if t == j+prevCol+1 {
					w.WriteString("\t")
					tab = true
					break
				}
			}
			if !tab {
				w.WriteString(" ")
			}
		}

		if isLast && printValueAt >= 0 {
			valueStr := iStrs.Truncate(info.Value(), maxPrettyExprVarWidth)
			if (i > 0 && col == l[i-1].col) || (i < len(l)-1 && col == l[i+1].col) {
				// There is another var on this column, so we need to include the name to differentiate them.
				w.WriteString(fmt.Sprintf("%s: %s", info.Title(), valueStr))
			} else {
				w.WriteString(valueStr)
			}
		} else {
			w.WriteString("|")
		}
		prevCol = col
		isFirst = false
	}
}

func init() {
	RegisterBuiltinFunc(ast.Trace.Name, builtinTrace)
}
