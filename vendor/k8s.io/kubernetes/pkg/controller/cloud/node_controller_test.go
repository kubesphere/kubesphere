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

package cloud

import (
	"errors"
	"testing"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/cloudprovider"
	fakecloud "k8s.io/kubernetes/pkg/cloudprovider/providers/fake"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/controller/testutil"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"

	"github.com/golang/glog"
	"github.com/stretchr/testify/assert"
)

func TestEnsureNodeExistsByProviderID(t *testing.T) {

	testCases := []struct {
		testName           string
		node               *v1.Node
		expectedCalls      []string
		expectedNodeExists bool
		hasInstanceID      bool
		existsByProviderID bool
		nodeNameErr        error
		providerIDErr      error
	}{
		{
			testName:           "node exists by provider id",
			existsByProviderID: true,
			providerIDErr:      nil,
			hasInstanceID:      true,
			nodeNameErr:        errors.New("unimplemented"),
			expectedCalls:      []string{"instance-exists-by-provider-id"},
			expectedNodeExists: true,
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node0",
				},
				Spec: v1.NodeSpec{
					ProviderID: "node0",
				},
			},
		},
		{
			testName:           "does not exist by provider id",
			existsByProviderID: false,
			providerIDErr:      nil,
			hasInstanceID:      true,
			nodeNameErr:        errors.New("unimplemented"),
			expectedCalls:      []string{"instance-exists-by-provider-id"},
			expectedNodeExists: false,
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node0",
				},
				Spec: v1.NodeSpec{
					ProviderID: "node0",
				},
			},
		},
		{
			testName:           "exists by instance id",
			existsByProviderID: true,
			providerIDErr:      nil,
			hasInstanceID:      true,
			nodeNameErr:        nil,
			expectedCalls:      []string{"instance-id", "instance-exists-by-provider-id"},
			expectedNodeExists: true,
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node0",
				},
			},
		},
		{
			testName:           "does not exist by no instance id",
			existsByProviderID: true,
			providerIDErr:      nil,
			hasInstanceID:      false,
			nodeNameErr:        cloudprovider.InstanceNotFound,
			expectedCalls:      []string{"instance-id"},
			expectedNodeExists: false,
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node0",
				},
			},
		},
		{
			testName:           "provider id returns error",
			existsByProviderID: false,
			providerIDErr:      errors.New("unimplemented"),
			hasInstanceID:      true,
			nodeNameErr:        cloudprovider.InstanceNotFound,
			expectedCalls:      []string{"instance-exists-by-provider-id"},
			expectedNodeExists: false,
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node0",
				},
				Spec: v1.NodeSpec{
					ProviderID: "node0",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			fc := &fakecloud.FakeCloud{
				ExistsByProviderID: tc.existsByProviderID,
				Err:                tc.nodeNameErr,
				ErrByProviderID:    tc.providerIDErr,
			}

			if tc.hasInstanceID {
				fc.ExtID = map[types.NodeName]string{
					types.NodeName(tc.node.Name): "provider-id://a",
				}
			}

			instances, _ := fc.Instances()
			exists, err := ensureNodeExistsByProviderID(instances, tc.node)
			assert.Equal(t, err, tc.providerIDErr)

			assert.EqualValues(t, tc.expectedCalls, fc.Calls,
				"expected cloud provider methods `%v` to be called but `%v` was called ",
				tc.expectedCalls, fc.Calls)

			assert.Equal(t, tc.expectedNodeExists, exists,
				"expected exists to be `%t` but got `%t`",
				tc.existsByProviderID, exists)
		})
	}

}

