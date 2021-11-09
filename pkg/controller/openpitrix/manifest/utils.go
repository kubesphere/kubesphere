package controllers

import (
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"kubesphere.io/api/application/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func convertObjState(state string) (frontState string) {
	switch state {
	case v1alpha1.ClusterInitState, v1alpha1.StatusCreating, v1alpha1.PgclusterStateProcessed:
		frontState = v1alpha1.FrontCreating
	case v1alpha1.ClusterUpdateState, v1alpha1.StatusInProgress:
		frontState = v1alpha1.FrontUpdating
	case v1alpha1.StatusCompleted, v1alpha1.PgclusterStateCreated:
		frontState = v1alpha1.FrontCompleted
	case v1alpha1.ClusterReadyState, v1alpha1.StatusRunning, v1alpha1.PgclusterStateInitialized:
		frontState = v1alpha1.FrontRunning
	case v1alpha1.ClusterCloseState, v1alpha1.PgclusterStateShutdown:
		frontState = v1alpha1.FrontClosed
	case v1alpha1.StatusCreateFailed:
		frontState = v1alpha1.FrontCreateFailed
	case v1alpha1.StatusUpdateFailed:
		frontState = v1alpha1.FrontUpdateFailed
	case v1alpha1.StatusTerminating:
		frontState = v1alpha1.FrontTerminating
	case v1alpha1.PgclusterStateBootstrapping:
		frontState = v1alpha1.StatusBootstrapping
	case v1alpha1.PgclusterStateBootstrapped:
		frontState = v1alpha1.StatusBootstrapped
	case v1alpha1.PgclusterStateRestore:
		frontState = v1alpha1.StatusRestoring
	default:
		frontState = v1alpha1.ClusterStatusUnknown
	}
	return
}

func (r *ManifestReconciler) newClusterClient(clusterName string) (client.Client, error) {
	var clusterCli client.Client
	clusterInfo, err := r.clusterClients.Get(clusterName)
	if err != nil {
		klog.Errorf("get cluster(%s) info error: %s", clusterName, err)
		return nil, err
	}
	if !r.clusterClients.IsHostCluster(clusterInfo) {
		clusterCli = r.Client
	} else {
		config, err := clientcmd.RESTConfigFromKubeConfig(clusterInfo.Spec.Connection.KubeConfig)
		if err != nil {
			klog.Errorf("get cluster config error: %s", err)
			return nil, err
		}
		clusterCli, err = client.New(config, client.Options{Scheme: r.Scheme})
		if err != nil {
			klog.Errorf("get cluster client with kubeconfig error: %s", err)
			return nil, err
		}
	}
	return clusterCli, nil
}
