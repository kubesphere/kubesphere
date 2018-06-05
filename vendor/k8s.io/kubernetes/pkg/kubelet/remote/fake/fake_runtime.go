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

package fake

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	kubeapi "k8s.io/kubernetes/pkg/kubelet/apis/cri/runtime/v1alpha2"
	apitest "k8s.io/kubernetes/pkg/kubelet/apis/cri/testing"
	"k8s.io/kubernetes/pkg/kubelet/util"
	utilexec "k8s.io/utils/exec"
)

// RemoteRuntime represents a fake remote container runtime.
type RemoteRuntime struct {
	server *grpc.Server
	// Fake runtime service.
	RuntimeService *apitest.FakeRuntimeService
	// Fake image service.
	ImageService *apitest.FakeImageService
}

// NewFakeRemoteRuntime creates a new RemoteRuntime.
func NewFakeRemoteRuntime() *RemoteRuntime {
	fakeRuntimeService := apitest.NewFakeRuntimeService()
	fakeImageService := apitest.NewFakeImageService()

	f := &RemoteRuntime{
		server:         grpc.NewServer(),
		RuntimeService: fakeRuntimeService,
		ImageService:   fakeImageService,
	}
	kubeapi.RegisterRuntimeServiceServer(f.server, f)
	kubeapi.RegisterImageServiceServer(f.server, f)

	return f
}

// Start starts the fake remote runtime.
func (f *RemoteRuntime) Start(endpoint string) error {
	l, err := util.CreateListener(endpoint)
	if err != nil {
		return fmt.Errorf("failed to listen on %q: %v", endpoint, err)
	}

	go f.server.Serve(l)
	return nil
}

// Stop stops the fake remote runtime.
func (f *RemoteRuntime) Stop() {
	f.server.Stop()
}

// Version returns the runtime name, runtime version, and runtime API version.
func (f *RemoteRuntime) Version(ctx context.Context, req *kubeapi.VersionRequest) (*kubeapi.VersionResponse, error) {
	return f.RuntimeService.Version(req.Version)
}

// RunPodSandbox creates and starts a pod-level sandbox. Runtimes must ensure
// the sandbox is in the ready state on success.
func (f *RemoteRuntime) RunPodSandbox(ctx context.Context, req *kubeapi.RunPodSandboxRequest) (*kubeapi.RunPodSandboxResponse, error) {
	sandboxID, err := f.RuntimeService.RunPodSandbox(req.Config)
	if err != nil {
		return nil, err
	}

	return &kubeapi.RunPodSandboxResponse{PodSandboxId: sandboxID}, nil
}

// StopPodSandbox stops any running process that is part of the sandbox and
// reclaims network resources (e.g., IP addresses) allocated to the sandbox.
// If there are any running containers in the sandbox, they must be forcibly
// terminated.
func (f *RemoteRuntime) StopPodSandbox(ctx context.Context, req *kubeapi.StopPodSandboxRequest) (*kubeapi.StopPodSandboxResponse, error) {
	err := f.RuntimeService.StopPodSandbox(req.PodSandboxId)
	if err != nil {
		return nil, err
	}

	return &kubeapi.StopPodSandboxResponse{}, nil
}

// RemovePodSandbox removes the sandbox. If there are any running containers
// in the sandbox, they must be forcibly terminated and removed.
// This call is idempotent, and must not return an error if the sandbox has
// already been removed.
func (f *RemoteRuntime) RemovePodSandbox(ctx context.Context, req *kubeapi.RemovePodSandboxRequest) (*kubeapi.RemovePodSandboxResponse, error) {
	err := f.RuntimeService.StopPodSandbox(req.PodSandboxId)
	if err != nil {
		return nil, err
	}

	return &kubeapi.RemovePodSandboxResponse{}, nil
}

