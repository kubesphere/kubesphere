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
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus"

	apiv3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/libcalico-go/lib/backend/model"
	cerrors "github.com/projectcalico/libcalico-go/lib/errors"
	cnet "github.com/projectcalico/libcalico-go/lib/net"
	"github.com/projectcalico/libcalico-go/lib/options"
	validator "github.com/projectcalico/libcalico-go/lib/validator/v3"
	"github.com/projectcalico/libcalico-go/lib/watch"
)

// IPPoolInterface has methods to work with IPPool resources.
type IPPoolInterface interface {
	Create(ctx context.Context, res *apiv3.IPPool, opts options.SetOptions) (*apiv3.IPPool, error)
	Update(ctx context.Context, res *apiv3.IPPool, opts options.SetOptions) (*apiv3.IPPool, error)
	Delete(ctx context.Context, name string, opts options.DeleteOptions) (*apiv3.IPPool, error)
	Get(ctx context.Context, name string, opts options.GetOptions) (*apiv3.IPPool, error)
	List(ctx context.Context, opts options.ListOptions) (*apiv3.IPPoolList, error)
	Watch(ctx context.Context, opts options.ListOptions) (watch.Interface, error)
}

// ipPools implements IPPoolInterface
type ipPools struct {
	client client
}

// Create takes the representation of a IPPool and creates it.  Returns the stored
// representation of the IPPool, and an error, if there is any.
func (r ipPools) Create(ctx context.Context, res *apiv3.IPPool, opts options.SetOptions) (*apiv3.IPPool, error) {
	if res != nil {
		// Since we're about to default some fields, take a (shallow) copy of the input data
		// before we do so.
		resCopy := *res
		res = &resCopy
	}
	// Validate the IPPool before creating the resource.
	if err := r.validateAndSetDefaults(ctx, res, nil); err != nil {
		return nil, err
	}

	if err := validator.Validate(res); err != nil {
		return nil, err
	}

	// Check that there are no existing blocks in the pool range that have a different block size.
	poolBlockSize := res.Spec.BlockSize
	poolIP, poolCIDR, err := net.ParseCIDR(res.Spec.CIDR)
	if err != nil {
		return nil, cerrors.ErrorParsingDatastoreEntry{
			RawKey:   "CIDR",
			RawValue: string(res.Spec.CIDR),
			Err:      err,
		}
	}

	ipVersion := 4
	if poolIP.To4() == nil {
		ipVersion = 6
	}

	blocks, err := r.client.backend.List(ctx, model.BlockListOptions{IPVersion: ipVersion}, "")
	if _, ok := err.(cerrors.ErrorOperationNotSupported); !ok && err != nil {
		// There was an error and it wasn't OperationNotSupported - return it.
		return nil, err
	} else if err == nil {
		// Skip the block check if the error is OperationUnsupported - listing blocks is not
		// supported with host-local IPAM on KDD.
		for _, b := range blocks.KVPairs {
			k := b.Key.(model.BlockKey)
			ones, _ := k.CIDR.Mask.Size()
			// Check if this block has a different size to the pool, and that it overlaps with the pool.
			if ones != poolBlockSize && k.CIDR.IsNetOverlap(*poolCIDR) {
				return nil, cerrors.ErrorValidation{
					ErroredFields: []cerrors.ErroredField{{
						Name:   "IPPool.Spec.BlockSize",
						Reason: "IPPool blocksSize conflicts with existing allocations that use a different blockSize",
						Value:  res.Spec.BlockSize,
					}},
				}
			}
		}
	}

	// Enable IPIP or VXLAN globally if required.  Do this before the Create so if it fails the user
	// can retry the same command.
	err = r.maybeEnableIPIP(ctx, res)
	if err != nil {
		return nil, err
	}
	err = r.maybeEnableVXLAN(ctx, res)
	if err != nil {
		return nil, err
	}

	out, err := r.client.resources.Create(ctx, opts, apiv3.KindIPPool, res)
	if out != nil {
		return out.(*apiv3.IPPool), err
	}
	return nil, err

}

