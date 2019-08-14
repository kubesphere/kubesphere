// Copyright (c) 2016-2019 Tigera, Inc. All rights reserved.
//
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

package ipam

import (
	"context"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/projectcalico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/libcalico-go/lib/set"

	bapi "github.com/projectcalico/libcalico-go/lib/backend/api"
	"github.com/projectcalico/libcalico-go/lib/backend/model"
	cerrors "github.com/projectcalico/libcalico-go/lib/errors"
	"github.com/projectcalico/libcalico-go/lib/names"
	"github.com/projectcalico/libcalico-go/lib/net"
)

const (
	// Number of retries when we have an error writing data
	// to etcd.
	datastoreRetries  = 100
	ipamKeyErrRetries = 3

	// Common attributes which may be set on allocations by clients.  Moved to the model package so they can be used
	// by the AllocationBlock code too.
	AttributePod       = model.IPAMBlockAttributePod
	AttributeNamespace = model.IPAMBlockAttributeNamespace
	AttributeNode      = model.IPAMBlockAttributeNode
	AttributeType      = model.IPAMBlockAttributeType
	AttributeTypeIPIP  = model.IPAMBlockAttributeTypeIPIP
	AttributeTypeVXLAN = model.IPAMBlockAttributeTypeVXLAN
)

var (
	ErrBlockLimit = errors.New("cannot allocate new block due to per host block limit")
)

// NewIPAMClient returns a new ipamClient, which implements Interface.
// Consumers of the Calico API should not create this directly, but should
// access IPAM through the main client IPAM accessor (e.g. clientv3.IPAM())
func NewIPAMClient(client bapi.Client, pools PoolAccessorInterface) Interface {
	return &ipamClient{
		client: client,
		pools:  pools,
		blockReaderWriter: blockReaderWriter{
			client: client,
			pools:  pools,
		},
	}
}

// ipamClient implements Interface
type ipamClient struct {
	client            bapi.Client
	pools             PoolAccessorInterface
	blockReaderWriter blockReaderWriter
}

// AutoAssign automatically assigns one or more IP addresses as specified by the
// provided AutoAssignArgs.  AutoAssign returns the list of the assigned IPv4 addresses,
// and the list of the assigned IPv6 addresses.
//
// In case of error, returns the IPs allocated so far along with the error.
func (c ipamClient) AutoAssign(ctx context.Context, args AutoAssignArgs) ([]net.IPNet, []net.IPNet, error) {
	// Determine the hostname to use - prefer the provided hostname if
	// non-nil, otherwise use the hostname reported by os.
	hostname, err := decideHostname(args.Hostname)
	if err != nil {
		return nil, nil, err
	}
	log.Infof("Auto-assign %d ipv4, %d ipv6 addrs for host '%s'", args.Num4, args.Num6, hostname)

	var v4list, v6list []net.IPNet

	if args.Num4 != 0 {
		// Assign IPv4 addresses.
		log.Debugf("Assigning IPv4 addresses")
		for _, pool := range args.IPv4Pools {
			if pool.IP.To4() == nil {
				return nil, nil, fmt.Errorf("provided IPv4 IPPools list contains one or more IPv6 IPPools")
			}
		}
		v4list, err = c.autoAssign(ctx, args.Num4, args.HandleID, args.Attrs, args.IPv4Pools, 4, hostname, args.MaxBlocksPerHost)
		if err != nil {
			log.Errorf("Error assigning IPV4 addresses: %v", err)
			return v4list, nil, err
		}
	}

	if args.Num6 != 0 {
		// If no err assigning V4, try to assign any V6.
		log.Debugf("Assigning IPv6 addresses")
		for _, pool := range args.IPv6Pools {
			if pool.IP.To4() != nil {
				return nil, nil, fmt.Errorf("provided IPv6 IPPools list contains one or more IPv4 IPPools")
			}
		}
		v6list, err = c.autoAssign(ctx, args.Num6, args.HandleID, args.Attrs, args.IPv6Pools, 6, hostname, args.MaxBlocksPerHost)
		if err != nil {
			log.Errorf("Error assigning IPV6 addresses: %v", err)
			return v4list, v6list, err
		}
	}

	return v4list, v6list, nil
}

// getBlockFromAffinity returns the block referenced by the given affinity, attempting to create it if
// it does not exist. getBlockFromAffinity will delete the provided affinity if it does not match the actual
// affinity of the block.
func (c ipamClient) getBlockFromAffinity(ctx context.Context, aff *model.KVPair) (*model.KVPair, error) {
	// Parse out affinity data.
	cidr := aff.Key.(model.BlockAffinityKey).CIDR
	host := aff.Key.(model.BlockAffinityKey).Host
	state := aff.Value.(*model.BlockAffinity).State
	logCtx := log.WithFields(log.Fields{"host": host, "cidr": cidr})

	// Get the block referenced by this affinity.
	logCtx.Info("Attempting to load block")
	b, err := c.blockReaderWriter.queryBlock(ctx, cidr, "")
	if err != nil {
		if _, ok := err.(cerrors.ErrorResourceDoesNotExist); ok {
			// The block referenced by the affinity doesn't exist. Try to create it.
			logCtx.Info("The referenced block doesn't exist, trying to create it")
			aff.Value.(*model.BlockAffinity).State = model.StatePending
			aff, err = c.blockReaderWriter.updateAffinity(ctx, aff)
			if err != nil {
				logCtx.WithError(err).Warn("Error updating block affinity")
				return nil, err
			}
			logCtx.Info("Wrote affinity as pending")

			cfg, err := c.GetIPAMConfig(ctx)
			if err != nil {
				logCtx.WithError(err).Errorf("Error getting IPAM Config")
				return nil, err
			}

			// Claim the block, which will also confirm the affinity.
			logCtx.Info("Attempting to claim the block")
			b, err := c.blockReaderWriter.claimAffineBlock(ctx, aff, *cfg)
			if err != nil {
				logCtx.WithError(err).Warn("Error claiming block")
				return nil, err
			}
			return b, nil
		}
		logCtx.WithError(err).Error("Error getting block")
		return nil, err
	}

	// If the block doesn't match the affinity, it means we've got a stale affininty hanging around.
	// We should remove it.
	blockAffinity := b.Value.(*model.AllocationBlock).Affinity
	if blockAffinity == nil || *blockAffinity != fmt.Sprintf("host:%s", host) {
		logCtx.WithField("blockAffinity", blockAffinity).Warn("Block does not match the provided affinity, deleting stale affinity")
		err := c.blockReaderWriter.deleteAffinity(ctx, aff)
		if err != nil {
			logCtx.WithError(err).Warn("Error deleting stale affinity")
			return nil, err
		}
		return nil, errStaleAffinity(fmt.Sprintf("Affinity is stale: %+v", aff))
	}

	// If the block does match the affinity but the affinity has not been confirmed,
	// try to confirm it. Treat empty string as confirmed for compatibility with older data.
	if state != model.StateConfirmed && state != "" {
		// Write the affinity as pending.
		logCtx.Info("Affinity has not been confirmed - attempt to confirm it")
		aff.Value.(*model.BlockAffinity).State = model.StatePending
		aff, err = c.blockReaderWriter.updateAffinity(ctx, aff)
		if err != nil {
			logCtx.WithError(err).Warn("Error marking affinity as pending as part of confirmation process")
			return nil, err
		}

		// CAS the block to get a new revision and invalidate any other instances
		// that might be trying to operate on the block.
		logCtx.Info("Writing block to get a new revision")
		b, err = c.blockReaderWriter.updateBlock(ctx, b)
		if err != nil {
			logCtx.WithError(err).Debug("Error writing block")
			return nil, err
		}

		// Confirm the affinity.
		logCtx.Info("Attempting to confirm affinity")
		aff.Value.(*model.BlockAffinity).State = model.StateConfirmed
		aff, err = c.blockReaderWriter.updateAffinity(ctx, aff)
		if err != nil {
			logCtx.WithError(err).Debug("Error confirming affinity")
			return nil, err
		}
		logCtx.Info("Affinity confirmed successfully")
	}
	logCtx.Info("Affinity is confirmed and block has been loaded")
	return b, nil
}

