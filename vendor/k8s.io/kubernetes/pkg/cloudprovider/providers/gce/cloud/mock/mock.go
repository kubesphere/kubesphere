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

// Package mock encapsulates mocks for testing GCE provider functionality.
// These methods are used to override the mock objects' methods in order to
// intercept the standard processing and to add custom logic for test purposes.
//
//  // Example usage:
// cloud := cloud.NewMockGCE()
// cloud.MockTargetPools.AddInstanceHook = mock.AddInstanceHook
package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	alpha "google.golang.org/api/compute/v0.alpha"
	beta "google.golang.org/api/compute/v0.beta"
	ga "google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	cloud "k8s.io/kubernetes/pkg/cloudprovider/providers/gce/cloud"
	"k8s.io/kubernetes/pkg/cloudprovider/providers/gce/cloud/filter"
	"k8s.io/kubernetes/pkg/cloudprovider/providers/gce/cloud/meta"
)

var (
	// InUseError is a shared variable with error code StatusBadRequest for error verification.
	InUseError = &googleapi.Error{Code: http.StatusBadRequest, Message: "It's being used by god."}
	// InternalServerError is shared variable with error code StatusInternalServerError for error verification.
	InternalServerError = &googleapi.Error{Code: http.StatusInternalServerError}
)

// gceObject is an abstraction of all GCE API object in go client
type gceObject interface {
	MarshalJSON() ([]byte, error)
}

// AddInstanceHook mocks adding a Instance to MockTargetPools
func AddInstanceHook(ctx context.Context, key *meta.Key, req *ga.TargetPoolsAddInstanceRequest, m *cloud.MockTargetPools) error {
	pool, err := m.Get(ctx, key)
	if err != nil {
		return &googleapi.Error{
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("Key: %s was not found in TargetPools", key.String()),
		}
	}

	for _, instance := range req.Instances {
		pool.Instances = append(pool.Instances, instance.Instance)
	}

	return nil
}

// RemoveInstanceHook mocks removing a Instance from MockTargetPools
func RemoveInstanceHook(ctx context.Context, key *meta.Key, req *ga.TargetPoolsRemoveInstanceRequest, m *cloud.MockTargetPools) error {
	pool, err := m.Get(ctx, key)
	if err != nil {
		return &googleapi.Error{
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("Key: %s was not found in TargetPools", key.String()),
		}
	}

	for _, instanceToRemove := range req.Instances {
		for i, instance := range pool.Instances {
			if instanceToRemove.Instance == instance {
				// Delete instance from pool.Instances without preserving order
				pool.Instances[i] = pool.Instances[len(pool.Instances)-1]
				pool.Instances = pool.Instances[:len(pool.Instances)-1]
				break
			}
		}
	}

	return nil
}

func convertAndInsertAlphaForwardingRule(key *meta.Key, obj gceObject, mRules map[meta.Key]*cloud.MockForwardingRulesObj, version meta.Version, projectID string) (bool, error) {
	if !key.Valid() {
		return false, fmt.Errorf("invalid GCE key (%+v)", key)
	}

	if _, ok := mRules[*key]; ok {
		err := &googleapi.Error{
			Code:    http.StatusConflict,
			Message: fmt.Sprintf("MockForwardingRule %v exists", key),
		}
		return false, err
	}

	enc, err := obj.MarshalJSON()
	if err != nil {
		return false, err
	}
	var fwdRule alpha.ForwardingRule
	if err := json.Unmarshal(enc, &fwdRule); err != nil {
		return false, err
	}
	// Set the default values for the Alpha fields.
	if fwdRule.NetworkTier == "" {
		fwdRule.NetworkTier = cloud.NetworkTierDefault.ToGCEValue()
	}

	fwdRule.Name = key.Name
	if fwdRule.SelfLink == "" {
		fwdRule.SelfLink = cloud.SelfLink(version, projectID, "forwardingRules", key)
	}

	mRules[*key] = &cloud.MockForwardingRulesObj{Obj: fwdRule}
	return true, nil
}

// InsertFwdRuleHook mocks inserting a ForwardingRule. ForwardingRules are
// expected to default to Premium tier if no NetworkTier is specified.
func InsertFwdRuleHook(ctx context.Context, key *meta.Key, obj *ga.ForwardingRule, m *cloud.MockForwardingRules) (bool, error) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	projectID := m.ProjectRouter.ProjectID(ctx, meta.VersionGA, "forwardingRules")
	return convertAndInsertAlphaForwardingRule(key, obj, m.Objects, meta.VersionGA, projectID)
}