// Update takes the representation of a IPPool and updates it. Returns the stored
// representation of the IPPool, and an error, if there is any.
func (r ipPools) Update(ctx context.Context, res *apiv3.IPPool, opts options.SetOptions) (*apiv3.IPPool, error) {
	if res != nil {
		// Since we're about to default some fields, take a (shallow) copy of the input data
		// before we do so.
		resCopy := *res
		res = &resCopy
	}

	// Get the existing settings, so that we can validate the CIDR and block size have not changed.
	old, err := r.Get(ctx, res.Name, options.GetOptions{})
	if err != nil {
		return nil, err
	}

	// Validate the IPPool updating the resource.
	if err := r.validateAndSetDefaults(ctx, res, old); err != nil {
		return nil, err
	}

	if err := validator.Validate(res); err != nil {
		return nil, err
	}

	// Enable IPIP globally if required.  Do this before the Update so if it fails the user
	// can retry the same command.
	err = r.maybeEnableIPIP(ctx, res)
	if err != nil {
		return nil, err
	}
	err = r.maybeEnableVXLAN(ctx, res)
	if err != nil {
		return nil, err
	}

	out, err := r.client.resources.Update(ctx, opts, apiv3.KindIPPool, res)
	if out != nil {
		return out.(*apiv3.IPPool), err
	}
	return nil, err
}

// Delete takes name of the IPPool and deletes it. Returns an error if one occurs.
func (r ipPools) Delete(ctx context.Context, name string, opts options.DeleteOptions) (*apiv3.IPPool, error) {
	// Deleting a pool requires a little care because of existing endpoints
	// using IP addresses allocated in the pool.  We do the deletion in
	// the following steps:
	// -  disable the pool so no more IPs are assigned from it
	// -  remove all affinities associated with the pool
	// -  delete the pool

	// Get the pool so that we can find the CIDR associated with it.
	pool, err := r.Get(ctx, name, options.GetOptions{})
	if err != nil {
		return nil, err
	}

	logCxt := log.WithFields(log.Fields{
		"CIDR": pool.Spec.CIDR,
		"Name": name,
	})

	// If the pool is active, set the disabled flag to ensure we stop allocating from this pool.
	if !pool.Spec.Disabled {
		logCxt.Info("Disabling pool to release affinities")
		pool.Spec.Disabled = true

		// If the Delete has been called with a ResourceVersion then use that to perform the
		// update - that way we'll catch update conflicts (we could actually check here, but
		// the most likely scenario is there isn't one - so just pass it through and let the
		// Update handle any conflicts).
		if opts.ResourceVersion != "" {
			pool.ResourceVersion = opts.ResourceVersion
		}
		if _, err := r.Update(ctx, pool, options.SetOptions{}); err != nil {
			return nil, err
		}

		// Reset the resource version before the actual delete since the version of that resource
		// will now have been updated.
		opts.ResourceVersion = ""
	}

	// Release affinities associated with this pool.  We do this even if the pool was disabled
	// (since it may have been enabled at one time, and if there are no affine blocks created
	// then this will be a no-op).  We've already validated the CIDR so we know it will parse.
	if _, cidrNet, err := cnet.ParseCIDR(pool.Spec.CIDR); err == nil {
		logCxt.Info("Releasing pool affinities")

		// Pause for a short period before releasing the affinities - this gives any in-progress
		// allocations an opportunity to finish.
		time.Sleep(500 * time.Millisecond)
		err = r.client.IPAM().ReleasePoolAffinities(ctx, *cidrNet)

		// Depending on the datastore, IPAM may not be supported.  If we get a not supported
		// error, then continue.  Any other error, fail.
		if _, ok := err.(cerrors.ErrorOperationNotSupported); !ok && err != nil {
			return nil, err
		}
	}

	// And finally, delete the pool.
	logCxt.Info("Deleting pool")
	out, err := r.client.resources.Delete(ctx, opts, apiv3.KindIPPool, noNamespace, name)
	if out != nil {
		return out.(*apiv3.IPPool), err
	}
	return nil, err
}

