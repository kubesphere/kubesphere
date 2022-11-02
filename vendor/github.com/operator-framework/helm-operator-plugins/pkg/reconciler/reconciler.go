/*
Copyright 2020 The Operator-SDK Authors.

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

package reconciler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	sdkhandler "github.com/operator-framework/operator-lib/handler"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	ctrlpredicate "sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/operator-framework/helm-operator-plugins/internal/sdk/controllerutil"
	"github.com/operator-framework/helm-operator-plugins/pkg/annotation"
	helmclient "github.com/operator-framework/helm-operator-plugins/pkg/client"
	"github.com/operator-framework/helm-operator-plugins/pkg/hook"
	"github.com/operator-framework/helm-operator-plugins/pkg/reconciler/internal/conditions"
	"github.com/operator-framework/helm-operator-plugins/pkg/reconciler/internal/diff"
	internalhook "github.com/operator-framework/helm-operator-plugins/pkg/reconciler/internal/hook"
	"github.com/operator-framework/helm-operator-plugins/pkg/reconciler/internal/updater"
	internalvalues "github.com/operator-framework/helm-operator-plugins/pkg/reconciler/internal/values"
	"github.com/operator-framework/helm-operator-plugins/pkg/values"
)

const uninstallFinalizer = "uninstall-helm-release"

// Reconciler reconciles a Helm object
type Reconciler struct {
	client             client.Client
	actionClientGetter helmclient.ActionClientGetter
	valueTranslator    values.Translator
	valueMapper        values.Mapper // nolint:staticcheck
	eventRecorder      record.EventRecorder
	preHooks           []hook.PreHook
	postHooks          []hook.PostHook

	log                              logr.Logger
	gvk                              *schema.GroupVersionKind
	chrt                             *chart.Chart
	selectorPredicate                predicate.Predicate
	overrideValues                   map[string]string
	skipDependentWatches             bool
	maxConcurrentReconciles          int
	reconcilePeriod                  time.Duration
	maxHistory                       int
	skipPrimaryGVKSchemeRegistration bool

	annotSetupOnce       sync.Once
	annotations          map[string]struct{}
	installAnnotations   map[string]annotation.Install
	upgradeAnnotations   map[string]annotation.Upgrade
	uninstallAnnotations map[string]annotation.Uninstall
}

// New creates a new Reconciler that reconciles custom resources that define a
// Helm release. New takes variadic Option arguments that are used to configure
// the Reconciler.
//
// Required options are:
//   - WithGroupVersionKind
//   - WithChart
//
// Other options are defaulted to sane defaults when SetupWithManager is called.
//
// If an error occurs configuring or validating the Reconciler, it is returned.
func New(opts ...Option) (*Reconciler, error) {
	r := &Reconciler{}
	r.annotSetupOnce.Do(r.setupAnnotationMaps)
	for _, o := range opts {
		if err := o(r); err != nil {
			return nil, err
		}
	}

	if err := r.validate(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Reconciler) setupAnnotationMaps() {
	r.annotations = make(map[string]struct{})
	r.installAnnotations = make(map[string]annotation.Install)
	r.upgradeAnnotations = make(map[string]annotation.Upgrade)
	r.uninstallAnnotations = make(map[string]annotation.Uninstall)
}

// SetupWithManager configures a controller for the Reconciler and registers
// watches. It also uses the passed Manager to initialize default values for the
// Reconciler and sets up the manager's scheme with the Reconciler's configured
// GroupVersionKind.
//
// If an error occurs setting up the Reconciler with the manager, it is
// returned.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	controllerName := fmt.Sprintf("%v-controller", strings.ToLower(r.gvk.Kind))

	r.addDefaults(mgr, controllerName)
	if !r.skipPrimaryGVKSchemeRegistration {
		r.setupScheme(mgr)
	}

	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: r, MaxConcurrentReconciles: r.maxConcurrentReconciles})
	if err != nil {
		return err
	}

	if err := r.setupWatches(mgr, c); err != nil {
		return err
	}

	r.log.Info("Watching resource",
		"group", r.gvk.Group,
		"version", r.gvk.Version,
		"kind", r.gvk.Kind,
	)

	return nil
}

// Option is a function that configures the helm Reconciler.
type Option func(r *Reconciler) error

// WithClient is an Option that configures a Reconciler's client.
//
// By default, manager.GetClient() is used if this option is not configured.
func WithClient(cl client.Client) Option {
	return func(r *Reconciler) error {
		r.client = cl
		return nil
	}
}

// WithActionClientGetter is an Option that configures a Reconciler's
// ActionClientGetter.
//
// A default ActionClientGetter is used if this option is not configured.
func WithActionClientGetter(actionClientGetter helmclient.ActionClientGetter) Option {
	return func(r *Reconciler) error {
		r.actionClientGetter = actionClientGetter
		return nil
	}
}

// WithEventRecorder is an Option that configures a Reconciler's EventRecorder.
//
// By default, manager.GetEventRecorderFor() is used if this option is not
// configured.
func WithEventRecorder(er record.EventRecorder) Option {
	return func(r *Reconciler) error {
		r.eventRecorder = er
		return nil
	}
}

// WithLog is an Option that configures a Reconciler's logger.
//
// A default logger is used if this option is not configured.
func WithLog(log logr.Logger) Option {
	return func(r *Reconciler) error {
		r.log = log
		return nil
	}
}

// WithGroupVersionKind is an Option that configures a Reconciler's
// GroupVersionKind.
//
// This option is required.
func WithGroupVersionKind(gvk schema.GroupVersionKind) Option {
	return func(r *Reconciler) error {
		r.gvk = &gvk
		return nil
	}
}

// WithChart is an Option that configures a Reconciler's helm chart.
//
// This option is required.
func WithChart(chrt chart.Chart) Option {
	return func(r *Reconciler) error {
		r.chrt = &chrt
		return nil
	}
}

// WithOverrideValues is an Option that configures a Reconciler's override
// values.
//
// Override values can be used to enforce that certain values provided by the
// chart's default values.yaml or by a CR spec are always overridden when
// rendering the chart. If a value in overrides is set by a CR, it is
// overridden by the override value. The override value can be static but can
// also refer to an environment variable.
//
// If an environment variable reference is listed in override values but is not
// present in the environment when this function runs, it will resolve to an
// empty string and override all other values. Therefore, when using
// environment variable expansion, ensure that the environment variable is set.
func WithOverrideValues(overrides map[string]string) Option {
	return func(r *Reconciler) error {
		// Validate that overrides can be parsed and applied
		// so that we fail fast during operator setup rather
		// than during the first reconciliation.
		obj := &unstructured.Unstructured{Object: map[string]interface{}{"spec": map[string]interface{}{}}}
		if err := internalvalues.ApplyOverrides(overrides, obj); err != nil {
			return err
		}

		r.overrideValues = overrides
		return nil
	}
}

// WithDependentWatchesEnabled is an Option that configures whether the
// Reconciler will register watches for dependent objects in releases and
// trigger reconciliations when they change.
//
// By default, dependent watches are enabled.
func SkipDependentWatches(skip bool) Option {
	return func(r *Reconciler) error {
		r.skipDependentWatches = skip
		return nil
	}
}

// SkipPrimaryGVKSchemeRegistration is an Option that allows to disable the default behaviour of
// registering unstructured.Unstructured as underlying type for the GVK scheme.
//
// Disabling this built-in registration is necessary when building operators
// for which it is desired to have the underlying GVK scheme backed by a
// custom struct type.
//
// Example for using a custom type for the GVK scheme instead of unstructured.Unstructured:
//
//   // Define custom type for GVK scheme.
//   //+kubebuilder:object:root=true
//   type Custom struct {
//     // [...]
//   }
//
//   // Register custom type along with common meta types in scheme.
//   scheme := runtime.NewScheme()
//   scheme.AddKnownTypes(SchemeGroupVersion, &Custom{})
//   metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
//
//   // Create new manager using the controller-runtime, injecting above scheme.
//   options := ctrl.Options{
//     Scheme = scheme,
//     // [...]
//   }
//   mgr, err := ctrl.NewManager(config, options)
//
//   // Create reconciler with generic scheme registration being disabled.
//   r, err := reconciler.New(
//     reconciler.WithChart(chart),
//     reconciler.SkipPrimaryGVKSchemeRegistration(true),
//     // [...]
//   )
//
//   // Setup reconciler with above manager.
//   err = r.SetupWithManager(mgr)
//
// By default, skipping of the generic scheme setup is disabled, which means that
// unstructured.Unstructured is used for the GVK scheme.
func SkipPrimaryGVKSchemeRegistration(skip bool) Option {
	return func(r *Reconciler) error {
		r.skipPrimaryGVKSchemeRegistration = skip
		return nil
	}
}

// WithMaxConcurrentReconciles is an Option that configures the number of
// concurrent reconciles that the controller will run.
//
// The default is 1.
func WithMaxConcurrentReconciles(max int) Option {
	return func(r *Reconciler) error {
		if max < 1 {
			return errors.New("maxConcurrentReconciles must be at least 1")
		}
		r.maxConcurrentReconciles = max
		return nil
	}
}

// WithReconcilePeriod is an Option that configures the reconcile period of the
// controller. This will cause the controller to reconcile CRs at least once
// every period. By default, the reconcile period is set to 0, which means no
// time-based reconciliations will occur.
func WithReconcilePeriod(rp time.Duration) Option {
	return func(r *Reconciler) error {
		if rp < 0 {
			return errors.New("reconcile period must not be negative")
		}
		r.reconcilePeriod = rp
		return nil
	}
}

// WithMaxReleaseHistory specifies the maximum size of the Helm release history maintained
// on upgrades/rollbacks. Zero (default) means unlimited.
func WithMaxReleaseHistory(maxHistory int) Option {
	return func(r *Reconciler) error {
		if maxHistory < 0 {
			return errors.New("maximum Helm release history size must not be negative")
		}
		r.maxHistory = maxHistory
		return nil
	}
}

// WithInstallAnnotations is an Option that configures Install annotations
// to enable custom action.Install fields to be set based on the value of
// annotations found in the custom resource watched by this reconciler.
// Duplicate annotation names will result in an error.
func WithInstallAnnotations(as ...annotation.Install) Option {
	return func(r *Reconciler) error {
		r.annotSetupOnce.Do(r.setupAnnotationMaps)

		for _, a := range as {
			name := a.Name()
			if _, ok := r.annotations[name]; ok {
				return fmt.Errorf("annotation %q already exists", name)
			}

			r.annotations[name] = struct{}{}
			r.installAnnotations[name] = a
		}
		return nil
	}
}

// WithUpgradeAnnotations is an Option that configures Upgrade annotations
// to enable custom action.Upgrade fields to be set based on the value of
// annotations found in the custom resource watched by this reconciler.
// Duplicate annotation names will result in an error.
func WithUpgradeAnnotations(as ...annotation.Upgrade) Option {
	return func(r *Reconciler) error {
		r.annotSetupOnce.Do(r.setupAnnotationMaps)

		for _, a := range as {
			name := a.Name()
			if _, ok := r.annotations[name]; ok {
				return fmt.Errorf("annotation %q already exists", name)
			}

			r.annotations[name] = struct{}{}
			r.upgradeAnnotations[name] = a
		}
		return nil
	}
}

// WithUninstallAnnotations is an Option that configures Uninstall annotations
// to enable custom action.Uninstall fields to be set based on the value of
// annotations found in the custom resource watched by this reconciler.
// Duplicate annotation names will result in an error.
func WithUninstallAnnotations(as ...annotation.Uninstall) Option {
	return func(r *Reconciler) error {
		r.annotSetupOnce.Do(r.setupAnnotationMaps)

		for _, a := range as {
			name := a.Name()
			if _, ok := r.annotations[name]; ok {
				return fmt.Errorf("annotation %q already exists", name)
			}

			r.annotations[name] = struct{}{}
			r.uninstallAnnotations[name] = a
		}
		return nil
	}
}

// WithPreHook is an Option that configures the reconciler to run the given
// PreHook just before performing any actions (e.g. install, upgrade, uninstall,
// or reconciliation).
func WithPreHook(h hook.PreHook) Option {
	return func(r *Reconciler) error {
		r.preHooks = append(r.preHooks, h)
		return nil
	}
}

// WithPostHook is an Option that configures the reconciler to run the given
// PostHook just after performing any non-uninstall release actions.
func WithPostHook(h hook.PostHook) Option {
	return func(r *Reconciler) error {
		r.postHooks = append(r.postHooks, h)
		return nil
	}
}

// WithValueTranslator is an Option that configures a function that translates a
// custom resource to the values passed to Helm.
// Use this if you need to customize the logic that translates your custom resource to Helm values.
// If you wish to, you can convert the Unstructured that is passed to your Translator to your own
// Custom Resource struct like this:
//
//   import "k8s.io/apimachinery/pkg/runtime"
//   foo := your.Foo{}
//   if err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &foo); err != nil {
//     return nil, err
//   }
//   // work with the type-safe foo
//
// Alternatively, your translator can also work similarly to a Mapper, by accessing the spec with:
//
//   u.Object["spec"].(map[string]interface{})
func WithValueTranslator(t values.Translator) Option {
	return func(r *Reconciler) error {
		r.valueTranslator = t
		return nil
	}
}

// WithValueMapper is an Option that configures a function that maps values
// from a custom resource spec to the values passed to Helm.
// Use this if you want to apply a transformation on the values obtained from your custom resource, before
// they are passed to Helm.
//
// Deprecated: Use WithValueTranslator instead.
// WithValueMapper will be removed in a future release.
func WithValueMapper(m values.Mapper) Option {
	return func(r *Reconciler) error {
		r.valueMapper = m
		return nil
	}
}

// WithSelector is an Option that configures the reconciler to creates a
// predicate that is used to filter resources based on the specified selector
func WithSelector(s metav1.LabelSelector) Option {
	return func(r *Reconciler) error {
		p, err := ctrlpredicate.LabelSelectorPredicate(s)
		if err != nil {
			return err
		}
		r.selectorPredicate = p
		return nil
	}
}

// Reconcile reconciles a CR that defines a Helm v3 release.
//
//   - If a release does not exist for this CR, a new release is installed.
//   - If a release exists and the CR spec has changed since the last,
//     reconciliation, the release is upgraded.
//   - If a release exists and the CR spec has not changed since the last
//     reconciliation, the release is reconciled. Any dependent resources that
//     have diverged from the release manifest are re-created or patched so that
//     they are re-aligned with the release.
//   - If the CR has been deleted, the release will be uninstalled. The
//     Reconciler uses a finalizer to ensure the release uninstall succeeds
//     before CR deletion occurs.
//
// If an error occurs during release installation or upgrade, the change will be
// rolled back to restore the previous state.
//
// Reconcile also manages the status field of the custom resource. It includes
// the release name and manifest in `status.deployedRelease`, and it updates
// `status.conditions` based on reconciliation progress and success. Condition
// types include:
//
//   - Deployed - a release for this CR is deployed (but not necessarily ready).
//   - ReleaseFailed - an installation or upgrade failed.
//   - Irreconcilable - an error occurred during reconciliation
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	log := r.log.WithValues(strings.ToLower(r.gvk.Kind), req.NamespacedName)
	log.V(1).Info("Reconciliation triggered")

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(*r.gvk)
	err = r.client.Get(ctx, req.NamespacedName, obj)
	if apierrors.IsNotFound(err) {
		log.V(1).Info("Resource %s/%s not found, nothing to do", req.NamespacedName.Namespace, req.NamespacedName.Name)
		return ctrl.Result{}, nil
	}
	if err != nil {
		return ctrl.Result{}, err
	}

	u := updater.New(r.client)
	defer func() {
		applyErr := u.Apply(ctx, obj)
		if err == nil && !apierrors.IsNotFound(applyErr) {
			err = applyErr
		}
	}()

	actionClient, err := r.actionClientGetter.ActionClientFor(obj)
	if err != nil {
		u.UpdateStatus(
			updater.EnsureCondition(conditions.Irreconcilable(corev1.ConditionTrue, conditions.ReasonErrorGettingClient, err)),
			updater.EnsureConditionUnknown(conditions.TypeDeployed),
			updater.EnsureConditionUnknown(conditions.TypeInitialized),
			updater.EnsureConditionUnknown(conditions.TypeReleaseFailed),
			updater.EnsureDeployedRelease(nil),
		)
		// NOTE: If obj has the uninstall finalizer, that means a release WAS deployed at some point
		//   in the past, but we don't know if it still is because we don't have an actionClient to check.
		//   So the question is, what do we do with the finalizer? We could:
		//      - Leave it in place. This would make the CR impossible to delete without either resolving this error, or
		//        manually uninstalling the release, deleting the finalizer, and deleting the CR.
		//      - Remove the finalizer. This would make it possible to delete the CR, but it would leave around any
		//        release resources that are not owned by the CR (those in the cluster scope or in other namespaces).
		//
		// The decision made for now is to leave the finalizer in place, so that the user can intervene and try to
		// resolve the issue, instead of the operator silently leaving some dangling resources hanging around after the
		// CR is deleted.
		return ctrl.Result{}, err
	}

	// As soon as we get the actionClient, lookup the release and
	// update the status with this info. We need to do this as
	// early as possible in case other irreconcilable errors occur.
	//
	// We also make sure not to return any errors we encounter so
	// we can still attempt an uninstall if the CR is being deleted.
	rel, err := actionClient.Get(obj.GetName())
	if errors.Is(err, driver.ErrReleaseNotFound) {
		u.UpdateStatus(updater.EnsureCondition(conditions.Deployed(corev1.ConditionFalse, "", "")))
	} else if err == nil {
		ensureDeployedRelease(&u, rel)
	}
	u.UpdateStatus(updater.EnsureCondition(conditions.Initialized(corev1.ConditionTrue, "", "")))

	if obj.GetDeletionTimestamp() != nil {
		err := r.handleDeletion(ctx, actionClient, obj, log)
		return ctrl.Result{}, err
	}

	vals, err := r.getValues(ctx, obj)
	if err != nil {
		u.UpdateStatus(
			updater.EnsureCondition(conditions.Irreconcilable(corev1.ConditionTrue, conditions.ReasonErrorGettingValues, err)),
			updater.EnsureConditionUnknown(conditions.TypeReleaseFailed),
		)
		return ctrl.Result{}, err
	}

	rel, state, err := r.getReleaseState(actionClient, obj, vals.AsMap())
	if err != nil {
		u.UpdateStatus(
			updater.EnsureCondition(conditions.Irreconcilable(corev1.ConditionTrue, conditions.ReasonErrorGettingReleaseState, err)),
			updater.EnsureConditionUnknown(conditions.TypeReleaseFailed),
			updater.EnsureConditionUnknown(conditions.TypeDeployed),
			updater.EnsureDeployedRelease(nil),
		)
		return ctrl.Result{}, err
	}
	u.UpdateStatus(updater.EnsureCondition(conditions.Irreconcilable(corev1.ConditionFalse, "", "")))

	for _, h := range r.preHooks {
		if err := h.Exec(obj, vals, log); err != nil {
			log.Error(err, "pre-release hook failed")
		}
	}

	switch state {
	case stateNeedsInstall:
		rel, err = r.doInstall(actionClient, &u, obj, vals.AsMap(), log)
		if err != nil {
			return ctrl.Result{}, err
		}

	case stateNeedsUpgrade:
		rel, err = r.doUpgrade(actionClient, &u, obj, vals.AsMap(), log)
		if err != nil {
			return ctrl.Result{}, err
		}

	case stateUnchanged:
		if err := r.doReconcile(actionClient, &u, rel, log); err != nil {
			return ctrl.Result{}, err
		}
	default:
		return ctrl.Result{}, fmt.Errorf("unexpected release state: %s", state)
	}

	for _, h := range r.postHooks {
		if err := h.Exec(obj, *rel, log); err != nil {
			log.Error(err, "post-release hook failed", "name", rel.Name, "version", rel.Version)
		}
	}

	ensureDeployedRelease(&u, rel)
	u.UpdateStatus(
		updater.EnsureCondition(conditions.ReleaseFailed(corev1.ConditionFalse, "", "")),
		updater.EnsureCondition(conditions.Irreconcilable(corev1.ConditionFalse, "", "")),
	)

	return ctrl.Result{RequeueAfter: r.reconcilePeriod}, nil
}

func (r *Reconciler) getValues(ctx context.Context, obj *unstructured.Unstructured) (chartutil.Values, error) {
	if err := internalvalues.ApplyOverrides(r.overrideValues, obj); err != nil {
		return chartutil.Values{}, err
	}
	vals, err := r.valueTranslator.Translate(ctx, obj)
	if err != nil {
		return chartutil.Values{}, err
	}
	vals = r.valueMapper.Map(vals)
	vals, err = chartutil.CoalesceValues(r.chrt, vals)
	if err != nil {
		return chartutil.Values{}, err
	}
	return vals, nil
}

type helmReleaseState string

const (
	stateNeedsInstall helmReleaseState = "needs install"
	stateNeedsUpgrade helmReleaseState = "needs upgrade"
	stateUnchanged    helmReleaseState = "unchanged"
	stateError        helmReleaseState = "error"
)

func (r *Reconciler) handleDeletion(ctx context.Context, actionClient helmclient.ActionInterface, obj *unstructured.Unstructured, log logr.Logger) error {
	if !controllerutil.ContainsFinalizer(obj, uninstallFinalizer) {
		log.Info("Resource is terminated, skipping reconciliation")
		return nil
	}

	// Use defer in a closure so that it executes before we wait for
	// the deletion of the CR. This might seem unnecessary since we're
	// applying changes to the CR after is has a deletion timestamp.
	// However, if uninstall fails, the finalizer will not be removed
	// and we need to be able to update the conditions on the CR to
	// indicate that the uninstall failed.
	if err := func() (err error) {
		uninstallUpdater := updater.New(r.client)
		defer func() {
			applyErr := uninstallUpdater.Apply(ctx, obj)
			if err == nil {
				err = applyErr
			}
		}()
		return r.doUninstall(actionClient, &uninstallUpdater, obj, log)
	}(); err != nil {
		return err
	}

	// Since the client is hitting a cache, waiting for the
	// deletion here will guarantee that the next reconciliation
	// will see that the CR has been deleted and that there's
	// nothing left to do.
	if err := controllerutil.WaitForDeletion(ctx, r.client, obj); err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) getReleaseState(client helmclient.ActionInterface, obj metav1.Object, vals map[string]interface{}) (*release.Release, helmReleaseState, error) {
	currentRelease, err := client.Get(obj.GetName())
	if err != nil && !errors.Is(err, driver.ErrReleaseNotFound) {
		return nil, stateError, err
	}

	if errors.Is(err, driver.ErrReleaseNotFound) {
		return nil, stateNeedsInstall, nil
	}

	var opts []helmclient.UpgradeOption
	if r.maxHistory > 0 {
		opts = append(opts, func(u *action.Upgrade) error {
			u.MaxHistory = r.maxHistory
			return nil
		})
	}
	for name, annot := range r.upgradeAnnotations {
		if v, ok := obj.GetAnnotations()[name]; ok {
			opts = append(opts, annot.UpgradeOption(v))
		}
	}
	opts = append(opts, func(u *action.Upgrade) error {
		u.DryRun = true
		return nil
	})
	specRelease, err := client.Upgrade(obj.GetName(), obj.GetNamespace(), r.chrt, vals, opts...)
	if err != nil {
		return currentRelease, stateError, err
	}
	if specRelease.Manifest != currentRelease.Manifest ||
		currentRelease.Info.Status == release.StatusFailed ||
		currentRelease.Info.Status == release.StatusSuperseded {
		return currentRelease, stateNeedsUpgrade, nil
	}
	return currentRelease, stateUnchanged, nil
}

func (r *Reconciler) doInstall(actionClient helmclient.ActionInterface, u *updater.Updater, obj *unstructured.Unstructured, vals map[string]interface{}, log logr.Logger) (*release.Release, error) {
	var opts []helmclient.InstallOption
	for name, annot := range r.installAnnotations {
		if v, ok := obj.GetAnnotations()[name]; ok {
			opts = append(opts, annot.InstallOption(v))
		}
	}
	rel, err := actionClient.Install(obj.GetName(), obj.GetNamespace(), r.chrt, vals, opts...)
	if err != nil {
		u.UpdateStatus(
			updater.EnsureCondition(conditions.Irreconcilable(corev1.ConditionTrue, conditions.ReasonReconcileError, err)),
			updater.EnsureCondition(conditions.ReleaseFailed(corev1.ConditionTrue, conditions.ReasonInstallError, err)),
		)
		return nil, err
	}
	r.reportOverrideEvents(obj)

	log.Info("Release installed", "name", rel.Name, "version", rel.Version)

	// If log verbosity is higher, output Helm Release Manifest that was installed
	if log.V(4).Enabled() {
		fmt.Println(diff.Generate("", rel.Manifest))
	}

	return rel, nil
}

func (r *Reconciler) doUpgrade(actionClient helmclient.ActionInterface, u *updater.Updater, obj *unstructured.Unstructured, vals map[string]interface{}, log logr.Logger) (*release.Release, error) {
	var opts []helmclient.UpgradeOption
	if r.maxHistory > 0 {
		opts = append(opts, func(u *action.Upgrade) error {
			u.MaxHistory = r.maxHistory
			return nil
		})
	}
	for name, annot := range r.upgradeAnnotations {
		if v, ok := obj.GetAnnotations()[name]; ok {
			opts = append(opts, annot.UpgradeOption(v))
		}
	}

	// Get the current release so we can compare the new release in the diff if the diff is being logged.
	curRel, err := actionClient.Get(obj.GetName())
	if err != nil {
		return nil, fmt.Errorf("could not get the current Helm Release: %w", err)
	}

	rel, err := actionClient.Upgrade(obj.GetName(), obj.GetNamespace(), r.chrt, vals, opts...)
	if err != nil {
		u.UpdateStatus(
			updater.EnsureCondition(conditions.Irreconcilable(corev1.ConditionTrue, conditions.ReasonReconcileError, err)),
			updater.EnsureCondition(conditions.ReleaseFailed(corev1.ConditionTrue, conditions.ReasonUpgradeError, err)),
		)
		return nil, err
	}
	r.reportOverrideEvents(obj)

	log.Info("Release upgraded", "name", rel.Name, "version", rel.Version)

	// If log verbosity is higher, output upgraded Helm Release Manifest
	if log.V(4).Enabled() {
		fmt.Println(diff.Generate(curRel.Manifest, rel.Manifest))
	}
	return rel, nil
}

func (r *Reconciler) reportOverrideEvents(obj runtime.Object) {
	for k, v := range r.overrideValues {
		r.eventRecorder.Eventf(obj, "Warning", "ValueOverridden",
			"Chart value %q overridden to %q by operator", k, v)
	}
}

func (r *Reconciler) doReconcile(actionClient helmclient.ActionInterface, u *updater.Updater, rel *release.Release, log logr.Logger) error {
	// If a change is made to the CR spec that causes a release failure, a
	// ConditionReleaseFailed is added to the status conditions. If that change
	// is then reverted to its previous state, the operator will stop
	// attempting the release and will resume reconciling. In this case, we
	// need to set the ConditionReleaseFailed to false because the failing
	// release is no longer being attempted.
	u.UpdateStatus(
		updater.EnsureCondition(conditions.ReleaseFailed(corev1.ConditionFalse, "", "")),
	)

	if err := actionClient.Reconcile(rel); err != nil {
		u.UpdateStatus(updater.EnsureCondition(conditions.Irreconcilable(corev1.ConditionTrue, conditions.ReasonReconcileError, err)))
		return err
	}

	log.Info("Release reconciled", "name", rel.Name, "version", rel.Version)
	return nil
}

func (r *Reconciler) doUninstall(actionClient helmclient.ActionInterface, u *updater.Updater, obj *unstructured.Unstructured, log logr.Logger) error {
	var opts []helmclient.UninstallOption
	for name, annot := range r.uninstallAnnotations {
		if v, ok := obj.GetAnnotations()[name]; ok {
			opts = append(opts, annot.UninstallOption(v))
		}
	}

	resp, err := actionClient.Uninstall(obj.GetName(), opts...)
	if errors.Is(err, driver.ErrReleaseNotFound) {
		log.Info("Release not found, removing finalizer")
	} else if err != nil {
		u.UpdateStatus(
			updater.EnsureCondition(conditions.Irreconcilable(corev1.ConditionTrue, conditions.ReasonReconcileError, err)),
			updater.EnsureCondition(conditions.ReleaseFailed(corev1.ConditionTrue, conditions.ReasonUninstallError, err)),
		)
		return err
	} else {
		log.Info("Release uninstalled", "name", resp.Release.Name, "version", resp.Release.Version)

		// If log verbosity is higher, output Helm Release Manifest that was uninstalled
		if log.V(4).Enabled() {
			fmt.Println(diff.Generate(resp.Release.Manifest, ""))
		}
	}
	u.Update(updater.RemoveFinalizer(uninstallFinalizer))
	u.UpdateStatus(
		updater.EnsureCondition(conditions.ReleaseFailed(corev1.ConditionFalse, "", "")),
		updater.EnsureCondition(conditions.Deployed(corev1.ConditionFalse, conditions.ReasonUninstallSuccessful, "")),
		updater.RemoveDeployedRelease(),
	)
	return nil
}

func (r *Reconciler) validate() error {
	if r.gvk == nil {
		return errors.New("gvk must not be nil")
	}
	if r.chrt == nil {
		return errors.New("chart must not be nil")
	}
	return nil
}

func (r *Reconciler) addDefaults(mgr ctrl.Manager, controllerName string) {
	if r.client == nil {
		r.client = mgr.GetClient()
	}
	if r.log.GetSink() == nil {
		r.log = ctrl.Log.WithName("controllers").WithName("Helm")
	}
	if r.actionClientGetter == nil {
		actionConfigGetter := helmclient.NewActionConfigGetter(mgr.GetConfig(), mgr.GetRESTMapper(), r.log)
		r.actionClientGetter = helmclient.NewActionClientGetter(actionConfigGetter)
	}
	if r.eventRecorder == nil {
		r.eventRecorder = mgr.GetEventRecorderFor(controllerName)
	}
	if r.valueTranslator == nil {
		r.valueTranslator = internalvalues.DefaultTranslator
	}
	if r.valueMapper == nil {
		r.valueMapper = internalvalues.DefaultMapper
	}
}

func (r *Reconciler) setupScheme(mgr ctrl.Manager) {
	mgr.GetScheme().AddKnownTypeWithName(*r.gvk, &unstructured.Unstructured{})
	metav1.AddToGroupVersion(mgr.GetScheme(), r.gvk.GroupVersion())
}

func (r *Reconciler) setupWatches(mgr ctrl.Manager, c controller.Controller) error {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(*r.gvk)

	var preds []ctrlpredicate.Predicate
	if r.selectorPredicate != nil {
		preds = append(preds, r.selectorPredicate)
	}

	if err := c.Watch(
		&source.Kind{Type: obj},
		&sdkhandler.InstrumentedEnqueueRequestForObject{},
		preds...,
	); err != nil {
		return err
	}

	secret := &corev1.Secret{}
	secret.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Secret",
	})

	if err := c.Watch(
		&source.Kind{Type: secret},
		&handler.EnqueueRequestForOwner{
			OwnerType:    obj,
			IsController: true,
		},
	); err != nil {
		return err
	}

	if !r.skipDependentWatches {
		r.postHooks = append([]hook.PostHook{internalhook.NewDependentResourceWatcher(c, mgr.GetRESTMapper())}, r.postHooks...)
	}
	return nil
}

func ensureDeployedRelease(u *updater.Updater, rel *release.Release) {
	reason := conditions.ReasonInstallSuccessful
	message := "release was successfully installed"
	if rel.Version > 1 {
		reason = conditions.ReasonUpgradeSuccessful
		message = "release was successfully upgraded"
	}
	if rel.Info != nil && len(rel.Info.Notes) > 0 {
		message = rel.Info.Notes
	}
	u.Update(updater.EnsureFinalizer(uninstallFinalizer))
	u.UpdateStatus(
		updater.EnsureCondition(conditions.Deployed(corev1.ConditionTrue, reason, message)),
		updater.EnsureDeployedRelease(rel),
	)
}
