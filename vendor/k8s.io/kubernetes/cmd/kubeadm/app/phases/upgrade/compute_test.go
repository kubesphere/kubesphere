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

package upgrade

import (
	"reflect"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	versionutil "k8s.io/kubernetes/pkg/util/version"
)

type fakeVersionGetter struct {
	clusterVersion, kubeadmVersion, stableVersion, latestVersion, latestDevBranchVersion, stablePatchVersion, kubeletVersion string
}

var _ VersionGetter = &fakeVersionGetter{}

// ClusterVersion gets a fake API server version
func (f *fakeVersionGetter) ClusterVersion() (string, *versionutil.Version, error) {
	return f.clusterVersion, versionutil.MustParseSemantic(f.clusterVersion), nil
}

// KubeadmVersion gets a fake kubeadm version
func (f *fakeVersionGetter) KubeadmVersion() (string, *versionutil.Version, error) {
	return f.kubeadmVersion, versionutil.MustParseSemantic(f.kubeadmVersion), nil
}

// VersionFromCILabel gets fake latest versions from CI
func (f *fakeVersionGetter) VersionFromCILabel(ciVersionLabel, _ string) (string, *versionutil.Version, error) {
	if ciVersionLabel == "stable" {
		return f.stableVersion, versionutil.MustParseSemantic(f.stableVersion), nil
	}
	if ciVersionLabel == "latest" {
		return f.latestVersion, versionutil.MustParseSemantic(f.latestVersion), nil
	}
	if ciVersionLabel == "latest-1.10" {
		return f.latestDevBranchVersion, versionutil.MustParseSemantic(f.latestDevBranchVersion), nil
	}
	return f.stablePatchVersion, versionutil.MustParseSemantic(f.stablePatchVersion), nil
}

// KubeletVersions gets the versions of the kubelets in the cluster
func (f *fakeVersionGetter) KubeletVersions() (map[string]uint16, error) {
	return map[string]uint16{
		f.kubeletVersion: 1,
	}, nil
}

type fakeEtcdCluster struct{ TLS bool }

func (f fakeEtcdCluster) HasTLS() bool { return f.TLS }

func (f fakeEtcdCluster) GetStatus() (*clientv3.StatusResponse, error) {
	client := &clientv3.StatusResponse{}
	client.Version = "3.1.12"
	return client, nil
}

func (f fakeEtcdCluster) WaitForStatus(delay time.Duration, retries int, retryInterval time.Duration) (*clientv3.StatusResponse, error) {
	return f.GetStatus()
}

