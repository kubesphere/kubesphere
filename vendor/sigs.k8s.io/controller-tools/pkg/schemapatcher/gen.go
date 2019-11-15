/*
Copyright 2019 The Kubernetes Authors.

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

package schemapatcher

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v3"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kyaml "sigs.k8s.io/yaml"

	crdgen "sigs.k8s.io/controller-tools/pkg/crd"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/markers"
	yamlop "sigs.k8s.io/controller-tools/pkg/schemapatcher/internal/yaml"
)

// NB(directxman12): this code is quite fragile, but there are a sufficient
// number of corner cases that it's hard to decompose into separate tools.
// When in doubt, ping @sttts.
//
// Namely:
// - It needs to only update existing versions
// - It needs to make "stable" changes that don't mess with map key ordering
//   (in order to facilitate validating that no change has occurred)
// - It needs to collapse identical schema versions into a top-level schema,
//   if all versions are identical (this is a common requirement to all CRDs,
//   but in this case it means simple jsonpatch wouldn't suffice)

// TODO(directxman12): When CRD v1 rolls around, consider splitting this into a
// tool that generates a patch, and a separate tool for applying stable YAML
// patches.

// Generator patches existing CRDs with new schemata.
//
// For single-version CRDs, it will simply replace the global schema.
//
// For multi-version CRDs, it will replace schemata of existing versions
// and *clear the schema* from any versions not specified in the Go code.
// It will *not* add new versions, or remove old ones.
//
// For multi-version CRDs with identical schemata, it will take care of
// lifting the per-version schema up to the global schema.
type Generator struct {
	// ManifestsPath contains the CustomResourceDefinition YAML files.
	ManifestsPath string `marker:"manifests"`

	// MaxDescLen specifies the maximum description length for fields in CRD's OpenAPI schema.
	//
	// 0 indicates drop the description for all fields completely.
	// n indicates limit the description to at most n characters and truncate the description to
	// closest sentence boundary if it exceeds n characters.
	MaxDescLen *int `marker:",optional"`
}

var _ genall.Generator = &Generator{}

func (Generator) RegisterMarkers(into *markers.Registry) error {
	return crdmarkers.Register(into)
}

func (g Generator) Generate(ctx *genall.GenerationContext) (result error) {
	parser := &crdgen.Parser{
		Collector: ctx.Collector,
		Checker:   ctx.Checker,
	}

	crdgen.AddKnownTypes(parser)
	for _, root := range ctx.Roots {
		parser.NeedPackage(root)
	}

	metav1Pkg := crdgen.FindMetav1(ctx.Roots)
	if metav1Pkg == nil {
		// no objects in the roots, since nothing imported metav1
		return nil
	}

	// load existing CRD manifests with group-kind and versions
	partialCRDs, err := crdsFromDirectory(ctx, g.ManifestsPath)
	if err != nil {
		return err
	}

	// generate schemata for the types we care about, and save them to be written later.
	for _, groupKind := range crdgen.FindKubeKinds(parser, metav1Pkg) {
		existingInfo, wanted := partialCRDs[groupKind]
		if !wanted {
			continue
		}

		for pkg, gv := range parser.GroupVersions {
			if gv.Group != groupKind.Group {
				continue
			}
			if _, wantedVersion := existingInfo.Versions[gv.Version]; !wantedVersion {
				continue
			}

			typeIdent := crdgen.TypeIdent{Package: pkg, Name: groupKind.Kind}
			parser.NeedFlattenedSchemaFor(typeIdent)

			fullSchema := parser.FlattenedSchemata[typeIdent]
			if g.MaxDescLen != nil {
				fullSchema = *fullSchema.DeepCopy()
				crdgen.TruncateDescription(&fullSchema, *g.MaxDescLen)
			}
			existingInfo.NewSchemata[gv.Version] = fullSchema
		}
	}

	// patch existing CRDs with new schemata
	for _, existingInfo := range partialCRDs {
		// first, figure out if we need to merge schemata together if they're *all*
		// identical (meaning we also don't have any "unset" versions)

		if len(existingInfo.NewSchemata) == 0 {
			continue
		}

		// copy over the new versions that we have, keeping old versions so
		// that we can tell if a schema would be nill
		var someVer string
		for ver := range existingInfo.NewSchemata {
			someVer = ver
			existingInfo.Versions[ver] = struct{}{}
		}

		allSame := true
		firstSchema := existingInfo.NewSchemata[someVer]
		for ver := range existingInfo.Versions {
			otherSchema, hasSchema := existingInfo.NewSchemata[ver]
			if !hasSchema || !equality.Semantic.DeepEqual(firstSchema, otherSchema) {
				allSame = false
				break
			}
		}

		if allSame {
			if err := existingInfo.setGlobalSchema(); err != nil {
				return fmt.Errorf("failed to set global firstSchema for %s: %v", existingInfo.GroupKind, err)
			}
		} else {
			if err := existingInfo.setVersionedSchemata(); err != nil {
				return fmt.Errorf("failed to set versioned schemas for %s: %v", existingInfo.GroupKind, err)
			}
		}
	}

	// write the final result out to the new location
	for _, crd := range partialCRDs {
		if err := func() error {
			outWriter, err := ctx.OutputRule.Open(nil, crd.FileName)
			if err != nil {
				return err
			}
			defer outWriter.Close()

			enc := yaml.NewEncoder(outWriter)
			// yaml.v2 defaults to indent=2, yaml.v3 defaults to indent=4,
			// so be compatible with everything else in k8s and choose 2.
			enc.SetIndent(2)
			return enc.Encode(crd.Yaml)
		}(); err != nil {
			return err
		}
	}

	return nil
}

// partialCRD tracks modifications to the schemata of a CRD.  It contains the
// raw YAML representation of a CRD, plus some structured content (versions,
// filename, etc) for easy lookup, and any new schemata registered.
type partialCRD struct {
	GroupKind schema.GroupKind
	Yaml      *yaml.Node
	Versions  map[string]struct{}
	FileName  string

	NewSchemata map[string]apiext.JSONSchemaProps
}

// setGlobalSchema sets the global schema to one of the schemata
// for this CRD.  All schemata must be identical for this to be a valid operation.
func (e *partialCRD) setGlobalSchema() error {
	// there's no easy way to get a "random" key from a go map :-/
	var schema apiext.JSONSchemaProps
	for ver := range e.NewSchemata {
		schema = e.NewSchemata[ver]
		break
	}

	schemaNodeTree, err := yamlop.ToYAML(schema)
	if err != nil {
		return err
	}
	schemaNodeTree = schemaNodeTree.Content[0] // get rid of the document node
	yamlop.SetStyle(schemaNodeTree, 0)         // clear the style so it defaults to auto-style-choice

	if err := yamlop.SetNode(e.Yaml, *schemaNodeTree, "spec", "validation", "openAPIV3Schema"); err != nil {
		return err
	}

	versions, found, err := e.getVersionsNode()
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	for i, verNode := range versions.Content {
		if err := yamlop.DeleteNode(verNode, "schema"); err != nil {
			return fmt.Errorf("spec.versions[%d]: %v", i, err)
		}
	}

	return nil
}

// getVersionsNode gets the YAML node of .spec.versions YAML mapping,
// if returning the node, and whether or not it was present.
func (e *partialCRD) getVersionsNode() (*yaml.Node, bool, error) {
	versions, found, err := yamlop.GetNode(e.Yaml, "spec", "versions")
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}
	if versions.Kind != yaml.SequenceNode {
		return nil, true, fmt.Errorf("unexpected non-sequence versions")
	}
	return versions, found, nil
}

// setVersionedSchemata populates all existing versions with new schemata,
// wiping the schema of any version that doesn't have a listed schema.
// Any "unknown" versions are ignored.
func (e *partialCRD) setVersionedSchemata() error {
	var err error
	if err := yamlop.DeleteNode(e.Yaml, "spec", "validation"); err != nil {
		return err
	}

	versions, found, err := e.getVersionsNode()
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("unexpected missing versions")
	}

	for i, verNode := range versions.Content {
		nameNode, _, _ := yamlop.GetNode(verNode, "name")
		if nameNode.Kind != yaml.ScalarNode || nameNode.ShortTag() != "!!str" {
			return fmt.Errorf("version name was not a string at spec.versions[%d]", i)
		}
		name := nameNode.Value
		if name == "" {
			return fmt.Errorf("unexpected empty name at spec.versions[%d]", i)
		}
		newSchema, found := e.NewSchemata[name]
		if !found {
			if err := yamlop.DeleteNode(verNode, "schema"); err != nil {
				return fmt.Errorf("spec.versions[%d]: %v", i, err)
			}
		} else {
			schemaNodeTree, err := yamlop.ToYAML(newSchema)
			if err != nil {
				return fmt.Errorf("failed to convert schema to YAML: %v", err)
			}
			schemaNodeTree = schemaNodeTree.Content[0] // get rid of the document node
			yamlop.SetStyle(schemaNodeTree, 0)         // clear the style so it defaults to an auto-chosen one
			if err := yamlop.SetNode(verNode, *schemaNodeTree, "schema", "openAPIV3Schema"); err != nil {
				return fmt.Errorf("spec.versions[%d]: %v", i, err)
			}
		}
	}
	return nil
}

// crdsFromDirectory returns loads all CRDs from the given directory in a
// manner that preserves ordering, comments, etc in order to make patching
// minimally invasive.  Returned CRDs are mapped by group-kind.
func crdsFromDirectory(ctx *genall.GenerationContext, dir string) (map[schema.GroupKind]*partialCRD, error) {
	apiextAPIVersion := apiext.SchemeGroupVersion.String()

	res := map[schema.GroupKind]*partialCRD{}
	dirEntries, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, fileInfo := range dirEntries {
		// find all files that are YAML
		if fileInfo.IsDir() || filepath.Ext(fileInfo.Name()) != ".yaml" {
			continue
		}

		rawContent, err := ctx.ReadFile(filepath.Join(dir, fileInfo.Name()))
		if err != nil {
			return nil, err
		}

		// NB(directxman12): we could use the universal deserializer for this, but it's
		// really pretty clunky, and the alternative is actually kinda easier to understand

		// ensure that this is a CRD
		var typeMeta metav1.TypeMeta
		if err := kyaml.Unmarshal(rawContent, &typeMeta); err != nil {
			continue
		}
		if typeMeta.APIVersion != apiextAPIVersion || typeMeta.Kind != "CustomResourceDefinition" {
			continue
		}

		// collect the group-kind and versions from the actual structured form
		var actualCRD apiext.CustomResourceDefinition
		if err := kyaml.Unmarshal(rawContent, &actualCRD); err != nil {
			continue
		}
		groupKind := schema.GroupKind{Group: actualCRD.Spec.Group, Kind: actualCRD.Spec.Names.Kind}
		var versions map[string]struct{}
		if len(actualCRD.Spec.Versions) == 0 {
			versions = map[string]struct{}{actualCRD.Spec.Version: struct{}{}}
		} else {
			versions = make(map[string]struct{}, len(actualCRD.Spec.Versions))
			for _, ver := range actualCRD.Spec.Versions {
				versions[ver.Name] = struct{}{}
			}
		}

		// then actually unmarshal in a manner that preserves ordering, etc
		var yamlNodeTree yaml.Node
		if err := yaml.Unmarshal(rawContent, &yamlNodeTree); err != nil {
			continue
		}

		res[groupKind] = &partialCRD{
			GroupKind:   groupKind,
			Yaml:        &yamlNodeTree,
			Versions:    versions,
			FileName:    fileInfo.Name(),
			NewSchemata: make(map[string]apiext.JSONSchemaProps),
		}
	}
	return res, nil
}
