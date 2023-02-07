/*
Copyright 2016 The Kubernetes Authors.

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

package util

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/pkg/errors"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	fedcommon "sigs.k8s.io/kubefed/pkg/apis/core/common"
	fedv1b1 "sigs.k8s.io/kubefed/pkg/apis/core/v1beta1"
	"sigs.k8s.io/kubefed/pkg/client/generic"
	"sigs.k8s.io/kubefed/pkg/metrics"
)

const (
	clusterSyncPeriod = 10 * time.Minute
	userAgentName     = "kubefed-controller"
)

// An object with an origin information.
type FederatedObject struct {
	Object      interface{}
	ClusterName string
}

// FederatedReadOnlyStore is an overlay over multiple stores created in federated clusters.
type FederatedReadOnlyStore interface {
	// Returns all items in the store.
	List() ([]FederatedObject, error)

	// Returns all items from a cluster.
	ListFromCluster(clusterName string) ([]interface{}, error)

	// GetKeyFor returns the key under which the item would be put in the store.
	GetKeyFor(item interface{}) string

	// GetByKey returns the item stored under the given key in the specified cluster (if exist).
	GetByKey(clusterName string, key string) (interface{}, bool, error)

	// Returns the items stored under the given key in all clusters.
	GetFromAllClusters(key string) ([]FederatedObject, error)

	// Checks whether stores for all clusters form the lists (and only these) are there and
	// are synced. This is only a basic check whether the data inside of the store is usable.
	// It is not a full synchronization/locking mechanism it only tries to ensure that out-of-sync
	// issues occur less often.	All users of the interface should assume
	// that there may be significant delays in content updates of all kinds and write their
	// code that it doesn't break if something is slightly out-of-sync.
	ClustersSynced(clusters []*fedv1b1.KubeFedCluster) bool
}

// An interface to retrieve both KubeFedCluster resources and clients
// to access the clusters they represent.
type RegisteredClustersView interface {
	// GetClientForCluster returns a client for the cluster, if present.
	GetClientForCluster(clusterName string) (generic.Client, error)

	// GetUnreadyClusters returns a list of all clusters that are not ready yet.
	GetUnreadyClusters() ([]*fedv1b1.KubeFedCluster, error)

	// GetReadyClusters returns all clusters for which the sub-informers are run.
	GetReadyClusters() ([]*fedv1b1.KubeFedCluster, error)

	// GetClusters returns a list of all clusters.
	GetClusters() ([]*fedv1b1.KubeFedCluster, error)

	// GetReadyCluster returns the cluster with the given name, if found.
	GetReadyCluster(name string) (*fedv1b1.KubeFedCluster, bool, error)

	// ClustersSynced returns true if the view is synced (for the first time).
	ClustersSynced() bool
}

// FederatedInformer provides access to clusters registered with a
// KubeFed control plane and watches a given resource type in
// registered clusters.
//
// Whenever a new cluster is registered with KubeFed, an informer is
// created for it using TargetInformerFactory. Informers are stopped
// when a cluster is either put offline of deleted. It is assumed that
// some controller keeps an eye on the cluster list and thus the
// clusters in ETCD are up to date.
type FederatedInformer interface {
	RegisteredClustersView

	// Returns a store created over all stores from target informers.
	GetTargetStore() FederatedReadOnlyStore

	// Starts all the processes.
	Start()

	// Stops all the processes inside the informer.
	Stop()
}

// A function that should be used to create an informer on the target object. Store should use
// cache.DeletionHandlingMetaNamespaceKeyFunc as a keying function.
type TargetInformerFactory func(*fedv1b1.KubeFedCluster, *restclient.Config) (cache.Store, cache.Controller, error)

// A structure with cluster lifecycle handler functions. Cluster is available (and ClusterAvailable is fired)
// when it is created in federated etcd and ready. Cluster becomes unavailable (and ClusterUnavailable is fired)
// when it is either deleted or becomes not ready. When cluster spec (IP)is modified both ClusterAvailable
// and ClusterUnavailable are fired.
type ClusterLifecycleHandlerFuncs struct {
	// Fired when the cluster becomes available.
	ClusterAvailable func(*fedv1b1.KubeFedCluster)
	// Fired when the cluster becomes unavailable. The second arg contains data that was present
	// in the cluster before deletion.
	ClusterUnavailable func(*fedv1b1.KubeFedCluster, []interface{})
}

// Builds a FederatedInformer for the given configuration.
func NewFederatedInformer(
	config *ControllerConfig,
	client generic.Client,
	apiResource *metav1.APIResource,
	triggerFunc func(runtimeclient.Object),
	clusterLifecycle *ClusterLifecycleHandlerFuncs) (FederatedInformer, error) {
	targetInformerFactory := func(cluster *fedv1b1.KubeFedCluster, clusterConfig *restclient.Config) (cache.Store, cache.Controller, error) {
		resourceClient, err := NewResourceClient(clusterConfig, apiResource)
		if err != nil {
			return nil, nil, err
		}
		targetNamespace := NamespaceForCluster(cluster.Name, config.TargetNamespace)
		store, controller := NewManagedResourceInformer(resourceClient, targetNamespace, apiResource, triggerFunc)
		return store, controller, nil
	}

	federatedInformer := &federatedInformerImpl{
		targetInformerFactory: targetInformerFactory,
		configFactory: func(cluster *fedv1b1.KubeFedCluster) (*restclient.Config, error) {
			clusterConfig, err := BuildClusterConfig(cluster, client, config.KubeFedNamespace)
			if err != nil {
				return nil, err
			}
			if clusterConfig == nil {
				return nil, errors.Errorf("Unable to load configuration for cluster %q", cluster.Name)
			}
			restclient.AddUserAgent(clusterConfig, userAgentName)
			return clusterConfig, nil
		},
		targetInformers: make(map[string]informer),
		fedNamespace:    config.KubeFedNamespace,
		clusterClients:  make(map[string]generic.Client),
	}

	getClusterData := func(name string) []interface{} {
		data, err := federatedInformer.GetTargetStore().ListFromCluster(name)
		if err != nil {
			klog.Errorf("Failed to list %s content: %v", name, err)
			return make([]interface{}, 0)
		}
		return data
	}

	var err error
	federatedInformer.clusterInformer.store, federatedInformer.clusterInformer.controller, err = NewGenericInformerWithEventHandler(
		config.KubeConfig,
		config.KubeFedNamespace,
		&fedv1b1.KubeFedCluster{},
		clusterSyncPeriod,
		&cache.ResourceEventHandlerFuncs{
			DeleteFunc: func(old interface{}) {
				oldCluster, ok := old.(*fedv1b1.KubeFedCluster)
				if ok {
					var data []interface{}
					if clusterLifecycle.ClusterUnavailable != nil {
						data = getClusterData(oldCluster.Name)
					}
					federatedInformer.deleteCluster(oldCluster)
					if clusterLifecycle.ClusterUnavailable != nil {
						clusterLifecycle.ClusterUnavailable(oldCluster, data)
					}
				}
			},
			AddFunc: func(cur interface{}) {
				curCluster, ok := cur.(*fedv1b1.KubeFedCluster)
				switch {
				case !ok:
					klog.Errorf("Cluster %v/%v not added; incorrect type", curCluster.Namespace, curCluster.Name)
				case IsClusterReady(&curCluster.Status):
					federatedInformer.addCluster(curCluster)
					klog.Infof("Cluster %v/%v is ready", curCluster.Namespace, curCluster.Name)
					if clusterLifecycle.ClusterAvailable != nil {
						clusterLifecycle.ClusterAvailable(curCluster)
					}
				default:
					klog.Infof("Cluster %v/%v not added; it is not ready.", curCluster.Namespace, curCluster.Name)
				}
			},
			UpdateFunc: func(old, cur interface{}) {
				oldCluster, ok := old.(*fedv1b1.KubeFedCluster)
				if !ok {
					klog.Errorf("Internal error: Cluster %v not updated. Old cluster not of correct type.", old)
					return
				}
				curCluster, ok := cur.(*fedv1b1.KubeFedCluster)
				if !ok {
					klog.Errorf("Internal error: Cluster %v not updated. New cluster not of correct type.", cur)
					return
				}
				if IsClusterReady(&oldCluster.Status) != IsClusterReady(&curCluster.Status) || !reflect.DeepEqual(oldCluster.Spec, curCluster.Spec) || !reflect.DeepEqual(oldCluster.ObjectMeta.Labels, curCluster.ObjectMeta.Labels) || !reflect.DeepEqual(oldCluster.ObjectMeta.Annotations, curCluster.ObjectMeta.Annotations) {
					var data []interface{}
					if clusterLifecycle.ClusterUnavailable != nil {
						data = getClusterData(oldCluster.Name)
					}
					federatedInformer.deleteCluster(oldCluster)
					if clusterLifecycle.ClusterUnavailable != nil {
						clusterLifecycle.ClusterUnavailable(oldCluster, data)
					}

					if IsClusterReady(&curCluster.Status) {
						federatedInformer.addCluster(curCluster)
						if clusterLifecycle.ClusterAvailable != nil {
							clusterLifecycle.ClusterAvailable(curCluster)
						}
					}
				} else {
					klog.V(7).Infof("Cluster %v not updated to %v as ready status and specs are identical", oldCluster, curCluster)
				}
			},
		},
	)
	return federatedInformer, err
}

func IsClusterReady(clusterStatus *fedv1b1.KubeFedClusterStatus) bool {
	for _, condition := range clusterStatus.Conditions {
		if condition.Type == fedcommon.ClusterReady {
			if condition.Status == apiv1.ConditionTrue {
				return true
			}
		}
	}
	return false
}

type informer struct {
	controller cache.Controller
	store      cache.Store
	stopChan   chan struct{}
}

type federatedInformerImpl struct {
	sync.Mutex

	// Informer on federated clusters.
	clusterInformer informer

	// Target informers factory
	targetInformerFactory TargetInformerFactory

	// Structures returned by targetInformerFactory
	targetInformers map[string]informer

	// Retrieves configuration to access a cluster.
	configFactory func(*fedv1b1.KubeFedCluster) (*restclient.Config, error)

	// Caches cluster clients (reduces client discovery and secret retrieval)
	clusterClients map[string]generic.Client

	// Namespace from which to source KubeFedCluster resources
	fedNamespace string
}

// *federatedInformerImpl implements FederatedInformer interface.
var _ FederatedInformer = &federatedInformerImpl{}

type federatedStoreImpl struct {
	federatedInformer *federatedInformerImpl
}

func (f *federatedInformerImpl) Stop() {
	klog.V(4).Infof("Stopping federated informer.")
	f.Lock()
	defer f.Unlock()

	klog.V(4).Infof("... Closing cluster informer channel.")
	close(f.clusterInformer.stopChan)
	for key, informer := range f.targetInformers {
		klog.V(4).Infof("... Closing informer channel for %q.", key)
		close(informer.stopChan)
		// Remove each informer after it has been stopped to prevent
		// subsequent cluster deletion from attempting to double close
		// an informer's stop channel.
		delete(f.targetInformers, key)
	}
}

func (f *federatedInformerImpl) Start() {
	f.Lock()
	defer f.Unlock()

	f.clusterInformer.stopChan = make(chan struct{})
	go f.clusterInformer.controller.Run(f.clusterInformer.stopChan)
}

// GetClientForCluster returns a client for the cluster, if present.
func (f *federatedInformerImpl) GetClientForCluster(clusterName string) (generic.Client, error) {
	defer metrics.ClusterClientConnectionDurationFromStart(time.Now())
	f.Lock()
	defer f.Unlock()

	// return cached client if one exists (to prevent frequent secret retrieval and rest discovery)
	if client, ok := f.clusterClients[clusterName]; ok {
		return client, nil
	}
	config, err := f.getConfigForClusterUnlocked(clusterName)
	if err != nil {
		return nil, errors.Wrap(err, "Client creation failed")
	}
	client, err := generic.New(config)
	if err != nil {
		return client, err
	}
	f.clusterClients[clusterName] = client

	return client, nil
}

func (f *federatedInformerImpl) getConfigForClusterUnlocked(clusterName string) (*restclient.Config, error) {
	// No locking needed. Will happen in f.GetCluster.
	klog.V(4).Infof("Getting config for cluster %q", clusterName)
	if cluster, found, err := f.getReadyClusterUnlocked(clusterName); found && err == nil {
		return f.configFactory(cluster)
	} else if err != nil {
		return nil, err
	}
	return nil, errors.Errorf("cluster %q not found", clusterName)
}

func (f *federatedInformerImpl) GetUnreadyClusters() ([]*fedv1b1.KubeFedCluster, error) {
	f.Lock()
	defer f.Unlock()

	items := f.clusterInformer.store.List()
	result := make([]*fedv1b1.KubeFedCluster, 0, len(items))
	for _, item := range items {
		if cluster, ok := item.(*fedv1b1.KubeFedCluster); ok {
			if !IsClusterReady(&cluster.Status) {
				result = append(result, cluster)
			}
		} else {
			return nil, errors.Errorf("wrong data in FederatedInformerImpl cluster store: %v", item)
		}
	}
	return result, nil
}

// GetReadyClusters returns all clusters for which the sub-informers are run.
func (f *federatedInformerImpl) GetReadyClusters() ([]*fedv1b1.KubeFedCluster, error) {
	return f.getClusters(true)
}

// GetClusters returns all clusters regardless of ready state.
func (f *federatedInformerImpl) GetClusters() ([]*fedv1b1.KubeFedCluster, error) {
	return f.getClusters(false)
}

// GetReadyClusters returns only ready clusters if onlyReady is true and all clusters otherwise.
func (f *federatedInformerImpl) getClusters(onlyReady bool) ([]*fedv1b1.KubeFedCluster, error) {
	f.Lock()
	defer f.Unlock()

	items := f.clusterInformer.store.List()
	result := make([]*fedv1b1.KubeFedCluster, 0, len(items))
	for _, item := range items {
		if cluster, ok := item.(*fedv1b1.KubeFedCluster); ok {
			if !onlyReady || IsClusterReady(&cluster.Status) {
				result = append(result, cluster)
			}
		} else {
			return nil, errors.Errorf("wrong data in FederatedInformerImpl cluster store: %v", item)
		}
	}
	return result, nil
}

// GetCluster returns the cluster with the given name, if found.
func (f *federatedInformerImpl) GetReadyCluster(name string) (*fedv1b1.KubeFedCluster, bool, error) {
	f.Lock()
	defer f.Unlock()
	return f.getReadyClusterUnlocked(name)
}

func (f *federatedInformerImpl) getReadyClusterUnlocked(name string) (*fedv1b1.KubeFedCluster, bool, error) {
	key := fmt.Sprintf("%s/%s", f.fedNamespace, name)
	if obj, exist, err := f.clusterInformer.store.GetByKey(key); exist && err == nil {
		if cluster, ok := obj.(*fedv1b1.KubeFedCluster); ok {
			if IsClusterReady(&cluster.Status) {
				return cluster, true, nil
			}
			return nil, false, nil
		}
		return nil, false, errors.Errorf("wrong data in FederatedInformerImpl cluster store: %v", obj)
	} else {
		return nil, false, err
	}
}

// Synced returns true if the view is synced (for the first time)
func (f *federatedInformerImpl) ClustersSynced() bool {
	return f.clusterInformer.controller.HasSynced()
}

// Adds the given cluster to federated informer.
func (f *federatedInformerImpl) addCluster(cluster *fedv1b1.KubeFedCluster) {
	f.Lock()
	defer f.Unlock()
	name := cluster.Name
	if config, err := f.getConfigForClusterUnlocked(name); err == nil {
		store, controller, err := f.targetInformerFactory(cluster, config)
		if err != nil {
			// TODO: create also an event for cluster.
			klog.Errorf("Failed to create an informer for cluster %q: %v", cluster.Name, err)
			return
		}
		targetInformer := informer{
			controller: controller,
			store:      store,
			stopChan:   make(chan struct{}),
		}
		f.targetInformers[name] = targetInformer
		go targetInformer.controller.Run(targetInformer.stopChan)
	} else {
		// TODO: create also an event for cluster.
		klog.Errorf("Failed to create a client for cluster: %v", err)
	}
}

// Removes the cluster from federated informer.
func (f *federatedInformerImpl) deleteCluster(cluster *fedv1b1.KubeFedCluster) {
	f.Lock()
	defer f.Unlock()
	name := cluster.Name
	if targetInformer, found := f.targetInformers[name]; found {
		close(targetInformer.stopChan)
	}
	delete(f.targetInformers, name)
	delete(f.clusterClients, name)
}

// Returns a store created over all stores from target informers.
func (f *federatedInformerImpl) GetTargetStore() FederatedReadOnlyStore {
	return &federatedStoreImpl{
		federatedInformer: f,
	}
}

// Returns all items in the store.
func (fs *federatedStoreImpl) List() ([]FederatedObject, error) {
	fs.federatedInformer.Lock()
	defer fs.federatedInformer.Unlock()

	result := make([]FederatedObject, 0)
	for clusterName, targetInformer := range fs.federatedInformer.targetInformers {
		for _, value := range targetInformer.store.List() {
			result = append(result, FederatedObject{ClusterName: clusterName, Object: value})
		}
	}
	return result, nil
}

// Returns all items in the given cluster.
func (fs *federatedStoreImpl) ListFromCluster(clusterName string) ([]interface{}, error) {
	fs.federatedInformer.Lock()
	defer fs.federatedInformer.Unlock()

	result := make([]interface{}, 0)
	if targetInformer, found := fs.federatedInformer.targetInformers[clusterName]; found {
		values := targetInformer.store.List()
		result = append(result, values...)
	}
	return result, nil
}

// GetByKey returns the item stored under the given key in the specified cluster (if exist).
func (fs *federatedStoreImpl) GetByKey(clusterName string, key string) (interface{}, bool, error) {
	fs.federatedInformer.Lock()
	defer fs.federatedInformer.Unlock()
	if targetInformer, found := fs.federatedInformer.targetInformers[clusterName]; found {
		return targetInformer.store.GetByKey(key)
	}
	return nil, false, nil
}

// Returns the items stored under the given key in all clusters.
func (fs *federatedStoreImpl) GetFromAllClusters(key string) ([]FederatedObject, error) {
	fs.federatedInformer.Lock()
	defer fs.federatedInformer.Unlock()

	result := make([]FederatedObject, 0)
	for clusterName, targetInformer := range fs.federatedInformer.targetInformers {
		value, exist, err := targetInformer.store.GetByKey(key)
		if err != nil {
			return nil, err
		}
		if exist {
			result = append(result, FederatedObject{ClusterName: clusterName, Object: value})
		}
	}
	return result, nil
}

// GetKeyFor returns the key under which the item would be put in the store.
func (fs *federatedStoreImpl) GetKeyFor(item interface{}) string {
	// TODO: support other keying functions.
	key, _ := cache.DeletionHandlingMetaNamespaceKeyFunc(item)
	return key
}

// ClustersSynced checks whether stores for all clusters form the lists (and only these) are there and
// are synced.
func (fs *federatedStoreImpl) ClustersSynced(clusters []*fedv1b1.KubeFedCluster) bool {
	fs.federatedInformer.Lock()
	defer fs.federatedInformer.Unlock()

	if len(fs.federatedInformer.targetInformers) != len(clusters) {
		klog.V(4).Infof("The number of target informers mismatch with given clusters")
		return false
	}
	for _, cluster := range clusters {
		if targetInformer, found := fs.federatedInformer.targetInformers[cluster.Name]; found {
			if !targetInformer.controller.HasSynced() {
				klog.V(4).Infof("Informer of cluster %q not synced", cluster.Name)
				return false
			}
		} else {
			klog.V(4).Infof("Informer of cluster %q not found", cluster.Name)
			return false
		}
	}
	return true
}
