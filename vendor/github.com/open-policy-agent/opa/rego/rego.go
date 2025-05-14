// Copyright 2017 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// Package rego exposes high level APIs for evaluating Rego policies.
package rego

import (
	"io"
	"time"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"
	"github.com/open-policy-agent/opa/loader"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/v1/metrics"
	v1 "github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/resolver"
	"github.com/open-policy-agent/opa/v1/topdown"
	"github.com/open-policy-agent/opa/v1/topdown/builtins"
	"github.com/open-policy-agent/opa/v1/topdown/cache"
	"github.com/open-policy-agent/opa/v1/topdown/print"
	"github.com/open-policy-agent/opa/v1/tracing"
)

// CompileResult represents the result of compiling a Rego query, zero or more
// Rego modules, and arbitrary contextual data into an executable.
type CompileResult = v1.CompileResult

// PartialQueries contains the queries and support modules produced by partial
// evaluation.
type PartialQueries = v1.PartialQueries

// PartialResult represents the result of partial evaluation. The result can be
// used to generate a new query that can be run when inputs are known.
type PartialResult = v1.PartialResult

// EvalContext defines the set of options allowed to be set at evaluation
// time. Any other options will need to be set on a new Rego object.
type EvalContext = v1.EvalContext

// EvalOption defines a function to set an option on an EvalConfig
type EvalOption = v1.EvalOption

// EvalInput configures the input for a Prepared Query's evaluation
func EvalInput(input interface{}) EvalOption {
	return v1.EvalInput(input)
}

// EvalParsedInput configures the input for a Prepared Query's evaluation
func EvalParsedInput(input ast.Value) EvalOption {
	return v1.EvalParsedInput(input)
}

// EvalMetrics configures the metrics for a Prepared Query's evaluation
func EvalMetrics(metric metrics.Metrics) EvalOption {
	return v1.EvalMetrics(metric)
}

// EvalTransaction configures the Transaction for a Prepared Query's evaluation
func EvalTransaction(txn storage.Transaction) EvalOption {
	return v1.EvalTransaction(txn)
}

// EvalInstrument enables or disables instrumenting for a Prepared Query's evaluation
func EvalInstrument(instrument bool) EvalOption {
	return v1.EvalInstrument(instrument)
}

// EvalTracer configures a tracer for a Prepared Query's evaluation
// Deprecated: Use EvalQueryTracer instead.
func EvalTracer(tracer topdown.Tracer) EvalOption {
	return v1.EvalTracer(tracer)
}

// EvalQueryTracer configures a tracer for a Prepared Query's evaluation
func EvalQueryTracer(tracer topdown.QueryTracer) EvalOption {
	return v1.EvalQueryTracer(tracer)
}

// EvalPartialNamespace returns an argument that sets the namespace to use for
// partial evaluation results. The namespace must be a valid package path
// component.
func EvalPartialNamespace(ns string) EvalOption {
	return v1.EvalPartialNamespace(ns)
}

// EvalUnknowns returns an argument that sets the values to treat as
// unknown during partial evaluation.
func EvalUnknowns(unknowns []string) EvalOption {
	return v1.EvalUnknowns(unknowns)
}

// EvalDisableInlining returns an argument that adds a set of paths to exclude from
// partial evaluation inlining.
func EvalDisableInlining(paths []ast.Ref) EvalOption {
	return v1.EvalDisableInlining(paths)
}

// EvalParsedUnknowns returns an argument that sets the values to treat
// as unknown during partial evaluation.
func EvalParsedUnknowns(unknowns []*ast.Term) EvalOption {
	return v1.EvalParsedUnknowns(unknowns)
}

// EvalRuleIndexing will disable indexing optimizations for the
// evaluation. This should only be used when tracing in debug mode.
func EvalRuleIndexing(enabled bool) EvalOption {
	return v1.EvalRuleIndexing(enabled)
}

// EvalEarlyExit will disable 'early exit' optimizations for the
// evaluation. This should only be used when tracing in debug mode.
func EvalEarlyExit(enabled bool) EvalOption {
	return v1.EvalEarlyExit(enabled)
}

// EvalTime sets the wall clock time to use during policy evaluation.
// time.now_ns() calls will return this value.
func EvalTime(x time.Time) EvalOption {
	return v1.EvalTime(x)
}

// EvalSeed sets a reader that will seed randomization required by built-in functions.
// If a seed is not provided crypto/rand.Reader is used.
func EvalSeed(r io.Reader) EvalOption {
	return v1.EvalSeed(r)
}

