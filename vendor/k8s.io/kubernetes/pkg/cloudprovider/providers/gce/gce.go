/*
Copyright 2014 The Kubernetes Authors.

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
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	gcfg "gopkg.in/gcfg.v1"

	"cloud.google.com/go/compute/metadata"
	"github.com/golang/glog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	computealpha "google.golang.org/api/compute/v0.alpha"
	computebeta "google.golang.org/api/compute/v0.beta"
	compute "google.golang.org/api/compute/v1"
	container "google.golang.org/api/container/v1"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/flowcontrol"

	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/cloudprovider/providers/gce/cloud"
	"k8s.io/kubernetes/pkg/controller"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
	"k8s.io/kubernetes/pkg/version"
)

const (
	ProviderName = "gce"

	k8sNodeRouteTag = "k8s-node-route"

	// AffinityTypeNone - no session affinity.
	gceAffinityTypeNone = "NONE"
	// AffinityTypeClientIP - affinity based on Client IP.
	gceAffinityTypeClientIP = "CLIENT_IP"
	// AffinityTypeClientIPProto - affinity based on Client IP and port.
	gceAffinityTypeClientIPProto = "CLIENT_IP_PROTO"

	operationPollInterval = 3 * time.Second
	// Creating Route in very large clusters, may take more than half an hour.
	operationPollTimeoutDuration = time.Hour

	// Each page can have 500 results, but we cap how many pages
	// are iterated through to prevent infinite loops if the API
	// were to continuously return a nextPageToken.
	maxPages = 25

	maxTargetPoolCreateInstances = 200

	// HTTP Load Balancer parameters
	// Configure 2 second period for external health checks.
	gceHcCheckIntervalSeconds = int64(2)
	gceHcTimeoutSeconds       = int64(1)
	// Start sending requests as soon as a pod is found on the node.
	gceHcHealthyThreshold = int64(1)
	// Defaults to 5 * 2 = 10 seconds before the LB will steer traffic away
	gceHcUnhealthyThreshold = int64(5)

	gceComputeAPIEndpoint     = "https://www.googleapis.com/compute/v1/"
	gceComputeAPIEndpointBeta = "https://www.googleapis.com/compute/beta/"
)

// gceObject is an abstraction of all GCE API object in go client
type gceObject interface {
	MarshalJSON() ([]byte, error)
}

// GCECloud is an implementation of Interface, LoadBalancer and Instances for Google Compute Engine.
type GCECloud struct {
	// ClusterID contains functionality for getting (and initializing) the ingress-uid. Call GCECloud.Initialize()
	// for the cloudprovider to start watching the configmap.
	ClusterID ClusterID

	service          *compute.Service
	serviceBeta      *computebeta.Service
	serviceAlpha     *computealpha.Service
	containerService *container.Service
	tpuService       *tpuService
	client           clientset.Interface
	clientBuilder    controller.ControllerClientBuilder
	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder
	projectID        string
	region           string
	localZone        string // The zone in which we are running
	// managedZones will be set to the 1 zone if running a single zone cluster
	// it will be set to ALL zones in region for any multi-zone cluster
	// Use GetAllCurrentZones to get only zones that contain nodes
	managedZones             []string
	networkURL               string
	isLegacyNetwork          bool
	subnetworkURL            string
	secondaryRangeName       string
	networkProjectID         string
	onXPN                    bool
	nodeTags                 []string    // List of tags to use on firewall rules for load balancers
	lastComputedNodeTags     []string    // List of node tags calculated in GetHostTags()
	lastKnownNodeNames       sets.String // List of hostnames used to calculate lastComputedHostTags in GetHostTags(names)
	computeNodeTagLock       sync.Mutex  // Lock for computing and setting node tags
	nodeInstancePrefix       string      // If non-"", an advisory prefix for all nodes in the cluster
	useMetadataServer        bool
	operationPollRateLimiter flowcontrol.RateLimiter
	manager                  diskServiceManager
	// Lock for access to nodeZones
	nodeZonesLock sync.Mutex
	// nodeZones is a mapping from Zone to a sets.String of Node's names in the Zone
	// it is updated by the nodeInformer
	nodeZones          map[string]sets.String
	nodeInformerSynced cache.InformerSynced
	// sharedResourceLock is used to serialize GCE operations that may mutate shared state to
	// prevent inconsistencies. For example, load balancers manipulation methods will take the
	// lock to prevent shared resources from being prematurely deleted while the operation is
	// in progress.
	sharedResourceLock sync.Mutex
	// AlphaFeatureGate gates gce alpha features in GCECloud instance.
	// Related wrapper functions that interacts with gce alpha api should examine whether
	// the corresponding api is enabled.
	// If not enabled, it should return error.
	AlphaFeatureGate *AlphaFeatureGate

	// New code generated interface to the GCE compute library.
	c cloud.Cloud

	// Keep a reference of this around so we can inject a new cloud.RateLimiter implementation.
	s *cloud.Service
}

// TODO: replace gcfg with json
type ConfigGlobal struct {
	TokenURL  string `gcfg:"token-url"`
	TokenBody string `gcfg:"token-body"`
	// ProjectID and NetworkProjectID can either be the numeric or string-based
	// unique identifier that starts with [a-z].
	ProjectID string `gcfg:"project-id"`
	// NetworkProjectID refers to the project which owns the network being used.
	NetworkProjectID string `gcfg:"network-project-id"`
	NetworkName      string `gcfg:"network-name"`
	SubnetworkName   string `gcfg:"subnetwork-name"`
	// SecondaryRangeName is the name of the secondary range to allocate IP
	// aliases. The secondary range must be present on the subnetwork the
	// cluster is attached to.
	SecondaryRangeName string   `gcfg:"secondary-range-name"`
	NodeTags           []string `gcfg:"node-tags"`
	NodeInstancePrefix string   `gcfg:"node-instance-prefix"`
	Multizone          bool     `gcfg:"multizone"`
	// ApiEndpoint is the GCE compute API endpoint to use. If this is blank,
	// then the default endpoint is used.
	ApiEndpoint string `gcfg:"api-endpoint"`
	// LocalZone specifies the GCE zone that gce cloud client instance is
	// located in (i.e. where the controller will be running). If this is
	// blank, then the local zone will be discovered via the metadata server.
	LocalZone string `gcfg:"local-zone"`
	// Default to none.
	// For example: MyFeatureFlag
	AlphaFeatures []string `gcfg:"alpha-features"`
}

// ConfigFile is the struct used to parse the /etc/gce.conf configuration file.
type ConfigFile struct {
	Global ConfigGlobal `gcfg:"global"`
}

// CloudConfig includes all the necessary configuration for creating GCECloud
type CloudConfig struct {
	ApiEndpoint        string
	ProjectID          string
	NetworkProjectID   string
	Region             string
	Zone               string
	ManagedZones       []string
	NetworkName        string
	NetworkURL         string
	SubnetworkName     string
	SubnetworkURL      string
	SecondaryRangeName string
	NodeTags           []string
	NodeInstancePrefix string
	TokenSource        oauth2.TokenSource
	UseMetadataServer  bool
	AlphaFeatureGate   *AlphaFeatureGate
}

func init() {
	cloudprovider.RegisterCloudProvider(
		ProviderName,
		func(config io.Reader) (cloudprovider.Interface, error) {
			return newGCECloud(config)
		})
}

// Services is the set of all versions of the compute service.
type Services struct {
	// GA, Alpha, Beta versions of the compute API.
	GA    *compute.Service
	Alpha *computealpha.Service
	Beta  *computebeta.Service
}

// ComputeServices returns access to the internal compute services.
func (g *GCECloud) ComputeServices() *Services {
	return &Services{g.service, g.serviceAlpha, g.serviceBeta}
}

// Compute returns the generated stubs for the compute API.
func (g *GCECloud) Compute() cloud.Cloud {
	return g.c
}

// newGCECloud creates a new instance of GCECloud.
func newGCECloud(config io.Reader) (gceCloud *GCECloud, err error) {
	var cloudConfig *CloudConfig
	var configFile *ConfigFile

	if config != nil {
		configFile, err = readConfig(config)
		if err != nil {
			return nil, err
		}
		glog.Infof("Using GCE provider config %+v", configFile)
	}

	cloudConfig, err = generateCloudConfig(configFile)
	if err != nil {
		return nil, err
	}
	return CreateGCECloud(cloudConfig)
}

func readConfig(reader io.Reader) (*ConfigFile, error) {
	cfg := &ConfigFile{}
	if err := gcfg.FatalOnly(gcfg.ReadInto(cfg, reader)); err != nil {
		glog.Errorf("Couldn't read config: %v", err)
		return nil, err
	}
	return cfg, nil
}

func generateCloudConfig(configFile *ConfigFile) (cloudConfig *CloudConfig, err error) {
	cloudConfig = &CloudConfig{}
	// By default, fetch token from GCE metadata server
	cloudConfig.TokenSource = google.ComputeTokenSource("")
	cloudConfig.UseMetadataServer = true
	cloudConfig.AlphaFeatureGate = NewAlphaFeatureGate([]string{})
	if configFile != nil {
		if configFile.Global.ApiEndpoint != "" {
			cloudConfig.ApiEndpoint = configFile.Global.ApiEndpoint
		}

		if configFile.Global.TokenURL != "" {
			// if tokenURL is nil, set tokenSource to nil. This will force the OAuth client to fall
			// back to use DefaultTokenSource. This allows running gceCloud remotely.
			if configFile.Global.TokenURL == "nil" {
				cloudConfig.TokenSource = nil
			} else {
				cloudConfig.TokenSource = NewAltTokenSource(configFile.Global.TokenURL, configFile.Global.TokenBody)
			}
		}

		cloudConfig.NodeTags = configFile.Global.NodeTags
		cloudConfig.NodeInstancePrefix = configFile.Global.NodeInstancePrefix
		cloudConfig.AlphaFeatureGate = NewAlphaFeatureGate(configFile.Global.AlphaFeatures)
	}

	// retrieve projectID and zone
	if configFile == nil || configFile.Global.ProjectID == "" || configFile.Global.LocalZone == "" {
		cloudConfig.ProjectID, cloudConfig.Zone, err = getProjectAndZone()
		if err != nil {
			return nil, err
		}
	}

	if configFile != nil {
		if configFile.Global.ProjectID != "" {
			cloudConfig.ProjectID = configFile.Global.ProjectID
		}
		if configFile.Global.LocalZone != "" {
			cloudConfig.Zone = configFile.Global.LocalZone
		}
		if configFile.Global.NetworkProjectID != "" {
			cloudConfig.NetworkProjectID = configFile.Global.NetworkProjectID
		}
	}

	// retrieve region
	cloudConfig.Region, err = GetGCERegion(cloudConfig.Zone)
	if err != nil {
		return nil, err
	}

	// generate managedZones
	cloudConfig.ManagedZones = []string{cloudConfig.Zone}
	if configFile != nil && configFile.Global.Multizone {
		cloudConfig.ManagedZones = nil // Use all zones in region
	}

	// Determine if network parameter is URL or Name
	if configFile != nil && configFile.Global.NetworkName != "" {
		if strings.Contains(configFile.Global.NetworkName, "/") {
			cloudConfig.NetworkURL = configFile.Global.NetworkName
		} else {
			cloudConfig.NetworkName = configFile.Global.NetworkName
		}
	} else {
		cloudConfig.NetworkName, err = getNetworkNameViaMetadata()
		if err != nil {
			return nil, err
		}
	}

	// Determine if subnetwork parameter is URL or Name
	// If cluster is on a GCP network of mode=custom, then `SubnetName` must be specified in config file.
	if configFile != nil && configFile.Global.SubnetworkName != "" {
		if strings.Contains(configFile.Global.SubnetworkName, "/") {
			cloudConfig.SubnetworkURL = configFile.Global.SubnetworkName
		} else {
			cloudConfig.SubnetworkName = configFile.Global.SubnetworkName
		}
	}

	if configFile != nil {
		cloudConfig.SecondaryRangeName = configFile.Global.SecondaryRangeName
	}

	return cloudConfig, err
}

// CreateGCECloud creates a GCECloud object using the specified parameters.
// If no networkUrl is specified, loads networkName via rest call.
// If no tokenSource is specified, uses oauth2.DefaultTokenSource.
// If managedZones is nil / empty all zones in the region will be managed.
func CreateGCECloud(config *CloudConfig) (*GCECloud, error) {
	// Remove any pre-release version and build metadata from the semver,
	// leaving only the MAJOR.MINOR.PATCH portion. See http://semver.org/.
	version := strings.TrimLeft(strings.Split(strings.Split(version.Get().GitVersion, "-")[0], "+")[0], "v")

	// Create a user-agent header append string to supply to the Google API
	// clients, to identify Kubernetes as the origin of the GCP API calls.
	userAgent := fmt.Sprintf("Kubernetes/%s (%s %s)", version, runtime.GOOS, runtime.GOARCH)

	// Use ProjectID for NetworkProjectID, if it wasn't explicitly set.
	if config.NetworkProjectID == "" {
		config.NetworkProjectID = config.ProjectID
	}

	client, err := newOauthClient(config.TokenSource)
	if err != nil {
		return nil, err
	}
	service, err := compute.New(client)
	if err != nil {
		return nil, err
	}
	service.UserAgent = userAgent

	client, err = newOauthClient(config.TokenSource)
	if err != nil {
		return nil, err
	}
	serviceBeta, err := computebeta.New(client)
	if err != nil {
		return nil, err
	}
	serviceBeta.UserAgent = userAgent

	client, err = newOauthClient(config.TokenSource)
	if err != nil {
		return nil, err
	}
	serviceAlpha, err := computealpha.New(client)
	if err != nil {
		return nil, err
	}
	serviceAlpha.UserAgent = userAgent

	// Expect override api endpoint to always be v1 api and follows the same pattern as prod.
	// Generate alpha and beta api endpoints based on override v1 api endpoint.
	// For example,
	// staging API endpoint: https://www.googleapis.com/compute/staging_v1/
	if config.ApiEndpoint != "" {
		service.BasePath = fmt.Sprintf("%sprojects/", config.ApiEndpoint)
		serviceBeta.BasePath = fmt.Sprintf("%sprojects/", strings.Replace(config.ApiEndpoint, "v1", "beta", -1))
		serviceAlpha.BasePath = fmt.Sprintf("%sprojects/", strings.Replace(config.ApiEndpoint, "v1", "alpha", -1))
	}

	containerService, err := container.New(client)
	if err != nil {
		return nil, err
	}
	containerService.UserAgent = userAgent

	tpuService, err := newTPUService(client)
	if err != nil {
		return nil, err
	}

	// ProjectID and.NetworkProjectID may be project number or name.
	projID, netProjID := tryConvertToProjectNames(config.ProjectID, config.NetworkProjectID, service)
	onXPN := projID != netProjID

	var networkURL string
	var subnetURL string
	var isLegacyNetwork bool

	if config.NetworkURL != "" {
		networkURL = config.NetworkURL
	} else if config.NetworkName != "" {
		networkURL = gceNetworkURL(config.ApiEndpoint, netProjID, config.NetworkName)
	} else {
		// Other consumers may use the cloudprovider without utilizing the wrapped GCE API functions
		// or functions requiring network/subnetwork URLs (e.g. Kubelet).
		glog.Warningf("No network name or URL specified.")
	}

	if config.SubnetworkURL != "" {
		subnetURL = config.SubnetworkURL
	} else if config.SubnetworkName != "" {
		subnetURL = gceSubnetworkURL(config.ApiEndpoint, netProjID, config.Region, config.SubnetworkName)
	} else {
		// Determine the type of network and attempt to discover the correct subnet for AUTO mode.
		// Gracefully fail because kubelet calls CreateGCECloud without any config, and minions
		// lack the proper credentials for API calls.
		if networkName := lastComponent(networkURL); networkName != "" {
			var n *compute.Network
			if n, err = getNetwork(service, netProjID, networkName); err != nil {
				glog.Warningf("Could not retrieve network %q; err: %v", networkName, err)
			} else {
				switch typeOfNetwork(n) {
				case netTypeLegacy:
					glog.Infof("Network %q is type legacy - no subnetwork", networkName)
					isLegacyNetwork = true
				case netTypeCustom:
					glog.Warningf("Network %q is type custom - cannot auto select a subnetwork", networkName)
				case netTypeAuto:
					subnetURL, err = determineSubnetURL(service, netProjID, networkName, config.Region)
					if err != nil {
						glog.Warningf("Could not determine subnetwork for network %q and region %v; err: %v", networkName, config.Region, err)
					} else {
						glog.Infof("Auto selecting subnetwork %q", subnetURL)
					}
				}
			}
		}
	}

	if len(config.ManagedZones) == 0 {
		config.ManagedZones, err = getZonesForRegion(service, config.ProjectID, config.Region)
		if err != nil {
			return nil, err
		}
	}
	if len(config.ManagedZones) > 1 {
		glog.Infof("managing multiple zones: %v", config.ManagedZones)
	}

	operationPollRateLimiter := flowcontrol.NewTokenBucketRateLimiter(10, 100) // 10 qps, 100 bucket size.

	gce := &GCECloud{
		service:                  service,
		serviceAlpha:             serviceAlpha,
		serviceBeta:              serviceBeta,
		containerService:         containerService,
		tpuService:               tpuService,
		projectID:                projID,
		networkProjectID:         netProjID,
		onXPN:                    onXPN,
		region:                   config.Region,
		localZone:                config.Zone,
		managedZones:             config.ManagedZones,
		networkURL:               networkURL,
		isLegacyNetwork:          isLegacyNetwork,
		subnetworkURL:            subnetURL,
		secondaryRangeName:       config.SecondaryRangeName,
		nodeTags:                 config.NodeTags,
		nodeInstancePrefix:       config.NodeInstancePrefix,
		useMetadataServer:        config.UseMetadataServer,
		operationPollRateLimiter: operationPollRateLimiter,
		AlphaFeatureGate:         config.AlphaFeatureGate,
		nodeZones:                map[string]sets.String{},
	}

	gce.manager = &gceServiceManager{gce}
	gce.s = &cloud.Service{
		GA:            service,
		Alpha:         serviceAlpha,
		Beta:          serviceBeta,
		ProjectRouter: &gceProjectRouter{gce},
		RateLimiter:   &gceRateLimiter{gce},
	}
	gce.c = cloud.NewGCE(gce.s)

	return gce, nil
}

// SetRateLimiter adds a custom cloud.RateLimiter implementation.
// WARNING: Calling this could have unexpected behavior if you have in-flight
// requests. It is best to use this immediately after creating a GCECloud.
func (g *GCECloud) SetRateLimiter(rl cloud.RateLimiter) {
	if rl != nil {
		g.s.RateLimiter = rl
	}
}

// determineSubnetURL queries for all subnetworks in a region for a given network and returns
// the URL of the subnetwork which exists in the auto-subnet range.
func determineSubnetURL(service *compute.Service, networkProjectID, networkName, region string) (string, error) {
	subnets, err := listSubnetworksOfNetwork(service, networkProjectID, networkName, region)
	if err != nil {
		return "", err
	}

	autoSubnets, err := subnetsInCIDR(subnets, autoSubnetIPRange)
	if err != nil {
		return "", err
	}

	if len(autoSubnets) == 0 {
		return "", fmt.Errorf("no subnet exists in auto CIDR")
	}

	if len(autoSubnets) > 1 {
		return "", fmt.Errorf("multiple subnetworks in the same region exist in auto CIDR")
	}

	return autoSubnets[0].SelfLink, nil
}

func tryConvertToProjectNames(configProject, configNetworkProject string, service *compute.Service) (projID, netProjID string) {
	projID = configProject
	if isProjectNumber(projID) {
		projName, err := getProjectID(service, projID)
		if err != nil {
			glog.Warningf("Failed to retrieve project %v while trying to retrieve its name. err %v", projID, err)
		} else {
			projID = projName
		}
	}

	netProjID = projID
	if configNetworkProject != configProject {
		netProjID = configNetworkProject
	}
	if isProjectNumber(netProjID) {
		netProjName, err := getProjectID(service, netProjID)
		if err != nil {
			glog.Warningf("Failed to retrieve network project %v while trying to retrieve its name. err %v", netProjID, err)
		} else {
			netProjID = netProjName
		}
	}

	return projID, netProjID
}

// Initialize takes in a clientBuilder and spawns a goroutine for watching the clusterid configmap.
// This must be called before utilizing the funcs of gce.ClusterID
func (gce *GCECloud) Initialize(clientBuilder controller.ControllerClientBuilder) {
	gce.clientBuilder = clientBuilder
	gce.client = clientBuilder.ClientOrDie("cloud-provider")

	if gce.OnXPN() {
		gce.eventBroadcaster = record.NewBroadcaster()
		gce.eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: gce.client.CoreV1().Events("")})
		gce.eventRecorder = gce.eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "gce-cloudprovider"})
	}

	go gce.watchClusterID()
}

// LoadBalancer returns an implementation of LoadBalancer for Google Compute Engine.
func (gce *GCECloud) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	return gce, true
}

// Instances returns an implementation of Instances for Google Compute Engine.
func (gce *GCECloud) Instances() (cloudprovider.Instances, bool) {
	return gce, true
}

// Zones returns an implementation of Zones for Google Compute Engine.
func (gce *GCECloud) Zones() (cloudprovider.Zones, bool) {
	return gce, true
}

func (gce *GCECloud) Clusters() (cloudprovider.Clusters, bool) {
	return gce, true
}

// Routes returns an implementation of Routes for Google Compute Engine.
func (gce *GCECloud) Routes() (cloudprovider.Routes, bool) {
	return gce, true
}

// ProviderName returns the cloud provider ID.
func (gce *GCECloud) ProviderName() string {
	return ProviderName
}

// ProjectID returns the ProjectID corresponding to the project this cloud is in.
func (g *GCECloud) ProjectID() string {
	return g.projectID
}

// NetworkProjectID returns the ProjectID corresponding to the project this cluster's network is in.
func (g *GCECloud) NetworkProjectID() string {
	return g.networkProjectID
}

// Region returns the region
func (gce *GCECloud) Region() string {
	return gce.region
}

// OnXPN returns true if the cluster is running on a cross project network (XPN)
func (gce *GCECloud) OnXPN() bool {
	return gce.onXPN
}

// NetworkURL returns the network url
func (gce *GCECloud) NetworkURL() string {
	return gce.networkURL
}

// SubnetworkURL returns the subnetwork url
func (gce *GCECloud) SubnetworkURL() string {
	return gce.subnetworkURL
}

func (gce *GCECloud) IsLegacyNetwork() bool {
	return gce.isLegacyNetwork
}

func (gce *GCECloud) SetInformers(informerFactory informers.SharedInformerFactory) {
	glog.Infof("Setting up informers for GCECloud")
	nodeInformer := informerFactory.Core().V1().Nodes().Informer()
	nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node := obj.(*v1.Node)
			gce.updateNodeZones(nil, node)
		},
		UpdateFunc: func(prev, obj interface{}) {
			prevNode := prev.(*v1.Node)
			newNode := obj.(*v1.Node)
			if newNode.Labels[kubeletapis.LabelZoneFailureDomain] ==
				prevNode.Labels[kubeletapis.LabelZoneFailureDomain] {
				return
			}
			gce.updateNodeZones(prevNode, newNode)
		},
		DeleteFunc: func(obj interface{}) {
			node, isNode := obj.(*v1.Node)
			// We can get DeletedFinalStateUnknown instead of *v1.Node here
			// and we need to handle that correctly.
			if !isNode {
				deletedState, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					glog.Errorf("Received unexpected object: %v", obj)
					return
				}
				node, ok = deletedState.Obj.(*v1.Node)
				if !ok {
					glog.Errorf("DeletedFinalStateUnknown contained non-Node object: %v", deletedState.Obj)
					return
				}
			}
			gce.updateNodeZones(node, nil)
		},
	})
	gce.nodeInformerSynced = nodeInformer.HasSynced
}

func (gce *GCECloud) updateNodeZones(prevNode, newNode *v1.Node) {
	gce.nodeZonesLock.Lock()
	defer gce.nodeZonesLock.Unlock()
	if prevNode != nil {
		prevZone, ok := prevNode.ObjectMeta.Labels[kubeletapis.LabelZoneFailureDomain]
		if ok {
			gce.nodeZones[prevZone].Delete(prevNode.ObjectMeta.Name)
			if gce.nodeZones[prevZone].Len() == 0 {
				gce.nodeZones[prevZone] = nil
			}
		}
	}
	if newNode != nil {
		newZone, ok := newNode.ObjectMeta.Labels[kubeletapis.LabelZoneFailureDomain]
		if ok {
			if gce.nodeZones[newZone] == nil {
				gce.nodeZones[newZone] = sets.NewString()
			}
			gce.nodeZones[newZone].Insert(newNode.ObjectMeta.Name)
		}
	}
}

// HasClusterID returns true if the cluster has a clusterID
func (gce *GCECloud) HasClusterID() bool {
	return true
}

// Project IDs cannot have a digit for the first characeter. If the id contains a digit,
// then it must be a project number.
func isProjectNumber(idOrNumber string) bool {
	_, err := strconv.ParseUint(idOrNumber, 10, 64)
	return err == nil
}

// GCECloud implements cloudprovider.Interface.
var _ cloudprovider.Interface = (*GCECloud)(nil)

func gceNetworkURL(apiEndpoint, project, network string) string {
	if apiEndpoint == "" {
		apiEndpoint = gceComputeAPIEndpoint
	}
	return apiEndpoint + strings.Join([]string{"projects", project, "global", "networks", network}, "/")
}

func gceSubnetworkURL(apiEndpoint, project, region, subnetwork string) string {
	if apiEndpoint == "" {
		apiEndpoint = gceComputeAPIEndpoint
	}
	return apiEndpoint + strings.Join([]string{"projects", project, "regions", region, "subnetworks", subnetwork}, "/")
}

// getRegionInURL parses full resource URLS and shorter URLS
// https://www.googleapis.com/compute/v1/projects/myproject/regions/us-central1/subnetworks/a
// projects/myproject/regions/us-central1/subnetworks/a
// All return "us-central1"
func getRegionInURL(urlStr string) string {
	fields := strings.Split(urlStr, "/")
	for i, v := range fields {
		if v == "regions" && i < len(fields)-1 {
			return fields[i+1]
		}
	}
	return ""
}

func getNetworkNameViaMetadata() (string, error) {
	result, err := metadata.Get("instance/network-interfaces/0/network")
	if err != nil {
		return "", err
	}
	parts := strings.Split(result, "/")
	if len(parts) != 4 {
		return "", fmt.Errorf("unexpected response: %s", result)
	}
	return parts[3], nil
}

// getNetwork returns a GCP network
func getNetwork(svc *compute.Service, networkProjectID, networkID string) (*compute.Network, error) {
	return svc.Networks.Get(networkProjectID, networkID).Do()
}

// listSubnetworksOfNetwork returns a list of subnetworks for a particular region of a network.
func listSubnetworksOfNetwork(svc *compute.Service, networkProjectID, networkID, region string) ([]*compute.Subnetwork, error) {
	var subnets []*compute.Subnetwork
	err := svc.Subnetworks.List(networkProjectID, region).Filter(fmt.Sprintf("network eq .*/%v$", networkID)).Pages(context.Background(), func(res *compute.SubnetworkList) error {
		subnets = append(subnets, res.Items...)
		return nil
	})
	return subnets, err
}