// InsertBetaFwdRuleHook mocks inserting a BetaForwardingRule.
func InsertBetaFwdRuleHook(ctx context.Context, key *meta.Key, obj *beta.ForwardingRule, m *cloud.MockForwardingRules) (bool, error) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	projectID := m.ProjectRouter.ProjectID(ctx, meta.VersionBeta, "forwardingRules")
	return convertAndInsertAlphaForwardingRule(key, obj, m.Objects, meta.VersionBeta, projectID)
}

// InsertAlphaFwdRuleHook mocks inserting an AlphaForwardingRule.
func InsertAlphaFwdRuleHook(ctx context.Context, key *meta.Key, obj *alpha.ForwardingRule, m *cloud.MockForwardingRules) (bool, error) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	projectID := m.ProjectRouter.ProjectID(ctx, meta.VersionAlpha, "forwardingRules")
	return convertAndInsertAlphaForwardingRule(key, obj, m.Objects, meta.VersionAlpha, projectID)
}

// AddressAttributes maps from Address key to a map of Instances
type AddressAttributes struct {
	IPCounter int // Used to assign Addresses with no IP a unique IP address
}

func convertAndInsertAlphaAddress(key *meta.Key, obj gceObject, mAddrs map[meta.Key]*cloud.MockAddressesObj, version meta.Version, projectID string, addressAttrs AddressAttributes) (bool, error) {
	if !key.Valid() {
		return false, fmt.Errorf("invalid GCE key (%+v)", key)
	}

	if _, ok := mAddrs[*key]; ok {
		err := &googleapi.Error{
			Code:    http.StatusConflict,
			Message: fmt.Sprintf("MockAddresses %v exists", key),
		}
		return false, err
	}

	enc, err := obj.MarshalJSON()
	if err != nil {
		return false, err
	}
	var addr alpha.Address
	if err := json.Unmarshal(enc, &addr); err != nil {
		return false, err
	}

	// Set default address type if not present.
	if addr.AddressType == "" {
		addr.AddressType = string(cloud.SchemeExternal)
	}

	var existingAddresses []*ga.Address
	for _, obj := range mAddrs {
		existingAddresses = append(existingAddresses, obj.ToGA())
	}

	for _, existingAddr := range existingAddresses {
		if addr.Address == existingAddr.Address {
			msg := fmt.Sprintf("MockAddresses IP %v in use", addr.Address)

			// When the IP is already in use, this call returns a StatusBadRequest
			// if the address is an external address, and StatusConflict if an
			// internal address. This is to be consistent with actual GCE API.
			errorCode := http.StatusConflict
			if addr.AddressType == string(cloud.SchemeExternal) {
				errorCode = http.StatusBadRequest
			}

			return false, &googleapi.Error{Code: errorCode, Message: msg}
		}
	}

	// Set default values used in tests
	addr.Name = key.Name
	if addr.SelfLink == "" {
		addr.SelfLink = cloud.SelfLink(version, projectID, "addresses", key)
	}

	if addr.Address == "" {
		addr.Address = fmt.Sprintf("1.2.3.%d", addressAttrs.IPCounter)
		addressAttrs.IPCounter++
	}

	// Set the default values for the Alpha fields.
	if addr.NetworkTier == "" {
		addr.NetworkTier = cloud.NetworkTierDefault.ToGCEValue()
	}

	mAddrs[*key] = &cloud.MockAddressesObj{Obj: addr}
	return true, nil
}

// InsertAddressHook mocks inserting an Address.
func InsertAddressHook(ctx context.Context, key *meta.Key, obj *ga.Address, m *cloud.MockAddresses) (bool, error) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	projectID := m.ProjectRouter.ProjectID(ctx, meta.VersionGA, "addresses")
	return convertAndInsertAlphaAddress(key, obj, m.Objects, meta.VersionGA, projectID, m.X.(AddressAttributes))
}

// InsertBetaAddressHook mocks inserting a BetaAddress.
func InsertBetaAddressHook(ctx context.Context, key *meta.Key, obj *beta.Address, m *cloud.MockAddresses) (bool, error) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	projectID := m.ProjectRouter.ProjectID(ctx, meta.VersionBeta, "addresses")
	return convertAndInsertAlphaAddress(key, obj, m.Objects, meta.VersionBeta, projectID, m.X.(AddressAttributes))
}

// InsertAlphaAddressHook mocks inserting an Address. Addresses are expected to
// default to Premium tier if no NetworkTier is specified.
func InsertAlphaAddressHook(ctx context.Context, key *meta.Key, obj *alpha.Address, m *cloud.MockAlphaAddresses) (bool, error) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	projectID := m.ProjectRouter.ProjectID(ctx, meta.VersionBeta, "addresses")
	return convertAndInsertAlphaAddress(key, obj, m.Objects, meta.VersionAlpha, projectID, m.X.(AddressAttributes))
}