// EvalInterQueryBuiltinCache sets the inter-query cache that built-in functions can utilize
// during evaluation.
func EvalInterQueryBuiltinCache(c cache.InterQueryCache) EvalOption {
	return v1.EvalInterQueryBuiltinCache(c)
}

// EvalInterQueryBuiltinValueCache sets the inter-query value cache that built-in functions can utilize
// during evaluation.
func EvalInterQueryBuiltinValueCache(c cache.InterQueryValueCache) EvalOption {
	return v1.EvalInterQueryBuiltinValueCache(c)
}

// EvalNDBuiltinCache sets the non-deterministic builtin cache that built-in functions can
// use during evaluation.
func EvalNDBuiltinCache(c builtins.NDBCache) EvalOption {
	return v1.EvalNDBuiltinCache(c)
}

// EvalResolver sets a Resolver for a specified ref path for this evaluation.
func EvalResolver(ref ast.Ref, r resolver.Resolver) EvalOption {
	return v1.EvalResolver(ref, r)
}

// EvalSortSets causes the evaluator to sort sets before returning them as JSON arrays.
func EvalSortSets(yes bool) EvalOption {
	return v1.EvalSortSets(yes)
}

// EvalCopyMaps causes the evaluator to copy `map[string]interface{}`s before returning them.
func EvalCopyMaps(yes bool) EvalOption {
	return v1.EvalCopyMaps(yes)
}

// EvalPrintHook sets the object to use for handling print statement outputs.
func EvalPrintHook(ph print.Hook) EvalOption {
	return v1.EvalPrintHook(ph)
}

// EvalVirtualCache sets the topdown.VirtualCache to use for evaluation. This is
// optional, and if not set, the default cache is used.
func EvalVirtualCache(vc topdown.VirtualCache) EvalOption {
	return v1.EvalVirtualCache(vc)
}

// PreparedEvalQuery holds the prepared Rego state that has been pre-processed
// for subsequent evaluations.
type PreparedEvalQuery = v1.PreparedEvalQuery

// PreparedPartialQuery holds the prepared Rego state that has been pre-processed
// for partial evaluations.
type PreparedPartialQuery = v1.PreparedPartialQuery

// Errors represents a collection of errors returned when evaluating Rego.
type Errors = v1.Errors

// IsPartialEvaluationNotEffectiveErr returns true if err is an error returned by
// this package to indicate that partial evaluation was ineffective.
func IsPartialEvaluationNotEffectiveErr(err error) bool {
	return v1.IsPartialEvaluationNotEffectiveErr(err)
}

// Rego constructs a query and can be evaluated to obtain results.
type Rego = v1.Rego

// Function represents a built-in function that is callable in Rego.
type Function = v1.Function

// BuiltinContext contains additional attributes from the evaluator that
// built-in functions can use, e.g., the request context.Context, caches, etc.
type BuiltinContext = v1.BuiltinContext

type (
	// Builtin1 defines a built-in function that accepts 1 argument.
	Builtin1 = v1.Builtin1

	// Builtin2 defines a built-in function that accepts 2 arguments.
	Builtin2 = v1.Builtin2

	// Builtin3 defines a built-in function that accepts 3 argument.
	Builtin3 = v1.Builtin3

	// Builtin4 defines a built-in function that accepts 4 argument.
	Builtin4 = v1.Builtin4

	// BuiltinDyn defines a built-in function  that accepts a list of arguments.
	BuiltinDyn = v1.BuiltinDyn
)

// RegisterBuiltin1 adds a built-in function globally inside the OPA runtime.
func RegisterBuiltin1(decl *Function, impl Builtin1) {
	v1.RegisterBuiltin1(decl, impl)
}

// RegisterBuiltin2 adds a built-in function globally inside the OPA runtime.
func RegisterBuiltin2(decl *Function, impl Builtin2) {
	v1.RegisterBuiltin2(decl, impl)
}

// RegisterBuiltin3 adds a built-in function globally inside the OPA runtime.
func RegisterBuiltin3(decl *Function, impl Builtin3) {
	v1.RegisterBuiltin3(decl, impl)
}

// RegisterBuiltin4 adds a built-in function globally inside the OPA runtime.
func RegisterBuiltin4(decl *Function, impl Builtin4) {
	v1.RegisterBuiltin4(decl, impl)
}

// RegisterBuiltinDyn adds a built-in function globally inside the OPA runtime.
func RegisterBuiltinDyn(decl *Function, impl BuiltinDyn) {
	v1.RegisterBuiltinDyn(decl, impl)
}

