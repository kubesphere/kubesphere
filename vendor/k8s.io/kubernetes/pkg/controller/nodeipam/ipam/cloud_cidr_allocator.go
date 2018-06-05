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

package ipam

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	informers "k8s.io/client-go/informers/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"

	"k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	v1node "k8s.io/kubernetes/pkg/api/v1/node"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/cloudprovider/providers/gce"
	"k8s.io/kubernetes/pkg/controller"
	nodeutil "k8s.io/kubernetes/pkg/controller/util/node"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"
	utilnode "k8s.io/kubernetes/pkg/util/node"
	utiltaints "k8s.io/kubernetes/pkg/util/taints"
)

// nodeProcessingInfo tracks information related to current nodes in processing
type nodeProcessingInfo struct {
	retries int
}

// cloudCIDRAllocator allocates node CIDRs according to IP address aliases
// assigned by the cloud provider. In this case, the allocation and
// deallocation is delegated to the external provider, and the controller
// merely takes the assignment and updates the node spec.
type cloudCIDRAllocator struct {
	client clientset.Interface
	cloud  *gce.GCECloud

	// nodeLister is able to list/get nodes and is populated by the shared informer passed to
	// NewCloudCIDRAllocator.
	nodeLister corelisters.NodeLister
	// nodesSynced returns true if the node shared informer has been synced at least once.
	nodesSynced cache.InformerSynced

	// Channel that is used to pass updating Nodes to the background.
	// This increases the throughput of CIDR assignment by parallelization
	// and not blocking on long operations (which shouldn't be done from
	// event handlers anyway).
	nodeUpdateChannel chan string
	recorder          record.EventRecorder

	// Keep a set of nodes that are currectly being processed to avoid races in CIDR allocation
	lock              sync.Mutex
	nodesInProcessing map[string]*nodeProcessingInfo
}

var _ CIDRAllocator = (*cloudCIDRAllocator)(nil)

// NewCloudCIDRAllocator creates a new cloud CIDR allocator.
func NewCloudCIDRAllocator(client clientset.Interface, cloud cloudprovider.Interface, nodeInformer informers.NodeInformer) (CIDRAllocator, error) {
	if client == nil {
		glog.Fatalf("kubeClient is nil when starting NodeController")
	}

	eventBroadcaster := record.NewBroadcaster()
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "cidrAllocator"})
	eventBroadcaster.StartLogging(glog.Infof)
	glog.V(0).Infof("Sending events to api server.")
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: client.CoreV1().Events("")})

	gceCloud, ok := cloud.(*gce.GCECloud)
	if !ok {
		err := fmt.Errorf("cloudCIDRAllocator does not support %v provider", cloud.ProviderName())
		return nil, err
	}

	ca := &cloudCIDRAllocator{
		client:            client,
		cloud:             gceCloud,
		nodeLister:        nodeInformer.Lister(),
		nodesSynced:       nodeInformer.Informer().HasSynced,
		nodeUpdateChannel: make(chan string, cidrUpdateQueueSize),
		recorder:          recorder,
		nodesInProcessing: map[string]*nodeProcessingInfo{},
	}

	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: nodeutil.CreateAddNodeHandler(ca.AllocateOrOccupyCIDR),
		UpdateFunc: nodeutil.CreateUpdateNodeHandler(func(_, newNode *v1.Node) error {
			if newNode.Spec.PodCIDR == "" {
				return ca.AllocateOrOccupyCIDR(newNode)
			}
			// Even if PodCIDR is assigned, but NetworkUnavailable condition is
			// set to true, we need to process the node to set the condition.
			networkUnavailableTaint := &v1.Taint{Key: algorithm.TaintNodeNetworkUnavailable, Effect: v1.TaintEffectNoSchedule}
			_, cond := v1node.GetNodeCondition(&newNode.Status, v1.NodeNetworkUnavailable)
			if cond == nil || cond.Status != v1.ConditionFalse || utiltaints.TaintExists(newNode.Spec.Taints, networkUnavailableTaint) {
				return ca.AllocateOrOccupyCIDR(newNode)
			}
			return nil
		}),
		DeleteFunc: nodeutil.CreateDeleteNodeHandler(ca.ReleaseCIDR),
	})

	glog.V(0).Infof("Using cloud CIDR allocator (provider: %v)", cloud.ProviderName())
	return ca, nil
}

func (ca *cloudCIDRAllocator) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()

	glog.Infof("Starting cloud CIDR allocator")
	defer glog.Infof("Shutting down cloud CIDR allocator")

	if !controller.WaitForCacheSync("cidrallocator", stopCh, ca.nodesSynced) {
		return
	}

	for i := 0; i < cidrUpdateWorkers; i++ {
		go ca.worker(stopCh)
	}

	<-stopCh
}