// determinePools compares a list of requested pools with the enabled pools and returns the intersect.
// If any requested pool does not exist, or is not enabled, an error is returned.
// If no pools are requested, all enabled pools are returned.
// Also applies selector logic on node labels to determine if the pool is a match.
// Returns the set of matching pools as well as the full set of ip pools.
func (c ipamClient) determinePools(requestedPoolNets []net.IPNet, version int, node v3.Node) (matchingPools, enabledPools []v3.IPPool, err error) {
	// Get all the enabled IP pools from the datastore.
	enabledPools, err = c.pools.GetEnabledPools(version)
	if err != nil {
		log.WithError(err).Errorf("Error getting IP pools")
		return
	}
	log.Debugf("enabled pools: %v", enabledPools)
	log.Debugf("requested pools: %v", requestedPoolNets)

	// Build a map so we can lookup existing pools by their CIDR.
	pm := map[string]v3.IPPool{}
	for _, p := range enabledPools {
		pm[p.Spec.CIDR] = p
	}

	// Build a list of requested IP pool objects based on the provided CIDRs, validating
	// that each one actually exists and is enabled for IPAM.
	requestedPools := []v3.IPPool{}
	for _, rp := range requestedPoolNets {
		if pool, ok := pm[rp.String()]; !ok {
			// The requested pool doesn't exist.
			err = fmt.Errorf("the given pool (%s) does not exist, or is not enabled", rp.IPNet.String())
			return
		} else {
			requestedPools = append(requestedPools, pool)
		}
	}

	// If requested IP pools are provided, use those unconditionally. We will ignore
	// IP pool selectors in this case. We need this for backwards compatibility, since IP pool
	// node selectors have not always existed.
	if len(requestedPools) > 0 {
		log.Debugf("Using the requested IP pools")
		matchingPools = requestedPools
		return
	}

	// At this point, we've determined the set of enabled IP pools which are valid for use.
	// We only want to use IP pools which actually match this node, so do a filter based on
	// selector.
	for _, pool := range enabledPools {
		var matches bool
		matches, err = pool.SelectsNode(node)
		if err != nil {
			log.WithError(err).WithField("pool", pool).Error("failed to determine if node matches pool")
			return
		}
		if !matches {
			// Do not consider pool enabled if the nodeSelector doesn't match the node's labels.
			log.Debugf("IP pool does not match this node: %s", pool.Name)
			continue
		}
		log.Debugf("IP pool matches this node: %s", pool.Name)
		matchingPools = append(matchingPools, pool)
	}

	return
}

