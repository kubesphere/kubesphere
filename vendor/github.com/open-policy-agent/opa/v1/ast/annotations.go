// Copyright 2022 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/open-policy-agent/opa/internal/deepcopy"
	astJSON "github.com/open-policy-agent/opa/v1/ast/json"
	"github.com/open-policy-agent/opa/v1/util"
)

const (
	annotationScopePackage     = "package"
	annotationScopeRule        = "rule"
	annotationScopeDocument    = "document"
	annotationScopeSubpackages = "subpackages"
)

var (
	scopeTerm            = StringTerm("scope")
	titleTerm            = StringTerm("title")
	entrypointTerm       = StringTerm("entrypoint")
	descriptionTerm      = StringTerm("description")
	organizationsTerm    = StringTerm("organizations")
	authorsTerm          = StringTerm("authors")
	relatedResourcesTerm = StringTerm("related_resources")
	schemasTerm          = StringTerm("schemas")
	customTerm           = StringTerm("custom")
	refTerm              = StringTerm("ref")
	nameTerm             = StringTerm("name")
	emailTerm            = StringTerm("email")
	schemaTerm           = StringTerm("schema")
	definitionTerm       = StringTerm("definition")
	documentTerm         = StringTerm(annotationScopeDocument)
	packageTerm          = StringTerm(annotationScopePackage)
	ruleTerm             = StringTerm(annotationScopeRule)
	subpackagesTerm      = StringTerm(annotationScopeSubpackages)
)

type (
	// Annotations represents metadata attached to other AST nodes such as rules.
	Annotations struct {
		Scope            string                       `json:"scope"`
		Title            string                       `json:"title,omitempty"`
		Entrypoint       bool                         `json:"entrypoint,omitempty"`
		Description      string                       `json:"description,omitempty"`
		Organizations    []string                     `json:"organizations,omitempty"`
		RelatedResources []*RelatedResourceAnnotation `json:"related_resources,omitempty"`
		Authors          []*AuthorAnnotation          `json:"authors,omitempty"`
		Schemas          []*SchemaAnnotation          `json:"schemas,omitempty"`
		Custom           map[string]interface{}       `json:"custom,omitempty"`
		Location         *Location                    `json:"location,omitempty"`

		comments []*Comment
		node     Node
	}

	// SchemaAnnotation contains a schema declaration for the document identified by the path.
	SchemaAnnotation struct {
		Path       Ref          `json:"path"`
		Schema     Ref          `json:"schema,omitempty"`
		Definition *interface{} `json:"definition,omitempty"`
	}

	AuthorAnnotation struct {
		Name  string `json:"name"`
		Email string `json:"email,omitempty"`
	}

	RelatedResourceAnnotation struct {
		Ref         url.URL `json:"ref"`
		Description string  `json:"description,omitempty"`
	}

	AnnotationSet struct {
		byRule    map[*Rule][]*Annotations
		byPackage map[int]*Annotations
		byPath    *annotationTreeNode
		modules   []*Module // Modules this set was constructed from
	}

	annotationTreeNode struct {
		Value    *Annotations
		Children map[Value]*annotationTreeNode // we assume key elements are hashable (vars and strings only!)
	}

	AnnotationsRef struct {
		Path        Ref          `json:"path"` // The path of the node the annotations are applied to
		Annotations *Annotations `json:"annotations,omitempty"`
		Location    *Location    `json:"location,omitempty"` // The location of the node the annotations are applied to

		node Node // The node the annotations are applied to
	}

	AnnotationsRefSet []*AnnotationsRef

	FlatAnnotationsRefSet AnnotationsRefSet
)

func (a *Annotations) String() string {
	bs, _ := a.MarshalJSON()
	return string(bs)
}

// Loc returns the location of this annotation.
func (a *Annotations) Loc() *Location {
	return a.Location
}

// SetLoc updates the location of this annotation.
func (a *Annotations) SetLoc(l *Location) {
	a.Location = l
}

