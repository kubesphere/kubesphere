package util

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"k8s.io/klog"
	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"
)

func GeneratePodName(repoName string) string {
	return fmt.Sprintf("%s-%s", "catalog", repoName)
}

func GetRecommendedPluginVersion(versions []extensionsv1alpha1.PluginVersion, k8sVersion string) *extensionsv1alpha1.PluginVersion {
	if len(versions) == 0 {
		return nil
	}

	kubeSemver, err := semver.NewVersion(k8sVersion)
	if err != nil {
		klog.V(2).Infof("parse kubernetes version failed, err: %s", err)
		return nil
	}

	var latestVersion *extensionsv1alpha1.PluginVersion
	var latestSemver *semver.Version

	for i := range versions {
		currSemver, err := semver.NewVersion(versions[i].Spec.MinKubeVersion)
		if err == nil {
			if latestSemver == nil {
				// the first valid semver
				latestSemver = currSemver
			} else if latestSemver.LessThan(currSemver) {
				// find a newer valid semver
				latestSemver = currSemver
			}

			if latestSemver.LessThan(kubeSemver) {
				latestVersion = &versions[i]
			}
		} else {
			// If the semver is invalid, just ignore it.
			klog.V(2).Infof("parse version failed, plugin version: %s, err: %s", versions[i].Name, err)
		}
	}

	return latestVersion
}

func GetLatestPluginVersion(versions []extensionsv1alpha1.PluginVersion) *extensionsv1alpha1.PluginVersion {
	if len(versions) == 0 {
		return nil
	}

	var latestVersion *extensionsv1alpha1.PluginVersion
	var latestSemver *semver.Version

	for i := range versions {
		currSemver, err := semver.NewVersion(versions[i].Spec.Version)
		if err == nil {
			if latestSemver == nil {
				// the first valid semver
				latestSemver = currSemver
				latestVersion = &versions[i]
			} else if latestSemver.LessThan(currSemver) {
				// find a newer valid semver
				latestSemver = currSemver
				latestVersion = &versions[i]
			}
		} else {
			// If the semver is invalid, just ignore it.
			klog.V(2).Infof("parse version failed, plugin version: %s, err: %s", versions[i].Name, err)
		}
	}
	return latestVersion
}

type PluginVersionList []extensionsv1alpha1.PluginVersionInfo

func (pvl PluginVersionList) Len() int           { return len(pvl) }
func (pvl PluginVersionList) Less(i, j int) bool { return pvl[i].Version < pvl[j].Version }
func (pvl PluginVersionList) Swap(i, j int)      { pvl[i], pvl[j] = pvl[j], pvl[i] }
