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

package azure

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2017-12-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2017-09-01/network"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2017-10-01/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/golang/glog"

	"k8s.io/client-go/util/flowcontrol"
)

// Helpers for rate limiting error/error channel creation
func createARMRateLimitErr(isWrite bool, opName string) error {
	opType := "read"
	if isWrite {
		opType = "write"
	}
	return fmt.Errorf("azure - ARM rate limited(%s) for operation:%s", opType, opName)
}

// VirtualMachinesClient defines needed functions for azure compute.VirtualMachinesClient
type VirtualMachinesClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName string, VMName string, parameters compute.VirtualMachine) (resp *http.Response, err error)
	Get(ctx context.Context, resourceGroupName string, VMName string, expand compute.InstanceViewTypes) (result compute.VirtualMachine, err error)
	List(ctx context.Context, resourceGroupName string) (result []compute.VirtualMachine, err error)
}

// InterfacesClient defines needed functions for azure network.InterfacesClient
type InterfacesClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName string, networkInterfaceName string, parameters network.Interface) (resp *http.Response, err error)
	Get(ctx context.Context, resourceGroupName string, networkInterfaceName string, expand string) (result network.Interface, err error)
	GetVirtualMachineScaleSetNetworkInterface(ctx context.Context, resourceGroupName string, virtualMachineScaleSetName string, virtualmachineIndex string, networkInterfaceName string, expand string) (result network.Interface, err error)
}

// LoadBalancersClient defines needed functions for azure network.LoadBalancersClient
type LoadBalancersClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName string, loadBalancerName string, parameters network.LoadBalancer) (resp *http.Response, err error)
	Delete(ctx context.Context, resourceGroupName string, loadBalancerName string) (resp *http.Response, err error)
	Get(ctx context.Context, resourceGroupName string, loadBalancerName string, expand string) (result network.LoadBalancer, err error)
	List(ctx context.Context, resourceGroupName string) (result []network.LoadBalancer, err error)
}

// PublicIPAddressesClient defines needed functions for azure network.PublicIPAddressesClient
type PublicIPAddressesClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName string, publicIPAddressName string, parameters network.PublicIPAddress) (resp *http.Response, err error)
	Delete(ctx context.Context, resourceGroupName string, publicIPAddressName string) (resp *http.Response, err error)
	Get(ctx context.Context, resourceGroupName string, publicIPAddressName string, expand string) (result network.PublicIPAddress, err error)
	List(ctx context.Context, resourceGroupName string) (result []network.PublicIPAddress, err error)
}

// SubnetsClient defines needed functions for azure network.SubnetsClient
type SubnetsClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName string, virtualNetworkName string, subnetName string, subnetParameters network.Subnet) (resp *http.Response, err error)
	Delete(ctx context.Context, resourceGroupName string, virtualNetworkName string, subnetName string) (resp *http.Response, err error)
	Get(ctx context.Context, resourceGroupName string, virtualNetworkName string, subnetName string, expand string) (result network.Subnet, err error)
	List(ctx context.Context, resourceGroupName string, virtualNetworkName string) (result []network.Subnet, err error)
}

// SecurityGroupsClient defines needed functions for azure network.SecurityGroupsClient
type SecurityGroupsClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName string, networkSecurityGroupName string, parameters network.SecurityGroup) (resp *http.Response, err error)
	Delete(ctx context.Context, resourceGroupName string, networkSecurityGroupName string) (resp *http.Response, err error)
	Get(ctx context.Context, resourceGroupName string, networkSecurityGroupName string, expand string) (result network.SecurityGroup, err error)
	List(ctx context.Context, resourceGroupName string) (result []network.SecurityGroup, err error)
}

// VirtualMachineScaleSetsClient defines needed functions for azure compute.VirtualMachineScaleSetsClient
type VirtualMachineScaleSetsClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName string, VMScaleSetName string, parameters compute.VirtualMachineScaleSet) (resp *http.Response, err error)
	Get(ctx context.Context, resourceGroupName string, VMScaleSetName string) (result compute.VirtualMachineScaleSet, err error)
	List(ctx context.Context, resourceGroupName string) (result []compute.VirtualMachineScaleSet, err error)
	UpdateInstances(ctx context.Context, resourceGroupName string, VMScaleSetName string, VMInstanceIDs compute.VirtualMachineScaleSetVMInstanceRequiredIDs) (resp *http.Response, err error)
}

// VirtualMachineScaleSetVMsClient defines needed functions for azure compute.VirtualMachineScaleSetVMsClient
type VirtualMachineScaleSetVMsClient interface {
	Get(ctx context.Context, resourceGroupName string, VMScaleSetName string, instanceID string) (result compute.VirtualMachineScaleSetVM, err error)
	GetInstanceView(ctx context.Context, resourceGroupName string, VMScaleSetName string, instanceID string) (result compute.VirtualMachineScaleSetVMInstanceView, err error)
	List(ctx context.Context, resourceGroupName string, virtualMachineScaleSetName string, filter string, selectParameter string, expand string) (result []compute.VirtualMachineScaleSetVM, err error)
	Update(ctx context.Context, resourceGroupName string, VMScaleSetName string, instanceID string, parameters compute.VirtualMachineScaleSetVM) (resp *http.Response, err error)
}

// RoutesClient defines needed functions for azure network.RoutesClient
type RoutesClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName string, routeTableName string, routeName string, routeParameters network.Route) (resp *http.Response, err error)
	Delete(ctx context.Context, resourceGroupName string, routeTableName string, routeName string) (resp *http.Response, err error)
}

// RouteTablesClient defines needed functions for azure network.RouteTablesClient
type RouteTablesClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName string, routeTableName string, parameters network.RouteTable) (resp *http.Response, err error)
	Get(ctx context.Context, resourceGroupName string, routeTableName string, expand string) (result network.RouteTable, err error)
}

// StorageAccountClient defines needed functions for azure storage.AccountsClient
type StorageAccountClient interface {
	Create(ctx context.Context, resourceGroupName string, accountName string, parameters storage.AccountCreateParameters) (result *http.Response, err error)
	Delete(ctx context.Context, resourceGroupName string, accountName string) (result autorest.Response, err error)
	ListKeys(ctx context.Context, resourceGroupName string, accountName string) (result storage.AccountListKeysResult, err error)
	ListByResourceGroup(ctx context.Context, resourceGroupName string) (result storage.AccountListResult, err error)
	GetProperties(ctx context.Context, resourceGroupName string, accountName string) (result storage.Account, err error)
}

// DisksClient defines needed functions for azure compute.DisksClient
type DisksClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName string, diskName string, diskParameter compute.Disk) (resp *http.Response, err error)
	Delete(ctx context.Context, resourceGroupName string, diskName string) (resp *http.Response, err error)
	Get(ctx context.Context, resourceGroupName string, diskName string) (result compute.Disk, err error)
}

// azClientConfig contains all essential information to create an Azure client.
type azClientConfig struct {
	subscriptionID          string
	resourceManagerEndpoint string
	servicePrincipalToken   *adal.ServicePrincipalToken
	// ARM Rate limiting for GET vs PUT/POST
	//Details: https://docs.microsoft.com/en-us/azure/azure-resource-manager/resource-manager-request-limits
	rateLimiterReader flowcontrol.RateLimiter
	rateLimiterWriter flowcontrol.RateLimiter
}