func TestGetAvailableUpgrades(t *testing.T) {
	featureGates := make(map[string]bool)
	tests := []struct {
		name                        string
		vg                          VersionGetter
		expectedUpgrades            []Upgrade
		allowExperimental, allowRCs bool
		errExpected                 bool
	}{
		{
			name: "no action needed, already up-to-date",
			vg: &fakeVersionGetter{
				clusterVersion: "v1.9.3",
				kubeletVersion: "v1.9.3",
				kubeadmVersion: "v1.9.3",

				stablePatchVersion: "v1.9.3",
				stableVersion:      "v1.9.3",
			},
			expectedUpgrades:  []Upgrade{},
			allowExperimental: false,
			errExpected:       false,
		},
		{
			name: "simple patch version upgrade",
			vg: &fakeVersionGetter{
				clusterVersion: "v1.9.1",
				kubeletVersion: "v1.9.1", // the kubelet are on the same version as the control plane
				kubeadmVersion: "v1.9.2",

				stablePatchVersion: "v1.9.3",
				stableVersion:      "v1.9.3",
			},
			expectedUpgrades: []Upgrade{
				{
					Description: "version in the v1.9 series",
					Before: ClusterState{
						KubeVersion: "v1.9.1",
						KubeletVersions: map[string]uint16{
							"v1.9.1": 1,
						},
						KubeadmVersion: "v1.9.2",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.1.12",
					},
					After: ClusterState{
						KubeVersion:    "v1.9.3",
						KubeadmVersion: "v1.9.3",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.1.12",
					},
				},
			},
			allowExperimental: false,
			errExpected:       false,
		},
		{
			name: "no version provided to offline version getter does not change behavior",
			vg: NewOfflineVersionGetter(&fakeVersionGetter{
				clusterVersion: "v1.9.1",
				kubeletVersion: "v1.9.1", // the kubelet are on the same version as the control plane
				kubeadmVersion: "v1.9.2",

				stablePatchVersion: "v1.9.3",
				stableVersion:      "v1.9.3",
			}, ""),
			expectedUpgrades: []Upgrade{
				{
					Description: "version in the v1.9 series",
					Before: ClusterState{
						KubeVersion: "v1.9.1",
						KubeletVersions: map[string]uint16{
							"v1.9.1": 1,
						},
						KubeadmVersion: "v1.9.2",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.1.12",
					},
					After: ClusterState{
						KubeVersion:    "v1.9.3",
						KubeadmVersion: "v1.9.3",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.1.12",
					},
				},
			},
			allowExperimental: false,
			errExpected:       false,
		},
		{
			name: "minor version upgrade only",
			vg: &fakeVersionGetter{
				clusterVersion: "v1.9.1",
				kubeletVersion: "v1.9.1", // the kubelet are on the same version as the control plane
				kubeadmVersion: "v1.10.0",

				stablePatchVersion: "v1.9.1",
				stableVersion:      "v1.10.0",
			},
			expectedUpgrades: []Upgrade{
				{
					Description: "stable version",
					Before: ClusterState{
						KubeVersion: "v1.9.1",
						KubeletVersions: map[string]uint16{
							"v1.9.1": 1,
						},
						KubeadmVersion: "v1.10.0",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.1.12",
					},
					After: ClusterState{
						KubeVersion:    "v1.10.0",
						KubeadmVersion: "v1.10.0",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.1.12",
					},
				},
			},
			allowExperimental: false,
			errExpected:       false,
		},
		{
			name: "both minor version upgrade and patch version upgrade available",
			vg: &fakeVersionGetter{
				clusterVersion: "v1.9.3",
				kubeletVersion: "v1.9.3", // the kubelet are on the same version as the control plane
				kubeadmVersion: "v1.9.5",

				stablePatchVersion: "v1.9.5",
				stableVersion:      "v1.10.1",
			},
			expectedUpgrades: []Upgrade{
				{
					Description: "version in the v1.9 series",
					Before: ClusterState{
						KubeVersion: "v1.9.3",
						KubeletVersions: map[string]uint16{
							"v1.9.3": 1,
						},
						KubeadmVersion: "v1.9.5",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.1.12",
					},
					After: ClusterState{
						KubeVersion:    "v1.9.5",
						KubeadmVersion: "v1.9.5", // Note: The kubeadm version mustn't be "downgraded" here
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.1.12",
					},
				},
				{
					Description: "stable version",
					Before: ClusterState{
						KubeVersion: "v1.9.3",
						KubeletVersions: map[string]uint16{
							"v1.9.3": 1,
						},
						KubeadmVersion: "v1.9.5",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.1.12",
					},
					After: ClusterState{
						KubeVersion:    "v1.10.1",
						KubeadmVersion: "v1.10.1",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.1.12",
					},
				},
			},
			allowExperimental: false,
			errExpected:       false,
		},
		{
			name: "allow experimental upgrades, but no upgrade available",
			vg: &fakeVersionGetter{
				clusterVersion: "v1.10.0-alpha.2",
				kubeletVersion: "v1.9.5",
				kubeadmVersion: "v1.9.5",

				stablePatchVersion: "v1.9.5",
				stableVersion:      "v1.9.5",
				latestVersion:      "v1.10.0-alpha.2",
			},
			expectedUpgrades:  []Upgrade{},
			allowExperimental: true,
			errExpected:       false,
		},
		{
			name: "upgrade to an unstable version should be supported",
			vg: &fakeVersionGetter{
				clusterVersion: "v1.9.5",
				kubeletVersion: "v1.9.5",
				kubeadmVersion: "v1.9.5",

				stablePatchVersion: "v1.9.5",
				stableVersion:      "v1.9.5",
				latestVersion:      "v1.10.0-alpha.2",
			},
			expectedUpgrades: []Upgrade{
				{
					Description: "experimental version",
					Before: ClusterState{
						KubeVersion: "v1.9.5",
						KubeletVersions: map[string]uint16{
							"v1.9.5": 1,
						},
						KubeadmVersion: "v1.9.5",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.1.12",
					},
					After: ClusterState{
						KubeVersion:    "v1.10.0-alpha.2",
						KubeadmVersion: "v1.10.0-alpha.2",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.1.12",
					},
				},
			},
			allowExperimental: true,
			errExpected:       false,
		},
		{
			name: "upgrade from an unstable version to an unstable version should be supported",
			vg: &fakeVersionGetter{
				clusterVersion: "v1.10.0-alpha.1",
				kubeletVersion: "v1.9.5",
				kubeadmVersion: "v1.9.5",

				stablePatchVersion: "v1.9.5",
				stableVersion:      "v1.9.5",
				latestVersion:      "v1.10.0-alpha.2",
			},
			expectedUpgrades: []Upgrade{
				{
					Description: "experimental version",
					Before: ClusterState{
						KubeVersion: "v1.10.0-alpha.1",
						KubeletVersions: map[string]uint16{
							"v1.9.5": 1,
						},
						KubeadmVersion: "v1.9.5",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.1.12",
					},
					After: ClusterState{
						KubeVersion:    "v1.10.0-alpha.2",
						KubeadmVersion: "v1.10.0-alpha.2",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.1.12",
					},
				},
			},
			allowExperimental: true,
			errExpected:       false,
		},
		{
			name: "v1.X.0-alpha.0 should be ignored",
			vg: &fakeVersionGetter{
				clusterVersion: "v1.9.5",
				kubeletVersion: "v1.9.5",
				kubeadmVersion: "v1.9.5",

				stablePatchVersion:     "v1.9.5",
				stableVersion:          "v1.9.5",
				latestDevBranchVersion: "v1.10.0-beta.1",
				latestVersion:          "v1.11.0-alpha.0",
			},
			expectedUpgrades: []Upgrade{
				{
					Description: "experimental version",
					Before: ClusterState{
						KubeVersion: "v1.9.5",
						KubeletVersions: map[string]uint16{
							"v1.9.5": 1,
						},
						KubeadmVersion: "v1.9.5",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.1.12",
					},
					After: ClusterState{
						KubeVersion:    "v1.10.0-beta.1",
						KubeadmVersion: "v1.10.0-beta.1",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.1.12",
					},
				},
			},
			allowExperimental: true,
			errExpected:       false,
		},
		{
			name: "upgrade to an RC version should be supported",
			vg: &fakeVersionGetter{
				clusterVersion: "v1.9.5",
				kubeletVersion: "v1.9.5",
				kubeadmVersion: "v1.9.5",

				stablePatchVersion:     "v1.9.5",
				stableVersion:          "v1.9.5",
				latestDevBranchVersion: "v1.10.0-rc.1",
				latestVersion:          "v1.11.0-alpha.1",
			},
			expectedUpgrades: []Upgrade{
				{
					Description: "release candidate version",
					Before: ClusterState{
						KubeVersion: "v1.9.5",
						KubeletVersions: map[string]uint16{
							"v1.9.5": 1,
						},
						KubeadmVersion: "v1.9.5",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.1.12",
					},
					After: ClusterState{
						KubeVersion:    "v1.10.0-rc.1",
						KubeadmVersion: "v1.10.0-rc.1",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.1.12",
					},
				},
			},
			allowRCs:    true,
			errExpected: false,
		},
		{
			name: "it is possible (but very uncommon) that the latest version from the previous branch is an rc and the current latest version is alpha.0. In that case, show the RC",
			vg: &fakeVersionGetter{
				clusterVersion: "v1.9.5",
				kubeletVersion: "v1.9.5",
				kubeadmVersion: "v1.9.5",

				stablePatchVersion:     "v1.9.5",
				stableVersion:          "v1.9.5",
				latestDevBranchVersion: "v1.10.6-rc.1",
				latestVersion:          "v1.11.1-alpha.0",
			},
			expectedUpgrades: []Upgrade{
				{
					Description: "experimental version", // Note that this is considered an experimental version in this uncommon scenario
					Before: ClusterState{
						KubeVersion: "v1.9.5",
						KubeletVersions: map[string]uint16{
							"v1.9.5": 1,
						},
						KubeadmVersion: "v1.9.5",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.1.12",
					},
					After: ClusterState{
						KubeVersion:    "v1.10.6-rc.1",
						KubeadmVersion: "v1.10.6-rc.1",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.1.12",
					},
				},
			},
			allowExperimental: true,
			errExpected:       false,
		},
		{
			name: "upgrade to an RC version should be supported. There may also be an even newer unstable version.",
			vg: &fakeVersionGetter{
				clusterVersion: "v1.9.5",
				kubeletVersion: "v1.9.5",
				kubeadmVersion: "v1.9.5",

				stablePatchVersion:     "v1.9.5",
				stableVersion:          "v1.9.5",
				latestDevBranchVersion: "v1.10.0-rc.1",
				latestVersion:          "v1.11.0-alpha.2",
			},
			expectedUpgrades: []Upgrade{
				{
					Description: "release candidate version",
					Before: ClusterState{
						KubeVersion: "v1.9.5",
						KubeletVersions: map[string]uint16{
							"v1.9.5": 1,
						},
						KubeadmVersion: "v1.9.5",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.1.12",
					},
					After: ClusterState{
						KubeVersion:    "v1.10.0-rc.1",
						KubeadmVersion: "v1.10.0-rc.1",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.1.12",
					},
				},
				{
					Description: "experimental version",
					Before: ClusterState{
						KubeVersion: "v1.9.5",
						KubeletVersions: map[string]uint16{
							"v1.9.5": 1,
						},
						KubeadmVersion: "v1.9.5",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.1.12",
					},
					After: ClusterState{
						KubeVersion:    "v1.11.0-alpha.2",
						KubeadmVersion: "v1.11.0-alpha.2",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.2.18",
					},
				},
			},
			allowRCs:          true,
			allowExperimental: true,
			errExpected:       false,
		},
		{
			name: "offline version getter",
			vg: NewOfflineVersionGetter(&fakeVersionGetter{
				clusterVersion: "v1.10.1",
				kubeletVersion: "v1.10.0",
				kubeadmVersion: "v1.10.1",
			}, "v1.11.1"),
			expectedUpgrades: []Upgrade{
				{
					Description: "version in the v1.1 series",
					Before: ClusterState{
						KubeVersion: "v1.10.1",
						KubeletVersions: map[string]uint16{
							"v1.10.0": 1,
						},
						KubeadmVersion: "v1.10.1",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.1.12",
					},
					After: ClusterState{
						KubeVersion:    "v1.11.1",
						KubeadmVersion: "v1.11.1",
						DNSVersion:     "1.14.10",
						EtcdVersion:    "3.2.18",
					},
				},
			},
		},
	}

	// Instantiating a fake etcd cluster for being able to get etcd version for a corresponding
	// kubernetes release.
	testCluster := fakeEtcdCluster{}
	for _, rt := range tests {
		t.Run(rt.name, func(t *testing.T) {
			actualUpgrades, actualErr := GetAvailableUpgrades(rt.vg, rt.allowExperimental, rt.allowRCs, testCluster, featureGates)
			if !reflect.DeepEqual(actualUpgrades, rt.expectedUpgrades) {
				t.Errorf("failed TestGetAvailableUpgrades\n\texpected upgrades: %v\n\tgot: %v", rt.expectedUpgrades, actualUpgrades)
			}
			if (actualErr != nil) != rt.errExpected {
				t.Errorf("failed TestGetAvailableUpgrades\n\texpected error: %t\n\tgot error: %t", rt.errExpected, (actualErr != nil))
			}
		})
	}
}