func (c ipamClient) autoAssign(ctx context.Context, num int, handleID *string, attrs map[string]string, requestedPools []net.IPNet, version int, host string, maxNumBlocks int) ([]net.IPNet, error) {
	// Retrieve node for given hostname to use for ip pool node selection
	node, err := c.client.Get(ctx, model.ResourceKey{Kind: v3.KindNode, Name: host}, "")
	if err != nil {
		log.WithError(err).WithField("node", host).Error("failed to get node for host")
		return nil, err
	}

	// Make sure the returned value is OK.
	v3n, ok := node.Value.(*v3.Node)
	if !ok {
		return nil, fmt.Errorf("Datastore returned malformed node object")
	}

	// Determine the correct set of IP pools to use for this request.
	pools, allPools, err := c.determinePools(requestedPools, version, *v3n)
	if err != nil {
		return nil, err
	}

	// If there are no pools, we cannot assign addresses.
	if len(pools) == 0 {
		return nil, fmt.Errorf("no configured Calico pools for node %v", host)
	}

	// First, we try to assign addresses from one of the existing host-affine blocks.  We
	// always do strict checking at this stage, so it doesn't matter whether
	// globally we have strict_affinity or not.
	logCtx := log.WithFields(log.Fields{"host": host})
	if handleID != nil {
		logCtx = logCtx.WithField("handle", *handleID)
	}
	logCtx.Info("Looking up existing affinities for host")
	affBlocks, affBlocksToRelease, err := c.blockReaderWriter.getAffineBlocks(ctx, host, version, pools)
	if err != nil {
		return nil, err
	}

	// Release any emptied blocks still affine to this host but no longer part of an IP Pool which selects this node.
	for _, block := range affBlocksToRelease {
		// Determine the pool for each block.
		pool, err := c.blockReaderWriter.getPoolForIP(net.IP{block.IP}, allPools)
		if err != nil {
			log.WithError(err).Warnf("Failed to get pool for IP")
			continue
		}
		if pool == nil {
			logCtx.WithFields(log.Fields{"block": block}).Warn("No pool found for block, skipping")
			continue
		}

		// Determine if the pool selects the current node, refusing to release this particular block affinity if so.
		blockSelectsNode, err := pool.SelectsNode(*v3n)
		if err != nil {
			logCtx.WithError(err).WithField("pool", pool).Error("Failed to determine if node matches pool, skipping")
			continue
		}
		if blockSelectsNode {
			logCtx.WithFields(log.Fields{"pool": pool, "block": block}).Debug("Block's pool still selects node, refusing to remove affinity")
			continue
		}

		// Release the block affinity, requiring it to be empty.
		for i := 0; i < datastoreRetries; i++ {
			if err = c.blockReaderWriter.releaseBlockAffinity(ctx, host, block, true); err != nil {
				if _, ok := err.(errBlockClaimConflict); ok {
					// Not claimed by this host - ignore.
				} else if _, ok := err.(errBlockNotEmpty); ok {
					// Block isn't empty - ignore.
				} else if _, ok := err.(cerrors.ErrorResourceDoesNotExist); ok {
					// Block does not exist - ignore.
				} else {
					logCtx.WithError(err).WithField("block", block).Warn("Error occurred releasing block, trying again")
					continue
				}
			}
			logCtx.WithField("block", block).Info("Released affine block that no longer selects this host")
			break
		}
	}

	logCtx.Debugf("Found %d affine IPv%d blocks for host: %v", len(affBlocks), version, affBlocks)
	ips := []net.IPNet{}
	newIPs := []net.IPNet{}

	// Record how many blocks we own so we can check against the limit later.
	numBlocksOwned := len(affBlocks)

	for len(ips) < num {
		if len(affBlocks) == 0 {
			logCtx.Infof("Ran out of existing affine blocks for host")
			break
		}
		cidr := affBlocks[0]
		affBlocks = affBlocks[1:]

		// Try to assign from this block - if we hit a CAS error, we'll try this block again.
		// For any other error, we'll break out and try the next affine block.
		for i := 0; i < datastoreRetries; i++ {
			// Get the affinity.
			logCtx.Infof("Trying affinity for %s", cidr)
			aff, err := c.blockReaderWriter.queryAffinity(ctx, host, cidr, "")
			if err != nil {
				logCtx.WithError(err).Warnf("Error getting affinity")
				break
			}

			// Get the block which is referenced by the affinity, creating it if necessary.
			b, err := c.getBlockFromAffinity(ctx, aff)
			if err != nil {
				// Couldn't get a block for this affinity.
				if _, ok := err.(cerrors.ErrorResourceUpdateConflict); ok {
					logCtx.WithError(err).Debug("CAS error getting affine block - retry")
					continue
				}
				logCtx.WithError(err).Warn("Couldn't get block for affinity, try next one")
				break
			}

			// Assign IPs from the block.
			newIPs, err = c.assignFromExistingBlock(ctx, b, num, handleID, attrs, host, true)
			if err != nil {
				if _, ok := err.(cerrors.ErrorResourceUpdateConflict); ok {
					logCtx.WithError(err).Debug("CAS error assigning from affine block - retry")
					continue
				}
				logCtx.WithError(err).Warn("Couldn't assign from affine block, try next one")
				break
			}
			ips = append(ips, newIPs...)
			break
		}
		logCtx.Infof("Block '%s' provided addresses: %v", cidr.String(), newIPs)
	}

	// If there are still addresses to allocate, then we've run out of
	// existing blocks with affinity to this host.  Before we can assign new blocks or assign in
	// non-affine blocks, we need to check that our IPAM configuration
	// allows that.
	config, err := c.GetIPAMConfig(ctx)
	if err != nil {
		return ips, err
	}
	logCtx.Debugf("Allocate new blocks? Config: %+v", config)
	if config.AutoAllocateBlocks == true {
		rem := num - len(ips)
		retries := datastoreRetries
		for rem > 0 && retries > 0 {
			if maxNumBlocks > 0 && numBlocksOwned >= maxNumBlocks {
				log.Warnf("Unable to allocate a new IPAM block; host already has %v blocks but "+
					"blocks per host limit is %v", numBlocksOwned, maxNumBlocks)
				return ips, ErrBlockLimit
			}

			// Claim a new block.
			logCtx.Infof("No more affine blocks, but need to allocate %d more addresses - allocate another block", rem)
			retries = retries - 1

			// First, try to find an unclaimed block.
			logCtx.Info("Looking for an unclaimed block")
			subnet, err := c.blockReaderWriter.findUnclaimedBlock(ctx, host, version, pools, *config)
			if err != nil {
				if _, ok := err.(noFreeBlocksError); ok {
					// No free blocks.  Break.
					logCtx.Info("No free blocks available for allocation")
					break
				}
				log.WithError(err).Error("Failed to find an unclaimed block")
				return ips, err
			}
			logCtx := log.WithFields(log.Fields{"host": host, "subnet": subnet})
			logCtx.Info("Found unclaimed block")

			for i := 0; i < datastoreRetries; i++ {
				// We found an unclaimed block - claim affinity for it.
				pa, err := c.blockReaderWriter.getPendingAffinity(ctx, host, *subnet)
				if err != nil {
					if _, ok := err.(cerrors.ErrorResourceUpdateConflict); ok {
						logCtx.WithError(err).Debug("CAS error claiming pending affinity, retry")
						continue
					}
					logCtx.WithError(err).Errorf("Error claiming pending affinity")
					return ips, err
				}

				// We have an affinity - try to get the block.
				b, err := c.getBlockFromAffinity(ctx, pa)
				if err != nil {
					if _, ok := err.(cerrors.ErrorResourceUpdateConflict); ok {
						logCtx.WithError(err).Debug("CAS error getting block, retry")
						continue
					} else if _, ok := err.(errBlockClaimConflict); ok {
						logCtx.WithError(err).Debug("Block taken by someone else, find a new one")
						break
					} else if _, ok := err.(errStaleAffinity); ok {
						logCtx.WithError(err).Debug("Affinity is stale, find a new one")
						break
					}
					logCtx.WithError(err).Errorf("Error getting block for affinity")
					return ips, err
				}

				// Claim successful.  Assign addresses from the new block.
				logCtx.Infof("Claimed new block %v - assigning %d addresses", b, rem)
				numBlocksOwned++
				newIPs, err := c.assignFromExistingBlock(ctx, b, rem, handleID, attrs, host, config.StrictAffinity)
				if err != nil {
					if _, ok := err.(cerrors.ErrorResourceUpdateConflict); ok {
						log.WithError(err).Debug("CAS Error assigning from new block - retry")
						continue
					}
					logCtx.WithError(err).Warningf("Failed to assign IPs in newly allocated block")
					break
				}
				logCtx.Debugf("Assigned IPs from new block: %s", newIPs)
				ips = append(ips, newIPs...)
				rem = num - len(ips)
				break
			}
		}

		if retries == 0 {
			return ips, errors.New("Max retries hit - excessive concurrent IPAM requests")
		}
	}

	// If there are still addresses to allocate, we've now tried all blocks
	// with some affinity to us, and tried (and failed) to allocate new
	// ones.  If we do not require strict host affinity, our last option is
	// a random hunt through any blocks we haven't yet tried.
	//
	// Note that this processing simply takes all of the IP pools and breaks
	// them up into block-sized CIDRs, then shuffles and searches through each
	// CIDR.  This algorithm does not work if we disallow auto-allocation of
	// blocks because the allocated blocks may be sparsely populated in the
	// pools resulting in a very slow search for free addresses.
	//
	// If we need to support non-strict affinity and no auto-allocation of
	// blocks, then we should query the actual allocation blocks and assign
	// from those.
	rem := num - len(ips)
	if config.StrictAffinity != true && rem != 0 {
		logCtx.Infof("Attempting to assign %d more addresses from non-affine blocks", rem)

		// Iterate over pools and assign addresses until we either run out of pools,
		// or the request has been satisfied.
		logCtx.Info("Looking for blocks with free IP addresses")
		for _, p := range pools {
			logCtx.Debugf("Assigning from non-affine blocks in pool %s", p.Spec.CIDR)
			newBlock := randomBlockGenerator(p, host)
			for rem > 0 {
				// Grab a new random block.
				blockCIDR := newBlock()
				if blockCIDR == nil {
					logCtx.Warningf("All addresses exhausted in pool %s", p.Spec.CIDR)
					break
				}

				for i := 0; i < datastoreRetries; i++ {
					b, err := c.blockReaderWriter.queryBlock(ctx, *blockCIDR, "")
					if err != nil {
						logCtx.WithError(err).Warn("Failed to get non-affine block")
						break
					}

					// Attempt to assign from the block.
					logCtx.Infof("Attempting to assign IPs from non-affine block %s", blockCIDR.String())
					newIPs, err := c.assignFromExistingBlock(ctx, b, rem, handleID, attrs, host, false)
					if err != nil {
						if _, ok := err.(cerrors.ErrorResourceUpdateConflict); ok {
							logCtx.WithError(err).Debug("CAS error assigning from non-affine block - retry")
							continue
						}
						logCtx.WithError(err).Warningf("Failed to assign IPs from non-affine block in pool %s", p.Spec.CIDR)
						break
					}
					if len(newIPs) == 0 {
						break
					}
					logCtx.Infof("Successfully assigned IPs from non-affine block %s", blockCIDR.String())
					ips = append(ips, newIPs...)
					rem = num - len(ips)
					break
				}
			}
		}
	}

	logCtx.Infof("Auto-assigned %d out of %d IPv%ds: %v", len(ips), num, version, ips)
	return ips, nil
}

