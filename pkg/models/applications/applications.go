/*

 Copyright 2019 The KubeSphere Authors.

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
package applications

import (
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/resources"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"strings"
	"time"
)

type Application struct {
	Name        string            `json:"name" description:"application name"`
	RepoName    string            `json:"repoName" description:"repo name"`
	Runtime     string            `json:"namespace" description:"namespace name"`
	RuntimeId   string            `json:"runtime_id" description:"It is the application runtime in OpenPitrix and represents k8s namespace by annotating the namespace"`
	Version     string            `json:"version" description:"application version"`
	VersionId   string            `json:"version_id" description:"application version id"`
	Status      string            `json:"status" description:"application status"`
	UpdateTime  time.Time         `json:"updateTime" description:"the last time this application was updated"`
	CreateTime  time.Time         `json:"createTime" description:"application creation time"`
	App         string            `json:"app" description:"application template name"`
	AppId       string            `json:"app_id" description:"application template id"`
	Description string            `json:"description,omitempty" description:"application description"`
	WorkLoads   *workLoads        `json:"workloads,omitempty" description:"application workloads"`
	Services    []v1.Service      `json:"services,omitempty" description:"application services"`
	Ingresses   []v1beta1.Ingress `json:"ingresses,omitempty" description:"application ingresses"`
	ClusterID   string            `json:"cluster_id" description:"application id"`
}

type workLoads struct {
	Deployments  []appsv1.Deployment  `json:"deployments,omitempty" description:"deployment list"`
	Statefulsets []appsv1.StatefulSet `json:"statefulsets,omitempty" description:"statefulset list"`
	Daemonsets   []appsv1.DaemonSet   `json:"daemonsets,omitempty" description:"daemonset list"`
}

func ListApplication(runtimeId string, conditions *params.Conditions, limit, offset int) (*models.PageableResponse, error) {
	openPitrixClient, err := client.ClientSets().OpenPitrix()
	if err != nil {
		return nil, err
	}
	clusterList, err := openPitrixClient.ListClusters(runtimeId, conditions.Match["keyword"], conditions.Match["status"], limit, offset)
	if err != nil {
		return nil, err
	}
	result := models.PageableResponse{TotalCount: clusterList.Total}
	result.Items = make([]interface{}, 0)
	for _, item := range clusterList.Clusters {
		var app Application

		app.Name = item.Name
		app.ClusterID = item.ClusterID
		app.UpdateTime = item.UpdateTime
		app.Status = item.Status
		versionInfo, _ := openPitrixClient.GetVersion(item.VersionID)
		app.Version = versionInfo
		app.VersionId = item.VersionID
		runtimeInfo, _ := openPitrixClient.GetRuntime(item.RunTimeId)
		app.Runtime = runtimeInfo
		app.RuntimeId = item.RunTimeId
		appInfo, _, appId, _ := openPitrixClient.GetAppInfo(item.AppID)
		app.App = appInfo
		app.AppId = appId
		app.Description = item.Description

		result.Items = append(result.Items, app)
	}

	return &result, nil
}

func GetApp(clusterId string) (*Application, error) {
	openPitrixClient, err := client.ClientSets().OpenPitrix()
	if err != nil {
		return nil, err
	}
	item, err := openPitrixClient.GetCluster(clusterId)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	var app Application

	app.Name = item.Name
	app.ClusterID = item.ClusterID
	app.UpdateTime = item.UpdateTime
	app.CreateTime = item.CreateTime
	app.Status = item.Status
	versionInfo, _ := openPitrixClient.GetVersion(item.VersionID)
	app.Version = versionInfo
	app.VersionId = item.VersionID

	runtimeInfo, _ := openPitrixClient.GetRuntime(item.RunTimeId)
	app.Runtime = runtimeInfo
	app.RuntimeId = item.RunTimeId
	appInfo, repoId, appId, _ := openPitrixClient.GetAppInfo(item.AppID)
	app.App = appInfo
	app.AppId = appId
	app.Description = item.Description

	app.RepoName, _ = openPitrixClient.GetRepo(repoId)

	workloads, err := getWorkLoads(app.Runtime, item.ClusterRoleSets)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	app.WorkLoads = workloads
	workloadLabels := getLabels(app.Runtime, app.WorkLoads)
	app.Services = getSvcs(app.Runtime, workloadLabels)
	app.Ingresses = getIng(app.Runtime, app.Services)

	return &app, nil
}

func getWorkLoads(namespace string, clusterRoles []openpitrix.ClusterRole) (*workLoads, error) {

	var works workLoads
	for _, clusterRole := range clusterRoles {
		workLoadName := clusterRole.Role
		if len(workLoadName) > 0 {
			if strings.HasSuffix(workLoadName, openpitrix.DeploySuffix) {
				name := strings.Split(workLoadName, openpitrix.DeploySuffix)[0]

				item, err := informers.SharedInformerFactory().Apps().V1().Deployments().Lister().Deployments(namespace).Get(name)

				if err != nil {
					// app not ready
					if errors.IsNotFound(err) {
						continue
					}
					klog.Error(err)
					return nil, err
				}

				works.Deployments = append(works.Deployments, *item)
				continue
			}

			if strings.HasSuffix(workLoadName, openpitrix.DaemonSuffix) {
				name := strings.Split(workLoadName, openpitrix.DaemonSuffix)[0]
				item, err := informers.SharedInformerFactory().Apps().V1().DaemonSets().Lister().DaemonSets(namespace).Get(name)
				if err != nil {
					// app not ready
					if errors.IsNotFound(err) {
						continue
					}
					klog.Error(err)
					return nil, err
				}
				works.Daemonsets = append(works.Daemonsets, *item)
				continue
			}

			if strings.HasSuffix(workLoadName, openpitrix.StateSuffix) {
				name := strings.Split(workLoadName, openpitrix.StateSuffix)[0]
				item, err := informers.SharedInformerFactory().Apps().V1().StatefulSets().Lister().StatefulSets(namespace).Get(name)
				if err != nil {
					// app not ready
					if errors.IsNotFound(err) {
						continue
					}
					klog.Error(err)
					return nil, err
				}
				works.Statefulsets = append(works.Statefulsets, *item)
				continue
			}
		}
	}
	return &works, nil
}

func getLabels(namespace string, workloads *workLoads) *[]map[string]string {
	k8sClient := client.ClientSets().K8s().Kubernetes()

	var workloadLables []map[string]string
	if workloads == nil {
		return nil
	}

	for _, workload := range workloads.Deployments {
		deploy, err := k8sClient.AppsV1().Deployments(namespace).Get(workload.Name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			continue
		}
		workloadLables = append(workloadLables, deploy.Labels)
	}

	for _, workload := range workloads.Daemonsets {
		daemonset, err := k8sClient.AppsV1().DaemonSets(namespace).Get(workload.Name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			continue
		}
		workloadLables = append(workloadLables, daemonset.Labels)
	}

	for _, workload := range workloads.Statefulsets {
		statefulset, err := k8sClient.AppsV1().StatefulSets(namespace).Get(workload.Name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			continue
		}
		workloadLables = append(workloadLables, statefulset.Labels)
	}

	return &workloadLables
}

func isExist(svcs []v1.Service, svc v1.Service) bool {
	for _, item := range svcs {
		if item.Name == svc.Name && item.Namespace == svc.Namespace {
			return true
		}
	}
	return false
}

func getSvcs(namespace string, workLoadLabels *[]map[string]string) []v1.Service {
	if len(*workLoadLabels) == 0 {
		return nil
	}
	k8sClient := client.ClientSets().K8s().Kubernetes()
	var services []v1.Service
	for _, label := range *workLoadLabels {
		labelSelector := labels.Set(label).AsSelector().String()
		svcs, err := k8sClient.CoreV1().Services(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
		if err != nil {
			klog.Errorf("get app's svc failed, reason: %v", err)
		}
		for _, item := range svcs.Items {
			if !isExist(services, item) {
				services = append(services, item)
			}
		}
	}

	return services
}

func getIng(namespace string, services []v1.Service) []v1beta1.Ingress {
	if services == nil {
		return nil
	}

	var ings []v1beta1.Ingress
	for _, svc := range services {
		result, err := resources.ListResources(namespace, "ingress", &params.Conditions{Fuzzy: map[string]string{"serviceName": svc.Name}}, "", false, -1, 0)
		if err != nil {
			klog.Error(err)
			return nil
		}

		klog.Error(result)
		for _, i := range result.Items {
			ingress := i.(*v1beta1.Ingress)

			exist := false
			var tmpRules []v1beta1.IngressRule
			for _, rule := range ingress.Spec.Rules {
				for _, p := range rule.HTTP.Paths {
					if p.Backend.ServiceName == svc.Name {
						exist = true
						tmpRules = append(tmpRules, rule)
					}
				}

			}

			if exist {
				ing := v1beta1.Ingress{}
				ing.Name = ingress.Name
				ing.Spec.Rules = tmpRules
				ings = append(ings, ing)
			}
		}
	}

	return ings
}

func DeployApplication(namespace string, app openpitrix.CreateClusterRequest) error {
	openPitrixClient, err := client.ClientSets().OpenPitrix()
	if err != nil {
		return err
	}

	ns, err := informers.SharedInformerFactory().Core().V1().Namespaces().Lister().Get(namespace)
	if err != nil {
		klog.Errorf("deploy application failed: %+v", err)
		return err
	}

	if runtimeId := ns.Annotations[constants.OpenPitrixRuntimeAnnotationKey]; runtimeId != "" {
		app.RuntimeId = runtimeId
	} else {
		return fmt.Errorf("runtime not init: namespace %s", namespace)
	}
	return openPitrixClient.CreateCluster(app)
}

func DeleteApplication(clusterId string) error {
	openPitrixClient, err := client.ClientSets().OpenPitrix()
	if err != nil {
		return err
	}

	return openPitrixClient.DeleteCluster(openpitrix.DeleteClusterRequest{ClusterId: []string{clusterId}})
}
