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

package resource

import (
	"errors"

	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	monitoringdashboardv1alpha2 "kubesphere.io/monitoring-dashboard/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	_ "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/application"
	_ "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/cluster"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/clusterrole"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/clusterrolebinding"
	_ "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/customresourcedefinition"
	_ "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/daemonset"
	_ "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/deployment"
	_ "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/federatedapplication"
	_ "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/federateddeployment"
	_ "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/federatedpersistentvolumeclaim"
	_ "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/federatedsecret"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/globalrole"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/globalrolebinding"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/groupbinding"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/ippool"
	_ "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/job"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/loginrecord"
	_ "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/namespace"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/node"
	_ "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/notification"
	_ "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/persistentvolume"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/persistentvolumeclaim"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/pod"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/role"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/rolebinding"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/user"
	_ "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/volumesnapshot"
	_ "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/volumesnapshotclass"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/workspacerole"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/workspacerolebinding"

	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	devopsv1alpha3 "kubesphere.io/api/devops/v1alpha3"
	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"
	networkv1alpha1 "kubesphere.io/api/network/v1alpha1"
	notificationv2beta1 "kubesphere.io/api/notification/v2beta1"
	tenantv1alpha1 "kubesphere.io/api/tenant/v1alpha1"
	tenantv1alpha2 "kubesphere.io/api/tenant/v1alpha2"
	typesv1beta1 "kubesphere.io/api/types/v1beta1"

	"kubesphere.io/kubesphere/pkg/models/crds"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
)

var ErrResourceNotSupported = errors.New("resource is not supported")

