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
package openpitrix

import (
	"fmt"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	cs "kubesphere.io/kubesphere/pkg/simple/client"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"openpitrix.io/openpitrix/pkg/pb"
	"strings"
)

type Application struct {
	Name      string            `json:"name" description:"application name"`
	Cluster   *Cluster          `json:"cluster,omitempty" description:"application cluster info"`
	Version   *AppVersion       `json:"version,omitempty" description:"application template version info"`
	App       *App              `json:"app,omitempty" description:"application template info"`
	WorkLoads *workLoads        `json:"workloads,omitempty" description:"application workloads"`
	Services  []v1.Service      `json:"services,omitempty" description:"application services"`
	Ingresses []v1beta1.Ingress `json:"ingresses,omitempty" description:"application ingresses"`
}

type workLoads struct {
	Deployments  []appsv1.Deployment  `json:"deployments,omitempty" description:"deployment list"`
	Statefulsets []appsv1.StatefulSet `json:"statefulsets,omitempty" description:"statefulset list"`
	Daemonsets   []appsv1.DaemonSet   `json:"daemonsets,omitempty" description:"daemonset list"`
}

func ListApplications(conditions *params.Conditions, limit, offset int, orderBy string, reverse bool) (*models.PageableResponse, error) {
	client, err := cs.ClientSets().OpenPitrix()
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	describeClustersRequest := &pb.DescribeClustersRequest{
		Limit:  uint32(limit),
		Offset: uint32(offset)}
	if keyword := conditions.Match["keyword"]; keyword != "" {
		describeClustersRequest.SearchWord = &wrappers.StringValue{Value: keyword}
	}
	if runtimeId := conditions.Match["runtime_id"]; runtimeId != "" {
		describeClustersRequest.RuntimeId = []string{runtimeId}
	}
	if appId := conditions.Match["app_id"]; appId != "" {
		describeClustersRequest.AppId = []string{appId}
	}
	if versionId := conditions.Match["version_id"]; versionId != "" {
		describeClustersRequest.VersionId = []string{versionId}
	}
	if status := conditions.Match["status"]; status != "" {
		describeClustersRequest.Status = strings.Split(status, "|")
	}
	if orderBy != "" {
		describeClustersRequest.SortKey = &wrappers.StringValue{Value: orderBy}
	}
	describeClustersRequest.Reverse = &wrappers.BoolValue{Value: !reverse}
	resp, err := client.Cluster().DescribeClusters(openpitrix.SystemContext(), describeClustersRequest)
	if err != nil {
		klog.Errorln(err)
		return nil, err
	}
	result := models.PageableResponse{TotalCount: int(resp.TotalCount)}
	result.Items = make([]interface{}, 0)
	for _, cluster := range resp.ClusterSet {
		app, err := describeApplication(cluster)
		if err != nil {
			klog.Errorln(err)
			return nil, err
		}
		result.Items = append(result.Items, app)
	}

	return &result, nil
}

func describeApplication(cluster *pb.Cluster) (*Application, error) {
	op, err := cs.ClientSets().OpenPitrix()
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	var app Application
	app.Name = cluster.Name.Value
	app.Cluster = convertCluster(cluster)
	versionInfo, err := op.App().DescribeAppVersions(openpitrix.SystemContext(), &pb.DescribeAppVersionsRequest{VersionId: []string{cluster.GetVersionId().GetValue()}})
	if err != nil {
		klog.Errorln(err)
		return nil, err
	}
	if len(versionInfo.AppVersionSet) > 0 {
		app.Version = convertAppVersion(versionInfo.AppVersionSet[0])
	}
	appInfo, err := op.App().DescribeApps(openpitrix.SystemContext(), &pb.DescribeAppsRequest{AppId: []string{cluster.GetAppId().GetValue()}, Limit: 1})
	if err != nil {
		klog.Errorln(err)
		return nil, err
	}
	if len(appInfo.AppSet) > 0 {
		app.App = convertApp(appInfo.GetAppSet()[0])
	}
	return &app, nil
}

func DescribeApplication(namespace string, clusterId string) (*Application, error) {

	client, err := cs.ClientSets().OpenPitrix()
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	clusters, err := client.Cluster().DescribeClusters(openpitrix.SystemContext(), &pb.DescribeClustersRequest{ClusterId: []string{clusterId}, Limit: 1})

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	var cluster *pb.Cluster
	if len(clusters.ClusterSet) > 0 {
		cluster = clusters.GetClusterSet()[0]
	} else {
		err := status.New(codes.NotFound, "resource not found").Err()
		klog.Errorln(err)
		return nil, err
	}
	app, err := describeApplication(cluster)
	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	workloads, err := getWorkLoads(namespace, cluster.ClusterRoleSet)

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}
	app.WorkLoads = workloads
	workloadLabels := getLabels(namespace, app.WorkLoads)
	app.Services = getSvcs(namespace, workloadLabels)
	app.Ingresses = getIng(namespace, app.Services)
	return app, nil
}

