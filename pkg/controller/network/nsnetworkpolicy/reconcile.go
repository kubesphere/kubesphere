package nsnetworkpolicy

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
	"kubesphere.io/kubesphere/pkg/apis/network/v1alpha1"
	"kubesphere.io/kubesphere/pkg/controller/network/utils"
)

const (
	controllerFinalizier = "nsnp.finalizers.networking.kubesphere.io"
)

var clog logr.Logger

func (c *controller) reconcile(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	clog = log.WithValues("name", name, "namespace", namespace)
	clog.V(1).Info("---------Begin to reconcile--------")
	defer clog.V(1).Info("---------Reconcile done--------")
	obj, err := c.nsnpLister.NamespaceNetworkPolicies(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			clog.V(2).Info("Object is removed")
			return nil
		}
		clog.Error(err, "Failed to get resource")
		return err
	}
	stop, err := c.addOrRemoveFinalizer(obj)
	if err != nil {
		return err
	}
	if stop {
		return nil
	}
	clog.V(2).Info("Check if we need a create or update")
	ok, err := c.nsNetworkPolicyProvider.CheckExist(obj)
	if err != nil {
		clog.Error(err, "Failed to check exist of network policy")
		return err
	}
	if !ok {
		clog.V(1).Info("Create a new object in backend")
		err = c.nsNetworkPolicyProvider.Add(obj)
		if err != nil {
			clog.Error(err, "Failed to create np")
			return err
		}
		return nil
	}

	needUpdate, err := c.nsNetworkPolicyProvider.NeedUpdate(obj)
	if err != nil {
		clog.Error(err, "Failed to check if object need a update")
		return err
	}
	if needUpdate {
		clog.V(1).Info("Update object in backend")
		err = c.nsNetworkPolicyProvider.Update(obj)
		if err != nil {
			clog.Error(err, "Failed to update object")
			return err
		}
	}
	return nil
}

func (c *controller) addOrRemoveFinalizer(obj *v1alpha1.NamespaceNetworkPolicy) (bool, error) {
	if obj.ObjectMeta.DeletionTimestamp.IsZero() {
		if !utils.ContainsString(obj.ObjectMeta.Finalizers, controllerFinalizier) {
			clog.V(2).Info("Detect no finalizer")
			obj.ObjectMeta.Finalizers = append(obj.ObjectMeta.Finalizers, controllerFinalizier)
			err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
				_, err := c.kubesphereClientset.NetworkV1alpha1().NamespaceNetworkPolicies(obj.Namespace).Update(obj)
				return err
			})
			if err != nil {
				clog.Error(err, "Failed to add finalizer")
				return false, err
			}
			return false, nil
		}
	} else {
		// The object is being deleted
		if utils.ContainsString(obj.ObjectMeta.Finalizers, controllerFinalizier) {
			// our finalizer is present, so lets handle any external dependency
			if err := c.deleteProviderNSNP(obj); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return false, err
			}
			clog.V(2).Info("Removing finalizer")
			// remove our finalizer from the list and update it.
			obj.ObjectMeta.Finalizers = utils.RemoveString(obj.ObjectMeta.Finalizers, controllerFinalizier)
			err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
				_, err := c.kubesphereClientset.NetworkV1alpha1().NamespaceNetworkPolicies(obj.Namespace).Update(obj)
				return err
			})
			if err != nil {
				clog.Error(err, "Failed to remove finalizer")
				return false, err
			}
			return true, nil
		}
	}
	return false, nil
}

// deleteProviderNSNP delete network policy in the backend
func (c *controller) deleteProviderNSNP(obj *v1alpha1.NamespaceNetworkPolicy) error {
	clog.V(2).Info("Deleting backend network policy")
	return c.nsNetworkPolicyProvider.Delete(obj)
}