// Function1 returns an option that adds a built-in function to the Rego object.
func Function1(decl *Function, f Builtin1) func(*Rego) {
	return v1.Function1(decl, f)
}

// Function2 returns an option that adds a built-in function to the Rego object.
func Function2(decl *Function, f Builtin2) func(*Rego) {
	return v1.Function2(decl, f)
}

// Function3 returns an option that adds a built-in function to the Rego object.
func Function3(decl *Function, f Builtin3) func(*Rego) {
	return v1.Function3(decl, f)
}

// Function4 returns an option that adds a built-in function to the Rego object.
func Function4(decl *Function, f Builtin4) func(*Rego) {
	return v1.Function4(decl, f)
}

// FunctionDyn returns an option that adds a built-in function to the Rego object.
func FunctionDyn(decl *Function, f BuiltinDyn) func(*Rego) {
	return v1.FunctionDyn(decl, f)
}

// FunctionDecl returns an option that adds a custom-built-in function
// __declaration__. NO implementation is provided. This is used for
// non-interpreter execution envs (e.g., Wasm).
func FunctionDecl(decl *Function) func(*Rego) {
	return v1.FunctionDecl(decl)
}

// Dump returns an argument that sets the writer to dump debugging information to.
func Dump(w io.Writer) func(r *Rego) {
	return v1.Dump(w)
}

// Query returns an argument that sets the Rego query.
func Query(q string) func(r *Rego) {
	return v1.Query(q)
}

// ParsedQuery returns an argument that sets the Rego query.
func ParsedQuery(q ast.Body) func(r *Rego) {
	return v1.ParsedQuery(q)
}

// Package returns an argument that sets the Rego package on the query's
// context.
func Package(p string) func(r *Rego) {
	return v1.Package(p)
}

// ParsedPackage returns an argument that sets the Rego package on the query's
// context.
func ParsedPackage(pkg *ast.Package) func(r *Rego) {
	return v1.ParsedPackage(pkg)
}

// Imports returns an argument that adds a Rego import to the query's context.
func Imports(p []string) func(r *Rego) {
	return v1.Imports(p)
}

// ParsedImports returns an argument that adds Rego imports to the query's
// context.
func ParsedImports(imp []*ast.Import) func(r *Rego) {
	return v1.ParsedImports(imp)
}

// Input returns an argument that sets the Rego input document. Input should be
// a native Go value representing the input document.
func Input(x interface{}) func(r *Rego) {
	return v1.Input(x)
}

// ParsedInput returns an argument that sets the Rego input document.
func ParsedInput(x ast.Value) func(r *Rego) {
	return v1.ParsedInput(x)
}

// Unknowns returns an argument that sets the values to treat as unknown during
// partial evaluation.
func Unknowns(unknowns []string) func(r *Rego) {
	return v1.Unknowns(unknowns)
}

// ParsedUnknowns returns an argument that sets the values to treat as unknown
// during partial evaluation.
func ParsedUnknowns(unknowns []*ast.Term) func(r *Rego) {
	return v1.ParsedUnknowns(unknowns)
}

// DisableInlining adds a set of paths to exclude from partial evaluation inlining.
func DisableInlining(paths []string) func(r *Rego) {
	return v1.DisableInlining(paths)
}

// ShallowInlining prevents rules that depend on unknown values from being inlined.
// Rules that only depend on known values are inlined.
func ShallowInlining(yes bool) func(r *Rego) {
	return v1.ShallowInlining(yes)
}

// SkipPartialNamespace disables namespacing of partial evalution results for support
// rules generated from policy. Synthetic support rules are still namespaced.
func SkipPartialNamespace(yes bool) func(r *Rego) {
	return v1.SkipPartialNamespace(yes)
}

// PartialNamespace returns an argument that sets the namespace to use for
// partial evaluation results. The namespace must be a valid package path
// component.
func PartialNamespace(ns string) func(r *Rego) {
	return v1.PartialNamespace(ns)
}

// Module returns an argument that adds a Rego module.
func Module(filename, input string) func(r *Rego) {
	return v1.Module(filename, input)
}

// ParsedModule returns an argument that adds a parsed Rego module. If a string
// module with the same filename name is added, it will override the parsed
// module.
func ParsedModule(module *ast.Module) func(*Rego) {
	return v1.ParsedModule(module)
}