func TestNodeShutdown(t *testing.T) {

	testCases := []struct {
		testName           string
		node               *v1.Node
		existsByProviderID bool
		shutdown           bool
	}{
		{
			testName:           "node shutdowned add taint",
			existsByProviderID: true,
			shutdown:           true,
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "node0",
					CreationTimestamp: metav1.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				Spec: v1.NodeSpec{
					ProviderID: "node0",
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:               v1.NodeReady,
							Status:             v1.ConditionUnknown,
							LastHeartbeatTime:  metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
							LastTransitionTime: metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
						},
					},
				},
			},
		},
		{
			testName:           "node started after shutdown remove taint",
			existsByProviderID: true,
			shutdown:           false,
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "node0",
					CreationTimestamp: metav1.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				Spec: v1.NodeSpec{
					ProviderID: "node0",
					Taints: []v1.Taint{
						{
							Key:    algorithm.TaintNodeShutdown,
							Effect: v1.TaintEffectNoSchedule,
						},
					},
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:               v1.NodeReady,
							Status:             v1.ConditionTrue,
							LastHeartbeatTime:  metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
							LastTransitionTime: metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			fc := &fakecloud.FakeCloud{
				ExistsByProviderID: tc.existsByProviderID,
				NodeShutdown:       tc.shutdown,
			}
			fnh := &testutil.FakeNodeHandler{
				Existing:      []*v1.Node{tc.node},
				Clientset:     fake.NewSimpleClientset(),
				PatchWaitChan: make(chan struct{}),
			}

			factory := informers.NewSharedInformerFactory(fnh, controller.NoResyncPeriodFunc())

			eventBroadcaster := record.NewBroadcaster()
			cloudNodeController := &CloudNodeController{
				kubeClient:                fnh,
				nodeInformer:              factory.Core().V1().Nodes(),
				cloud:                     fc,
				nodeMonitorPeriod:         1 * time.Second,
				recorder:                  eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "cloud-node-controller"}),
				nodeStatusUpdateFrequency: 1 * time.Second,
			}
			eventBroadcaster.StartLogging(glog.Infof)

			cloudNodeController.Run()

			select {
			case <-fnh.PatchWaitChan:
			case <-time.After(1 * time.Second):
				t.Errorf("Timed out waiting %v for node to be updated", wait.ForeverTestTimeout)
			}

			assert.Equal(t, 1, len(fnh.UpdatedNodes), "Node was not updated")
			if tc.shutdown {
				assert.Equal(t, 1, len(fnh.UpdatedNodes[0].Spec.Taints), "Node Taint was not added")
				assert.Equal(t, "node.cloudprovider.kubernetes.io/shutdown", fnh.UpdatedNodes[0].Spec.Taints[0].Key, "Node Taint key is not correct")
			} else {
				assert.Equal(t, 0, len(fnh.UpdatedNodes[0].Spec.Taints), "Node Taint was not removed after node is back in ready state")
			}

		})
	}

}

// This test checks that the node is deleted when kubelet stops reporting
// and cloud provider says node is gone
func TestNodeDeleted(t *testing.T) {
	pod0 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "pod0",
		},
		Spec: v1.PodSpec{
			NodeName: "node0",
		},
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				},
			},
		},
	}

	pod1 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "pod1",
		},
		Spec: v1.PodSpec{
			NodeName: "node0",
		},
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				},
			},
		},
	}

	fnh := &testutil.FakeNodeHandler{
		Existing: []*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "node0",
					CreationTimestamp: metav1.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:               v1.NodeReady,
							Status:             v1.ConditionUnknown,
							LastHeartbeatTime:  metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
							LastTransitionTime: metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
						},
					},
				},
			},
		},
		Clientset:      fake.NewSimpleClientset(&v1.PodList{Items: []v1.Pod{*pod0, *pod1}}),
		DeleteWaitChan: make(chan struct{}),
	}

	factory := informers.NewSharedInformerFactory(fnh, controller.NoResyncPeriodFunc())

	eventBroadcaster := record.NewBroadcaster()
	cloudNodeController := &CloudNodeController{
		kubeClient:   fnh,
		nodeInformer: factory.Core().V1().Nodes(),
		cloud: &fakecloud.FakeCloud{
			ExistsByProviderID: false,
			Err:                nil,
		},
		nodeMonitorPeriod:         1 * time.Second,
		recorder:                  eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "cloud-node-controller"}),
		nodeStatusUpdateFrequency: 1 * time.Second,
	}
	eventBroadcaster.StartLogging(glog.Infof)

	cloudNodeController.Run()

	select {
	case <-fnh.DeleteWaitChan:
	case <-time.After(wait.ForeverTestTimeout):
		t.Errorf("Timed out waiting %v for node to be deleted", wait.ForeverTestTimeout)
	}

	assert.Equal(t, 1, len(fnh.DeletedNodes), "Node was not deleted")
	assert.Equal(t, "node0", fnh.DeletedNodes[0].Name, "Node was not deleted")
}