// getProjectID returns the project's string ID given a project number or string
func getProjectID(svc *compute.Service, projectNumberOrID string) (string, error) {
	proj, err := svc.Projects.Get(projectNumberOrID).Do()
	if err != nil {
		return "", err
	}

	return proj.Name, nil
}

func getZonesForRegion(svc *compute.Service, projectID, region string) ([]string, error) {
	// TODO: use PageToken to list all not just the first 500
	listCall := svc.Zones.List(projectID)

	// Filtering by region doesn't seem to work
	// (tested in https://cloud.google.com/compute/docs/reference/latest/zones/list)
	// listCall = listCall.Filter("region eq " + region)

	res, err := listCall.Do()
	if err != nil {
		return nil, fmt.Errorf("unexpected response listing zones: %v", err)
	}
	zones := []string{}
	for _, zone := range res.Items {
		regionName := lastComponent(zone.Region)
		if regionName == region {
			zones = append(zones, zone.Name)
		}
	}
	return zones, nil
}

func findSubnetForRegion(subnetURLs []string, region string) string {
	for _, url := range subnetURLs {
		if thisRegion := getRegionInURL(url); thisRegion == region {
			return url
		}
	}
	return ""
}

func newOauthClient(tokenSource oauth2.TokenSource) (*http.Client, error) {
	if tokenSource == nil {
		var err error
		tokenSource, err = google.DefaultTokenSource(
			oauth2.NoContext,
			compute.CloudPlatformScope,
			compute.ComputeScope)
		glog.Infof("Using DefaultTokenSource %#v", tokenSource)
		if err != nil {
			return nil, err
		}
	} else {
		glog.Infof("Using existing Token Source %#v", tokenSource)
	}

	backoff := wait.Backoff{
		// These values will add up to about a minute. See #56293 for background.
		Duration: time.Second,
		Factor:   1.4,
		Steps:    10,
	}
	if err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		if _, err := tokenSource.Token(); err != nil {
			glog.Errorf("error fetching initial token: %v", err)
			return false, nil
		}
		return true, nil
	}); err != nil {
		return nil, err
	}

	return oauth2.NewClient(oauth2.NoContext, tokenSource), nil
}

func (manager *gceServiceManager) getProjectsAPIEndpoint() string {
	projectsApiEndpoint := gceComputeAPIEndpoint + "projects/"
	if manager.gce.service != nil {
		projectsApiEndpoint = manager.gce.service.BasePath
	}

	return projectsApiEndpoint
}

func (manager *gceServiceManager) getProjectsAPIEndpointBeta() string {
	projectsApiEndpoint := gceComputeAPIEndpointBeta + "projects/"
	if manager.gce.service != nil {
		projectsApiEndpoint = manager.gce.serviceBeta.BasePath
	}

	return projectsApiEndpoint
}
