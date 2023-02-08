/*
Copyright 2018 The Kubernetes Authors.

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

package util

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"

	fedv1b1 "sigs.k8s.io/kubefed/pkg/apis/core/v1beta1"
)

// LeaderElectionConfiguration defines the configuration of leader election
// clients for controller that can run with leader election enabled.
type LeaderElectionConfiguration struct {
	// leaseDuration is the duration that non-leader candidates will wait
	// after observing a leadership renewal until attempting to acquire
	// leadership of a led but unrenewed leader slot. This is effectively the
	// maximum duration that a leader can be stopped before it is replaced
	// by another candidate. This is only applicable if leader election is
	// enabled.
	LeaseDuration time.Duration
	// renewDeadline is the interval between attempts by the acting master to
	// renew a leadership slot before it stops leading. This must be less
	// than or equal to the lease duration. This is only applicable if leader
	// election is enabled.
	RenewDeadline time.Duration
	// retryPeriod is the duration the clients should wait between attempting
	// acquisition and renewal of a leadership. This is only applicable if
	// leader election is enabled.
	RetryPeriod time.Duration
	// resourceLock indicates the resource object type that will be used to lock
	// during leader election cycles.
	ResourceLock fedv1b1.ResourceLockType
}

// KubeFedNamespaces defines the namespace configuration shared by
// most kubefed controllers.
type KubeFedNamespaces struct {
	KubeFedNamespace string
	TargetNamespace  string
}

// ClusterHealthCheckConfig defines the configurable parameters for cluster health check
type ClusterHealthCheckConfig struct {
	Period           time.Duration
	FailureThreshold int64
	SuccessThreshold int64
	Timeout          time.Duration
}

// ControllerConfig defines the configuration common to KubeFed
// controllers.
type ControllerConfig struct {
	KubeFedNamespaces
	KubeConfig                    *restclient.Config
	ClusterAvailableDelay         time.Duration
	ClusterUnavailableDelay       time.Duration
	MinimizeLatency               bool
	CacheSyncTimeout              time.Duration
	MaxConcurrentSyncReconciles   int64
	MaxConcurrentStatusReconciles int64
	SkipAdoptingResources         bool
	RawResourceStatusCollection   bool
}

func (c *ControllerConfig) LimitedScope() bool {
	return c.KubeFedNamespaces.TargetNamespace != metav1.NamespaceAll
}