// This test checks that a node with the external cloud provider taint is cloudprovider initialized
func TestNodeInitialized(t *testing.T) {
	fnh := &testutil.FakeNodeHandler{
		Existing: []*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "node0",
					CreationTimestamp: metav1.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:               v1.NodeReady,
							Status:             v1.ConditionUnknown,
							LastHeartbeatTime:  metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
							LastTransitionTime: metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
						},
					},
				},
				Spec: v1.NodeSpec{
					Taints: []v1.Taint{
						{
							Key:    algorithm.TaintExternalCloudProvider,
							Value:  "true",
							Effect: v1.TaintEffectNoSchedule,
						},
					},
				},
			},
		},
		Clientset:      fake.NewSimpleClientset(&v1.PodList{}),
		DeleteWaitChan: make(chan struct{}),
	}

	factory := informers.NewSharedInformerFactory(fnh, controller.NoResyncPeriodFunc())

	fakeCloud := &fakecloud.FakeCloud{
		InstanceTypes: map[types.NodeName]string{
			types.NodeName("node0"): "t1.micro",
		},
		Addresses: []v1.NodeAddress{
			{
				Type:    v1.NodeHostName,
				Address: "node0.cloud.internal",
			},
			{
				Type:    v1.NodeInternalIP,
				Address: "10.0.0.1",
			},
			{
				Type:    v1.NodeExternalIP,
				Address: "132.143.154.163",
			},
		},
		Err: nil,
	}

	eventBroadcaster := record.NewBroadcaster()
	cloudNodeController := &CloudNodeController{
		kubeClient:                fnh,
		nodeInformer:              factory.Core().V1().Nodes(),
		cloud:                     fakeCloud,
		nodeMonitorPeriod:         1 * time.Second,
		recorder:                  eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "cloud-node-controller"}),
		nodeStatusUpdateFrequency: 1 * time.Second,
	}
	eventBroadcaster.StartLogging(glog.Infof)

	cloudNodeController.AddCloudNode(fnh.Existing[0])

	assert.Equal(t, 1, len(fnh.UpdatedNodes), "Node was not updated")
	assert.Equal(t, "node0", fnh.UpdatedNodes[0].Name, "Node was not updated")
	assert.Equal(t, 0, len(fnh.UpdatedNodes[0].Spec.Taints), "Node Taint was not removed after cloud init")
}

// This test checks that a node without the external cloud provider taint are NOT cloudprovider initialized
func TestNodeIgnored(t *testing.T) {
	fnh := &testutil.FakeNodeHandler{
		Existing: []*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "node0",
					CreationTimestamp: metav1.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:               v1.NodeReady,
							Status:             v1.ConditionUnknown,
							LastHeartbeatTime:  metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
							LastTransitionTime: metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
						},
					},
				},
			},
		},
		Clientset:      fake.NewSimpleClientset(&v1.PodList{}),
		DeleteWaitChan: make(chan struct{}),
	}

	factory := informers.NewSharedInformerFactory(fnh, controller.NoResyncPeriodFunc())

	fakeCloud := &fakecloud.FakeCloud{
		InstanceTypes: map[types.NodeName]string{
			types.NodeName("node0"): "t1.micro",
		},
		Addresses: []v1.NodeAddress{
			{
				Type:    v1.NodeHostName,
				Address: "node0.cloud.internal",
			},
			{
				Type:    v1.NodeInternalIP,
				Address: "10.0.0.1",
			},
			{
				Type:    v1.NodeExternalIP,
				Address: "132.143.154.163",
			},
		},
		Err: nil,
	}

	eventBroadcaster := record.NewBroadcaster()
	cloudNodeController := &CloudNodeController{
		kubeClient:        fnh,
		nodeInformer:      factory.Core().V1().Nodes(),
		cloud:             fakeCloud,
		nodeMonitorPeriod: 5 * time.Second,
		recorder:          eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "cloud-node-controller"}),
	}
	eventBroadcaster.StartLogging(glog.Infof)

	cloudNodeController.AddCloudNode(fnh.Existing[0])
	assert.Equal(t, 0, len(fnh.UpdatedNodes), "Node was wrongly updated")

}

