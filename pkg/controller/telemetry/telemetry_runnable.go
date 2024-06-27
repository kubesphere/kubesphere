/*
Copyright 2024 The KubeSphere Authors.

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

package telemetry

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"
	"kubesphere.io/kubesphere/pkg/controller/options"
)

var _ manager.LeaderElectionRunnable = &Runnable{}
var _ manager.Runnable = &Runnable{}
var _ kscontroller.Controller = &Runnable{}

const (
	runnableName = "telemetry"
	// defaultPeriod for collect data
	defaultPeriod = time.Hour * 24
)

type Runnable struct {
	*options.TelemetryOptions
}

func (r *Runnable) Name() string {
	return runnableName
}

func (r *Runnable) SetupWithManager(mgr *kscontroller.Manager) error {
	if mgr.TelemetryOptions == nil || mgr.TelemetryOptions.KSCloudURL == "" {
		klog.V(4).Infof("telemetry runnable is disabled")
		return nil
	}
	r.TelemetryOptions = mgr.TelemetryOptions
	if r.TelemetryOptions.Period == nil {
		r.TelemetryOptions.Period = ptr.To[time.Duration](defaultPeriod)
	}
	return mgr.Add(r)
}

func (r *Runnable) Enabled(clusterRole string) bool {
	return strings.EqualFold(clusterRole, string(clusterv1alpha1.ClusterRoleHost))
}

func (r *Runnable) Start(ctx context.Context) error {
	t := time.NewTicker(*r.Period)
	for {
		select {
		case <-t.C:
			var args = []string{
				"--url", r.KSCloudURL,
			}

			cmd := exec.CommandContext(ctx, "telemetry", args...)
			if _, err := cmd.CombinedOutput(); err != nil {
				klog.Errorf("failed to exec command for telemetry %v", err)
			}
		case <-ctx.Done():
			t.Stop()
			return nil
		}
	}
}

func (r *Runnable) NeedLeaderElection() bool {
	return true
}