// AssignIP assigns the provided IP address to the provided host.  The IP address
// must fall within a configured pool.  AssignIP will claim block affinity as needed
// in order to satisfy the assignment.  An error will be returned if the IP address
// is already assigned, or if StrictAffinity is enabled and the address is within
// a block that does not have affinity for the given host.
func (c ipamClient) AssignIP(ctx context.Context, args AssignIPArgs) error {
	hostname, err := decideHostname(args.Hostname)
	if err != nil {
		return err
	}
	log.Infof("Assigning IP %s to host: %s", args.IP, hostname)

	pool, err := c.blockReaderWriter.getPoolForIP(args.IP, nil)
	if err != nil {
		return err
	}
	if pool == nil {
		return errors.New("The provided IP address is not in a configured pool\n")
	}

	blockCIDR := getBlockCIDRForAddress(args.IP, pool)
	log.Debugf("IP %s is in block '%s'", args.IP.String(), blockCIDR.String())
	for i := 0; i < datastoreRetries; i++ {
		obj, err := c.blockReaderWriter.queryBlock(ctx, blockCIDR, "")
		if err != nil {
			if _, ok := err.(cerrors.ErrorResourceDoesNotExist); !ok {
				log.WithError(err).Error("Error getting block")
				return err
			}

			log.Debugf("Block for IP %s does not yet exist, creating", args.IP)
			cfg, err := c.GetIPAMConfig(ctx)
			if err != nil {
				log.Errorf("Error getting IPAM Config: %v", err)
				return err
			}

			pa, err := c.blockReaderWriter.getPendingAffinity(ctx, hostname, blockCIDR)
			if err != nil {
				if _, ok := err.(cerrors.ErrorResourceUpdateConflict); ok {
					log.WithError(err).Debug("CAS error claiming affinity for block - retry")
					continue
				}
				return err
			}

			obj, err = c.blockReaderWriter.claimAffineBlock(ctx, pa, *cfg)
			if err != nil {
				if _, ok := err.(*errBlockClaimConflict); ok {
					log.Warningf("Someone else claimed block %s before us", blockCIDR.String())
					continue
				} else if _, ok := err.(cerrors.ErrorResourceUpdateConflict); ok {
					log.WithError(err).Debug("CAS error claiming affine block - retry")
					continue
				}
				log.WithError(err).Error("Error claiming block")
				return err
			}
			log.Infof("Claimed new block: %s", blockCIDR)
		}

		block := allocationBlock{obj.Value.(*model.AllocationBlock)}
		err = block.assign(args.IP, args.HandleID, args.Attrs, hostname)
		if err != nil {
			log.Errorf("Failed to assign address %v: %v", args.IP, err)
			return err
		}

		// Increment handle.
		if args.HandleID != nil {
			c.incrementHandle(ctx, *args.HandleID, blockCIDR, 1)
		}

		// Update the block using the original KVPair to do a CAS.  No need to
		// update the Value since we have been manipulating the Value pointed to
		// in the KVPair.
		_, err = c.blockReaderWriter.updateBlock(ctx, obj)
		if err != nil {
			if _, ok := err.(cerrors.ErrorResourceUpdateConflict); ok {
				log.WithError(err).Debug("CAS error assigning IP - retry")
				continue
			}

			log.WithError(err).Warningf("Update failed on block %s", block.CIDR.String())
			if args.HandleID != nil {
				if err := c.decrementHandle(ctx, *args.HandleID, blockCIDR, 1); err != nil {
					log.WithError(err).Warn("Failed to decrement handle")
				}
			}
			return err
		}
		return nil
	}
	return errors.New("Max retries hit - excessive concurrent IPAM requests")
}

// ReleaseIPs releases any of the given IP addresses that are currently assigned,
// so that they are available to be used in another assignment.
func (c ipamClient) ReleaseIPs(ctx context.Context, ips []net.IP) ([]net.IP, error) {
	log.Infof("Releasing IP addresses: %v", ips)
	unallocated := []net.IP{}

	// Group IP addresses by block to minimize the number of writes
	// to the datastore required to release the given addresses.
	ipsByBlock := map[string][]net.IP{}
	for _, ip := range ips {
		var cidrStr string

		pool, err := c.blockReaderWriter.getPoolForIP(ip, nil)
		if err != nil {
			log.WithError(err).Warnf("Failed to get pool for IP")
			return nil, err
		}
		if pool == nil {
			if cidr, err := c.blockReaderWriter.getBlockForIP(ctx, ip); err != nil {
				return nil, err
			} else {
				if cidr == nil {
					// The IP isn't in any block so it's already unallocated.
					unallocated = append(unallocated, ip)

					// Move on to the next IP
					continue
				}
				cidrStr = cidr.String()
			}
		} else {
			cidrStr = getBlockCIDRForAddress(ip, pool).String()
		}

		// Check if we've already got an entry for this block.
		if _, exists := ipsByBlock[cidrStr]; !exists {
			// Entry does not exist, create it.
			ipsByBlock[cidrStr] = []net.IP{}
		}

		// Append to the list.
		ipsByBlock[cidrStr] = append(ipsByBlock[cidrStr], ip)
	}

	// Release IPs for each block.
	for cidrStr, ips := range ipsByBlock {
		_, cidr, _ := net.ParseCIDR(cidrStr)
		unalloc, err := c.releaseIPsFromBlock(ctx, ips, *cidr)
		if err != nil {
			log.Errorf("Error releasing IPs: %v", err)
			return nil, err
		}
		unallocated = append(unallocated, unalloc...)
	}
	return unallocated, nil
}

func (c ipamClient) releaseIPsFromBlock(ctx context.Context, ips []net.IP, blockCIDR net.IPNet) ([]net.IP, error) {
	logCtx := log.WithField("cidr", blockCIDR)
	for i := 0; i < datastoreRetries; i++ {
		logCtx.Info("Getting block so we can release IPs")

		// Get allocation block for cidr.
		obj, err := c.blockReaderWriter.queryBlock(ctx, blockCIDR, "")
		if err != nil {
			if _, ok := err.(cerrors.ErrorResourceDoesNotExist); ok {
				// The block does not exist - all addresses must be unassigned.
				return ips, nil
			} else {
				// Unexpected error reading block.
				return nil, err
			}
		}

		// Release the IPs.
		b := allocationBlock{obj.Value.(*model.AllocationBlock)}
		unallocated, handles, err2 := b.release(ips)
		if err2 != nil {
			return nil, err2
		}
		if len(ips) == len(unallocated) {
			// All the given IP addresses are already unallocated.
			// Just return.
			logCtx.Info("No IPs need to be released")
			return unallocated, nil
		}

		// If the block is empty and has no affinity, we can delete it.
		// Otherwise, update the block using CAS.  There is no need to update
		// the Value since we have updated the structure pointed to in the
		// KVPair.
		var updateErr error
		if b.empty() && b.Affinity == nil {
			logCtx.Info("Deleting non-affine block")
			updateErr = c.blockReaderWriter.deleteBlock(ctx, obj)
		} else {
			logCtx.Info("Updating assignments in block")
			_, updateErr = c.blockReaderWriter.updateBlock(ctx, obj)
		}

		if updateErr != nil {
			if _, ok := updateErr.(cerrors.ErrorResourceUpdateConflict); ok {
				// Comparison error - retry.
				logCtx.Warningf("Failed to update block - retry #%d", i)
				continue
			} else {
				// Something else - return the error.
				logCtx.WithError(updateErr).Errorf("Error updating block")
				return nil, updateErr
			}
		}

		// Success - decrement handles.
		logCtx.Debugf("Decrementing handles: %v", handles)
		for handleID, amount := range handles {
			if err := c.decrementHandle(ctx, handleID, blockCIDR, amount); err != nil {
				logCtx.WithError(err).Warn("Failed to decrement handle")
			}
		}

		// Determine whether or not the block's pool still matches the node.
		if err := c.ensureConsistentAffinity(ctx, obj.Value.(*model.AllocationBlock)); err != nil {
			logCtx.WithError(err).Warn("Error ensuring consistent affinity but IP already released. Returning no error.")
		}
		return unallocated, nil
	}
	return nil, errors.New("Max retries hit - excessive concurrent IPAM requests")
}