// PodSandboxStatus returns the status of the PodSandbox. If the PodSandbox is not
// present, returns an error.
func (f *RemoteRuntime) PodSandboxStatus(ctx context.Context, req *kubeapi.PodSandboxStatusRequest) (*kubeapi.PodSandboxStatusResponse, error) {
	podStatus, err := f.RuntimeService.PodSandboxStatus(req.PodSandboxId)
	if err != nil {
		return nil, err
	}

	return &kubeapi.PodSandboxStatusResponse{Status: podStatus}, nil
}

// ListPodSandbox returns a list of PodSandboxes.
func (f *RemoteRuntime) ListPodSandbox(ctx context.Context, req *kubeapi.ListPodSandboxRequest) (*kubeapi.ListPodSandboxResponse, error) {
	items, err := f.RuntimeService.ListPodSandbox(req.Filter)
	if err != nil {
		return nil, err
	}

	return &kubeapi.ListPodSandboxResponse{Items: items}, nil
}

// CreateContainer creates a new container in specified PodSandbox
func (f *RemoteRuntime) CreateContainer(ctx context.Context, req *kubeapi.CreateContainerRequest) (*kubeapi.CreateContainerResponse, error) {
	containerID, err := f.RuntimeService.CreateContainer(req.PodSandboxId, req.Config, req.SandboxConfig)
	if err != nil {
		return nil, err
	}

	return &kubeapi.CreateContainerResponse{ContainerId: containerID}, nil
}

// StartContainer starts the container.
func (f *RemoteRuntime) StartContainer(ctx context.Context, req *kubeapi.StartContainerRequest) (*kubeapi.StartContainerResponse, error) {
	err := f.RuntimeService.StartContainer(req.ContainerId)
	if err != nil {
		return nil, err
	}

	return &kubeapi.StartContainerResponse{}, nil
}

// StopContainer stops a running container with a grace period (i.e., timeout).
// This call is idempotent, and must not return an error if the container has
// already been stopped.
func (f *RemoteRuntime) StopContainer(ctx context.Context, req *kubeapi.StopContainerRequest) (*kubeapi.StopContainerResponse, error) {
	err := f.RuntimeService.StopContainer(req.ContainerId, req.Timeout)
	if err != nil {
		return nil, err
	}

	return &kubeapi.StopContainerResponse{}, nil
}

// RemoveContainer removes the container. If the container is running, the
// container must be forcibly removed.
// This call is idempotent, and must not return an error if the container has
// already been removed.
func (f *RemoteRuntime) RemoveContainer(ctx context.Context, req *kubeapi.RemoveContainerRequest) (*kubeapi.RemoveContainerResponse, error) {
	err := f.RuntimeService.RemoveContainer(req.ContainerId)
	if err != nil {
		return nil, err
	}

	return &kubeapi.RemoveContainerResponse{}, nil
}

// ListContainers lists all containers by filters.
func (f *RemoteRuntime) ListContainers(ctx context.Context, req *kubeapi.ListContainersRequest) (*kubeapi.ListContainersResponse, error) {
	items, err := f.RuntimeService.ListContainers(req.Filter)
	if err != nil {
		return nil, err
	}

	return &kubeapi.ListContainersResponse{Containers: items}, nil
}

// ContainerStatus returns status of the container. If the container is not
// present, returns an error.
func (f *RemoteRuntime) ContainerStatus(ctx context.Context, req *kubeapi.ContainerStatusRequest) (*kubeapi.ContainerStatusResponse, error) {
	status, err := f.RuntimeService.ContainerStatus(req.ContainerId)
	if err != nil {
		return nil, err
	}

	return &kubeapi.ContainerStatusResponse{Status: status}, nil
}

