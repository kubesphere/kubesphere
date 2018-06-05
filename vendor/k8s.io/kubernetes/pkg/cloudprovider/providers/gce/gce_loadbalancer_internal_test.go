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

package gce

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	v1_service "k8s.io/kubernetes/pkg/api/v1/service"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/cloudprovider/providers/gce/cloud"
	"k8s.io/kubernetes/pkg/cloudprovider/providers/gce/cloud/mock"
)

func createInternalLoadBalancer(gce *GCECloud, svc *v1.Service, existingFwdRule *compute.ForwardingRule, nodeNames []string, clusterName, clusterID, zoneName string) (*v1.LoadBalancerStatus, error) {
	nodes, err := createAndInsertNodes(gce, nodeNames, zoneName)
	if err != nil {
		return nil, err
	}

	return gce.ensureInternalLoadBalancer(
		clusterName,
		clusterID,
		svc,
		existingFwdRule,
		nodes,
	)
}

func TestEnsureInternalBackendServiceUpdates(t *testing.T) {
	t.Parallel()

	vals := DefaultTestClusterValues()
	nodeNames := []string{"test-node-1"}

	gce, err := fakeGCECloud(vals)
	require.NoError(t, err)

	svc := fakeLoadbalancerService(string(LBTypeInternal))
	lbName := cloudprovider.GetLoadBalancerName(svc)
	nodes, err := createAndInsertNodes(gce, nodeNames, vals.ZoneName)
	igName := makeInstanceGroupName(vals.ClusterID)
	igLinks, err := gce.ensureInternalInstanceGroups(igName, nodes)
	require.NoError(t, err)

	sharedBackend := shareBackendService(svc)
	bsName := makeBackendServiceName(lbName, vals.ClusterID, sharedBackend, cloud.SchemeInternal, "TCP", svc.Spec.SessionAffinity)
	err = gce.ensureInternalBackendService(bsName, "description", svc.Spec.SessionAffinity, cloud.SchemeInternal, "TCP", igLinks, "")
	require.NoError(t, err)

	// Update the Internal Backend Service with a new ServiceAffinity
	err = gce.ensureInternalBackendService(bsName, "description", v1.ServiceAffinityNone, cloud.SchemeInternal, "TCP", igLinks, "")
	require.NoError(t, err)

	bs, err := gce.GetRegionBackendService(bsName, gce.region)
	assert.NoError(t, err)
	assert.Equal(t, bs.SessionAffinity, strings.ToUpper(string(v1.ServiceAffinityNone)))
}

func TestEnsureInternalBackendServiceGroups(t *testing.T) {
	for desc, tc := range map[string]struct {
		mockModifier func(*cloud.MockGCE)
	}{
		"Basic workflow": {},
		"GetRegionBackendService failed": {
			mockModifier: func(c *cloud.MockGCE) {
				c.MockRegionBackendServices.GetHook = mock.GetRegionBackendServicesErrHook
			},
		},
		"UpdateRegionBackendServices failed": {
			mockModifier: func(c *cloud.MockGCE) {
				c.MockRegionBackendServices.UpdateHook = mock.UpdateRegionBackendServicesErrHook
			},
		},
	} {
		t.Run(desc, func(t *testing.T) {
			t.Parallel()

			vals := DefaultTestClusterValues()
			nodeNames := []string{"test-node-1"}

			gce, err := fakeGCECloud(vals)
			require.NoError(t, err)

			svc := fakeLoadbalancerService(string(LBTypeInternal))
			lbName := cloudprovider.GetLoadBalancerName(svc)
			nodes, err := createAndInsertNodes(gce, nodeNames, vals.ZoneName)
			igName := makeInstanceGroupName(vals.ClusterID)
			igLinks, err := gce.ensureInternalInstanceGroups(igName, nodes)
			require.NoError(t, err)

			sharedBackend := shareBackendService(svc)
			bsName := makeBackendServiceName(lbName, vals.ClusterID, sharedBackend, cloud.SchemeInternal, "TCP", svc.Spec.SessionAffinity)

			err = gce.ensureInternalBackendService(bsName, "description", svc.Spec.SessionAffinity, cloud.SchemeInternal, "TCP", igLinks, "")
			require.NoError(t, err)

			// Update the BackendService with new Instances
			if tc.mockModifier != nil {
				tc.mockModifier(gce.c.(*cloud.MockGCE))
			}
			newNodeNames := []string{"new-test-node-1", "new-test-node-2"}
			err = gce.ensureInternalBackendServiceGroups(bsName, newNodeNames)
			if tc.mockModifier != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			bs, err := gce.GetRegionBackendService(bsName, gce.region)
			assert.NoError(t, err)

			// Check that the instances are updated
			newNodes, err := createAndInsertNodes(gce, newNodeNames, vals.ZoneName)
			newIgLinks, err := gce.ensureInternalInstanceGroups(igName, newNodes)
			backends := backendsFromGroupLinks(newIgLinks)
			assert.Equal(t, bs.Backends, backends)
		})
	}
}

