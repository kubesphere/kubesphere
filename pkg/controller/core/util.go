/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package core

import (
	"bytes"
	"encoding/base64"
	goerrors "errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/Masterminds/semver/v3"
	yaml3 "gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/storage/driver"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog/v2"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
	"kubesphere.io/utils/helm"

	"kubesphere.io/kubesphere/pkg/utils/hashutil"
	"kubesphere.io/kubesphere/pkg/version"
)

func getRecommendedExtensionVersion(versions []corev1alpha1.ExtensionVersion, k8sVersion *semver.Version) (string, error) {
	if len(versions) == 0 {
		return "", nil
	}

	ksVersion, err := semver.NewVersion(version.Get().GitVersion)
	if err != nil {
		return "", fmt.Errorf("parse KubeSphere version failed: %v", err)
	}

	var matchedVersions []*semver.Version

	for _, v := range versions {
		var kubeVersionMatched, ksVersionMatched bool

		if v.Spec.KubeVersion == "" {
			kubeVersionMatched = true
		} else {
			targetKubeVersion, err := semver.NewConstraint(v.Spec.KubeVersion)
			if err != nil {
				// If the semver is invalid, just ignore it.
				klog.Warningf("failed to parse Kubernetes version constraints: kubeVersion: %s, err: %s", v.Spec.KubeVersion, err)
				continue
			}
			kubeVersionMatched = targetKubeVersion.Check(k8sVersion)
		}

		if v.Spec.KSVersion == "" {
			ksVersionMatched = true
		} else {
			targetKSVersion, err := semver.NewConstraint(v.Spec.KSVersion)
			if err != nil {
				klog.Warningf("failed to parse KubeSphere version constraints: ksVersion: %s, err: %s", v.Spec.KSVersion, err)
				continue
			}
			ksVersionMatched = targetKSVersion.Check(ksVersion)
		}

		if kubeVersionMatched && ksVersionMatched {
			targetVersion, err := semver.NewVersion(v.Spec.Version)
			if err != nil {
				klog.V(2).Infof("parse version failed, extension version: %s, err: %s", v.Spec.Version, err)
				continue
			}
			matchedVersions = append(matchedVersions, targetVersion)
		}
	}

	if len(matchedVersions) == 0 {
		return "", nil
	}

	sort.Slice(matchedVersions, func(i, j int) bool {
		return matchedVersions[i].Compare(matchedVersions[j]) >= 0
	})

	return matchedVersions[0].Original(), nil
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
			klog.Warningf("parse version failed, extension version: %s, err: %s", versions[i].Name, err)
		}
	}
	return latestVersion
}

func isReleaseNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), driver.ErrReleaseNotFound.Error())
}

func clusterConfig(sub *corev1alpha1.InstallPlan, clusterName string) []byte {
	if clusterName == "" {
		return []byte(sub.Spec.Config)
	}
	for cluster, config := range sub.Spec.ClusterScheduling.Overrides {
		if cluster == clusterName {
			return merge(sub.Spec.Config, config)
		}
	}
	return []byte(sub.Spec.Config)
}

func merge(config string, override string) []byte {
	config = strings.TrimSpace(config)
	override = strings.TrimSpace(override)

	if config == "" && override == "" {
		return []byte("")
	}

	if override == "" {
		return []byte(config)
	}

	if config == "" {
		return []byte(override)
	}

	baseConf := map[string]interface{}{}
	if err := yaml3.Unmarshal([]byte(config), &baseConf); err != nil {
		klog.Warningf("failed to unmarshal config: %v", err)
	}

	overrideConf := map[string]interface{}{}
	if err := yaml3.Unmarshal([]byte(override), overrideConf); err != nil {
		klog.Warningf("failed to unmarshal config: %v", err)
	}

	finalConf := mergeValues(baseConf, overrideConf)
	data, _ := yaml3.Marshal(finalConf)
	return data
}