// EndLoc returns the location of this annotation's last comment line.
func (a *Annotations) EndLoc() *Location {
	count := len(a.comments)
	if count == 0 {
		return a.Location
	}
	return a.comments[count-1].Location
}

// Compare returns an integer indicating if a is less than, equal to, or greater
// than other.
func (a *Annotations) Compare(other *Annotations) int {

	if a == nil && other == nil {
		return 0
	}

	if a == nil {
		return -1
	}

	if other == nil {
		return 1
	}

	if cmp := scopeCompare(a.Scope, other.Scope); cmp != 0 {
		return cmp
	}

	if cmp := strings.Compare(a.Title, other.Title); cmp != 0 {
		return cmp
	}

	if cmp := strings.Compare(a.Description, other.Description); cmp != 0 {
		return cmp
	}

	if cmp := compareStringLists(a.Organizations, other.Organizations); cmp != 0 {
		return cmp
	}

	if cmp := compareRelatedResources(a.RelatedResources, other.RelatedResources); cmp != 0 {
		return cmp
	}

	if cmp := compareAuthors(a.Authors, other.Authors); cmp != 0 {
		return cmp
	}

	if cmp := compareSchemas(a.Schemas, other.Schemas); cmp != 0 {
		return cmp
	}

	if a.Entrypoint != other.Entrypoint {
		if a.Entrypoint {
			return 1
		}
		return -1
	}

	if cmp := util.Compare(a.Custom, other.Custom); cmp != 0 {
		return cmp
	}

	return 0
}

// GetTargetPath returns the path of the node these Annotations are applied to (the target)
func (a *Annotations) GetTargetPath() Ref {
	switch n := a.node.(type) {
	case *Package:
		return n.Path
	case *Rule:
		return n.Ref().GroundPrefix()
	default:
		return nil
	}
}

func (a *Annotations) MarshalJSON() ([]byte, error) {
	if a == nil {
		return []byte(`{"scope":""}`), nil
	}

	data := map[string]interface{}{
		"scope": a.Scope,
	}

	if a.Title != "" {
		data["title"] = a.Title
	}

	if a.Description != "" {
		data["description"] = a.Description
	}

	if a.Entrypoint {
		data["entrypoint"] = a.Entrypoint
	}

	if len(a.Organizations) > 0 {
		data["organizations"] = a.Organizations
	}

	if len(a.RelatedResources) > 0 {
		data["related_resources"] = a.RelatedResources
	}

	if len(a.Authors) > 0 {
		data["authors"] = a.Authors
	}

	if len(a.Schemas) > 0 {
		data["schemas"] = a.Schemas
	}

	if len(a.Custom) > 0 {
		data["custom"] = a.Custom
	}

	if astJSON.GetOptions().MarshalOptions.IncludeLocation.Annotations {
		if a.Location != nil {
			data["location"] = a.Location
		}
	}

	return json.Marshal(data)
}

func NewAnnotationsRef(a *Annotations) *AnnotationsRef {
	var loc *Location
	if a.node != nil {
		loc = a.node.Loc()
	}

	return &AnnotationsRef{
		Location:    loc,
		Path:        a.GetTargetPath(),
		Annotations: a,
		node:        a.node,
	}
}

func (ar *AnnotationsRef) GetPackage() *Package {
	switch n := ar.node.(type) {
	case *Package:
		return n
	case *Rule:
		return n.Module.Package
	default:
		return nil
	}
}

func (ar *AnnotationsRef) GetRule() *Rule {
	switch n := ar.node.(type) {
	case *Rule:
		return n
	default:
		return nil
	}
}

func (ar *AnnotationsRef) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{
		"path": ar.Path,
	}

	if ar.Annotations != nil {
		data["annotations"] = ar.Annotations
	}

	if astJSON.GetOptions().MarshalOptions.IncludeLocation.AnnotationsRef {
		if ar.Location != nil {
			data["location"] = ar.Location
		}

		// The location set for the schema ref terms is wrong (always set to
		// row 1) and not really useful anyway.. so strip it out before marshalling
		for _, schema := range ar.Annotations.Schemas {
			if schema.Path != nil {
				for _, term := range schema.Path {
					term.Location = nil
				}
			}
		}
	}

	return json.Marshal(data)
}

