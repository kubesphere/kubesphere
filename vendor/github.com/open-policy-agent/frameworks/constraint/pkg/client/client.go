package client

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"regexp"
	"strings"
	"sync"

	"github.com/open-policy-agent/frameworks/constraint/pkg/client/drivers"
	"github.com/open-policy-agent/frameworks/constraint/pkg/client/regolib"
	constraintlib "github.com/open-policy-agent/frameworks/constraint/pkg/core/constraints"
	"github.com/open-policy-agent/frameworks/constraint/pkg/core/templates"
	"github.com/open-policy-agent/frameworks/constraint/pkg/regorewriter"
	"github.com/open-policy-agent/frameworks/constraint/pkg/types"
	"github.com/open-policy-agent/opa/format"
	"github.com/pkg/errors"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const constraintGroup = "constraints.gatekeeper.sh"

type Opt func(*Client) error

// Client options

var targetNameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9.]*$`)

func Targets(ts ...TargetHandler) Opt {
	return func(c *Client) error {
		var errs Errors
		handlers := make(map[string]TargetHandler, len(ts))
		for _, t := range ts {
			if t.GetName() == "" {
				errs = append(errs, errors.New("Invalid target: a target is returning an empty string for GetName()"))
			} else if !targetNameRegex.MatchString(t.GetName()) {
				errs = append(errs, fmt.Errorf("Target name \"%s\" is not of the form %s", t.GetName(), targetNameRegex.String()))
			} else {
				handlers[t.GetName()] = t
			}
		}
		c.targets = handlers
		if len(errs) > 0 {
			return errs
		}
		return nil
	}
}

// AllowedDataFields sets the fields under `data` that Rego in ConstraintTemplates
// can access. If unset, all fields can be accessed. Only fields recognized by
// the system can be enabled.
func AllowedDataFields(fields ...string) Opt {
	return func(c *Client) error {
		c.allowedDataFields = fields
		return nil
	}
}

type templateEntry struct {
	template *templates.ConstraintTemplate
	CRD      *apiextensions.CustomResourceDefinition
	Targets  []string
}

type Client struct {
	backend           *Backend
	targets           map[string]TargetHandler
	constraintsMux    sync.RWMutex
	templates         map[templateKey]*templateEntry
	constraints       map[schema.GroupKind]map[string]*unstructured.Unstructured
	allowedDataFields []string
}

// createDataPath compiles the data destination: data.external.<target>.<path>
func createDataPath(target, subpath string) string {
	subpaths := strings.Split(subpath, "/")
	p := []string{"external", target}
	p = append(p, subpaths...)

	return "/" + path.Join(p...)
}

// AddData inserts the provided data into OPA for every target that can handle the data.
// On error, the responses return value will still be populated so that
// partial results can be analyzed.
func (c *Client) AddData(ctx context.Context, data interface{}) (*types.Responses, error) {
	resp := types.NewResponses()
	errMap := make(ErrorMap)
	for target, h := range c.targets {
		handled, path, processedData, err := h.ProcessData(data)
		if err != nil {
			errMap[target] = err
			continue
		}
		if !handled {
			continue
		}
		if err := c.backend.driver.PutData(ctx, createDataPath(target, path), processedData); err != nil {
			errMap[target] = err
			continue
		}
		resp.Handled[target] = true
	}
	if len(errMap) == 0 {
		return resp, nil
	}
	return resp, errMap
}

// RemoveData removes data from OPA for every target that can handle the data.
// On error, the responses return value will still be populated so that
// partial results can be analyzed.
func (c *Client) RemoveData(ctx context.Context, data interface{}) (*types.Responses, error) {
	resp := types.NewResponses()
	errMap := make(ErrorMap)
	for target, h := range c.targets {
		handled, path, _, err := h.ProcessData(data)
		if err != nil {
			errMap[target] = err
			continue
		}
		if !handled {
			continue
		}
		if _, err := c.backend.driver.DeleteData(ctx, createDataPath(target, path)); err != nil {
			errMap[target] = err
			continue
		}
		resp.Handled[target] = true
	}
	if len(errMap) == 0 {
		return resp, nil
	}
	return resp, errMap
}

// createTemplatePath returns the package path for a given template: templates.<target>.<name>
func createTemplatePath(target, name string) string {
	return fmt.Sprintf(`templates["%s"]["%s"]`, target, name)
}

// templateLibPrefix returns the new lib prefix for the libs that are specified in the CT.
func templateLibPrefix(target, name string) string {
	return fmt.Sprintf("libs.%s.%s", target, name)
}

// validateTargets handles validating the targets section of the CT.
func (c *Client) validateTargets(templ *templates.ConstraintTemplate) (*templates.Target, TargetHandler, error) {
	if err := validateTargets(templ); err != nil {
		return nil, nil, err
	}

	if len(templ.Spec.Targets) != 1 {
		return nil, nil, errors.Errorf("expected exactly 1 item in targets, got %v", templ.Spec.Targets)
	}

	targetSpec := &templ.Spec.Targets[0]
	targetHandler, found := c.targets[targetSpec.Target]
	if !found {
		return nil, nil, fmt.Errorf("target %s not recognized", targetSpec.Target)
	}

	return targetSpec, targetHandler, nil
}

type templateKey string

type keyableArtifact interface {
	Key() templateKey
}

var _ keyableArtifact = &basicCTArtifacts{}

func templateKeyFromConstraint(cst *unstructured.Unstructured) templateKey {
	return templateKey(strings.ToLower(cst.GetKind()))
}

// rawCTArtifacts have no processing and are only useful for looking things up
// from the cache
type rawCTArtifacts struct {
	// template is the template itself
	template *templates.ConstraintTemplate
}

func (a *rawCTArtifacts) Key() templateKey {
	return templateKey(a.template.GetName())
}

// createRawTemplateArtifacts creates the "free" artifacts for a template, avoiding more
// complex tasks like rewriting Rego. Provides minimal validation.
func (c *Client) createRawTemplateArtifacts(templ *templates.ConstraintTemplate) (*rawCTArtifacts, error) {
	if templ.ObjectMeta.Name == "" {
		return nil, errors.New("Template has no name")
	}
	return &rawCTArtifacts{template: templ}, nil
}

// basicCTArtifacts are the artifacts created by processing a constraint template
// that require little compute effort
type basicCTArtifacts struct {
	rawCTArtifacts

	// namePrefix is the name prefix by which the modules will be identified during create / delete
	// calls to the drivers.Driver interface.
	namePrefix string

	// gk is the groupKind of the constraints the template creates
	gk schema.GroupKind

	// crd is the CustomResourceDefinition created from the CT.
	crd *apiextensions.CustomResourceDefinition

	// targetHandler is the target handler indicated by the CT.  This isn't generated, but is used by
	// consumers of createTemplateArtifacts
	targetHandler TargetHandler

	// targetSpec is the target-oriented portion of a CT's Spec field.
	targetSpec *templates.Target
}

func (a basicCTArtifacts) CRD() *apiextensions.CustomResourceDefinition {
	return a.crd
}

// ctArtifacts are all artifacts created by processing a constraint template
type ctArtifacts struct {
	basicCTArtifacts

	// modules is the rewritten set of modules that the constraint template declares in Rego and Libs
	modules []string
}

// createBasicTemplateArtifacts creates the low-cost artifacts for a template, avoiding more
// complex tasks like rewriting Rego.
func (c *Client) createBasicTemplateArtifacts(templ *templates.ConstraintTemplate) (*basicCTArtifacts, error) {
	rawArtifacts, err := c.createRawTemplateArtifacts(templ)
	if err != nil {
		return nil, err
	}
	if templ.ObjectMeta.Name != strings.ToLower(templ.Spec.CRD.Spec.Names.Kind) {
		return nil, fmt.Errorf("Template's name %s is not equal to the lowercase of CRD's Kind: %s", templ.ObjectMeta.Name, strings.ToLower(templ.Spec.CRD.Spec.Names.Kind))
	}

	targetSpec, targetHandler, err := c.validateTargets(templ)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to validate targets for template %s", templ.Name)
	}

	sch, err := c.backend.crd.createSchema(templ, targetHandler)
	if err != nil {
		return nil, err
	}
	crd, err := c.backend.crd.createCRD(templ, sch)
	if err != nil {
		return nil, err
	}
	if err = c.backend.crd.validateCRD(crd); err != nil {
		return nil, err
	}

	entryPointPath := createTemplatePath(targetHandler.GetName(), templ.Spec.CRD.Spec.Names.Kind)

	return &basicCTArtifacts{
		rawCTArtifacts: *rawArtifacts,
		gk:             schema.GroupKind{Group: crd.Spec.Group, Kind: crd.Spec.Names.Kind},
		crd:            crd,
		targetHandler:  targetHandler,
		targetSpec:     targetSpec,
		namePrefix:     entryPointPath,
	}, nil
}

// createTemplateArtifacts will validate the CT, create the CRD for the CT's constraints, then
// validate and rewrite the rego sources specified in the CT.
func (c *Client) createTemplateArtifacts(templ *templates.ConstraintTemplate) (*ctArtifacts, error) {
	artifacts, err := c.createBasicTemplateArtifacts(templ)
	if err != nil {
		return nil, err
	}

	var externs []string
	for _, field := range c.allowedDataFields {
		externs = append(externs, fmt.Sprintf("data.%s", field))
	}

	libPrefix := templateLibPrefix(artifacts.targetHandler.GetName(), artifacts.crd.Spec.Names.Kind)
	rr, err := regorewriter.New(
		regorewriter.NewPackagePrefixer(libPrefix),
		[]string{"data.lib"},
		externs)
	if err != nil {
		return nil, err
	}

	entryPoint, err := parseModule(artifacts.namePrefix, artifacts.targetSpec.Rego)
	if err != nil {
		return nil, err
	}
	if entryPoint == nil {
		return nil, errors.Errorf("Failed to parse module for unknown reason")
	}

	if err := rewriteModulePackage(artifacts.namePrefix, entryPoint); err != nil {
		return nil, err
	}

	req := map[string]struct{}{"violation": {}}

	if err := requireRulesModule(entryPoint, req); err != nil {
		return nil, fmt.Errorf("Invalid rego: %s", err)
	}

	rr.AddEntryPointModule(artifacts.namePrefix, entryPoint)
	for idx, libSrc := range artifacts.targetSpec.Libs {
		libPath := fmt.Sprintf(`%s["lib_%d"]`, libPrefix, idx)
		if err := rr.AddLib(libPath, libSrc); err != nil {
			return nil, err
		}
	}

	sources, err := rr.Rewrite()
	if err != nil {
		return nil, err
	}

	var mods []string
	if err := sources.ForEachModule(func(m *regorewriter.Module) error {
		content, err := m.Content()
		if err != nil {
			return err
		}
		mods = append(mods, string(content))
		return nil
	}); err != nil {
		return nil, err
	}

	return &ctArtifacts{
		basicCTArtifacts: *artifacts,
		modules:          mods,
	}, nil
}

// CreateCRD creates a CRD from template
func (c *Client) CreateCRD(ctx context.Context, templ *templates.ConstraintTemplate) (*apiextensions.CustomResourceDefinition, error) {
	artifacts, err := c.createTemplateArtifacts(templ)
	if err != nil {
		return nil, err
	}
	return artifacts.crd, nil
}

// AddTemplate adds the template source code to OPA and registers the CRD with the client for
// schema validation on calls to AddConstraint. On error, the responses return value
// will still be populated so that partial results can be analyzed.
func (c *Client) AddTemplate(ctx context.Context, templ *templates.ConstraintTemplate) (*types.Responses, error) {
	resp := types.NewResponses()

	basicArtifacts, err := c.createBasicTemplateArtifacts(templ)
	if err != nil {
		return resp, err
	}

	// return immediately if no change
	if cached, err := c.GetTemplate(ctx, templ); err == nil && cached.SemanticEqual(templ) {
		resp.Handled[basicArtifacts.targetHandler.GetName()] = true
		return resp, nil
	}

	artifacts, err := c.createTemplateArtifacts(templ)
	if err != nil {
		return resp, err
	}

	c.constraintsMux.Lock()
	defer c.constraintsMux.Unlock()

	if err := c.backend.driver.PutModules(ctx, artifacts.namePrefix, artifacts.modules); err != nil {
		return resp, err
	}

	cpy := templ.DeepCopy()
	cpy.Status = templates.ConstraintTemplateStatus{}
	c.templates[artifacts.Key()] = &templateEntry{
		template: cpy,
		CRD:      artifacts.crd,
		Targets:  []string{artifacts.targetHandler.GetName()},
	}
	if _, ok := c.constraints[artifacts.gk]; !ok {
		c.constraints[artifacts.gk] = make(map[string]*unstructured.Unstructured)
	}
	resp.Handled[artifacts.targetHandler.GetName()] = true
	return resp, nil
}

// RemoveTemplate removes the template source code from OPA and removes the CRD from the validation
// registry. Any constraints relying on the template will also be removed.
// On error, the responses return value will still be populated so that
// partial results can be analyzed.
func (c *Client) RemoveTemplate(ctx context.Context, templ *templates.ConstraintTemplate) (*types.Responses, error) {
	resp := types.NewResponses()

	rawArtifacts, err := c.createRawTemplateArtifacts(templ)
	if err != nil {
		return resp, err
	}

	c.constraintsMux.Lock()
	defer c.constraintsMux.Unlock()

	template, err := c.getTemplateNoLock(ctx, rawArtifacts)
	if err != nil {
		if IsMissingTemplateError(err) {
			return resp, nil
		}
		return resp, err
	}

	artifacts, err := c.createBasicTemplateArtifacts(template)
	if err != nil {
		return resp, err
	}

	if _, err := c.backend.driver.DeleteModules(ctx, artifacts.namePrefix); err != nil {
		return resp, err
	}

	for _, cstr := range c.constraints[artifacts.gk] {
		if r, err := c.removeConstraintNoLock(ctx, cstr); err != nil {
			return r, err
		}
	}
	delete(c.constraints, artifacts.gk)
	// Also clean up root path to avoid memory leaks
	constraintRoot := createConstraintGKPath(artifacts.targetHandler.GetName(), artifacts.gk)
	if _, err := c.backend.driver.DeleteData(ctx, constraintRoot); err != nil {
		return resp, err
	}
	delete(c.templates, artifacts.Key())
	resp.Handled[artifacts.targetHandler.GetName()] = true
	return resp, nil
}

// GetTemplate gets the currently recognized template.
func (c *Client) GetTemplate(ctx context.Context, templ *templates.ConstraintTemplate) (*templates.ConstraintTemplate, error) {

	artifacts, err := c.createRawTemplateArtifacts(templ)
	if err != nil {
		return nil, err
	}

	c.constraintsMux.Lock()
	defer c.constraintsMux.Unlock()
	return c.getTemplateNoLock(ctx, artifacts)
}

func (c *Client) getTemplateNoLock(ctx context.Context, artifacts keyableArtifact) (*templates.ConstraintTemplate, error) {
	t, ok := c.templates[artifacts.Key()]
	if !ok {
		return nil, NewMissingTemplateError(string(artifacts.Key()))
	}
	ret := t.template.DeepCopy()
	return ret, nil
}

// createConstraintSubPath returns the key where we will store the constraint
// for each target: cluster.<group>.<kind>.<name>
func createConstraintSubPath(constraint *unstructured.Unstructured) (string, error) {
	if constraint.GetName() == "" {
		return "", errors.New("Constraint has no name")
	}
	gvk := constraint.GroupVersionKind()
	if gvk.Group == "" {
		return "", fmt.Errorf("Empty group for the constrant named %s", constraint.GetName())
	}
	if gvk.Kind == "" {
		return "", fmt.Errorf("Empty kind for the constraint named %s", constraint.GetName())
	}
	return path.Join(createConstraintGKSubPath(gvk.GroupKind()), constraint.GetName()), nil
}

// createConstraintGKPath returns the subpath for given a constraint GK
func createConstraintGKSubPath(gk schema.GroupKind) string {
	return "/" + path.Join("cluster", gk.Group, gk.Kind)
}

// createConstraintGKPath returns the storage path for a given constrain GK: constraints.<target>.cluster.<group>.<kind>
func createConstraintGKPath(target string, gk schema.GroupKind) string {
	return constraintPathMerge(target, createConstraintGKSubPath(gk))
}

// createConstraintPath returns the storage path for a given constraint: constraints.<target>.cluster.<group>.<kind>.<name>
func createConstraintPath(target string, constraint *unstructured.Unstructured) (string, error) {
	p, err := createConstraintSubPath(constraint)
	if err != nil {
		return "", err
	}
	return constraintPathMerge(target, p), nil
}

// constraintPathMerge is a shared function for creating constraint paths to
// ensure uniformity, it is not meant to be called directly
func constraintPathMerge(target, subpath string) string {
	return "/" + path.Join("constraints", target, subpath)
}

// getTemplateEntry returns the template entry for a given constraint
func (c *Client) getTemplateEntry(constraint *unstructured.Unstructured, lock bool) (*templateEntry, error) {
	kind := constraint.GetKind()
	if kind == "" {
		return nil, fmt.Errorf("Constraint %s has no kind", constraint.GetName())
	}
	if constraint.GroupVersionKind().Group != constraintGroup {
		return nil, fmt.Errorf("Constraint %s has the wrong group", constraint.GetName())
	}
	if lock {
		c.constraintsMux.RLock()
		defer c.constraintsMux.RUnlock()
	}
	entry, ok := c.templates[templateKeyFromConstraint(constraint)]
	if !ok {
		return nil, NewUnrecognizedConstraintError(kind)
	}
	return entry, nil
}

// AddConstraint validates the constraint and, if valid, inserts it into OPA.
// On error, the responses return value will still be populated so that
// partial results can be analyzed.
func (c *Client) AddConstraint(ctx context.Context, constraint *unstructured.Unstructured) (*types.Responses, error) {
	c.constraintsMux.RLock()
	defer c.constraintsMux.RUnlock()
	resp := types.NewResponses()
	errMap := make(ErrorMap)
	entry, err := c.getTemplateEntry(constraint, false)
	if err != nil {
		return resp, err
	}
	subPath, err := createConstraintSubPath(constraint)
	if err != nil {
		return resp, err
	}
	// return immediately if no change
	if cached, err := c.getConstraintNoLock(ctx, constraint); err == nil && constraintlib.SemanticEqual(cached, constraint) {
		for _, target := range entry.Targets {
			resp.Handled[target] = true
		}
		return resp, nil
	}
	if err := c.validateConstraint(constraint, false); err != nil {
		return resp, err
	}
	for _, target := range entry.Targets {
		path, err := createConstraintPath(target, constraint)
		// If we ever create multi-target constraints we will need to handle this more cleverly.
		// the short-circuiting question, cleanup, etc.
		if err != nil {
			errMap[target] = err
			continue
		}
		if err := c.backend.driver.PutData(ctx, path, constraint.Object); err != nil {
			errMap[target] = err
			continue
		}
		resp.Handled[target] = true
	}
	if len(errMap) == 0 {
		c.constraints[constraint.GroupVersionKind().GroupKind()][subPath] = constraint.DeepCopy()
		return resp, nil
	}
	return resp, errMap
}

// RemoveConstraint removes a constraint from OPA. On error, the responses
// return value will still be populated so that partial results can be analyzed.
func (c *Client) RemoveConstraint(ctx context.Context, constraint *unstructured.Unstructured) (*types.Responses, error) {
	c.constraintsMux.RLock()
	defer c.constraintsMux.RUnlock()
	return c.removeConstraintNoLock(ctx, constraint)
}

func (c *Client) removeConstraintNoLock(ctx context.Context, constraint *unstructured.Unstructured) (*types.Responses, error) {
	resp := types.NewResponses()
	errMap := make(ErrorMap)
	entry, err := c.getTemplateEntry(constraint, false)
	if err != nil {
		return resp, err
	}
	subPath, err := createConstraintSubPath(constraint)
	if err != nil {
		return resp, err
	}
	for _, target := range entry.Targets {
		path, err := createConstraintPath(target, constraint)
		// If we ever create multi-target constraints we will need to handle this more cleverly.
		// the short-circuiting question, cleanup, etc.
		if err != nil {
			errMap[target] = err
			continue
		}
		if _, err := c.backend.driver.DeleteData(ctx, path); err != nil {
			errMap[target] = err
		}
		resp.Handled[target] = true
	}
	if len(errMap) == 0 {
		// If we ever create multi-target constraints we will need to handle this more cleverly.
		// the short-circuiting question, cleanup, etc.
		delete(c.constraints[constraint.GroupVersionKind().GroupKind()], subPath)
		return resp, nil
	}
	return resp, errMap
}

// getConstraintNoLock gets the currently recognized constraint without the lock
func (c *Client) getConstraintNoLock(ctx context.Context, constraint *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	subPath, err := createConstraintSubPath(constraint)
	if err != nil {
		return nil, err
	}

	cstr, ok := c.constraints[constraint.GroupVersionKind().GroupKind()][subPath]
	if !ok {
		return nil, NewMissingConstraintError(subPath)
	}
	return cstr.DeepCopy(), nil
}

// GetConstraint gets the currently recognized constraint.
func (c *Client) GetConstraint(ctx context.Context, constraint *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	c.constraintsMux.Lock()
	defer c.constraintsMux.Unlock()
	return c.getConstraintNoLock(ctx, constraint)
}

// validateConstraint is an internal function that allows us to toggle whether we use a read lock
// when validating a constraint
func (c *Client) validateConstraint(constraint *unstructured.Unstructured, lock bool) error {
	entry, err := c.getTemplateEntry(constraint, lock)
	if err != nil {
		return err
	}
	if err = c.backend.crd.validateCR(constraint, entry.CRD); err != nil {
		return err
	}

	for _, target := range entry.Targets {
		if err := c.targets[target].ValidateConstraint(constraint); err != nil {
			return err
		}
	}
	return nil
}

// ValidateConstraint returns an error if the constraint is not recognized or does not conform to
// the registered CRD for that constraint.
func (c *Client) ValidateConstraint(ctx context.Context, constraint *unstructured.Unstructured) error {
	return c.validateConstraint(constraint, true)
}

// init initializes the OPA backend for the client
func (c *Client) init() error {
	for _, t := range c.targets {
		hooks := fmt.Sprintf(`hooks["%s"]`, t.GetName())
		templMap := map[string]string{"Target": t.GetName()}

		libBuiltin := &bytes.Buffer{}
		if err := regolib.TargetLib.Execute(libBuiltin, templMap); err != nil {
			return err
		}
		if err := c.backend.driver.PutModule(
			context.Background(),
			fmt.Sprintf("%s.hooks_builtin", hooks),
			libBuiltin.String()); err != nil {
			return err
		}

		libTempl := t.Library()
		if libTempl == nil {
			return fmt.Errorf("Target %s has no Rego library template", t.GetName())
		}
		libBuf := &bytes.Buffer{}
		if err := libTempl.Execute(libBuf, map[string]string{
			"ConstraintsRoot": fmt.Sprintf(`data.constraints["%s"].cluster["%s"]`, t.GetName(), constraintGroup),
			"DataRoot":        fmt.Sprintf(`data.external["%s"]`, t.GetName()),
		}); err != nil {
			return err
		}
		lib := libBuf.String()
		req := map[string]struct{}{
			"autoreject_review":                {},
			"matching_reviews_and_constraints": {},
			"matching_constraints":             {},
		}
		path := fmt.Sprintf("%s.library", hooks)
		libModule, err := parseModule(path, lib)
		if err != nil {
			return errors.Wrapf(err, "failed to parse module")
		}
		if err := requireRulesModule(libModule, req); err != nil {
			return fmt.Errorf("Problem with the below Rego for %s target:\n\n====%s\n====\n%s", t.GetName(), lib, err)
		}
		err = rewriteModulePackage(path, libModule)
		if err != nil {
			return err
		}
		src, err := format.Ast(libModule)
		if err != nil {
			return fmt.Errorf("Could not re-format Rego source: %v", err)
		}
		if err := c.backend.driver.PutModule(context.Background(), path, string(src)); err != nil {
			return fmt.Errorf("Error %s from compiled source:\n%s", err, src)
		}
	}

	return nil
}

// Reset the state of OPA.
func (c *Client) Reset(ctx context.Context) error {
	c.constraintsMux.Lock()
	defer c.constraintsMux.Unlock()
	for name := range c.targets {
		if _, err := c.backend.driver.DeleteData(ctx, fmt.Sprintf("/external/%s", name)); err != nil {
			return err
		}
		if _, err := c.backend.driver.DeleteData(ctx, fmt.Sprintf("/constraints/%s", name)); err != nil {
			return err
		}
	}
	for name, v := range c.templates {
		for _, t := range v.Targets {
			if _, err := c.backend.driver.DeleteModule(ctx, fmt.Sprintf(`templates["%s"]["%s"]`, t, name)); err != nil {
				return err
			}
		}
	}
	c.templates = make(map[templateKey]*templateEntry)
	c.constraints = make(map[schema.GroupKind]map[string]*unstructured.Unstructured)
	return nil
}

type queryCfg struct {
	enableTracing bool
}

type QueryOpt func(*queryCfg)

func Tracing(enabled bool) QueryOpt {
	return func(cfg *queryCfg) {
		cfg.enableTracing = enabled
	}
}

// Review makes sure the provided object satisfies all stored constraints.
// On error, the responses return value will still be populated so that
// partial results can be analyzed.
func (c *Client) Review(ctx context.Context, obj interface{}, opts ...QueryOpt) (*types.Responses, error) {
	cfg := &queryCfg{}
	for _, opt := range opts {
		opt(cfg)
	}
	responses := types.NewResponses()
	errMap := make(ErrorMap)
TargetLoop:
	for name, target := range c.targets {
		handled, review, err := target.HandleReview(obj)
		// Short-circuiting question applies here as well
		if err != nil {
			errMap[name] = err
			continue
		}
		if !handled {
			continue
		}
		input := map[string]interface{}{"review": review}
		resp, err := c.backend.driver.Query(ctx, fmt.Sprintf(`hooks["%s"].violation`, name), input, drivers.Tracing(cfg.enableTracing))
		if err != nil {
			errMap[name] = err
			continue
		}
		for _, r := range resp.Results {
			if err := target.HandleViolation(r); err != nil {
				errMap[name] = err
				continue TargetLoop
			}
		}
		resp.Target = name
		responses.ByTarget[name] = resp
	}
	if len(errMap) == 0 {
		return responses, nil
	}
	return responses, errMap
}

// Audit makes sure the cached state of the system satisfies all stored constraints.
// On error, the responses return value will still be populated so that
// partial results can be analyzed.
func (c *Client) Audit(ctx context.Context, opts ...QueryOpt) (*types.Responses, error) {
	cfg := &queryCfg{}
	for _, opt := range opts {
		opt(cfg)
	}
	responses := types.NewResponses()
	errMap := make(ErrorMap)
TargetLoop:
	for name, target := range c.targets {
		// Short-circuiting question applies here as well
		resp, err := c.backend.driver.Query(ctx, fmt.Sprintf(`hooks["%s"].audit`, name), nil, drivers.Tracing(cfg.enableTracing))
		if err != nil {
			errMap[name] = err
			continue
		}
		for _, r := range resp.Results {
			if err := target.HandleViolation(r); err != nil {
				errMap[name] = err
				continue TargetLoop
			}
		}
		resp.Target = name
		responses.ByTarget[name] = resp
	}
	if len(errMap) == 0 {
		return responses, nil
	}
	return responses, errMap
}

// Dump dumps the state of OPA to aid in debugging.
func (c *Client) Dump(ctx context.Context) (string, error) {
	return c.backend.driver.Dump(ctx)
}