// InstanceGroupAttributes maps from InstanceGroup key to a map of Instances
type InstanceGroupAttributes struct {
	InstanceMap map[meta.Key]map[string]*ga.InstanceWithNamedPorts
	Lock        *sync.Mutex
}

// AddInstances adds a list of Instances passed by InstanceReference
func (igAttrs *InstanceGroupAttributes) AddInstances(key *meta.Key, instanceRefs []*ga.InstanceReference) error {
	igAttrs.Lock.Lock()
	defer igAttrs.Lock.Unlock()

	instancesWithNamedPorts, ok := igAttrs.InstanceMap[*key]
	if !ok {
		instancesWithNamedPorts = make(map[string]*ga.InstanceWithNamedPorts)
	}

	for _, instance := range instanceRefs {
		iWithPort := &ga.InstanceWithNamedPorts{
			Instance: instance.Instance,
		}

		instancesWithNamedPorts[instance.Instance] = iWithPort
	}

	igAttrs.InstanceMap[*key] = instancesWithNamedPorts
	return nil
}

// RemoveInstances removes a list of Instances passed by InstanceReference
func (igAttrs *InstanceGroupAttributes) RemoveInstances(key *meta.Key, instanceRefs []*ga.InstanceReference) error {
	igAttrs.Lock.Lock()
	defer igAttrs.Lock.Unlock()

	instancesWithNamedPorts, ok := igAttrs.InstanceMap[*key]
	if !ok {
		instancesWithNamedPorts = make(map[string]*ga.InstanceWithNamedPorts)
	}

	for _, instanceToRemove := range instanceRefs {
		if _, ok := instancesWithNamedPorts[instanceToRemove.Instance]; ok {
			delete(instancesWithNamedPorts, instanceToRemove.Instance)
		} else {
			return &googleapi.Error{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("%s is not a member of %s", instanceToRemove.Instance, key.String()),
			}
		}
	}

	igAttrs.InstanceMap[*key] = instancesWithNamedPorts
	return nil
}

// List gets a list of InstanceWithNamedPorts
func (igAttrs *InstanceGroupAttributes) List(key *meta.Key) []*ga.InstanceWithNamedPorts {
	igAttrs.Lock.Lock()
	defer igAttrs.Lock.Unlock()

	instancesWithNamedPorts, ok := igAttrs.InstanceMap[*key]
	if !ok {
		instancesWithNamedPorts = make(map[string]*ga.InstanceWithNamedPorts)
	}

	var instanceList []*ga.InstanceWithNamedPorts
	for _, val := range instancesWithNamedPorts {
		instanceList = append(instanceList, val)
	}

	return instanceList
}

// AddInstancesHook mocks adding instances from an InstanceGroup
func AddInstancesHook(ctx context.Context, key *meta.Key, req *ga.InstanceGroupsAddInstancesRequest, m *cloud.MockInstanceGroups) error {
	_, err := m.Get(ctx, key)
	if err != nil {
		return &googleapi.Error{
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("Key: %s was not found in InstanceGroups", key.String()),
		}
	}

	var attrs InstanceGroupAttributes
	attrs = m.X.(InstanceGroupAttributes)
	attrs.AddInstances(key, req.Instances)
	m.X = attrs
	return nil
}

// ListInstancesHook mocks listing instances from an InstanceGroup
func ListInstancesHook(ctx context.Context, key *meta.Key, req *ga.InstanceGroupsListInstancesRequest, filter *filter.F, m *cloud.MockInstanceGroups) ([]*ga.InstanceWithNamedPorts, error) {
	_, err := m.Get(ctx, key)
	if err != nil {
		return nil, &googleapi.Error{
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("Key: %s was not found in InstanceGroups", key.String()),
		}
	}

	var attrs InstanceGroupAttributes
	attrs = m.X.(InstanceGroupAttributes)
	instances := attrs.List(key)

	return instances, nil
}

// RemoveInstancesHook mocks removing instances from an InstanceGroup
func RemoveInstancesHook(ctx context.Context, key *meta.Key, req *ga.InstanceGroupsRemoveInstancesRequest, m *cloud.MockInstanceGroups) error {
	_, err := m.Get(ctx, key)
	if err != nil {
		return &googleapi.Error{
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("Key: %s was not found in InstanceGroups", key.String()),
		}
	}

	var attrs InstanceGroupAttributes
	attrs = m.X.(InstanceGroupAttributes)
	attrs.RemoveInstances(key, req.Instances)
	m.X = attrs
	return nil
}