func scopeCompare(s1, s2 string) int {
	o1 := scopeOrder(s1)
	o2 := scopeOrder(s2)

	if o2 < o1 {
		return 1
	} else if o2 > o1 {
		return -1
	}

	if s1 < s2 {
		return -1
	} else if s2 < s1 {
		return 1
	}

	return 0
}

func scopeOrder(s string) int {
	if s == annotationScopeRule {
		return 1
	}
	return 0
}

func compareAuthors(a, b []*AuthorAnnotation) int {
	if len(a) > len(b) {
		return 1
	} else if len(a) < len(b) {
		return -1
	}

	for i := range a {
		if cmp := a[i].Compare(b[i]); cmp != 0 {
			return cmp
		}
	}

	return 0
}

func compareRelatedResources(a, b []*RelatedResourceAnnotation) int {
	if len(a) > len(b) {
		return 1
	} else if len(a) < len(b) {
		return -1
	}

	for i := range a {
		if cmp := a[i].Compare(b[i]); cmp != 0 {
			return cmp
		}
	}

	return 0
}

func compareSchemas(a, b []*SchemaAnnotation) int {
	maxLen := len(a)
	if len(b) < maxLen {
		maxLen = len(b)
	}

	for i := range maxLen {
		if cmp := a[i].Compare(b[i]); cmp != 0 {
			return cmp
		}
	}

	if len(a) > len(b) {
		return 1
	} else if len(a) < len(b) {
		return -1
	}

	return 0
}

func compareStringLists(a, b []string) int {
	if len(a) > len(b) {
		return 1
	} else if len(a) < len(b) {
		return -1
	}

	for i := range a {
		if cmp := strings.Compare(a[i], b[i]); cmp != 0 {
			return cmp
		}
	}

	return 0
}

// Copy returns a deep copy of s.
func (a *Annotations) Copy(node Node) *Annotations {
	cpy := *a

	cpy.Organizations = make([]string, len(a.Organizations))
	copy(cpy.Organizations, a.Organizations)

	cpy.RelatedResources = make([]*RelatedResourceAnnotation, len(a.RelatedResources))
	for i := range a.RelatedResources {
		cpy.RelatedResources[i] = a.RelatedResources[i].Copy()
	}

	cpy.Authors = make([]*AuthorAnnotation, len(a.Authors))
	for i := range a.Authors {
		cpy.Authors[i] = a.Authors[i].Copy()
	}

	cpy.Schemas = make([]*SchemaAnnotation, len(a.Schemas))
	for i := range a.Schemas {
		cpy.Schemas[i] = a.Schemas[i].Copy()
	}

	if a.Custom != nil {
		cpy.Custom = deepcopy.Map(a.Custom)
	}

	cpy.node = node

	return &cpy
}