var resourceToGVK map[string]schema.GroupVersionKind = map[string]schema.GroupVersionKind{
	// Kubernetes
	"deployments":               {Group: "apps", Version: "v1", Kind: "Deployment"},
	"daemonsets":                {Group: "apps", Version: "v1", Kind: "DaemonSet"},
	"statefulsets":              {Group: "apps", Version: "v1", Kind: "StatefulSet"},
	"services":                  {Group: "", Version: "v1", Kind: "Service"},
	"configmaps":                {Group: "", Version: "v1", Kind: "ConfigMap"},
	"secrets":                   {Group: "", Version: "v1", Kind: "Secret"},
	"serviceaccounts":           {Group: "", Version: "v1", Kind: "ServiceAccount"},
	"ingresses":                 {Group: "networking.k8s.io", Version: "v1", Kind: "Ingress"},
	"networkpolicies":           {Group: "networking.k8s.io", Version: "v1", Kind: "NetworkPolicie"},
	"jobs":                      {Group: "batch", Version: "v1", Kind: "Job"},
	"applications":              {Group: "app.k8s.io", Version: "v1beta1", Kind: "Application"},
	"persistentvolumes":         {Group: "", Version: "v1", Kind: "PersistentVolume"},
	"volumesnapshots":           snapshotv1.SchemeGroupVersion.WithKind("VolumeSnapshot"),
	"volumesnapshotclasses":     snapshotv1.SchemeGroupVersion.WithKind("VolumeSnapshotClass"),
	"namespaces":                {Group: "", Version: "v1", Kind: "Namespace"},
	"customresourcedefinitions": {Group: "apiextensions.k8s.io", Version: "v1", Kind: "CustomResourceDefinition"},
	// TODO later
	// "nodes":                     {Group: "", Version: "v1", Kind: "Node"},
	// "persistentvolumeclaims":               {Group: "", Version: "v1", Kind: "PersistentVolumeClaim"},
	// "pods":              {Group: "", Version: "v1", Kind: "Pod"},
	// iamv1alpha2.ResourcesPluralRoleBinding: rbacv1.SchemeGroupVersion.WithKind(iamv1alpha2.ResourceKindRoleBinding),
	// iamv1alpha2.ResourcesPluralRole: rbacv1.SchemeGroupVersion.WithKind(iamv1alpha2.ResourceKindRole),

	// KubeSphere
	// TODO later
	// networkv1alpha1.ResourcePluralIPPool:            networkv1alpha1.SchemeGroupVersion.WithKind(networkv1alpha1.ResourceKindIPPool),
	// iamv1alpha2.ResourcesPluralGlobalRole:           iamv1alpha2.SchemeGroupVersion.WithKind(iamv1alpha2.ResourceKindGlobalRole),
	// iamv1alpha2.ResourcesPluralUser:          iamv1alpha2.SchemeGroupVersion.WithKind(iamv1alpha2.ResourceKindUser),
	// iamv1alpha2.ResourcesPluralGlobalRoleBinding:    iamv1alpha2.SchemeGroupVersion.WithKind(iamv1alpha2.ResourceKindGlobalRoleBinding),
	// iamv1alpha2.ResourcesPluralWorkspaceRoleBinding: iamv1alpha2.SchemeGroupVersion.WithKind(iamv1alpha2.ResourceKindWorkspaceRoleBinding),
	// iamv1alpha2.ResourcesPluralLoginRecord: iamv1alpha2.SchemeGroupVersion.WithKind(iamv1alpha2.ResourceKindLoginRecord),
	// iamv1alpha2.ResourcePluralGroupBinding: iamv1alpha2.SchemeGroupVersion.WithKind("GroupBinding"),
	// iamv1alpha2.ResourcesPluralClusterRole:          rbacv1.SchemeGroupVersion.WithKind(iamv1alpha2.ResourceKindClusterRole),
	// iamv1alpha2.ResourcesPluralClusterRoleBinding:   rbacv1.SchemeGroupVersion.WithKind(iamv1alpha2.ResourceKindClusterRoleBinding),

	devopsv1alpha3.ResourcePluralDevOpsProject:     devopsv1alpha3.SchemeGroupVersion.WithKind(devopsv1alpha3.ResourceKindDevOpsProject),
	tenantv1alpha1.ResourcePluralWorkspace:         tenantv1alpha1.SchemeGroupVersion.WithKind(tenantv1alpha1.ResourceKindWorkspace),
	tenantv1alpha2.ResourcePluralWorkspaceTemplate: tenantv1alpha2.SchemeGroupVersion.WithKind(tenantv1alpha2.ResourceKindWorkspaceTemplate),
	iamv1alpha2.ResourcesPluralWorkspaceRole:       iamv1alpha2.SchemeGroupVersion.WithKind(iamv1alpha2.ResourceKindWorkspaceRole),
	iamv1alpha2.ResourcePluralGroup:                iamv1alpha2.SchemeGroupVersion.WithKind("Group"),
	clusterv1alpha1.ResourcesPluralCluster:         clusterv1alpha1.SchemeGroupVersion.WithKind(clusterv1alpha1.ResourceKindCluster),
	notificationv2beta1.ResourcesPluralConfig:      notificationv2beta1.SchemeGroupVersion.WithKind(notificationv2beta1.ResourceKindConfig),
	notificationv2beta1.ResourcesPluralReceiver:    notificationv2beta1.SchemeGroupVersion.WithKind(notificationv2beta1.ResourceKindReceiver),
	"clusterdashboards":                            monitoringdashboardv1alpha2.GroupVersion.WithKind("ClusterDashboard"),
	"dashboards":                                   monitoringdashboardv1alpha2.GroupVersion.WithKind("Dashboard"),
	// federated resources
	typesv1beta1.ResourcePluralFederatedNamespace:             typesv1beta1.SchemeGroupVersion.WithKind(typesv1beta1.FederatedNamespaceKind),
	typesv1beta1.ResourcePluralFederatedDeployment:            typesv1beta1.SchemeGroupVersion.WithKind(typesv1beta1.FederatedDeploymentKind),
	typesv1beta1.ResourcePluralFederatedSecret:                typesv1beta1.SchemeGroupVersion.WithKind(typesv1beta1.FederatedSecretKind),
	typesv1beta1.ResourcePluralFederatedConfigmap:             typesv1beta1.SchemeGroupVersion.WithKind(typesv1beta1.FederatedConfigMapKind),
	typesv1beta1.ResourcePluralFederatedService:               typesv1beta1.SchemeGroupVersion.WithKind(typesv1beta1.FederatedServiceKind),
	typesv1beta1.ResourcePluralFederatedApplication:           typesv1beta1.SchemeGroupVersion.WithKind(typesv1beta1.FederatedApplicationKind),
	typesv1beta1.ResourcePluralFederatedPersistentVolumeClaim: typesv1beta1.SchemeGroupVersion.WithKind(typesv1beta1.FederatedPersistentVolumeClaimKind),
	typesv1beta1.ResourcePluralFederatedStatefulSet:           typesv1beta1.SchemeGroupVersion.WithKind(typesv1beta1.FederatedStatefulSetKind),
	typesv1beta1.ResourcePluralFederatedIngress:               typesv1beta1.SchemeGroupVersion.WithKind(typesv1beta1.FederatedIngressKind),
}

