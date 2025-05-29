/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package applyconfiguration

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/code-generator/cmd/applyconfiguration-gen/args"
	"k8s.io/code-generator/cmd/applyconfiguration-gen/generators"

	"k8s.io/gengo/v2"
	"k8s.io/gengo/v2/generator"
	"k8s.io/gengo/v2/parser"

	kerrors "k8s.io/apimachinery/pkg/util/errors"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

// Based on deepcopy gen but with legacy marker support removed.

var (
	isCRDMarker      = markers.Must(markers.MakeDefinition("kubebuilder:resource", markers.DescribesType, crdmarkers.Resource{}))
	enablePkgMarker  = markers.Must(markers.MakeDefinition("kubebuilder:ac:generate", markers.DescribesPackage, false))
	outputPkgMarker  = markers.Must(markers.MakeDefinition("kubebuilder:ac:output:package", markers.DescribesPackage, ""))
	enableTypeMarker = markers.Must(markers.MakeDefinition("kubebuilder:ac:generate", markers.DescribesType, false))
)

const defaultOutputPackage = "applyconfiguration"

// +controllertools:marker:generateHelp

// Generator generates code containing apply configuration type implementations.
type Generator struct {
	// HeaderFile specifies the header text (e.g. license) to prepend to generated files.
	HeaderFile string `marker:",optional"`
}

func (Generator) CheckFilter() loader.NodeFilter {
	return func(node ast.Node) bool {
		// ignore interfaces
		_, isIface := node.(*ast.InterfaceType)
		return !isIface
	}
}

func (Generator) RegisterMarkers(into *markers.Registry) error {
	if err := markers.RegisterAll(into,
		isCRDMarker, enablePkgMarker, enableTypeMarker, outputPkgMarker); err != nil {
		return err
	}

	into.AddHelp(isCRDMarker,
		markers.SimpleHelp("apply", "enables apply configuration generation for this type"))
	into.AddHelp(
		enableTypeMarker, markers.SimpleHelp("apply", "overrides enabling or disabling applyconfiguration generation for the type, can be used to generate applyconfiguration for a single type when the package generation is disabled, or to disable generation for a single type when the package generation is enabled"))
	into.AddHelp(
		enablePkgMarker, markers.SimpleHelp("apply", "overrides enabling or disabling applyconfiguration generation for the package"))
	into.AddHelp(
		outputPkgMarker, markers.SimpleHelp("apply", "overrides the default output package for the applyconfiguration generation, supports relative paths to the API directory. The default value is \"applyconfiguration\""))
	return nil
}

func enabledOnPackage(col *markers.Collector, pkg *loader.Package) (bool, error) {
	pkgMarkers, err := markers.PackageMarkers(col, pkg)
	if err != nil {
		return false, err
	}
	pkgMarker := pkgMarkers.Get(enablePkgMarker.Name)
	if pkgMarker != nil {
		return pkgMarker.(bool), nil
	}
	return false, nil
}

func enabledOnType(info *markers.TypeInfo) bool {
	if typeMarker := info.Markers.Get(enableTypeMarker.Name); typeMarker != nil {
		return typeMarker.(bool)
	}
	return isCRD(info)
}

func outputPkg(col *markers.Collector, pkg *loader.Package) string {
	pkgMarkers, err := markers.PackageMarkers(col, pkg)
	if err != nil {
		// Use the default when there's an error.
		return defaultOutputPackage
	}

	pkgMarker := pkgMarkers.Get(outputPkgMarker.Name)
	if pkgMarker != nil {
		return pkgMarker.(string)
	}

	return defaultOutputPackage
}

func isCRD(info *markers.TypeInfo) bool {
	objectEnabled := info.Markers.Get(isCRDMarker.Name)
	return objectEnabled != nil
}

func (d Generator) Generate(ctx *genall.GenerationContext) error {
	headerFilePath := d.HeaderFile

	if headerFilePath == "" {
		tmpFile, err := os.CreateTemp("", "applyconfig-header-*.txt")
		if err != nil {
			return fmt.Errorf("failed to create temporary file: %w", err)
		}
		if err := tmpFile.Close(); err != nil {
			return fmt.Errorf("failed to close temporary file: %w", err)
		}

		defer os.Remove(tmpFile.Name())

		headerFilePath = tmpFile.Name()
	}

	objGenCtx := ObjectGenCtx{
		Collector:      ctx.Collector,
		Checker:        ctx.Checker,
		HeaderFilePath: headerFilePath,
	}

	errs := []error{}
	for _, pkg := range ctx.Roots {
		if err := objGenCtx.generateForPackage(pkg); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return kerrors.NewAggregate(errs)
	}

	return nil
}

// ObjectGenCtx contains the common info for generating apply configuration implementations.
// It mostly exists so that generating for a package can be easily tested without
// requiring a full set of output rules, etc.
type ObjectGenCtx struct {
	Collector      *markers.Collector
	Checker        *loader.TypeChecker
	HeaderFilePath string
}

// generateForPackage generates apply configuration implementations for
// types in the given package, writing the formatted result to given writer.
func (ctx *ObjectGenCtx) generateForPackage(root *loader.Package) error {
	enabled, _ := enabledOnPackage(ctx.Collector, root)
	if !enabled {
		return nil
	}
	if len(root.GoFiles) == 0 {
		return nil
	}

	arguments := args.New()
	arguments.GoHeaderFile = ctx.HeaderFilePath

	outpkg := outputPkg(ctx.Collector, root)

	arguments.OutputDir = filepath.Join(root.Dir, outpkg)
	arguments.OutputPkg = filepath.Join(root.Package.PkgPath, outpkg)

	// The following code is based on gengo/v2.Execute.
	// We have lifted it from there so that we can adjust the markers on the types to make sure
	// that Kubebuilder generation markers are converted into the genclient marker
	// prior to executing the targets.
	buildTags := []string{gengo.StdBuildTag}
	p := parser.NewWithOptions(parser.Options{BuildTags: buildTags})
	if err := p.LoadPackages(root.PkgPath); err != nil {
		return fmt.Errorf("failed making a parser: %w", err)
	}

	c, err := generator.NewContext(p, generators.NameSystems(), generators.DefaultNameSystem())
	if err != nil {
		return fmt.Errorf("failed making a context: %w", err)
	}

	pkg, ok := c.Universe[root.PkgPath]
	if !ok {
		return fmt.Errorf("package %q not found in universe", root.Name)
	}

	// For each type we think should be generated, make sure it has a genclient
	// marker else the apply generator will not generate it.
	if err := markers.EachType(ctx.Collector, root, func(info *markers.TypeInfo) {
		if !enabledOnType(info) {
			return
		}

		typ, ok := pkg.Types[info.Name]
		if !ok {
			return
		}

		comments := sets.NewString(typ.CommentLines...)
		comments.Insert(typ.SecondClosestCommentLines...)

		if !comments.Has("// +genclient") {
			typ.CommentLines = append(typ.CommentLines, "+genclient")
		}
	}); err != nil {
		return err
	}

	targets := generators.GetTargets(c, arguments)
	if err := c.ExecuteTargets(targets); err != nil {
		return fmt.Errorf("failed executing generator: %w", err)
	}

	return nil
}