// toObject constructs an AST Object from the annotation.
func (a *Annotations) toObject() (*Object, *Error) {
	obj := NewObject()

	if a == nil {
		return &obj, nil
	}

	if len(a.Scope) > 0 {
		switch a.Scope {
		case annotationScopeDocument:
			obj.Insert(scopeTerm, documentTerm)
		case annotationScopePackage:
			obj.Insert(scopeTerm, packageTerm)
		case annotationScopeRule:
			obj.Insert(scopeTerm, ruleTerm)
		case annotationScopeSubpackages:
			obj.Insert(scopeTerm, subpackagesTerm)
		default:
			obj.Insert(scopeTerm, StringTerm(a.Scope))
		}
	}

	if len(a.Title) > 0 {
		obj.Insert(titleTerm, StringTerm(a.Title))
	}

	if a.Entrypoint {
		obj.Insert(entrypointTerm, InternedBooleanTerm(true))
	}

	if len(a.Description) > 0 {
		obj.Insert(descriptionTerm, StringTerm(a.Description))
	}

	if len(a.Organizations) > 0 {
		orgs := make([]*Term, 0, len(a.Organizations))
		for _, org := range a.Organizations {
			orgs = append(orgs, StringTerm(org))
		}
		obj.Insert(organizationsTerm, ArrayTerm(orgs...))
	}

	if len(a.RelatedResources) > 0 {
		rrs := make([]*Term, 0, len(a.RelatedResources))
		for _, rr := range a.RelatedResources {
			rrObj := NewObject(Item(refTerm, StringTerm(rr.Ref.String())))
			if len(rr.Description) > 0 {
				rrObj.Insert(descriptionTerm, StringTerm(rr.Description))
			}
			rrs = append(rrs, NewTerm(rrObj))
		}
		obj.Insert(relatedResourcesTerm, ArrayTerm(rrs...))
	}

	if len(a.Authors) > 0 {
		as := make([]*Term, 0, len(a.Authors))
		for _, author := range a.Authors {
			aObj := NewObject()
			if len(author.Name) > 0 {
				aObj.Insert(nameTerm, StringTerm(author.Name))
			}
			if len(author.Email) > 0 {
				aObj.Insert(emailTerm, StringTerm(author.Email))
			}
			as = append(as, NewTerm(aObj))
		}
		obj.Insert(authorsTerm, ArrayTerm(as...))
	}

	if len(a.Schemas) > 0 {
		ss := make([]*Term, 0, len(a.Schemas))
		for _, s := range a.Schemas {
			sObj := NewObject()
			if len(s.Path) > 0 {
				sObj.Insert(pathTerm, NewTerm(s.Path.toArray()))
			}
			if len(s.Schema) > 0 {
				sObj.Insert(schemaTerm, NewTerm(s.Schema.toArray()))
			}
			if s.Definition != nil {
				def, err := InterfaceToValue(s.Definition)
				if err != nil {
					return nil, NewError(CompileErr, a.Location, "invalid definition in schema annotation: %s", err.Error())
				}
				sObj.Insert(definitionTerm, NewTerm(def))
			}
			ss = append(ss, NewTerm(sObj))
		}
		obj.Insert(schemasTerm, ArrayTerm(ss...))
	}

	if len(a.Custom) > 0 {
		c, err := InterfaceToValue(a.Custom)
		if err != nil {
			return nil, NewError(CompileErr, a.Location, "invalid custom annotation %s", err.Error())
		}
		obj.Insert(customTerm, NewTerm(c))
	}

	return &obj, nil
}

func attachRuleAnnotations(mod *Module) {
	// make a copy of the annotations
	cpy := make([]*Annotations, len(mod.Annotations))
	for i, a := range mod.Annotations {
		cpy[i] = a.Copy(a.node)
	}

	for _, rule := range mod.Rules {
		var j int
		var found bool
		for i, a := range cpy {
			if rule.Ref().GroundPrefix().Equal(a.GetTargetPath()) {
				if a.Scope == annotationScopeDocument {
					rule.Annotations = append(rule.Annotations, a)
				} else if a.Scope == annotationScopeRule && rule.Loc().Row > a.Location.Row {
					j = i
					found = true
					rule.Annotations = append(rule.Annotations, a)
				}
			}
		}

		if found && j < len(cpy) {
			cpy = append(cpy[:j], cpy[j+1:]...)
		}
	}
}

