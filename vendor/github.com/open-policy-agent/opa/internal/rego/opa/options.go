package opa

import (
	"io"
	"time"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/metrics"
	"github.com/open-policy-agent/opa/v1/topdown/builtins"
	"github.com/open-policy-agent/opa/v1/topdown/cache"
	"github.com/open-policy-agent/opa/v1/topdown/print"
)

// Result holds the evaluation result.
type Result struct {
	Result []byte
}

// EvalOpts define options for performing an evaluation.
type EvalOpts struct {
	Input                       *interface{}
	Metrics                     metrics.Metrics
	Entrypoint                  int32
	Time                        time.Time
	Seed                        io.Reader
	InterQueryBuiltinCache      cache.InterQueryCache
	InterQueryBuiltinValueCache cache.InterQueryValueCache
	NDBuiltinCache              builtins.NDBCache
	PrintHook                   print.Hook
	Capabilities                *ast.Capabilities
}
