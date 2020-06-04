/*

 Copyright 2020 The KubeSphere Authors.

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
package capability

import (
	"context"
	"errors"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/keepalive"
	"kubesphere.io/kubesphere/pkg/apis/storage/v1alpha1"
	"net"
	"net/url"
	"time"
)

const (
	dialDuration    = time.Second * 5
	requestDuration = time.Second * 10
)

func csiCapability(csiAddress string) (*v1alpha1.StorageClassCapabilitySpec, error) {
	csiConn, err := connect(csiAddress)
	if err != nil {
		return nil, err
	}
	defer func() { _ = csiConn.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), requestDuration)
	defer cancel()

	spec := &v1alpha1.StorageClassCapabilitySpec{}
	err = addPluginCapabilities(ctx, csiConn, spec)
	if err != nil {
		return nil, err
	}
	err = addControllerCapabilities(ctx, csiConn, spec)
	if err != nil {
		return nil, err
	}
	err = addNodeCapabilities(ctx, csiConn, spec)
	if err != nil {
		return nil, err
	}
	return spec, nil

}

func addPluginCapabilities(ctx context.Context, conn *grpc.ClientConn, spec *v1alpha1.StorageClassCapabilitySpec) error {
	identityClient := csi.NewIdentityClient(conn)
	pluginCapabilitiesResponse, err := identityClient.GetPluginCapabilities(ctx, &csi.GetPluginCapabilitiesRequest{})
	if err != nil {
		return err
	}

	for _, capability := range pluginCapabilitiesResponse.GetCapabilities() {
		if capability == nil {
			continue
		}
		if capability.GetService().GetType() == csi.PluginCapability_Service_VOLUME_ACCESSIBILITY_CONSTRAINTS {
			spec.Features.Topology = true
		}
		volumeExpansion := capability.GetVolumeExpansion()
		if volumeExpansion != nil {
			switch volumeExpansion.GetType() {
			case csi.PluginCapability_VolumeExpansion_ONLINE:
				spec.Features.Volume.Expand = v1alpha1.ExpandModeOnline
			case csi.PluginCapability_VolumeExpansion_OFFLINE:
				spec.Features.Volume.Expand = v1alpha1.ExpandModeOffline
			}
		}
	}
	return nil
}

func addControllerCapabilities(ctx context.Context, conn *grpc.ClientConn, spec *v1alpha1.StorageClassCapabilitySpec) error {
	controllerClient := csi.NewControllerClient(conn)
	controllerCapabilitiesResponse, err := controllerClient.ControllerGetCapabilities(ctx, &csi.ControllerGetCapabilitiesRequest{})
	if err != nil {
		return err
	}
	for _, capability := range controllerCapabilitiesResponse.GetCapabilities() {
		switch capability.GetRpc().GetType() {
		case csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME:
			spec.Features.Volume.Create = true
		case csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME:
			spec.Features.Volume.Attach = true
		case csi.ControllerServiceCapability_RPC_LIST_VOLUMES:
			spec.Features.Volume.List = true
		case csi.ControllerServiceCapability_RPC_CLONE_VOLUME:
			spec.Features.Volume.Clone = true
		case csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT:
			spec.Features.Snapshot.Create = true
		case csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS:
			spec.Features.Snapshot.List = true
		}
	}
	return nil
}

func addNodeCapabilities(ctx context.Context, conn *grpc.ClientConn, spec *v1alpha1.StorageClassCapabilitySpec) error {
	nodeClient := csi.NewNodeClient(conn)
	controllerCapabilitiesResponse, err := nodeClient.NodeGetCapabilities(ctx, &csi.NodeGetCapabilitiesRequest{})
	if err != nil {
		return err
	}
	for _, capability := range controllerCapabilitiesResponse.GetCapabilities() {
		switch capability.GetRpc().GetType() {
		case csi.NodeServiceCapability_RPC_GET_VOLUME_STATS:
			spec.Features.Volume.Stats = true
		}
	}
	return nil
}

// Connect address by GRPC
func connect(address string) (*grpc.ClientConn, error) {
	dialOptions := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	u, err := url.Parse(address)
	if err == nil && (!u.IsAbs() || u.Scheme == "unix") {
		dialOptions = append(dialOptions,
			grpc.WithDialer(
				func(addr string, timeout time.Duration) (net.Conn, error) {
					return net.DialTimeout("unix", u.Path, timeout)
				}))
	}
	// This is necessary when connecting via TCP and does not hurt
	// when using Unix domain sockets. It ensures that gRPC detects a dead connection
	// in a timely manner.
	dialOptions = append(dialOptions,
		grpc.WithKeepaliveParams(keepalive.ClientParameters{PermitWithoutStream: true}))

	conn, err := grpc.Dial(address, dialOptions...)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), dialDuration)
	defer cancel()
	for {
		if !conn.WaitForStateChange(ctx, conn.GetState()) {
			return conn, errors.New("connection timed out")
		}
		if conn.GetState() == connectivity.Ready {
			return conn, nil
		}
	}
}
