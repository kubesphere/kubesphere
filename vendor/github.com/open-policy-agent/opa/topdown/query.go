package topdown

import (
	"context"
	"sort"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/metrics"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/topdown/builtins"
	"github.com/open-policy-agent/opa/topdown/copypropagation"
)

// QueryResultSet represents a collection of results returned by a query.
type QueryResultSet []QueryResult

// QueryResult represents a single result returned by a query. The result
// contains bindings for all variables that appear in the query.
type QueryResult map[ast.Var]*ast.Term

// Query provides a configurable interface for performing query evaluation.
type Query struct {
	cancel           Cancel
	query            ast.Body
	queryCompiler    ast.QueryCompiler
	compiler         *ast.Compiler
	store            storage.Store
	txn              storage.Transaction
	input            *ast.Term
	tracers          []Tracer
	unknowns         []*ast.Term
	partialNamespace string
	metrics          metrics.Metrics
	instr            *Instrumentation
	disableInlining  []ast.Ref
	genvarprefix     string
	runtime          *ast.Term
	builtins         map[string]*Builtin
	indexing         bool
}

// Builtin represents a built-in function that queries can call.
type Builtin struct {
	Decl *ast.Builtin
	Func BuiltinFunc
}

// NewQuery returns a new Query object that can be run.
func NewQuery(query ast.Body) *Query {
	return &Query{
		query:        query,
		genvarprefix: ast.WildcardPrefix,
		indexing:     true,
	}
}

// WithQueryCompiler sets the queryCompiler used for the query.
func (q *Query) WithQueryCompiler(queryCompiler ast.QueryCompiler) *Query {
	q.queryCompiler = queryCompiler
	return q
}

// WithCompiler sets the compiler to use for the query.
func (q *Query) WithCompiler(compiler *ast.Compiler) *Query {
	q.compiler = compiler
	return q
}

// WithStore sets the store to use for the query.
func (q *Query) WithStore(store storage.Store) *Query {
	q.store = store
	return q
}

// WithTransaction sets the transaction to use for the query. All queries
// should be performed over a consistent snapshot of the storage layer.
func (q *Query) WithTransaction(txn storage.Transaction) *Query {
	q.txn = txn
	return q
}

// WithCancel sets the cancellation object to use for the query. Set this if
// you need to abort queries based on a deadline. This is optional.
func (q *Query) WithCancel(cancel Cancel) *Query {
	q.cancel = cancel
	return q
}

// WithInput sets the input object to use for the query. References rooted at
// input will be evaluated against this value. This is optional.
func (q *Query) WithInput(input *ast.Term) *Query {
	q.input = input
	return q
}

// WithTracer adds a query tracer to use during evaluation. This is optional.
func (q *Query) WithTracer(tracer Tracer) *Query {
	q.tracers = append(q.tracers, tracer)
	return q
}

// WithMetrics sets the metrics collection to add evaluation metrics to. This
// is optional.
func (q *Query) WithMetrics(m metrics.Metrics) *Query {
	q.metrics = m
	return q
}

// WithInstrumentation sets the instrumentation configuration to enable on the
// evaluation process. By default, instrumentation is turned off.
func (q *Query) WithInstrumentation(instr *Instrumentation) *Query {
	q.instr = instr
	return q
}

// WithUnknowns sets the initial set of variables or references to treat as
// unknown during query evaluation. This is required for partial evaluation.
func (q *Query) WithUnknowns(terms []*ast.Term) *Query {
	q.unknowns = terms
	return q
}

// WithPartialNamespace sets the namespace to use for supporting rules
// generated as part of the partial evaluation process. The ns value must be a
// valid package path component.
func (q *Query) WithPartialNamespace(ns string) *Query {
	q.partialNamespace = ns
	return q
}

// WithDisableInlining adds a set of paths to the query that should be excluded from
// inlining. Inlining during partial evaluation can be expensive in some cases
// (e.g., when a cross-product is computed.) Disabling inlining avoids expensive
// computation at the cost of generating support rules.
func (q *Query) WithDisableInlining(paths []ast.Ref) *Query {
	q.disableInlining = paths
	return q
}

// WithRuntime sets the runtime data to execute the query with. The runtime data
// can be returned by the `opa.runtime` built-in function.
func (q *Query) WithRuntime(runtime *ast.Term) *Query {
	q.runtime = runtime
	return q
}

// WithBuiltins adds a set of built-in functions that can be called by the
// query.
func (q *Query) WithBuiltins(builtins map[string]*Builtin) *Query {
	q.builtins = builtins
	return q
}

// WithIndexing will enable or disable using rule indexing for the evaluation
// of the query. The default is enabled.
func (q *Query) WithIndexing(enabled bool) *Query {
	q.indexing = enabled
	return q
}