// This test checks that a node with the external cloud provider taint is cloudprovider initialized and
// the GCE route condition is added if cloudprovider is GCE
func TestGCECondition(t *testing.T) {
	fnh := &testutil.FakeNodeHandler{
		Existing: []*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "node0",
					CreationTimestamp: metav1.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:               v1.NodeReady,
							Status:             v1.ConditionUnknown,
							LastHeartbeatTime:  metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
							LastTransitionTime: metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
						},
					},
				},
				Spec: v1.NodeSpec{
					Taints: []v1.Taint{
						{
							Key:    algorithm.TaintExternalCloudProvider,
							Value:  "true",
							Effect: v1.TaintEffectNoSchedule,
						},
					},
				},
			},
		},
		Clientset:      fake.NewSimpleClientset(&v1.PodList{}),
		DeleteWaitChan: make(chan struct{}),
	}

	factory := informers.NewSharedInformerFactory(fnh, controller.NoResyncPeriodFunc())

	fakeCloud := &fakecloud.FakeCloud{
		InstanceTypes: map[types.NodeName]string{
			types.NodeName("node0"): "t1.micro",
		},
		Addresses: []v1.NodeAddress{
			{
				Type:    v1.NodeHostName,
				Address: "node0.cloud.internal",
			},
			{
				Type:    v1.NodeInternalIP,
				Address: "10.0.0.1",
			},
			{
				Type:    v1.NodeExternalIP,
				Address: "132.143.154.163",
			},
		},
		Provider: "gce",
		Err:      nil,
	}

	eventBroadcaster := record.NewBroadcaster()
	cloudNodeController := &CloudNodeController{
		kubeClient:        fnh,
		nodeInformer:      factory.Core().V1().Nodes(),
		cloud:             fakeCloud,
		nodeMonitorPeriod: 1 * time.Second,
		recorder:          eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "cloud-node-controller"}),
	}
	eventBroadcaster.StartLogging(glog.Infof)

	cloudNodeController.AddCloudNode(fnh.Existing[0])

	assert.Equal(t, 1, len(fnh.UpdatedNodes), "Node was not updated")
	assert.Equal(t, "node0", fnh.UpdatedNodes[0].Name, "Node was not updated")
	assert.Equal(t, 2, len(fnh.UpdatedNodes[0].Status.Conditions), "No new conditions were added for GCE")

	conditionAdded := false
	for _, cond := range fnh.UpdatedNodes[0].Status.Conditions {
		if cond.Status == "True" && cond.Type == "NetworkUnavailable" && cond.Reason == "NoRouteCreated" {
			conditionAdded = true
		}
	}

	assert.True(t, conditionAdded, "Network Route Condition for GCE not added by external cloud initializer")
}

// This test checks that a node with the external cloud provider taint is cloudprovider initialized and
// and that zone labels are added correctly
func TestZoneInitialized(t *testing.T) {
	fnh := &testutil.FakeNodeHandler{
		Existing: []*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "node0",
					CreationTimestamp: metav1.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
					Labels:            map[string]string{},
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:               v1.NodeReady,
							Status:             v1.ConditionUnknown,
							LastHeartbeatTime:  metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
							LastTransitionTime: metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
						},
					},
				},
				Spec: v1.NodeSpec{
					Taints: []v1.Taint{
						{
							Key:    algorithm.TaintExternalCloudProvider,
							Value:  "true",
							Effect: v1.TaintEffectNoSchedule,
						},
					},
				},
			},
		},
		Clientset:      fake.NewSimpleClientset(&v1.PodList{}),
		DeleteWaitChan: make(chan struct{}),
	}

	factory := informers.NewSharedInformerFactory(fnh, controller.NoResyncPeriodFunc())

	fakeCloud := &fakecloud.FakeCloud{
		InstanceTypes: map[types.NodeName]string{
			types.NodeName("node0"): "t1.micro",
		},
		Addresses: []v1.NodeAddress{
			{
				Type:    v1.NodeHostName,
				Address: "node0.cloud.internal",
			},
			{
				Type:    v1.NodeInternalIP,
				Address: "10.0.0.1",
			},
			{
				Type:    v1.NodeExternalIP,
				Address: "132.143.154.163",
			},
		},
		Provider: "aws",
		Zone: cloudprovider.Zone{
			FailureDomain: "us-west-1a",
			Region:        "us-west",
		},
		Err: nil,
	}

	eventBroadcaster := record.NewBroadcaster()
	cloudNodeController := &CloudNodeController{
		kubeClient:        fnh,
		nodeInformer:      factory.Core().V1().Nodes(),
		cloud:             fakeCloud,
		nodeMonitorPeriod: 5 * time.Second,
		recorder:          eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "cloud-node-controller"}),
	}
	eventBroadcaster.StartLogging(glog.Infof)

	cloudNodeController.AddCloudNode(fnh.Existing[0])

	assert.Equal(t, 1, len(fnh.UpdatedNodes), "Node was not updated")
	assert.Equal(t, "node0", fnh.UpdatedNodes[0].Name, "Node was not updated")
	assert.Equal(t, 2, len(fnh.UpdatedNodes[0].ObjectMeta.Labels),
		"Node label for Region and Zone were not set")
	assert.Equal(t, "us-west", fnh.UpdatedNodes[0].ObjectMeta.Labels[kubeletapis.LabelZoneRegion],
		"Node Region not correctly updated")
	assert.Equal(t, "us-west-1a", fnh.UpdatedNodes[0].ObjectMeta.Labels[kubeletapis.LabelZoneFailureDomain],
		"Node FailureDomain not correctly updated")
}

