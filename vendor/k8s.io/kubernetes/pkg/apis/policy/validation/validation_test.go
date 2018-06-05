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

package validation

import (
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation/field"
	api "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/apis/policy"
	"k8s.io/kubernetes/pkg/security/apparmor"
	"k8s.io/kubernetes/pkg/security/podsecuritypolicy/seccomp"
	psputil "k8s.io/kubernetes/pkg/security/podsecuritypolicy/util"
)

func TestValidatePodDisruptionBudgetSpec(t *testing.T) {
	minAvailable := intstr.FromString("0%")
	maxUnavailable := intstr.FromString("10%")

	spec := policy.PodDisruptionBudgetSpec{
		MinAvailable:   &minAvailable,
		MaxUnavailable: &maxUnavailable,
	}
	errs := ValidatePodDisruptionBudgetSpec(spec, field.NewPath("foo"))
	if len(errs) == 0 {
		t.Errorf("unexpected success for %v", spec)
	}
}

func TestValidateMinAvailablePodDisruptionBudgetSpec(t *testing.T) {
	successCases := []intstr.IntOrString{
		intstr.FromString("0%"),
		intstr.FromString("1%"),
		intstr.FromString("100%"),
		intstr.FromInt(0),
		intstr.FromInt(1),
		intstr.FromInt(100),
	}
	for _, c := range successCases {
		spec := policy.PodDisruptionBudgetSpec{
			MinAvailable: &c,
		}
		errs := ValidatePodDisruptionBudgetSpec(spec, field.NewPath("foo"))
		if len(errs) != 0 {
			t.Errorf("unexpected failure %v for %v", errs, spec)
		}
	}

	failureCases := []intstr.IntOrString{
		intstr.FromString("1.1%"),
		intstr.FromString("nope"),
		intstr.FromString("-1%"),
		intstr.FromString("101%"),
		intstr.FromInt(-1),
	}
	for _, c := range failureCases {
		spec := policy.PodDisruptionBudgetSpec{
			MinAvailable: &c,
		}
		errs := ValidatePodDisruptionBudgetSpec(spec, field.NewPath("foo"))
		if len(errs) == 0 {
			t.Errorf("unexpected success for %v", spec)
		}
	}
}

func TestValidateMinAvailablePodAndMaxUnavailableDisruptionBudgetSpec(t *testing.T) {
	c1 := intstr.FromString("10%")
	c2 := intstr.FromInt(1)

	spec := policy.PodDisruptionBudgetSpec{
		MinAvailable:   &c1,
		MaxUnavailable: &c2,
	}
	errs := ValidatePodDisruptionBudgetSpec(spec, field.NewPath("foo"))
	if len(errs) == 0 {
		t.Errorf("unexpected success for %v", spec)
	}
}

func TestValidatePodDisruptionBudgetStatus(t *testing.T) {
	successCases := []policy.PodDisruptionBudgetStatus{
		{PodDisruptionsAllowed: 10},
		{CurrentHealthy: 5},
		{DesiredHealthy: 3},
		{ExpectedPods: 2}}
	for _, c := range successCases {
		errors := ValidatePodDisruptionBudgetStatus(c, field.NewPath("status"))
		if len(errors) > 0 {
			t.Errorf("unexpected failure %v for %v", errors, c)
		}
	}
	failureCases := []policy.PodDisruptionBudgetStatus{
		{PodDisruptionsAllowed: -10},
		{CurrentHealthy: -5},
		{DesiredHealthy: -3},
		{ExpectedPods: -2}}
	for _, c := range failureCases {
		errors := ValidatePodDisruptionBudgetStatus(c, field.NewPath("status"))
		if len(errors) == 0 {
			t.Errorf("unexpected success for %v", c)
		}
	}
}

