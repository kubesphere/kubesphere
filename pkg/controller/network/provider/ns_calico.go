package provider

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	v3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/libcalico-go/lib/clientv3"
	"github.com/projectcalico/libcalico-go/lib/errors"
	"github.com/projectcalico/libcalico-go/lib/options"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/klogr"
	api "kubesphere.io/kubesphere/pkg/apis/network/v1alpha1"
)

var log = klogr.New().WithName("calico-client")
var defaultBackoff = wait.Backoff{
	Steps:    4,
	Duration: 10 * time.Millisecond,
	Factor:   5.0,
	Jitter:   0.1,
}

type calicoNetworkProvider struct {
	np clientv3.NetworkPolicyInterface
}

func NewCalicoNetworkProvider(np clientv3.NetworkPolicyInterface) NsNetworkPolicyProvider {
	return &calicoNetworkProvider{
		np: np,
	}
}
func convertSpec(n *api.NamespaceNetworkPolicySpec) *v3.NetworkPolicySpec {
	bytes, err := json.Marshal(&n)
	if err != nil {
		panic(err)
	}
	m := new(v3.NetworkPolicySpec)
	err = json.Unmarshal(bytes, m)
	if err != nil {
		panic(err)
	}
	return m
}

// ConvertAPIToCalico convert our api to calico api
func ConvertAPIToCalico(n *api.NamespaceNetworkPolicy) *v3.NetworkPolicy {
	output := v3.NewNetworkPolicy()
	//Object Metadata
	output.ObjectMeta.Name = n.Name
	output.Namespace = n.Namespace
	output.Annotations = n.Annotations
	output.Labels = n.Labels
	//spec
	output.Spec = *(convertSpec(&n.Spec))
	return output
}

func (k *calicoNetworkProvider) Get(o *api.NamespaceNetworkPolicy) (interface{}, error) {
	return k.np.Get(context.TODO(), o.Namespace, o.Name, options.GetOptions{})
}

func (k *calicoNetworkProvider) Add(o *api.NamespaceNetworkPolicy) error {
	log.V(3).Info("Creating network policy", "name", o.Name, "namespace", o.Namespace)
	obj := ConvertAPIToCalico(o)
	log.V(4).Info("Show object spe detail", "name", o.Name, "namespace", o.Namespace, "Spec", obj.Spec)
	_, err := k.np.Create(context.TODO(), obj, options.SetOptions{})
	return err
}

func (k *calicoNetworkProvider) CheckExist(o *api.NamespaceNetworkPolicy) (bool, error) {
	log.V(3).Info("Checking network policy whether exsits or not", "name", o.Name, "namespace", o.Namespace)
	out, err := k.np.Get(context.Background(), o.Namespace, o.Name, options.GetOptions{})
	if err != nil {
		if _, ok := err.(errors.ErrorResourceDoesNotExist); ok {
			return false, nil
		}
		return false, err
	}
	if out != nil {
		return true, nil
	}
	return false, nil
}

func (k *calicoNetworkProvider) Delete(o *api.NamespaceNetworkPolicy) error {
	log.V(3).Info("Deleting network policy", "name", o.Name, "namespace", o.Namespace)
	_, err := k.np.Delete(context.Background(), o.Namespace, o.Name, options.DeleteOptions{})
	return err
}

func (k *calicoNetworkProvider) NeedUpdate(o *api.NamespaceNetworkPolicy) (bool, error) {
	store, err := k.np.Get(context.Background(), o.Namespace, o.Name, options.GetOptions{})
	if err != nil {
		log.Error(err, "Failed to get resource", "name", o.Name, "namespace", o.Namespace)
	}
	expected := ConvertAPIToCalico(o)
	log.V(4).Info("Comparing Spec", "store", store.Spec, "current", expected.Spec)
	if !reflect.DeepEqual(store.Spec, expected.Spec) {
		return true, nil
	}
	return false, nil
}

func (k *calicoNetworkProvider) Update(o *api.NamespaceNetworkPolicy) error {
	log.V(3).Info("Updating network policy", "name", o.Name, "namespace", o.Namespace)
	updateObject, err := k.Get(o)
	if err != nil {
		log.Error(err, "Failed to get resource in store")
		return err
	}
	up := updateObject.(*v3.NetworkPolicy)
	up.Spec = *convertSpec(&o.Spec)
	err = RetryOnConflict(defaultBackoff, func() error {
		_, err := k.np.Update(context.Background(), up, options.SetOptions{})
		return err
	})
	if err != nil {
		log.Error(err, "Failed to update resource", "name", o.Name, "namespace", o.Namespace)
	}
	return err
}

// RetryOnConflict is same as the function in k8s, but replaced with error in calico
func RetryOnConflict(backoff wait.Backoff, fn func() error) error {
	var lastConflictErr error
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := fn()
		if err == nil {
			return true, nil
		}
		if _, ok := err.(errors.ErrorResourceUpdateConflict); ok {
			lastConflictErr = err
			return false, nil
		}
		return false, err
	})
	if err == wait.ErrWaitTimeout {
		err = lastConflictErr
	}
	return err
}