// Get takes name of the IPPool, and returns the corresponding IPPool object,
// and an error if there is any.
func (r ipPools) Get(ctx context.Context, name string, opts options.GetOptions) (*apiv3.IPPool, error) {
	out, err := r.client.resources.Get(ctx, opts, apiv3.KindIPPool, noNamespace, name)
	if out != nil {
		convertIpPoolFromStorage(out.(*apiv3.IPPool))
		return out.(*apiv3.IPPool), err
	}

	return nil, err
}

// List returns the list of IPPool objects that match the supplied options.
func (r ipPools) List(ctx context.Context, opts options.ListOptions) (*apiv3.IPPoolList, error) {
	res := &apiv3.IPPoolList{}
	if err := r.client.resources.List(ctx, opts, apiv3.KindIPPool, apiv3.KindIPPoolList, res); err != nil {
		return nil, err
	}

	// Default values when reading from backend.
	for i := range res.Items {
		convertIpPoolFromStorage(&res.Items[i])
	}

	return res, nil
}

// Default pool values when reading from storage
func convertIpPoolFromStorage(pool *apiv3.IPPool) error {
	// Default the blockSize if it wasn't previously set
	if pool.Spec.BlockSize == 0 {
		// Get the IP address of the CIDR to find the IP version
		ipAddr, _, err := cnet.ParseCIDR(pool.Spec.CIDR)
		if err != nil {
			return cerrors.ErrorValidation{
				ErroredFields: []cerrors.ErroredField{{
					Name:   "IPPool.Spec.CIDR",
					Reason: "IPPool CIDR must be a valid subnet",
					Value:  pool.Spec.CIDR,
				}},
			}
		}

		if ipAddr.Version() == 4 {
			pool.Spec.BlockSize = 26
		} else {
			pool.Spec.BlockSize = 122
		}
	}

	// Default the nodeSelector if it wasn't previously set.
	if pool.Spec.NodeSelector == "" {
		pool.Spec.NodeSelector = "all()"
	}
	return nil
}

// Watch returns a watch.Interface that watches the IPPools that match the
// supplied options.
func (r ipPools) Watch(ctx context.Context, opts options.ListOptions) (watch.Interface, error) {
	return r.client.resources.Watch(ctx, opts, apiv3.KindIPPool, nil)
}