func TestValidatePodDisruptionBudgetUpdate(t *testing.T) {
	c1 := intstr.FromString("10%")
	c2 := intstr.FromInt(1)
	c3 := intstr.FromInt(2)
	oldPdb := &policy.PodDisruptionBudget{}
	pdb := &policy.PodDisruptionBudget{}
	testCases := []struct {
		generations []int64
		name        string
		specs       []policy.PodDisruptionBudgetSpec
		status      []policy.PodDisruptionBudgetStatus
		ok          bool
	}{
		{
			name:        "only update status",
			generations: []int64{int64(2), int64(3)},
			specs: []policy.PodDisruptionBudgetSpec{
				{
					MinAvailable:   &c1,
					MaxUnavailable: &c2,
				},
				{
					MinAvailable:   &c1,
					MaxUnavailable: &c2,
				},
			},
			status: []policy.PodDisruptionBudgetStatus{
				{
					PodDisruptionsAllowed: 10,
					CurrentHealthy:        5,
					ExpectedPods:          2,
				},
				{
					PodDisruptionsAllowed: 8,
					CurrentHealthy:        5,
					DesiredHealthy:        3,
				},
			},
			ok: true,
		},
		{
			name:        "only update pdb spec",
			generations: []int64{int64(2), int64(3)},
			specs: []policy.PodDisruptionBudgetSpec{
				{
					MaxUnavailable: &c2,
				},
				{
					MinAvailable:   &c1,
					MaxUnavailable: &c3,
				},
			},
			status: []policy.PodDisruptionBudgetStatus{
				{
					PodDisruptionsAllowed: 10,
				},
				{
					PodDisruptionsAllowed: 10,
				},
			},
			ok: false,
		},
		{
			name:        "update spec and status",
			generations: []int64{int64(2), int64(3)},
			specs: []policy.PodDisruptionBudgetSpec{
				{
					MaxUnavailable: &c2,
				},
				{
					MinAvailable:   &c1,
					MaxUnavailable: &c3,
				},
			},
			status: []policy.PodDisruptionBudgetStatus{
				{
					PodDisruptionsAllowed: 10,
					CurrentHealthy:        5,
					ExpectedPods:          2,
				},
				{
					PodDisruptionsAllowed: 8,
					CurrentHealthy:        5,
					DesiredHealthy:        3,
				},
			},
			ok: false,
		},
	}

	for i, tc := range testCases {
		oldPdb.Spec = tc.specs[0]
		oldPdb.Generation = tc.generations[0]
		oldPdb.Status = tc.status[0]

		pdb.Spec = tc.specs[1]
		pdb.Generation = tc.generations[1]
		oldPdb.Status = tc.status[1]

		errs := ValidatePodDisruptionBudgetUpdate(oldPdb, pdb)
		if tc.ok && len(errs) > 0 {
			t.Errorf("[%d:%s] unexpected errors: %v", i, tc.name, errs)
		} else if !tc.ok && len(errs) == 0 {
			t.Errorf("[%d:%s] expected errors: %v", i, tc.name, errs)
		}
	}
}

