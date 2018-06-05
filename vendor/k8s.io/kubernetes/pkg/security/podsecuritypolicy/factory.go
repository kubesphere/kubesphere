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

	"k8s.io/apimachinery/pkg/util/errors"
	api "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/apis/policy"
	"k8s.io/kubernetes/pkg/security/podsecuritypolicy/apparmor"
	"k8s.io/kubernetes/pkg/security/podsecuritypolicy/capabilities"
	"k8s.io/kubernetes/pkg/security/podsecuritypolicy/group"
	"k8s.io/kubernetes/pkg/security/podsecuritypolicy/seccomp"
	"k8s.io/kubernetes/pkg/security/podsecuritypolicy/selinux"
	"k8s.io/kubernetes/pkg/security/podsecuritypolicy/sysctl"
	"k8s.io/kubernetes/pkg/security/podsecuritypolicy/user"
)

type simpleStrategyFactory struct{}

var _ StrategyFactory = &simpleStrategyFactory{}

func NewSimpleStrategyFactory() StrategyFactory {
	return &simpleStrategyFactory{}
}

func (f *simpleStrategyFactory) CreateStrategies(psp *policy.PodSecurityPolicy, namespace string) (*ProviderStrategies, error) {
	errs := []error{}

	userStrat, err := createUserStrategy(&psp.Spec.RunAsUser)
	if err != nil {
		errs = append(errs, err)
	}

	seLinuxStrat, err := createSELinuxStrategy(&psp.Spec.SELinux)
	if err != nil {
		errs = append(errs, err)
	}

	appArmorStrat, err := createAppArmorStrategy(psp)
	if err != nil {
		errs = append(errs, err)
	}

	seccompStrat, err := createSeccompStrategy(psp)
	if err != nil {
		errs = append(errs, err)
	}

	fsGroupStrat, err := createFSGroupStrategy(&psp.Spec.FSGroup)
	if err != nil {
		errs = append(errs, err)
	}

	supGroupStrat, err := createSupplementalGroupStrategy(&psp.Spec.SupplementalGroups)
	if err != nil {
		errs = append(errs, err)
	}

	capStrat, err := createCapabilitiesStrategy(psp.Spec.DefaultAddCapabilities, psp.Spec.RequiredDropCapabilities, psp.Spec.AllowedCapabilities)
	if err != nil {
		errs = append(errs, err)
	}

	var unsafeSysctls []string
	if ann, found := psp.Annotations[policy.SysctlsPodSecurityPolicyAnnotationKey]; found {
		var err error
		unsafeSysctls, err = policy.SysctlsFromPodSecurityPolicyAnnotation(ann)
		if err != nil {
			errs = append(errs, err)
		}
	}
	sysctlsStrat := createSysctlsStrategy(unsafeSysctls)

	if len(errs) > 0 {
		return nil, errors.NewAggregate(errs)
	}

	strategies := &ProviderStrategies{
		RunAsUserStrategy:         userStrat,
		SELinuxStrategy:           seLinuxStrat,
		AppArmorStrategy:          appArmorStrat,
		FSGroupStrategy:           fsGroupStrat,
		SupplementalGroupStrategy: supGroupStrat,
		CapabilitiesStrategy:      capStrat,
		SeccompStrategy:           seccompStrat,
		SysctlsStrategy:           sysctlsStrat,
	}

	return strategies, nil
}

// createUserStrategy creates a new user strategy.
func createUserStrategy(opts *policy.RunAsUserStrategyOptions) (user.RunAsUserStrategy, error) {
	switch opts.Rule {
	case policy.RunAsUserStrategyMustRunAs:
		return user.NewMustRunAs(opts)
	case policy.RunAsUserStrategyMustRunAsNonRoot:
		return user.NewRunAsNonRoot(opts)
	case policy.RunAsUserStrategyRunAsAny:
		return user.NewRunAsAny(opts)
	default:
		return nil, fmt.Errorf("Unrecognized RunAsUser strategy type %s", opts.Rule)
	}
}

// createSELinuxStrategy creates a new selinux strategy.
func createSELinuxStrategy(opts *policy.SELinuxStrategyOptions) (selinux.SELinuxStrategy, error) {
	switch opts.Rule {
	case policy.SELinuxStrategyMustRunAs:
		return selinux.NewMustRunAs(opts)
	case policy.SELinuxStrategyRunAsAny:
		return selinux.NewRunAsAny(opts)
	default:
		return nil, fmt.Errorf("Unrecognized SELinuxContext strategy type %s", opts.Rule)
	}
}

// createAppArmorStrategy creates a new AppArmor strategy.
func createAppArmorStrategy(psp *policy.PodSecurityPolicy) (apparmor.Strategy, error) {
	return apparmor.NewStrategy(psp.Annotations), nil
}

// createSeccompStrategy creates a new seccomp strategy.
func createSeccompStrategy(psp *policy.PodSecurityPolicy) (seccomp.Strategy, error) {
	return seccomp.NewStrategy(psp.Annotations), nil
}

// createFSGroupStrategy creates a new fsgroup strategy
func createFSGroupStrategy(opts *policy.FSGroupStrategyOptions) (group.GroupStrategy, error) {
	switch opts.Rule {
	case policy.FSGroupStrategyRunAsAny:
		return group.NewRunAsAny()
	case policy.FSGroupStrategyMustRunAs:
		return group.NewMustRunAs(opts.Ranges, fsGroupField)
	default:
		return nil, fmt.Errorf("Unrecognized FSGroup strategy type %s", opts.Rule)
	}
}

// createSupplementalGroupStrategy creates a new supplemental group strategy
func createSupplementalGroupStrategy(opts *policy.SupplementalGroupsStrategyOptions) (group.GroupStrategy, error) {
	switch opts.Rule {
	case policy.SupplementalGroupsStrategyRunAsAny:
		return group.NewRunAsAny()
	case policy.SupplementalGroupsStrategyMustRunAs:
		return group.NewMustRunAs(opts.Ranges, supplementalGroupsField)
	default:
		return nil, fmt.Errorf("Unrecognized SupplementalGroups strategy type %s", opts.Rule)
	}
}

// createCapabilitiesStrategy creates a new capabilities strategy.
func createCapabilitiesStrategy(defaultAddCaps, requiredDropCaps, allowedCaps []api.Capability) (capabilities.Strategy, error) {
	return capabilities.NewDefaultCapabilities(defaultAddCaps, requiredDropCaps, allowedCaps)
}

// createSysctlsStrategy creates a new unsafe sysctls strategy.
func createSysctlsStrategy(sysctlsPatterns []string) sysctl.SysctlsStrategy {
	return sysctl.NewMustMatchPatterns(sysctlsPatterns)
}