// This test checks that a node with the external cloud provider taint is cloudprovider initialized and
// and nodeAddresses are updated from the cloudprovider
func TestNodeAddresses(t *testing.T) {
	fnh := &testutil.FakeNodeHandler{
		Existing: []*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "node0",
					CreationTimestamp: metav1.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
					Labels:            map[string]string{},
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:               v1.NodeReady,
							Status:             v1.ConditionUnknown,
							LastHeartbeatTime:  metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
							LastTransitionTime: metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
						},
					},
				},
				Spec: v1.NodeSpec{
					Taints: []v1.Taint{
						{
							Key:    "ImproveCoverageTaint",
							Value:  "true",
							Effect: v1.TaintEffectNoSchedule,
						},
						{
							Key:    algorithm.TaintExternalCloudProvider,
							Value:  "true",
							Effect: v1.TaintEffectNoSchedule,
						},
					},
				},
			},
		},
		Clientset:      fake.NewSimpleClientset(&v1.PodList{}),
		DeleteWaitChan: make(chan struct{}),
	}

	factory := informers.NewSharedInformerFactory(fnh, controller.NoResyncPeriodFunc())

	fakeCloud := &fakecloud.FakeCloud{
		InstanceTypes: map[types.NodeName]string{},
		Addresses: []v1.NodeAddress{
			{
				Type:    v1.NodeHostName,
				Address: "node0.cloud.internal",
			},
			{
				Type:    v1.NodeInternalIP,
				Address: "10.0.0.1",
			},
			{
				Type:    v1.NodeExternalIP,
				Address: "132.143.154.163",
			},
		},
		Provider: "aws",
		Zone: cloudprovider.Zone{
			FailureDomain: "us-west-1a",
			Region:        "us-west",
		},
		ExistsByProviderID: true,
		Err:                nil,
	}

	eventBroadcaster := record.NewBroadcaster()
	cloudNodeController := &CloudNodeController{
		kubeClient:                fnh,
		nodeInformer:              factory.Core().V1().Nodes(),
		cloud:                     fakeCloud,
		nodeMonitorPeriod:         5 * time.Second,
		nodeStatusUpdateFrequency: 1 * time.Second,
		recorder:                  eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "cloud-node-controller"}),
	}
	eventBroadcaster.StartLogging(glog.Infof)

	cloudNodeController.AddCloudNode(fnh.Existing[0])

	assert.Equal(t, 1, len(fnh.UpdatedNodes), "Node was not updated")
	assert.Equal(t, "node0", fnh.UpdatedNodes[0].Name, "Node was not updated")
	assert.Equal(t, 3, len(fnh.UpdatedNodes[0].Status.Addresses), "Node status not updated")

	fakeCloud.Addresses = []v1.NodeAddress{
		{
			Type:    v1.NodeHostName,
			Address: "node0.cloud.internal",
		},
		{
			Type:    v1.NodeInternalIP,
			Address: "10.0.0.1",
		},
	}

	cloudNodeController.Run()

	<-time.After(2 * time.Second)

	updatedNodes := fnh.GetUpdatedNodesCopy()

	assert.Equal(t, 2, len(updatedNodes[0].Status.Addresses), "Node Addresses not correctly updated")

}