// validateAndSetDefaults validates IPPool fields and sets default values that are
// not assigned.
// The old pool will be unassigned for a Create.
func (r ipPools) validateAndSetDefaults(ctx context.Context, new, old *apiv3.IPPool) error {
	errFields := []cerrors.ErroredField{}

	// Spec.CIDR field must not be empty.
	if new.Spec.CIDR == "" {
		return cerrors.ErrorValidation{
			ErroredFields: []cerrors.ErroredField{{
				Name:   "IPPool.Spec.CIDR",
				Reason: "IPPool CIDR must be specified",
			}},
		}
	}

	// Make sure the CIDR is parsable.
	ipAddr, cidr, err := cnet.ParseCIDR(new.Spec.CIDR)
	if err != nil {
		return cerrors.ErrorValidation{
			ErroredFields: []cerrors.ErroredField{{
				Name:   "IPPool.Spec.CIDR",
				Reason: "IPPool CIDR must be a valid subnet",
				Value:  new.Spec.CIDR,
			}},
		}
	}

	// Normalize the CIDR before persisting.
	new.Spec.CIDR = cidr.String()

	// If a nodeSelector is not specified, then this IP pool selects all nodes.
	if new.Spec.NodeSelector == "" {
		new.Spec.NodeSelector = "all()"
	}

	// If there was a previous pool then this must be an Update, validate that the
	// CIDR has not changed.  Since we are using normalized CIDRs we can just do a
	// simple string comparison.
	if old != nil && old.Spec.CIDR != new.Spec.CIDR {
		errFields = append(errFields, cerrors.ErroredField{
			Name:   "IPPool.Spec.CIDR",
			Reason: "IPPool CIDR cannot be modified",
			Value:  new.Spec.CIDR,
		})
	}

	// Default the blockSize
	if new.Spec.BlockSize == 0 {
		if ipAddr.Version() == 4 {
			new.Spec.BlockSize = 26
		} else {
			new.Spec.BlockSize = 122
		}
	}

	// Check that the blockSize hasn't changed since updates are not supported.
	if old != nil && old.Spec.BlockSize != new.Spec.BlockSize {
		errFields = append(errFields, cerrors.ErroredField{
			Name:   "IPPool.Spec.BlockSize",
			Reason: "IPPool BlockSize cannot be modified",
			Value:  new.Spec.BlockSize,
		})
	}

	if ipAddr.Version() == 4 {
		if new.Spec.BlockSize > 32 || new.Spec.BlockSize < 20 {
			errFields = append(errFields, cerrors.ErroredField{
				Name:   "IPPool.Spec.BlockSize",
				Reason: "IPv4 block size must be between 20 and 32",
				Value:  new.Spec.BlockSize,
			})

		}
	} else {
		if new.Spec.BlockSize > 128 || new.Spec.BlockSize < 116 {
			errFields = append(errFields, cerrors.ErroredField{
				Name:   "IPPool.Spec.BlockSize",
				Reason: "IPv6 block size must be between 116 and 128",
				Value:  new.Spec.BlockSize,
			})
		}
	}

	// The Calico IPAM places restrictions on the minimum IP pool size.  If
	// the ippool is enabled, check that the pool is at least the minimum size.
	if !new.Spec.Disabled {
		ones, _ := cidr.Mask.Size()
		log.Debugf("Pool CIDR: %s, mask: %d, blockSize: %d", cidr.String(), ones, new.Spec.BlockSize)
		if ones > new.Spec.BlockSize {
			errFields = append(errFields, cerrors.ErroredField{
				Name:   "IPPool.Spec.CIDR",
				Reason: "IP pool size is too small for use with Calico IPAM. It must be equal to or greater than the block size.",
				Value:  new.Spec.CIDR,
			})
		}
	}

	// If there was no previous pool then this must be a Create.  Check that the CIDR
	// does not overlap with any other pool CIDRs.
	if old == nil {
		allPools, err := r.List(ctx, options.ListOptions{})
		if err != nil {
			return err
		}

		for _, otherPool := range allPools.Items {
			// It's possible that Create is called for a pre-existing pool, so skip our own
			// pool and let the generic processing handle the pre-existing resource error case.
			if otherPool.Name == new.Name {
				continue
			}
			_, otherCIDR, err := cnet.ParseCIDR(otherPool.Spec.CIDR)
			if err != nil {
				log.WithField("Name", otherPool.Name).WithError(err).Error("IPPool is configured with an invalid CIDR")
				continue
			}
			if otherCIDR.IsNetOverlap(cidr.IPNet) {
				errFields = append(errFields, cerrors.ErroredField{
					Name:   "IPPool.Spec.CIDR",
					Reason: fmt.Sprintf("IPPool(%s) CIDR overlaps with IPPool(%s) CIDR %s", new.Name, otherPool.Name, otherPool.Spec.CIDR),
					Value:  new.Spec.CIDR,
				})
			}
		}
	}

	// Make sure IPIPMode is defaulted to "Never".
	if len(new.Spec.IPIPMode) == 0 {
		new.Spec.IPIPMode = apiv3.IPIPModeNever
	}

	// Make sure VXLANMode is defaulted to "Never".
	if len(new.Spec.VXLANMode) == 0 {
		new.Spec.VXLANMode = apiv3.VXLANModeNever
	}

	// Make sure only one of VXLAN and IPIP is enabled.
	if new.Spec.VXLANMode != apiv3.VXLANModeNever && new.Spec.IPIPMode != apiv3.IPIPModeNever {
		errFields = append(errFields, cerrors.ErroredField{
			Name:   "IPPool.Spec.VXLANMode",
			Reason: "Cannot enable both VXLAN and IPIP on the same IPPool",
			Value:  new.Spec.VXLANMode,
		})
	}

	// IPIP cannot be enabled for IPv6.
	if cidr.Version() == 6 && new.Spec.IPIPMode != apiv3.IPIPModeNever {
		errFields = append(errFields, cerrors.ErroredField{
			Name:   "IPPool.Spec.IPIPMode",
			Reason: "IPIP is not supported on an IPv6 IP pool",
			Value:  new.Spec.IPIPMode,
		})
	}

	// The Calico CIDR should be strictly masked
	log.Debugf("IPPool CIDR: %s, Masked IP: %d", new.Spec.CIDR, cidr.IP)
	if cidr.IP.String() != ipAddr.String() {
		errFields = append(errFields, cerrors.ErroredField{
			Name:   "IPPool.Spec.CIDR",
			Reason: "IPPool CIDR is not strictly masked",
			Value:  new.Spec.CIDR,
		})
	}

	// IPv4 link local subnet.
	ipv4LinkLocalNet := net.IPNet{
		IP:   net.ParseIP("169.254.0.0"),
		Mask: net.CIDRMask(16, 32),
	}
	// IPv6 link local subnet.
	ipv6LinkLocalNet := net.IPNet{
		IP:   net.ParseIP("fe80::"),
		Mask: net.CIDRMask(10, 128),
	}

	// IP Pool CIDR cannot overlap with IPv4 or IPv6 link local address range.
	if cidr.Version() == 4 && cidr.IsNetOverlap(ipv4LinkLocalNet) {
		errFields = append(errFields, cerrors.ErroredField{
			Name:   "IPPool.Spec.CIDR",
			Reason: "IPPool CIDR overlaps with IPv4 Link Local range 169.254.0.0/16",
			Value:  new.Spec.CIDR,
		})
	}

	if cidr.Version() == 6 && cidr.IsNetOverlap(ipv6LinkLocalNet) {
		errFields = append(errFields, cerrors.ErroredField{
			Name:   "IPPool.Spec.CIDR",
			Reason: "IPPool CIDR overlaps with IPv6 Link Local range fe80::/10",
			Value:  new.Spec.CIDR,
		})
	}

	// Return the errors if we have one or more validation errors.
	if len(errFields) > 0 {
		return cerrors.ErrorValidation{
			ErroredFields: errFields,
		}
	}

	return nil
}