func TestValidatePodSecurityPolicy(t *testing.T) {
	validPSP := func() *policy.PodSecurityPolicy {
		return &policy.PodSecurityPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "foo",
				Annotations: map[string]string{},
			},
			Spec: policy.PodSecurityPolicySpec{
				SELinux: policy.SELinuxStrategyOptions{
					Rule: policy.SELinuxStrategyRunAsAny,
				},
				RunAsUser: policy.RunAsUserStrategyOptions{
					Rule: policy.RunAsUserStrategyRunAsAny,
				},
				FSGroup: policy.FSGroupStrategyOptions{
					Rule: policy.FSGroupStrategyRunAsAny,
				},
				SupplementalGroups: policy.SupplementalGroupsStrategyOptions{
					Rule: policy.SupplementalGroupsStrategyRunAsAny,
				},
				AllowedHostPaths: []policy.AllowedHostPath{
					{PathPrefix: "/foo/bar"},
					{PathPrefix: "/baz/"},
				},
			},
		}
	}

	noUserOptions := validPSP()
	noUserOptions.Spec.RunAsUser.Rule = ""

	noSELinuxOptions := validPSP()
	noSELinuxOptions.Spec.SELinux.Rule = ""

	invalidUserStratType := validPSP()
	invalidUserStratType.Spec.RunAsUser.Rule = "invalid"

	invalidSELinuxStratType := validPSP()
	invalidSELinuxStratType.Spec.SELinux.Rule = "invalid"

	invalidUIDPSP := validPSP()
	invalidUIDPSP.Spec.RunAsUser.Rule = policy.RunAsUserStrategyMustRunAs
	invalidUIDPSP.Spec.RunAsUser.Ranges = []policy.IDRange{{Min: -1, Max: 1}}

	missingObjectMetaName := validPSP()
	missingObjectMetaName.ObjectMeta.Name = ""

	noFSGroupOptions := validPSP()
	noFSGroupOptions.Spec.FSGroup.Rule = ""

	invalidFSGroupStratType := validPSP()
	invalidFSGroupStratType.Spec.FSGroup.Rule = "invalid"

	noSupplementalGroupsOptions := validPSP()
	noSupplementalGroupsOptions.Spec.SupplementalGroups.Rule = ""

	invalidSupGroupStratType := validPSP()
	invalidSupGroupStratType.Spec.SupplementalGroups.Rule = "invalid"

	invalidRangeMinGreaterThanMax := validPSP()
	invalidRangeMinGreaterThanMax.Spec.FSGroup.Ranges = []policy.IDRange{
		{Min: 2, Max: 1},
	}

	invalidRangeNegativeMin := validPSP()
	invalidRangeNegativeMin.Spec.FSGroup.Ranges = []policy.IDRange{
		{Min: -1, Max: 10},
	}

	invalidRangeNegativeMax := validPSP()
	invalidRangeNegativeMax.Spec.FSGroup.Ranges = []policy.IDRange{
		{Min: 1, Max: -10},
	}

	wildcardAllowedCapAndRequiredDrop := validPSP()
	wildcardAllowedCapAndRequiredDrop.Spec.RequiredDropCapabilities = []api.Capability{"foo"}
	wildcardAllowedCapAndRequiredDrop.Spec.AllowedCapabilities = []api.Capability{policy.AllowAllCapabilities}

	requiredCapAddAndDrop := validPSP()
	requiredCapAddAndDrop.Spec.DefaultAddCapabilities = []api.Capability{"foo"}
	requiredCapAddAndDrop.Spec.RequiredDropCapabilities = []api.Capability{"foo"}

	allowedCapListedInRequiredDrop := validPSP()
	allowedCapListedInRequiredDrop.Spec.RequiredDropCapabilities = []api.Capability{"foo"}
	allowedCapListedInRequiredDrop.Spec.AllowedCapabilities = []api.Capability{"foo"}

	invalidAppArmorDefault := validPSP()
	invalidAppArmorDefault.Annotations = map[string]string{
		apparmor.DefaultProfileAnnotationKey: "not-good",
	}
	invalidAppArmorAllowed := validPSP()
	invalidAppArmorAllowed.Annotations = map[string]string{
		apparmor.AllowedProfilesAnnotationKey: apparmor.ProfileRuntimeDefault + ",not-good",
	}

	invalidSysctlPattern := validPSP()
	invalidSysctlPattern.Annotations[policy.SysctlsPodSecurityPolicyAnnotationKey] = "a.*.b"

	invalidSeccompDefault := validPSP()
	invalidSeccompDefault.Annotations = map[string]string{
		seccomp.DefaultProfileAnnotationKey: "not-good",
	}
	invalidSeccompAllowAnyDefault := validPSP()
	invalidSeccompAllowAnyDefault.Annotations = map[string]string{
		seccomp.DefaultProfileAnnotationKey: "*",
	}
	invalidSeccompAllowed := validPSP()
	invalidSeccompAllowed.Annotations = map[string]string{
		seccomp.AllowedProfilesAnnotationKey: api.SeccompProfileRuntimeDefault + ",not-good",
	}

	invalidAllowedHostPathMissingPath := validPSP()
	invalidAllowedHostPathMissingPath.Spec.AllowedHostPaths = []policy.AllowedHostPath{
		{PathPrefix: ""},
	}

	invalidAllowedHostPathBacksteps := validPSP()
	invalidAllowedHostPathBacksteps.Spec.AllowedHostPaths = []policy.AllowedHostPath{
		{PathPrefix: "/dont/allow/backsteps/.."},
	}

	invalidDefaultAllowPrivilegeEscalation := validPSP()
	pe := true
	invalidDefaultAllowPrivilegeEscalation.Spec.DefaultAllowPrivilegeEscalation = &pe

	emptyFlexDriver := validPSP()
	emptyFlexDriver.Spec.Volumes = []policy.FSType{policy.FlexVolume}
	emptyFlexDriver.Spec.AllowedFlexVolumes = []policy.AllowedFlexVolume{{}}

	nonEmptyFlexVolumes := validPSP()
	nonEmptyFlexVolumes.Spec.AllowedFlexVolumes = []policy.AllowedFlexVolume{{Driver: "example/driver"}}

	type testCase struct {
		psp         *policy.PodSecurityPolicy
		errorType   field.ErrorType
		errorDetail string
	}
	errorCases := map[string]testCase{
		"no user options": {
			psp:         noUserOptions,
			errorType:   field.ErrorTypeNotSupported,
			errorDetail: `supported values: "MustRunAs", "MustRunAsNonRoot", "RunAsAny"`,
		},
		"no selinux options": {
			psp:         noSELinuxOptions,
			errorType:   field.ErrorTypeNotSupported,
			errorDetail: `supported values: "MustRunAs", "RunAsAny"`,
		},
		"no fsgroup options": {
			psp:         noFSGroupOptions,
			errorType:   field.ErrorTypeNotSupported,
			errorDetail: `supported values: "MustRunAs", "RunAsAny"`,
		},
		"no sup group options": {
			psp:         noSupplementalGroupsOptions,
			errorType:   field.ErrorTypeNotSupported,
			errorDetail: `supported values: "MustRunAs", "RunAsAny"`,
		},
		"invalid user strategy type": {
			psp:         invalidUserStratType,
			errorType:   field.ErrorTypeNotSupported,
			errorDetail: `supported values: "MustRunAs", "MustRunAsNonRoot", "RunAsAny"`,
		},
		"invalid selinux strategy type": {
			psp:         invalidSELinuxStratType,
			errorType:   field.ErrorTypeNotSupported,
			errorDetail: `supported values: "MustRunAs", "RunAsAny"`,
		},
		"invalid sup group strategy type": {
			psp:         invalidSupGroupStratType,
			errorType:   field.ErrorTypeNotSupported,
			errorDetail: `supported values: "MustRunAs", "RunAsAny"`,
		},
		"invalid fs group strategy type": {
			psp:         invalidFSGroupStratType,
			errorType:   field.ErrorTypeNotSupported,
			errorDetail: `supported values: "MustRunAs", "RunAsAny"`,
		},
		"invalid uid": {
			psp:         invalidUIDPSP,
			errorType:   field.ErrorTypeInvalid,
			errorDetail: "min cannot be negative",
		},
		"missing object meta name": {
			psp:         missingObjectMetaName,
			errorType:   field.ErrorTypeRequired,
			errorDetail: "name or generateName is required",
		},
		"invalid range min greater than max": {
			psp:         invalidRangeMinGreaterThanMax,
			errorType:   field.ErrorTypeInvalid,
			errorDetail: "min cannot be greater than max",
		},
		"invalid range negative min": {
			psp:         invalidRangeNegativeMin,
			errorType:   field.ErrorTypeInvalid,
			errorDetail: "min cannot be negative",
		},
		"invalid range negative max": {
			psp:         invalidRangeNegativeMax,
			errorType:   field.ErrorTypeInvalid,
			errorDetail: "max cannot be negative",
		},
		"non-empty required drops and all caps are allowed by a wildcard": {
			psp:         wildcardAllowedCapAndRequiredDrop,
			errorType:   field.ErrorTypeInvalid,
			errorDetail: "must be empty when all capabilities are allowed by a wildcard",
		},
		"invalid required caps": {
			psp:         requiredCapAddAndDrop,
			errorType:   field.ErrorTypeInvalid,
			errorDetail: "capability is listed in defaultAddCapabilities and requiredDropCapabilities",
		},
		"allowed cap listed in required drops": {
			psp:         allowedCapListedInRequiredDrop,
			errorType:   field.ErrorTypeInvalid,
			errorDetail: "capability is listed in allowedCapabilities and requiredDropCapabilities",
		},
		"invalid AppArmor default profile": {
			psp:         invalidAppArmorDefault,
			errorType:   field.ErrorTypeInvalid,
			errorDetail: "invalid AppArmor profile name: \"not-good\"",
		},
		"invalid AppArmor allowed profile": {
			psp:         invalidAppArmorAllowed,
			errorType:   field.ErrorTypeInvalid,
			errorDetail: "invalid AppArmor profile name: \"not-good\"",
		},
		"invalid sysctl pattern": {
			psp:         invalidSysctlPattern,
			errorType:   field.ErrorTypeInvalid,
			errorDetail: fmt.Sprintf("must have at most 253 characters and match regex %s", SysctlPatternFmt),
		},
		"invalid seccomp default profile": {
			psp:         invalidSeccompDefault,
			errorType:   field.ErrorTypeInvalid,
			errorDetail: "must be a valid seccomp profile",
		},
		"invalid seccomp allow any default profile": {
			psp:         invalidSeccompAllowAnyDefault,
			errorType:   field.ErrorTypeInvalid,
			errorDetail: "must be a valid seccomp profile",
		},
		"invalid seccomp allowed profile": {
			psp:         invalidSeccompAllowed,
			errorType:   field.ErrorTypeInvalid,
			errorDetail: "must be a valid seccomp profile",
		},
		"invalid defaultAllowPrivilegeEscalation": {
			psp:         invalidDefaultAllowPrivilegeEscalation,
			errorType:   field.ErrorTypeInvalid,
			errorDetail: "Cannot set DefaultAllowPrivilegeEscalation to true without also setting AllowPrivilegeEscalation to true",
		},
		"invalid allowed host path empty path": {
			psp:         invalidAllowedHostPathMissingPath,
			errorType:   field.ErrorTypeRequired,
			errorDetail: "is required",
		},
		"invalid allowed host path with backsteps": {
			psp:         invalidAllowedHostPathBacksteps,
			errorType:   field.ErrorTypeInvalid,
			errorDetail: "must not contain '..'",
		},
		"empty flex volume driver": {
			psp:         emptyFlexDriver,
			errorType:   field.ErrorTypeRequired,
			errorDetail: "must specify a driver",
		},
	}

	for k, v := range errorCases {
		errs := ValidatePodSecurityPolicy(v.psp)
		if len(errs) == 0 {
			t.Errorf("%s expected errors but got none", k)
			continue
		}
		if errs[0].Type != v.errorType {
			t.Errorf("[%s] received an unexpected error type.  Expected: '%s' got: '%s'", k, v.errorType, errs[0].Type)
		}
		if errs[0].Detail != v.errorDetail {
			t.Errorf("[%s] received an unexpected error detail.  Expected '%s' got: '%s'", k, v.errorDetail, errs[0].Detail)
		}
	}

	// Update error is different for 'missing object meta name'.
	errorCases["missing object meta name"] = testCase{
		psp:         errorCases["missing object meta name"].psp,
		errorType:   field.ErrorTypeInvalid,
		errorDetail: "field is immutable",
	}

	// Should not be able to update to an invalid policy.
	for k, v := range errorCases {
		v.psp.ResourceVersion = "444" // Required for updates.
		errs := ValidatePodSecurityPolicyUpdate(validPSP(), v.psp)
		if len(errs) == 0 {
			t.Errorf("[%s] expected update errors but got none", k)
			continue
		}
		if errs[0].Type != v.errorType {
			t.Errorf("[%s] received an unexpected error type.  Expected: '%s' got: '%s'", k, v.errorType, errs[0].Type)
		}
		if errs[0].Detail != v.errorDetail {
			t.Errorf("[%s] received an unexpected error detail.  Expected '%s' got: '%s'", k, v.errorDetail, errs[0].Detail)
		}
	}

	mustRunAs := validPSP()
	mustRunAs.Spec.FSGroup.Rule = policy.FSGroupStrategyMustRunAs
	mustRunAs.Spec.SupplementalGroups.Rule = policy.SupplementalGroupsStrategyMustRunAs
	mustRunAs.Spec.RunAsUser.Rule = policy.RunAsUserStrategyMustRunAs
	mustRunAs.Spec.RunAsUser.Ranges = []policy.IDRange{
		{Min: 1, Max: 1},
	}
	mustRunAs.Spec.SELinux.Rule = policy.SELinuxStrategyMustRunAs

	runAsNonRoot := validPSP()
	runAsNonRoot.Spec.RunAsUser.Rule = policy.RunAsUserStrategyMustRunAsNonRoot

	caseInsensitiveAddDrop := validPSP()
	caseInsensitiveAddDrop.Spec.DefaultAddCapabilities = []api.Capability{"foo"}
	caseInsensitiveAddDrop.Spec.RequiredDropCapabilities = []api.Capability{"FOO"}

	caseInsensitiveAllowedDrop := validPSP()
	caseInsensitiveAllowedDrop.Spec.RequiredDropCapabilities = []api.Capability{"FOO"}
	caseInsensitiveAllowedDrop.Spec.AllowedCapabilities = []api.Capability{"foo"}

	validAppArmor := validPSP()
	validAppArmor.Annotations = map[string]string{
		apparmor.DefaultProfileAnnotationKey:  apparmor.ProfileRuntimeDefault,
		apparmor.AllowedProfilesAnnotationKey: apparmor.ProfileRuntimeDefault + "," + apparmor.ProfileNamePrefix + "foo",
	}

	withSysctl := validPSP()
	withSysctl.Annotations[policy.SysctlsPodSecurityPolicyAnnotationKey] = "net.*"

	validSeccomp := validPSP()
	validSeccomp.Annotations = map[string]string{
		seccomp.DefaultProfileAnnotationKey:  api.SeccompProfileRuntimeDefault,
		seccomp.AllowedProfilesAnnotationKey: api.SeccompProfileRuntimeDefault + ",unconfined,localhost/foo,*",
	}

	validDefaultAllowPrivilegeEscalation := validPSP()
	pe = true
	validDefaultAllowPrivilegeEscalation.Spec.DefaultAllowPrivilegeEscalation = &pe
	validDefaultAllowPrivilegeEscalation.Spec.AllowPrivilegeEscalation = true

	flexvolumeWhenFlexVolumesAllowed := validPSP()
	flexvolumeWhenFlexVolumesAllowed.Spec.Volumes = []policy.FSType{policy.FlexVolume}
	flexvolumeWhenFlexVolumesAllowed.Spec.AllowedFlexVolumes = []policy.AllowedFlexVolume{
		{Driver: "example/driver1"},
	}

	flexvolumeWhenAllVolumesAllowed := validPSP()
	flexvolumeWhenAllVolumesAllowed.Spec.Volumes = []policy.FSType{policy.All}
	flexvolumeWhenAllVolumesAllowed.Spec.AllowedFlexVolumes = []policy.AllowedFlexVolume{
		{Driver: "example/driver2"},
	}
	successCases := map[string]struct {
		psp *policy.PodSecurityPolicy
	}{
		"must run as": {
			psp: mustRunAs,
		},
		"run as any": {
			psp: validPSP(),
		},
		"run as non-root (user only)": {
			psp: runAsNonRoot,
		},
		"comparison for add -> drop is case sensitive": {
			psp: caseInsensitiveAddDrop,
		},
		"comparison for allowed -> drop is case sensitive": {
			psp: caseInsensitiveAllowedDrop,
		},
		"valid AppArmor annotations": {
			psp: validAppArmor,
		},
		"with network sysctls": {
			psp: withSysctl,
		},
		"valid seccomp annotations": {
			psp: validSeccomp,
		},
		"valid defaultAllowPrivilegeEscalation as true": {
			psp: validDefaultAllowPrivilegeEscalation,
		},
		"allow white-listed flexVolume when flex volumes are allowed": {
			psp: flexvolumeWhenFlexVolumesAllowed,
		},
		"allow white-listed flexVolume when all volumes are allowed": {
			psp: flexvolumeWhenAllVolumesAllowed,
		},
	}

	for k, v := range successCases {
		if errs := ValidatePodSecurityPolicy(v.psp); len(errs) != 0 {
			t.Errorf("Expected success for %s, got %v", k, errs)
		}

		// Should be able to update to a valid PSP.
		v.psp.ResourceVersion = "444" // Required for updates.
		if errs := ValidatePodSecurityPolicyUpdate(validPSP(), v.psp); len(errs) != 0 {
			t.Errorf("Expected success for %s update, got %v", k, errs)
		}
	}
}

