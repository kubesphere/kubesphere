/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package cluster

import (
	"context"
	"fmt"
	"os"
	"path"

	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"kubesphere.io/kubesphere/pkg/config"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/utils/hashutil"
)

const releaseName = "ks-core"

func configChanged(cluster *clusterv1alpha1.Cluster) bool {
	return hashutil.FNVString(cluster.Spec.Config) != cluster.Annotations[constants.ConfigHashAnnotation]
}

func setConfigHash(cluster *clusterv1alpha1.Cluster) {
	configHash := hashutil.FNVString(cluster.Spec.Config)
	if cluster.Annotations == nil {
		cluster.Annotations = map[string]string{
			constants.ConfigHashAnnotation: configHash,
		}
	} else {
		cluster.Annotations[constants.ConfigHashAnnotation] = configHash
	}
}

func getKubeSphereConfig(ctx context.Context, client runtimeclient.Client) (*config.Config, error) {
	cm := &corev1.ConfigMap{}
	if err := client.Get(ctx, types.NamespacedName{Name: constants.KubeSphereConfigName, Namespace: constants.KubeSphereNamespace}, cm); err != nil {
		return nil, err
	}
	configData, err := config.FromConfigMap(cm)
	if err != nil {
		return nil, err
	}
	return configData, nil
}

// generateChartValueBytes generates the chart value bytes for the cluster
func generateChartValueBytes(chartConfig []byte, jwtSecret string) ([]byte, error) {
	values := make(map[string]interface{})
	if chartConfig != nil {
		if err := yaml.Unmarshal(chartConfig, &values); err != nil {
			return nil, err
		}
	}

	// Override some necessary values
	values["role"] = "member"
	values["multicluster"] = map[string]string{"role": "member"}
	// disable upgrade to prevent execution of kse-upgrade
	values["upgrade"] = map[string]interface{}{
		"enabled": false,
	}
	if err := unstructured.SetNestedField(values, jwtSecret, "authentication", "issuer", "jwtSecret"); err != nil {
		return nil, err
	}

	valuesBytes, err := yaml.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal values: %v", err)
	}
	return valuesBytes, nil
}

func getChartBytes(chartPath string) ([]byte, error) {
	prefix := "/var/helm-charts"
	if chartPath == "" {
		chartPath = path.Join(prefix, releaseName)
	}

	tgzFile := path.Join(prefix, fmt.Sprintf("%s.tgz", releaseName))
	if _, err := os.Stat(tgzFile); os.IsNotExist(err) {
		chart, err := loader.Load(chartPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load chart: %v", err)
		}

		saveFile, err := chartutil.Save(chart, prefix)
		if err != nil {
			return nil, fmt.Errorf("failed to save chart: %v", err)
		}

		klog.Infof("saveFile %s, tgzFile %s", saveFile, tgzFile)
		if saveFile != tgzFile {
			if err := os.Rename(saveFile, tgzFile); err != nil {
				return nil, fmt.Errorf("failed to rename chart file: %v", err)
			}
		}
	}

	chartBytes, err := os.ReadFile(tgzFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read chart files: %v", err)
	}
	return chartBytes, nil
}