// UpdateFirewallHook defines the hook for updating a Firewall. It replaces the
// object with the same key in the mock with the updated object.
func UpdateFirewallHook(ctx context.Context, key *meta.Key, obj *ga.Firewall, m *cloud.MockFirewalls) error {
	_, err := m.Get(ctx, key)
	if err != nil {
		return &googleapi.Error{
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("Key: %s was not found in Firewalls", key.String()),
		}
	}

	obj.Name = key.Name
	projectID := m.ProjectRouter.ProjectID(ctx, "ga", "firewalls")
	obj.SelfLink = cloud.SelfLink(meta.VersionGA, projectID, "firewalls", key)

	m.Objects[*key] = &cloud.MockFirewallsObj{Obj: obj}
	return nil
}

// UpdateHealthCheckHook defines the hook for updating a HealthCheck. It
// replaces the object with the same key in the mock with the updated object.
func UpdateHealthCheckHook(ctx context.Context, key *meta.Key, obj *ga.HealthCheck, m *cloud.MockHealthChecks) error {
	_, err := m.Get(ctx, key)
	if err != nil {
		return &googleapi.Error{
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("Key: %s was not found in HealthChecks", key.String()),
		}
	}

	obj.Name = key.Name
	projectID := m.ProjectRouter.ProjectID(ctx, "ga", "healthChecks")
	obj.SelfLink = cloud.SelfLink(meta.VersionGA, projectID, "healthChecks", key)

	m.Objects[*key] = &cloud.MockHealthChecksObj{Obj: obj}
	return nil
}

// UpdateRegionBackendServiceHook defines the hook for updating a Region
// BackendsService. It replaces the object with the same key in the mock with
// the updated object.
func UpdateRegionBackendServiceHook(ctx context.Context, key *meta.Key, obj *ga.BackendService, m *cloud.MockRegionBackendServices) error {
	_, err := m.Get(ctx, key)
	if err != nil {
		return &googleapi.Error{
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("Key: %s was not found in RegionBackendServices", key.String()),
		}
	}

	obj.Name = key.Name
	projectID := m.ProjectRouter.ProjectID(ctx, "ga", "backendServices")
	obj.SelfLink = cloud.SelfLink(meta.VersionGA, projectID, "backendServices", key)

	m.Objects[*key] = &cloud.MockRegionBackendServicesObj{Obj: obj}
	return nil
}

// InsertFirewallsUnauthorizedErrHook mocks firewall insertion. A forbidden error will be thrown as return.
func InsertFirewallsUnauthorizedErrHook(ctx context.Context, key *meta.Key, obj *ga.Firewall, m *cloud.MockFirewalls) (bool, error) {
	return true, &googleapi.Error{Code: http.StatusForbidden}
}

// UpdateFirewallsUnauthorizedErrHook mocks firewall updating. A forbidden error will be thrown as return.
func UpdateFirewallsUnauthorizedErrHook(ctx context.Context, key *meta.Key, obj *ga.Firewall, m *cloud.MockFirewalls) error {
	return &googleapi.Error{Code: http.StatusForbidden}
}

// DeleteFirewallsUnauthorizedErrHook mocks firewall deletion. A forbidden error will be thrown as return.
func DeleteFirewallsUnauthorizedErrHook(ctx context.Context, key *meta.Key, m *cloud.MockFirewalls) (bool, error) {
	return true, &googleapi.Error{Code: http.StatusForbidden}
}

// GetFirewallsUnauthorizedErrHook mocks firewall information retrival. A forbidden error will be thrown as return.
func GetFirewallsUnauthorizedErrHook(ctx context.Context, key *meta.Key, m *cloud.MockFirewalls) (bool, *ga.Firewall, error) {
	return true, nil, &googleapi.Error{Code: http.StatusForbidden}
}

// GetTargetPoolInternalErrHook mocks getting target pool. It returns a internal server error.
func GetTargetPoolInternalErrHook(ctx context.Context, key *meta.Key, m *cloud.MockTargetPools) (bool, *ga.TargetPool, error) {
	return true, nil, InternalServerError
}

// GetForwardingRulesInternalErrHook mocks getting forwarding rules and returns an internal server error.
func GetForwardingRulesInternalErrHook(ctx context.Context, key *meta.Key, m *cloud.MockForwardingRules) (bool, *ga.ForwardingRule, error) {
	return true, nil, InternalServerError
}

// GetAddressesInternalErrHook mocks getting network address and returns an internal server error.
func GetAddressesInternalErrHook(ctx context.Context, key *meta.Key, m *cloud.MockAddresses) (bool, *ga.Address, error) {
	return true, nil, InternalServerError
}