func (r *resourceGetter) initResourceGetters(factory informers.InformerFactory) {
	namespacedResourceGetters := make(map[schema.GroupVersionResource]v1alpha3.Interface)
	namespacedResourceGetters[networkv1alpha1.SchemeGroupVersion.WithResource(networkv1alpha1.ResourcePluralIPPool)] = ippool.New(factory.KubeSphereSharedInformerFactory(), factory.KubernetesSharedInformerFactory())
	namespacedResourceGetters[schema.GroupVersionResource{Group: "", Version: "v1", Resource: "persistentvolumeclaims"}] = persistentvolumeclaim.New(factory.KubernetesSharedInformerFactory(), factory.SnapshotSharedInformerFactory())
	namespacedResourceGetters[schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}] = pod.New(factory.KubernetesSharedInformerFactory())
	namespacedResourceGetters[rbacv1.SchemeGroupVersion.WithResource(iamv1alpha2.ResourcesPluralRoleBinding)] = rolebinding.New(factory.KubernetesSharedInformerFactory())
	namespacedResourceGetters[rbacv1.SchemeGroupVersion.WithResource(iamv1alpha2.ResourcesPluralRole)] = role.New(factory.KubernetesSharedInformerFactory())
	clusterResourceGetters := make(map[schema.GroupVersionResource]v1alpha3.Interface)
	clusterResourceGetters[rbacv1.SchemeGroupVersion.WithResource(iamv1alpha2.ResourcesPluralClusterRole)] = clusterrole.New(factory.KubernetesSharedInformerFactory())
	clusterResourceGetters[rbacv1.SchemeGroupVersion.WithResource(iamv1alpha2.ResourcesPluralClusterRoleBinding)] = clusterrolebinding.New(factory.KubernetesSharedInformerFactory())
	clusterResourceGetters[iamv1alpha2.SchemeGroupVersion.WithResource(iamv1alpha2.ResourcesPluralGlobalRole)] = globalrole.New(factory.KubeSphereSharedInformerFactory())
	clusterResourceGetters[iamv1alpha2.SchemeGroupVersion.WithResource(iamv1alpha2.ResourcesPluralGlobalRoleBinding)] = globalrolebinding.New(factory.KubeSphereSharedInformerFactory())
	clusterResourceGetters[iamv1alpha2.SchemeGroupVersion.WithResource(iamv1alpha2.ResourcePluralGroupBinding)] = groupbinding.New(factory.KubeSphereSharedInformerFactory())
	clusterResourceGetters[iamv1alpha2.SchemeGroupVersion.WithResource(iamv1alpha2.ResourcesPluralWorkspaceRole)] = workspacerole.New(factory.KubeSphereSharedInformerFactory())
	clusterResourceGetters[iamv1alpha2.SchemeGroupVersion.WithResource(iamv1alpha2.ResourcesPluralWorkspaceRoleBinding)] = workspacerolebinding.New(factory.KubeSphereSharedInformerFactory())
	clusterResourceGetters[iamv1alpha2.SchemeGroupVersion.WithResource(iamv1alpha2.ResourcesPluralUser)] = user.New(factory.KubeSphereSharedInformerFactory(), factory.KubernetesSharedInformerFactory())
	clusterResourceGetters[schema.GroupVersionResource{Group: "", Version: "v1", Resource: "nodes"}] = node.New(factory.KubernetesSharedInformerFactory())
	clusterResourceGetters[iamv1alpha2.SchemeGroupVersion.WithResource(iamv1alpha2.ResourcesPluralLoginRecord)] = loginrecord.New(factory.KubeSphereSharedInformerFactory())
	r.clusterResourceGetters = clusterResourceGetters
	r.namespacedResourceGetters = namespacedResourceGetters
}

type resourceGetter struct {
	handleMaps map[string]crds.Reader

	clusterResourceGetters    map[schema.GroupVersionResource]v1alpha3.Interface
	namespacedResourceGetters map[schema.GroupVersionResource]v1alpha3.Interface
}

type ResourceGetter interface {
	Get(resource, namespace, name string) (runtime.Object, error)
	List(resource, namespace string, query *query.Query) (*api.ListResult, error)
}

// NewResourceGetter creates a ResourceGetter instance that can query all registered resources
func NewResourceGetter(factory informers.InformerFactory, sch *runtime.Scheme, cache client.Reader) ResourceGetter {
	return NewResourceGetterWithKind(factory, sch, cache, resourceToGVK)
}

func NewResourceGetterWithKind(factory informers.InformerFactory, sch *runtime.Scheme, cache client.Reader, gvk map[string]schema.GroupVersionKind) ResourceGetter {

	res := &resourceGetter{handleMaps: make(map[string]crds.Reader)}
	for k, v := range gvk {
		h := crds.NewTyped(cache, v, sch)
		res.handleMaps[k] = h
	}

	//For legecy resources with complex logic only
	if factory != nil {
		res.initResourceGetters(factory)
	}

	return res
}

func (r *resourceGetter) Get(resource, namespace, name string) (runtime.Object, error) {

	if getter, ok := r.handleMaps[resource]; ok {
		return getter.Get(types.NamespacedName{Name: name, Namespace: namespace})
	}

	clusterScope := namespace == ""
	if getter := r.TryResource(clusterScope, resource); getter != nil {
		return getter.Get(namespace, name)
	}
	return nil, ErrResourceNotSupported
}

func (r *resourceGetter) List(resource, namespace string, query *query.Query) (*api.ListResult, error) {

	if getter, ok := r.handleMaps[resource]; ok {
		return getter.List(namespace, query)
	}

	clusterScope := namespace == ""
	if getter := r.TryResource(clusterScope, resource); getter != nil {
		return getter.List(namespace, query)
	}

	return nil, ErrResourceNotSupported
}

// TryResource will retrieve a getter with resource name, it doesn't guarantee find resource with correct group version
// need to refactor this use schema.GroupVersionResource
func (r *resourceGetter) TryResource(clusterScope bool, resource string) v1alpha3.Interface {
	if clusterScope {
		for k, v := range r.clusterResourceGetters {
			if k.Resource == resource {
				return v
			}
		}
	}
	for k, v := range r.namespacedResourceGetters {
		if k.Resource == resource {
			return v
		}
	}
	return nil
}