// azVirtualMachinesClient implements VirtualMachinesClient.
type azVirtualMachinesClient struct {
	client            compute.VirtualMachinesClient
	rateLimiterReader flowcontrol.RateLimiter
	rateLimiterWriter flowcontrol.RateLimiter
}

func getContextWithCancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

func newAzVirtualMachinesClient(config *azClientConfig) *azVirtualMachinesClient {
	virtualMachinesClient := compute.NewVirtualMachinesClient(config.subscriptionID)
	virtualMachinesClient.BaseURI = config.resourceManagerEndpoint
	virtualMachinesClient.Authorizer = autorest.NewBearerAuthorizer(config.servicePrincipalToken)
	virtualMachinesClient.PollingDelay = 5 * time.Second
	configureUserAgent(&virtualMachinesClient.Client)

	return &azVirtualMachinesClient{
		rateLimiterReader: config.rateLimiterReader,
		rateLimiterWriter: config.rateLimiterWriter,
		client:            virtualMachinesClient,
	}
}

func (az *azVirtualMachinesClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, VMName string, parameters compute.VirtualMachine) (resp *http.Response, err error) {
	// /* Write rate limiting */
	if !az.rateLimiterWriter.TryAccept() {
		err = createARMRateLimitErr(true, "VMCreateOrUpdate")
		return
	}

	glog.V(10).Infof("azVirtualMachinesClient.CreateOrUpdate(%q, %q): start", resourceGroupName, VMName)
	defer func() {
		glog.V(10).Infof("azVirtualMachinesClient.CreateOrUpdate(%q, %q): end", resourceGroupName, VMName)
	}()

	mc := newMetricContext("vm", "create_or_update", resourceGroupName, az.client.SubscriptionID)
	future, err := az.client.CreateOrUpdate(ctx, resourceGroupName, VMName, parameters)
	if err != nil {
		return future.Response(), err
	}

	err = future.WaitForCompletion(ctx, az.client.Client)
	mc.Observe(err)
	return future.Response(), err
}

func (az *azVirtualMachinesClient) Get(ctx context.Context, resourceGroupName string, VMName string, expand compute.InstanceViewTypes) (result compute.VirtualMachine, err error) {
	if !az.rateLimiterReader.TryAccept() {
		err = createARMRateLimitErr(false, "VMGet")
		return
	}

	glog.V(10).Infof("azVirtualMachinesClient.Get(%q, %q): start", resourceGroupName, VMName)
	defer func() {
		glog.V(10).Infof("azVirtualMachinesClient.Get(%q, %q): end", resourceGroupName, VMName)
	}()

	mc := newMetricContext("vm", "get", resourceGroupName, az.client.SubscriptionID)
	result, err = az.client.Get(ctx, resourceGroupName, VMName, expand)
	mc.Observe(err)
	return
}

func (az *azVirtualMachinesClient) List(ctx context.Context, resourceGroupName string) (result []compute.VirtualMachine, err error) {
	if !az.rateLimiterReader.TryAccept() {
		err = createARMRateLimitErr(false, "VMList")
		return
	}

	glog.V(10).Infof("azVirtualMachinesClient.List(%q): start", resourceGroupName)
	defer func() {
		glog.V(10).Infof("azVirtualMachinesClient.List(%q): end", resourceGroupName)
	}()

	mc := newMetricContext("vm", "list", resourceGroupName, az.client.SubscriptionID)
	iterator, err := az.client.ListComplete(ctx, resourceGroupName)
	mc.Observe(err)
	if err != nil {
		return nil, err
	}

	result = make([]compute.VirtualMachine, 0)
	for ; iterator.NotDone(); err = iterator.Next() {
		if err != nil {
			return nil, err
		}

		result = append(result, iterator.Value())
	}

	return result, nil
}

// azInterfacesClient implements InterfacesClient.
type azInterfacesClient struct {
	client            network.InterfacesClient
	rateLimiterReader flowcontrol.RateLimiter
	rateLimiterWriter flowcontrol.RateLimiter
}

func newAzInterfacesClient(config *azClientConfig) *azInterfacesClient {
	interfacesClient := network.NewInterfacesClient(config.subscriptionID)
	interfacesClient.BaseURI = config.resourceManagerEndpoint
	interfacesClient.Authorizer = autorest.NewBearerAuthorizer(config.servicePrincipalToken)
	interfacesClient.PollingDelay = 5 * time.Second
	configureUserAgent(&interfacesClient.Client)

	return &azInterfacesClient{
		rateLimiterReader: config.rateLimiterReader,
		rateLimiterWriter: config.rateLimiterWriter,
		client:            interfacesClient,
	}
}

func (az *azInterfacesClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, networkInterfaceName string, parameters network.Interface) (resp *http.Response, err error) {
	/* Write rate limiting */
	if !az.rateLimiterWriter.TryAccept() {
		err = createARMRateLimitErr(true, "NiCreateOrUpdate")
		return
	}

	glog.V(10).Infof("azInterfacesClient.CreateOrUpdate(%q,%q): start", resourceGroupName, networkInterfaceName)
	defer func() {
		glog.V(10).Infof("azInterfacesClient.CreateOrUpdate(%q,%q): end", resourceGroupName, networkInterfaceName)
	}()

	mc := newMetricContext("interfaces", "create_or_update", resourceGroupName, az.client.SubscriptionID)
	future, err := az.client.CreateOrUpdate(ctx, resourceGroupName, networkInterfaceName, parameters)
	if err != nil {
		mc.Observe(err)
		return future.Response(), err
	}

	err = future.WaitForCompletion(ctx, az.client.Client)
	mc.Observe(err)
	return future.Response(), err
}

func (az *azInterfacesClient) Get(ctx context.Context, resourceGroupName string, networkInterfaceName string, expand string) (result network.Interface, err error) {
	if !az.rateLimiterReader.TryAccept() {
		err = createARMRateLimitErr(false, "NicGet")
		return
	}

	glog.V(10).Infof("azInterfacesClient.Get(%q,%q): start", resourceGroupName, networkInterfaceName)
	defer func() {
		glog.V(10).Infof("azInterfacesClient.Get(%q,%q): end", resourceGroupName, networkInterfaceName)
	}()

	mc := newMetricContext("interfaces", "get", resourceGroupName, az.client.SubscriptionID)
	result, err = az.client.Get(ctx, resourceGroupName, networkInterfaceName, expand)
	mc.Observe(err)
	return
}

func (az *azInterfacesClient) GetVirtualMachineScaleSetNetworkInterface(ctx context.Context, resourceGroupName string, virtualMachineScaleSetName string, virtualmachineIndex string, networkInterfaceName string, expand string) (result network.Interface, err error) {
	if !az.rateLimiterReader.TryAccept() {
		err = createARMRateLimitErr(false, "NicGetVirtualMachineScaleSetNetworkInterface")
		return
	}

	glog.V(10).Infof("azInterfacesClient.GetVirtualMachineScaleSetNetworkInterface(%q,%q,%q,%q): start", resourceGroupName, virtualMachineScaleSetName, virtualmachineIndex, networkInterfaceName)
	defer func() {
		glog.V(10).Infof("azInterfacesClient.GetVirtualMachineScaleSetNetworkInterface(%q,%q,%q,%q): end", resourceGroupName, virtualMachineScaleSetName, virtualmachineIndex, networkInterfaceName)
	}()

	mc := newMetricContext("interfaces", "get_vmss_ni", resourceGroupName, az.client.SubscriptionID)
	result, err = az.client.GetVirtualMachineScaleSetNetworkInterface(ctx, resourceGroupName, virtualMachineScaleSetName, virtualmachineIndex, networkInterfaceName, expand)
	mc.Observe(err)
	return
}