// This test checks that a node with the external cloud provider taint is cloudprovider initialized and
// and the provided node ip is validated with the cloudprovider and nodeAddresses are updated from the cloudprovider
func TestNodeProvidedIPAddresses(t *testing.T) {
	fnh := &testutil.FakeNodeHandler{
		Existing: []*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "node0",
					CreationTimestamp: metav1.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
					Labels:            map[string]string{},
					Annotations: map[string]string{
						kubeletapis.AnnotationProvidedIPAddr: "10.0.0.1",
					},
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:               v1.NodeReady,
							Status:             v1.ConditionUnknown,
							LastHeartbeatTime:  metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
							LastTransitionTime: metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
						},
					},
					Addresses: []v1.NodeAddress{
						{
							Type:    v1.NodeHostName,
							Address: "node0.cloud.internal",
						},
					},
				},
				Spec: v1.NodeSpec{
					Taints: []v1.Taint{
						{
							Key:    "ImproveCoverageTaint",
							Value:  "true",
							Effect: v1.TaintEffectNoSchedule,
						},
						{
							Key:    algorithm.TaintExternalCloudProvider,
							Value:  "true",
							Effect: v1.TaintEffectNoSchedule,
						},
					},
					ProviderID: "node0.aws.12345",
				},
			},
		},
		Clientset:      fake.NewSimpleClientset(&v1.PodList{}),
		DeleteWaitChan: make(chan struct{}),
	}

	factory := informers.NewSharedInformerFactory(fnh, controller.NoResyncPeriodFunc())

	fakeCloud := &fakecloud.FakeCloud{
		InstanceTypes: map[types.NodeName]string{
			types.NodeName("node0"):           "t1.micro",
			types.NodeName("node0.aws.12345"): "t2.macro",
		},
		Addresses: []v1.NodeAddress{
			{
				Type:    v1.NodeInternalIP,
				Address: "10.0.0.1",
			},
			{
				Type:    v1.NodeExternalIP,
				Address: "132.143.154.163",
			},
		},
		Provider: "aws",
		Zone: cloudprovider.Zone{
			FailureDomain: "us-west-1a",
			Region:        "us-west",
		},
		ExistsByProviderID: true,
		Err:                nil,
	}

	eventBroadcaster := record.NewBroadcaster()
	cloudNodeController := &CloudNodeController{
		kubeClient:                fnh,
		nodeInformer:              factory.Core().V1().Nodes(),
		cloud:                     fakeCloud,
		nodeMonitorPeriod:         5 * time.Second,
		nodeStatusUpdateFrequency: 1 * time.Second,
		recorder:                  eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "cloud-node-controller"}),
	}
	eventBroadcaster.StartLogging(glog.Infof)

	cloudNodeController.AddCloudNode(fnh.Existing[0])

	assert.Equal(t, 1, len(fnh.UpdatedNodes), "Node was not updated")
	assert.Equal(t, "node0", fnh.UpdatedNodes[0].Name, "Node was not updated")
	assert.Equal(t, 1, len(fnh.UpdatedNodes[0].Status.Addresses), "Node status unexpectedly updated")

	cloudNodeController.Run()

	<-time.After(2 * time.Second)

	updatedNodes := fnh.GetUpdatedNodesCopy()

	assert.Equal(t, 1, len(updatedNodes[0].Status.Addresses), 1, "Node Addresses not correctly updated")
	assert.Equal(t, "10.0.0.1", updatedNodes[0].Status.Addresses[0].Address, "Node Addresses not correctly updated")
}

