package provider

import (
	"reflect"

	"github.com/projectcalico/libcalico-go/lib/errors"
	"k8s.io/client-go/tools/cache"
	api "kubesphere.io/kubesphere/pkg/apis/network/v1alpha1"
)

func NewFakeCalicoNetworkProvider() *FakeCalicoNetworkProvider {
	f := new(FakeCalicoNetworkProvider)
	f.NSNPData = make(map[string]*api.NamespaceNetworkPolicy)
	return f
}

type FakeCalicoNetworkProvider struct {
	NSNPData map[string]*api.NamespaceNetworkPolicy
}

func (f *FakeCalicoNetworkProvider) Get(o *api.NamespaceNetworkPolicy) (interface{}, error) {
	namespacename, _ := cache.MetaNamespaceKeyFunc(o)
	obj, ok := f.NSNPData[namespacename]
	if !ok {
		return nil, errors.ErrorResourceDoesNotExist{}
	}
	return obj, nil
}

func (f *FakeCalicoNetworkProvider) Add(o *api.NamespaceNetworkPolicy) error {
	namespacename, _ := cache.MetaNamespaceKeyFunc(o)
	if _, ok := f.NSNPData[namespacename]; ok {
		return errors.ErrorResourceAlreadyExists{}
	}
	f.NSNPData[namespacename] = o
	return nil
}

func (f *FakeCalicoNetworkProvider) CheckExist(o *api.NamespaceNetworkPolicy) (bool, error) {
	namespacename, _ := cache.MetaNamespaceKeyFunc(o)
	if _, ok := f.NSNPData[namespacename]; ok {
		return true, nil
	}
	return false, nil
}

func (f *FakeCalicoNetworkProvider) NeedUpdate(o *api.NamespaceNetworkPolicy) (bool, error) {
	namespacename, _ := cache.MetaNamespaceKeyFunc(o)
	store := f.NSNPData[namespacename]
	if !reflect.DeepEqual(store, o) {
		return true, nil
	}
	return false, nil
}

func (f *FakeCalicoNetworkProvider) Update(o *api.NamespaceNetworkPolicy) error {
	namespacename, _ := cache.MetaNamespaceKeyFunc(o)
	f.NSNPData[namespacename] = o
	return nil
}

func (f *FakeCalicoNetworkProvider) Delete(o *api.NamespaceNetworkPolicy) error {
	namespacename, _ := cache.MetaNamespaceKeyFunc(o)
	delete(f.NSNPData, namespacename)
	return nil
}