// ExecSync runs a command in a container synchronously.
func (f *RemoteRuntime) ExecSync(ctx context.Context, req *kubeapi.ExecSyncRequest) (*kubeapi.ExecSyncResponse, error) {
	var exitCode int32
	stdout, stderr, err := f.RuntimeService.ExecSync(req.ContainerId, req.Cmd, time.Duration(req.Timeout)*time.Second)
	if err != nil {
		exitError, ok := err.(utilexec.ExitError)
		if !ok {
			return nil, err
		}
		exitCode = int32(exitError.ExitStatus())

		return nil, err
	}

	return &kubeapi.ExecSyncResponse{
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: exitCode,
	}, nil
}

// Exec prepares a streaming endpoint to execute a command in the container.
func (f *RemoteRuntime) Exec(ctx context.Context, req *kubeapi.ExecRequest) (*kubeapi.ExecResponse, error) {
	return f.RuntimeService.Exec(req)
}

// Attach prepares a streaming endpoint to attach to a running container.
func (f *RemoteRuntime) Attach(ctx context.Context, req *kubeapi.AttachRequest) (*kubeapi.AttachResponse, error) {
	return f.RuntimeService.Attach(req)
}

// PortForward prepares a streaming endpoint to forward ports from a PodSandbox.
func (f *RemoteRuntime) PortForward(ctx context.Context, req *kubeapi.PortForwardRequest) (*kubeapi.PortForwardResponse, error) {
	return f.RuntimeService.PortForward(req)
}

// ContainerStats returns stats of the container. If the container does not
// exist, the call returns an error.
func (f *RemoteRuntime) ContainerStats(ctx context.Context, req *kubeapi.ContainerStatsRequest) (*kubeapi.ContainerStatsResponse, error) {
	stats, err := f.RuntimeService.ContainerStats(req.ContainerId)
	if err != nil {
		return nil, err
	}

	return &kubeapi.ContainerStatsResponse{Stats: stats}, nil
}

// ListContainerStats returns stats of all running containers.
func (f *RemoteRuntime) ListContainerStats(ctx context.Context, req *kubeapi.ListContainerStatsRequest) (*kubeapi.ListContainerStatsResponse, error) {
	stats, err := f.RuntimeService.ListContainerStats(req.Filter)
	if err != nil {
		return nil, err
	}

	return &kubeapi.ListContainerStatsResponse{Stats: stats}, nil
}

// UpdateRuntimeConfig updates the runtime configuration based on the given request.
func (f *RemoteRuntime) UpdateRuntimeConfig(ctx context.Context, req *kubeapi.UpdateRuntimeConfigRequest) (*kubeapi.UpdateRuntimeConfigResponse, error) {
	err := f.RuntimeService.UpdateRuntimeConfig(req.RuntimeConfig)
	if err != nil {
		return nil, err
	}

	return &kubeapi.UpdateRuntimeConfigResponse{}, nil
}

// Status returns the status of the runtime.
func (f *RemoteRuntime) Status(ctx context.Context, req *kubeapi.StatusRequest) (*kubeapi.StatusResponse, error) {
	status, err := f.RuntimeService.Status()
	if err != nil {
		return nil, err
	}

	return &kubeapi.StatusResponse{Status: status}, nil
}

// UpdateContainerResources updates ContainerConfig of the container.
func (f *RemoteRuntime) UpdateContainerResources(ctx context.Context, req *kubeapi.UpdateContainerResourcesRequest) (*kubeapi.UpdateContainerResourcesResponse, error) {
	err := f.RuntimeService.UpdateContainerResources(req.ContainerId, req.Linux)
	if err != nil {
		return nil, err
	}

	return &kubeapi.UpdateContainerResourcesResponse{}, nil
}

// ReopenContainerLog reopens the container log file.
func (f *RemoteRuntime) ReopenContainerLog(ctx context.Context, req *kubeapi.ReopenContainerLogRequest) (*kubeapi.ReopenContainerLogResponse, error) {
	err := f.RuntimeService.ReopenContainerLog(req.ContainerId)
	if err != nil {
		return nil, err
	}

	return &kubeapi.ReopenContainerLogResponse{}, nil
}