// Tests that node address changes are detected correctly
func TestNodeAddressesChangeDetected(t *testing.T) {
	addressSet1 := []v1.NodeAddress{
		{
			Type:    v1.NodeInternalIP,
			Address: "10.0.0.1",
		},
		{
			Type:    v1.NodeExternalIP,
			Address: "132.143.154.163",
		},
	}
	addressSet2 := []v1.NodeAddress{
		{
			Type:    v1.NodeInternalIP,
			Address: "10.0.0.1",
		},
		{
			Type:    v1.NodeExternalIP,
			Address: "132.143.154.163",
		},
	}

	assert.False(t, nodeAddressesChangeDetected(addressSet1, addressSet2),
		"Node address changes are not detected correctly")

	addressSet1 = []v1.NodeAddress{
		{
			Type:    v1.NodeInternalIP,
			Address: "10.0.0.1",
		},
		{
			Type:    v1.NodeExternalIP,
			Address: "132.143.154.164",
		},
	}
	addressSet2 = []v1.NodeAddress{
		{
			Type:    v1.NodeInternalIP,
			Address: "10.0.0.1",
		},
		{
			Type:    v1.NodeExternalIP,
			Address: "132.143.154.163",
		},
	}

	assert.True(t, nodeAddressesChangeDetected(addressSet1, addressSet2),
		"Node address changes are not detected correctly")

	addressSet1 = []v1.NodeAddress{
		{
			Type:    v1.NodeInternalIP,
			Address: "10.0.0.1",
		},
		{
			Type:    v1.NodeExternalIP,
			Address: "132.143.154.164",
		},
		{
			Type:    v1.NodeHostName,
			Address: "hostname.zone.region.aws.test",
		},
	}
	addressSet2 = []v1.NodeAddress{
		{
			Type:    v1.NodeInternalIP,
			Address: "10.0.0.1",
		},
		{
			Type:    v1.NodeExternalIP,
			Address: "132.143.154.164",
		},
	}

	assert.True(t, nodeAddressesChangeDetected(addressSet1, addressSet2),
		"Node address changes are not detected correctly")

	addressSet1 = []v1.NodeAddress{
		{
			Type:    v1.NodeInternalIP,
			Address: "10.0.0.1",
		},
		{
			Type:    v1.NodeExternalIP,
			Address: "132.143.154.164",
		},
	}
	addressSet2 = []v1.NodeAddress{
		{
			Type:    v1.NodeInternalIP,
			Address: "10.0.0.1",
		},
		{
			Type:    v1.NodeExternalIP,
			Address: "132.143.154.164",
		},
		{
			Type:    v1.NodeHostName,
			Address: "hostname.zone.region.aws.test",
		},
	}

	assert.True(t, nodeAddressesChangeDetected(addressSet1, addressSet2),
		"Node address changes are not detected correctly")

	addressSet1 = []v1.NodeAddress{
		{
			Type:    v1.NodeExternalIP,
			Address: "10.0.0.1",
		},
		{
			Type:    v1.NodeInternalIP,
			Address: "132.143.154.163",
		},
	}
	addressSet2 = []v1.NodeAddress{
		{
			Type:    v1.NodeInternalIP,
			Address: "10.0.0.1",
		},
		{
			Type:    v1.NodeExternalIP,
			Address: "132.143.154.163",
		},
	}

	assert.True(t, nodeAddressesChangeDetected(addressSet1, addressSet2),
		"Node address changes are not detected correctly")
}

// This test checks that a node with the external cloud provider taint is cloudprovider initialized and
// and node addresses will not be updated when node isn't present according to the cloudprovider
func TestNodeAddressesNotUpdate(t *testing.T) {
	fnh := &testutil.FakeNodeHandler{
		Existing: []*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "node0",
					CreationTimestamp: metav1.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
					Labels:            map[string]string{},
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:               v1.NodeReady,
							Status:             v1.ConditionUnknown,
							LastHeartbeatTime:  metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
							LastTransitionTime: metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
						},
					},
				},
				Spec: v1.NodeSpec{
					Taints: []v1.Taint{
						{
							Key:    "ImproveCoverageTaint",
							Value:  "true",
							Effect: v1.TaintEffectNoSchedule,
						},
					},
				},
			},
		},
	}

	factory := informers.NewSharedInformerFactory(fnh, controller.NoResyncPeriodFunc())

	fakeCloud := &fakecloud.FakeCloud{
		InstanceTypes: map[types.NodeName]string{},
		Addresses: []v1.NodeAddress{
			{
				Type:    v1.NodeHostName,
				Address: "node0.cloud.internal",
			},
			{
				Type:    v1.NodeInternalIP,
				Address: "10.0.0.1",
			},
			{
				Type:    v1.NodeExternalIP,
				Address: "132.143.154.163",
			},
		},
		ExistsByProviderID: false,
		Err:                nil,
	}

	cloudNodeController := &CloudNodeController{
		kubeClient:   fnh,
		nodeInformer: factory.Core().V1().Nodes(),
		cloud:        fakeCloud,
	}

	cloudNodeController.updateNodeAddress(fnh.Existing[0], fakeCloud)

	if len(fnh.UpdatedNodes) != 0 {
		t.Errorf("Node was not correctly updated, the updated len(nodes) got: %v, wanted=0", len(fnh.UpdatedNodes))
	}
}

