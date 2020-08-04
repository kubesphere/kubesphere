/*
Copyright 2020 KubeSphere Authors

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

package cluster

import (
	"context"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	fedv1b1 "sigs.k8s.io/kubefed/pkg/apis/core/v1beta1"
	genericclient "sigs.k8s.io/kubefed/pkg/client/generic"
	"sigs.k8s.io/kubefed/pkg/kubefedctl/util"
)

// Following code copied from sigs.k8s.io/kubefed to avoid import collision

// UnjoinCluster performs all the necessary steps to remove the
// registration of a cluster from a KubeFed control plane provided the
// required set of parameters are passed in.
func unjoinCluster(hostConfig, clusterConfig *rest.Config, kubefedNamespace, hostClusterName, unjoiningClusterName string, forceDeletion, dryRun bool, skipMemberClusterResources bool) error {

	hostClientset, err := util.HostClientset(hostConfig)
	if err != nil {
		klog.V(2).Infof("Failed to get host cluster clientset: %v", err)
		return err
	}

	var clusterClientset *kubeclient.Clientset
	if clusterConfig != nil {
		clusterClientset, err = util.ClusterClientset(clusterConfig)
		if err != nil {
			klog.V(2).Infof("Failed to get unjoining cluster clientset: %v", err)
			if !forceDeletion {
				return err
			}
		}
	}

	client, err := genericclient.New(hostConfig)
	if err != nil {
		klog.V(2).Infof("Failed to get kubefed clientset: %v", err)
		return err
	}

	if clusterClientset != nil && !skipMemberClusterResources {
		err := deleteRBACResources(clusterClientset, kubefedNamespace, unjoiningClusterName, hostClusterName, forceDeletion, dryRun)
		if err != nil {
			if !forceDeletion {
				return err
			}
			klog.V(2).Infof("Failed to delete RBAC resources: %v", err)
		}

		err = deleteFedNSFromUnjoinCluster(hostClientset, clusterClientset, kubefedNamespace, unjoiningClusterName, dryRun)
		if err != nil {
			if !forceDeletion {
				return err
			}
			klog.V(2).Infof("Failed to delete kubefed namespace: %v", err)
		}
	}

	// deletionSucceeded when all operations in deleteRBACResources and deleteFedNSFromUnjoinCluster succeed.
	err = deleteFederatedClusterAndSecret(hostClientset, client, kubefedNamespace, unjoiningClusterName, forceDeletion, dryRun)
	if err != nil {
		return err
	}
	return nil
}

// deleteKubeFedClusterAndSecret deletes a federated cluster resource that associates
// the cluster and secret.
func deleteFederatedClusterAndSecret(hostClientset kubeclient.Interface, client genericclient.Client,
	kubefedNamespace, unjoiningClusterName string, forceDeletion, dryRun bool) error {
	if dryRun {
		return nil
	}

	klog.V(2).Infof("Deleting kubefed cluster resource from namespace %q for unjoin cluster %q",
		kubefedNamespace, unjoiningClusterName)

	fedCluster := &fedv1b1.KubeFedCluster{}
	err := client.Get(context.TODO(), fedCluster, kubefedNamespace, unjoiningClusterName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return errors.Wrapf(err, "Failed to get kubefed cluster \"%s/%s\"", kubefedNamespace, unjoiningClusterName)
	}

	err = hostClientset.CoreV1().Secrets(kubefedNamespace).Delete(fedCluster.Spec.SecretRef.Name,
		&metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		klog.V(2).Infof("Secret \"%s/%s\" does not exist in the host cluster.", kubefedNamespace, fedCluster.Spec.SecretRef.Name)
	} else if err != nil {
		wrappedErr := errors.Wrapf(err, "Failed to delete secret \"%s/%s\" for unjoin cluster %q",
			kubefedNamespace, fedCluster.Spec.SecretRef.Name, unjoiningClusterName)
		if !forceDeletion {
			return wrappedErr
		}
		klog.V(2).Infof("%v", wrappedErr)
	} else {
		klog.V(2).Infof("Deleted secret \"%s/%s\" for unjoin cluster %q", kubefedNamespace, fedCluster.Spec.SecretRef.Name, unjoiningClusterName)
	}

	err = client.Delete(context.TODO(), fedCluster, fedCluster.Namespace, fedCluster.Name)
	if apierrors.IsNotFound(err) {
		klog.V(2).Infof("KubeFed cluster \"%s/%s\" does not exist in the host cluster.", fedCluster.Namespace, fedCluster.Name)
	} else if err != nil {
		wrappedErr := errors.Wrapf(err, "Failed to delete kubefed cluster \"%s/%s\" for unjoin cluster %q", fedCluster.Namespace, fedCluster.Name, unjoiningClusterName)
		if !forceDeletion {
			return wrappedErr
		}
		klog.V(2).Infof("%v", wrappedErr)
	} else {
		klog.V(2).Infof("Deleted kubefed cluster \"%s/%s\" for unjoin cluster %q.", fedCluster.Namespace, fedCluster.Name, unjoiningClusterName)
	}

	return nil
}

// deleteRBACResources deletes the cluster role, cluster rolebindings and service account
// from the unjoining cluster.
func deleteRBACResources(unjoiningClusterClientset kubeclient.Interface,
	namespace, unjoiningClusterName, hostClusterName string, forceDeletion, dryRun bool) error {

	saName := ClusterServiceAccountName(unjoiningClusterName, hostClusterName)

	err := deleteClusterRoleAndBinding(unjoiningClusterClientset, saName, namespace, unjoiningClusterName, forceDeletion, dryRun)
	if err != nil {
		return err
	}

	err = deleteServiceAccount(unjoiningClusterClientset, saName, namespace, unjoiningClusterName, dryRun)
	if err != nil {
		return err
	}

	return nil
}

// deleteFedNSFromUnjoinCluster deletes the kubefed namespace from
// the unjoining cluster so long as the unjoining cluster is not the
// host cluster.
func deleteFedNSFromUnjoinCluster(hostClientset, unjoiningClusterClientset kubeclient.Interface,
	kubefedNamespace, unjoiningClusterName string, dryRun bool) error {

	if dryRun {
		return nil
	}

	hostClusterNamespace, err := hostClientset.CoreV1().Namespaces().Get(kubefedNamespace, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "Error retrieving namespace %q from host cluster", kubefedNamespace)
	}

	unjoiningClusterNamespace, err := unjoiningClusterClientset.CoreV1().Namespaces().Get(kubefedNamespace, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "Error retrieving namespace %q from unjoining cluster %q", kubefedNamespace, unjoiningClusterName)
	}

	if IsPrimaryCluster(hostClusterNamespace, unjoiningClusterNamespace) {
		klog.V(2).Infof("The kubefed namespace %q does not need to be deleted from the host cluster by unjoin.", kubefedNamespace)
		return nil
	}

	klog.V(2).Infof("Deleting kubefed namespace %q from unjoining cluster %q.", kubefedNamespace, unjoiningClusterName)
	err = unjoiningClusterClientset.CoreV1().Namespaces().Delete(kubefedNamespace, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		klog.V(2).Infof("The kubefed namespace %q no longer exists in unjoining cluster %q.", kubefedNamespace, unjoiningClusterName)
		return nil
	} else if err != nil {
		return errors.Wrapf(err, "Could not delete kubefed namespace %q from unjoining cluster %q", kubefedNamespace, unjoiningClusterName)
	} else {
		klog.V(2).Infof("Deleted kubefed namespace %q from unjoining cluster %q.", kubefedNamespace, unjoiningClusterName)
	}

	return nil
}

// deleteServiceAccount deletes a service account in the cluster associated
// with clusterClientset with credentials that are used by the host cluster
// to access its API server.
func deleteServiceAccount(clusterClientset kubeclient.Interface, saName,
	namespace, unjoiningClusterName string, dryRun bool) error {
	if dryRun {
		return nil
	}

	klog.V(2).Infof("Deleting service account \"%s/%s\" in unjoining cluster %q.", namespace, saName, unjoiningClusterName)

	// Delete a service account.
	err := clusterClientset.CoreV1().ServiceAccounts(namespace).Delete(saName,
		&metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		klog.V(2).Infof("Service account \"%s/%s\" does not exist.", namespace, saName)
	} else if err != nil {
		return errors.Wrapf(err, "Could not delete service account \"%s/%s\"", namespace, saName)
	} else {
		klog.V(2).Infof("Deleted service account \"%s/%s\" in unjoining cluster %q.", namespace, saName, unjoiningClusterName)
	}

	return nil
}

// deleteClusterRoleAndBinding deletes an RBAC cluster role and binding that
// allows the service account identified by saName to access all resources in
// all namespaces in the cluster associated with clusterClientset.
func deleteClusterRoleAndBinding(clusterClientset kubeclient.Interface,
	saName, namespace, unjoiningClusterName string, forceDeletion, dryRun bool) error {
	if dryRun {
		return nil
	}

	roleName := util.RoleName(saName)
	healthCheckRoleName := util.HealthCheckRoleName(saName, namespace)

	// Attempt to delete all role and role bindings created by join
	for _, name := range []string{roleName, healthCheckRoleName} {
		klog.V(2).Infof("Deleting cluster role binding %q for service account %q in unjoining cluster %q.",
			name, saName, unjoiningClusterName)

		err := clusterClientset.RbacV1().ClusterRoleBindings().Delete(name, &metav1.DeleteOptions{})
		if apierrors.IsNotFound(err) {
			klog.V(2).Infof("Cluster role binding %q for service account %q does not exist in unjoining cluster %q.",
				name, saName, unjoiningClusterName)
		} else if err != nil {
			wrappedErr := errors.Wrapf(err, "Could not delete cluster role binding %q for service account %q in unjoining cluster %q",
				name, saName, unjoiningClusterName)
			if !forceDeletion {
				return wrappedErr
			}
			klog.V(2).Infof("%v", wrappedErr)
		} else {
			klog.V(2).Infof("Deleted cluster role binding %q for service account %q in unjoining cluster %q.",
				name, saName, unjoiningClusterName)
		}

		klog.V(2).Infof("Deleting cluster role %q for service account %q in unjoining cluster %q.",
			name, saName, unjoiningClusterName)
		err = clusterClientset.RbacV1().ClusterRoles().Delete(name, &metav1.DeleteOptions{})
		if apierrors.IsNotFound(err) {
			klog.V(2).Infof("Cluster role %q for service account %q does not exist in unjoining cluster %q.",
				name, saName, unjoiningClusterName)
		} else if err != nil {
			wrappedErr := errors.Wrapf(err, "Could not delete cluster role %q for service account %q in unjoining cluster %q",
				name, saName, unjoiningClusterName)
			if !forceDeletion {
				return wrappedErr
			}
			klog.V(2).Infof("%v", wrappedErr)
		} else {
			klog.V(2).Infof("Deleted cluster role %q for service account %q in unjoining cluster %q.",
				name, saName, unjoiningClusterName)
		}
	}

	klog.V(2).Infof("Deleting role binding \"%s/%s\" for service account %q in unjoining cluster %q.",
		namespace, roleName, saName, unjoiningClusterName)
	err := clusterClientset.RbacV1().RoleBindings(namespace).Delete(roleName, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		klog.V(2).Infof("Role binding \"%s/%s\" for service account %q does not exist in unjoining cluster %q.",
			namespace, roleName, saName, unjoiningClusterName)
	} else if err != nil {
		wrappedErr := errors.Wrapf(err, "Could not delete role binding \"%s/%s\" for service account %q in unjoining cluster %q",
			namespace, roleName, saName, unjoiningClusterName)
		if !forceDeletion {
			return wrappedErr
		}
		klog.V(2).Infof("%v", wrappedErr)
	} else {
		klog.V(2).Infof("Deleted role binding \"%s/%s\" for service account %q in unjoining cluster %q.",
			namespace, roleName, saName, unjoiningClusterName)
	}

	klog.V(2).Infof("Deleting role \"%s/%s\" for service account %q in unjoining cluster %q.",
		namespace, roleName, saName, unjoiningClusterName)
	err = clusterClientset.RbacV1().Roles(namespace).Delete(roleName, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		klog.V(2).Infof("Role \"%s/%s\" for service account %q does not exist in unjoining cluster %q.",
			namespace, roleName, saName, unjoiningClusterName)
	} else if err != nil {
		wrappedErr := errors.Wrapf(err, "Could not delete role \"%s/%s\" for service account %q in unjoining cluster %q",
			namespace, roleName, saName, unjoiningClusterName)
		if !forceDeletion {
			return wrappedErr
		}
		klog.V(2).Infof("%v", wrappedErr)
	} else {
		klog.V(2).Infof("Deleting Role \"%s/%s\" for service account %q in unjoining cluster %q.",
			namespace, roleName, saName, unjoiningClusterName)
	}

	return nil
}