func attachAnnotationsNodes(mod *Module) Errors {
	var errs Errors

	// Find first non-annotation statement following each annotation and attach
	// the annotation to that statement.
	for _, a := range mod.Annotations {
		for _, stmt := range mod.stmts {
			_, ok := stmt.(*Annotations)
			if !ok {
				if stmt.Loc().Row > a.Location.Row {
					a.node = stmt
					break
				}
			}
		}

		if a.Scope == "" {
			switch a.node.(type) {
			case *Rule:
				if a.Entrypoint {
					a.Scope = annotationScopeDocument
				} else {
					a.Scope = annotationScopeRule
				}
			case *Package:
				a.Scope = annotationScopePackage
			case *Import:
				// Note that this isn't a valid scope, but set here so that the
				// validate function called below can print an error message with
				// a context that makes sense ("invalid scope: 'import'" instead of
				// "invalid scope: '')
				a.Scope = "import"
			}
		}

		if err := validateAnnotationScopeAttachment(a); err != nil {
			errs = append(errs, err)
		}

		if err := validateAnnotationEntrypointAttachment(a); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func validateAnnotationScopeAttachment(a *Annotations) *Error {

	switch a.Scope {
	case annotationScopeRule, annotationScopeDocument:
		if _, ok := a.node.(*Rule); ok {
			return nil
		}
		return newScopeAttachmentErr(a, "rule")
	case annotationScopePackage, annotationScopeSubpackages:
		if _, ok := a.node.(*Package); ok {
			return nil
		}
		return newScopeAttachmentErr(a, "package")
	}

	return NewError(ParseErr, a.Loc(), "invalid annotation scope '%v'. Use one of '%s', '%s', '%s', or '%s'",
		a.Scope, annotationScopeRule, annotationScopeDocument, annotationScopePackage, annotationScopeSubpackages)
}

func validateAnnotationEntrypointAttachment(a *Annotations) *Error {
	if a.Entrypoint && !(a.Scope == annotationScopeDocument || a.Scope == annotationScopePackage) {
		return NewError(
			ParseErr, a.Loc(), "annotation entrypoint applied to non-document or package scope '%v'", a.Scope)
	}
	return nil
}

// Copy returns a deep copy of a.
func (a *AuthorAnnotation) Copy() *AuthorAnnotation {
	cpy := *a
	return &cpy
}

// Compare returns an integer indicating if s is less than, equal to, or greater
// than other.
func (a *AuthorAnnotation) Compare(other *AuthorAnnotation) int {
	if cmp := strings.Compare(a.Name, other.Name); cmp != 0 {
		return cmp
	}

	if cmp := strings.Compare(a.Email, other.Email); cmp != 0 {
		return cmp
	}

	return 0
}

func (a *AuthorAnnotation) String() string {
	if len(a.Email) == 0 {
		return a.Name
	} else if len(a.Name) == 0 {
		return fmt.Sprintf("<%s>", a.Email)
	}
	return fmt.Sprintf("%s <%s>", a.Name, a.Email)
}

// Copy returns a deep copy of rr.
func (rr *RelatedResourceAnnotation) Copy() *RelatedResourceAnnotation {
	cpy := *rr
	return &cpy
}

// Compare returns an integer indicating if s is less than, equal to, or greater
// than other.
func (rr *RelatedResourceAnnotation) Compare(other *RelatedResourceAnnotation) int {
	if cmp := strings.Compare(rr.Description, other.Description); cmp != 0 {
		return cmp
	}

	if cmp := strings.Compare(rr.Ref.String(), other.Ref.String()); cmp != 0 {
		return cmp
	}

	return 0
}

func (rr *RelatedResourceAnnotation) String() string {
	bs, _ := json.Marshal(rr)
	return string(bs)
}

func (rr *RelatedResourceAnnotation) MarshalJSON() ([]byte, error) {
	d := map[string]interface{}{
		"ref": rr.Ref.String(),
	}

	if len(rr.Description) > 0 {
		d["description"] = rr.Description
	}

	return json.Marshal(d)
}

// Copy returns a deep copy of s.
func (s *SchemaAnnotation) Copy() *SchemaAnnotation {
	cpy := *s
	return &cpy
}

// Compare returns an integer indicating if s is less than, equal to, or greater
// than other.
func (s *SchemaAnnotation) Compare(other *SchemaAnnotation) int {
	if cmp := s.Path.Compare(other.Path); cmp != 0 {
		return cmp
	}

	if cmp := s.Schema.Compare(other.Schema); cmp != 0 {
		return cmp
	}

	if s.Definition != nil && other.Definition == nil {
		return -1
	} else if s.Definition == nil && other.Definition != nil {
		return 1
	} else if s.Definition != nil && other.Definition != nil {
		return util.Compare(*s.Definition, *other.Definition)
	}

	return 0
}

func (s *SchemaAnnotation) String() string {
	bs, _ := json.Marshal(s)
	return string(bs)
}

func newAnnotationSet() *AnnotationSet {
	return &AnnotationSet{
		byRule:    map[*Rule][]*Annotations{},
		byPackage: map[int]*Annotations{},
		byPath:    newAnnotationTree(),
	}
}

func BuildAnnotationSet(modules []*Module) (*AnnotationSet, Errors) {
	as := newAnnotationSet()
	var errs Errors
	for _, m := range modules {
		for _, a := range m.Annotations {
			if err := as.add(a); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		return nil, errs
	}
	as.modules = modules
	return as, nil
}

// NOTE(philipc): During copy propagation, the underlying Nodes can be
// stripped away from the annotations, leading to nil deref panics. We
// silently ignore these cases for now, as a workaround.
func (as *AnnotationSet) add(a *Annotations) *Error {
	switch a.Scope {
	case annotationScopeRule:
		if rule, ok := a.node.(*Rule); ok {
			as.byRule[rule] = append(as.byRule[rule], a)
		}
	case annotationScopePackage:
		if pkg, ok := a.node.(*Package); ok {
			hash := pkg.Path.Hash()
			if exist, ok := as.byPackage[hash]; ok {
				return errAnnotationRedeclared(a, exist.Location)
			}
			as.byPackage[hash] = a
		}
	case annotationScopeDocument:
		if rule, ok := a.node.(*Rule); ok {
			path := rule.Ref().GroundPrefix()
			x := as.byPath.get(path)
			if x != nil {
				return errAnnotationRedeclared(a, x.Value.Location)
			}
			as.byPath.insert(path, a)
		}
	case annotationScopeSubpackages:
		if pkg, ok := a.node.(*Package); ok {
			x := as.byPath.get(pkg.Path)
			if x != nil && x.Value != nil {
				return errAnnotationRedeclared(a, x.Value.Location)
			}
			as.byPath.insert(pkg.Path, a)
		}
	}
	return nil
}

func (as *AnnotationSet) GetRuleScope(r *Rule) []*Annotations {
	if as == nil {
		return nil
	}
	return as.byRule[r]
}

func (as *AnnotationSet) GetSubpackagesScope(path Ref) []*Annotations {
	if as == nil {
		return nil
	}
	return as.byPath.ancestors(path)
}

func (as *AnnotationSet) GetDocumentScope(path Ref) *Annotations {
	if as == nil {
		return nil
	}
	if node := as.byPath.get(path); node != nil {
		return node.Value
	}
	return nil
}

func (as *AnnotationSet) GetPackageScope(pkg *Package) *Annotations {
	if as == nil {
		return nil
	}
	return as.byPackage[pkg.Path.Hash()]
}

// Flatten returns a flattened list view of this AnnotationSet.
// The returned slice is sorted, first by the annotations' target path, then by their target location
func (as *AnnotationSet) Flatten() FlatAnnotationsRefSet {
	// This preallocation often won't be optimal, but it's superior to starting with a nil slice.
	refs := make([]*AnnotationsRef, 0, len(as.byPath.Children)+len(as.byRule)+len(as.byPackage))

	refs = as.byPath.flatten(refs)

	for _, a := range as.byPackage {
		refs = append(refs, NewAnnotationsRef(a))
	}

	for _, as := range as.byRule {
		for _, a := range as {
			refs = append(refs, NewAnnotationsRef(a))
		}
	}

	// Sort by path, then annotation location, for stable output
	slices.SortStableFunc(refs, (*AnnotationsRef).Compare)

	return refs
}

// Chain returns the chain of annotations leading up to the given rule.
// The returned slice is ordered as follows
// 0. Entries for the given rule, ordered from the METADATA block declared immediately above the rule, to the block declared farthest away (always at least one entry)
// 1. The 'document' scope entry, if any
// 2. The 'package' scope entry, if any
// 3. Entries for the 'subpackages' scope, if any; ordered from the closest package path to the fartest. E.g.: 'do.re.mi', 'do.re', 'do'
// The returned slice is guaranteed to always contain at least one entry, corresponding to the given rule.
func (as *AnnotationSet) Chain(rule *Rule) AnnotationsRefSet {
	var refs []*AnnotationsRef

	ruleAnnots := as.GetRuleScope(rule)

	if len(ruleAnnots) >= 1 {
		for _, a := range ruleAnnots {
			refs = append(refs, NewAnnotationsRef(a))
		}
	} else {
		// Make sure there is always a leading entry representing the passed rule, even if it has no annotations
		refs = append(refs, &AnnotationsRef{
			Location: rule.Location,
			Path:     rule.Ref().GroundPrefix(),
			node:     rule,
		})
	}

	if len(refs) > 1 {
		// Sort by annotation location; chain must start with annotations declared closest to rule, then going outward
		slices.SortStableFunc(refs, func(a, b *AnnotationsRef) int {
			return -a.Annotations.Location.Compare(b.Annotations.Location)
		})
	}

	docAnnots := as.GetDocumentScope(rule.Ref().GroundPrefix())
	if docAnnots != nil {
		refs = append(refs, NewAnnotationsRef(docAnnots))
	}

	pkg := rule.Module.Package
	pkgAnnots := as.GetPackageScope(pkg)
	if pkgAnnots != nil {
		refs = append(refs, NewAnnotationsRef(pkgAnnots))
	}

	subPkgAnnots := as.GetSubpackagesScope(pkg.Path)
	// We need to reverse the order, as subPkgAnnots ordering will start at the root,
	// whereas we want to end at the root.
	for i := len(subPkgAnnots) - 1; i >= 0; i-- {
		refs = append(refs, NewAnnotationsRef(subPkgAnnots[i]))
	}

	return refs
}

func (ars FlatAnnotationsRefSet) Insert(ar *AnnotationsRef) FlatAnnotationsRefSet {
	result := make(FlatAnnotationsRefSet, 0, len(ars)+1)

	// insertion sort, first by path, then location
	for i, current := range ars {
		if ar.Compare(current) < 0 {
			result = append(result, ar)
			result = append(result, ars[i:]...)
			break
		}
		result = append(result, current)
	}

	if len(result) < len(ars)+1 {
		result = append(result, ar)
	}

	return result
}

func newAnnotationTree() *annotationTreeNode {
	return &annotationTreeNode{
		Value:    nil,
		Children: map[Value]*annotationTreeNode{},
	}
}

func (t *annotationTreeNode) insert(path Ref, value *Annotations) {
	node := t
	for _, k := range path {
		child, ok := node.Children[k.Value]
		if !ok {
			child = newAnnotationTree()
			node.Children[k.Value] = child
		}
		node = child
	}
	node.Value = value
}

func (t *annotationTreeNode) get(path Ref) *annotationTreeNode {
	node := t
	for _, k := range path {
		if node == nil {
			return nil
		}
		child, ok := node.Children[k.Value]
		if !ok {
			return nil
		}
		node = child
	}
	return node
}

// ancestors returns a slice of annotations in ascending order, starting with the root of ref; e.g.: 'root', 'root.foo', 'root.foo.bar'.
func (t *annotationTreeNode) ancestors(path Ref) (result []*Annotations) {
	node := t
	for _, k := range path {
		if node == nil {
			return result
		}
		child, ok := node.Children[k.Value]
		if !ok {
			return result
		}
		if child.Value != nil {
			result = append(result, child.Value)
		}
		node = child
	}
	return result
}

func (t *annotationTreeNode) flatten(refs []*AnnotationsRef) []*AnnotationsRef {
	if a := t.Value; a != nil {
		refs = append(refs, NewAnnotationsRef(a))
	}
	for _, c := range t.Children {
		refs = c.flatten(refs)
	}
	return refs
}

func (ar *AnnotationsRef) Compare(other *AnnotationsRef) int {
	if c := ar.Path.Compare(other.Path); c != 0 {
		return c
	}

	if c := ar.Annotations.Location.Compare(other.Annotations.Location); c != 0 {
		return c
	}

	return ar.Annotations.Compare(other.Annotations)
}
