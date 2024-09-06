/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package quota

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	quotav1alpha2 "kubesphere.io/api/quota/v1alpha2"

	lru "github.com/hashicorp/golang-lru"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/storage"

	utilquota "kubesphere.io/kubesphere/kube/pkg/quota/v1"
)

// Following code copied from github.com/openshift/apiserver-library-go/pkg/admission/quota/clusterresourcequota

type accessor struct {
	client client.Client

	// updatedResourceQuotas holds a cache of quotas that we've updated.  This is used to pull the "really latest" during back to
	// back quota evaluations that touch the same quota doc.  This only works because we can compare etcd resourceVersions
	// for the same resource as integers.  Before this change: 22 updates with 12 conflicts.  after this change: 15 updates with 0 conflicts
	updatedResourceQuotas *lru.Cache
}

// newQuotaAccessor creates an object that conforms to the QuotaAccessor interface to be used to retrieve quota objects.
func newQuotaAccessor(client client.Client) *accessor {
	updatedCache, err := lru.New(100)
	if err != nil {
		// this should never happen
		panic(err)
	}

	return &accessor{
		client:                client,
		updatedResourceQuotas: updatedCache,
	}
}

// UpdateQuotaStatus the newQuota coming in will be incremented from the original.  The difference between the original
// and the new is the amount to add to the namespace total, but the total status is the used value itself
func (a *accessor) UpdateQuotaStatus(newQuota *corev1.ResourceQuota) error {
	// skipping namespaced resource quota
	if newQuota.APIVersion != quotav1alpha2.SchemeGroupVersion.String() {
		klog.V(6).Infof("skipping namespaced resource quota %v %v", newQuota.Namespace, newQuota.Name)
		return nil
	}
	ctx := context.TODO()
	resourceQuota := &quotav1alpha2.ResourceQuota{}
	err := a.client.Get(ctx, types.NamespacedName{Name: newQuota.Name}, resourceQuota)
	if err != nil {
		klog.Errorf("failed to fetch resource quota: %s, %v", newQuota.Name, err)
		return err
	}
	resourceQuota = a.checkCache(resourceQuota)

	// re-assign objectmeta
	// make a copy
	updatedQuota := resourceQuota.DeepCopy()
	updatedQuota.ObjectMeta = newQuota.ObjectMeta
	updatedQuota.Namespace = ""

	// determine change in usage
	usageDiff := utilquota.Subtract(newQuota.Status.Used, updatedQuota.Status.Total.Used)

	// update aggregate usage
	updatedQuota.Status.Total.Used = newQuota.Status.Used

	// update per namespace totals
	oldNamespaceTotals, _ := getResourceQuotasStatusByNamespace(updatedQuota.Status.Namespaces, newQuota.Namespace)
	namespaceTotalCopy := oldNamespaceTotals.DeepCopy()
	newNamespaceTotals := *namespaceTotalCopy
	newNamespaceTotals.Used = utilquota.Add(oldNamespaceTotals.Used, usageDiff)
	insertResourceQuotasStatus(&updatedQuota.Status.Namespaces, quotav1alpha2.ResourceQuotaStatusByNamespace{
		Namespace:           newQuota.Namespace,
		ResourceQuotaStatus: newNamespaceTotals,
	})

	klog.V(6).Infof("update resource quota: %+v", updatedQuota)
	err = a.client.Status().Update(ctx, updatedQuota)
	if err != nil {
		klog.Errorf("failed to update resource quota: %v", err)
		return err
	}

	a.updatedResourceQuotas.Add(resourceQuota.Name, updatedQuota)
	return nil
}

var storageVersioner = storage.APIObjectVersioner{}