func (ca *cloudCIDRAllocator) worker(stopChan <-chan struct{}) {
	for {
		select {
		case workItem, ok := <-ca.nodeUpdateChannel:
			if !ok {
				glog.Warning("Channel nodeCIDRUpdateChannel was unexpectedly closed")
				return
			}
			if err := ca.updateCIDRAllocation(workItem); err != nil {
				if ca.canRetry(workItem) {
					time.AfterFunc(updateRetryTimeout, func() {
						// Requeue the failed node for update again.
						ca.nodeUpdateChannel <- workItem
					})
					continue
				}
			}
			ca.removeNodeFromProcessing(workItem)
		case <-stopChan:
			return
		}
	}
}

func (ca *cloudCIDRAllocator) insertNodeToProcessing(nodeName string) bool {
	ca.lock.Lock()
	defer ca.lock.Unlock()
	if _, found := ca.nodesInProcessing[nodeName]; found {
		return false
	}
	ca.nodesInProcessing[nodeName] = &nodeProcessingInfo{}
	return true
}

func (ca *cloudCIDRAllocator) canRetry(nodeName string) bool {
	ca.lock.Lock()
	defer ca.lock.Unlock()
	count := ca.nodesInProcessing[nodeName].retries + 1
	if count > updateMaxRetries {
		return false
	}
	ca.nodesInProcessing[nodeName].retries = count
	return true
}

func (ca *cloudCIDRAllocator) removeNodeFromProcessing(nodeName string) {
	ca.lock.Lock()
	defer ca.lock.Unlock()
	delete(ca.nodesInProcessing, nodeName)
}

// WARNING: If you're adding any return calls or defer any more work from this
// function you have to make sure to update nodesInProcessing properly with the
// disposition of the node when the work is done.
func (ca *cloudCIDRAllocator) AllocateOrOccupyCIDR(node *v1.Node) error {
	if node == nil {
		return nil
	}
	if !ca.insertNodeToProcessing(node.Name) {
		glog.V(2).Infof("Node %v is already in a process of CIDR assignment.", node.Name)
		return nil
	}

	glog.V(4).Infof("Putting node %s into the work queue", node.Name)
	ca.nodeUpdateChannel <- node.Name
	return nil
}

// updateCIDRAllocation assigns CIDR to Node and sends an update to the API server.
func (ca *cloudCIDRAllocator) updateCIDRAllocation(nodeName string) error {
	node, err := ca.nodeLister.Get(nodeName)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil // node no longer available, skip processing
		}
		glog.Errorf("Failed while getting node %v for updating Node.Spec.PodCIDR: %v", nodeName, err)
		return err
	}

	cidrs, err := ca.cloud.AliasRanges(types.NodeName(nodeName))
	if err != nil {
		nodeutil.RecordNodeStatusChange(ca.recorder, node, "CIDRNotAvailable")
		return fmt.Errorf("failed to allocate cidr: %v", err)
	}
	if len(cidrs) == 0 {
		nodeutil.RecordNodeStatusChange(ca.recorder, node, "CIDRNotAvailable")
		return fmt.Errorf("failed to allocate cidr: Node %v has no CIDRs", node.Name)
	}
	_, cidr, err := net.ParseCIDR(cidrs[0])
	if err != nil {
		return fmt.Errorf("failed to parse string '%s' as a CIDR: %v", cidrs[0], err)
	}
	podCIDR := cidr.String()

	if node.Spec.PodCIDR == podCIDR {
		glog.V(4).Infof("Node %v already has allocated CIDR %v. It matches the proposed one.", node.Name, podCIDR)
		// We don't return here, in order to set the NetworkUnavailable condition later below.
	} else {
		if node.Spec.PodCIDR != "" {
			glog.Errorf("PodCIDR being reassigned! Node %v spec has %v, but cloud provider has assigned %v", node.Name, node.Spec.PodCIDR, podCIDR)
			// We fall through and set the CIDR despite this error. This
			// implements the same logic as implemented in the
			// rangeAllocator.
			//
			// See https://github.com/kubernetes/kubernetes/pull/42147#discussion_r103357248
		}
		for i := 0; i < cidrUpdateRetries; i++ {
			if err = utilnode.PatchNodeCIDR(ca.client, types.NodeName(node.Name), podCIDR); err == nil {
				glog.Infof("Set node %v PodCIDR to %v", node.Name, podCIDR)
				break
			}
		}
	}
	if err != nil {
		nodeutil.RecordNodeStatusChange(ca.recorder, node, "CIDRAssignmentFailed")
		glog.Errorf("Failed to update node %v PodCIDR to %v after multiple attempts: %v", node.Name, podCIDR, err)
		return err
	}

	err = utilnode.SetNodeCondition(ca.client, types.NodeName(node.Name), v1.NodeCondition{
		Type:               v1.NodeNetworkUnavailable,
		Status:             v1.ConditionFalse,
		Reason:             "RouteCreated",
		Message:            "NodeController create implicit route",
		LastTransitionTime: metav1.Now(),
	})
	if err != nil {
		glog.Errorf("Error setting route status for node %v: %v", node.Name, err)
	}
	return err
}

func (ca *cloudCIDRAllocator) ReleaseCIDR(node *v1.Node) error {
	glog.V(2).Infof("Node %v PodCIDR (%v) will be released by external cloud provider (not managed by controller)",
		node.Name, node.Spec.PodCIDR)
	return nil
}