// azLoadBalancersClient implements LoadBalancersClient.
type azLoadBalancersClient struct {
	client            network.LoadBalancersClient
	rateLimiterReader flowcontrol.RateLimiter
	rateLimiterWriter flowcontrol.RateLimiter
}

func newAzLoadBalancersClient(config *azClientConfig) *azLoadBalancersClient {
	loadBalancerClient := network.NewLoadBalancersClient(config.subscriptionID)
	loadBalancerClient.BaseURI = config.resourceManagerEndpoint
	loadBalancerClient.Authorizer = autorest.NewBearerAuthorizer(config.servicePrincipalToken)
	loadBalancerClient.PollingDelay = 5 * time.Second
	configureUserAgent(&loadBalancerClient.Client)

	return &azLoadBalancersClient{
		rateLimiterReader: config.rateLimiterReader,
		rateLimiterWriter: config.rateLimiterWriter,
		client:            loadBalancerClient,
	}
}

func (az *azLoadBalancersClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, loadBalancerName string, parameters network.LoadBalancer) (resp *http.Response, err error) {
	/* Write rate limiting */
	if !az.rateLimiterWriter.TryAccept() {
		err = createARMRateLimitErr(true, "LBCreateOrUpdate")
		return nil, err
	}

	glog.V(10).Infof("azLoadBalancersClient.CreateOrUpdate(%q,%q): start", resourceGroupName, loadBalancerName)
	defer func() {
		glog.V(10).Infof("azLoadBalancersClient.CreateOrUpdate(%q,%q): end", resourceGroupName, loadBalancerName)
	}()

	mc := newMetricContext("load_balancers", "create_or_update", resourceGroupName, az.client.SubscriptionID)
	future, err := az.client.CreateOrUpdate(ctx, resourceGroupName, loadBalancerName, parameters)
	mc.Observe(err)
	if err != nil {
		return future.Response(), err
	}

	err = future.WaitForCompletion(ctx, az.client.Client)
	mc.Observe(err)
	return future.Response(), err
}

func (az *azLoadBalancersClient) Delete(ctx context.Context, resourceGroupName string, loadBalancerName string) (resp *http.Response, err error) {
	/* Write rate limiting */
	if !az.rateLimiterWriter.TryAccept() {
		err = createARMRateLimitErr(true, "LBDelete")
		return nil, err
	}

	glog.V(10).Infof("azLoadBalancersClient.Delete(%q,%q): start", resourceGroupName, loadBalancerName)
	defer func() {
		glog.V(10).Infof("azLoadBalancersClient.Delete(%q,%q): end", resourceGroupName, loadBalancerName)
	}()

	mc := newMetricContext("load_balancers", "delete", resourceGroupName, az.client.SubscriptionID)
	future, err := az.client.Delete(ctx, resourceGroupName, loadBalancerName)
	mc.Observe(err)
	if err != nil {
		return future.Response(), err
	}

	err = future.WaitForCompletion(ctx, az.client.Client)
	mc.Observe(err)
	return future.Response(), err
}

func (az *azLoadBalancersClient) Get(ctx context.Context, resourceGroupName string, loadBalancerName string, expand string) (result network.LoadBalancer, err error) {
	if !az.rateLimiterReader.TryAccept() {
		err = createARMRateLimitErr(false, "LBGet")
		return
	}

	glog.V(10).Infof("azLoadBalancersClient.Get(%q,%q): start", resourceGroupName, loadBalancerName)
	defer func() {
		glog.V(10).Infof("azLoadBalancersClient.Get(%q,%q): end", resourceGroupName, loadBalancerName)
	}()

	mc := newMetricContext("load_balancers", "get", resourceGroupName, az.client.SubscriptionID)
	result, err = az.client.Get(ctx, resourceGroupName, loadBalancerName, expand)
	mc.Observe(err)
	return
}

func (az *azLoadBalancersClient) List(ctx context.Context, resourceGroupName string) ([]network.LoadBalancer, error) {
	if !az.rateLimiterReader.TryAccept() {
		err := createARMRateLimitErr(false, "LBList")
		return nil, err
	}

	glog.V(10).Infof("azLoadBalancersClient.List(%q): start", resourceGroupName)
	defer func() {
		glog.V(10).Infof("azLoadBalancersClient.List(%q): end", resourceGroupName)
	}()

	mc := newMetricContext("load_balancers", "list", resourceGroupName, az.client.SubscriptionID)
	iterator, err := az.client.ListComplete(ctx, resourceGroupName)
	mc.Observe(err)
	if err != nil {
		return nil, err
	}

	result := make([]network.LoadBalancer, 0)
	for ; iterator.NotDone(); err = iterator.Next() {
		if err != nil {
			return nil, err
		}

		result = append(result, iterator.Value())
	}

	return result, nil
}

// azPublicIPAddressesClient implements PublicIPAddressesClient.
type azPublicIPAddressesClient struct {
	client            network.PublicIPAddressesClient
	rateLimiterReader flowcontrol.RateLimiter
	rateLimiterWriter flowcontrol.RateLimiter
}

func newAzPublicIPAddressesClient(config *azClientConfig) *azPublicIPAddressesClient {
	publicIPAddressClient := network.NewPublicIPAddressesClient(config.subscriptionID)
	publicIPAddressClient.BaseURI = config.resourceManagerEndpoint
	publicIPAddressClient.Authorizer = autorest.NewBearerAuthorizer(config.servicePrincipalToken)
	publicIPAddressClient.PollingDelay = 5 * time.Second
	configureUserAgent(&publicIPAddressClient.Client)

	return &azPublicIPAddressesClient{
		rateLimiterReader: config.rateLimiterReader,
		rateLimiterWriter: config.rateLimiterWriter,
		client:            publicIPAddressClient,
	}
}

func (az *azPublicIPAddressesClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, publicIPAddressName string, parameters network.PublicIPAddress) (resp *http.Response, err error) {
	/* Write rate limiting */
	if !az.rateLimiterWriter.TryAccept() {
		err = createARMRateLimitErr(true, "PublicIPCreateOrUpdate")
		return nil, err
	}

	glog.V(10).Infof("azPublicIPAddressesClient.CreateOrUpdate(%q,%q): start", resourceGroupName, publicIPAddressName)
	defer func() {
		glog.V(10).Infof("azPublicIPAddressesClient.CreateOrUpdate(%q,%q): end", resourceGroupName, publicIPAddressName)
	}()

	mc := newMetricContext("public_ip_addresses", "create_or_update", resourceGroupName, az.client.SubscriptionID)
	future, err := az.client.CreateOrUpdate(ctx, resourceGroupName, publicIPAddressName, parameters)
	mc.Observe(err)
	if err != nil {
		return future.Response(), err
	}

	err = future.WaitForCompletion(ctx, az.client.Client)
	mc.Observe(err)
	return future.Response(), err
}