// checkCache compares the passed quota against the value in the look-aside cache and returns the newer
// if the cache is out of date, it deletes the stale entry.  This only works because of etcd resourceVersions
// being monotonically increasing integers
func (a *accessor) checkCache(resourceQuota *quotav1alpha2.ResourceQuota) *quotav1alpha2.ResourceQuota {
	uncastCachedQuota, ok := a.updatedResourceQuotas.Get(resourceQuota.Name)
	if !ok {
		return resourceQuota
	}
	cachedQuota := uncastCachedQuota.(*quotav1alpha2.ResourceQuota)

	if storageVersioner.CompareResourceVersion(resourceQuota, cachedQuota) >= 0 {
		a.updatedResourceQuotas.Remove(resourceQuota.Name)
		return resourceQuota
	}
	return cachedQuota
}

func (a *accessor) GetQuotas(namespaceName string) ([]corev1.ResourceQuota, error) {
	resourceQuotaNames, err := a.waitForReadyResourceQuotaNames(namespaceName)
	if err != nil {
		klog.Errorf("failed to fetch resource quota names: %v, %v", namespaceName, err)
		return nil, err
	}
	var result []corev1.ResourceQuota
	for _, resourceQuotaName := range resourceQuotaNames {
		resourceQuota := &quotav1alpha2.ResourceQuota{}
		err = a.client.Get(context.TODO(), types.NamespacedName{Name: resourceQuotaName}, resourceQuota)
		if err != nil {
			klog.Errorf("failed to fetch resource quota %s: %v", resourceQuotaName, err)
			return result, err
		}
		resourceQuota = a.checkCache(resourceQuota)

		// now convert to a ResourceQuota
		convertedQuota := corev1.ResourceQuota{}
		convertedQuota.APIVersion = quotav1alpha2.SchemeGroupVersion.String()
		convertedQuota.ObjectMeta = resourceQuota.ObjectMeta
		convertedQuota.Namespace = namespaceName
		convertedQuota.Spec = resourceQuota.Spec.Quota
		convertedQuota.Status = resourceQuota.Status.Total
		result = append(result, convertedQuota)
	}

	// avoid conflicts with namespaced resource quota
	namespacedResourceQuotas, err := a.waitForReadyNamespacedResourceQuotas(namespaceName)
	if err != nil {
		klog.Errorf("failed to fetch namespaced resource quotas: %v, %v", namespaceName, err)
		return nil, err
	}
	for _, resourceQuota := range namespacedResourceQuotas {
		resourceQuota.APIVersion = corev1.SchemeGroupVersion.String()
		result = append(result, resourceQuota)
	}
	return result, nil
}

func (a *accessor) waitForReadyResourceQuotaNames(namespaceName string) ([]string, error) {
	var resourceQuotaNames []string
	// wait for a valid mapping cache.  The overall response can be delayed for up to 10 seconds.
	err := utilwait.PollUntilContextTimeout(context.TODO(), 100*time.Millisecond, 8*time.Second, true, func(ctx context.Context) (done bool, err error) {
		resourceQuotaNames, err = resourceQuotaNamesFor(ctx, a.client, namespaceName)
		// if we can't find the namespace yet, just wait for the cache to update.  Requests to non-existent namespaces
		// may hang, but those people are doing something wrong and namespace lifecycle should reject them.
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		return true, nil
	})
	return resourceQuotaNames, err
}

func (a *accessor) waitForReadyNamespacedResourceQuotas(namespaceName string) ([]corev1.ResourceQuota, error) {
	var resourceQuotas []corev1.ResourceQuota
	// wait for a valid mapping cache.  The overall response can be delayed for up to 10 seconds.
	err := utilwait.PollUntilContextTimeout(context.TODO(), 100*time.Millisecond, 8*time.Second, true, func(ctx context.Context) (done bool, err error) {
		resourceQuotaList := &corev1.ResourceQuotaList{}
		err = a.client.List(ctx, resourceQuotaList, &client.ListOptions{Namespace: namespaceName})
		if err != nil {
			return false, err
		}
		resourceQuotas = resourceQuotaList.Items
		return true, nil
	})
	return resourceQuotas, err
}
