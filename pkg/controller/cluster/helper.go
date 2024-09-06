/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package cluster

import (
	"context"
	"errors"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/storage/driver"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	"kubesphere.io/utils/helm"
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

func installKSCoreInMemberCluster(kubeConfig []byte, jwtSecret, chartPath string, chartConfig []byte) error {
	helmConf, err := helm.InitHelmConf(kubeConfig, constants.KubeSphereNamespace)
	if err != nil {
		return err
	}

	if chartPath == "" {
		chartPath = "/var/helm-charts/ks-core"
	}
	chart, err := loader.Load(chartPath) // in-container chart path
	if err != nil {
		return err
	}

	// values example:
	// 	map[string]interface{}{
	//		"nestedKey": map[string]interface{}{
	//			"simpleKey": "simpleValue",
	//		},
	//  }
	values := make(map[string]interface{})
	if chartConfig != nil {
		if err = yaml.Unmarshal(chartConfig, &values); err != nil {
			return err
		}
	}

	// Override some necessary values
	values["role"] = "member"
	// disable upgrade to prevent execution of ks-upgrade
	values["upgrade"] = map[string]interface{}{
		"enabled": false,
	}
	if err = unstructured.SetNestedField(values, jwtSecret, "authentication", "issuer", "jwtSecret"); err != nil {
		return err
	}

	helmStatus := action.NewStatus(helmConf)
	if _, err = helmStatus.Run(releaseName); err != nil {
		if !errors.Is(err, driver.ErrReleaseNotFound) {
			return err
		}

		// the release not exists
		install := action.NewInstall(helmConf)
		install.Namespace = constants.KubeSphereNamespace
		install.CreateNamespace = true
		install.Wait = true
		install.ReleaseName = releaseName
		install.Timeout = time.Minute * 5
		if _, err = install.Run(chart, values); err != nil {
			return err
		}
		return nil
	}

	upgrade := action.NewUpgrade(helmConf)
	upgrade.Namespace = constants.KubeSphereNamespace
	upgrade.Install = true
	upgrade.Wait = true
	upgrade.Timeout = time.Minute * 5
	if _, err = upgrade.Run(releaseName, chart, values); err != nil {
		return err
	}
	return nil
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

func hasCondition(conditions []clusterv1alpha1.ClusterCondition, conditionsType clusterv1alpha1.ClusterConditionType) bool {
	for _, condition := range conditions {
		if condition.Type == conditionsType && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
