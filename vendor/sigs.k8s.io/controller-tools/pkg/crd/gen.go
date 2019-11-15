/*
Copyright 2018 The Kubernetes Authors.

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

package crd

import (
	"fmt"
	"go/types"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"sigs.k8s.io/controller-tools/pkg/version"
)

// +controllertools:marker:generateHelp

// Generator generates CustomResourceDefinition objects.
type Generator struct {
	// TrivialVersions indicates that we should produce a single-version CRD.
	//
	// Single "trivial-version" CRDs are compatible with older (pre 1.13)
	// Kubernetes API servers.  The storage version's schema will be used as
	// the CRD's schema.
	TrivialVersions bool `marker:",optional"`

	// MaxDescLen specifies the maximum description length for fields in CRD's OpenAPI schema.
	//
	// 0 indicates drop the description for all fields completely.
	// n indicates limit the description to at most n characters and truncate the description to
	// closest sentence boundary if it exceeds n characters.
	MaxDescLen *int `marker:",optional"`
}

func (Generator) RegisterMarkers(into *markers.Registry) error {
	return crdmarkers.Register(into)
}
func (g Generator) Generate(ctx *genall.GenerationContext) error {
	parser := &Parser{
		Collector: ctx.Collector,
		Checker:   ctx.Checker,
	}

	AddKnownTypes(parser)
	for _, root := range ctx.Roots {
		parser.NeedPackage(root)
	}

	metav1Pkg := FindMetav1(ctx.Roots)
	if metav1Pkg == nil {
		// no objects in the roots, since nothing imported metav1
		return nil
	}

	// TODO: allow selecting a specific object
	kubeKinds := FindKubeKinds(parser, metav1Pkg)
	if len(kubeKinds) == 0 {
		// no objects in the roots
		return nil
	}

	for _, groupKind := range kubeKinds {
		parser.NeedCRDFor(groupKind, g.MaxDescLen)
		crd := parser.CustomResourceDefinitions[groupKind]
		if g.TrivialVersions {
			toTrivialVersions(&crd)
		}
		addAttribution(&crd)
		fileName := fmt.Sprintf("%s_%s.yaml", crd.Spec.Group, crd.Spec.Names.Plural)
		if err := ctx.WriteYAML(fileName, crd); err != nil {
			return err
		}
	}

	return nil
}

// toTrivialVersions strips out all schemata except for the storage schema,
// and moves that up into the root object.  This makes the CRD compatible
// with pre 1.13 clusters.
func toTrivialVersions(crd *apiext.CustomResourceDefinition) {
	var canonicalSchema *apiext.CustomResourceValidation
	var canonicalSubresources *apiext.CustomResourceSubresources
	var canonicalColumns []apiext.CustomResourceColumnDefinition
	for i, ver := range crd.Spec.Versions {
		if ver.Storage == true {
			canonicalSchema = ver.Schema
			canonicalSubresources = ver.Subresources
			canonicalColumns = ver.AdditionalPrinterColumns
		}
		crd.Spec.Versions[i].Schema = nil
		crd.Spec.Versions[i].Subresources = nil
		crd.Spec.Versions[i].AdditionalPrinterColumns = nil
	}
	if canonicalSchema == nil {
		return
	}

	crd.Spec.Validation = canonicalSchema
	crd.Spec.Subresources = canonicalSubresources
	crd.Spec.AdditionalPrinterColumns = canonicalColumns
}

// addAttribution adds attribution info to indicate controller-gen tool was used
// to generate this CRD definition along with the version info.
func addAttribution(crd *apiext.CustomResourceDefinition) {
	if crd.ObjectMeta.Annotations == nil {
		crd.ObjectMeta.Annotations = map[string]string{}
	}
	crd.ObjectMeta.Annotations["controller-gen.kubebuilder.io/version"] = version.Version()
}

// FindMetav1 locates the actual package representing metav1 amongst
// the imports of the roots.
func FindMetav1(roots []*loader.Package) *loader.Package {
	for _, root := range roots {
		pkg := root.Imports()["k8s.io/apimachinery/pkg/apis/meta/v1"]
		if pkg != nil {
			return pkg
		}
	}
	return nil
}

// FindKubeKinds locates all types that contain TypeMeta and ObjectMeta
// (and thus may be a Kubernetes object), and returns the corresponding
// group-kinds.
func FindKubeKinds(parser *Parser, metav1Pkg *loader.Package) []schema.GroupKind {
	// TODO(directxman12): technically, we should be finding metav1 per-package
	var kubeKinds []schema.GroupKind
	for typeIdent, info := range parser.Types {
		hasObjectMeta := false
		hasTypeMeta := false

		pkg := typeIdent.Package
		pkg.NeedTypesInfo()
		typesInfo := pkg.TypesInfo

		for _, field := range info.Fields {
			if field.Name != "" {
				// type and object meta are embedded,
				// so they can't be this
				continue
			}

			fieldType := typesInfo.TypeOf(field.RawField.Type)
			namedField, isNamed := fieldType.(*types.Named)
			if !isNamed {
				// ObjectMeta and TypeMeta are named types
				continue
			}
			if namedField.Obj().Pkg() == nil {
				// Embedded non-builtin universe type (specifically, it's probably `error`),
				// so it can't be ObjectMeta or TypeMeta
				continue
			}
			fieldPkgPath := loader.NonVendorPath(namedField.Obj().Pkg().Path())
			fieldPkg := pkg.Imports()[fieldPkgPath]
			if fieldPkg != metav1Pkg {
				continue
			}

			switch namedField.Obj().Name() {
			case "ObjectMeta":
				hasObjectMeta = true
			case "TypeMeta":
				hasTypeMeta = true
			}
		}

		if !hasObjectMeta || !hasTypeMeta {
			continue
		}

		groupKind := schema.GroupKind{
			Group: parser.GroupVersions[pkg].Group,
			Kind:  typeIdent.Name,
		}
		kubeKinds = append(kubeKinds, groupKind)
	}

	return kubeKinds
}