// Load returns an argument that adds a filesystem path to load data
// and Rego modules from. Any file with a *.rego, *.yaml, or *.json
// extension will be loaded. The path can be either a directory or file,
// directories are loaded recursively. The optional ignore string patterns
// can be used to filter which files are used.
// The Load option can only be used once.
// Note: Loading files will require a write transaction on the store.
func Load(paths []string, filter loader.Filter) func(r *Rego) {
	return v1.Load(paths, filter)
}

// LoadBundle returns an argument that adds a filesystem path to load
// a bundle from. The path can be a compressed bundle file or a directory
// to be loaded as a bundle.
// Note: Loading bundles will require a write transaction on the store.
func LoadBundle(path string) func(r *Rego) {
	return v1.LoadBundle(path)
}

// ParsedBundle returns an argument that adds a bundle to be loaded.
func ParsedBundle(name string, b *bundle.Bundle) func(r *Rego) {
	return v1.ParsedBundle(name, b)
}

// Compiler returns an argument that sets the Rego compiler.
func Compiler(c *ast.Compiler) func(r *Rego) {
	return v1.Compiler(c)
}

// Store returns an argument that sets the policy engine's data storage layer.
//
// If using the Load, LoadBundle, or ParsedBundle options then a transaction
// must also be provided via the Transaction() option. After loading files
// or bundles the transaction should be aborted or committed.
func Store(s storage.Store) func(r *Rego) {
	return v1.Store(s)
}

// StoreReadAST returns an argument that sets whether the store should eagerly convert data to AST values.
//
// Only applicable when no store has been set on the Rego object through the Store option.
func StoreReadAST(enabled bool) func(r *Rego) {
	return v1.StoreReadAST(enabled)
}

// Transaction returns an argument that sets the transaction to use for storage
// layer operations.
//
// Requires the store associated with the transaction to be provided via the
// Store() option. If using Load(), LoadBundle(), or ParsedBundle() options
// the transaction will likely require write params.
func Transaction(txn storage.Transaction) func(r *Rego) {
	return v1.Transaction(txn)
}

// Metrics returns an argument that sets the metrics collection.
func Metrics(m metrics.Metrics) func(r *Rego) {
	return v1.Metrics(m)
}

// Instrument returns an argument that enables instrumentation for diagnosing
// performance issues.
func Instrument(yes bool) func(r *Rego) {
	return v1.Instrument(yes)
}

// Trace returns an argument that enables tracing on r.
func Trace(yes bool) func(r *Rego) {
	return v1.Trace(yes)
}

// Tracer returns an argument that adds a query tracer to r.
// Deprecated: Use QueryTracer instead.
func Tracer(t topdown.Tracer) func(r *Rego) {
	return v1.Tracer(t)
}

// QueryTracer returns an argument that adds a query tracer to r.
func QueryTracer(t topdown.QueryTracer) func(r *Rego) {
	return v1.QueryTracer(t)
}

// Runtime returns an argument that sets the runtime data to provide to the
// evaluation engine.
func Runtime(term *ast.Term) func(r *Rego) {
	return v1.Runtime(term)
}

// Time sets the wall clock time to use during policy evaluation. Prepared queries
// do not inherit this parameter. Use EvalTime to set the wall clock time when
// executing a prepared query.
func Time(x time.Time) func(r *Rego) {
	return v1.Time(x)
}

// Seed sets a reader that will seed randomization required by built-in functions.
// If a seed is not provided crypto/rand.Reader is used.
func Seed(r io.Reader) func(*Rego) {
	return v1.Seed(r)
}

// PrintTrace is a helper function to write a human-readable version of the
// trace to the writer w.
func PrintTrace(w io.Writer, r *Rego) {
	v1.PrintTrace(w, r)
}

// PrintTraceWithLocation is a helper function to write a human-readable version of the
// trace to the writer w.
func PrintTraceWithLocation(w io.Writer, r *Rego) {
	v1.PrintTraceWithLocation(w, r)
}

// UnsafeBuiltins sets the built-in functions to treat as unsafe and not allow.
// This option is ignored for module compilation if the caller supplies the
// compiler. This option is always honored for query compilation. Provide an
// empty (non-nil) map to disable checks on queries.
func UnsafeBuiltins(unsafeBuiltins map[string]struct{}) func(r *Rego) {
	return v1.UnsafeBuiltins(unsafeBuiltins)
}

// SkipBundleVerification skips verification of a signed bundle.
func SkipBundleVerification(yes bool) func(r *Rego) {
	return v1.SkipBundleVerification(yes)
}

// InterQueryBuiltinCache sets the inter-query cache that built-in functions can utilize
// during evaluation.
func InterQueryBuiltinCache(c cache.InterQueryCache) func(r *Rego) {
	return v1.InterQueryBuiltinCache(c)
}

