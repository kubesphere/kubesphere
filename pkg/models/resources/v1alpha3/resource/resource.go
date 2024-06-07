/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package resource

import (
	"errors"

	"github.com/Masterminds/semver/v3"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	"kubesphere.io/api/tenant/v1beta1"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/cluster"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/clusterrole"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/clusterrolebinding"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/configmap"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/cronjob"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/customresourcedefinition"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/daemonset"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/deployment"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/globalrole"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/globalrolebinding"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/group"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/groupbinding"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/hpa"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/ingress"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/job"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/label"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/loginrecord"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/namespace"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/node"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/persistentvolume"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/persistentvolumeclaim"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/pod"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/role"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/rolebinding"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/secret"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/service"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/serviceaccount"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/statefulset"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/user"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/workspace"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/workspacerole"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/workspacerolebinding"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/workspacetemplate"
)

var ErrResourceNotSupported = errors.New("resource is not supported")

type Getter struct {
	clusterResourceGetters    map[schema.GroupVersionResource]v1alpha3.Interface
	namespacedResourceGetters map[schema.GroupVersionResource]v1alpha3.Interface
}

func NewResourceGetter(cache runtimeclient.Reader, k8sVersion *semver.Version) *Getter {
	namespacedResourceGetters := make(map[schema.GroupVersionResource]v1alpha3.Interface)
	clusterResourceGetters := make(map[schema.GroupVersionResource]v1alpha3.Interface)

	namespacedResourceGetters[schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}] = deployment.New(cache)
	namespacedResourceGetters[schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"}] = daemonset.New(cache)
	namespacedResourceGetters[schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"}] = statefulset.New(cache)
	namespacedResourceGetters[schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}] = service.New(cache)
	namespacedResourceGetters[schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}] = configmap.New(cache)
	namespacedResourceGetters[schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}] = secret.New(cache)
	namespacedResourceGetters[schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}] = pod.New(cache)
	namespacedResourceGetters[schema.GroupVersionResource{Group: "", Version: "v1", Resource: "serviceaccounts"}] = serviceaccount.New(cache)
	namespacedResourceGetters[schema.GroupVersionResource{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"}] = ingress.New(cache)
	namespacedResourceGetters[schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "jobs"}] = job.New(cache)
	namespacedResourceGetters[schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "cronjobs"}] = cronjob.New(cache, k8sVersion)
	namespacedResourceGetters[schema.GroupVersionResource{Group: "", Version: "v1", Resource: "persistentvolumeclaims"}] = persistentvolumeclaim.New(cache)
	namespacedResourceGetters[schema.GroupVersionResource{Group: "autoscaling", Version: "v2", Resource: "horizontalpodautoscalers"}] = hpa.New(cache, k8sVersion)
	namespacedResourceGetters[rbacv1.SchemeGroupVersion.WithResource(iamv1beta1.ResourcesPluralRoleBinding)] = rolebinding.New(cache)
	namespacedResourceGetters[rbacv1.SchemeGroupVersion.WithResource(iamv1beta1.ResourcesPluralRole)] = role.New(cache)

	clusterResourceGetters[schema.GroupVersionResource{Group: "", Version: "v1", Resource: "persistentvolumes"}] = persistentvolume.New(cache)
	clusterResourceGetters[schema.GroupVersionResource{Group: "", Version: "v1", Resource: "nodes"}] = node.New(cache)
	clusterResourceGetters[schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}] = namespace.New(cache)
	clusterResourceGetters[schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}] = customresourcedefinition.New(cache)

	// kubesphere resources
	clusterResourceGetters[v1beta1.SchemeGroupVersion.WithResource(v1beta1.ResourcePluralWorkspace)] = workspace.New(cache)
	clusterResourceGetters[v1beta1.SchemeGroupVersion.WithResource(tenantv1beta1.ResourcePluralWorkspaceTemplate)] = workspacetemplate.New(cache)
	clusterResourceGetters[iamv1beta1.SchemeGroupVersion.WithResource(iamv1beta1.ResourcesPluralGlobalRole)] = globalrole.New(cache)
	clusterResourceGetters[iamv1beta1.SchemeGroupVersion.WithResource(iamv1beta1.ResourcesPluralWorkspaceRole)] = workspacerole.New(cache)
	clusterResourceGetters[iamv1beta1.SchemeGroupVersion.WithResource(iamv1beta1.ResourcesPluralUser)] = user.New(cache)
	clusterResourceGetters[iamv1beta1.SchemeGroupVersion.WithResource(iamv1beta1.ResourcesPluralGlobalRoleBinding)] = globalrolebinding.New(cache)
	clusterResourceGetters[iamv1beta1.SchemeGroupVersion.WithResource(iamv1beta1.ResourcesPluralWorkspaceRoleBinding)] = workspacerolebinding.New(cache)
	clusterResourceGetters[iamv1beta1.SchemeGroupVersion.WithResource(iamv1beta1.ResourcesPluralLoginRecord)] = loginrecord.New(cache)
	clusterResourceGetters[iamv1beta1.SchemeGroupVersion.WithResource(iamv1beta1.ResourcePluralGroup)] = group.New(cache)
	clusterResourceGetters[iamv1beta1.SchemeGroupVersion.WithResource(iamv1beta1.ResourcePluralGroupBinding)] = groupbinding.New(cache)
	clusterResourceGetters[rbacv1.SchemeGroupVersion.WithResource(iamv1beta1.ResourcesPluralClusterRole)] = clusterrole.New(cache)
	clusterResourceGetters[rbacv1.SchemeGroupVersion.WithResource(iamv1beta1.ResourcesPluralClusterRoleBinding)] = clusterrolebinding.New(cache)
	clusterResourceGetters[clusterv1alpha1.SchemeGroupVersion.WithResource(clusterv1alpha1.ResourcesPluralCluster)] = cluster.New(cache)
	clusterResourceGetters[clusterv1alpha1.SchemeGroupVersion.WithResource(clusterv1alpha1.ResourcesPluralLabel)] = label.New(cache)

	return &Getter{
		namespacedResourceGetters: namespacedResourceGetters,
		clusterResourceGetters:    clusterResourceGetters,
	}
}

// TryResource will retrieve a getter with resource name, it doesn't guarantee find resource with correct group version
// need to refactor this use schema.GroupVersionResource
func (r *Getter) TryResource(clusterScope bool, resource string) v1alpha3.Interface {
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

func (r *Getter) Get(resource, namespace, name string) (runtime.Object, error) {
	clusterScope := namespace == ""
	getter := r.TryResource(clusterScope, resource)
	if getter == nil {
		return nil, ErrResourceNotSupported
	}
	return getter.Get(namespace, name)
}

func (r *Getter) List(resource, namespace string, query *query.Query) (*api.ListResult, error) {
	clusterScope := namespace == ""
	getter := r.TryResource(clusterScope, resource)
	if getter == nil {
		return nil, ErrResourceNotSupported
	}
	return getter.List(namespace, query)
}