func (c ipamClient) assignFromExistingBlock(ctx context.Context, block *model.KVPair, num int, handleID *string, attrs map[string]string, host string, affCheck bool) ([]net.IPNet, error) {
	blockCIDR := block.Key.(model.BlockKey).CIDR
	logCtx := log.WithFields(log.Fields{"host": host, "block": blockCIDR})
	if handleID != nil {
		logCtx = logCtx.WithField("handle", *handleID)
	}
	logCtx.Infof("Attempting to assign %d addresses from block", num)

	// Pull out the block.
	b := allocationBlock{block.Value.(*model.AllocationBlock)}

	ips, err := b.autoAssign(num, handleID, host, attrs, affCheck)
	if err != nil {
		logCtx.WithError(err).Errorf("Error in auto assign")
		return nil, err
	}
	if len(ips) == 0 {
		logCtx.Infof("Block is full")
		return []net.IPNet{}, nil
	}

	// Increment handle count.
	if handleID != nil {
		logCtx.Debug("Incrementing handle")
		c.incrementHandle(ctx, *handleID, blockCIDR, num)
	}

	// Update the block using CAS by passing back the original
	// KVPair.
	logCtx.Info("Writing block in order to claim IPs")
	block.Value = b.AllocationBlock
	_, err = c.blockReaderWriter.updateBlock(ctx, block)
	if err != nil {
		logCtx.WithError(err).Infof("Failed to update block")
		if handleID != nil {
			logCtx.Debug("Decrementing handle since we failed to allocate IP(s)")
			if err := c.decrementHandle(ctx, *handleID, blockCIDR, num); err != nil {
				logCtx.WithError(err).Warnf("Failed to decrement handle")
			}
		}
		return nil, err
	}
	logCtx.Infof("Successfully claimed IPs: %v", ips)
	return ips, nil
}

// ClaimAffinity makes a best effort to claim affinity to the given host for all blocks
// within the given CIDR.  The given CIDR must fall within a configured
// pool.  Returns a list of blocks that were claimed, as well as a
// list of blocks that were claimed by another host.
// If an empty string is passed as the host, then the hostname is automatically detected.
func (c ipamClient) ClaimAffinity(ctx context.Context, cidr net.IPNet, host string) ([]net.IPNet, []net.IPNet, error) {
	logCtx := log.WithFields(log.Fields{"host": host, "cidr": cidr})

	// Verify the requested CIDR falls within a configured pool.
	pool, err := c.blockReaderWriter.getPoolForIP(net.IP{IP: cidr.IP}, nil)
	if err != nil {
		return nil, nil, err
	}
	if pool == nil {
		estr := fmt.Sprintf("The requested CIDR (%s) is not within any configured pools.", cidr.String())
		return nil, nil, errors.New(estr)
	}

	// Validate that the given CIDR is at least as big as a block.
	if !largerThanOrEqualToBlock(cidr, pool) {
		estr := fmt.Sprintf("The requested CIDR (%s) is smaller than the minimum.", cidr.String())
		return nil, nil, invalidSizeError(estr)
	}

	// Determine the hostname to use.
	hostname, err := decideHostname(host)
	if err != nil {
		return nil, nil, err
	}

	failed := []net.IPNet{}
	claimed := []net.IPNet{}

	// Get IPAM config.
	cfg, err := c.GetIPAMConfig(ctx)
	if err != nil {
		logCtx.Errorf("Failed to get IPAM Config: %v", err)
		return nil, nil, err
	}

	// Claim all blocks within the given cidr.
	blocks := blockGenerator(pool, cidr)
	for blockCIDR := blocks(); blockCIDR != nil; blockCIDR = blocks() {
		for i := 0; i < datastoreRetries; i++ {
			// First, claim a pending affinity.
			pa, err := c.blockReaderWriter.getPendingAffinity(ctx, hostname, *blockCIDR)
			if err != nil {
				if _, ok := err.(cerrors.ErrorResourceUpdateConflict); ok {
					logCtx.WithError(err).Debug("CAS error getting pending affinity - retry")
					continue
				}
				return claimed, failed, err
			}

			// Once we have the affinity, claim the block, which will confirm the affinity.
			_, err = c.blockReaderWriter.claimAffineBlock(ctx, pa, *cfg)
			if err != nil {
				if _, ok := err.(cerrors.ErrorResourceUpdateConflict); ok {
					logCtx.WithError(err).Debug("CAS error claiming affine block - retry")
					continue
				} else if _, ok := err.(errBlockClaimConflict); ok {
					logCtx.Debugf("Block %s is claimed by another host", blockCIDR.String())
					failed = append(failed, *blockCIDR)
				} else {
					logCtx.Errorf("Failed to claim block: %v", err)
					return claimed, failed, err
				}
			} else {
				logCtx.Debugf("Claimed CIDR %s", blockCIDR.String())
				claimed = append(claimed, *blockCIDR)
			}
			break
		}
	}
	return claimed, failed, nil
}

// ReleaseAffinity releases affinity for all blocks within the given CIDR
// on the given host.  If a block does not have affinity for the given host,
// its affinity will not be released and no error will be returned.
// If an empty string is passed as the host, then the hostname is automatically detected.
func (c ipamClient) ReleaseAffinity(ctx context.Context, cidr net.IPNet, host string, mustBeEmpty bool) error {
	// Verify the requested CIDR falls within a configured pool.
	fields := log.Fields{"cidr": cidr.String(), "host": host, "mustBeEmpty": mustBeEmpty}
	log.WithFields(fields).Debugf("Releasing affinity for CIDR")
	pool, err := c.blockReaderWriter.getPoolForIP(net.IP{IP: cidr.IP}, nil)
	if pool == nil {
		estr := fmt.Sprintf("The requested CIDR (%s) is not within any configured pools.", cidr.String())
		return errors.New(estr)
	}

	// Validate that the given CIDR is at least as big as a block.
	if !largerThanOrEqualToBlock(cidr, pool) {
		estr := fmt.Sprintf("The requested CIDR (%s) is smaller than the minimum.", cidr.String())
		return invalidSizeError(estr)
	}

	// Determine the hostname to use.
	hostname, err := decideHostname(host)
	if err != nil {
		return err
	}

	// Release all blocks within the given cidr.
	blocks := blockGenerator(pool, cidr)
	for blockCIDR := blocks(); blockCIDR != nil; blockCIDR = blocks() {
		logCtx := log.WithField("cidr", blockCIDR)
		for i := 0; i < datastoreRetries; i++ {
			err := c.blockReaderWriter.releaseBlockAffinity(ctx, hostname, *blockCIDR, mustBeEmpty)
			if err != nil {
				if _, ok := err.(errBlockClaimConflict); ok {
					// Not claimed by this host - ignore.
				} else if _, ok := err.(cerrors.ErrorResourceDoesNotExist); ok {
					// Block does not exist - ignore.
				} else if _, ok := err.(cerrors.ErrorResourceUpdateConflict); ok {
					logCtx.WithError(err).Debug("CAS error releasing block affinity - retry")
					continue
				} else {
					logCtx.WithError(err).Errorf("Error releasing affinity")
					return err
				}
			}
			break
		}
	}
	return nil
}

