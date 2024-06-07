/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package core

import (
	"testing"

	"github.com/Masterminds/semver/v3"

	corev1alpha1 "kubesphere.io/api/core/v1alpha1"

	"kubesphere.io/kubesphere/pkg/version"
)

func TestGetRecommendedExtensionVersion(t *testing.T) {
	k8sVersion120, _ := semver.NewVersion("1.20.0")
	k8sVersion125, _ := semver.NewVersion("1.25.4")
	tests := []struct {
		name       string
		versions   []corev1alpha1.ExtensionVersion
		k8sVersion *semver.Version
		ksVersion  string
		wanted     string
	}{
		{
			name: "normal test",
			versions: []corev1alpha1.ExtensionVersion{
				{
					Spec: corev1alpha1.ExtensionVersionSpec{ // match
						Version:     "1.0.0",
						KubeVersion: ">=1.19.0",
						KSVersion:   ">=4.0.0",
					},
				},
				{
					Spec: corev1alpha1.ExtensionVersionSpec{ // match
						Version:     "1.1.0",
						KubeVersion: ">=1.20.0",
						KSVersion:   ">=4.0.0",
					},
				},
				{
					Spec: corev1alpha1.ExtensionVersionSpec{ // KubeVersion not match
						Version:     "1.2.0",
						KubeVersion: ">=1.21.0",
						KSVersion:   ">=4.0.0",
					},
				},
				{
					Spec: corev1alpha1.ExtensionVersionSpec{ // KSVersion not match
						Version:     "1.3.0",
						KubeVersion: ">=1.20.0",
						KSVersion:   ">=4.1.0",
					},
				},
			},
			k8sVersion: k8sVersion120,
			ksVersion:  "4.0.0",
			wanted:     "1.1.0",
		},
		{
			name: "no matches test",
			versions: []corev1alpha1.ExtensionVersion{
				{
					Spec: corev1alpha1.ExtensionVersionSpec{ // KubeVersion not match
						Version:     "1.2.0",
						KubeVersion: ">=1.21.0",
						KSVersion:   ">=4.0.0",
					},
				},
				{
					Spec: corev1alpha1.ExtensionVersionSpec{ // KSVersion not match
						Version:     "1.3.0",
						KubeVersion: ">=1.20.0",
						KSVersion:   ">=4.1.0",
					},
				},
			},
			k8sVersion: k8sVersion120,
			ksVersion:  "4.0.0",
			wanted:     "",
		},
		{
			name: "match 1.3.0",
			versions: []corev1alpha1.ExtensionVersion{
				{
					Spec: corev1alpha1.ExtensionVersionSpec{
						Version:     "1.2.0",
						KubeVersion: ">=1.19.0",
						KSVersion:   ">=3.0.0",
					},
				},
				{
					Spec: corev1alpha1.ExtensionVersionSpec{
						Version:     "1.3.0",
						KubeVersion: ">=1.19.0",
						KSVersion:   ">=4.0.0-alpha",
					},
				},
			},
			k8sVersion: k8sVersion125,
			ksVersion:  "4.0.0-beta.5+ae34",
			wanted:     "1.3.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version.SetGitVersion(tt.ksVersion)
			if got, _ := getRecommendedExtensionVersion(tt.versions, tt.k8sVersion); got != tt.wanted {
				t.Errorf("getRecommendedExtensionVersion() = %v, want %v", got, tt.wanted)
			}
		})
	}
}