// mergeValues will merge source and destination map, preferring values from the source map
func mergeValues(dest map[string]interface{}, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}
		nextMap, ok := v.(map[string]interface{})
		// If it isn't another map, overwrite the value
		if !ok {
			dest[k] = v
			continue
		}
		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := dest[k].(map[string]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		dest[k] = mergeValues(destMap, nextMap)
	}
	return dest
}

func usesPermissions(mainChart *chart.Chart) (rbacv1.ClusterRole, rbacv1.Role) {
	var clusterRole rbacv1.ClusterRole
	var role rbacv1.Role
	for _, file := range mainChart.Files {
		if file.Name == permissionDefinitionFile {
			// decoder := yaml.NewDecoder(bytes.NewReader(file.Data))
			decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(file.Data), 1024)
			for {
				result := new(rbacv1.Role)
				// create new spec here
				// pass a reference to spec reference
				err := decoder.Decode(&result)
				// check it was parsed
				if result == nil {
					continue
				}
				// break the loop in case of EOF
				if goerrors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					return clusterRole, role
				}
				if result.Kind == "ClusterRole" {
					clusterRole.Rules = append(clusterRole.Rules, result.Rules...)
				}
				if result.Kind == "Role" {
					role.Rules = append(role.Rules, result.Rules...)
				}
			}
		}
	}
	return clusterRole, role
}

func hasCluster(clusters []clusterv1alpha1.Cluster, clusterName string) bool {
	for _, cluster := range clusters {
		if cluster.Name == clusterName {
			return true
		}
	}
	return false
}

func versionChanged(plan *corev1alpha1.InstallPlan, cluster string) bool {
	var oldVersion string
	if cluster == "" {
		oldVersion = plan.Status.Version
	} else if plan.Status.ClusterSchedulingStatuses != nil {
		oldVersion = plan.Status.ClusterSchedulingStatuses[cluster].Version
	}
	newVersion := plan.Spec.Extension.Version
	if oldVersion == "" {
		return false
	}
	return newVersion != oldVersion
}

func configChanged(sub *corev1alpha1.InstallPlan, cluster string) bool {
	var oldConfigHash string
	if cluster == "" {
		oldConfigHash = sub.Status.InstallationStatus.ConfigHash
	} else {
		oldConfigHash = sub.Status.ClusterSchedulingStatuses[cluster].ConfigHash
	}
	newConfigHash := hashutil.FNVString(clusterConfig(sub, cluster))
	if oldConfigHash == "" {
		return true
	}
	return newConfigHash != oldConfigHash
}

// newHelmCred from Repository
func newHelmCred(repo *corev1alpha1.Repository) (helm.RepoCredential, error) {
	cred := helm.RepoCredential{
		InsecureSkipTLSVerify: repo.Spec.Insecure,
	}
	if repo.Spec.CABundle != "" {
		caFile, err := storeCAFile(repo.Spec.CABundle, repo.Name)
		if err != nil {
			return cred, err
		}
		cred.CAFile = caFile
	}
	if repo.Spec.BasicAuth != nil {
		cred.Username = repo.Spec.BasicAuth.Username
		cred.Password = repo.Spec.BasicAuth.Password
	}
	return cred, nil
}

// storeCAFile in local file from caTemplate.
func storeCAFile(caBundle string, repoName string) (string, error) {
	var buff = &bytes.Buffer{}
	tmpl, err := template.New("repositoryCABundle").Parse(caTemplate)
	if err != nil {
		return "", err
	}
	if err := tmpl.Execute(buff, map[string]string{
		"TempDIR":        os.TempDir(),
		"RepositoryName": repoName,
	}); err != nil {
		return "", err
	}
	caFile := buff.String()
	if _, err := os.Stat(filepath.Dir(caFile)); err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}

		if err := os.MkdirAll(filepath.Dir(caFile), os.ModePerm); err != nil {
			return "", err
		}
	}

	data, err := base64.StdEncoding.DecodeString(caBundle)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(caFile, data, os.ModePerm); err != nil {
		return "", err
	}

	return caFile, nil
}