// ReleaseHostAffinities releases affinity for all blocks that are affine
// to the given host.  If an empty string is passed as the host,
// then the hostname is automatically detected.
func (c ipamClient) ReleaseHostAffinities(ctx context.Context, host string, mustBeEmpty bool) error {
	log.Debugf("Releasing affinities for host %s. MustBeEmpty? %v", host, mustBeEmpty)
	hostname, err := decideHostname(host)
	if err != nil {
		return err
	}

	versions := []int{4, 6}
	for _, version := range versions {
		blockCIDRs, _, err := c.blockReaderWriter.getAffineBlocks(ctx, hostname, version, nil)
		if err != nil {
			return err
		}

		for _, blockCIDR := range blockCIDRs {
			err := c.ReleaseAffinity(ctx, blockCIDR, hostname, mustBeEmpty)
			if err != nil {
				if _, ok := err.(errBlockClaimConflict); ok {
					// Claimed by a different host.
				} else {
					return err
				}
			}
		}
	}
	return nil
}

// ReleasePoolAffinities releases affinity for all blocks within
// the specified pool across all hosts.
func (c ipamClient) ReleasePoolAffinities(ctx context.Context, pool net.IPNet) error {
	log.Infof("Releasing block affinities within pool '%s'", pool.String())
	for i := 0; i < ipamKeyErrRetries; i++ {
		retry := false
		pairs, err := c.hostBlockPairs(ctx, pool)
		if err != nil {
			return err
		}

		if len(pairs) == 0 {
			log.Debugf("No blocks have affinity")
			return nil
		}

		for blockString, host := range pairs {
			_, blockCIDR, _ := net.ParseCIDR(blockString)
			logCtx := log.WithField("cidr", blockCIDR)
			for i := 0; i < datastoreRetries; i++ {
				err = c.blockReaderWriter.releaseBlockAffinity(ctx, host, *blockCIDR, false)
				if err != nil {
					if _, ok := err.(errBlockClaimConflict); ok {
						retry = true
					} else if _, ok := err.(cerrors.ErrorResourceDoesNotExist); ok {
						logCtx.Debugf("No such block")
						break
					} else if _, ok := err.(cerrors.ErrorResourceUpdateConflict); ok {
						logCtx.WithError(err).Debug("CAS error releasing block affinity - retry")
						continue
					} else {
						logCtx.WithError(err).Errorf("Error releasing affinity")
						return err
					}
				}
				break
			}
		}

		if !retry {
			return nil
		}
	}
	return errors.New("Max retries hit - excessive concurrent IPAM requests")
}

// RemoveIPAMHost releases affinity for all blocks on the given host,
// and removes all host-specific IPAM data from the datastore.
// RemoveIPAMHost does not release any IP addresses claimed on the given host.
// If an empty string is passed as the host, then the hostname is automatically detected.
func (c ipamClient) RemoveIPAMHost(ctx context.Context, host string) error {
	// Determine the hostname to use.
	hostname, err := decideHostname(host)
	if err != nil {
		return err
	}
	logCtx := log.WithField("host", hostname)
	logCtx.Info("Removing IPAM data for host")

	for i := 0; i < datastoreRetries; i++ {
		// Release affinities for this host.
		logCtx.Info("Releasing IPAM affinities for host")
		if err := c.ReleaseHostAffinities(ctx, hostname, false); err != nil {
			logCtx.WithError(err).Errorf("Failed to release IPAM affinities for host")
			return err
		}

		// Get the IPAM host.
		logCtx.Info("Querying IPAM host tree in data store")
		k := model.IPAMHostKey{Host: hostname}
		kvp, err := c.client.Get(ctx, k, "")
		if err != nil {
			if _, ok := err.(cerrors.ErrorOperationNotSupported); ok {
				// KDD mode doesn't have this object - this is a no-op.
				logCtx.Debugf("No need to remove IPAM host for this datastore")
				return nil
			}
			if _, ok := err.(cerrors.ErrorResourceDoesNotExist); !ok {
				logCtx.WithError(err).Errorf("Failed to get IPAM host")
				return err
			}

			// Resource does not exist, no need to remove it.
			logCtx.Info("IPAM host data does not exist")
			return nil
		}

		// Remove the host tree from the datastore.
		logCtx.Info("Deleting IPAM host tree from data store")
		_, err = c.client.Delete(ctx, k, kvp.Revision)
		if err != nil {
			if _, ok := err.(cerrors.ErrorResourceUpdateConflict); ok {
				// We hit a compare-and-delete error - retry.
				continue
			}

			// Return the error unless the resource does not exist.
			if _, ok := err.(cerrors.ErrorResourceDoesNotExist); !ok {
				logCtx.Errorf("Error removing IPAM host: %v", err)
				return err
			}
		}
		logCtx.Info("Successfully deleted IPAM host data")
		return nil
	}

	return errors.New("Max retries hit")
}

func (c ipamClient) hostBlockPairs(ctx context.Context, pool net.IPNet) (map[string]string, error) {
	pairs := map[string]string{}

	// Get all blocks and their affinities.
	objs, err := c.client.List(ctx, model.BlockAffinityListOptions{}, "")
	if err != nil {
		log.Errorf("Error querying block affinities: %v", err)
		return nil, err
	}

	// Iterate through each block affinity and build up a mapping
	// of blockCidr -> host.
	log.Debugf("Getting block -> host mappings")
	for _, o := range objs.KVPairs {
		k := o.Key.(model.BlockAffinityKey)

		// Only add the pair to the map if the block belongs to the pool.
		if pool.Contains(k.CIDR.IPNet.IP) {
			pairs[k.CIDR.String()] = k.Host
		}
		log.Debugf("Block %s -> %s", k.CIDR.String(), k.Host)
	}

	return pairs, nil
}

// IpsByHandle returns a list of all IP addresses that have been
// assigned using the provided handle.
func (c ipamClient) IPsByHandle(ctx context.Context, handleID string) ([]net.IP, error) {
	obj, err := c.blockReaderWriter.queryHandle(ctx, handleID, "")
	if err != nil {
		return nil, err
	}
	handle := allocationHandle{obj.Value.(*model.IPAMHandle)}

	assignments := []net.IP{}
	for k, _ := range handle.Block {
		_, blockCIDR, _ := net.ParseCIDR(k)
		obj, err := c.blockReaderWriter.queryBlock(ctx, *blockCIDR, "")
		if err != nil {
			log.WithError(err).Warningf("Couldn't read block %s referenced by handle %s", blockCIDR, handleID)
			continue
		}

		// Pull out the allocationBlock and get all the assignments from it.
		b := allocationBlock{obj.Value.(*model.AllocationBlock)}
		assignments = append(assignments, b.ipsByHandle(handleID)...)
	}
	return assignments, nil
}

// ReleaseByHandle releases all IP addresses that have been assigned
// using the provided handle.
func (c ipamClient) ReleaseByHandle(ctx context.Context, handleID string) error {
	log.Infof("Releasing all IPs with handle '%s'", handleID)
	obj, err := c.blockReaderWriter.queryHandle(ctx, handleID, "")
	if err != nil {
		return err
	}
	handle := allocationHandle{obj.Value.(*model.IPAMHandle)}

	for blockStr, _ := range handle.Block {
		_, blockCIDR, _ := net.ParseCIDR(blockStr)
		if err := c.releaseByHandle(ctx, handleID, *blockCIDR); err != nil {
			return err
		}
	}
	return nil
}

