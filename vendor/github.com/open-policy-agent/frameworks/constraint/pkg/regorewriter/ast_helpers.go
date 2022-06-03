package regorewriter

import (
	"github.com/golang/glog"
	"github.com/open-policy-agent/opa/ast"
	"github.com/pkg/errors"
)

var (
	dataVarTerm    = ast.VarTerm("data")
	inputRefPrefix = ast.MustParseRef("input")
)

// isDataRef returns true if the ast.Ref is referring to the "data" document.
func isDataRef(ref ast.Ref) bool {
	if len(ref) == 0 {
		return false
	}

	firstTerm := ref[0]
	return firstTerm.Equal(dataVarTerm)
}

// isSubRef returns true if sub is contained within base.
func isSubRef(base, sub ast.Ref) bool {
	glog.V(vLogDetail).Infof("Subref check %s %s", base, sub)
	if len(sub) < len(base) {
		return false
	}
	return base.Equal(sub[0:len(base)])
}

// packagesAsRefs parses a list of refs in the form "data.foo.bar" into ast.Ref values.
func packagesAsRefs(strs []string) ([]ast.Ref, error) {
	var refs []ast.Ref
	for _, s := range strs {
		ref, err := ast.ParseRef(s)
		if err != nil {
			return nil, err
		}
		if len(ref) == 0 {
			return nil, errors.Errorf("invalid ref input %s", s)
		}
		if !dataVarTerm.Equal(ref[0]) {
			return nil, errors.Wrapf(err, "ref must start with data")
		}
		refs = append(refs, ref)
	}
	return refs, nil
}