func (az *azPublicIPAddressesClient) Delete(ctx context.Context, resourceGroupName string, publicIPAddressName string) (resp *http.Response, err error) {
	/* Write rate limiting */
	if !az.rateLimiterWriter.TryAccept() {
		err = createARMRateLimitErr(true, "PublicIPDelete")
		return nil, err
	}

	glog.V(10).Infof("azPublicIPAddressesClient.Delete(%q,%q): start", resourceGroupName, publicIPAddressName)
	defer func() {
		glog.V(10).Infof("azPublicIPAddressesClient.Delete(%q,%q): end", resourceGroupName, publicIPAddressName)
	}()

	mc := newMetricContext("public_ip_addresses", "delete", resourceGroupName, az.client.SubscriptionID)
	future, err := az.client.Delete(ctx, resourceGroupName, publicIPAddressName)
	mc.Observe(err)
	if err != nil {
		return future.Response(), err
	}

	err = future.WaitForCompletion(ctx, az.client.Client)
	mc.Observe(err)
	return future.Response(), err
}

func (az *azPublicIPAddressesClient) Get(ctx context.Context, resourceGroupName string, publicIPAddressName string, expand string) (result network.PublicIPAddress, err error) {
	if !az.rateLimiterReader.TryAccept() {
		err = createARMRateLimitErr(false, "PublicIPGet")
		return
	}

	glog.V(10).Infof("azPublicIPAddressesClient.Get(%q,%q): start", resourceGroupName, publicIPAddressName)
	defer func() {
		glog.V(10).Infof("azPublicIPAddressesClient.Get(%q,%q): end", resourceGroupName, publicIPAddressName)
	}()

	mc := newMetricContext("public_ip_addresses", "get", resourceGroupName, az.client.SubscriptionID)
	result, err = az.client.Get(ctx, resourceGroupName, publicIPAddressName, expand)
	mc.Observe(err)
	return
}

func (az *azPublicIPAddressesClient) List(ctx context.Context, resourceGroupName string) ([]network.PublicIPAddress, error) {
	if !az.rateLimiterReader.TryAccept() {
		return nil, createARMRateLimitErr(false, "PublicIPList")
	}

	glog.V(10).Infof("azPublicIPAddressesClient.List(%q): start", resourceGroupName)
	defer func() {
		glog.V(10).Infof("azPublicIPAddressesClient.List(%q): end", resourceGroupName)
	}()

	mc := newMetricContext("public_ip_addresses", "list", resourceGroupName, az.client.SubscriptionID)
	iterator, err := az.client.ListComplete(ctx, resourceGroupName)
	mc.Observe(err)
	if err != nil {
		return nil, err
	}

	result := make([]network.PublicIPAddress, 0)
	for ; iterator.NotDone(); err = iterator.Next() {
		if err != nil {
			return nil, err
		}

		result = append(result, iterator.Value())
	}

	return result, nil
}

// azSubnetsClient implements SubnetsClient.
type azSubnetsClient struct {
	client            network.SubnetsClient
	rateLimiterReader flowcontrol.RateLimiter
	rateLimiterWriter flowcontrol.RateLimiter
}

func newAzSubnetsClient(config *azClientConfig) *azSubnetsClient {
	subnetsClient := network.NewSubnetsClient(config.subscriptionID)
	subnetsClient.BaseURI = config.resourceManagerEndpoint
	subnetsClient.Authorizer = autorest.NewBearerAuthorizer(config.servicePrincipalToken)
	subnetsClient.PollingDelay = 5 * time.Second
	configureUserAgent(&subnetsClient.Client)

	return &azSubnetsClient{
		client:            subnetsClient,
		rateLimiterReader: config.rateLimiterReader,
		rateLimiterWriter: config.rateLimiterWriter,
	}
}

func (az *azSubnetsClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, virtualNetworkName string, subnetName string, subnetParameters network.Subnet) (resp *http.Response, err error) {
	/* Write rate limiting */
	if !az.rateLimiterWriter.TryAccept() {
		err = createARMRateLimitErr(true, "SubnetCreateOrUpdate")
		return
	}

	glog.V(10).Infof("azSubnetsClient.CreateOrUpdate(%q,%q,%q): start", resourceGroupName, virtualNetworkName, subnetName)
	defer func() {
		glog.V(10).Infof("azSubnetsClient.CreateOrUpdate(%q,%q,%q): end", resourceGroupName, virtualNetworkName, subnetName)
	}()

	mc := newMetricContext("subnets", "create_or_update", resourceGroupName, az.client.SubscriptionID)
	future, err := az.client.CreateOrUpdate(ctx, resourceGroupName, virtualNetworkName, subnetName, subnetParameters)
	if err != nil {
		mc.Observe(err)
		return future.Response(), err
	}

	err = future.WaitForCompletion(ctx, az.client.Client)
	mc.Observe(err)
	return future.Response(), err
}

func (az *azSubnetsClient) Delete(ctx context.Context, resourceGroupName string, virtualNetworkName string, subnetName string) (resp *http.Response, err error) {
	/* Write rate limiting */
	if !az.rateLimiterWriter.TryAccept() {
		err = createARMRateLimitErr(true, "SubnetDelete")
		return
	}

	glog.V(10).Infof("azSubnetsClient.Delete(%q,%q,%q): start", resourceGroupName, virtualNetworkName, subnetName)
	defer func() {
		glog.V(10).Infof("azSubnetsClient.Delete(%q,%q,%q): end", resourceGroupName, virtualNetworkName, subnetName)
	}()

	mc := newMetricContext("subnets", "delete", resourceGroupName, az.client.SubscriptionID)
	future, err := az.client.Delete(ctx, resourceGroupName, virtualNetworkName, subnetName)
	if err != nil {
		mc.Observe(err)
		return future.Response(), err
	}

	err = future.WaitForCompletion(ctx, az.client.Client)
	mc.Observe(err)
	return future.Response(), err
}

func (az *azSubnetsClient) Get(ctx context.Context, resourceGroupName string, virtualNetworkName string, subnetName string, expand string) (result network.Subnet, err error) {
	if !az.rateLimiterReader.TryAccept() {
		err = createARMRateLimitErr(false, "SubnetGet")
		return
	}

	glog.V(10).Infof("azSubnetsClient.Get(%q,%q,%q): start", resourceGroupName, virtualNetworkName, subnetName)
	defer func() {
		glog.V(10).Infof("azSubnetsClient.Get(%q,%q,%q): end", resourceGroupName, virtualNetworkName, subnetName)
	}()

	mc := newMetricContext("subnets", "get", resourceGroupName, az.client.SubscriptionID)
	result, err = az.client.Get(ctx, resourceGroupName, virtualNetworkName, subnetName, expand)
	mc.Observe(err)
	return
}

func (az *azSubnetsClient) List(ctx context.Context, resourceGroupName string, virtualNetworkName string) ([]network.Subnet, error) {
	if !az.rateLimiterReader.TryAccept() {
		return nil, createARMRateLimitErr(false, "SubnetList")
	}

	glog.V(10).Infof("azSubnetsClient.List(%q,%q): start", resourceGroupName, virtualNetworkName)
	defer func() {
		glog.V(10).Infof("azSubnetsClient.List(%q,%q): end", resourceGroupName, virtualNetworkName)
	}()

	mc := newMetricContext("subnets", "list", resourceGroupName, az.client.SubscriptionID)
	iterator, err := az.client.ListComplete(ctx, resourceGroupName, virtualNetworkName)
	if err != nil {
		mc.Observe(err)
		return nil, err
	}

	result := make([]network.Subnet, 0)
	for ; iterator.NotDone(); err = iterator.Next() {
		if err != nil {
			return nil, err
		}

		result = append(result, iterator.Value())
	}

	return result, nil
}