func TestEnsureInternalLoadBalancer(t *testing.T) {
	t.Parallel()

	vals := DefaultTestClusterValues()
	nodeNames := []string{"test-node-1"}

	gce, err := fakeGCECloud(vals)
	require.NoError(t, err)

	svc := fakeLoadbalancerService(string(LBTypeInternal))
	status, err := createInternalLoadBalancer(gce, svc, nil, nodeNames, vals.ClusterName, vals.ClusterID, vals.ZoneName)
	assert.NoError(t, err)
	assert.NotEmpty(t, status.Ingress)
	assertInternalLbResources(t, gce, svc, vals, nodeNames)
}

func TestEnsureInternalLoadBalancerWithExistingResources(t *testing.T) {
	t.Parallel()

	vals := DefaultTestClusterValues()
	nodeNames := []string{"test-node-1"}

	gce, err := fakeGCECloud(vals)
	require.NoError(t, err)
	svc := fakeLoadbalancerService(string(LBTypeInternal))

	// Create the expected resources necessary for an Internal Load Balancer
	nm := types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}
	lbName := cloudprovider.GetLoadBalancerName(svc)

	sharedHealthCheck := !v1_service.RequestsOnlyLocalTraffic(svc)
	hcName := makeHealthCheckName(lbName, vals.ClusterID, sharedHealthCheck)
	hcPath, hcPort := GetNodesHealthCheckPath(), GetNodesHealthCheckPort()
	existingHC := newInternalLBHealthCheck(hcName, nm, sharedHealthCheck, hcPath, hcPort)
	err = gce.CreateHealthCheck(existingHC)
	require.NoError(t, err)

	nodes, err := createAndInsertNodes(gce, nodeNames, vals.ZoneName)
	igName := makeInstanceGroupName(vals.ClusterID)
	igLinks, err := gce.ensureInternalInstanceGroups(igName, nodes)
	require.NoError(t, err)

	sharedBackend := shareBackendService(svc)
	bsDescription := makeBackendServiceDescription(nm, sharedBackend)
	bsName := makeBackendServiceName(lbName, vals.ClusterID, sharedBackend, cloud.SchemeInternal, "TCP", svc.Spec.SessionAffinity)
	err = gce.ensureInternalBackendService(bsName, bsDescription, svc.Spec.SessionAffinity, cloud.SchemeInternal, "TCP", igLinks, existingHC.SelfLink)
	require.NoError(t, err)

	_, err = createInternalLoadBalancer(gce, svc, nil, nodeNames, vals.ClusterName, vals.ClusterID, vals.ZoneName)
	assert.NoError(t, err)
}

