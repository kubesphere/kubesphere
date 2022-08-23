package core

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"k8s.io/klog"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
)

func generatePodName(repoName string) string {
	return fmt.Sprintf("%s-%s", "catalog", repoName)
}

func getRecommendedExtensionVersion(versions []corev1alpha1.ExtensionVersion, k8sVersion string) *corev1alpha1.ExtensionVersion {
	if len(versions) == 0 {
		return nil
	}

	kubeSemver, err := semver.NewVersion(k8sVersion)
	if err != nil {
		klog.V(2).Infof("parse kubernetes version failed, err: %s", err)
		return nil
	}

	var latestVersion *corev1alpha1.ExtensionVersion
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
			klog.V(2).Infof("parse version failed, extension version: %s, err: %s", versions[i].Name, err)
		}
	}

	return latestVersion
}

func getLatestExtensionVersion(versions []corev1alpha1.ExtensionVersion) *corev1alpha1.ExtensionVersion {
	if len(versions) == 0 {
		return nil
	}

	var latestVersion *corev1alpha1.ExtensionVersion
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
			klog.V(2).Infof("parse version failed, extension version: %s, err: %s", versions[i].Name, err)
		}
	}
	return latestVersion
}

type VersionList []corev1alpha1.ExtensionVersionInfo

func (pvl VersionList) Len() int           { return len(pvl) }
func (pvl VersionList) Less(i, j int) bool { return pvl[i].Version < pvl[j].Version }
func (pvl VersionList) Swap(i, j int)      { pvl[i], pvl[j] = pvl[j], pvl[i] }