func TestValidatePSPVolumes(t *testing.T) {
	validPSP := func() *policy.PodSecurityPolicy {
		return &policy.PodSecurityPolicy{
			ObjectMeta: metav1.ObjectMeta{Name: "foo"},
			Spec: policy.PodSecurityPolicySpec{
				SELinux: policy.SELinuxStrategyOptions{
					Rule: policy.SELinuxStrategyRunAsAny,
				},
				RunAsUser: policy.RunAsUserStrategyOptions{
					Rule: policy.RunAsUserStrategyRunAsAny,
				},
				FSGroup: policy.FSGroupStrategyOptions{
					Rule: policy.FSGroupStrategyRunAsAny,
				},
				SupplementalGroups: policy.SupplementalGroupsStrategyOptions{
					Rule: policy.SupplementalGroupsStrategyRunAsAny,
				},
			},
		}
	}

	volumes := psputil.GetAllFSTypesAsSet()
	// add in the * value since that is a pseudo type that is not included by default
	volumes.Insert(string(policy.All))

	for _, strVolume := range volumes.List() {
		psp := validPSP()
		psp.Spec.Volumes = []policy.FSType{policy.FSType(strVolume)}
		errs := ValidatePodSecurityPolicy(psp)
		if len(errs) != 0 {
			t.Errorf("%s validation expected no errors but received %v", strVolume, errs)
		}
	}
}