func TestEnsureInternalLoadBalancerClearPreviousResources(t *testing.T) {
	t.Parallel()

	vals := DefaultTestClusterValues()
	gce, err := fakeGCECloud(vals)
	require.NoError(t, err)

	svc := fakeLoadbalancerService(string(LBTypeInternal))
	lbName := cloudprovider.GetLoadBalancerName(svc)

	// Create a ForwardingRule that's missing an IP address
	existingFwdRule := &compute.ForwardingRule{
		Name:                lbName,
		IPAddress:           "",
		Ports:               []string{"123"},
		IPProtocol:          "TCP",
		LoadBalancingScheme: string(cloud.SchemeInternal),
	}
	gce.CreateRegionForwardingRule(existingFwdRule, gce.region)

	// Create a Firewall that's missing a Description
	existingFirewall := &compute.Firewall{
		Name:    lbName,
		Network: gce.networkURL,
		Allowed: []*compute.FirewallAllowed{
			{
				IPProtocol: "tcp",
				Ports:      []string{"123"},
			},
		},
	}
	gce.CreateFirewall(existingFirewall)

	sharedHealthCheck := !v1_service.RequestsOnlyLocalTraffic(svc)
	hcName := makeHealthCheckName(lbName, vals.ClusterID, sharedHealthCheck)
	hcPath, hcPort := GetNodesHealthCheckPath(), GetNodesHealthCheckPort()
	nm := types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}

	// Create a healthcheck with an incorrect threshold
	existingHC := newInternalLBHealthCheck(hcName, nm, sharedHealthCheck, hcPath, hcPort)
	existingHC.HealthyThreshold = gceHcHealthyThreshold * 10
	gce.CreateHealthCheck(existingHC)

	// Create a backend Service that's missing Description and Backends
	sharedBackend := shareBackendService(svc)
	backendServiceName := makeBackendServiceName(lbName, vals.ClusterID, sharedBackend, cloud.SchemeInternal, "TCP", svc.Spec.SessionAffinity)
	existingBS := &compute.BackendService{
		Name:                lbName,
		Protocol:            "TCP",
		HealthChecks:        []string{existingHC.SelfLink},
		SessionAffinity:     translateAffinityType(svc.Spec.SessionAffinity),
		LoadBalancingScheme: string(cloud.SchemeInternal),
	}

	gce.CreateRegionBackendService(existingBS, gce.region)
	existingFwdRule.BackendService = existingBS.Name

	_, err = createInternalLoadBalancer(gce, svc, existingFwdRule, []string{"test-node-1"}, vals.ClusterName, vals.ClusterID, vals.ZoneName)
	assert.NoError(t, err)

	// Expect new resources with the correct attributes to be created
	rule, _ := gce.GetRegionForwardingRule(lbName, gce.region)
	assert.NotEqual(t, existingFwdRule, rule)

	firewall, err := gce.GetFirewall(lbName)
	require.NoError(t, err)
	assert.NotEqual(t, firewall, existingFirewall)

	healthcheck, err := gce.GetHealthCheck(hcName)
	require.NoError(t, err)
	assert.NotEqual(t, healthcheck, existingHC)

	bs, err := gce.GetRegionBackendService(backendServiceName, gce.region)
	require.NoError(t, err)
	assert.NotEqual(t, bs, existingBS)
}

