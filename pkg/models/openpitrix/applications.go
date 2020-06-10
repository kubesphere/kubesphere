/*
Copyright 2020 The KubeSphere Authors.
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
	"encoding/json"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"openpitrix.io/openpitrix/pkg/pb"
	"openpitrix.io/openpitrix/pkg/util/pbutil"
	"strings"
)

type ApplicationInterface interface {
	ListApplications(conditions *params.Conditions, limit, offset int, orderBy string, reverse bool) (*models.PageableResponse, error)
	DescribeApplication(namespace, applicationId, clusterName string) (*Application, error)
	CreateApplication(clusterName, namespace string, request CreateClusterRequest) error
	ModifyApplication(request ModifyClusterAttributesRequest) error
	DeleteApplication(id string) error
	UpgradeApplication(request UpgradeClusterRequest) error
}

type applicationOperator struct {
	informers informers.SharedInformerFactory
	opClient  openpitrix.Client
}

func newApplicationOperator(informers informers.SharedInformerFactory, opClient openpitrix.Client) ApplicationInterface {
	return &applicationOperator{informers: informers, opClient: opClient}
}

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

type resourceInfo struct {
	Deployments  []appsv1.Deployment  `json:"deployments,omitempty" description:"deployment list"`
	Statefulsets []appsv1.StatefulSet `json:"statefulsets,omitempty" description:"statefulset list"`
	Daemonsets   []appsv1.DaemonSet   `json:"daemonsets,omitempty" description:"daemonset list"`
	Services     []v1.Service         `json:"services,omitempty" description:"application services"`
	Ingresses    []v1beta1.Ingress    `json:"ingresses,omitempty" description:"application ingresses"`
}

func (c *applicationOperator) ListApplications(conditions *params.Conditions, limit, offset int, orderBy string, reverse bool) (*models.PageableResponse, error) {
	describeClustersRequest := &pb.DescribeClustersRequest{
		Limit:  uint32(limit),
		Offset: uint32(offset)}
	if keyword := conditions.Match[Keyword]; keyword != "" {
		describeClustersRequest.SearchWord = &wrappers.StringValue{Value: keyword}
	}
	if runtimeId := conditions.Match[RuntimeId]; runtimeId != "" {
		describeClustersRequest.RuntimeId = []string{runtimeId}
	}
	if appId := conditions.Match[AppId]; appId != "" {
		describeClustersRequest.AppId = []string{appId}
	}
	if versionId := conditions.Match[VersionId]; versionId != "" {
		describeClustersRequest.VersionId = []string{versionId}
	}
	if status := conditions.Match[Status]; status != "" {
		describeClustersRequest.Status = strings.Split(status, "|")
	}
	if zone := conditions.Match[Zone]; zone != "" {
		describeClustersRequest.Zone = []string{zone}
	}
	if orderBy != "" {
		describeClustersRequest.SortKey = &wrappers.StringValue{Value: orderBy}
	}
	describeClustersRequest.Reverse = &wrappers.BoolValue{Value: reverse}
	resp, err := c.opClient.DescribeClusters(openpitrix.SystemContext(), describeClustersRequest)
	if err != nil {
		klog.Errorln(err)
		return nil, err
	}
	result := models.PageableResponse{TotalCount: int(resp.TotalCount)}
	result.Items = make([]interface{}, 0)
	for _, cluster := range resp.ClusterSet {
		app, err := c.describeApplication(cluster)
		if err != nil {
			klog.Errorln(err)
			return nil, err
		}
		result.Items = append(result.Items, app)
	}

	return &result, nil
}

func (c *applicationOperator) describeApplication(cluster *pb.Cluster) (*Application, error) {
	var app Application
	app.Name = cluster.Name.Value
	app.Cluster = convertCluster(cluster)
	versionInfo, err := c.opClient.DescribeAppVersions(openpitrix.SystemContext(), &pb.DescribeAppVersionsRequest{VersionId: []string{cluster.GetVersionId().GetValue()}})
	if err != nil {
		klog.Errorln(err)
		return nil, err
	}
	if len(versionInfo.AppVersionSet) > 0 {
		app.Version = convertAppVersion(versionInfo.AppVersionSet[0])
	}
	appInfo, err := c.opClient.DescribeApps(openpitrix.SystemContext(), &pb.DescribeAppsRequest{AppId: []string{cluster.GetAppId().GetValue()}, Limit: 1})
	if err != nil {
		klog.Errorln(err)
		return nil, err
	}
	if len(appInfo.AppSet) > 0 {
		app.App = convertApp(appInfo.GetAppSet()[0])
	}
	return &app, nil
}

func (c *applicationOperator) DescribeApplication(namespace string, applicationId string, clusterName string) (*Application, error) {
	describeClusterRequest := &pb.DescribeClustersRequest{
		ClusterId:  []string{applicationId},
		RuntimeId:  []string{clusterName},
		Zone:       []string{namespace},
		WithDetail: pbutil.ToProtoBool(true),
		Limit:      1,
	}
	clusters, err := c.opClient.DescribeClusters(openpitrix.SystemContext(), describeClusterRequest)

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
	app, err := c.describeApplication(cluster)
	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	resource := new(resourceInfo)
	workloads := cluster.AdditionalInfo.GetValue()
	if workloads == "" {
		err := status.New(codes.NotFound, "cannot get workload").Err()
		klog.Errorln(err)
		return nil, err
	}
	err = json.Unmarshal([]byte(workloads), resource)
	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	app.WorkLoads = &workLoads{
		Deployments:  resource.Deployments,
		Statefulsets: resource.Statefulsets,
		Daemonsets:   resource.Daemonsets,
	}
	app.Services = resource.Services
	app.Ingresses = resource.Ingresses
	return app, nil
}

func (c *applicationOperator) getWorkLoads(namespace string, clusterRoles []*pb.ClusterRole) (*workLoads, error) {

	var works workLoads
	for _, clusterRole := range clusterRoles {
		workLoadName := clusterRole.Role.Value
		if len(workLoadName) > 0 {
			if strings.HasSuffix(workLoadName, openpitrix.DeploySuffix) {
				name := strings.Split(workLoadName, openpitrix.DeploySuffix)[0]

				item, err := c.informers.Apps().V1().Deployments().Lister().Deployments(namespace).Get(name)

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
				item, err := c.informers.Apps().V1().DaemonSets().Lister().DaemonSets(namespace).Get(name)
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
				item, err := c.informers.Apps().V1().StatefulSets().Lister().StatefulSets(namespace).Get(name)
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

func (c *applicationOperator) getLabels(namespace string, workloads *workLoads) *[]map[string]string {

	var workloadLabels []map[string]string
	if workloads == nil {
		return nil
	}

	for _, workload := range workloads.Deployments {
		deploy, err := c.informers.Apps().V1().Deployments().Lister().Deployments(namespace).Get(workload.Name)
		if errors.IsNotFound(err) {
			continue
		}
		workloadLabels = append(workloadLabels, deploy.Labels)
	}

	for _, workload := range workloads.Daemonsets {
		daemonset, err := c.informers.Apps().V1().DaemonSets().Lister().DaemonSets(namespace).Get(workload.Name)
		if errors.IsNotFound(err) {
			continue
		}
		workloadLabels = append(workloadLabels, daemonset.Labels)
	}

	for _, workload := range workloads.Statefulsets {
		statefulset, err := c.informers.Apps().V1().StatefulSets().Lister().StatefulSets(namespace).Get(workload.Name)
		if errors.IsNotFound(err) {
			continue
		}
		workloadLabels = append(workloadLabels, statefulset.Labels)
	}

	return &workloadLabels
}

func (c *applicationOperator) isExist(svcs []v1.Service, svc *v1.Service) bool {
	for _, item := range svcs {
		if item.Name == svc.Name && item.Namespace == svc.Namespace {
			return true
		}
	}
	return false
}

func (c *applicationOperator) getSvcs(namespace string, workLoadLabels *[]map[string]string) []v1.Service {
	if len(*workLoadLabels) == 0 {
		return nil
	}
	var services []v1.Service
	for _, label := range *workLoadLabels {
		labelSelector := labels.Set(label).AsSelector()
		svcs, err := c.informers.Core().V1().Services().Lister().Services(namespace).List(labelSelector)
		if err != nil {
			klog.Errorf("get app's svc failed, reason: %v", err)
		}
		for _, item := range svcs {
			if !c.isExist(services, item) {
				services = append(services, *item)
			}
		}
	}

	return services
}

func (c *applicationOperator) getIng(namespace string, services []v1.Service) []v1beta1.Ingress {
	if services == nil {
		return nil
	}

	var ings []v1beta1.Ingress
	for _, svc := range services {
		ingresses, err := c.informers.Extensions().V1beta1().Ingresses().Lister().Ingresses(namespace).List(labels.Everything())
		if err != nil {
			klog.Error(err)
			return ings
		}

		for _, ingress := range ingresses {
			if ingress.Spec.Backend.ServiceName != svc.Name {
				continue
			}

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

func (c *applicationOperator) CreateApplication(clusterName, namespace string, request CreateClusterRequest) error {
	_, err := c.opClient.CreateCluster(openpitrix.ContextWithUsername(request.Username), &pb.CreateClusterRequest{
		AppId:     &wrappers.StringValue{Value: request.AppId},
		VersionId: &wrappers.StringValue{Value: request.VersionId},
		RuntimeId: &wrappers.StringValue{Value: clusterName},
		Conf:      &wrappers.StringValue{Value: request.Conf},
		Zone:      &wrappers.StringValue{Value: namespace},
	})

	if err != nil {
		klog.Errorln(err)
		return err
	}

	return nil
}

func (c *applicationOperator) ModifyApplication(request ModifyClusterAttributesRequest) error {

	modifyClusterAttributesRequest := &pb.ModifyClusterAttributesRequest{ClusterId: &wrappers.StringValue{Value: request.ClusterID}}
	if request.Name != nil {
		modifyClusterAttributesRequest.Name = &wrappers.StringValue{Value: *request.Name}
	}
	if request.Description != nil {
		modifyClusterAttributesRequest.Description = &wrappers.StringValue{Value: *request.Description}
	}

	_, err := c.opClient.ModifyClusterAttributes(openpitrix.SystemContext(), modifyClusterAttributesRequest)

	if err != nil {
		klog.Errorln(err)
		return err
	}

	return nil
}

func (c *applicationOperator) DeleteApplication(applicationId string) error {
	_, err := c.opClient.DeleteClusters(openpitrix.SystemContext(), &pb.DeleteClustersRequest{ClusterId: []string{applicationId}, Force: &wrappers.BoolValue{Value: true}})

	if err != nil {
		klog.Errorln(err)
		return err
	}

	return nil
}

func (c *applicationOperator) UpgradeApplication(request UpgradeClusterRequest) error {
	_, err := c.opClient.UpgradeCluster(openpitrix.ContextWithUsername(request.Username), &pb.UpgradeClusterRequest{
		ClusterId: &wrappers.StringValue{Value: request.ClusterId},
		VersionId: &wrappers.StringValue{Value: request.VersionId},
	})

	if err != nil {
		klog.Errorln(err)
		return err
	}

	return nil
}