// InterQueryBuiltinValueCache sets the inter-query value cache that built-in functions can utilize
// during evaluation.
func InterQueryBuiltinValueCache(c cache.InterQueryValueCache) func(r *Rego) {
	return v1.InterQueryBuiltinValueCache(c)
}

// NDBuiltinCache sets the non-deterministic builtins cache.
func NDBuiltinCache(c builtins.NDBCache) func(r *Rego) {
	return v1.NDBuiltinCache(c)
}

// StrictBuiltinErrors tells the evaluator to treat all built-in function errors as fatal errors.
func StrictBuiltinErrors(yes bool) func(r *Rego) {
	return v1.StrictBuiltinErrors(yes)
}

// BuiltinErrorList supplies an error slice to store built-in function errors.
func BuiltinErrorList(list *[]topdown.Error) func(r *Rego) {
	return v1.BuiltinErrorList(list)
}

// Resolver sets a Resolver for a specified ref path.
func Resolver(ref ast.Ref, r resolver.Resolver) func(r *Rego) {
	return v1.Resolver(ref, r)
}

// Schemas sets the schemaSet
func Schemas(x *ast.SchemaSet) func(r *Rego) {
	return v1.Schemas(x)
}

// Capabilities configures the underlying compiler's capabilities.
// This option is ignored for module compilation if the caller supplies the
// compiler.
func Capabilities(c *ast.Capabilities) func(r *Rego) {
	return v1.Capabilities(c)
}

// Target sets the runtime to exercise.
func Target(t string) func(r *Rego) {
	return v1.Target(t)
}

// GenerateJSON sets the AST to JSON converter for the results.
func GenerateJSON(f func(*ast.Term, *EvalContext) (interface{}, error)) func(r *Rego) {
	return v1.GenerateJSON(f)
}

// PrintHook sets the object to use for handling print statement outputs.
func PrintHook(h print.Hook) func(r *Rego) {
	return v1.PrintHook(h)
}

// DistributedTracingOpts sets the options to be used by distributed tracing.
func DistributedTracingOpts(tr tracing.Options) func(r *Rego) {
	return v1.DistributedTracingOpts(tr)
}

// EnablePrintStatements enables print() calls. If this option is not provided,
// print() calls will be erased from the policy. This option only applies to
// queries and policies that passed as raw strings, i.e., this function will not
// have any affect if the caller supplies the ast.Compiler instance.
func EnablePrintStatements(yes bool) func(r *Rego) {
	return v1.EnablePrintStatements(yes)
}

// Strict enables or disables strict-mode in the compiler
func Strict(yes bool) func(r *Rego) {
	return v1.Strict(yes)
}

func SetRegoVersion(version ast.RegoVersion) func(r *Rego) {
	return v1.SetRegoVersion(version)
}

// New returns a new Rego object.
func New(options ...func(r *Rego)) *Rego {
	opts := make([]func(r *Rego), 0, len(options)+1)
	opts = append(opts, options...)
	opts = append(opts, func(r *Rego) {
		if r.RegoVersion() == ast.RegoUndefined {
			SetRegoVersion(ast.DefaultRegoVersion)(r)
		}
	})

	return v1.New(opts...)
}

// CompileOption defines a function to set options on Compile calls.
type CompileOption = v1.CompileOption

// CompileContext contains options for Compile calls.
type CompileContext = v1.CompileContext

// CompilePartial defines an option to control whether partial evaluation is run
// before the query is planned and compiled.
func CompilePartial(yes bool) CompileOption {
	return v1.CompilePartial(yes)
}

// PrepareOption defines a function to set an option to control
// the behavior of the Prepare call.
type PrepareOption = v1.PrepareOption

// PrepareConfig holds settings to control the behavior of the
// Prepare call.
type PrepareConfig = v1.PrepareConfig

// WithPartialEval configures an option for PrepareForEval
// which will have it perform partial evaluation while preparing
// the query (similar to rego.Rego#PartialResult)
func WithPartialEval() PrepareOption {
	return v1.WithPartialEval()
}

// WithNoInline adds a set of paths to exclude from partial evaluation inlining.
func WithNoInline(paths []string) PrepareOption {
	return v1.WithNoInline(paths)
}

// WithBuiltinFuncs carries the rego.Function{1,2,3} per-query function definitions
// to the target plugins.
func WithBuiltinFuncs(bis map[string]*topdown.Builtin) PrepareOption {
	return v1.WithBuiltinFuncs(bis)
}