func TestUpdateInternalLoadBalancerBackendServices(t *testing.T) {
	t.Parallel()

	vals := DefaultTestClusterValues()
	nodeName := "test-node-1"

	gce, err := fakeGCECloud(vals)
	require.NoError(t, err)

	svc := fakeLoadbalancerService(string(LBTypeInternal))
	_, err = createInternalLoadBalancer(gce, svc, nil, []string{"test-node-1"}, vals.ClusterName, vals.ClusterID, vals.ZoneName)
	assert.NoError(t, err)

	// BackendService exists prior to updateInternalLoadBalancer call, but has
	// incorrect (missing) attributes.
	// ensureInternalBackendServiceGroups is called and creates the correct
	// BackendService
	lbName := cloudprovider.GetLoadBalancerName(svc)
	sharedBackend := shareBackendService(svc)
	backendServiceName := makeBackendServiceName(lbName, vals.ClusterID, sharedBackend, cloud.SchemeInternal, "TCP", svc.Spec.SessionAffinity)
	existingBS := &compute.BackendService{
		Name:                backendServiceName,
		Protocol:            "TCP",
		SessionAffinity:     translateAffinityType(svc.Spec.SessionAffinity),
		LoadBalancingScheme: string(cloud.SchemeInternal),
	}

	gce.CreateRegionBackendService(existingBS, gce.region)

	nodes, err := createAndInsertNodes(gce, []string{nodeName}, vals.ZoneName)
	require.NoError(t, err)

	err = gce.updateInternalLoadBalancer(vals.ClusterName, vals.ClusterID, svc, nodes)
	assert.NoError(t, err)

	bs, err := gce.GetRegionBackendService(backendServiceName, gce.region)
	require.NoError(t, err)

	// Check that the new BackendService has the correct attributes
	url_base := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s", vals.ProjectID)

	assert.NotEqual(t, existingBS, bs)
	assert.Equal(
		t,
		bs.SelfLink,
		fmt.Sprintf("%s/regions/%s/backendServices/%s", url_base, vals.Region, bs.Name),
	)
	assert.Equal(t, bs.Description, `{"kubernetes.io/service-name":"/"}`)
	assert.Equal(
		t,
		bs.HealthChecks,
		[]string{fmt.Sprintf("%s/global/healthChecks/k8s-%s-node", url_base, vals.ClusterID)},
	)
}