func TestKubeletUpgrade(t *testing.T) {
	tests := []struct {
		before   map[string]uint16
		after    string
		expected bool
	}{
		{ // upgrade available
			before: map[string]uint16{
				"v1.9.1": 1,
			},
			after:    "v1.9.3",
			expected: true,
		},
		{ // upgrade available
			before: map[string]uint16{
				"v1.9.1": 1,
				"v1.9.3": 100,
			},
			after:    "v1.9.3",
			expected: true,
		},
		{ // upgrade not available
			before: map[string]uint16{
				"v1.9.3": 1,
			},
			after:    "v1.9.3",
			expected: false,
		},
		{ // upgrade not available
			before: map[string]uint16{
				"v1.9.3": 100,
			},
			after:    "v1.9.3",
			expected: false,
		},
		{ // upgrade not available if we don't know anything about the earlier state
			before:   map[string]uint16{},
			after:    "v1.9.3",
			expected: false,
		},
	}

	for _, rt := range tests {

		upgrade := Upgrade{
			Before: ClusterState{
				KubeletVersions: rt.before,
			},
			After: ClusterState{
				KubeVersion: rt.after,
			},
		}
		actual := upgrade.CanUpgradeKubelets()
		if actual != rt.expected {
			t.Errorf("failed TestKubeletUpgrade\n\texpected: %t\n\tgot: %t\n\ttest object: %v", rt.expected, actual, upgrade)
		}
	}
}
