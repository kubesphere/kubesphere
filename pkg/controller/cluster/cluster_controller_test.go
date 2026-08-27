/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package cluster

import (
	"context"
	"sync"
	"testing"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"

	"kubesphere.io/kubesphere/pkg/config"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/controller/options"
)

// TestReconcileMemberClusterRetriesAfterFailedInstall guards against a regression where a
// failed KS Core install/upgrade attempt was recorded as if it had applied Spec.Config,
// so configChanged() never reported a change again and the controller stopped retrying
// (e.g. a private registry configured on "Add Cluster" that was unreachable on the first
// attempt would never be retried).
func TestReconcileMemberClusterRetriesAfterFailedInstall(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := clusterv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	cfgBytes, err := yaml.Marshal(config.New())
	if err != nil {
		t.Fatal(err)
	}
	kubesphereConfig := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.KubeSphereConfigName,
			Namespace: constants.KubeSphereNamespace,
		},
		Data: map[string]string{constants.KubeSphereConfigMapDataKey: string(cfgBytes)},
	}

	memberCluster := &clusterv1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: "member-with-private-registry"},
		Spec: clusterv1alpha1.ClusterSpec{
			Config: []byte("global:\n  imageRegistry: registry.example.com\n"),
		},
	}

	r := &Reconciler{
		Client:              fake.NewClientBuilder().WithScheme(scheme).WithObjects(kubesphereConfig, memberCluster).Build(),
		installLock:         &sync.Map{},
		HelmExecutorOptions: options.NewHelmExecutorOptions(),
	}

	// The ks-core chart directory doesn't exist in this test environment, so
	// installOrUpgradeKSCoreInMemberCluster fails here just like it would in production
	// if the configured registry were unreachable on the first attempt.
	if err := r.reconcileMemberCluster(context.Background(), memberCluster, nil); err != nil {
		t.Fatalf("reconcileMemberCluster returned an unexpected error: %v", err)
	}

	if !configChanged(memberCluster) {
		t.Fatal("configChanged() returned false after a failed install; the controller will never retry applying Spec.Config again")
	}
}
