package regorewriter

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/golang/glog"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/format"
	"github.com/pkg/errors"
)

const vLog = 2
const vLogDetail = vLog + 1

// RegoRewriter rewrites rego code by updating library package paths by prepending a prefix
// and updating references to library code accordingly.
type RegoRewriter struct {
	// entryPoints are the files that contain a violation rule which serves as the entry point for a
	// constraint template.  These sources will not have their package path updated, but refs to any
	// libs will be updated accordingly.
	entryPoints []*Module
	// libs are the library files that will have their package paths updated and have refs updated
	// as well.
	libs []*Module
	// testData are files that are found in the 'test' directory and should not be
	testData []*TestData
	// packageTransform is the transform that modifies the package path and refs.
	packageTransform PackageTransformer
	// allowedLibPrefixes are the allowed package path prefixes for libs, for example "data.lib".
	allowedLibPrefixes []ast.Ref
	// allowedExterns are the allowed external references for entryPoints/libs, for example "data.inventory"
	allowedExterns []ast.Ref
}

// New returns a new RegoRewriter
// args:
// 	it - the PackageTransformer that will be used for updating the path
//	libs - a list of package prefixes that are allowed for library use
//	externs - a list of packages that the rego is allowed to reference but not declared in any libs
func New(pt PackageTransformer, libs []string, externs []string) (*RegoRewriter, error) {
	externRefs, err := packagesAsRefs(externs)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse externs")
	}
	libRefs, err := packagesAsRefs(libs)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse libs")
	}

	return &RegoRewriter{
		packageTransform:   pt,
		allowedLibPrefixes: libRefs,
		allowedExterns:     externRefs,
	}, nil
}

// add is the internal method for parsing a module and entering it into the bookkeeping.
func (r *RegoRewriter) add(path, src string, slice *[]*Module) error {
	m, err := ast.ParseModule(path, src)
	if err != nil {
		return err
	}

	r.addModule(path, m, slice)
	return nil
}

func (r *RegoRewriter) addModule(path string, m *ast.Module, slice *[]*Module) {
	*slice = append(*slice, &Module{FilePath: FilePath{path}, Module: m})
}

func (r *RegoRewriter) AddEntryPointModule(path string, m *ast.Module) {
	r.addModule(path, m, &r.entryPoints)
}

// AddEntryPoint adds a base source which will not have it's package path rewritten.  These correspond
// to the rego that will be populated into a ConstraintTemplate with the 'violation' rule.
func (r *RegoRewriter) AddEntryPoint(path, src string) error {
	return r.add(path, src, &r.entryPoints)
}

// AddLib adds a library source which will have the package path updated.
func (r *RegoRewriter) AddLib(path, src string) error {
	return r.add(path, src, &r.libs)
}

// addTestDir adds a test dir inside one of the provided paths.
func (r *RegoRewriter) addTestDir(testDirPath string) error {
	glog.V(vLog).Infof("Walking test dir %s", testDirPath)
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "walk error on path %s", path)
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".json") && !strings.HasSuffix(path, ".yaml") {
			return nil
		}

		glog.V(vLog).Infof("reading %s", path)
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		r.testData = append(r.testData, &TestData{FilePath: FilePath{path: path}, content: bytes})
		return nil
	}
	if err := filepath.Walk(testDirPath, walkFn); err != nil {
		return err
	}
	return nil
}

