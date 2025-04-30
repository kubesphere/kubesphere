/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package core

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"path"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	yaml3 "gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/storage/driver"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog/v2"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
	"kubesphere.io/utils/helm"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/utils/hashutil"
	"kubesphere.io/kubesphere/pkg/version"
)

const ExtensionVersionMaxLength = validation.LabelValueMaxLength
const ExtensionNameMaxLength = validation.LabelValueMaxLength

func getRecommendedExtensionVersion(versions []corev1alpha1.ExtensionVersion, k8sVersion *semver.Version) (string, error) {
	if len(versions) == 0 {
		return "", nil
	}

	ksVersion, err := semver.NewVersion(version.Get().GitVersion)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse KS version: %s", version.Get().GitVersion)
	}

	var matchedVersions []*semver.Version
	for _, v := range versions {
		kubeVersionMatched, ksVersionMatched := matchVersionConstraints(v, k8sVersion, ksVersion)
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

func matchVersionConstraints(v corev1alpha1.ExtensionVersion, k8sVersion, ksVersion *semver.Version) (bool, bool) {
	kubeVersionMatched := v.Spec.KubeVersion == "" || checkVersionConstraint(v.Spec.KubeVersion, k8sVersion)
	ksVersionMatched := v.Spec.KSVersion == "" || checkVersionConstraint(v.Spec.KSVersion, ksVersion)
	return kubeVersionMatched, ksVersionMatched
}

func checkVersionConstraint(constraint string, version *semver.Version) bool {
	targetVersion, err := semver.NewConstraint(constraint)
	if err != nil {
		klog.Warningf("failed to parse version constraints: %s, err: %s", constraint, err)
		return false
	}
	return targetVersion.Check(version)
}

// filterExtensionVersions filters and sorts a slice of ExtensionVersion objects based on semantic versioning.
// It first validates and removes entries with invalid versions (non-semver format) and logs warnings for them.
// The remaining entries are sorted in descending order by version (latest first).
// Finally, the slice is truncated to the specified depth:
//   - If depth is nil, it defaults to a pre-configured depth (DefaultRepositoryDepth).
//   - If depth is 0, all valid entries are kept.
//
// The function returns the filtered and truncated list of ExtensionVersion objects.
func filterExtensionVersions(versions []corev1alpha1.ExtensionVersion, depth *int) []corev1alpha1.ExtensionVersion {
	// Filter and parse valid versions.
	parsedVersions := make([]struct {
		semver   *semver.Version
		original corev1alpha1.ExtensionVersion
	}, 0, len(versions))

	for _, v := range versions {
		parsed, err := semver.NewVersion(v.Spec.Version)
		if err != nil {
			klog.Warningf("failed to parse version, extension %s version: %s, err: %s", v.Name, v.Spec.Version, err)
			continue
		}
		parsedVersions = append(parsedVersions, struct {
			semver   *semver.Version
			original corev1alpha1.ExtensionVersion
		}{semver: parsed, original: v})
	}

	// Sort by descending semantic version.
	slices.SortFunc(parsedVersions, func(a, b struct {
		semver   *semver.Version
		original corev1alpha1.ExtensionVersion
	}) int {
		return b.semver.Compare(a.semver)
	})

	// Determine truncation length.
	end := len(parsedVersions)
	if depth == nil {
		end = corev1alpha1.DefaultRepositoryDepth
	} else if *depth > 0 && *depth < len(parsedVersions) {
		end = *depth
	}
	if end > len(parsedVersions) {
		end = len(parsedVersions)
	}

	// Extract the truncated versions.
	filteredVersions := make([]corev1alpha1.ExtensionVersion, end)
	for i := 0; i < end; i++ {
		filteredVersions[i] = parsedVersions[i].original
	}
	return filteredVersions
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
				if errors.Is(err, io.EOF) {
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

// createHelmCredential from Repository
func createHelmCredential(repo *corev1alpha1.Repository) (helm.RepoCredential, error) {
	cred := helm.RepoCredential{
		InsecureSkipTLSVerify: repo.Spec.Insecure,
		CABundle:              repo.Spec.CABundle,
	}
	if repo.Spec.BasicAuth != nil {
		cred.Username = repo.Spec.BasicAuth.Username
		cred.Password = repo.Spec.BasicAuth.Password
	}
	return cred, nil
}

func fetchExtensionVersionSpec(ctx context.Context, client client.Reader, extensionVersion *corev1alpha1.ExtensionVersion) (corev1alpha1.ExtensionVersionSpec, error) {
	extensionVersionSpec := extensionVersion.Spec
	logger := klog.FromContext(ctx)
	data, err := fetchChartData(ctx, client, extensionVersion)
	if err != nil {
		return extensionVersionSpec, errors.Wrapf(err, "failed to fetch chart data")
	}
	helmChart, err := loader.LoadArchive(bytes.NewReader(data))
	if err != nil {
		return extensionVersionSpec, errors.Wrapf(err, "failed to load chart archive")
	}
	errs := isValidExtensionVersion(helmChart.Metadata.Version)
	if len(errs) > 0 {
		logger.V(4).Info("invalid extension version", "errors", errs)
		return extensionVersionSpec, nil
	}
	for _, file := range helmChart.Files {
		if file.Name == extensionFileName {
			if err := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(file.Data), 1024).Decode(&extensionVersionSpec); err != nil {
				logger.V(4).Info("failed to decode extension.yaml", "error", err)
				return extensionVersionSpec, nil
			}
			break
		}
	}
	extensionVersionSpec.Name = helmChart.Name()
	absPath := strings.TrimPrefix(extensionVersionSpec.Icon, "./")
	var iconData []byte
	for _, file := range helmChart.Files {
		if file.Name == absPath {
			iconData = file.Data
			break
		}
	}

	if iconData != nil {
		mimeType := mime.TypeByExtension(path.Ext(extensionVersionSpec.Icon))
		if mimeType == "" {
			mimeType = http.DetectContentType(iconData)
		}
		base64EncodedData := base64.StdEncoding.EncodeToString(iconData)
		extensionVersionSpec.Icon = fmt.Sprintf("data:%s;base64,%s", mimeType, base64EncodedData)
	}

	return extensionVersionSpec, nil
}

func fetchChartData(ctx context.Context, client client.Reader, extensionVersion *corev1alpha1.ExtensionVersion) ([]byte, error) {
	if extensionVersion.Spec.ChartDataRef != nil {
		return fetchChartDataFromConfigMap(ctx, client, extensionVersion.Spec.ChartDataRef)
	}

	chartURL, err := url.Parse(extensionVersion.Spec.ChartURL)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse chart URL: %s", extensionVersion.Spec.ChartURL)
	}

	repo, err := fetchRepository(ctx, client, extensionVersion.Spec.Repository)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch repository: %s", extensionVersion.Spec.Repository)
	}

	repoURL, err := url.Parse(repo.Spec.URL)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse repo URL: %s", extensionVersion.Spec.ChartURL)
	}

	if chartURL.Host == "" {
		chartURL.Scheme = repoURL.Scheme
		chartURL.Host = repoURL.Host
	}

	transport, err := createTransport(repo, chartURL.Hostname())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create transport")
	}

	opts := createGetterOptions(repo, transport)
	chartGetter, err := createChartGetter(chartURL.Scheme, opts)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create chart getter")
	}

	return getChartData(chartGetter, chartURL.String())
}