func TestUpdateInternalLoadBalancerNodes(t *testing.T) {
	t.Parallel()

	vals := DefaultTestClusterValues()
	gce, err := fakeGCECloud(vals)
	require.NoError(t, err)
	node1Name := []string{"test-node-1"}

	svc := fakeLoadbalancerService(string(LBTypeInternal))
	nodes, err := createAndInsertNodes(gce, node1Name, vals.ZoneName)
	require.NoError(t, err)

	_, err = gce.ensureInternalLoadBalancer(vals.ClusterName, vals.ClusterID, svc, nil, nodes)
	assert.NoError(t, err)

	// Replace the node in initial zone; add new node in a new zone.
	node2Name, node3Name := "test-node-2", "test-node-3"
	newNodesZoneA, err := createAndInsertNodes(gce, []string{node2Name}, vals.ZoneName)
	require.NoError(t, err)
	newNodesZoneB, err := createAndInsertNodes(gce, []string{node3Name}, vals.SecondaryZoneName)
	require.NoError(t, err)

	nodes = append(newNodesZoneA, newNodesZoneB...)
	err = gce.updateInternalLoadBalancer(vals.ClusterName, vals.ClusterID, svc, nodes)
	assert.NoError(t, err)

	lbName := cloudprovider.GetLoadBalancerName(svc)
	sharedBackend := shareBackendService(svc)
	backendServiceName := makeBackendServiceName(lbName, vals.ClusterID, sharedBackend, cloud.SchemeInternal, "TCP", svc.Spec.SessionAffinity)
	bs, err := gce.GetRegionBackendService(backendServiceName, gce.region)
	require.NoError(t, err)
	assert.Equal(t, 2, len(bs.Backends), "Want two backends referencing two instances groups")

	for _, zone := range []string{vals.ZoneName, vals.SecondaryZoneName} {
		var found bool
		for _, be := range bs.Backends {
			if strings.Contains(be.Group, zone) {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected list of backends to have zone %q", zone)
	}

	// Expect initial zone to have test-node-2
	igName := makeInstanceGroupName(vals.ClusterID)
	instances, err := gce.ListInstancesInInstanceGroup(igName, vals.ZoneName, "ALL")
	require.NoError(t, err)
	assert.Equal(t, 1, len(instances))
	assert.Contains(
		t,
		instances[0].Instance,
		fmt.Sprintf("projects/%s/zones/%s/instances/%s", vals.ProjectID, vals.ZoneName, node2Name),
	)

	// Expect initial zone to have test-node-3
	instances, err = gce.ListInstancesInInstanceGroup(igName, vals.SecondaryZoneName, "ALL")
	require.NoError(t, err)
	assert.Equal(t, 1, len(instances))
	assert.Contains(
		t,
		instances[0].Instance,
		fmt.Sprintf("projects/%s/zones/%s/instances/%s", vals.ProjectID, vals.SecondaryZoneName, node3Name),
	)
}

func TestEnsureInternalLoadBalancerDeleted(t *testing.T) {
	t.Parallel()

	vals := DefaultTestClusterValues()
	gce, err := fakeGCECloud(vals)
	require.NoError(t, err)

	svc := fakeLoadbalancerService(string(LBTypeInternal))
	_, err = createInternalLoadBalancer(gce, svc, nil, []string{"test-node-1"}, vals.ClusterName, vals.ClusterID, vals.ZoneName)
	assert.NoError(t, err)

	err = gce.ensureInternalLoadBalancerDeleted(vals.ClusterName, vals.ClusterID, svc)
	assert.NoError(t, err)

	assertInternalLbResourcesDeleted(t, gce, svc, vals, true)
}

func TestEnsureInternalLoadBalancerDeletedTwiceDoesNotError(t *testing.T) {
	t.Parallel()

	vals := DefaultTestClusterValues()
	gce, err := fakeGCECloud(vals)
	require.NoError(t, err)
	svc := fakeLoadbalancerService(string(LBTypeInternal))

	_, err = createInternalLoadBalancer(gce, svc, nil, []string{"test-node-1"}, vals.ClusterName, vals.ClusterID, vals.ZoneName)
	assert.NoError(t, err)

	err = gce.ensureInternalLoadBalancerDeleted(vals.ClusterName, vals.ClusterID, svc)
	assert.NoError(t, err)

	// Deleting the loadbalancer and resources again should not cause an error.
	err = gce.ensureInternalLoadBalancerDeleted(vals.ClusterName, vals.ClusterID, svc)
	assert.NoError(t, err)
	assertInternalLbResourcesDeleted(t, gce, svc, vals, true)
}

func TestEnsureInternalLoadBalancerWithSpecialHealthCheck(t *testing.T) {
	vals := DefaultTestClusterValues()
	nodeName := "test-node-1"
	gce, err := fakeGCECloud(vals)
	require.NoError(t, err)

	healthCheckNodePort := int32(10101)
	svc := fakeLoadbalancerService(string(LBTypeInternal))
	svc.Spec.HealthCheckNodePort = healthCheckNodePort
	svc.Spec.Type = v1.ServiceTypeLoadBalancer
	svc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal

	status, err := createInternalLoadBalancer(gce, svc, nil, []string{nodeName}, vals.ClusterName, vals.ClusterID, vals.ZoneName)
	assert.NoError(t, err)
	assert.NotEmpty(t, status.Ingress)

	loadBalancerName := cloudprovider.GetLoadBalancerName(svc)
	hc, err := gce.GetHealthCheck(loadBalancerName)
	assert.NoError(t, err)
	assert.NotNil(t, hc)
	assert.Equal(t, int64(healthCheckNodePort), hc.HttpHealthCheck.Port)
}

func TestClearPreviousInternalResources(t *testing.T) {
	// Configure testing environment.
	vals := DefaultTestClusterValues()
	svc := fakeLoadbalancerService(string(LBTypeInternal))
	loadBalancerName := cloudprovider.GetLoadBalancerName(svc)
	nm := types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}
	gce, err := fakeGCECloud(vals)
	c := gce.c.(*cloud.MockGCE)
	require.NoError(t, err)

	hc_1, err := gce.ensureInternalHealthCheck("hc_1", nm, false, "healthz", 12345)
	require.NoError(t, err)

	hc_2, err := gce.ensureInternalHealthCheck("hc_2", nm, false, "healthz", 12346)
	require.NoError(t, err)

	err = gce.ensureInternalBackendService(svc.ObjectMeta.Name, "", svc.Spec.SessionAffinity, cloud.SchemeInternal, v1.ProtocolTCP, []string{}, "")
	require.NoError(t, err)
	backendSvc, err := gce.GetRegionBackendService(svc.ObjectMeta.Name, gce.region)
	backendSvc.HealthChecks = []string{hc_1.SelfLink, hc_2.SelfLink}

	c.MockRegionBackendServices.DeleteHook = mock.DeleteRegionBackendServicesErrHook
	c.MockHealthChecks.DeleteHook = mock.DeleteHealthChecksInternalErrHook
	gce.clearPreviousInternalResources(svc, loadBalancerName, backendSvc, "expectedBSName", "expectedHCName")

	backendSvc, err = gce.GetRegionBackendService(svc.ObjectMeta.Name, gce.region)
	assert.NoError(t, err)
	assert.NotNil(t, backendSvc, "BackendService should not be deleted when api is mocked out.")
	hc_1, err = gce.GetHealthCheck("hc_1")
	assert.NoError(t, err)
	assert.NotNil(t, hc_1, "HealthCheck should not be deleted when there are more than one healthcheck attached.")
	hc_2, err = gce.GetHealthCheck("hc_2")
	assert.NoError(t, err)
	assert.NotNil(t, hc_2, "HealthCheck should not be deleted when there are more than one healthcheck attached.")

	c.MockRegionBackendServices.DeleteHook = mock.DeleteRegionBackendServicesInUseErrHook
	backendSvc.HealthChecks = []string{hc_1.SelfLink}
	gce.clearPreviousInternalResources(svc, loadBalancerName, backendSvc, "expectedBSName", "expectedHCName")

	hc_1, err = gce.GetHealthCheck("hc_1")
	assert.NoError(t, err)
	assert.NotNil(t, hc_1, "HealthCheck should not be deleted when api is mocked out.")

	c.MockHealthChecks.DeleteHook = mock.DeleteHealthChecksInuseErrHook
	gce.clearPreviousInternalResources(svc, loadBalancerName, backendSvc, "expectedBSName", "expectedHCName")

	hc_1, err = gce.GetHealthCheck("hc_1")
	assert.NoError(t, err)
	assert.NotNil(t, hc_1, "HealthCheck should not be deleted when api is mocked out.")

	c.MockRegionBackendServices.DeleteHook = nil
	c.MockHealthChecks.DeleteHook = nil
	gce.clearPreviousInternalResources(svc, loadBalancerName, backendSvc, "expectedBSName", "expectedHCName")

	backendSvc, err = gce.GetRegionBackendService(svc.ObjectMeta.Name, gce.region)
	assert.Error(t, err)
	assert.Nil(t, backendSvc, "BackendService should be deleted.")
	hc_1, err = gce.GetHealthCheck("hc_1")
	assert.Error(t, err)
	assert.Nil(t, hc_1, "HealthCheck should be deleted.")
}

func TestEnsureInternalFirewallSucceedsOnXPN(t *testing.T) {
	gce, err := fakeGCECloud(DefaultTestClusterValues())
	require.NoError(t, err)
	vals := DefaultTestClusterValues()
	svc := fakeLoadbalancerService(string(LBTypeInternal))
	fwName := cloudprovider.GetLoadBalancerName(svc)

	c := gce.c.(*cloud.MockGCE)
	c.MockFirewalls.InsertHook = mock.InsertFirewallsUnauthorizedErrHook
	c.MockFirewalls.UpdateHook = mock.UpdateFirewallsUnauthorizedErrHook
	gce.onXPN = true
	require.True(t, gce.OnXPN())

	recorder := record.NewFakeRecorder(1024)
	gce.eventRecorder = recorder

	nodes, err := createAndInsertNodes(gce, []string{"test-node-1"}, vals.ZoneName)
	require.NoError(t, err)
	sourceRange := []string{"10.0.0.0/20"}
	gce.ensureInternalFirewall(
		svc,
		fwName,
		"A sad little firewall",
		sourceRange,
		[]string{"123"},
		v1.ProtocolTCP,
		nodes)
	require.Nil(t, err, "Should success when XPN is on.")

	checkEvent(t, recorder, FilewallChangeMsg, true)

	// Create a firewall.
	c.MockFirewalls.InsertHook = nil
	c.MockFirewalls.UpdateHook = nil
	gce.onXPN = false

	gce.ensureInternalFirewall(
		svc,
		fwName,
		"A sad little firewall",
		sourceRange,
		[]string{"123"},
		v1.ProtocolTCP,
		nodes)
	require.Nil(t, err)
	existingFirewall, err := gce.GetFirewall(fwName)
	require.Nil(t, err)
	require.NotNil(t, existingFirewall)

	gce.onXPN = true
	c.MockFirewalls.InsertHook = mock.InsertFirewallsUnauthorizedErrHook
	c.MockFirewalls.UpdateHook = mock.UpdateFirewallsUnauthorizedErrHook

	// Try to update the firewall just created.
	gce.ensureInternalFirewall(
		svc,
		fwName,
		"A happy little firewall",
		sourceRange,
		[]string{"123"},
		v1.ProtocolTCP,
		nodes)
	require.Nil(t, err, "Should success when XPN is on.")

	checkEvent(t, recorder, FilewallChangeMsg, true)
}

func TestEnsureLoadBalancerDeletedSucceedsOnXPN(t *testing.T) {
	vals := DefaultTestClusterValues()
	gce, err := fakeGCECloud(vals)
	c := gce.c.(*cloud.MockGCE)
	recorder := record.NewFakeRecorder(1024)
	gce.eventRecorder = recorder
	require.NoError(t, err)

	_, err = createInternalLoadBalancer(gce, fakeLoadbalancerService(string(LBTypeInternal)), nil, []string{"test-node-1"}, vals.ClusterName, vals.ClusterID, vals.ZoneName)
	assert.NoError(t, err)

	c.MockFirewalls.DeleteHook = mock.DeleteFirewallsUnauthorizedErrHook
	gce.onXPN = true

	err = gce.ensureInternalLoadBalancerDeleted(vals.ClusterName, vals.ClusterID, fakeLoadbalancerService(string(LBTypeInternal)))
	assert.NoError(t, err)
	checkEvent(t, recorder, FilewallChangeMsg, true)
}

func TestEnsureInternalInstanceGroupsDeleted(t *testing.T) {
	vals := DefaultTestClusterValues()
	gce, err := fakeGCECloud(vals)
	c := gce.c.(*cloud.MockGCE)
	recorder := record.NewFakeRecorder(1024)
	gce.eventRecorder = recorder
	require.NoError(t, err)

	igName := makeInstanceGroupName(vals.ClusterID)

	svc := fakeLoadbalancerService(string(LBTypeInternal))
	_, err = createInternalLoadBalancer(gce, svc, nil, []string{"test-node-1"}, vals.ClusterName, vals.ClusterID, vals.ZoneName)
	assert.NoError(t, err)

	c.MockZones.ListHook = mock.ListZonesInternalErrHook

	err = gce.ensureInternalLoadBalancerDeleted(igName, vals.ClusterID, svc)
	assert.Error(t, err, mock.InternalServerError)
	ig, err := gce.GetInstanceGroup(igName, vals.ZoneName)
	assert.NoError(t, err)
	assert.NotNil(t, ig)

	c.MockZones.ListHook = nil
	c.MockInstanceGroups.DeleteHook = mock.DeleteInstanceGroupInternalErrHook

	err = gce.ensureInternalInstanceGroupsDeleted(igName)
	assert.Error(t, err, mock.InternalServerError)
	ig, err = gce.GetInstanceGroup(igName, vals.ZoneName)
	assert.NoError(t, err)
	assert.NotNil(t, ig)

	c.MockInstanceGroups.DeleteHook = nil
	err = gce.ensureInternalInstanceGroupsDeleted(igName)
	assert.NoError(t, err)
	ig, err = gce.GetInstanceGroup(igName, vals.ZoneName)
	assert.Error(t, err)
	assert.Nil(t, ig)
}

type EnsureILBParams struct {
	clusterName     string
	clusterID       string
	service         *v1.Service
	existingFwdRule *compute.ForwardingRule
	nodes           []*v1.Node
}

// newEnsureILBParams is the constructor of EnsureILBParams.
func newEnsureILBParams(nodes []*v1.Node) *EnsureILBParams {
	vals := DefaultTestClusterValues()
	return &EnsureILBParams{
		vals.ClusterName,
		vals.ClusterID,
		fakeLoadbalancerService(string(LBTypeInternal)),
		nil,
		nodes,
	}
}

// TestEnsureInternalLoadBalancerErrors tests the function
// ensureInternalLoadBalancer, making sure the system won't panic when
// exceptions raised by gce.
func TestEnsureInternalLoadBalancerErrors(t *testing.T) {
	vals := DefaultTestClusterValues()
	var params *EnsureILBParams

	for desc, tc := range map[string]struct {
		adjustParams func(*EnsureILBParams)
		injectMock   func(*cloud.MockGCE)
	}{
		"Create internal instance groups failed": {
			injectMock: func(c *cloud.MockGCE) {
				c.MockInstanceGroups.GetHook = mock.GetInstanceGroupInternalErrHook
			},
		},
		"Invalid existing forwarding rules given": {
			adjustParams: func(params *EnsureILBParams) {
				params.existingFwdRule = &compute.ForwardingRule{BackendService: "badBackendService"}
			},
			injectMock: func(c *cloud.MockGCE) {
				c.MockRegionBackendServices.GetHook = mock.GetRegionBackendServicesErrHook
			},
		},
		"EnsureInternalBackendService failed": {
			injectMock: func(c *cloud.MockGCE) {
				c.MockRegionBackendServices.GetHook = mock.GetRegionBackendServicesErrHook
			},
		},
		"Create internal health check failed": {
			injectMock: func(c *cloud.MockGCE) {
				c.MockHealthChecks.GetHook = mock.GetHealthChecksInternalErrHook
			},
		},
		"Create firewall failed": {
			injectMock: func(c *cloud.MockGCE) {
				c.MockFirewalls.InsertHook = mock.InsertFirewallsUnauthorizedErrHook
			},
		},
		"Create region forwarding rule failed": {
			injectMock: func(c *cloud.MockGCE) {
				c.MockForwardingRules.InsertHook = mock.InsertForwardingRulesInternalErrHook
			},
		},
		"Get region forwarding rule failed": {
			injectMock: func(c *cloud.MockGCE) {
				c.MockForwardingRules.GetHook = mock.GetForwardingRulesInternalErrHook
			},
		},
		"Delete region forwarding rule failed": {
			adjustParams: func(params *EnsureILBParams) {
				params.existingFwdRule = &compute.ForwardingRule{BackendService: "badBackendService"}
			},
			injectMock: func(c *cloud.MockGCE) {
				c.MockForwardingRules.DeleteHook = mock.DeleteForwardingRuleErrHook
			},
		},
	} {
		t.Run(desc, func(t *testing.T) {
			gce, err := fakeGCECloud(DefaultTestClusterValues())
			nodes, err := createAndInsertNodes(gce, []string{"test-node-1"}, vals.ZoneName)
			require.NoError(t, err)
			params = newEnsureILBParams(nodes)
			if tc.adjustParams != nil {
				tc.adjustParams(params)
			}
			if tc.injectMock != nil {
				tc.injectMock(gce.c.(*cloud.MockGCE))
			}
			status, err := gce.ensureInternalLoadBalancer(
				params.clusterName,
				params.clusterID,
				params.service,
				params.existingFwdRule,
				params.nodes,
			)
			assert.Error(t, err, "Should return an error when "+desc)
			assert.Nil(t, status, "Should not return a status when "+desc)
		})
	}
}