func (c ipamClient) releaseByHandle(ctx context.Context, handleID string, blockCIDR net.IPNet) error {
	logCtx := log.WithFields(log.Fields{"handle": handleID, "cidr": blockCIDR})
	for i := 0; i < datastoreRetries; i++ {
		logCtx.Debug("Querying block so we can release IPs by handle")
		obj, err := c.blockReaderWriter.queryBlock(ctx, blockCIDR, "")
		if err != nil {
			if _, ok := err.(cerrors.ErrorResourceDoesNotExist); ok {
				// Block doesn't exist, so all addresses are already
				// unallocated.  This can happen when a handle is
				// overestimating the number of assigned addresses.
				return nil
			} else {
				return err
			}
		}

		// Release the IP by handle.
		block := allocationBlock{obj.Value.(*model.AllocationBlock)}
		num := block.releaseByHandle(handleID)
		if num == 0 {
			// Block has no addresses with this handle, so
			// all addresses are already unallocated.
			logCtx.Debug("Block has no addresses with the given handle")
			return nil
		}
		logCtx.Debugf("Block has %d IPs with the given handle", num)

		if block.empty() && block.Affinity == nil {
			logCtx.Info("Deleting block because it is now empty and has no affinity")
			err = c.blockReaderWriter.deleteBlock(ctx, obj)
			if err != nil {
				if _, ok := err.(cerrors.ErrorResourceUpdateConflict); ok {
					logCtx.Debug("CAD error deleting block - retry")
					continue
				}

				// Return the error unless the resource does not exist.
				if _, ok := err.(cerrors.ErrorResourceDoesNotExist); !ok {
					logCtx.Errorf("Error deleting block: %v", err)
					return err
				}
			}
			logCtx.Info("Successfully deleted empty block")
		} else {
			// Compare and swap the AllocationBlock using the original
			// KVPair read from before.  No need to update the Value since we
			// have been directly manipulating the value referenced by the KVPair.
			logCtx.Debug("Updating block to release IPs")
			_, err = c.blockReaderWriter.updateBlock(ctx, obj)
			if err != nil {
				if _, ok := err.(cerrors.ErrorResourceUpdateConflict); ok {
					// Comparison failed - retry.
					logCtx.Warningf("CAS error for block, retry #%d: %v", i, err)
					continue
				} else {
					// Something else - return the error.
					logCtx.Errorf("Error updating block '%s': %v", block.CIDR.String(), err)
					return err
				}
			}
			logCtx.Debug("Successfully released IPs from block")
		}
		if err = c.decrementHandle(ctx, handleID, blockCIDR, num); err != nil {
			logCtx.WithError(err).Warn("Failed to decrement handle")
		}

		// Determine whether or not the block's pool still matches the node.
		if err = c.ensureConsistentAffinity(ctx, block.AllocationBlock); err != nil {
			logCtx.WithError(err).Warn("Error ensuring consistent affinity but IP already released. Returning no error.")
		}
		return nil
	}
	return errors.New("Hit max retries")
}

func (c ipamClient) incrementHandle(ctx context.Context, handleID string, blockCIDR net.IPNet, num int) error {
	var obj *model.KVPair
	var err error
	for i := 0; i < datastoreRetries; i++ {
		obj, err = c.blockReaderWriter.queryHandle(ctx, handleID, "")
		if err != nil {
			if _, ok := err.(cerrors.ErrorResourceDoesNotExist); ok {
				// Handle doesn't exist - create it.
				log.Infof("Creating new handle: %s", handleID)
				bh := model.IPAMHandle{
					HandleID: handleID,
					Block:    map[string]int{},
				}
				obj = &model.KVPair{
					Key:   model.IPAMHandleKey{HandleID: handleID},
					Value: &bh,
				}
			} else {
				// Unexpected error reading handle.
				return err
			}
		}

		// Get the handle from the KVPair.
		handle := allocationHandle{obj.Value.(*model.IPAMHandle)}

		// Increment the handle for this block.
		handle.incrementBlock(blockCIDR, num)

		// Compare and swap the handle using the KVPair from above.  We've been
		// manipulating the structure in the KVPair, so pass straight back to
		// apply the changes.
		if obj.Revision != "" {
			// This is an existing handle - update it.
			_, err = c.blockReaderWriter.updateHandle(ctx, obj)
			if err != nil {
				log.WithError(err).Warning("Failed to update handle, retry")
				continue
			}
		} else {
			// This is a new handle - create it.
			_, err = c.client.Create(ctx, obj)
			if err != nil {
				log.WithError(err).Warning("Failed to create handle, retry")
				continue
			}
		}
		return nil
	}
	return errors.New("Max retries hit - excessive concurrent IPAM requests")

}

func (c ipamClient) decrementHandle(ctx context.Context, handleID string, blockCIDR net.IPNet, num int) error {
	for i := 0; i < datastoreRetries; i++ {
		obj, err := c.blockReaderWriter.queryHandle(ctx, handleID, "")
		if err != nil {
			return err
		}
		handle := allocationHandle{obj.Value.(*model.IPAMHandle)}

		_, err = handle.decrementBlock(blockCIDR, num)
		if err != nil {
			return err
		}

		// Update / Delete as appropriate.  Since we have been manipulating the
		// data in the KVPair, just pass this straight back to the client.
		if handle.empty() {
			log.Debugf("Deleting handle: %s", handleID)
			if err = c.blockReaderWriter.deleteHandle(ctx, obj); err != nil {
				if err != nil {
					if _, ok := err.(cerrors.ErrorResourceUpdateConflict); ok {
						// Update conflict - retry.
						continue
					} else if _, ok := err.(cerrors.ErrorResourceDoesNotExist); !ok {
						return err
					}
					// Already deleted.
				}
			}
		} else {
			log.Debugf("Updating handle: %s", handleID)
			if _, err = c.blockReaderWriter.updateHandle(ctx, obj); err != nil {
				if _, ok := err.(cerrors.ErrorResourceUpdateConflict); ok {
					// Update conflict - retry.
					continue
				}
				return err
			}
		}

		log.Debugf("Decremented handle '%s' by %d", handleID, num)
		return nil
	}
	return errors.New("Max retries hit - excessive concurrent IPAM requests")
}

// GetAssignmentAttributes returns the attributes stored with the given IP address
// upon assignment.
func (c ipamClient) GetAssignmentAttributes(ctx context.Context, addr net.IP) (map[string]string, error) {
	pool, err := c.blockReaderWriter.getPoolForIP(addr, nil)
	if err != nil {
		return nil, err
	}
	if pool == nil {
		log.Errorf("Error reading pool for %s", addr.String())
		return nil, errors.New(fmt.Sprintf("%s is not part of a configured pool", addr))
	}
	blockCIDR := getBlockCIDRForAddress(addr, pool)
	obj, err := c.blockReaderWriter.queryBlock(ctx, blockCIDR, "")
	if err != nil {
		log.Errorf("Error reading block %s: %v", blockCIDR, err)
		return nil, errors.New(fmt.Sprintf("%s is not assigned", addr))
	}
	block := allocationBlock{obj.Value.(*model.AllocationBlock)}
	return block.attributesForIP(addr)
}

// GetIPAMConfig returns the global IPAM configuration.  If no IPAM configuration
// has been set, returns a default configuration with StrictAffinity disabled
// and AutoAllocateBlocks enabled.
func (c ipamClient) GetIPAMConfig(ctx context.Context) (*IPAMConfig, error) {
	obj, err := c.client.Get(ctx, model.IPAMConfigKey{}, "")
	if err != nil {
		if _, ok := err.(cerrors.ErrorResourceDoesNotExist); ok {
			// IPAMConfig has not been explicitly set.  Return
			// a default IPAM configuration.
			return &IPAMConfig{AutoAllocateBlocks: true, StrictAffinity: false}, nil
		}
		log.Errorf("Error getting IPAMConfig: %v", err)
		return nil, err
	}
	return c.convertBackendToIPAMConfig(obj.Value.(*model.IPAMConfig)), nil
}

