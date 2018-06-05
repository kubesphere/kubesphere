/*
Copyright 2018 The Kubernetes Authors.

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

// This file is used to deploy the CSI hostPath plugin
// More Information: https://github.com/kubernetes-csi/drivers/tree/master/pkg/hostpath

package storage

import (
	"fmt"
	"time"

	"k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/manifest"

	. "github.com/onsi/ginkgo"
)

const (
	csiHostPathPluginImage      string = "quay.io/k8scsi/hostpathplugin:v0.2.0"
	csiExternalAttacherImage    string = "quay.io/k8scsi/csi-attacher:v0.2.0"
	csiExternalProvisionerImage string = "quay.io/k8scsi/csi-provisioner:v0.2.1"
	csiDriverRegistrarImage     string = "quay.io/k8scsi/driver-registrar:v0.2.0"
)

// Create the driver registrar cluster role if it doesn't exist, no teardown so that tests
// are parallelizable. This role will be shared with many of the CSI tests.
func csiDriverRegistrarClusterRole(
	config framework.VolumeTestConfig,
) *rbacv1.ClusterRole {
	// TODO(Issue: #62237) Remove impersonation workaround and cluster role when issue resolved
	By("Creating an impersonating superuser kubernetes clientset to define cluster role")
	rc, err := framework.LoadConfig()
	framework.ExpectNoError(err)
	rc.Impersonate = restclient.ImpersonationConfig{
		UserName: "superuser",
		Groups:   []string{"system:masters"},
	}
	superuserClientset, err := clientset.NewForConfig(rc)
	framework.ExpectNoError(err, "Failed to create superuser clientset: %v", err)
	By("Creating the CSI driver registrar cluster role")
	clusterRoleClient := superuserClientset.RbacV1().ClusterRoles()
	role := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: csiDriverRegistrarClusterRoleName,
		},
		Rules: []rbacv1.PolicyRule{

			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"get", "update", "patch"},
			},
		},
	}

	ret, err := clusterRoleClient.Create(role)
	if err != nil {
		if apierrs.IsAlreadyExists(err) {
			return ret
		}
		framework.ExpectNoError(err, "Failed to create %s cluster role: %v", role.GetName(), err)
	}

	return ret
}

func csiServiceAccount(
	client clientset.Interface,
	config framework.VolumeTestConfig,
	componentName string,
	teardown bool,
) *v1.ServiceAccount {
	creatingString := "Creating"
	if teardown {
		creatingString = "Deleting"
	}
	By(fmt.Sprintf("%v a CSI service account for %v", creatingString, componentName))
	serviceAccountName := config.Prefix + "-" + componentName + "-service-account"
	serviceAccountClient := client.CoreV1().ServiceAccounts(config.Namespace)
	sa := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceAccountName,
		},
	}

	serviceAccountClient.Delete(sa.GetName(), &metav1.DeleteOptions{})
	err := wait.Poll(2*time.Second, 10*time.Minute, func() (bool, error) {
		_, err := serviceAccountClient.Get(sa.GetName(), metav1.GetOptions{})
		return apierrs.IsNotFound(err), nil
	})
	framework.ExpectNoError(err, "Timed out waiting for deletion: %v", err)

	if teardown {
		return nil
	}

	ret, err := serviceAccountClient.Create(sa)
	if err != nil {
		framework.ExpectNoError(err, "Failed to create %s service account: %v", sa.GetName(), err)
	}

	return ret
}

func csiClusterRoleBindings(
	client clientset.Interface,
	config framework.VolumeTestConfig,
	teardown bool,
	sa *v1.ServiceAccount,
	clusterRolesNames []string,
) {
	bindingString := "Binding"
	if teardown {
		bindingString = "Unbinding"
	}
	By(fmt.Sprintf("%v cluster roles %v to the CSI service account %v", bindingString, clusterRolesNames, sa.GetName()))
	clusterRoleBindingClient := client.RbacV1().ClusterRoleBindings()
	for _, clusterRoleName := range clusterRolesNames {

		binding := &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: config.Prefix + "-" + clusterRoleName + "-" + config.Namespace + "-role-binding",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      sa.GetName(),
					Namespace: sa.GetNamespace(),
				},
			},
			RoleRef: rbacv1.RoleRef{
				Kind:     "ClusterRole",
				Name:     clusterRoleName,
				APIGroup: "rbac.authorization.k8s.io",
			},
		}

		clusterRoleBindingClient.Delete(binding.GetName(), &metav1.DeleteOptions{})
		err := wait.Poll(2*time.Second, 10*time.Minute, func() (bool, error) {
			_, err := clusterRoleBindingClient.Get(binding.GetName(), metav1.GetOptions{})
			return apierrs.IsNotFound(err), nil
		})
		framework.ExpectNoError(err, "Timed out waiting for deletion: %v", err)

		if teardown {
			return
		}

		_, err = clusterRoleBindingClient.Create(binding)
		if err != nil {
			framework.ExpectNoError(err, "Failed to create %s role binding: %v", binding.GetName(), err)
		}
	}
}

func csiHostPathPod(
	client clientset.Interface,
	config framework.VolumeTestConfig,
	teardown bool,
	f *framework.Framework,
	sa *v1.ServiceAccount,
) *v1.Pod {
	podClient := client.CoreV1().Pods(config.Namespace)

	priv := true
	mountPropagation := v1.MountPropagationBidirectional
	hostPathType := v1.HostPathDirectoryOrCreate
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.Prefix + "-pod",
			Namespace: config.Namespace,
			Labels: map[string]string{
				"app": "hostpath-driver",
			},
		},
		Spec: v1.PodSpec{
			ServiceAccountName: sa.GetName(),
			NodeName:           config.ServerNodeName,
			RestartPolicy:      v1.RestartPolicyNever,
			Containers: []v1.Container{
				{
					Name:            "external-provisioner",
					Image:           csiExternalProvisionerImage,
					ImagePullPolicy: v1.PullAlways,
					Args: []string{
						"--v=5",
						"--provisioner=csi-hostpath",
						"--csi-address=/csi/csi.sock",
					},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "socket-dir",
							MountPath: "/csi",
						},
					},
				},
				{
					Name:            "driver-registrar",
					Image:           csiDriverRegistrarImage,
					ImagePullPolicy: v1.PullAlways,
					Args: []string{
						"--v=5",
						"--csi-address=/csi/csi.sock",
					},
					Env: []v1.EnvVar{
						{
							Name: "KUBE_NODE_NAME",
							ValueFrom: &v1.EnvVarSource{
								FieldRef: &v1.ObjectFieldSelector{
									FieldPath: "spec.nodeName",
								},
							},
						},
					},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "socket-dir",
							MountPath: "/csi",
						},
					},
				},
				{
					Name:            "external-attacher",
					Image:           csiExternalAttacherImage,
					ImagePullPolicy: v1.PullAlways,
					Args: []string{
						"--v=5",
						"--csi-address=$(ADDRESS)",
					},
					Env: []v1.EnvVar{
						{
							Name:  "ADDRESS",
							Value: "/csi/csi.sock",
						},
					},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "socket-dir",
							MountPath: "/csi",
						},
					},
				},
				{
					Name:            "hostpath-driver",
					Image:           csiHostPathPluginImage,
					ImagePullPolicy: v1.PullAlways,
					SecurityContext: &v1.SecurityContext{
						Privileged: &priv,
					},
					Args: []string{
						"--v=5",
						"--endpoint=$(CSI_ENDPOINT)",
						"--nodeid=$(KUBE_NODE_NAME)",
					},
					Env: []v1.EnvVar{
						{
							Name:  "CSI_ENDPOINT",
							Value: "unix://" + "/csi/csi.sock",
						},
						{
							Name: "KUBE_NODE_NAME",
							ValueFrom: &v1.EnvVarSource{
								FieldRef: &v1.ObjectFieldSelector{
									FieldPath: "spec.nodeName",
								},
							},
						},
					},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "socket-dir",
							MountPath: "/csi",
						},
						{
							Name:             "mountpoint-dir",
							MountPath:        "/var/lib/kubelet/pods",
							MountPropagation: &mountPropagation,
						},
					},
				},
			},
			Volumes: []v1.Volume{
				{
					Name: "socket-dir",
					VolumeSource: v1.VolumeSource{
						HostPath: &v1.HostPathVolumeSource{
							Path: "/var/lib/kubelet/plugins/csi-hostpath",
							Type: &hostPathType,
						},
					},
				},
				{
					Name: "mountpoint-dir",
					VolumeSource: v1.VolumeSource{
						HostPath: &v1.HostPathVolumeSource{
							Path: "/var/lib/kubelet/pods",
							Type: &hostPathType,
						},
					},
				},
			},
		},
	}

	err := framework.DeletePodWithWait(f, client, pod)
	framework.ExpectNoError(err, "Failed to delete pod %s/%s: %v",
		pod.GetNamespace(), pod.GetName(), err)

	if teardown {
		return nil
	}

	ret, err := podClient.Create(pod)
	if err != nil {
		framework.ExpectNoError(err, "Failed to create %q pod: %v", pod.GetName(), err)
	}

	// Wait for pod to come up
	framework.ExpectNoError(framework.WaitForPodRunningInNamespace(client, ret))
	return ret
}

func deployGCEPDCSIDriver(
	client clientset.Interface,
	config framework.VolumeTestConfig,
	teardown bool,
	f *framework.Framework,
	nodeSA *v1.ServiceAccount,
	controllerSA *v1.ServiceAccount,
) {
	// Get API Objects from manifests
	nodeds, err := manifest.DaemonSetFromManifest("test/e2e/testing-manifests/storage-csi/gce-pd/node_ds.yaml", config.Namespace)
	framework.ExpectNoError(err, "Failed to create DaemonSet from manifest")
	nodeds.Spec.Template.Spec.ServiceAccountName = nodeSA.GetName()

	controllerss, err := manifest.StatefulSetFromManifest("test/e2e/testing-manifests/storage-csi/gce-pd/controller_ss.yaml", config.Namespace)
	framework.ExpectNoError(err, "Failed to create StatefulSet from manifest")
	controllerss.Spec.Template.Spec.ServiceAccountName = controllerSA.GetName()

	controllerservice, err := manifest.SvcFromManifest("test/e2e/testing-manifests/storage-csi/gce-pd/controller_service.yaml")
	framework.ExpectNoError(err, "Failed to create Service from manifest")

	// Got all objects from manifests now try to delete objects
	err = client.CoreV1().Services(config.Namespace).Delete(controllerservice.GetName(), nil)
	if err != nil {
		if !apierrs.IsNotFound(err) {
			framework.ExpectNoError(err, "Failed to delete Service: %v", controllerservice.GetName())
		}
	}

	err = client.AppsV1().StatefulSets(config.Namespace).Delete(controllerss.Name, nil)
	if err != nil {
		if !apierrs.IsNotFound(err) {
			framework.ExpectNoError(err, "Failed to delete StatefulSet: %v", controllerss.GetName())
		}
	}
	err = client.AppsV1().DaemonSets(config.Namespace).Delete(nodeds.Name, nil)
	if err != nil {
		if !apierrs.IsNotFound(err) {
			framework.ExpectNoError(err, "Failed to delete DaemonSet: %v", nodeds.GetName())
		}
	}
	if teardown {
		return
	}

	// Create new API Objects through client
	_, err = client.CoreV1().Services(config.Namespace).Create(controllerservice)
	framework.ExpectNoError(err, "Failed to create Service: %v", controllerservice.Name)

	_, err = client.AppsV1().StatefulSets(config.Namespace).Create(controllerss)
	framework.ExpectNoError(err, "Failed to create StatefulSet: %v", controllerss.Name)

	_, err = client.AppsV1().DaemonSets(config.Namespace).Create(nodeds)
	framework.ExpectNoError(err, "Failed to create DaemonSet: %v", nodeds.Name)

}
