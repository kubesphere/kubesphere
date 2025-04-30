/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package core

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
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

func TestFilterExtensionVersions(t *testing.T) {
	var getExtensionVersion = func(version string) corev1alpha1.ExtensionVersion {
		return corev1alpha1.ExtensionVersion{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-" + version,
			},
			Spec: corev1alpha1.ExtensionVersionSpec{
				Version: version,
			},
		}
	}

	tests := []struct {
		name           string
		versions       []corev1alpha1.ExtensionVersion
		depth          *int
		exceptVersions []corev1alpha1.ExtensionVersion
	}{
		{
			name: "invalid versions",
			versions: []corev1alpha1.ExtensionVersion{
				getExtensionVersion("abc"),
				getExtensionVersion("v1.1.1"),
				getExtensionVersion("v1.1.2"),
			},
			exceptVersions: []corev1alpha1.ExtensionVersion{
				getExtensionVersion("v1.1.2"),
				getExtensionVersion("v1.1.1"),
			},
		},
		{
			name: "depth is null", // default value
			versions: []corev1alpha1.ExtensionVersion{
				getExtensionVersion("v1.1.1"),
				getExtensionVersion("v1.1.2"),
				getExtensionVersion("v1.1.3"),
				getExtensionVersion("v1.1.4"),
			},
			exceptVersions: []corev1alpha1.ExtensionVersion{
				getExtensionVersion("v1.1.4"),
				getExtensionVersion("v1.1.3"),
				getExtensionVersion("v1.1.2"),
			},
		},
		{
			name: "depth is 0", // all value
			versions: []corev1alpha1.ExtensionVersion{
				getExtensionVersion("v1.1.1"),
				getExtensionVersion("v1.1.2"),
				getExtensionVersion("v1.1.3"),
				getExtensionVersion("v1.1.4"),
			},
			depth: ptr.To(0),
			exceptVersions: []corev1alpha1.ExtensionVersion{
				getExtensionVersion("v1.1.4"),
				getExtensionVersion("v1.1.3"),
				getExtensionVersion("v1.1.2"),
				getExtensionVersion("v1.1.1"),
			},
		},
		{
			name: "depth over length range", // all value
			versions: []corev1alpha1.ExtensionVersion{
				getExtensionVersion("v1.1.1"),
				getExtensionVersion("v1.1.2"),
				getExtensionVersion("v1.1.3"),
				getExtensionVersion("v1.1.4"),
			},
			depth: ptr.To(10),
			exceptVersions: []corev1alpha1.ExtensionVersion{
				getExtensionVersion("v1.1.4"),
				getExtensionVersion("v1.1.3"),
				getExtensionVersion("v1.1.2"),
				getExtensionVersion("v1.1.1"),
			},
		},
		{
			name: "depth in length range", // specific value
			versions: []corev1alpha1.ExtensionVersion{
				getExtensionVersion("v1.1.1"),
				getExtensionVersion("v1.1.2"),
				getExtensionVersion("v1.1.3"),
				getExtensionVersion("v1.1.4"),
			},
			depth: ptr.To(2),
			exceptVersions: []corev1alpha1.ExtensionVersion{
				getExtensionVersion("v1.1.4"),
				getExtensionVersion("v1.1.3"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			versions := filterExtensionVersions(tt.versions, tt.depth)
			assert.Equal(t, versions, tt.exceptVersions)
		})
	}
}
