/*
Copyright 2022 The KubeSphere Authors.

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

package helm

import (
	"context"
	"time"
)

// Executor is used to manage a helm release, you can install/uninstall and upgrade a chart
// or get the status and manifest data of the release, etc.
type Executor interface {
	// Install installs the specified chart and returns the name of the Job that executed the task.
	Install(ctx context.Context, chartName, chartData, values string) (string, error)
	// Upgrade upgrades the specified chart and returns the name of the Job that executed the task.
	Upgrade(ctx context.Context, chartName, chartData, values string) (string, error)
	// Uninstall is used to uninstall the specified chart and returns the name of the Job that executed the task.
	Uninstall(ctx context.Context) (string, error)
	// Manifest returns the manifest data for this release.
	Manifest() (string, error)
	// IsReleaseReady checks if the helm release is ready.
	IsReleaseReady(timeout time.Duration) (bool, error)
}

type executor struct {
}

type Option func(*executor)

// NewExecutor generates a new Executor instance with the following parameters:
//   - kubeConfig: kube config data of the target cluster
//   - namespace: the namespace of the helm release
//   - releaseName: the helm release name
//   - options: functions to set optional parameters
func NewExecutor(kubeConfig, namespace, releaseName string, options ...Option) (Executor, error) {
	return &executor{}, nil
}

// Install installs the specified chart, returns the name of the Job that executed the task.
func (e *executor) Install(ctx context.Context, chartName, chartData, values string) (string, error) {
	return "", nil
}

// Upgrade upgrades the specified chart, returns the name of the Job that executed the task.
func (e *executor) Upgrade(ctx context.Context, chartName, chartData, values string) (string, error) {
	return "", nil
}

// Uninstall uninstalls the specified chart, returns the name of the Job that executed the task.
func (e *executor) Uninstall(ctx context.Context) (string, error) {
	return "", nil
}

// Manifest returns the manifest data for this release.
func (e *executor) Manifest() (string, error) {
	return "", nil
}

// IsReleaseReady checks if the helm release is ready.
func (e *executor) IsReleaseReady(timeout time.Duration) (bool, error) {
	return false, nil
}