// maybeEnableIPIP enables global IPIP if a default setting is not already configured
// and the pool has IPIP enabled.
func (c ipPools) maybeEnableIPIP(ctx context.Context, pool *apiv3.IPPool) error {
	if pool.Spec.IPIPMode == apiv3.IPIPModeNever {
		log.Debug("IPIP is not enabled for this pool - no need to check global setting")
		return nil
	}

	var err error
	ipEnabled := true
	for i := 0; i < maxApplyRetries; i++ {
		log.WithField("Retry", i).Debug("Checking global IPIP setting")
		res, err := c.client.FelixConfigurations().Get(ctx, "default", options.GetOptions{})
		if _, ok := err.(cerrors.ErrorResourceDoesNotExist); !ok && err != nil {
			log.WithError(err).Debug("Error getting current FelixConfiguration resource")
			return err
		}

		if res == nil {
			log.Debug("Global FelixConfiguration does not exist - creating")
			res = apiv3.NewFelixConfiguration()
			res.Name = "default"
		} else if res.Spec.IPIPEnabled != nil {
			// A value for the default config is set so leave unchanged.  It may be set to false,
			// so log the actual value - but we shouldn't update it if someone has explicitly
			// disabled it globally.
			log.WithField("IPIPEnabled", res.Spec.IPIPEnabled).Debug("Global IPIPEnabled setting is already configured")
			return nil
		}

		// Enable IpInIp and do the Create or Update.
		res.Spec.IPIPEnabled = &ipEnabled
		if res.ResourceVersion == "" {
			res, err = c.client.FelixConfigurations().Create(ctx, res, options.SetOptions{})
			if _, ok := err.(cerrors.ErrorResourceAlreadyExists); ok {
				log.Debug("FelixConfiguration already exists - retry update")
				continue
			}
		} else {
			res, err = c.client.FelixConfigurations().Update(ctx, res, options.SetOptions{})
			if _, ok := err.(cerrors.ErrorResourceUpdateConflict); ok {
				log.Debug("FelixConfiguration update conflict - retry update")
				continue
			}
		}

		if err == nil {
			log.Debug("FelixConfiguration updated successfully")
			return nil
		}

		log.WithError(err).Debug("Error updating FelixConfiguration to enable IPIP")
		return err
	}

	// Return the error from the final Update.
	log.WithError(err).Info("Too many conflict failures attempting to update FelixConfiguration to enable IPIP")
	return err
}