// GetHTTPHealthChecksInternalErrHook mocks getting http health check and returns an internal server error.
func GetHTTPHealthChecksInternalErrHook(ctx context.Context, key *meta.Key, m *cloud.MockHttpHealthChecks) (bool, *ga.HttpHealthCheck, error) {
	return true, nil, InternalServerError
}

// InsertTargetPoolsInternalErrHook mocks getting target pool and returns an internal server error.
func InsertTargetPoolsInternalErrHook(ctx context.Context, key *meta.Key, obj *ga.TargetPool, m *cloud.MockTargetPools) (bool, error) {
	return true, InternalServerError
}

// InsertForwardingRulesInternalErrHook mocks getting forwarding rule and returns an internal server error.
func InsertForwardingRulesInternalErrHook(ctx context.Context, key *meta.Key, obj *ga.ForwardingRule, m *cloud.MockForwardingRules) (bool, error) {
	return true, InternalServerError
}

// DeleteAddressesNotFoundErrHook mocks deleting network address and returns a not found error.
func DeleteAddressesNotFoundErrHook(ctx context.Context, key *meta.Key, m *cloud.MockAddresses) (bool, error) {
	return true, &googleapi.Error{Code: http.StatusNotFound}
}

// DeleteAddressesInternalErrHook mocks deleting address and returns an internal server error.
func DeleteAddressesInternalErrHook(ctx context.Context, key *meta.Key, m *cloud.MockAddresses) (bool, error) {
	return true, InternalServerError
}

// GetRegionBackendServicesErrHook mocks getting region backend service and returns an internal server error.
func GetRegionBackendServicesErrHook(ctx context.Context, key *meta.Key, m *cloud.MockRegionBackendServices) (bool, *ga.BackendService, error) {
	return true, nil, InternalServerError
}

// UpdateRegionBackendServicesErrHook mocks updating a reegion backend service and returns an internal server error.
func UpdateRegionBackendServicesErrHook(ctx context.Context, key *meta.Key, svc *ga.BackendService, m *cloud.MockRegionBackendServices) error {
	return InternalServerError
}

// DeleteRegionBackendServicesErrHook mocks deleting region backend service and returns an internal server error.
func DeleteRegionBackendServicesErrHook(ctx context.Context, key *meta.Key, m *cloud.MockRegionBackendServices) (bool, error) {
	return true, InternalServerError
}

// DeleteRegionBackendServicesInUseErrHook mocks deleting region backend service and returns an InUseError.
func DeleteRegionBackendServicesInUseErrHook(ctx context.Context, key *meta.Key, m *cloud.MockRegionBackendServices) (bool, error) {
	return true, InUseError
}

// GetInstanceGroupInternalErrHook mocks getting instance group and returns an internal server error.
func GetInstanceGroupInternalErrHook(ctx context.Context, key *meta.Key, m *cloud.MockInstanceGroups) (bool, *ga.InstanceGroup, error) {
	return true, nil, InternalServerError
}

// GetHealthChecksInternalErrHook mocks getting health check and returns an internal server erorr.
func GetHealthChecksInternalErrHook(ctx context.Context, key *meta.Key, m *cloud.MockHealthChecks) (bool, *ga.HealthCheck, error) {
	return true, nil, InternalServerError
}

// DeleteHealthChecksInternalErrHook mocks deleting health check and returns an internal server error.
func DeleteHealthChecksInternalErrHook(ctx context.Context, key *meta.Key, m *cloud.MockHealthChecks) (bool, error) {
	return true, InternalServerError
}

// DeleteHealthChecksInuseErrHook mocks deleting health check and returns an in use error.
func DeleteHealthChecksInuseErrHook(ctx context.Context, key *meta.Key, m *cloud.MockHealthChecks) (bool, error) {
	return true, InUseError
}

// DeleteForwardingRuleErrHook mocks deleting forwarding rule and returns an internal server error.
func DeleteForwardingRuleErrHook(ctx context.Context, key *meta.Key, m *cloud.MockForwardingRules) (bool, error) {
	return true, InternalServerError
}

// ListZonesInternalErrHook mocks listing zone and returns an internal server error.
func ListZonesInternalErrHook(ctx context.Context, fl *filter.F, m *cloud.MockZones) (bool, []*ga.Zone, error) {
	return true, []*ga.Zone{}, InternalServerError
}

// DeleteInstanceGroupInternalErrHook mocks deleting instance group and returns an internal server error.
func DeleteInstanceGroupInternalErrHook(ctx context.Context, key *meta.Key, m *cloud.MockInstanceGroups) (bool, error) {
	return true, InternalServerError
}
