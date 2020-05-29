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

package resource

import (
	"fmt"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/application"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/clusterrole"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/configmap"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/cronjob"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/daemonset"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/deployment"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/hpa"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/ingress"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/job"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/namespace"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/node"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/persistentvolumeclaim"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/pod"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/role"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/s2buildertemplate"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/s2ibuilder"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/s2irun"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/secret"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/service"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/statefulset"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/storageclass"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/workspace"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
)

type ResourceGetter struct {
	resourcesGetters map[string]v1alpha2.Interface
}

func (r ResourceGetter) Add(resource string, getter v1alpha2.Interface) {
	if r.resourcesGetters == nil {
		r.resourcesGetters = make(map[string]v1alpha2.Interface)
	}
	r.resourcesGetters[resource] = getter
}

func NewResourceGetter(factory informers.InformerFactory) *ResourceGetter {
	resourceGetters := make(map[string]v1alpha2.Interface)

	resourceGetters[v1alpha2.ConfigMaps] = configmap.NewConfigmapSearcher(factory.KubernetesSharedInformerFactory())
	resourceGetters[v1alpha2.CronJobs] = cronjob.NewCronJobSearcher(factory.KubernetesSharedInformerFactory())
	resourceGetters[v1alpha2.DaemonSets] = daemonset.NewDaemonSetSearcher(factory.KubernetesSharedInformerFactory())
	resourceGetters[v1alpha2.Deployments] = deployment.NewDeploymentSetSearcher(factory.KubernetesSharedInformerFactory())
	resourceGetters[v1alpha2.Ingresses] = ingress.NewIngressSearcher(factory.KubernetesSharedInformerFactory())
	resourceGetters[v1alpha2.Jobs] = job.NewJobSearcher(factory.KubernetesSharedInformerFactory())
	resourceGetters[v1alpha2.PersistentVolumeClaims] = persistentvolumeclaim.NewPersistentVolumeClaimSearcher(factory.KubernetesSharedInformerFactory(), factory.SnapshotSharedInformerFactory())
	resourceGetters[v1alpha2.Secrets] = secret.NewSecretSearcher(factory.KubernetesSharedInformerFactory())
	resourceGetters[v1alpha2.Services] = service.NewServiceSearcher(factory.KubernetesSharedInformerFactory())
	resourceGetters[v1alpha2.StatefulSets] = statefulset.NewStatefulSetSearcher(factory.KubernetesSharedInformerFactory())
	resourceGetters[v1alpha2.Pods] = pod.NewPodSearcher(factory.KubernetesSharedInformerFactory())
	resourceGetters[v1alpha2.Roles] = role.NewRoleSearcher(factory.KubernetesSharedInformerFactory())

	resourceGetters[v1alpha2.Nodes] = node.NewNodeSearcher(factory.KubernetesSharedInformerFactory())
	resourceGetters[v1alpha2.Namespaces] = namespace.NewNamespaceSearcher(factory.KubernetesSharedInformerFactory())
	resourceGetters[v1alpha2.ClusterRoles] = clusterrole.NewClusterRoleSearcher(factory.KubernetesSharedInformerFactory())
	resourceGetters[v1alpha2.StorageClasses] = storageclass.NewStorageClassesSearcher(factory.KubernetesSharedInformerFactory(), factory.SnapshotSharedInformerFactory())
	resourceGetters[v1alpha2.HorizontalPodAutoscalers] = hpa.NewHpaSearcher(factory.KubernetesSharedInformerFactory())
	resourceGetters[v1alpha2.S2iBuilders] = s2ibuilder.NewS2iBuilderSearcher(factory.KubeSphereSharedInformerFactory())
	resourceGetters[v1alpha2.S2iRuns] = s2irun.NewS2iRunSearcher(factory.KubeSphereSharedInformerFactory())
	resourceGetters[v1alpha2.S2iBuilderTemplates] = s2buildertemplate.NewS2iBuidlerTemplateSearcher(factory.KubeSphereSharedInformerFactory())
	resourceGetters[v1alpha2.Workspaces] = workspace.NewWorkspaceSearcher(factory.KubeSphereSharedInformerFactory())
	resourceGetters[v1alpha2.Applications] = application.NewApplicationSearcher(factory.ApplicationSharedInformerFactory())

	return &ResourceGetter{resourcesGetters: resourceGetters}

}

var (
	//injector         = v1alpha2.extraAnnotationInjector{}
	clusterResources = []string{v1alpha2.Nodes, v1alpha2.Workspaces, v1alpha2.Namespaces, v1alpha2.ClusterRoles, v1alpha2.StorageClasses, v1alpha2.S2iBuilderTemplates}
)

func (r *ResourceGetter) GetResource(namespace, resource, name string) (interface{}, error) {
	if searcher, ok := r.resourcesGetters[resource]; ok {
		resource, err := searcher.Get(namespace, name)
		if err != nil {
			klog.Errorf("resource %s.%s.%s not found: %s", namespace, resource, name, err)
			return nil, err
		}
		return resource, nil
	}
	return nil, fmt.Errorf("resource %s.%s.%s not found", namespace, resource, name)
}

func (r *ResourceGetter) ListResources(namespace, resource string, conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	items := make([]interface{}, 0)
	var err error
	var result []interface{}

	// none namespace resource
	if namespace != "" && sliceutil.HasString(clusterResources, resource) {
		err = fmt.Errorf("resource %s is not supported", resource)
		klog.Errorln(err)
		return nil, err
	}

	if searcher, ok := r.resourcesGetters[resource]; ok {
		result, err = searcher.Search(namespace, conditions, orderBy, reverse)
	} else {
		err = fmt.Errorf("resource %s is not supported", resource)
		klog.Errorln(err)
		return nil, err
	}

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	if limit == -1 || limit+offset > len(result) {
		limit = len(result) - offset
	}

	items = result[offset : offset+limit]

	return &models.PageableResponse{TotalCount: len(result), Items: items}, nil
}
