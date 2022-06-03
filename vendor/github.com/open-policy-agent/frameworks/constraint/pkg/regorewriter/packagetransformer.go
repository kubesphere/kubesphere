package regorewriter

import (
	"strings"

	"github.com/open-policy-agent/opa/ast"
)

// PackageTransformer takes a package path and transforms it to the new package path it will be
// re-written to.
type PackageTransformer interface {
	// Transform returns a modified ast.Ref with an updated package path.
	Transform(ref ast.Ref) ast.Ref
}

// PackagePrefixer is an implementation of PackageTransformer that prepends a prefix to the package
// path after the 'data' reference, for example, if prefix is specified as ["x", "y", "z"], then
// the path "data.whatever.checkpolicy" will be updated to "data.x.y.z.whatever.checkpolicy".
type PackagePrefixer struct {
	// The prefix to prepend to the package path.
	prefix []string
}

// NewPackagePrefixer returns a new PackagePrefixer
func NewPackagePrefixer(pkgPrefix string) *PackagePrefixer {
	return &PackagePrefixer{
		prefix: strings.Split(pkgPrefix, "."),
	}
}

// Transform implements PackageTransformer
func (p *PackagePrefixer) Transform(ref ast.Ref) ast.Ref {
	var newRef ast.Ref
	newRef = append(newRef, ref[0])
	for _, prefix := range p.prefix {
		newRef = append(newRef, ast.StringTerm(prefix))
	}
	newRef = append(newRef, ref[1:]...)
	return newRef
}