func TestIsValidSysctlPattern(t *testing.T) {
	valid := []string{
		"a.b.c.d",
		"a",
		"a_b",
		"a-b",
		"abc",
		"abc.def",
		"*",
		"a.*",
		"*",
		"abc*",
		"a.abc*",
		"a.b.*",
	}
	invalid := []string{
		"",
		"ä",
		"a_",
		"_",
		"_a",
		"_a._b",
		"__",
		"-",
		".",
		"a.",
		".a",
		"a.b.",
		"a*.b",
		"a*b",
		"*a",
		"Abc",
		func(n int) string {
			x := make([]byte, n)
			for i := range x {
				x[i] = byte('a')
			}
			return string(x)
		}(256),
	}
	for _, s := range valid {
		if !IsValidSysctlPattern(s) {
			t.Errorf("%q expected to be a valid sysctl pattern", s)
		}
	}
	for _, s := range invalid {
		if IsValidSysctlPattern(s) {
			t.Errorf("%q expected to be an invalid sysctl pattern", s)
		}
	}
}

func Test_validatePSPRunAsUser(t *testing.T) {
	var testCases = []struct {
		name              string
		runAsUserStrategy policy.RunAsUserStrategyOptions
		fail              bool
	}{
		{"Invalid RunAsUserStrategy", policy.RunAsUserStrategyOptions{Rule: policy.RunAsUserStrategy("someInvalidStrategy")}, true},
		{"RunAsUserStrategyMustRunAs", policy.RunAsUserStrategyOptions{Rule: policy.RunAsUserStrategyMustRunAs}, false},
		{"RunAsUserStrategyMustRunAsNonRoot", policy.RunAsUserStrategyOptions{Rule: policy.RunAsUserStrategyMustRunAsNonRoot}, false},
		{"RunAsUserStrategyMustRunAsNonRoot With Valid Range", policy.RunAsUserStrategyOptions{Rule: policy.RunAsUserStrategyMustRunAs, Ranges: []policy.IDRange{{Min: 2, Max: 3}, {Min: 4, Max: 5}}}, false},
		{"RunAsUserStrategyMustRunAsNonRoot With Invalid Range", policy.RunAsUserStrategyOptions{Rule: policy.RunAsUserStrategyMustRunAs, Ranges: []policy.IDRange{{Min: 2, Max: 3}, {Min: 5, Max: 4}}}, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			errList := validatePSPRunAsUser(field.NewPath("status"), &testCase.runAsUserStrategy)
			actualErrors := len(errList)
			expectedErrors := 1
			if !testCase.fail {
				expectedErrors = 0
			}
			if actualErrors != expectedErrors {
				t.Errorf("In testCase %v, expected %v errors, got %v errors", testCase.name, expectedErrors, actualErrors)
			}
		})
	}
}
