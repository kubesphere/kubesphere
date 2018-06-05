/*
Copyright 2016 The Kubernetes Authors.

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

package podsecuritypolicy

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/golang/glog"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/admission"
	genericadmissioninit "k8s.io/apiserver/pkg/admission/initializer"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	api "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/apis/policy"
	informers "k8s.io/kubernetes/pkg/client/informers/informers_generated/internalversion"
	policylisters "k8s.io/kubernetes/pkg/client/listers/policy/internalversion"
	kubeapiserveradmission "k8s.io/kubernetes/pkg/kubeapiserver/admission"
	rbacregistry "k8s.io/kubernetes/pkg/registry/rbac"
	psp "k8s.io/kubernetes/pkg/security/podsecuritypolicy"
	psputil "k8s.io/kubernetes/pkg/security/podsecuritypolicy/util"
	"k8s.io/kubernetes/pkg/serviceaccount"
)

const (
	PluginName = "PodSecurityPolicy"
)

// Register registers a plugin
func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(config io.Reader) (admission.Interface, error) {
		plugin := newPlugin(psp.NewSimpleStrategyFactory(), true)
		return plugin, nil
	})
}

// PodSecurityPolicyPlugin holds state for and implements the admission plugin.
type PodSecurityPolicyPlugin struct {
	*admission.Handler
	strategyFactory  psp.StrategyFactory
	failOnNoPolicies bool
	authz            authorizer.Authorizer
	lister           policylisters.PodSecurityPolicyLister
}

// SetAuthorizer sets the authorizer.
func (plugin *PodSecurityPolicyPlugin) SetAuthorizer(authz authorizer.Authorizer) {
	plugin.authz = authz
}

// ValidateInitialization ensures an authorizer is set.
func (plugin *PodSecurityPolicyPlugin) ValidateInitialization() error {
	if plugin.authz == nil {
		return fmt.Errorf("%s requires an authorizer", PluginName)
	}
	if plugin.lister == nil {
		return fmt.Errorf("%s requires a lister", PluginName)
	}
	return nil
}

var _ admission.MutationInterface = &PodSecurityPolicyPlugin{}
var _ admission.ValidationInterface = &PodSecurityPolicyPlugin{}
var _ genericadmissioninit.WantsAuthorizer = &PodSecurityPolicyPlugin{}
var _ kubeapiserveradmission.WantsInternalKubeInformerFactory = &PodSecurityPolicyPlugin{}

// newPlugin creates a new PSP admission plugin.
func newPlugin(strategyFactory psp.StrategyFactory, failOnNoPolicies bool) *PodSecurityPolicyPlugin {
	return &PodSecurityPolicyPlugin{
		Handler:          admission.NewHandler(admission.Create, admission.Update),
		strategyFactory:  strategyFactory,
		failOnNoPolicies: failOnNoPolicies,
	}
}

func (a *PodSecurityPolicyPlugin) SetInternalKubeInformerFactory(f informers.SharedInformerFactory) {
	podSecurityPolicyInformer := f.Policy().InternalVersion().PodSecurityPolicies()
	a.lister = podSecurityPolicyInformer.Lister()
	a.SetReadyFunc(podSecurityPolicyInformer.Informer().HasSynced)
}

// Admit determines if the pod should be admitted based on the requested security context
// and the available PSPs.
//
// 1.  Find available PSPs.
// 2.  Create the providers, includes setting pre-allocated values if necessary.
// 3.  Try to generate and validate a PSP with providers.  If we find one then admit the pod
//     with the validated PSP.  If we don't find any reject the pod and give all errors from the
//     failed attempts.
func (c *PodSecurityPolicyPlugin) Admit(a admission.Attributes) error {
	if ignore, err := shouldIgnore(a); err != nil {
		return err
	} else if ignore {
		return nil
	}

	// only mutate if this is a CREATE request. On updates we only validate.
	// TODO(liggitt): allow spec mutation during initializing updates?
	if a.GetOperation() != admission.Create {
		return nil
	}

	pod := a.GetObject().(*api.Pod)

	// compute the context. Mutation is allowed. ValidatedPSPAnnotation is not taken into account.
	allowedPod, pspName, validationErrs, err := c.computeSecurityContext(a, pod, true, "")
	if err != nil {
		return admission.NewForbidden(a, err)
	}
	if allowedPod != nil {
		*pod = *allowedPod
		// annotate and accept the pod
		glog.V(4).Infof("pod %s (generate: %s) in namespace %s validated against provider %s", pod.Name, pod.GenerateName, a.GetNamespace(), pspName)
		if pod.ObjectMeta.Annotations == nil {
			pod.ObjectMeta.Annotations = map[string]string{}
		}
		pod.ObjectMeta.Annotations[psputil.ValidatedPSPAnnotation] = pspName
		return nil
	}

	// we didn't validate against any provider, reject the pod and give the errors for each attempt
	glog.V(4).Infof("unable to validate pod %s (generate: %s) in namespace %s against any pod security policy: %v", pod.Name, pod.GenerateName, a.GetNamespace(), validationErrs)
	return admission.NewForbidden(a, fmt.Errorf("unable to validate against any pod security policy: %v", validationErrs))
}

func (c *PodSecurityPolicyPlugin) Validate(a admission.Attributes) error {
	if ignore, err := shouldIgnore(a); err != nil {
		return err
	} else if ignore {
		return nil
	}

	pod := a.GetObject().(*api.Pod)

	// compute the context. Mutation is not allowed. ValidatedPSPAnnotation is used as a hint to gain same speed-up.
	allowedPod, _, validationErrs, err := c.computeSecurityContext(a, pod, false, pod.ObjectMeta.Annotations[psputil.ValidatedPSPAnnotation])
	if err != nil {
		return admission.NewForbidden(a, err)
	}
	if apiequality.Semantic.DeepEqual(pod, allowedPod) {
		return nil
	}

	// we didn't validate against any provider, reject the pod and give the errors for each attempt
	glog.V(4).Infof("unable to validate pod %s (generate: %s) in namespace %s against any pod security policy: %v", pod.Name, pod.GenerateName, a.GetNamespace(), validationErrs)
	return admission.NewForbidden(a, fmt.Errorf("unable to validate against any pod security policy: %v", validationErrs))
}

func shouldIgnore(a admission.Attributes) (bool, error) {
	if a.GetResource().GroupResource() != api.Resource("pods") {
		return true, nil
	}
	if len(a.GetSubresource()) != 0 {
		return true, nil
	}

	// if we can't convert then fail closed since we've already checked that this is supposed to be a pod object.
	// this shouldn't normally happen during admission but could happen if an integrator passes a versioned
	// pod object rather than an internal object.
	if _, ok := a.GetObject().(*api.Pod); !ok {
		return false, admission.NewForbidden(a, fmt.Errorf("unexpected type %T", a.GetObject()))
	}

	// if this is an update, see if we are only updating the ownerRef/finalizers.  Garbage collection does this
	// and we should allow it in general, since you had the power to update and the power to delete.
	// The worst that happens is that you delete something, but you aren't controlling the privileged object itself
	if a.GetOperation() == admission.Update && rbacregistry.IsOnlyMutatingGCFields(a.GetObject(), a.GetOldObject(), apiequality.Semantic) {
		return true, nil
	}

	return false, nil
}

// computeSecurityContext derives a valid security context while trying to avoid any changes to the given pod. I.e.
// if there is a matching policy with the same security context as given, it will be reused. If there is no
// matching policy the returned pod will be nil and the pspName empty. validatedPSPHint is the validated psp name
// saved in kubernetes.io/psp annotation. This psp is usually the one we are looking for.
func (c *PodSecurityPolicyPlugin) computeSecurityContext(a admission.Attributes, pod *api.Pod, specMutationAllowed bool, validatedPSPHint string) (*api.Pod, string, field.ErrorList, error) {
	// get all constraints that are usable by the user
	glog.V(4).Infof("getting pod security policies for pod %s (generate: %s)", pod.Name, pod.GenerateName)
	var saInfo user.Info
	if len(pod.Spec.ServiceAccountName) > 0 {
		saInfo = serviceaccount.UserInfo(a.GetNamespace(), pod.Spec.ServiceAccountName, "")
	}

	policies, err := c.lister.List(labels.Everything())
	if err != nil {
		return nil, "", nil, err
	}

	// if we have no policies and want to succeed then return.  Otherwise we'll end up with no
	// providers and fail with "unable to validate against any pod security policy" below.
	if len(policies) == 0 && !c.failOnNoPolicies {
		return pod, "", nil, nil
	}

	// sort policies by name to make order deterministic
	// If mutation is not allowed and validatedPSPHint is provided, check the validated policy first.
	// TODO(liggitt): add priority field to allow admins to bucket differently
	sort.SliceStable(policies, func(i, j int) bool {
		if !specMutationAllowed {
			if policies[i].Name == validatedPSPHint {
				return true
			}
			if policies[j].Name == validatedPSPHint {
				return false
			}
		}
		return strings.Compare(policies[i].Name, policies[j].Name) < 0
	})

	providers, errs := c.createProvidersFromPolicies(policies, pod.Namespace)
	for _, err := range errs {
		glog.V(4).Infof("provider creation error: %v", err)
	}

	if len(providers) == 0 {
		return nil, "", nil, fmt.Errorf("no providers available to validate pod request")
	}

	var (
		allowedMutatedPod   *api.Pod
		allowingMutatingPSP string
		// Map of PSP name to associated validation errors.
		validationErrs = map[string]field.ErrorList{}
	)

	for _, provider := range providers {
		podCopy := pod.DeepCopy()

		if errs := assignSecurityContext(provider, podCopy, field.NewPath(fmt.Sprintf("provider %s: ", provider.GetPSPName()))); len(errs) > 0 {
			validationErrs[provider.GetPSPName()] = errs
			continue
		}

		// the entire pod validated
		mutated := !apiequality.Semantic.DeepEqual(pod, podCopy)
		if mutated && !specMutationAllowed {
			continue
		}

		if !isAuthorizedForPolicy(a.GetUserInfo(), saInfo, a.GetNamespace(), provider.GetPSPName(), c.authz) {
			continue
		}

		switch {
		case !mutated:
			// if it validated without mutating anything, use this result
			return podCopy, provider.GetPSPName(), nil, nil

		case specMutationAllowed && allowedMutatedPod == nil:
			// if mutation is allowed and this is the first PSP to allow the pod, remember it,
			// but continue to see if another PSP allows without mutating
			allowedMutatedPod = podCopy
			allowingMutatingPSP = provider.GetPSPName()
		}
	}

	if allowedMutatedPod != nil {
		return allowedMutatedPod, allowingMutatingPSP, nil, nil
	}

	// Pod is rejected. Filter the validation errors to only include errors from authorized PSPs.
	aggregate := field.ErrorList{}
	for psp, errs := range validationErrs {
		if isAuthorizedForPolicy(a.GetUserInfo(), saInfo, a.GetNamespace(), psp, c.authz) {
			aggregate = append(aggregate, errs...)
		}
	}
	return nil, "", aggregate, nil
}

// assignSecurityContext creates a security context for each container in the pod
// and validates that the sc falls within the psp constraints.  All containers must validate against
// the same psp or is not considered valid.
func assignSecurityContext(provider psp.Provider, pod *api.Pod, fldPath *field.Path) field.ErrorList {
	errs := field.ErrorList{}

	err := provider.DefaultPodSecurityContext(pod)
	if err != nil {
		errs = append(errs, field.Invalid(field.NewPath("spec", "securityContext"), pod.Spec.SecurityContext, err.Error()))
	}

	errs = append(errs, provider.ValidatePod(pod, field.NewPath("spec", "securityContext"))...)

	for i := range pod.Spec.InitContainers {
		err := provider.DefaultContainerSecurityContext(pod, &pod.Spec.InitContainers[i])
		if err != nil {
			errs = append(errs, field.Invalid(field.NewPath("spec", "initContainers").Index(i).Child("securityContext"), "", err.Error()))
			continue
		}
		errs = append(errs, provider.ValidateContainerSecurityContext(pod, &pod.Spec.InitContainers[i], field.NewPath("spec", "initContainers").Index(i).Child("securityContext"))...)
	}

	for i := range pod.Spec.Containers {
		err := provider.DefaultContainerSecurityContext(pod, &pod.Spec.Containers[i])
		if err != nil {
			errs = append(errs, field.Invalid(field.NewPath("spec", "containers").Index(i).Child("securityContext"), "", err.Error()))
			continue
		}
		errs = append(errs, provider.ValidateContainerSecurityContext(pod, &pod.Spec.Containers[i], field.NewPath("spec", "containers").Index(i).Child("securityContext"))...)
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// createProvidersFromPolicies creates providers from the constraints supplied.
func (c *PodSecurityPolicyPlugin) createProvidersFromPolicies(psps []*policy.PodSecurityPolicy, namespace string) ([]psp.Provider, []error) {
	var (
		// collected providers
		providers []psp.Provider
		// collected errors to return
		errs []error
	)

	for _, constraint := range psps {
		provider, err := psp.NewSimpleProvider(constraint, namespace, c.strategyFactory)
		if err != nil {
			errs = append(errs, fmt.Errorf("error creating provider for PSP %s: %v", constraint.Name, err))
			continue
		}
		providers = append(providers, provider)
	}
	return providers, errs
}

func isAuthorizedForPolicy(user, sa user.Info, namespace, policyName string, authz authorizer.Authorizer) bool {
	// Check the service account first, as that is the more common use case.
	return authorizedForPolicy(sa, namespace, policyName, authz) ||
		authorizedForPolicy(user, namespace, policyName, authz)
}

// authorizedForPolicy returns true if info is authorized to perform the "use" verb on the policy resource.
// TODO: check against only the policy group when PSP will be completely moved out of the extensions
func authorizedForPolicy(info user.Info, namespace string, policyName string, authz authorizer.Authorizer) bool {
	// Check against extensions API group for backward compatibility
	return authorizedForPolicyInAPIGroup(info, namespace, policyName, policy.GroupName, authz) ||
		authorizedForPolicyInAPIGroup(info, namespace, policyName, extensions.GroupName, authz)
}

// authorizedForPolicyInAPIGroup returns true if info is authorized to perform the "use" verb on the policy resource in the specified API group.
func authorizedForPolicyInAPIGroup(info user.Info, namespace, policyName, apiGroupName string, authz authorizer.Authorizer) bool {
	if info == nil {
		return false
	}
	attr := buildAttributes(info, namespace, policyName, apiGroupName)
	decision, reason, err := authz.Authorize(attr)
	if err != nil {
		glog.V(5).Infof("cannot authorize for policy: %v,%v", reason, err)
	}
	return (decision == authorizer.DecisionAllow)
}

// buildAttributes builds an attributes record for a SAR based on the user info and policy.
func buildAttributes(info user.Info, namespace, policyName, apiGroupName string) authorizer.Attributes {
	// check against the namespace that the pod is being created in to allow per-namespace PSP grants.
	attr := authorizer.AttributesRecord{
		User:            info,
		Verb:            "use",
		Namespace:       namespace,
		Name:            policyName,
		APIGroup:        apiGroupName,
		Resource:        "podsecuritypolicies",
		ResourceRequest: true,
	}
	return attr
}
