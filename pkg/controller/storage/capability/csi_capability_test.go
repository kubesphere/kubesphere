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
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/storage/v1alpha1"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var DefaultControllerRPCType = []csi.ControllerServiceCapability_RPC_Type{
	csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
	csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
	csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
	csi.ControllerServiceCapability_RPC_CLONE_VOLUME,
}

var DefaultNodeRPCType = []csi.NodeServiceCapability_RPC_Type{
	csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
	csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
	csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
}

var DefaultPluginCapability = []*csi.PluginCapability{
	{
		Type: &csi.PluginCapability_Service_{
			Service: &csi.PluginCapability_Service{
				Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
			},
		},
	},
	{
		Type: &csi.PluginCapability_VolumeExpansion_{
			VolumeExpansion: &csi.PluginCapability_VolumeExpansion{
				Type: csi.PluginCapability_VolumeExpansion_OFFLINE,
			},
		},
	},
}

type fakeCSIServer struct {
	csi.UnimplementedIdentityServer
	csi.UnimplementedControllerServer
	csi.UnimplementedNodeServer
	network string
	address string
	server  *grpc.Server
}

func newTestCSIServer(port int) (csiServer *fakeCSIServer, address string) {
	if runtime.GOOS == "windows" {
		address = fmt.Sprintf("localhost:%d", +port)
		csiServer = newFakeCSIServer("tcp", address)
	} else {
		address = filepath.Join(os.TempDir(), "csi.sock"+rand.String(4))
		csiServer = newFakeCSIServer("unix", address)
		address = "unix://" + address
	}
	return csiServer, address
}

func newFakeCSIServer(network, address string) *fakeCSIServer {
	return &fakeCSIServer{
		network: network,
		address: address,
	}
}

func (f *fakeCSIServer) run() {
	listener, err := net.Listen(f.network, f.address)
	if err != nil {
		klog.Error("fake CSI server listen failed, ", err)
		return
	}
	server := grpc.NewServer()
	csi.RegisterIdentityServer(server, f)
	csi.RegisterControllerServer(server, f)
	csi.RegisterNodeServer(server, f)
	go func() {
		err = server.Serve(listener)
		if err != nil && !strings.Contains(err.Error(), "stopped") {
			klog.Error("fake CSI server serve failed, ", err)
		}
	}()
	f.server = server
}

func (f *fakeCSIServer) stop() {
	if f.server != nil {
		f.server.Stop()
	}
}

func (*fakeCSIServer) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	return &csi.GetPluginCapabilitiesResponse{Capabilities: DefaultPluginCapability}, nil
}

func (*fakeCSIServer) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	var capabilities []*csi.ControllerServiceCapability
	for _, rpcType := range DefaultControllerRPCType {
		capability := &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: rpcType,
				},
			},
		}
		capabilities = append(capabilities, capability)
	}
	return &csi.ControllerGetCapabilitiesResponse{Capabilities: capabilities}, nil
}

func (*fakeCSIServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	var capabilities []*csi.NodeServiceCapability
	for _, rpcType := range DefaultNodeRPCType {
		capability := &csi.NodeServiceCapability{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: rpcType,
				},
			},
		}
		capabilities = append(capabilities, capability)
	}
	return &csi.NodeGetCapabilitiesResponse{Capabilities: capabilities}, nil
}

func Test_CSICapability(t *testing.T) {
	fakeCSIServer, address := newTestCSIServer(30087)
	fakeCSIServer.run()
	defer fakeCSIServer.stop()

	specGot, err := csiCapability(address)
	if err != nil {
		t.Error(err)
	}

	specExpected := newStorageClassCapabilitySpec()
	if diff := cmp.Diff(specGot, specExpected); diff != "" {
		t.Errorf("%T differ (-got, +want): %s", specExpected, diff)
	}
}

func newStorageClassCapabilitySpec() *v1alpha1.StorageClassCapabilitySpec {
	return &v1alpha1.StorageClassCapabilitySpec{
		Features: v1alpha1.CapabilityFeatures{
			Topology: false,
			Volume: v1alpha1.VolumeFeature{
				Create: true,
				Attach: false,
				List:   false,
				Clone:  true,
				Stats:  true,
				Expand: v1alpha1.ExpandModeOffline,
			},
			Snapshot: v1alpha1.SnapshotFeature{
				Create: true,
				List:   false,
			},
		},
	}
}