// addFileFromFs reads a file from the filesystem, parses it then appends it to slice.
func (r *RegoRewriter) addFileFromFs(path string, slice *[]*Module) error {
	glog.V(vLog).Infof("adding file %s", path)
	if !strings.HasSuffix(path, ".rego") {
		return errors.Errorf("invalid file specified %s", path)
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return r.add(path, string(bytes), slice)
}

// addPathFromFs adds a module from the local filesystem.
// Loading from the filesystem is based on how "opa test" operates in terms of scoping.
// 1. the 'test' directory must exist as a member of one of the paths passed to 'opa test'.
// 2. the '.rego' source can exist anywhere in the subtree of the specified path
// 3. any '.rego' loaded by "opa test" can reference any "test" data member that is loaded by
//    opa test, for example, if "opa test foo/ bar/" is specified, a test in foo/ can see test data
//    from bar/test/.
func (r *RegoRewriter) addPathFromFs(path string, slice *[]*Module) error {
	fileStat, err := os.Stat(path)
	if err != nil {
		return err
	}

	if fileStat.IsDir() {
		infos, err := ioutil.ReadDir(path)
		if err != nil {
			return err
		}

		// handle test dirs
		for _, info := range infos {
			if info.IsDir() && info.Name() == "test" {
				// load test data
				if err := r.addTestDir(filepath.Join(path, info.Name())); err != nil {
					return err
				}
			}
		}

		walkFn := func(path string, info os.FileInfo, err error) error {
			if info.IsDir() || !strings.HasSuffix(path, ".rego") {
				return nil
			}
			return r.addFileFromFs(path, slice)
		}
		if err := filepath.Walk(path, walkFn); err != nil {
			return err
		}
		return nil
	}

	return r.addFileFromFs(path, slice)
}

// AddBaseFromFs adds a base source which will not have it's package path rewritten.  These correspond
// to the rego that will be populated into a ConstraintTemplate with the 'violation' rule.
func (r *RegoRewriter) AddBaseFromFs(path string) error {
	return r.addPathFromFs(path, &r.entryPoints)
}

// AddLibFromFs adds a library source which will have the package path updated.
func (r *RegoRewriter) AddLibFromFs(path string) error {
	return r.addPathFromFs(path, &r.libs)
}

// forAllModules runs f on all rego modules (both entrypoints and libraries)
func (r *RegoRewriter) forAllModules(f func(*Module) error) error {
	for _, m := range r.libs {
		if err := f(m); err != nil {
			return err
		}
	}
	for _, m := range r.entryPoints {
		if err := f(m); err != nil {
			return err
		}
	}
	return nil
}

// checkImports checks that the imports for all sources are referencing a known lib or a declared
// extern.
func (r *RegoRewriter) checkImports() error {
	return r.forAllModules(func(m *Module) error {
		if m.IsTestFile() {
			glog.V(vLog).Infof("skipping import check for %s", m.FilePath)
			return nil
		}

		glog.V(vLogDetail).Infof("checking %s", m.FilePath)
		for _, i := range m.Module.Imports {
			if err := r.checkImport(i); err != nil {
				return err
			}
		}
		return nil
	})
}

// checkLibPackages validates that defined lib packages are within the allowed libs.
func (r *RegoRewriter) checkLibPackages() error {
	for _, mod := range r.libs {
		path := mod.Module.Package.Path
		if !r.allowedLibPackage(path) {
			return errors.Errorf("path %s not found in lib prefixes", path)
		}
	}
	return nil
}

// allowedLibPackage returns true if the lib package is an allowed package name which is
// defined as a subref of any allowed lib prefix (note that it cannot be exactly the lib prefix).
func (r *RegoRewriter) allowedLibPackage(ref ast.Ref) bool {
	for _, libRef := range r.allowedLibPrefixes {
		if libRef.Equal(ref) {
			return false
		}
		if ref.HasPrefix(libRef) {
			return true
		}
	}

	return false
}

// checkRef will check that a ref is allowed based on externs and known libs.
func (r *RegoRewriter) checkRef(ref ast.Ref) error {
	glog.V(vLogDetail).Infof("  Checking ref %s", ref)
	if !isDataRef(ref) {
		return nil
	}

	for _, extern := range r.allowedExterns {
		if isSubRef(extern, ref) {
			glog.V(vLogDetail).Infof("Found extern ref %s for %s", extern, ref)
			return nil
		}
	}

	for _, lib := range r.allowedLibPrefixes {
		if isSubRef(lib, ref) {
			glog.V(vLogDetail).Infof("Found lib ref %s for %s", lib, ref)
			return nil
		}
	}

	return errors.Errorf("disallowed ref %s", ref)
}

// checkImport checks the import statement to ensure that it's a subref of an allowed lib prefix.
func (r *RegoRewriter) checkImport(i *ast.Import) error {
	want := i.Path.String()
	glog.V(vLog).Infof("checking import %s", want)

	importRef := i.Path.Value.(ast.Ref)
	if isSubRef(inputRefPrefix, importRef) {
		return errors.Errorf("bad import")
	}

	for _, libPrefix := range r.allowedLibPrefixes {
		if isSubRef(libPrefix, importRef) {
			return nil
		}
	}

	return errors.Errorf("bad import")
}

// checkDataReferences checks that all data references are directed to allowed lib prefixes or
// externs.
func (r *RegoRewriter) checkDataReferences() error {
	// walk AST, look for data references
	return r.forAllModules(func(m *Module) error {
		if m.IsTestFile() {
			glog.V(vLogDetail).Infof("skipping check data references for %s", m.FilePath)
			return nil
		}

		glog.V(vLogDetail).Infof("checking data references for %s", m.FilePath)
		var errs Errors
		for _, rule := range m.Module.Rules {
			ast.WalkRefs(rule, func(ref ast.Ref) bool {
				if err := r.checkRef(ref); err != nil {
					errs = append(errs, err)
				}
				return true
			})
		}
		if errs != nil {
			return errors.Wrapf(errs, "check refs failed on module %s", m.FilePath)
		}
		return nil
	})
}

// checkSources runs all checks on the rego sources.
func (r *RegoRewriter) checkSources() error {
	if err := r.checkLibPackages(); err != nil {
		return err
	}
	if err := r.checkImports(); err != nil {
		return err
	}
	return r.checkDataReferences()
}

// refNeedsRewrite checks if the Ref refers to the 'data' element.
func (r *RegoRewriter) refNeedsRewrite(ref ast.Ref) bool {
	if !isDataRef(ref) {
		return false
	}
	for _, extRef := range r.allowedExterns {
		if isSubRef(extRef, ref) {
			return false
		}
	}

	glog.V(1).Infof("ref needs rewrite: %s   %#v", ref, ref)
	for _, t := range ref {
		glog.V(3).Infof("  term: %s %#v %#v", t, t, reflect.TypeOf(t.Value).String())
	}
	return true
}

// rewriteDataRef will update a data ref based on the import transform.
func (r *RegoRewriter) rewriteDataRef(ref ast.Ref) ast.Ref {
	if !r.refNeedsRewrite(ref) {
		return ref
	}
	return r.packageTransform.Transform(ref)
}

// rewriteImportPath updates an import path to the new value.
func (r *RegoRewriter) rewriteImportPath(path *ast.Term) *ast.Term {
	pathRef := path.Value.(ast.Ref)
	if !r.refNeedsRewrite(pathRef) {
		return path
	}
	return ast.NewTerm(r.rewriteDataRef(pathRef))
}

// Rewrite will check the input source and update the package paths and refs as appropriate.
func (r *RegoRewriter) Rewrite() (*Sources, error) {
	if err := r.checkSources(); err != nil {
		return nil, err
	}

	// libs - update package
	for _, l := range r.libs {
		l.Module.Package.Path = r.rewriteDataRef(l.Module.Package.Path)
	}

	// libs, entryPoints - update import and other refs
	err := r.forAllModules(func(mod *Module) error {
		for _, i := range mod.Module.Imports {
			glog.V(vLogDetail).Infof("import: %s %#v", i.Path, i.Path)
			for _, t := range i.Path.Value.(ast.Ref) {
				glog.V(vLogDetail).Infof("  term: %s %#v %#v", t, t, reflect.TypeOf(t.Value).String())
			}
			i.Path = r.rewriteImportPath(i.Path)
		}

		for _, rule := range mod.Module.Rules {
			ast.WalkTerms(rule, func(term *ast.Term) bool {
				if ref, ok := term.Value.(ast.Ref); ok {
					term.Value = r.rewriteDataRef(ref)
				}
				return true
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// write updated modules
	err = r.forAllModules(func(mod *Module) error {
		b, err := format.Ast(mod.Module)
		if err != nil {
			return err
		}
		glog.V(2).Infof("Formatted rego:\n%s\n", b)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Sources{EntryPoints: r.entryPoints, Libs: r.libs, TestData: r.testData}, nil
}
