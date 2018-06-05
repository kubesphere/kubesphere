/*
Copyright 2017 The Kubernetes Authors.

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

package auth

import (
	"fmt"

	"k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	policy "k8s.io/api/policy/v1beta1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/authentication/serviceaccount"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/security/apparmor"
	"k8s.io/kubernetes/pkg/security/podsecuritypolicy/seccomp"
	psputil "k8s.io/kubernetes/pkg/security/podsecuritypolicy/util"
	utilpointer "k8s.io/kubernetes/pkg/util/pointer"
	"k8s.io/kubernetes/test/e2e/common"
	"k8s.io/kubernetes/test/e2e/framework"
	imageutils "k8s.io/kubernetes/test/utils/image"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = SIGDescribe("PodSecurityPolicy", func() {
	f := framework.NewDefaultFramework("podsecuritypolicy")
	f.SkipPrivilegedPSPBinding = true

	// Client that will impersonate the default service account, in order to run
	// with reduced privileges.
	var c clientset.Interface
	var ns string // Test namespace, for convenience
	BeforeEach(func() {
		if !framework.IsPodSecurityPolicyEnabled(f) {
			framework.Skipf("PodSecurityPolicy not enabled")
		}
		if !framework.IsRBACEnabled(f) {
			framework.Skipf("RBAC not enabled")
		}
		ns = f.Namespace.Name

		By("Creating a kubernetes client that impersonates the default service account")
		config, err := framework.LoadConfig()
		framework.ExpectNoError(err)
		config.Impersonate = restclient.ImpersonationConfig{
			UserName: serviceaccount.MakeUsername(ns, "default"),
			Groups:   serviceaccount.MakeGroupNames(ns),
		}
		c, err = clientset.NewForConfig(config)
		framework.ExpectNoError(err)

		By("Binding the edit role to the default SA")
		framework.BindClusterRole(f.ClientSet.RbacV1beta1(), "edit", ns,
			rbacv1beta1.Subject{Kind: rbacv1beta1.ServiceAccountKind, Namespace: ns, Name: "default"})
	})

	It("should forbid pod creation when no PSP is available", func() {
		By("Running a restricted pod")
		_, err := c.CoreV1().Pods(ns).Create(restrictedPod(f, "restricted"))
		expectForbidden(err)
	})

	// TODO: merge tests for extensions/policy API groups when PSP will be completely moved out of the extensions

	It("should enforce the restricted extensions.PodSecurityPolicy", func() {
		By("Creating & Binding a restricted policy for the test service account")
		_, cleanup := createAndBindPSP(f, restrictedPSP("restrictive"))
		defer cleanup()

		By("Running a restricted pod")
		pod, err := c.CoreV1().Pods(ns).Create(restrictedPod(f, "allowed"))
		framework.ExpectNoError(err)
		framework.ExpectNoError(framework.WaitForPodNameRunningInNamespace(c, pod.Name, pod.Namespace))

		testPrivilegedPods(f, func(pod *v1.Pod) {
			_, err := c.CoreV1().Pods(ns).Create(pod)
			expectForbidden(err)
		})
	})

	It("should enforce the restricted policy.PodSecurityPolicy", func() {
		By("Creating & Binding a restricted policy for the test service account")
		_, cleanup := createAndBindPSPInPolicy(f, restrictedPSPInPolicy("restrictive"))
		defer cleanup()

		By("Running a restricted pod")
		pod, err := c.CoreV1().Pods(ns).Create(restrictedPod(f, "allowed"))
		framework.ExpectNoError(err)
		framework.ExpectNoError(framework.WaitForPodNameRunningInNamespace(c, pod.Name, pod.Namespace))

		testPrivilegedPods(f, func(pod *v1.Pod) {
			_, err := c.CoreV1().Pods(ns).Create(pod)
			expectForbidden(err)
		})
	})

	It("should allow pods under the privileged extensions.PodSecurityPolicy", func() {
		By("Creating & Binding a privileged policy for the test service account")
		// Ensure that the permissive policy is used even in the presence of the restricted policy.
		_, cleanup := createAndBindPSP(f, restrictedPSP("restrictive"))
		defer cleanup()
		expectedPSP, cleanup := createAndBindPSP(f, framework.PrivilegedPSP("permissive"))
		defer cleanup()

		testPrivilegedPods(f, func(pod *v1.Pod) {
			p, err := c.CoreV1().Pods(ns).Create(pod)
			framework.ExpectNoError(err)
			framework.ExpectNoError(framework.WaitForPodNameRunningInNamespace(c, p.Name, p.Namespace))

			// Verify expected PSP was used.
			p, err = c.CoreV1().Pods(ns).Get(p.Name, metav1.GetOptions{})
			framework.ExpectNoError(err)
			validated, found := p.Annotations[psputil.ValidatedPSPAnnotation]
			Expect(found).To(BeTrue(), "PSP annotation not found")
			Expect(validated).To(Equal(expectedPSP.Name), "Unexpected validated PSP")
		})
	})

	It("should allow pods under the privileged policy.PodSecurityPolicy", func() {
		By("Creating & Binding a privileged policy for the test service account")
		// Ensure that the permissive policy is used even in the presence of the restricted policy.
		_, cleanup := createAndBindPSPInPolicy(f, restrictedPSPInPolicy("restrictive"))
		defer cleanup()
		expectedPSP, cleanup := createAndBindPSPInPolicy(f, privilegedPSPInPolicy("permissive"))
		defer cleanup()

		testPrivilegedPods(f, func(pod *v1.Pod) {
			p, err := c.CoreV1().Pods(ns).Create(pod)
			framework.ExpectNoError(err)
			framework.ExpectNoError(framework.WaitForPodNameRunningInNamespace(c, p.Name, p.Namespace))

			// Verify expected PSP was used.
			p, err = c.CoreV1().Pods(ns).Get(p.Name, metav1.GetOptions{})
			framework.ExpectNoError(err)
			validated, found := p.Annotations[psputil.ValidatedPSPAnnotation]
			Expect(found).To(BeTrue(), "PSP annotation not found")
			Expect(validated).To(Equal(expectedPSP.Name), "Unexpected validated PSP")
		})
	})
})

func expectForbidden(err error) {
	Expect(err).To(HaveOccurred(), "should be forbidden")
	Expect(apierrs.IsForbidden(err)).To(BeTrue(), "should be forbidden error")
}

func testPrivilegedPods(f *framework.Framework, tester func(pod *v1.Pod)) {
	By("Running a privileged pod", func() {
		privileged := restrictedPod(f, "privileged")
		privileged.Spec.Containers[0].SecurityContext.Privileged = boolPtr(true)
		privileged.Spec.Containers[0].SecurityContext.AllowPrivilegeEscalation = nil
		tester(privileged)
	})

	By("Running a HostPath pod", func() {
		hostpath := restrictedPod(f, "hostpath")
		hostpath.Spec.Containers[0].VolumeMounts = []v1.VolumeMount{{
			Name:      "hp",
			MountPath: "/hp",
		}}
		hostpath.Spec.Volumes = []v1.Volume{{
			Name: "hp",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{Path: "/tmp"},
			},
		}}
		tester(hostpath)
	})

	By("Running a HostNetwork pod", func() {
		hostnet := restrictedPod(f, "hostnet")
		hostnet.Spec.HostNetwork = true
		tester(hostnet)
	})

	By("Running a HostPID pod", func() {
		hostpid := restrictedPod(f, "hostpid")
		hostpid.Spec.HostPID = true
		tester(hostpid)
	})

	By("Running a HostIPC pod", func() {
		hostipc := restrictedPod(f, "hostipc")
		hostipc.Spec.HostIPC = true
		tester(hostipc)
	})

	if common.IsAppArmorSupported() {
		By("Running a custom AppArmor profile pod", func() {
			aa := restrictedPod(f, "apparmor")
			// Every node is expected to have the docker-default profile.
			aa.Annotations[apparmor.ContainerAnnotationKeyPrefix+"pause"] = "localhost/docker-default"
			tester(aa)
		})
	}

	By("Running an unconfined Seccomp pod", func() {
		unconfined := restrictedPod(f, "seccomp")
		unconfined.Annotations[v1.SeccompPodAnnotationKey] = "unconfined"
		tester(unconfined)
	})

	By("Running a SYS_ADMIN pod", func() {
		sysadmin := restrictedPod(f, "sysadmin")
		sysadmin.Spec.Containers[0].SecurityContext.Capabilities = &v1.Capabilities{
			Add: []v1.Capability{"SYS_ADMIN"},
		}
		sysadmin.Spec.Containers[0].SecurityContext.AllowPrivilegeEscalation = nil
		tester(sysadmin)
	})
}

func createAndBindPSP(f *framework.Framework, pspTemplate *extensionsv1beta1.PodSecurityPolicy) (psp *extensionsv1beta1.PodSecurityPolicy, cleanup func()) {
	// Create the PodSecurityPolicy object.
	psp = pspTemplate.DeepCopy()
	// Add the namespace to the name to ensure uniqueness and tie it to the namespace.
	ns := f.Namespace.Name
	name := fmt.Sprintf("%s-%s", ns, psp.Name)
	psp.Name = name
	psp, err := f.ClientSet.ExtensionsV1beta1().PodSecurityPolicies().Create(psp)
	framework.ExpectNoError(err, "Failed to create PSP")

	// Create the Role to bind it to the namespace.
	_, err = f.ClientSet.RbacV1beta1().Roles(ns).Create(&rbacv1beta1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Rules: []rbacv1beta1.PolicyRule{{
			APIGroups:     []string{"extensions"},
			Resources:     []string{"podsecuritypolicies"},
			ResourceNames: []string{name},
			Verbs:         []string{"use"},
		}},
	})
	framework.ExpectNoError(err, "Failed to create PSP role")

	// Bind the role to the namespace.
	framework.BindRoleInNamespace(f.ClientSet.RbacV1beta1(), name, ns, rbacv1beta1.Subject{
		Kind:      rbacv1beta1.ServiceAccountKind,
		Namespace: ns,
		Name:      "default",
	})
	framework.ExpectNoError(framework.WaitForNamedAuthorizationUpdate(f.ClientSet.AuthorizationV1beta1(),
		serviceaccount.MakeUsername(ns, "default"), ns, "use", name,
		schema.GroupResource{Group: "extensions", Resource: "podsecuritypolicies"}, true))

	return psp, func() {
		// Cleanup non-namespaced PSP object.
		f.ClientSet.ExtensionsV1beta1().PodSecurityPolicies().Delete(name, &metav1.DeleteOptions{})
	}
}

// createAndBindPSPInPolicy creates a PSP in the policy API group (unlike createAndBindPSP()).
// TODO: merge these functions when PSP will be completely moved out of the extensions
func createAndBindPSPInPolicy(f *framework.Framework, pspTemplate *policy.PodSecurityPolicy) (psp *policy.PodSecurityPolicy, cleanup func()) {
	// Create the PodSecurityPolicy object.
	psp = pspTemplate.DeepCopy()
	// Add the namespace to the name to ensure uniqueness and tie it to the namespace.
	ns := f.Namespace.Name
	name := fmt.Sprintf("%s-%s", ns, psp.Name)
	psp.Name = name
	psp, err := f.ClientSet.PolicyV1beta1().PodSecurityPolicies().Create(psp)
	framework.ExpectNoError(err, "Failed to create PSP")

	// Create the Role to bind it to the namespace.
	_, err = f.ClientSet.RbacV1beta1().Roles(ns).Create(&rbacv1beta1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Rules: []rbacv1beta1.PolicyRule{{
			APIGroups:     []string{"policy"},
			Resources:     []string{"podsecuritypolicies"},
			ResourceNames: []string{name},
			Verbs:         []string{"use"},
		}},
	})
	framework.ExpectNoError(err, "Failed to create PSP role")

	// Bind the role to the namespace.
	framework.BindRoleInNamespace(f.ClientSet.RbacV1beta1(), name, ns, rbacv1beta1.Subject{
		Kind:      rbacv1beta1.ServiceAccountKind,
		Namespace: ns,
		Name:      "default",
	})
	framework.ExpectNoError(framework.WaitForNamedAuthorizationUpdate(f.ClientSet.AuthorizationV1beta1(),
		serviceaccount.MakeUsername(ns, "default"), ns, "use", name,
		schema.GroupResource{Group: "policy", Resource: "podsecuritypolicies"}, true))

	return psp, func() {
		// Cleanup non-namespaced PSP object.
		f.ClientSet.PolicyV1beta1().PodSecurityPolicies().Delete(name, &metav1.DeleteOptions{})
	}
}

func restrictedPod(f *framework.Framework, name string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				v1.SeccompPodAnnotationKey:                      v1.SeccompProfileRuntimeDefault,
				apparmor.ContainerAnnotationKeyPrefix + "pause": apparmor.ProfileRuntimeDefault,
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{{
				Name:  "pause",
				Image: imageutils.GetPauseImageName(),
				SecurityContext: &v1.SecurityContext{
					AllowPrivilegeEscalation: boolPtr(false),
					RunAsUser:                utilpointer.Int64Ptr(65534),
				},
			}},
		},
	}
}

// privilegedPSPInPolicy creates a PodSecurityPolicy (in the "policy" API Group) that allows everything.
// TODO: replace by PrivilegedPSP when PSP will be completely moved out of the extensions
func privilegedPSPInPolicy(name string) *policy.PodSecurityPolicy {
	return &policy.PodSecurityPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: map[string]string{seccomp.AllowedProfilesAnnotationKey: seccomp.AllowAny},
		},
		Spec: policy.PodSecurityPolicySpec{
			Privileged:               true,
			AllowPrivilegeEscalation: utilpointer.BoolPtr(true),
			AllowedCapabilities:      []v1.Capability{"*"},
			Volumes:                  []policy.FSType{policy.All},
			HostNetwork:              true,
			HostPorts:                []policy.HostPortRange{{Min: 0, Max: 65535}},
			HostIPC:                  true,
			HostPID:                  true,
			RunAsUser: policy.RunAsUserStrategyOptions{
				Rule: policy.RunAsUserStrategyRunAsAny,
			},
			SELinux: policy.SELinuxStrategyOptions{
				Rule: policy.SELinuxStrategyRunAsAny,
			},
			SupplementalGroups: policy.SupplementalGroupsStrategyOptions{
				Rule: policy.SupplementalGroupsStrategyRunAsAny,
			},
			FSGroup: policy.FSGroupStrategyOptions{
				Rule: policy.FSGroupStrategyRunAsAny,
			},
			ReadOnlyRootFilesystem: false,
		},
	}
}

// restrictedPSPInPolicy creates a PodSecurityPolicy (in the "policy" API Group) that is most strict.
// TODO: replace by restrictedPSP when PSP will be completely moved out of the extensions
func restrictedPSPInPolicy(name string) *policy.PodSecurityPolicy {
	return &policy.PodSecurityPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				seccomp.AllowedProfilesAnnotationKey:  v1.SeccompProfileRuntimeDefault,
				seccomp.DefaultProfileAnnotationKey:   v1.SeccompProfileRuntimeDefault,
				apparmor.AllowedProfilesAnnotationKey: apparmor.ProfileRuntimeDefault,
				apparmor.DefaultProfileAnnotationKey:  apparmor.ProfileRuntimeDefault,
			},
		},
		Spec: policy.PodSecurityPolicySpec{
			Privileged:               false,
			AllowPrivilegeEscalation: utilpointer.BoolPtr(false),
			RequiredDropCapabilities: []v1.Capability{
				"AUDIT_WRITE",
				"CHOWN",
				"DAC_OVERRIDE",
				"FOWNER",
				"FSETID",
				"KILL",
				"MKNOD",
				"NET_RAW",
				"SETGID",
				"SETUID",
				"SYS_CHROOT",
			},
			Volumes: []policy.FSType{
				policy.ConfigMap,
				policy.EmptyDir,
				policy.PersistentVolumeClaim,
				"projected",
				policy.Secret,
			},
			HostNetwork: false,
			HostIPC:     false,
			HostPID:     false,
			RunAsUser: policy.RunAsUserStrategyOptions{
				Rule: policy.RunAsUserStrategyMustRunAsNonRoot,
			},
			SELinux: policy.SELinuxStrategyOptions{
				Rule: policy.SELinuxStrategyRunAsAny,
			},
			SupplementalGroups: policy.SupplementalGroupsStrategyOptions{
				Rule: policy.SupplementalGroupsStrategyRunAsAny,
			},
			FSGroup: policy.FSGroupStrategyOptions{
				Rule: policy.FSGroupStrategyRunAsAny,
			},
			ReadOnlyRootFilesystem: false,
		},
	}
}

// restrictedPSP creates a PodSecurityPolicy that is most strict.
func restrictedPSP(name string) *extensionsv1beta1.PodSecurityPolicy {
	return &extensionsv1beta1.PodSecurityPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				seccomp.AllowedProfilesAnnotationKey:  v1.SeccompProfileRuntimeDefault,
				seccomp.DefaultProfileAnnotationKey:   v1.SeccompProfileRuntimeDefault,
				apparmor.AllowedProfilesAnnotationKey: apparmor.ProfileRuntimeDefault,
				apparmor.DefaultProfileAnnotationKey:  apparmor.ProfileRuntimeDefault,
			},
		},
		Spec: extensionsv1beta1.PodSecurityPolicySpec{
			Privileged:               false,
			AllowPrivilegeEscalation: utilpointer.BoolPtr(false),
			RequiredDropCapabilities: []v1.Capability{
				"AUDIT_WRITE",
				"CHOWN",
				"DAC_OVERRIDE",
				"FOWNER",
				"FSETID",
				"KILL",
				"MKNOD",
				"NET_RAW",
				"SETGID",
				"SETUID",
				"SYS_CHROOT",
			},
			Volumes: []extensionsv1beta1.FSType{
				extensionsv1beta1.ConfigMap,
				extensionsv1beta1.EmptyDir,
				extensionsv1beta1.PersistentVolumeClaim,
				"projected",
				extensionsv1beta1.Secret,
			},
			HostNetwork: false,
			HostIPC:     false,
			HostPID:     false,
			RunAsUser: extensionsv1beta1.RunAsUserStrategyOptions{
				Rule: extensionsv1beta1.RunAsUserStrategyMustRunAsNonRoot,
			},
			SELinux: extensionsv1beta1.SELinuxStrategyOptions{
				Rule: extensionsv1beta1.SELinuxStrategyRunAsAny,
			},
			SupplementalGroups: extensionsv1beta1.SupplementalGroupsStrategyOptions{
				Rule: extensionsv1beta1.SupplementalGroupsStrategyRunAsAny,
			},
			FSGroup: extensionsv1beta1.FSGroupStrategyOptions{
				Rule: extensionsv1beta1.FSGroupStrategyRunAsAny,
			},
			ReadOnlyRootFilesystem: false,
		},
	}
}

func boolPtr(b bool) *bool {
	return &b
}