// azSecurityGroupsClient implements SecurityGroupsClient.
type azSecurityGroupsClient struct {
	client            network.SecurityGroupsClient
	rateLimiterReader flowcontrol.RateLimiter
	rateLimiterWriter flowcontrol.RateLimiter
}

func newAzSecurityGroupsClient(config *azClientConfig) *azSecurityGroupsClient {
	securityGroupsClient := network.NewSecurityGroupsClient(config.subscriptionID)
	securityGroupsClient.BaseURI = config.resourceManagerEndpoint
	securityGroupsClient.Authorizer = autorest.NewBearerAuthorizer(config.servicePrincipalToken)
	securityGroupsClient.PollingDelay = 5 * time.Second
	configureUserAgent(&securityGroupsClient.Client)

	return &azSecurityGroupsClient{
		client:            securityGroupsClient,
		rateLimiterReader: config.rateLimiterReader,
		rateLimiterWriter: config.rateLimiterWriter,
	}
}

func (az *azSecurityGroupsClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, networkSecurityGroupName string, parameters network.SecurityGroup) (resp *http.Response, err error) {
	/* Write rate limiting */
	if !az.rateLimiterWriter.TryAccept() {
		err = createARMRateLimitErr(true, "NSGCreateOrUpdate")
		return
	}

	glog.V(10).Infof("azSecurityGroupsClient.CreateOrUpdate(%q,%q): start", resourceGroupName, networkSecurityGroupName)
	defer func() {
		glog.V(10).Infof("azSecurityGroupsClient.CreateOrUpdate(%q,%q): end", resourceGroupName, networkSecurityGroupName)
	}()

	mc := newMetricContext("security_groups", "create_or_update", resourceGroupName, az.client.SubscriptionID)
	future, err := az.client.CreateOrUpdate(ctx, resourceGroupName, networkSecurityGroupName, parameters)
	if err != nil {
		mc.Observe(err)
		return future.Response(), err
	}

	err = future.WaitForCompletion(ctx, az.client.Client)
	mc.Observe(err)
	return future.Response(), err
}

func (az *azSecurityGroupsClient) Delete(ctx context.Context, resourceGroupName string, networkSecurityGroupName string) (resp *http.Response, err error) {
	/* Write rate limiting */
	if !az.rateLimiterWriter.TryAccept() {
		err = createARMRateLimitErr(true, "NSGDelete")
		return
	}

	glog.V(10).Infof("azSecurityGroupsClient.Delete(%q,%q): start", resourceGroupName, networkSecurityGroupName)
	defer func() {
		glog.V(10).Infof("azSecurityGroupsClient.Delete(%q,%q): end", resourceGroupName, networkSecurityGroupName)
	}()

	mc := newMetricContext("security_groups", "delete", resourceGroupName, az.client.SubscriptionID)
	future, err := az.client.Delete(ctx, resourceGroupName, networkSecurityGroupName)
	if err != nil {
		mc.Observe(err)
		return future.Response(), err
	}

	err = future.WaitForCompletion(ctx, az.client.Client)
	mc.Observe(err)
	return future.Response(), err
}

func (az *azSecurityGroupsClient) Get(ctx context.Context, resourceGroupName string, networkSecurityGroupName string, expand string) (result network.SecurityGroup, err error) {
	if !az.rateLimiterReader.TryAccept() {
		err = createARMRateLimitErr(false, "NSGGet")
		return
	}

	glog.V(10).Infof("azSecurityGroupsClient.Get(%q,%q): start", resourceGroupName, networkSecurityGroupName)
	defer func() {
		glog.V(10).Infof("azSecurityGroupsClient.Get(%q,%q): end", resourceGroupName, networkSecurityGroupName)
	}()

	mc := newMetricContext("security_groups", "get", resourceGroupName, az.client.SubscriptionID)
	result, err = az.client.Get(ctx, resourceGroupName, networkSecurityGroupName, expand)
	mc.Observe(err)
	return
}

func (az *azSecurityGroupsClient) List(ctx context.Context, resourceGroupName string) ([]network.SecurityGroup, error) {
	if !az.rateLimiterReader.TryAccept() {
		return nil, createARMRateLimitErr(false, "NSGList")
	}

	glog.V(10).Infof("azSecurityGroupsClient.List(%q): start", resourceGroupName)
	defer func() {
		glog.V(10).Infof("azSecurityGroupsClient.List(%q): end", resourceGroupName)
	}()

	mc := newMetricContext("security_groups", "list", resourceGroupName, az.client.SubscriptionID)
	iterator, err := az.client.ListComplete(ctx, resourceGroupName)
	mc.Observe(err)
	if err != nil {
		return nil, err
	}

	result := make([]network.SecurityGroup, 0)
	for ; iterator.NotDone(); err = iterator.Next() {
		if err != nil {
			return nil, err
		}

		result = append(result, iterator.Value())
	}

	return result, nil
}

// azVirtualMachineScaleSetsClient implements VirtualMachineScaleSetsClient.
type azVirtualMachineScaleSetsClient struct {
	client            compute.VirtualMachineScaleSetsClient
	rateLimiterReader flowcontrol.RateLimiter
	rateLimiterWriter flowcontrol.RateLimiter
}

func newAzVirtualMachineScaleSetsClient(config *azClientConfig) *azVirtualMachineScaleSetsClient {
	virtualMachineScaleSetsClient := compute.NewVirtualMachineScaleSetsClient(config.subscriptionID)
	virtualMachineScaleSetsClient.BaseURI = config.resourceManagerEndpoint
	virtualMachineScaleSetsClient.Authorizer = autorest.NewBearerAuthorizer(config.servicePrincipalToken)
	virtualMachineScaleSetsClient.PollingDelay = 5 * time.Second
	configureUserAgent(&virtualMachineScaleSetsClient.Client)

	return &azVirtualMachineScaleSetsClient{
		client:            virtualMachineScaleSetsClient,
		rateLimiterReader: config.rateLimiterReader,
		rateLimiterWriter: config.rateLimiterWriter,
	}
}

func (az *azVirtualMachineScaleSetsClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, VMScaleSetName string, parameters compute.VirtualMachineScaleSet) (resp *http.Response, err error) {
	/* Write rate limiting */
	if !az.rateLimiterWriter.TryAccept() {
		err = createARMRateLimitErr(true, "VMSSCreateOrUpdate")
		return
	}

	glog.V(10).Infof("azVirtualMachineScaleSetsClient.CreateOrUpdate(%q,%q): start", resourceGroupName, VMScaleSetName)
	defer func() {
		glog.V(10).Infof("azVirtualMachineScaleSetsClient.CreateOrUpdate(%q,%q): end", resourceGroupName, VMScaleSetName)
	}()

	mc := newMetricContext("vmss", "create_or_update", resourceGroupName, az.client.SubscriptionID)
	future, err := az.client.CreateOrUpdate(ctx, resourceGroupName, VMScaleSetName, parameters)
	mc.Observe(err)
	if err != nil {
		return future.Response(), err
	}

	err = future.WaitForCompletion(ctx, az.client.Client)
	mc.Observe(err)
	return future.Response(), err
}