// This test checks that a node is set with the correct providerID
func TestNodeProviderID(t *testing.T) {
	fnh := &testutil.FakeNodeHandler{
		Existing: []*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "node0",
					CreationTimestamp: metav1.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
					Labels:            map[string]string{},
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:               v1.NodeReady,
							Status:             v1.ConditionUnknown,
							LastHeartbeatTime:  metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
							LastTransitionTime: metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
						},
					},
				},
				Spec: v1.NodeSpec{
					Taints: []v1.Taint{
						{
							Key:    "ImproveCoverageTaint",
							Value:  "true",
							Effect: v1.TaintEffectNoSchedule,
						},
						{
							Key:    algorithm.TaintExternalCloudProvider,
							Value:  "true",
							Effect: v1.TaintEffectNoSchedule,
						},
					},
				},
			},
		},
		Clientset:      fake.NewSimpleClientset(&v1.PodList{}),
		DeleteWaitChan: make(chan struct{}),
	}

	factory := informers.NewSharedInformerFactory(fnh, controller.NoResyncPeriodFunc())

	fakeCloud := &fakecloud.FakeCloud{
		InstanceTypes: map[types.NodeName]string{},
		Addresses: []v1.NodeAddress{
			{
				Type:    v1.NodeHostName,
				Address: "node0.cloud.internal",
			},
			{
				Type:    v1.NodeInternalIP,
				Address: "10.0.0.1",
			},
			{
				Type:    v1.NodeExternalIP,
				Address: "132.143.154.163",
			},
		},
		Provider: "test",
		ExtID: map[types.NodeName]string{
			types.NodeName("node0"): "12345",
		},
		Err: nil,
	}

	eventBroadcaster := record.NewBroadcaster()
	cloudNodeController := &CloudNodeController{
		kubeClient:                fnh,
		nodeInformer:              factory.Core().V1().Nodes(),
		cloud:                     fakeCloud,
		nodeMonitorPeriod:         5 * time.Second,
		nodeStatusUpdateFrequency: 1 * time.Second,
		recorder:                  eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "cloud-node-controller"}),
	}
	eventBroadcaster.StartLogging(glog.Infof)

	cloudNodeController.AddCloudNode(fnh.Existing[0])

	assert.Equal(t, 1, len(fnh.UpdatedNodes), "Node was not updated")
	assert.Equal(t, "node0", fnh.UpdatedNodes[0].Name, "Node was not updated")
	assert.Equal(t, "test://12345", fnh.UpdatedNodes[0].Spec.ProviderID, "Node ProviderID not set correctly")
}

// This test checks that a node's provider ID will not be overwritten
func TestNodeProviderIDAlreadySet(t *testing.T) {
	fnh := &testutil.FakeNodeHandler{
		Existing: []*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "node0",
					CreationTimestamp: metav1.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
					Labels:            map[string]string{},
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:               v1.NodeReady,
							Status:             v1.ConditionUnknown,
							LastHeartbeatTime:  metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
							LastTransitionTime: metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
						},
					},
				},
				Spec: v1.NodeSpec{
					ProviderID: "test-provider-id",
					Taints: []v1.Taint{
						{
							Key:    "ImproveCoverageTaint",
							Value:  "true",
							Effect: v1.TaintEffectNoSchedule,
						},
						{
							Key:    algorithm.TaintExternalCloudProvider,
							Value:  "true",
							Effect: v1.TaintEffectNoSchedule,
						},
					},
				},
			},
		},
		Clientset:      fake.NewSimpleClientset(&v1.PodList{}),
		DeleteWaitChan: make(chan struct{}),
	}

	factory := informers.NewSharedInformerFactory(fnh, controller.NoResyncPeriodFunc())

	fakeCloud := &fakecloud.FakeCloud{
		InstanceTypes: map[types.NodeName]string{},
		Addresses: []v1.NodeAddress{
			{
				Type:    v1.NodeHostName,
				Address: "node0.cloud.internal",
			},
			{
				Type:    v1.NodeInternalIP,
				Address: "10.0.0.1",
			},
			{
				Type:    v1.NodeExternalIP,
				Address: "132.143.154.163",
			},
		},
		Provider: "test",
		ExtID: map[types.NodeName]string{
			types.NodeName("node0"): "12345",
		},
		Err: nil,
	}

	eventBroadcaster := record.NewBroadcaster()
	cloudNodeController := &CloudNodeController{
		kubeClient:                fnh,
		nodeInformer:              factory.Core().V1().Nodes(),
		cloud:                     fakeCloud,
		nodeMonitorPeriod:         5 * time.Second,
		nodeStatusUpdateFrequency: 1 * time.Second,
		recorder:                  eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "cloud-node-controller"}),
	}
	eventBroadcaster.StartLogging(glog.Infof)

	cloudNodeController.AddCloudNode(fnh.Existing[0])

	assert.Equal(t, 1, len(fnh.UpdatedNodes), "Node was not updated")
	assert.Equal(t, "node0", fnh.UpdatedNodes[0].Name, "Node was not updated")
	// CCM node controller should not overwrite provider if it's already set
	assert.Equal(t, "test-provider-id", fnh.UpdatedNodes[0].Spec.ProviderID, "Node ProviderID not set correctly")
}
