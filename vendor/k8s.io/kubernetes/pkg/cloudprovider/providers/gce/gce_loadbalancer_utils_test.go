/*
Copyright 2017 The Kubernetes Authors.

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

// This file contains shared functions and variables to set up for tests for
// ExternalLoadBalancer and InternalLoadBalancers. It currently cannot live in a
// separate package from GCE because then it would cause a circular import.

package gce

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	compute "google.golang.org/api/compute/v1"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	v1_service "k8s.io/kubernetes/pkg/api/v1/service"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/cloudprovider/providers/gce/cloud"
	"k8s.io/kubernetes/pkg/cloudprovider/providers/gce/cloud/meta"
	"k8s.io/kubernetes/pkg/cloudprovider/providers/gce/cloud/mock"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
)

// TODO(yankaiz): Create shared error types for both test/non-test codes.
const (
	eventReasonManualChange = "LoadBalancerManualChange"
	eventMsgFirewallChange  = "Firewall change required by network admin"
	errPrefixGetTargetPool  = "error getting load balancer's target pool:"
	errStrLbNoHosts         = "Cannot EnsureLoadBalancer() with no hosts"
	wrongTier               = "SupremeLuxury"
	errStrUnsupportedTier   = "unsupported network tier: \"" + wrongTier + "\""
)

type TestClusterValues struct {
	ProjectID         string
	Region            string
	ZoneName          string
	SecondaryZoneName string
	ClusterID         string
	ClusterName       string
}

func DefaultTestClusterValues() TestClusterValues {
	return TestClusterValues{
		ProjectID:         "test-project",
		Region:            "us-central1",
		ZoneName:          "us-central1-b",
		SecondaryZoneName: "us-central1-c",
		ClusterID:         "test-cluster-id",
		ClusterName:       "Test Cluster Name",
	}
}

func fakeLoadbalancerService(lbType string) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "",
			Annotations: map[string]string{ServiceAnnotationLoadBalancerType: lbType},
		},
		Spec: v1.ServiceSpec{
			SessionAffinity: v1.ServiceAffinityClientIP,
			Type:            v1.ServiceTypeLoadBalancer,
			Ports:           []v1.ServicePort{{Protocol: v1.ProtocolTCP, Port: int32(123)}},
		},
	}
}

var (
	FilewallChangeMsg = fmt.Sprintf("%s %s %s", v1.EventTypeNormal, eventReasonManualChange, eventMsgFirewallChange)
)

type fakeRoundTripper struct{}

func (*fakeRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, fmt.Errorf("err: test used fake http client")
}

func fakeGCECloud(vals TestClusterValues) (*GCECloud, error) {
	client := &http.Client{Transport: &fakeRoundTripper{}}

	service, err := compute.New(client)
	if err != nil {
		return nil, err
	}

	// Used in disk unit tests
	fakeManager := newFakeManager(vals.ProjectID, vals.Region)
	zonesWithNodes := createNodeZones([]string{vals.ZoneName})

	alphaFeatureGate := NewAlphaFeatureGate([]string{})
	if err != nil {
		return nil, err
	}

	gce := &GCECloud{
		region:             vals.Region,
		service:            service,
		manager:            fakeManager,
		managedZones:       []string{vals.ZoneName},
		projectID:          vals.ProjectID,
		networkProjectID:   vals.ProjectID,
		AlphaFeatureGate:   alphaFeatureGate,
		nodeZones:          zonesWithNodes,
		nodeInformerSynced: func() bool { return true },
		ClusterID:          fakeClusterID(vals.ClusterID),
	}

	c := cloud.NewMockGCE(&gceProjectRouter{gce})
	c.MockTargetPools.AddInstanceHook = mock.AddInstanceHook
	c.MockTargetPools.RemoveInstanceHook = mock.RemoveInstanceHook
	c.MockForwardingRules.InsertHook = mock.InsertFwdRuleHook
	c.MockAddresses.InsertHook = mock.InsertAddressHook
	c.MockAlphaAddresses.InsertHook = mock.InsertAlphaAddressHook
	c.MockAlphaAddresses.X = mock.AddressAttributes{}
	c.MockAddresses.X = mock.AddressAttributes{}

	c.MockInstanceGroups.X = mock.InstanceGroupAttributes{
		InstanceMap: make(map[meta.Key]map[string]*compute.InstanceWithNamedPorts),
		Lock:        &sync.Mutex{},
	}
	c.MockInstanceGroups.AddInstancesHook = mock.AddInstancesHook
	c.MockInstanceGroups.RemoveInstancesHook = mock.RemoveInstancesHook
	c.MockInstanceGroups.ListInstancesHook = mock.ListInstancesHook

	c.MockRegionBackendServices.UpdateHook = mock.UpdateRegionBackendServiceHook
	c.MockHealthChecks.UpdateHook = mock.UpdateHealthCheckHook
	c.MockFirewalls.UpdateHook = mock.UpdateFirewallHook

	keyGA := meta.GlobalKey("key-ga")
	c.MockZones.Objects[*keyGA] = &cloud.MockZonesObj{
		Obj: &compute.Zone{Name: vals.ZoneName, Region: gce.getRegionLink(vals.Region)},
	}

	gce.c = c

	return gce, nil
}

func createAndInsertNodes(gce *GCECloud, nodeNames []string, zoneName string) ([]*v1.Node, error) {
	nodes := []*v1.Node{}

	for _, name := range nodeNames {
		// Inserting the same node name twice causes an error - here we check if
		// the instance exists already before insertion.
		// TestUpdateExternalLoadBalancer inserts a new node, and relies on an older
		// node to already have been inserted.
		instance, _ := gce.getInstanceByName(name)

		if instance == nil {
			err := gce.InsertInstance(
				gce.ProjectID(),
				zoneName,
				&compute.Instance{
					Name: name,
					Tags: &compute.Tags{
						Items: []string{name},
					},
				},
			)
			if err != nil {
				return nodes, err
			}
		}

		nodes = append(
			nodes,
			&v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
					Labels: map[string]string{
						kubeletapis.LabelHostname:          name,
						kubeletapis.LabelZoneFailureDomain: zoneName,
					},
				},
				Status: v1.NodeStatus{
					NodeInfo: v1.NodeSystemInfo{
						KubeProxyVersion: "v1.7.2",
					},
				},
			},
		)

	}

	return nodes, nil
}

// Stubs ClusterID so that ClusterID.getOrInitialize() does not require calling
// gce.Initialize()
func fakeClusterID(clusterID string) ClusterID {
	return ClusterID{
		clusterID: &clusterID,
		store: cache.NewStore(func(obj interface{}) (string, error) {
			return "", nil
		}),
	}
}

func assertExternalLbResources(t *testing.T, gce *GCECloud, apiService *v1.Service, vals TestClusterValues, nodeNames []string) {
	lbName := cloudprovider.GetLoadBalancerName(apiService)
	hcName := MakeNodesHealthCheckName(vals.ClusterID)

	// Check that Firewalls are created for the LoadBalancer and the HealthCheck
	fwNames := []string{
		MakeFirewallName(lbName), // Firewalls for external LBs are prefixed with k8s-fw-
		MakeHealthCheckFirewallName(vals.ClusterID, hcName, true),
	}

	for _, fwName := range fwNames {
		firewall, err := gce.GetFirewall(fwName)
		require.NoError(t, err)
		assert.Equal(t, nodeNames, firewall.TargetTags)
		assert.NotEmpty(t, firewall.SourceRanges)
	}

	// Check that TargetPool is Created
	pool, err := gce.GetTargetPool(lbName, gce.region)
	require.NoError(t, err)
	assert.Equal(t, lbName, pool.Name)
	assert.NotEmpty(t, pool.HealthChecks)
	assert.Equal(t, 1, len(pool.Instances))

	// Check that HealthCheck is created
	healthcheck, err := gce.GetHttpHealthCheck(hcName)
	require.NoError(t, err)
	assert.Equal(t, hcName, healthcheck.Name)

	// Check that ForwardingRule is created
	fwdRule, err := gce.GetRegionForwardingRule(lbName, gce.region)
	require.NoError(t, err)
	assert.Equal(t, lbName, fwdRule.Name)
	assert.Equal(t, "TCP", fwdRule.IPProtocol)
	assert.Equal(t, "123-123", fwdRule.PortRange)
}

func assertExternalLbResourcesDeleted(t *testing.T, gce *GCECloud, apiService *v1.Service, vals TestClusterValues, firewallsDeleted bool) {
	lbName := cloudprovider.GetLoadBalancerName(apiService)
	hcName := MakeNodesHealthCheckName(vals.ClusterID)

	if firewallsDeleted {
		// Check that Firewalls are deleted for the LoadBalancer and the HealthCheck
		fwNames := []string{
			MakeFirewallName(lbName),
			MakeHealthCheckFirewallName(vals.ClusterID, hcName, true),
		}

		for _, fwName := range fwNames {
			firewall, err := gce.GetFirewall(fwName)
			require.Error(t, err)
			assert.Nil(t, firewall)
		}

		// Check forwarding rule is deleted
		fwdRule, err := gce.GetRegionForwardingRule(lbName, gce.region)
		require.Error(t, err)
		assert.Nil(t, fwdRule)
	}

	// Check that TargetPool is deleted
	pool, err := gce.GetTargetPool(lbName, gce.region)
	require.Error(t, err)
	assert.Nil(t, pool)

	// Check that HealthCheck is deleted
	healthcheck, err := gce.GetHttpHealthCheck(hcName)
	require.Error(t, err)
	assert.Nil(t, healthcheck)

}

func assertInternalLbResources(t *testing.T, gce *GCECloud, apiService *v1.Service, vals TestClusterValues, nodeNames []string) {
	lbName := cloudprovider.GetLoadBalancerName(apiService)

	// Check that Instance Group is created
	igName := makeInstanceGroupName(vals.ClusterID)
	ig, err := gce.GetInstanceGroup(igName, vals.ZoneName)
	assert.NoError(t, err)
	assert.Equal(t, igName, ig.Name)

	// Check that Firewalls are created for the LoadBalancer and the HealthCheck
	fwNames := []string{
		lbName, // Firewalls for internal LBs are named the same name as the loadbalancer.
		makeHealthCheckFirewallName(lbName, vals.ClusterID, true),
	}

	for _, fwName := range fwNames {
		firewall, err := gce.GetFirewall(fwName)
		require.NoError(t, err)
		assert.Equal(t, nodeNames, firewall.TargetTags)
		assert.NotEmpty(t, firewall.SourceRanges)
	}

	// Check that HealthCheck is created
	sharedHealthCheck := !v1_service.RequestsOnlyLocalTraffic(apiService)
	hcName := makeHealthCheckName(lbName, vals.ClusterID, sharedHealthCheck)
	healthcheck, err := gce.GetHealthCheck(hcName)
	require.NoError(t, err)
	assert.Equal(t, hcName, healthcheck.Name)

	// Check that BackendService exists
	sharedBackend := shareBackendService(apiService)
	backendServiceName := makeBackendServiceName(lbName, vals.ClusterID, sharedBackend, cloud.SchemeInternal, "TCP", apiService.Spec.SessionAffinity)
	backendServiceLink := gce.getBackendServiceLink(backendServiceName)

	bs, err := gce.GetRegionBackendService(backendServiceName, gce.region)
	require.NoError(t, err)
	assert.Equal(t, "TCP", bs.Protocol)
	assert.Equal(
		t,
		[]string{healthcheck.SelfLink},
		bs.HealthChecks,
	)

	// Check that ForwardingRule is created
	fwdRule, err := gce.GetRegionForwardingRule(lbName, gce.region)
	require.NoError(t, err)
	assert.Equal(t, lbName, fwdRule.Name)
	assert.Equal(t, "TCP", fwdRule.IPProtocol)
	assert.Equal(t, backendServiceLink, fwdRule.BackendService)
	// if no Subnetwork specified, defaults to the GCE NetworkURL
	assert.Equal(t, gce.NetworkURL(), fwdRule.Subnetwork)
}

func assertInternalLbResourcesDeleted(t *testing.T, gce *GCECloud, apiService *v1.Service, vals TestClusterValues, firewallsDeleted bool) {
	lbName := cloudprovider.GetLoadBalancerName(apiService)
	sharedHealthCheck := !v1_service.RequestsOnlyLocalTraffic(apiService)
	hcName := makeHealthCheckName(lbName, vals.ClusterID, sharedHealthCheck)

	// ensureExternalLoadBalancer and ensureInternalLoadBalancer both create
	// Firewalls with the same name.
	if firewallsDeleted {
		// Check that Firewalls are deleted for the LoadBalancer and the HealthCheck
		fwNames := []string{
			MakeFirewallName(lbName),
			MakeHealthCheckFirewallName(vals.ClusterID, hcName, true),
		}

		for _, fwName := range fwNames {
			firewall, err := gce.GetFirewall(fwName)
			require.Error(t, err)
			assert.Nil(t, firewall)
		}

		// Check forwarding rule is deleted
		fwdRule, err := gce.GetRegionForwardingRule(lbName, gce.region)
		require.Error(t, err)
		assert.Nil(t, fwdRule)
	}

	// Check that Instance Group is deleted
	igName := makeInstanceGroupName(vals.ClusterID)
	ig, err := gce.GetInstanceGroup(igName, vals.ZoneName)
	assert.Error(t, err)
	assert.Nil(t, ig)

	// Check that HealthCheck is deleted
	healthcheck, err := gce.GetHealthCheck(hcName)
	require.Error(t, err)
	assert.Nil(t, healthcheck)
}

func checkEvent(t *testing.T, recorder *record.FakeRecorder, expected string, shouldMatch bool) bool {
	select {
	case received := <-recorder.Events:
		if strings.HasPrefix(received, expected) != shouldMatch {
			t.Errorf(received)
			if shouldMatch {
				t.Errorf("Should receive message \"%v\" but got \"%v\".", expected, received)
			} else {
				t.Errorf("Unexpected event \"%v\".", received)
			}
		}
		return false
	case <-time.After(2 * time.Second):
		if shouldMatch {
			t.Errorf("Should receive message \"%v\" but got timed out.", expected)
		}
		return true
	}
}