func (az *azVirtualMachineScaleSetsClient) Get(ctx context.Context, resourceGroupName string, VMScaleSetName string) (result compute.VirtualMachineScaleSet, err error) {
	if !az.rateLimiterReader.TryAccept() {
		err = createARMRateLimitErr(false, "VMSSGet")
		return
	}

	glog.V(10).Infof("azVirtualMachineScaleSetsClient.Get(%q,%q): start", resourceGroupName, VMScaleSetName)
	defer func() {
		glog.V(10).Infof("azVirtualMachineScaleSetsClient.Get(%q,%q): end", resourceGroupName, VMScaleSetName)
	}()

	mc := newMetricContext("vmss", "get", resourceGroupName, az.client.SubscriptionID)
	result, err = az.client.Get(ctx, resourceGroupName, VMScaleSetName)
	mc.Observe(err)
	return
}

func (az *azVirtualMachineScaleSetsClient) List(ctx context.Context, resourceGroupName string) (result []compute.VirtualMachineScaleSet, err error) {
	if !az.rateLimiterReader.TryAccept() {
		err = createARMRateLimitErr(false, "VMSSList")
		return
	}

	glog.V(10).Infof("azVirtualMachineScaleSetsClient.List(%q,%q): start", resourceGroupName)
	defer func() {
		glog.V(10).Infof("azVirtualMachineScaleSetsClient.List(%q,%q): end", resourceGroupName)
	}()

	mc := newMetricContext("vmss", "list", resourceGroupName, az.client.SubscriptionID)
	iterator, err := az.client.ListComplete(ctx, resourceGroupName)
	mc.Observe(err)
	if err != nil {
		return nil, err
	}

	result = make([]compute.VirtualMachineScaleSet, 0)
	for ; iterator.NotDone(); err = iterator.Next() {
		if err != nil {
			return nil, err
		}

		result = append(result, iterator.Value())
	}

	return result, nil
}

func (az *azVirtualMachineScaleSetsClient) UpdateInstances(ctx context.Context, resourceGroupName string, VMScaleSetName string, VMInstanceIDs compute.VirtualMachineScaleSetVMInstanceRequiredIDs) (resp *http.Response, err error) {
	/* Write rate limiting */
	if !az.rateLimiterWriter.TryAccept() {
		err = createARMRateLimitErr(true, "VMSSUpdateInstances")
		return
	}

	glog.V(10).Infof("azVirtualMachineScaleSetsClient.UpdateInstances(%q,%q,%q): start", resourceGroupName, VMScaleSetName, VMInstanceIDs)
	defer func() {
		glog.V(10).Infof("azVirtualMachineScaleSetsClient.UpdateInstances(%q,%q,%q): end", resourceGroupName, VMScaleSetName, VMInstanceIDs)
	}()

	mc := newMetricContext("vmss", "update_instances", resourceGroupName, az.client.SubscriptionID)
	future, err := az.client.UpdateInstances(ctx, resourceGroupName, VMScaleSetName, VMInstanceIDs)
	mc.Observe(err)
	if err != nil {
		return future.Response(), err
	}

	err = future.WaitForCompletion(ctx, az.client.Client)
	mc.Observe(err)
	return future.Response(), err
}

// azVirtualMachineScaleSetVMsClient implements VirtualMachineScaleSetVMsClient.
type azVirtualMachineScaleSetVMsClient struct {
	client            compute.VirtualMachineScaleSetVMsClient
	rateLimiterReader flowcontrol.RateLimiter
	rateLimiterWriter flowcontrol.RateLimiter
}

func newAzVirtualMachineScaleSetVMsClient(config *azClientConfig) *azVirtualMachineScaleSetVMsClient {
	virtualMachineScaleSetVMsClient := compute.NewVirtualMachineScaleSetVMsClient(config.subscriptionID)
	virtualMachineScaleSetVMsClient.BaseURI = config.resourceManagerEndpoint
	virtualMachineScaleSetVMsClient.Authorizer = autorest.NewBearerAuthorizer(config.servicePrincipalToken)
	virtualMachineScaleSetVMsClient.PollingDelay = 5 * time.Second
	configureUserAgent(&virtualMachineScaleSetVMsClient.Client)

	return &azVirtualMachineScaleSetVMsClient{
		client:            virtualMachineScaleSetVMsClient,
		rateLimiterReader: config.rateLimiterReader,
		rateLimiterWriter: config.rateLimiterWriter,
	}
}

func (az *azVirtualMachineScaleSetVMsClient) Get(ctx context.Context, resourceGroupName string, VMScaleSetName string, instanceID string) (result compute.VirtualMachineScaleSetVM, err error) {
	if !az.rateLimiterReader.TryAccept() {
		err = createARMRateLimitErr(false, "VMSSGet")
		return
	}

	glog.V(10).Infof("azVirtualMachineScaleSetVMsClient.Get(%q,%q,%q): start", resourceGroupName, VMScaleSetName, instanceID)
	defer func() {
		glog.V(10).Infof("azVirtualMachineScaleSetVMsClient.Get(%q,%q,%q): end", resourceGroupName, VMScaleSetName, instanceID)
	}()

	mc := newMetricContext("vmssvm", "get", resourceGroupName, az.client.SubscriptionID)
	result, err = az.client.Get(ctx, resourceGroupName, VMScaleSetName, instanceID)
	mc.Observe(err)
	return
}

func (az *azVirtualMachineScaleSetVMsClient) GetInstanceView(ctx context.Context, resourceGroupName string, VMScaleSetName string, instanceID string) (result compute.VirtualMachineScaleSetVMInstanceView, err error) {
	if !az.rateLimiterReader.TryAccept() {
		err = createARMRateLimitErr(false, "VMSSGetInstanceView")
		return
	}

	glog.V(10).Infof("azVirtualMachineScaleSetVMsClient.GetInstanceView(%q,%q,%q): start", resourceGroupName, VMScaleSetName, instanceID)
	defer func() {
		glog.V(10).Infof("azVirtualMachineScaleSetVMsClient.GetInstanceView(%q,%q,%q): end", resourceGroupName, VMScaleSetName, instanceID)
	}()

	mc := newMetricContext("vmssvm", "get_instance_view", resourceGroupName, az.client.SubscriptionID)
	result, err = az.client.GetInstanceView(ctx, resourceGroupName, VMScaleSetName, instanceID)
	mc.Observe(err)
	return
}

func (az *azVirtualMachineScaleSetVMsClient) List(ctx context.Context, resourceGroupName string, virtualMachineScaleSetName string, filter string, selectParameter string, expand string) (result []compute.VirtualMachineScaleSetVM, err error) {
	if !az.rateLimiterReader.TryAccept() {
		err = createARMRateLimitErr(false, "VMSSList")
		return
	}

	glog.V(10).Infof("azVirtualMachineScaleSetVMsClient.List(%q,%q,%q): start", resourceGroupName, virtualMachineScaleSetName, filter)
	defer func() {
		glog.V(10).Infof("azVirtualMachineScaleSetVMsClient.List(%q,%q,%q): end", resourceGroupName, virtualMachineScaleSetName, filter)
	}()

	mc := newMetricContext("vmssvm", "list", resourceGroupName, az.client.SubscriptionID)
	iterator, err := az.client.ListComplete(ctx, resourceGroupName, virtualMachineScaleSetName, filter, selectParameter, expand)
	mc.Observe(err)
	if err != nil {
		return nil, err
	}

	result = make([]compute.VirtualMachineScaleSetVM, 0)
	for ; iterator.NotDone(); err = iterator.Next() {
		if err != nil {
			return nil, err
		}

		result = append(result, iterator.Value())
	}

	return result, nil
}

