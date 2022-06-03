package types

import (
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Result struct {
	Msg string `json:"msg,omitempty"`

	// Metadata includes the contents of `details` from the Rego rule signature
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// The constraint that was violated
	Constraint *unstructured.Unstructured `json:"constraint,omitempty"`

	// The violating review
	Review interface{} `json:"review,omitempty"`

	// The violating Resource, filled out by the Target
	Resource interface{}

	// The enforcement action of the constraint
	EnforcementAction string `json:"enforcementAction,omitempty"`
}

type Response struct {
	Trace   *string
	Input   *string
	Target  string
	Results []*Result
}

func (r *Response) TraceDump() string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "Target: %s\n", r.Target)
	if r.Input == nil {
		fmt.Fprintf(b, "Input: TRACING DISABLED\n\n")
	} else {
		fmt.Fprintf(b, "Input:\n%s\n\n", *r.Input)
	}
	if r.Trace == nil {
		fmt.Fprintf(b, "Trace: TRACING DISABLED\n\n")
	} else {
		fmt.Fprintf(b, "Trace:\n%s\n\n", *r.Trace)
	}
	for i, r := range r.Results {
		fmt.Fprintf(b, "Result(%d):\n%s\n\n", i, spew.Sdump(r))
	}
	return b.String()
}

func NewResponses() *Responses {
	return &Responses{
		ByTarget: make(map[string]*Response),
		Handled:  make(map[string]bool),
	}
}

type Responses struct {
	ByTarget map[string]*Response
	Handled  map[string]bool
}

func (r *Responses) Results() []*Result {
	if r == nil {
		return nil
	}
	var res []*Result
	for _, resp := range r.ByTarget {
		res = append(res, resp.Results...)
	}
	return res
}

func (r *Responses) HandledCount() int {
	if r == nil {
		return 0
	}
	c := 0
	for _, h := range r.Handled {
		if h {
			c++
		}
	}
	return c
}

func (r *Responses) TraceDump() string {
	b := &strings.Builder{}
	for _, resp := range r.ByTarget {
		fmt.Fprintln(b, resp.TraceDump())
		fmt.Fprintln(b, "")
	}
	return b.String()
}