// maybeEnableVXLAN enables global VXLAN if a default setting is not already configured
// and the pool has VXLAN enabled.
func (c ipPools) maybeEnableVXLAN(ctx context.Context, pool *apiv3.IPPool) error {
	if pool.Spec.VXLANMode == apiv3.VXLANModeNever {
		log.Debug("VXLAN is not enabled for this pool - no need to check global setting")
		return nil
	}

	var err error
	ipEnabled := true
	for i := 0; i < maxApplyRetries; i++ {
		log.WithField("Retry", i).Debug("Checking global VXLAN setting")
		res, err := c.client.FelixConfigurations().Get(ctx, "default", options.GetOptions{})
		if _, ok := err.(cerrors.ErrorResourceDoesNotExist); !ok && err != nil {
			log.WithError(err).Debug("Error getting current FelixConfiguration resource")
			return err
		}

		if res == nil {
			log.Debug("Global FelixConfiguration does not exist - creating")
			res = apiv3.NewFelixConfiguration()
			res.Name = "default"
		} else if res.Spec.VXLANEnabled != nil {
			// A value for the default config is set so leave unchanged.  It may be set to false,
			// so log the actual value - but we shouldn't update it if someone has explicitly
			// disabled it globally.
			log.WithField("VXLANEnabled", res.Spec.VXLANEnabled).Debug("Global VXLANEnabled setting is already configured")
			return nil
		}

		// Enable IpInIp and do the Create or Update.
		res.Spec.VXLANEnabled = &ipEnabled
		if res.ResourceVersion == "" {
			res, err = c.client.FelixConfigurations().Create(ctx, res, options.SetOptions{})
			if _, ok := err.(cerrors.ErrorResourceAlreadyExists); ok {
				log.Debug("FelixConfiguration already exists - retry update")
				continue
			}
		} else {
			res, err = c.client.FelixConfigurations().Update(ctx, res, options.SetOptions{})
			if _, ok := err.(cerrors.ErrorResourceUpdateConflict); ok {
				log.Debug("FelixConfiguration update conflict - retry update")
				continue
			}
		}

		if err == nil {
			log.Debug("FelixConfiguration updated successfully")
			return nil
		}

		log.WithError(err).Debug("Error updating FelixConfiguration to enable VXLAN")
		return err
	}

	// Return the error from the final Update.
	log.WithError(err).Info("Too many conflict failures attempting to update FelixConfiguration to enable VXLAN")
	return err
}
