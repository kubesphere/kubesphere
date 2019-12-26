// Copyright (c) 2017-2019 Tigera, Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package clientv3

import (
	"context"

	log "github.com/sirupsen/logrus"

	"encoding/hex"
	"fmt"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/projectcalico/libcalico-go/lib/apiconfig"
	v3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/libcalico-go/lib/backend"
	bapi "github.com/projectcalico/libcalico-go/lib/backend/api"
	cerrors "github.com/projectcalico/libcalico-go/lib/errors"
	"github.com/projectcalico/libcalico-go/lib/ipam"
	"github.com/projectcalico/libcalico-go/lib/net"
	"github.com/projectcalico/libcalico-go/lib/options"
	"github.com/projectcalico/libcalico-go/lib/set"
)

// client implements the client.Interface.
type client struct {
	// The config we were created with.
	config apiconfig.CalicoAPIConfig

	// The backend client.
	backend bapi.Client

	// The resources client used internally.
	resources resourceInterface
}

// New returns a connected client. The ClientConfig can either be created explicitly,
// or can be loaded from a config file or environment variables using the LoadClientConfig() function.
func New(config apiconfig.CalicoAPIConfig) (Interface, error) {
	be, err := backend.NewClient(config)
	if err != nil {
		return nil, err
	}
	return client{
		config:    config,
		backend:   be,
		resources: &resources{backend: be},
	}, nil
}

// NewFromEnv loads the config from ENV variables and returns a connected client.
func NewFromEnv() (Interface, error) {

	config, err := apiconfig.LoadClientConfigFromEnvironment()
	if err != nil {
		return nil, err
	}

	return New(*config)
}

// Nodes returns an interface for managing node resources.
func (c client) Nodes() NodeInterface {
	return nodes{client: c}
}

// NetworkPolicies returns an interface for managing policy resources.
func (c client) NetworkPolicies() NetworkPolicyInterface {
	return networkPolicies{client: c}
}

// GlobalNetworkPolicies returns an interface for managing policy resources.
func (c client) GlobalNetworkPolicies() GlobalNetworkPolicyInterface {
	return globalNetworkPolicies{client: c}
}

// IPPools returns an interface for managing IP pool resources.
func (c client) IPPools() IPPoolInterface {
	return ipPools{client: c}
}

// Profiles returns an interface for managing profile resources.
func (c client) Profiles() ProfileInterface {
	return profiles{client: c}
}

// GlobalNetworkSets returns an interface for managing host endpoint resources.
func (c client) GlobalNetworkSets() GlobalNetworkSetInterface {
	return globalNetworkSets{client: c}
}

// NetworkSets returns an interface for managing host endpoint resources.
func (c client) NetworkSets() NetworkSetInterface {
	return networkSets{client: c}
}

// HostEndpoints returns an interface for managing host endpoint resources.
func (c client) HostEndpoints() HostEndpointInterface {
	return hostEndpoints{client: c}
}

// WorkloadEndpoints returns an interface for managing workload endpoint resources.
func (c client) WorkloadEndpoints() WorkloadEndpointInterface {
	return workloadEndpoints{client: c}
}

// BGPPeers returns an interface for managing BGP peer resources.
func (c client) BGPPeers() BGPPeerInterface {
	return bgpPeers{client: c}
}

// IPAM returns an interface for managing IP address assignment and releasing.
func (c client) IPAM() ipam.Interface {
	return ipam.NewIPAMClient(c.backend, poolAccessor{client: &c})
}

// BGPConfigurations returns an interface for managing the BGP configuration resources.
func (c client) BGPConfigurations() BGPConfigurationInterface {
	return bgpConfigurations{client: c}
}

// FelixConfigurations returns an interface for managing the Felix configuration resources.
func (c client) FelixConfigurations() FelixConfigurationInterface {
	return felixConfigurations{client: c}
}

// ClusterInformation returns an interface for managing the cluster information resource.
func (c client) ClusterInformation() ClusterInformationInterface {
	return clusterInformation{client: c}
}

type poolAccessor struct {
	client *client
}

func (p poolAccessor) GetEnabledPools(ipVersion int) ([]v3.IPPool, error) {
	return p.getPools(func(pool *v3.IPPool) bool {
		if pool.Spec.Disabled {
			log.Debugf("Skipping disabled IP pool (%s)", pool.Name)
			return false
		}
		if _, cidr, err := net.ParseCIDR(pool.Spec.CIDR); err == nil && cidr.Version() == ipVersion {
			log.Debugf("Adding pool (%s) to the IPPool list", cidr.String())
			return true
		} else if err != nil {
			log.Warnf("Failed to parse the IPPool: %s. Ignoring that IPPool", pool.Spec.CIDR)
		} else {
			log.Debugf("Ignoring IPPool: %s. IP version is different.", pool.Spec.CIDR)
		}
		return false
	})
}

func (p poolAccessor) getPools(filter func(pool *v3.IPPool) bool) ([]v3.IPPool, error) {
	pools, err := p.client.IPPools().List(context.Background(), options.ListOptions{})
	if err != nil {
		return nil, err
	}
	log.Debugf("Got list of all IPPools: %v", pools)
	var filtered []v3.IPPool
	for _, pool := range pools.Items {
		if filter(&pool) {
			filtered = append(filtered, pool)
		}
	}
	return filtered, nil
}

func (p poolAccessor) GetAllPools() ([]v3.IPPool, error) {
	return p.getPools(func(pool *v3.IPPool) bool {
		return true
	})
}