func fetchChartDataFromConfigMap(ctx context.Context, client client.Reader, ref *corev1alpha1.ConfigMapKeyRef) ([]byte, error) {
	configMap := &corev1.ConfigMap{}
	if err := client.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, configMap); err != nil {
		return nil, errors.Wrapf(err, "failed to get config map: %s", ref.Name)
	}
	data := configMap.BinaryData[ref.Key]
	if data != nil {
		return data, nil
	}
	return nil, errors.New("chart data not found in config map")
}

func fetchRepository(ctx context.Context, client client.Reader, repoName string) (*corev1alpha1.Repository, error) {
	if repoName == "" {
		return &corev1alpha1.Repository{}, nil
	}
	repo := &corev1alpha1.Repository{}
	if err := client.Get(ctx, types.NamespacedName{Name: repoName}, repo); err != nil {
		return nil, errors.Wrapf(err, "failed to get repository: %s", repoName)
	}
	return repo, nil
}

func createTransport(repo *corev1alpha1.Repository, serverName string) (*http.Transport, error) {
	tlsConf, err := helm.NewTLSConfig(repo.Spec.CABundle, repo.Spec.Insecure)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create tls config")
	}
	tlsConf.ServerName = serverName

	return &http.Transport{
		DisableCompression:    true,
		DialContext:           (&net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		Proxy:                 http.ProxyFromEnvironment,
		TLSClientConfig:       tlsConf,
	}, nil
}

func createGetterOptions(repo *corev1alpha1.Repository, transport *http.Transport) []getter.Option {
	opts := []getter.Option{getter.WithTransport(transport)}
	if repo.Spec.BasicAuth != nil {
		opts = append(opts, getter.WithBasicAuth(repo.Spec.BasicAuth.Username, repo.Spec.BasicAuth.Password))
	}
	return opts
}

func createChartGetter(scheme string, opts []getter.Option) (getter.Getter, error) {
	switch scheme {
	case registry.OCIScheme:
		return getter.NewOCIGetter(opts...)
	case "http", "https":
		return getter.NewHTTPGetter(opts...)
	default:
		return nil, errors.Errorf("unsupported scheme: %s", scheme)
	}
}

func getChartData(chartGetter getter.Getter, url string) ([]byte, error) {
	buffer, err := chartGetter.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch chart data: %s", url)
	}

	data, err := io.ReadAll(buffer)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read chart data: %s", url)
	}
	return data, nil
}

func isValidExtensionVersion(version string) []string {
	var errs []string
	if len(version) > ExtensionVersionMaxLength {
		errs = append(errs, fmt.Sprintf("extension version length exceeds %d", ExtensionVersionMaxLength))
	}
	if _, err := semver.NewVersion(version); err != nil {
		errs = append(errs, fmt.Sprintf("invalid semver format: %s", err))
	}
	if len(validation.IsDNS1123Subdomain(version)) > 0 {
		errs = append(errs, "invalid DNS-1123 subdomain")
	}
	return errs
}

func isValidExtensionName(name string) []string {
	var errs []string
	if name == "" {
		errs = append(errs, "extension name should not be empty")
	}
	if len(name) > ExtensionNameMaxLength {
		errs = append(errs, fmt.Sprintf("extension name length exceeds %d", ExtensionNameMaxLength))
	}
	if len(validation.IsDNS1123Subdomain(name)) > 0 {
		errs = append(errs, "invalid DNS-1123 subdomain")
	}
	return errs
}
