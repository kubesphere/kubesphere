/*
Copyright 2018 The KubeSphere Authors.

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

package models

import (
	"io/ioutil"
	"strings"

	"github.com/golang/glog"
	coreV1 "k8s.io/api/core/v1"
	extensionsV1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/api/rbac/v1beta1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"

	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/constants"
)

func GetAllRouters() ([]*coreV1.Service, error) {

	k8sClient := client.NewK8sClient()

	routers := make([]*coreV1.Service, 0)

	opts := metaV1.ListOptions{}

	namespaces, err := k8sClient.CoreV1().Namespaces().List(opts)

	if err != nil {
		glog.Error(err)
		return routers, err
	}

	opts = metaV1.ListOptions{
		LabelSelector: "app=kubesphere,component=kubesphere-router",
		FieldSelector: "metadata.name=kubesphere-router-gateway",
	}

	for _, namespace := range namespaces.Items {
		services, err := k8sClient.CoreV1().Services(namespace.Name).List(opts)

		if err != nil {
			glog.Error(err)
			return nil, err
		}

		if len(services.Items) > 0 {
			routers = append(routers, &services.Items[0])
		}
	}

	return routers, nil
}

// Get router from a namespace
func GetRouter(namespace string) (*coreV1.Service, error) {
	k8sClient := client.NewK8sClient()

	var router *coreV1.Service

	opts := metaV1.ListOptions{
		LabelSelector: "app=kubesphere,component=kubesphere-router",
		FieldSelector: "metadata.name=kubesphere-router-gateway",
	}

	services, err := k8sClient.CoreV1().Services(namespace).List(opts)

	if err != nil {
		glog.Error(err)
		return nil, err
	}

	if len(services.Items) > 0 {
		router = &services.Items[0]
	}

	return router, nil
}

// Load all resource yamls
func LoadYamls() ([]string, error) {

	var yamls []string

	files, err := ioutil.ReadDir(constants.INGRESS_CONTROLLER_FOLDER)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	for _, file := range files {
		content, err := ioutil.ReadFile(constants.INGRESS_CONTROLLER_FOLDER + "/" + file.Name())

		if err != nil {
			glog.Error(err)
			return nil, err
		} else {
			yamls = append(yamls, string(content))
		}
	}

	return yamls, nil
}

func IsRouterService(serviceName string) bool {
	if strings.Compare(strings.ToLower(serviceName), "default-http-backend") == 0 {
		return false
	}
	return true
}

// Create a ingress controller in a namespace
func CreateRouter(namespace string, routerType coreV1.ServiceType, annotations map[string]string) (*coreV1.Service, error) {

	k8sClient := client.NewK8sClient()

	var router *coreV1.Service

	yamls, err := LoadYamls()

	if err != nil {
		glog.Error(err)
	}

	for _, f := range yamls {
		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, _, err := decode([]byte(f), nil, nil)

		if err != nil {
			glog.Error(err)
			return router, err
		}

		switch obj.(type) {
		case *v1beta1.Role:
			role := obj.(*v1beta1.Role)
			role, err := k8sClient.RbacV1beta1().Roles(namespace).Create(role)
			if err != nil {
				glog.Error(err)
			}

		case *v1beta1.ClusterRole:
			clusterRole := obj.(*v1beta1.ClusterRole)

			clusterRole, err := k8sClient.RbacV1beta1().ClusterRoles().Create(clusterRole)
			if err != nil {
				glog.Error(err)
			}
		case *v1beta1.ClusterRoleBinding:
			clusterRoleBinding := obj.(*v1beta1.ClusterRoleBinding)
			clusterRoleBinding.Subjects[0].Namespace = namespace
			clusterRoleBinding, err := k8sClient.RbacV1beta1().ClusterRoleBindings().Create(clusterRoleBinding)
			if err != nil {
				glog.Error(err)
			}
		case *v1beta1.RoleBinding:
			roleBinding := obj.(*v1beta1.RoleBinding)
			roleBinding.Subjects[0].Namespace = namespace
			roleBinding, err := k8sClient.RbacV1beta1().RoleBindings(namespace).Create(roleBinding)

			if err != nil {
				glog.Error(err)
			}
		case *coreV1.ServiceAccount:
			sa := obj.(*coreV1.ServiceAccount)
			sa, err := k8sClient.CoreV1().ServiceAccounts(namespace).Create(sa)
			if err != nil {
				glog.Error(err)
			}
		case *coreV1.Service:
			service := obj.(*coreV1.Service)

			if IsRouterService(service.Name) {
				service.SetAnnotations(annotations)
				service.Spec.Type = routerType
			}

			service, err := k8sClient.CoreV1().Services(namespace).Create(service)
			if err != nil {
				glog.Error(err)
				return nil, err
			}

			if IsRouterService(service.Name) {
				router = service
			}

		case *extensionsV1beta1.Deployment:
			deployment := obj.(*extensionsV1beta1.Deployment)
			deployment, err := k8sClient.ExtensionsV1beta1().Deployments(namespace).Create(deployment)
			if err != nil {
				glog.Error(err)
			}
		default:
			//glog.Info("Default resource")
		}
	}

	return router, nil
}

// DeleteRouter is used to delete ingress controller related resources in namespace
// It will not delete ClusterRole resource cause it maybe used other controllers
func DeleteRouter(namespace string) (*coreV1.Service, error) {
	k8sClient := client.NewK8sClient()

	var router *coreV1.Service
	yamls, err := LoadYamls()

	if err != nil {
		glog.Error(err)
	}

	for _, f := range yamls {
		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, _, err := decode([]byte(f), nil, nil)

		if err != nil {
			glog.Error(err)
			return router, err
		}

		options := metaV1.DeleteOptions{}

		switch obj.(type) {
		case *v1beta1.Role:
			role := obj.(*v1beta1.Role)
			err = k8sClient.RbacV1beta1().Roles(namespace).Delete(role.Name, &options)
			if err != nil {
				glog.Error(err)
			}
		case *v1beta1.ClusterRoleBinding:
			clusterRoleBinding := obj.(*v1beta1.ClusterRoleBinding)
			err = k8sClient.RbacV1beta1().ClusterRoleBindings().Delete(clusterRoleBinding.Name, &options)
			if err != nil {
				glog.Error(err)
			}
		case *v1beta1.RoleBinding:
			roleBinding := obj.(*v1beta1.RoleBinding)
			err = k8sClient.RbacV1beta1().RoleBindings(namespace).Delete(roleBinding.Name, &options)
			if err != nil {
				glog.Error(err)
			}
		case *coreV1.ServiceAccount:
			sa := obj.(*coreV1.ServiceAccount)
			err = k8sClient.CoreV1().ServiceAccounts(namespace).Delete(sa.Name, &options)
			if err != nil {
				glog.Error(err)
			}
		case *coreV1.Service:
			service := obj.(*coreV1.Service)

			err = k8sClient.CoreV1().Services(namespace).Delete(service.Name, &options)
			if err != nil {
				glog.Error(err)
			}

			if IsRouterService(service.Name) {
				router = service
			}

		case *extensionsV1beta1.Deployment:
			deployment := obj.(*extensionsV1beta1.Deployment)
			err = k8sClient.ExtensionsV1beta1().Deployments(namespace).Delete(deployment.Name, &options)
			if err != nil {
				glog.Error(err)
			}
		default:
			//glog.Info("Default resource")
		}
	}

	return router, nil

}

// Update Ingress Controller Service, change type from NodePort to Loadbalancer or vice versa.
func UpdateRouter(namespace string, routerType coreV1.ServiceType, annotations map[string]string) (*coreV1.Service, error) {
	k8sClient := client.NewK8sClient()

	var router *coreV1.Service

	router, err := GetRouter(namespace)

	if err != nil {
		glog.Error(err)
		return router, nil
	}

	router.Spec.Type = routerType
	router.SetAnnotations(annotations)

	router, err = k8sClient.CoreV1().Services(namespace).Update(router)

	if err != nil {
		glog.Error(err)
		return router, err
	}

	return router, nil

}
