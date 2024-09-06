/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package quotas

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver/v3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
)

const (
	podsKey                   = "count/pods"
	daemonsetsKey             = "count/daemonsets.apps"
	deploymentsKey            = "count/deployments.apps"
	ingressKey                = "count/ingresses.extensions"
	servicesKey               = "count/services"
	statefulsetsKey           = "count/statefulsets.apps"
	persistentvolumeclaimsKey = "persistentvolumeclaims"
	jobsKey                   = "count/jobs.batch"
	cronJobsKey               = "count/cronjobs.batch"
)

type ResourceQuotaGetter interface {
	GetClusterQuota() (*api.ResourceQuota, error)
	GetNamespaceQuota(namespace string) (*api.NamespacedResourceQuota, error)
}

type resourceQuotaGetter struct {
	client             runtimeclient.Client
	supportedResources map[string]schema.GroupVersionKind
}

func NewResourceQuotaGetter(client runtimeclient.Client, k8sVersion *semver.Version) ResourceQuotaGetter {
	supportedResources := map[string]schema.GroupVersionKind{
		deploymentsKey:            {Group: "apps", Version: "v1", Kind: "DeploymentList"},
		daemonsetsKey:             {Group: "apps", Version: "v1", Kind: "DaemonSetList"},
		statefulsetsKey:           {Group: "apps", Version: "v1", Kind: "StatefulSetList"},
		podsKey:                   {Group: "", Version: "v1", Kind: "PodList"},
		servicesKey:               {Group: "", Version: "v1", Kind: "ServiceList"},
		persistentvolumeclaimsKey: {Group: "", Version: "v1", Kind: "PersistentVolumeClaimList"},
		ingressKey:                {Group: "networking.k8s.io", Version: "v1", Kind: "IngressList"},
		jobsKey:                   {Group: "batch", Version: "v1", Kind: "JobList"},
		cronJobsKey:               {Group: "batch", Version: "v1", Kind: "CronJobList"},
	}
	if k8sutil.ServeBatchV1beta1(k8sVersion) {
		cronJobs := supportedResources[cronJobsKey]
		cronJobs.Version = "v1beta1"
		supportedResources[cronJobsKey] = cronJobs
	}

	return &resourceQuotaGetter{client: client, supportedResources: supportedResources}
}

func (c *resourceQuotaGetter) getUsage(namespace, resource string) (int, error) {
	var obj runtimeclient.ObjectList
	gvk, ok := c.supportedResources[resource]
	if !ok {
		return 0, fmt.Errorf("resource %s is not supported", resource)
	}

	if c.client.Scheme().Recognizes(gvk) {
		gvkObject, err := c.client.Scheme().New(gvk)
		if err != nil {
			return 0, nil
		}
		obj = gvkObject.(runtimeclient.ObjectList)
	} else {
		u := &unstructured.UnstructuredList{}
		u.SetGroupVersionKind(gvk)
		obj = u
	}

	if err := c.client.List(context.Background(), obj, runtimeclient.InNamespace(namespace)); err != nil {
		return 0, err
	}

	items, err := meta.ExtractList(obj)
	if err != nil {
		return 0, err
	}
	return len(items), nil
}

// GetClusterQuota no one use this api anymore， marked as deprecated
func (c *resourceQuotaGetter) GetClusterQuota() (*api.ResourceQuota, error) {

	quota := v1.ResourceQuotaStatus{Hard: make(v1.ResourceList), Used: make(v1.ResourceList)}

	for r := range c.supportedResources {
		used, err := c.getUsage("", r)
		if err != nil {
			return nil, err
		}
		var quantity resource.Quantity
		quantity.Set(int64(used))
		quota.Used[v1.ResourceName(r)] = quantity
	}

	return &api.ResourceQuota{Namespace: "\"\"", Data: quota}, nil

}

func (c *resourceQuotaGetter) GetNamespaceQuota(namespace string) (*api.NamespacedResourceQuota, error) {
	quota, err := c.getNamespaceResourceQuota(namespace)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	if quota == nil {
		quota = &v1.ResourceQuotaStatus{Hard: make(v1.ResourceList), Used: make(v1.ResourceList)}
	}

	var resourceQuotaLeft = v1.ResourceList{}

	for key, hardLimit := range quota.Hard {
		if used, ok := quota.Used[key]; ok {
			left := hardLimit.DeepCopy()
			left.Sub(used)
			if hardLimit.Cmp(used) < 0 {
				left = resource.MustParse("0")
			}

			resourceQuotaLeft[key] = left
		}
	}

	// add extra quota usage, cause user may not specify them
	for key := range c.supportedResources {
		// only add them when they don't exist in quotastatus
		if _, ok := quota.Used[v1.ResourceName(key)]; !ok {
			used, err := c.getUsage(namespace, key)
			if err != nil {
				klog.Error(err)
				return nil, err
			}

			quota.Used[v1.ResourceName(key)] = *(resource.NewQuantity(int64(used), resource.DecimalSI))
		}
	}

	var result = api.NamespacedResourceQuota{
		Namespace: namespace,
	}
	result.Data.Hard = quota.Hard
	result.Data.Used = quota.Used
	result.Data.Left = resourceQuotaLeft

	return &result, nil

}

func updateNamespaceQuota(tmpResourceList, resourceList v1.ResourceList) {
	if tmpResourceList == nil {
		tmpResourceList = resourceList
	}
	for res, usage := range resourceList {
		tmpUsage, exist := tmpResourceList[res]
		if !exist {
			tmpResourceList[res] = usage
		}
		if tmpUsage.Cmp(usage) == 1 {
			tmpResourceList[res] = usage
		}
	}
}

func (c *resourceQuotaGetter) getNamespaceResourceQuota(namespace string) (*v1.ResourceQuotaStatus, error) {
	resourceQuotaList := &v1.ResourceQuotaList{}
	if err := c.client.List(context.Background(), resourceQuotaList, runtimeclient.InNamespace(namespace)); err != nil {
		klog.Error(err)
		return nil, err
	} else if len(resourceQuotaList.Items) == 0 {
		return nil, nil
	}

	quotaStatus := v1.ResourceQuotaStatus{Hard: make(v1.ResourceList), Used: make(v1.ResourceList)}

	for _, quota := range resourceQuotaList.Items {
		updateNamespaceQuota(quotaStatus.Hard, quota.Status.Hard)
		updateNamespaceQuota(quotaStatus.Used, quota.Status.Used)
	}

	return &quotaStatus, nil
}