func getWorkLoads(namespace string, clusterRoles []*pb.ClusterRole) (*workLoads, error) {

	var works workLoads
	for _, clusterRole := range clusterRoles {
		workLoadName := clusterRole.Role.Value
		if len(workLoadName) > 0 {
			if strings.HasSuffix(workLoadName, openpitrix.DeploySuffix) {
				name := strings.Split(workLoadName, openpitrix.DeploySuffix)[0]

				item, err := informers.SharedInformerFactory().Apps().V1().Deployments().Lister().Deployments(namespace).Get(name)

				if err != nil {
					// app not ready
					if errors.IsNotFound(err) {
						continue
					}
					klog.Errorln(err)
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
					klog.Errorln(err)
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
					klog.Errorln(err)
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
	k8sClient := cs.ClientSets().K8s().Kubernetes()

	var workloadLabels []map[string]string
	if workloads == nil {
		return nil
	}

	for _, workload := range workloads.Deployments {
		deploy, err := k8sClient.AppsV1().Deployments(namespace).Get(workload.Name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			continue
		}
		workloadLabels = append(workloadLabels, deploy.Labels)
	}

	for _, workload := range workloads.Daemonsets {
		daemonset, err := k8sClient.AppsV1().DaemonSets(namespace).Get(workload.Name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			continue
		}
		workloadLabels = append(workloadLabels, daemonset.Labels)
	}

	for _, workload := range workloads.Statefulsets {
		statefulset, err := k8sClient.AppsV1().StatefulSets(namespace).Get(workload.Name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			continue
		}
		workloadLabels = append(workloadLabels, statefulset.Labels)
	}

	return &workloadLabels
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
	k8sClient := cs.ClientSets().K8s().Kubernetes()
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
			klog.Errorln(err)
			return nil
		}

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

func CreateApplication(namespace string, request CreateClusterRequest) error {
	ns, err := informers.SharedInformerFactory().Core().V1().Namespaces().Lister().Get(namespace)
	if err != nil {
		klog.Error(err)
		return err
	}

	if runtimeId := ns.Annotations[constants.OpenPitrixRuntimeAnnotationKey]; runtimeId != "" {
		request.RuntimeId = runtimeId
	} else {
		return fmt.Errorf("runtime not init: namespace %s", namespace)
	}

	client, err := cs.ClientSets().OpenPitrix()

	if err != nil {
		klog.Error(err)
		return err
	}

	_, err = client.Cluster().CreateCluster(openpitrix.ContextWithUsername(request.Username), &pb.CreateClusterRequest{
		AppId:     &wrappers.StringValue{Value: request.AppId},
		VersionId: &wrappers.StringValue{Value: request.VersionId},
		RuntimeId: &wrappers.StringValue{Value: request.RuntimeId},
		Conf:      &wrappers.StringValue{Value: request.Conf},
	})

	if err != nil {
		klog.Errorln(err)
		return err
	}

	return nil
}

func PatchApplication(request *ModifyClusterAttributesRequest) error {
	op, err := cs.ClientSets().OpenPitrix()

	if err != nil {
		klog.Error(err)
		return err
	}

	modifyClusterAttributesRequest := &pb.ModifyClusterAttributesRequest{ClusterId: &wrappers.StringValue{Value: request.ClusterID}}
	if request.Name != nil {
		modifyClusterAttributesRequest.Name = &wrappers.StringValue{Value: *request.Name}
	}
	if request.Description != nil {
		modifyClusterAttributesRequest.Description = &wrappers.StringValue{Value: *request.Description}
	}

	_, err = op.Cluster().ModifyClusterAttributes(openpitrix.SystemContext(), modifyClusterAttributesRequest)

	if err != nil {
		klog.Errorln(err)
		return err
	}
	return nil
}

func DeleteApplication(clusterId string) error {
	client, err := cs.ClientSets().OpenPitrix()

	if err != nil {
		klog.Error(err)
		return err
	}

	_, err = client.Cluster().DeleteClusters(openpitrix.SystemContext(), &pb.DeleteClustersRequest{ClusterId: []string{clusterId}, Force: &wrappers.BoolValue{Value: true}})

	if err != nil {
		klog.Errorln(err)
		return err
	}

	return nil
}