func (az *azVirtualMachineScaleSetVMsClient) Update(ctx context.Context, resourceGroupName string, VMScaleSetName string, instanceID string, parameters compute.VirtualMachineScaleSetVM) (resp *http.Response, err error) {
	if !az.rateLimiterWriter.TryAccept() {
		err = createARMRateLimitErr(true, "VMSSUpdate")
		return
	}

	glog.V(10).Infof("azVirtualMachineScaleSetVMsClient.Update(%q,%q,%q): start", resourceGroupName, VMScaleSetName, instanceID)
	defer func() {
		glog.V(10).Infof("azVirtualMachineScaleSetVMsClient.Update(%q,%q,%q): end", resourceGroupName, VMScaleSetName, instanceID)
	}()

	mc := newMetricContext("vmssvm", "update", resourceGroupName, az.client.SubscriptionID)
	future, err := az.client.Update(ctx, resourceGroupName, VMScaleSetName, instanceID, parameters)
	mc.Observe(err)
	if err != nil {
		return future.Response(), err
	}

	err = future.WaitForCompletion(ctx, az.client.Client)
	mc.Observe(err)
	return future.Response(), err
}

// azRoutesClient implements RoutesClient.
type azRoutesClient struct {
	client            network.RoutesClient
	rateLimiterReader flowcontrol.RateLimiter
	rateLimiterWriter flowcontrol.RateLimiter
}

func newAzRoutesClient(config *azClientConfig) *azRoutesClient {
	routesClient := network.NewRoutesClient(config.subscriptionID)
	routesClient.BaseURI = config.resourceManagerEndpoint
	routesClient.Authorizer = autorest.NewBearerAuthorizer(config.servicePrincipalToken)
	routesClient.PollingDelay = 5 * time.Second
	configureUserAgent(&routesClient.Client)

	return &azRoutesClient{
		client:            routesClient,
		rateLimiterReader: config.rateLimiterReader,
		rateLimiterWriter: config.rateLimiterWriter,
	}
}

func (az *azRoutesClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, routeTableName string, routeName string, routeParameters network.Route) (resp *http.Response, err error) {
	/* Write rate limiting */
	if !az.rateLimiterWriter.TryAccept() {
		err = createARMRateLimitErr(true, "RouteCreateOrUpdate")
		return
	}

	glog.V(10).Infof("azRoutesClient.CreateOrUpdate(%q,%q,%q): start", resourceGroupName, routeTableName, routeName)
	defer func() {
		glog.V(10).Infof("azRoutesClient.CreateOrUpdate(%q,%q,%q): end", resourceGroupName, routeTableName, routeName)
	}()

	mc := newMetricContext("routes", "create_or_update", resourceGroupName, az.client.SubscriptionID)
	future, err := az.client.CreateOrUpdate(ctx, resourceGroupName, routeTableName, routeName, routeParameters)
	if err != nil {
		mc.Observe(err)
		return future.Response(), err
	}

	err = future.WaitForCompletion(ctx, az.client.Client)
	mc.Observe(err)
	return future.Response(), err
}

func (az *azRoutesClient) Delete(ctx context.Context, resourceGroupName string, routeTableName string, routeName string) (resp *http.Response, err error) {
	/* Write rate limiting */
	if !az.rateLimiterWriter.TryAccept() {
		err = createARMRateLimitErr(true, "RouteDelete")
		return
	}

	glog.V(10).Infof("azRoutesClient.Delete(%q,%q,%q): start", resourceGroupName, routeTableName, routeName)
	defer func() {
		glog.V(10).Infof("azRoutesClient.Delete(%q,%q,%q): end", resourceGroupName, routeTableName, routeName)
	}()

	mc := newMetricContext("routes", "delete", resourceGroupName, az.client.SubscriptionID)
	future, err := az.client.Delete(ctx, resourceGroupName, routeTableName, routeName)
	if err != nil {
		mc.Observe(err)
		return future.Response(), err
	}

	err = future.WaitForCompletion(ctx, az.client.Client)
	mc.Observe(err)
	return future.Response(), err
}

// azRouteTablesClient implements RouteTablesClient.
type azRouteTablesClient struct {
	client            network.RouteTablesClient
	rateLimiterReader flowcontrol.RateLimiter
	rateLimiterWriter flowcontrol.RateLimiter
}

func newAzRouteTablesClient(config *azClientConfig) *azRouteTablesClient {
	routeTablesClient := network.NewRouteTablesClient(config.subscriptionID)
	routeTablesClient.BaseURI = config.resourceManagerEndpoint
	routeTablesClient.Authorizer = autorest.NewBearerAuthorizer(config.servicePrincipalToken)
	routeTablesClient.PollingDelay = 5 * time.Second
	configureUserAgent(&routeTablesClient.Client)

	return &azRouteTablesClient{
		client:            routeTablesClient,
		rateLimiterReader: config.rateLimiterReader,
		rateLimiterWriter: config.rateLimiterWriter,
	}
}

func (az *azRouteTablesClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, routeTableName string, parameters network.RouteTable) (resp *http.Response, err error) {
	/* Write rate limiting */
	if !az.rateLimiterWriter.TryAccept() {
		err = createARMRateLimitErr(true, "RouteTableCreateOrUpdate")
		return
	}

	glog.V(10).Infof("azRouteTablesClient.CreateOrUpdate(%q,%q): start", resourceGroupName, routeTableName)
	defer func() {
		glog.V(10).Infof("azRouteTablesClient.CreateOrUpdate(%q,%q): end", resourceGroupName, routeTableName)
	}()

	mc := newMetricContext("route_tables", "create_or_update", resourceGroupName, az.client.SubscriptionID)
	future, err := az.client.CreateOrUpdate(ctx, resourceGroupName, routeTableName, parameters)
	mc.Observe(err)
	if err != nil {
		return future.Response(), err
	}

	err = future.WaitForCompletion(ctx, az.client.Client)
	mc.Observe(err)
	return future.Response(), err
}

func (az *azRouteTablesClient) Get(ctx context.Context, resourceGroupName string, routeTableName string, expand string) (result network.RouteTable, err error) {
	if !az.rateLimiterReader.TryAccept() {
		err = createARMRateLimitErr(false, "GetRouteTable")
		return
	}

	glog.V(10).Infof("azRouteTablesClient.Get(%q,%q): start", resourceGroupName, routeTableName)
	defer func() {
		glog.V(10).Infof("azRouteTablesClient.Get(%q,%q): end", resourceGroupName, routeTableName)
	}()

	mc := newMetricContext("route_tables", "get", resourceGroupName, az.client.SubscriptionID)
	result, err = az.client.Get(ctx, resourceGroupName, routeTableName, expand)
	mc.Observe(err)
	return
}

// azStorageAccountClient implements StorageAccountClient.
type azStorageAccountClient struct {
	client            storage.AccountsClient
	rateLimiterReader flowcontrol.RateLimiter
	rateLimiterWriter flowcontrol.RateLimiter
}

func newAzStorageAccountClient(config *azClientConfig) *azStorageAccountClient {
	storageAccountClient := storage.NewAccountsClientWithBaseURI(config.resourceManagerEndpoint, config.subscriptionID)
	storageAccountClient.Authorizer = autorest.NewBearerAuthorizer(config.servicePrincipalToken)
	storageAccountClient.PollingDelay = 5 * time.Second
	configureUserAgent(&storageAccountClient.Client)

	return &azStorageAccountClient{
		client:            storageAccountClient,
		rateLimiterReader: config.rateLimiterReader,
		rateLimiterWriter: config.rateLimiterWriter,
	}
}