// SetIPAMConfig sets global IPAM configuration.  This can only
// be done when there are no allocated blocks and IP addresses.
func (c ipamClient) SetIPAMConfig(ctx context.Context, cfg IPAMConfig) error {
	current, err := c.GetIPAMConfig(ctx)
	if err != nil {
		return err
	}

	if *current == cfg {
		return nil
	}

	if !cfg.StrictAffinity && !cfg.AutoAllocateBlocks {
		return errors.New("Cannot disable 'StrictAffinity' and 'AutoAllocateBlocks' at the same time")
	}

	allObjs, err := c.client.List(ctx, model.BlockListOptions{}, "")
	if len(allObjs.KVPairs) != 0 {
		return errors.New("Cannot change IPAM config while allocations exist")
	}

	// Write to datastore.
	obj := model.KVPair{
		Key:   model.IPAMConfigKey{},
		Value: c.convertIPAMConfigToBackend(&cfg),
	}
	_, err = c.client.Apply(ctx, &obj)
	if err != nil {
		log.Errorf("Error applying IPAMConfig: %v", err)
		return err
	}
	return nil
}

func (c ipamClient) convertIPAMConfigToBackend(cfg *IPAMConfig) *model.IPAMConfig {
	return &model.IPAMConfig{
		StrictAffinity:     cfg.StrictAffinity,
		AutoAllocateBlocks: cfg.AutoAllocateBlocks,
	}
}

func (c ipamClient) convertBackendToIPAMConfig(cfg *model.IPAMConfig) *IPAMConfig {
	return &IPAMConfig{
		StrictAffinity:     cfg.StrictAffinity,
		AutoAllocateBlocks: cfg.AutoAllocateBlocks,
	}
}

// ensureConsistentAffinity retrieves the pool and node for the given block and determines
// if the pool still selects node. If it no longer matches, it will release the block
// affinity for that node.
// Returns a bool indicating if the block affinity was released.
func (c ipamClient) ensureConsistentAffinity(ctx context.Context, b *model.AllocationBlock) error {

	// Retrieve node for this allocation. We do this so we can clean up affinity for blocks
	// which should no longer be affine to this host.
	host := getHostAffinity(b)
	logCtx := log.WithFields(log.Fields{"cidr": b.CIDR, "host": host})

	// If no hostname is found on the block affinity,
	// there is no need to do an ip pool node selection check.
	if host == "" {
		logCtx.Debug("Block already has no affinity")
		return nil
	}

	// If the IP pool which owns this block no longer selects this node,
	// we should release the block's affinity to this node so it can be
	// used elsewhere.
	logCtx.Debugf("Looking up node labels for host affinity")
	node, err := c.client.Get(ctx, model.ResourceKey{Kind: v3.KindNode, Name: host}, "")
	if err != nil {
		if _, ok := err.(cerrors.ErrorResourceDoesNotExist); !ok {
			logCtx.WithError(err).WithField("node", host).Error("Failed to get node for host")
			return err
		}
		logCtx.Info("Node doesn't exist, no need to release affinity")
		return nil
	}

	// Make sure the returned value is a valid node.
	v3n, ok := node.Value.(*v3.Node)
	if !ok {
		return fmt.Errorf("Datastore returned malformed node object")
	}

	// Fetch the pool for the given CIDR and check if it selects the node.
	pool, err := c.blockReaderWriter.getPoolForIP(net.IP{b.CIDR.IPNet.IP}, nil)
	if err != nil {
		return err
	}

	if pool == nil {
		logCtx.Debug("No pools own this block")
		return nil
	} else if sel, err := pool.SelectsNode(*v3n); err != nil {
		logCtx.WithField("selector", pool.Spec.NodeSelector).WithError(err).Error("Failed to determine node selection")
		return err
	} else if sel {
		logCtx.Debug("Pool selects node, no change")
		return nil
	}
	logCtx.WithField("selector", pool.Spec.NodeSelector).Debug("Pool no longer selects node, releasing block affinity")

	// Pool does not match this node's label, release this block's affinity.
	if err = c.blockReaderWriter.releaseBlockAffinity(ctx, host, b.CIDR, true); err != nil {
		if _, ok := err.(errBlockClaimConflict); ok {
			// Not claimed by this host - ignore.
		} else if _, ok := err.(errBlockNotEmpty); ok {
			// Block isn't empty - ignore.
		} else if _, ok := err.(cerrors.ErrorResourceDoesNotExist); ok {
			// Block does not exist - ignore.
		} else {
			return err
		}
	}

	return nil
}

func decideHostname(host string) (string, error) {
	// Determine the hostname to use - prefer the provided hostname if
	// non-nil, otherwise use the hostname reported by os.
	var hostname string
	var err error
	if host != "" {
		hostname = host
	} else {
		hostname, err = names.Hostname()
		if err != nil {
			return "", fmt.Errorf("Failed to acquire hostname: %+v", err)
		}
	}
	log.Debugf("Using hostname=%s", hostname)
	return hostname, nil
}

// GetUtilization returns IP utilization info for the specified pools, or for all pools.
func (c ipamClient) GetUtilization(ctx context.Context, args GetUtilizationArgs) ([]*PoolUtilization, error) {
	var usage []*PoolUtilization

	// Read all pools.
	allPools, err := c.pools.GetAllPools()
	if err != nil {
		log.WithError(err).Errorf("Error getting IP pools")
		return nil, err
	}

	// Identify the ones we want and create a PoolUtilization for each of those.
	wantAllPools := len(args.Pools) == 0
	wantedPools := set.FromArray(args.Pools)
	for _, pool := range allPools {
		if wantAllPools ||
			wantedPools.Contains(pool.Name) ||
			wantedPools.Contains(pool.Spec.CIDR) {
			usage = append(usage, &PoolUtilization{
				Name: pool.Name,
				CIDR: net.MustParseNetwork(pool.Spec.CIDR).IPNet,
			})
		}
	}

	// If we've been asked for all pools, also report utilization for any allocation
	// blocks for which there is no longer an IP pool.  Note: following code depends
	// on this being at the end of the list; otherwise it will suck in allocation
	// blocks that should be reported under other pools.
	if wantAllPools {
		usage = append(usage, &PoolUtilization{
			Name: "orphaned allocation blocks",
			CIDR: net.MustParseNetwork("0.0.0.0/0").IPNet,
		})
	}

	// Read all allocation blocks.
	blocks, err := c.client.List(ctx, model.BlockListOptions{}, "")
	if err != nil {
		return nil, err
	}
	for _, kvp := range blocks.KVPairs {
		b := kvp.Value.(*model.AllocationBlock)
		log.Debugf("Got block: %v", b)

		// Find which pool this block belongs to.
		for _, poolUse := range usage {
			if b.CIDR.IsNetOverlap(poolUse.CIDR) {
				log.Debugf("Block CIDR %v belongs to pool %v", b.CIDR, poolUse.Name)
				poolUse.Blocks = append(poolUse.Blocks, BlockUtilization{
					CIDR:      b.CIDR.IPNet,
					Capacity:  b.NumAddresses(),
					Available: len(b.Unallocated),
				})
				break
			}
		}
	}
	return usage, nil
}