// EnsureInitialized is used to ensure the backend datastore is correctly
// initialized for use by Calico.  This method may be called multiple times, and
// will have no effect if the datastore is already correctly initialized.
//
// Most Calico deployment scenarios will automatically implicitly invoke this
// method and so a general consumer of this API can assume that the datastore
// is already initialized.
func (c client) EnsureInitialized(ctx context.Context, calicoVersion, clusterType string) error {
	// Perform datastore specific initialization first.
	if err := c.backend.EnsureInitialized(); err != nil {
		return err
	}

	if err := c.ensureClusterInformation(ctx, calicoVersion, clusterType); err != nil {
		return err
	}

	return nil
}

const globalClusterInfoName = "default"

// ensureClusterInformation ensures that the ClusterInformation fields i.e. ClusterType,
// CalicoVersion and ClusterGUID are set.  It creates/updates the ClusterInformation as needed.
func (c client) ensureClusterInformation(ctx context.Context, calicoVersion, clusterType string) error {
	// Append "kdd" last if the datastoreType is 'kubernetes'.
	if c.config.Spec.DatastoreType == apiconfig.Kubernetes {
		// If clusterType is already set then append ",kdd" at the end.
		if clusterType != "" {
			// Trim the trailing ",", if any.
			clusterType = strings.TrimSuffix(clusterType, ",")
			// Append "kdd" very last thing in the list.
			clusterType = fmt.Sprintf("%s,%s", clusterType, "kdd")
		} else {
			clusterType = "kdd"
		}
	}

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		clusterInfo, err := c.ClusterInformation().Get(ctx, globalClusterInfoName, options.GetOptions{})
		if err != nil {
			// Create the default config if it doesn't already exist.
			if _, ok := err.(cerrors.ErrorResourceDoesNotExist); ok {
				newClusterInfo := v3.NewClusterInformation()
				newClusterInfo.Name = globalClusterInfoName
				newClusterInfo.Spec.CalicoVersion = calicoVersion
				newClusterInfo.Spec.ClusterType = clusterType
				newClusterInfo.Spec.ClusterGUID = fmt.Sprintf("%s", hex.EncodeToString(uuid.NewV4().Bytes()))
				datastoreReady := true
				newClusterInfo.Spec.DatastoreReady = &datastoreReady
				_, err = c.ClusterInformation().Create(ctx, newClusterInfo, options.SetOptions{})
				if err != nil {
					if _, ok := err.(cerrors.ErrorResourceAlreadyExists); ok {
						log.Info("Failed to create global ClusterInformation; another node got there first.")
						time.Sleep(1 * time.Second)
						continue
					}
					log.WithError(err).WithField("ClusterInformation", newClusterInfo).Errorf("Error creating cluster information config")
					return err
				}
			} else {
				log.WithError(err).WithField("ClusterInformation", globalClusterInfoName).Errorf("Error getting cluster information config")
				return err
			}
			break
		}

		updateNeeded := false
		if calicoVersion != "" {
			// Only update the version if it's different from what we have.
			if clusterInfo.Spec.CalicoVersion != calicoVersion {
				clusterInfo.Spec.CalicoVersion = calicoVersion
				updateNeeded = true
			} else {
				log.WithField("CalicoVersion", clusterInfo.Spec.CalicoVersion).Debug("Calico version value already assigned")
			}
		}

		if clusterInfo.Spec.ClusterGUID == "" {
			clusterInfo.Spec.ClusterGUID = fmt.Sprintf("%s", hex.EncodeToString(uuid.NewV4().Bytes()))
			updateNeeded = true
		} else {
			log.WithField("ClusterGUID", clusterInfo.Spec.ClusterGUID).Debug("Cluster GUID value already set")
		}

		if clusterInfo.Spec.DatastoreReady == nil {
			// If the ready flag is nil, default it to true (but if it's explicitly false, leave
			// it as-is).
			datastoreReady := true
			clusterInfo.Spec.DatastoreReady = &datastoreReady
			updateNeeded = true
		} else {
			log.WithField("DatastoreReady", clusterInfo.Spec.DatastoreReady).Debug("DatastoreReady value already set")
		}

		if clusterType != "" {
			if clusterInfo.Spec.ClusterType == "" {
				clusterInfo.Spec.ClusterType = clusterType
				updateNeeded = true

			} else {
				allClusterTypes := strings.Split(clusterInfo.Spec.ClusterType, ",")
				existingClusterTypes := set.FromArray(allClusterTypes)
				localClusterTypes := strings.Split(clusterType, ",")

				clusterTypeUpdateNeeded := false
				for _, lct := range localClusterTypes {
					if existingClusterTypes.Contains(lct) {
						continue
					}
					clusterTypeUpdateNeeded = true
					allClusterTypes = append(allClusterTypes, lct)
				}

				if clusterTypeUpdateNeeded {
					clusterInfo.Spec.ClusterType = strings.Join(allClusterTypes, ",")
					updateNeeded = true
				}
			}
		}

		if updateNeeded {
			_, err = c.ClusterInformation().Update(ctx, clusterInfo, options.SetOptions{})
			if _, ok := err.(cerrors.ErrorResourceUpdateConflict); ok {
				log.WithError(err).WithField("ClusterInformation", clusterInfo).Warning(
					"Conflict while updating cluster information, may retry")
				time.Sleep(1 * time.Second)
				continue
			} else if err != nil {
				log.WithError(err).WithField("ClusterInformation", clusterInfo).Errorf(
					"Error updating cluster information")
				return err
			}
		}
		break
	}

	return nil
}

// Backend returns the backend client used by the v3 client.  Not exposed on the main
// client API, but available publicly for consumers that require access to the backend
// client (e.g. for syncer support).
func (c client) Backend() bapi.Client {
	return c.backend
}