func (az *azStorageAccountClient) Create(ctx context.Context, resourceGroupName string, accountName string, parameters storage.AccountCreateParameters) (result *http.Response, err error) {
	/* Write rate limiting */
	if !az.rateLimiterWriter.TryAccept() {
		err = createARMRateLimitErr(true, "StorageAccountCreate")
		return
	}

	glog.V(10).Infof("azStorageAccountClient.Create(%q,%q): start", resourceGroupName, accountName)
	defer func() {
		glog.V(10).Infof("azStorageAccountClient.Create(%q,%q): end", resourceGroupName, accountName)
	}()

	mc := newMetricContext("storage_account", "create", resourceGroupName, az.client.SubscriptionID)
	future, err := az.client.Create(ctx, resourceGroupName, accountName, parameters)
	if err != nil {
		return future.Response(), err
	}

	err = future.WaitForCompletion(ctx, az.client.Client)
	mc.Observe(err)
	return future.Response(), err
}

func (az *azStorageAccountClient) Delete(ctx context.Context, resourceGroupName string, accountName string) (result autorest.Response, err error) {
	if !az.rateLimiterReader.TryAccept() {
		err = createARMRateLimitErr(false, "DeleteStorageAccount")
		return
	}

	glog.V(10).Infof("azStorageAccountClient.Delete(%q,%q): start", resourceGroupName, accountName)
	defer func() {
		glog.V(10).Infof("azStorageAccountClient.Delete(%q,%q): end", resourceGroupName, accountName)
	}()

	mc := newMetricContext("storage_account", "delete", resourceGroupName, az.client.SubscriptionID)
	result, err = az.client.Delete(ctx, resourceGroupName, accountName)
	mc.Observe(err)
	return
}

func (az *azStorageAccountClient) ListKeys(ctx context.Context, resourceGroupName string, accountName string) (result storage.AccountListKeysResult, err error) {
	if !az.rateLimiterReader.TryAccept() {
		err = createARMRateLimitErr(false, "ListStorageAccountKeys")
		return
	}

	glog.V(10).Infof("azStorageAccountClient.ListKeys(%q,%q): start", resourceGroupName, accountName)
	defer func() {
		glog.V(10).Infof("azStorageAccountClient.ListKeys(%q,%q): end", resourceGroupName, accountName)
	}()

	mc := newMetricContext("storage_account", "list_keys", resourceGroupName, az.client.SubscriptionID)
	result, err = az.client.ListKeys(ctx, resourceGroupName, accountName)
	mc.Observe(err)
	return
}

func (az *azStorageAccountClient) ListByResourceGroup(ctx context.Context, resourceGroupName string) (result storage.AccountListResult, err error) {
	if !az.rateLimiterReader.TryAccept() {
		err = createARMRateLimitErr(false, "ListStorageAccountsByResourceGroup")
		return
	}

	glog.V(10).Infof("azStorageAccountClient.ListByResourceGroup(%q): start", resourceGroupName)
	defer func() {
		glog.V(10).Infof("azStorageAccountClient.ListByResourceGroup(%q): end", resourceGroupName)
	}()

	mc := newMetricContext("storage_account", "list_by_resource_group", resourceGroupName, az.client.SubscriptionID)
	result, err = az.client.ListByResourceGroup(ctx, resourceGroupName)
	mc.Observe(err)
	return
}

func (az *azStorageAccountClient) GetProperties(ctx context.Context, resourceGroupName string, accountName string) (result storage.Account, err error) {
	if !az.rateLimiterReader.TryAccept() {
		err = createARMRateLimitErr(false, "GetStorageAccount/Properties")
		return
	}

	glog.V(10).Infof("azStorageAccountClient.GetProperties(%q,%q): start", resourceGroupName, accountName)
	defer func() {
		glog.V(10).Infof("azStorageAccountClient.GetProperties(%q,%q): end", resourceGroupName, accountName)
	}()

	mc := newMetricContext("storage_account", "get_properties", resourceGroupName, az.client.SubscriptionID)
	result, err = az.client.GetProperties(ctx, resourceGroupName, accountName)
	mc.Observe(err)
	return
}

// azDisksClient implements DisksClient.
type azDisksClient struct {
	client            compute.DisksClient
	rateLimiterReader flowcontrol.RateLimiter
	rateLimiterWriter flowcontrol.RateLimiter
}

func newAzDisksClient(config *azClientConfig) *azDisksClient {
	disksClient := compute.NewDisksClientWithBaseURI(config.resourceManagerEndpoint, config.subscriptionID)
	disksClient.Authorizer = autorest.NewBearerAuthorizer(config.servicePrincipalToken)
	disksClient.PollingDelay = 5 * time.Second
	configureUserAgent(&disksClient.Client)

	return &azDisksClient{
		client:            disksClient,
		rateLimiterReader: config.rateLimiterReader,
		rateLimiterWriter: config.rateLimiterWriter,
	}
}

func (az *azDisksClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, diskName string, diskParameter compute.Disk) (resp *http.Response, err error) {
	/* Write rate limiting */
	if !az.rateLimiterWriter.TryAccept() {
		err = createARMRateLimitErr(true, "DiskCreateOrUpdate")
		return
	}

	glog.V(10).Infof("azDisksClient.CreateOrUpdate(%q,%q): start", resourceGroupName, diskName)
	defer func() {
		glog.V(10).Infof("azDisksClient.CreateOrUpdate(%q,%q): end", resourceGroupName, diskName)
	}()

	mc := newMetricContext("disks", "create_or_update", resourceGroupName, az.client.SubscriptionID)
	future, err := az.client.CreateOrUpdate(ctx, resourceGroupName, diskName, diskParameter)
	mc.Observe(err)
	if err != nil {
		return future.Response(), err
	}

	err = future.WaitForCompletion(ctx, az.client.Client)
	mc.Observe(err)
	return future.Response(), err
}

func (az *azDisksClient) Delete(ctx context.Context, resourceGroupName string, diskName string) (resp *http.Response, err error) {
	/* Write rate limiting */
	if !az.rateLimiterWriter.TryAccept() {
		err = createARMRateLimitErr(true, "DiskDelete")
		return
	}

	glog.V(10).Infof("azDisksClient.Delete(%q,%q): start", resourceGroupName, diskName)
	defer func() {
		glog.V(10).Infof("azDisksClient.Delete(%q,%q): end", resourceGroupName, diskName)
	}()

	mc := newMetricContext("disks", "delete", resourceGroupName, az.client.SubscriptionID)
	future, err := az.client.Delete(ctx, resourceGroupName, diskName)
	mc.Observe(err)
	if err != nil {
		return future.Response(), err
	}

	err = future.WaitForCompletion(ctx, az.client.Client)
	mc.Observe(err)
	return future.Response(), err
}

func (az *azDisksClient) Get(ctx context.Context, resourceGroupName string, diskName string) (result compute.Disk, err error) {
	if !az.rateLimiterReader.TryAccept() {
		err = createARMRateLimitErr(false, "GetDisk")
		return
	}

	glog.V(10).Infof("azDisksClient.Get(%q,%q): start", resourceGroupName, diskName)
	defer func() {
		glog.V(10).Infof("azDisksClient.Get(%q,%q): end", resourceGroupName, diskName)
	}()

	mc := newMetricContext("disks", "get", resourceGroupName, az.client.SubscriptionID)
	result, err = az.client.Get(ctx, resourceGroupName, diskName)
	mc.Observe(err)
	return
}