// PartialRun executes partial evaluation on the query with respect to unknown
// values. Partial evaluation attempts to evaluate as much of the query as
// possible without requiring values for the unknowns set on the query. The
// result of partial evaluation is a new set of queries that can be evaluated
// once the unknown value is known. In addition to new queries, partial
// evaluation may produce additional support modules that should be used in
// conjunction with the partially evaluated queries.
func (q *Query) PartialRun(ctx context.Context) (partials []ast.Body, support []*ast.Module, err error) {
	if q.partialNamespace == "" {
		q.partialNamespace = "partial" // lazily initialize partial namespace
	}
	f := &queryIDFactory{}
	b := newBindings(0, q.instr)
	e := &eval{
		ctx:           ctx,
		cancel:        q.cancel,
		query:         q.query,
		queryCompiler: q.queryCompiler,
		queryIDFact:   f,
		queryID:       f.Next(),
		bindings:      b,
		compiler:      q.compiler,
		store:         q.store,
		baseCache:     newBaseCache(),
		targetStack:   newRefStack(),
		txn:           q.txn,
		input:         q.input,
		tracers:       q.tracers,
		instr:         q.instr,
		builtins:      q.builtins,
		builtinCache:  builtins.Cache{},
		virtualCache:  newVirtualCache(),
		saveSet:       newSaveSet(q.unknowns, b, q.instr),
		saveStack:     newSaveStack(),
		saveSupport:   newSaveSupport(),
		saveNamespace: ast.StringTerm(q.partialNamespace),
		genvarprefix:  q.genvarprefix,
		runtime:       q.runtime,
		indexing:      q.indexing,
	}

	if len(q.disableInlining) > 0 {
		e.disableInlining = [][]ast.Ref{q.disableInlining}
	}

	e.caller = e
	q.startTimer(metrics.RegoPartialEval)
	defer q.stopTimer(metrics.RegoPartialEval)

	livevars := ast.NewVarSet()

	ast.WalkVars(q.query, func(x ast.Var) bool {
		if !x.IsGenerated() {
			livevars.Add(x)
		}
		return false
	})

	p := copypropagation.New(livevars)

	err = e.Run(func(e *eval) error {

		// Build output from saved expressions.
		body := ast.NewBody()

		for _, elem := range e.saveStack.Stack[len(e.saveStack.Stack)-1] {
			body.Append(elem.Plug(e.bindings))
		}

		// Include bindings as exprs so that when caller evals the result, they
		// can obtain values for the vars in their query.
		bindingExprs := []*ast.Expr{}
		e.bindings.Iter(e.bindings, func(a, b *ast.Term) error {
			bindingExprs = append(bindingExprs, ast.Equality.Expr(a, b))
			return nil
		})

		// Sort binding expressions so that results are deterministic.
		sort.Slice(bindingExprs, func(i, j int) bool {
			return bindingExprs[i].Compare(bindingExprs[j]) < 0
		})

		for i := range bindingExprs {
			body.Append(bindingExprs[i])
		}

		partials = append(partials, applyCopyPropagation(p, e.instr, body))
		return nil
	})

	support = e.saveSupport.List()

	return partials, support, err
}

// Run is a wrapper around Iter that accumulates query results and returns them
// in one shot.
func (q *Query) Run(ctx context.Context) (QueryResultSet, error) {
	qrs := QueryResultSet{}
	return qrs, q.Iter(ctx, func(qr QueryResult) error {
		qrs = append(qrs, qr)
		return nil
	})
}

// Iter executes the query and invokes the iter function with query results
// produced by evaluating the query.
func (q *Query) Iter(ctx context.Context, iter func(QueryResult) error) error {
	f := &queryIDFactory{}
	e := &eval{
		ctx:           ctx,
		cancel:        q.cancel,
		query:         q.query,
		queryCompiler: q.queryCompiler,
		queryIDFact:   f,
		queryID:       f.Next(),
		bindings:      newBindings(0, q.instr),
		compiler:      q.compiler,
		store:         q.store,
		baseCache:     newBaseCache(),
		targetStack:   newRefStack(),
		txn:           q.txn,
		input:         q.input,
		tracers:       q.tracers,
		instr:         q.instr,
		builtins:      q.builtins,
		builtinCache:  builtins.Cache{},
		virtualCache:  newVirtualCache(),
		genvarprefix:  q.genvarprefix,
		runtime:       q.runtime,
		indexing:      q.indexing,
	}
	e.caller = e
	q.startTimer(metrics.RegoQueryEval)
	err := e.Run(func(e *eval) error {
		qr := QueryResult{}
		e.bindings.Iter(nil, func(k, v *ast.Term) error {
			qr[k.Value.(ast.Var)] = v
			return nil
		})
		return iter(qr)
	})
	q.stopTimer(metrics.RegoQueryEval)
	return err
}

func (q *Query) startTimer(name string) {
	if q.metrics != nil {
		q.metrics.Timer(name).Start()
	}
}

func (q *Query) stopTimer(name string) {
	if q.metrics != nil {
		q.metrics.Timer(name).Stop()
	}
}
