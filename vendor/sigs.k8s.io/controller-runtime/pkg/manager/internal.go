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

package manager

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/internal/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/recorder"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const (
	// Values taken from: https://github.com/kubernetes/apiserver/blob/master/pkg/apis/config/v1alpha1/defaults.go
	defaultLeaseDuration = 15 * time.Second
	defaultRenewDeadline = 10 * time.Second
	defaultRetryPeriod   = 2 * time.Second
)

var log = logf.RuntimeLog.WithName("manager")

type controllerManager struct {
	// config is the rest.config used to talk to the apiserver.  Required.
	config *rest.Config

	// scheme is the scheme injected into Controllers, EventHandlers, Sources and Predicates.  Defaults
	// to scheme.scheme.
	scheme *runtime.Scheme

	// leaderElectionRunnables is the set of Controllers that the controllerManager injects deps into and Starts.
	// These Runnables are managed by lead election.
	leaderElectionRunnables []Runnable
	// nonLeaderElectionRunnables is the set of webhook servers that the controllerManager injects deps into and Starts.
	// These Runnables will not be blocked by lead election.
	nonLeaderElectionRunnables []Runnable

	cache cache.Cache

	// TODO(directxman12): Provide an escape hatch to get individual indexers
	// client is the client injected into Controllers (and EventHandlers, Sources and Predicates).
	client client.Client

	// apiReader is the reader that will make requests to the api server and not the cache.
	apiReader client.Reader

	// fieldIndexes knows how to add field indexes over the Cache used by this controller,
	// which can later be consumed via field selectors from the injected client.
	fieldIndexes client.FieldIndexer

	// recorderProvider is used to generate event recorders that will be injected into Controllers
	// (and EventHandlers, Sources and Predicates).
	recorderProvider recorder.Provider

	// resourceLock forms the basis for leader election
	resourceLock resourcelock.Interface

	// mapper is used to map resources to kind, and map kind and version.
	mapper meta.RESTMapper

	// metricsListener is used to serve prometheus metrics
	metricsListener net.Listener

	mu            sync.Mutex
	started       bool
	startedLeader bool
	errChan       chan error

	// internalStop is the stop channel *actually* used by everything involved
	// with the manager as a stop channel, so that we can pass a stop channel
	// to things that need it off the bat (like the Channel source).  It can
	// be closed via `internalStopper` (by being the same underlying channel).
	internalStop <-chan struct{}

	// internalStopper is the write side of the internal stop channel, allowing us to close it.
	// It and `internalStop` should point to the same channel.
	internalStopper chan<- struct{}

	startCache func(stop <-chan struct{}) error

	// port is the port that the webhook server serves at.
	port int
	// host is the hostname that the webhook server binds to.
	host string

	webhookServer *webhook.Server

	// leaseDuration is the duration that non-leader candidates will
	// wait to force acquire leadership.
	leaseDuration time.Duration
	// renewDeadline is the duration that the acting master will retry
	// refreshing leadership before giving up.
	renewDeadline time.Duration
	// retryPeriod is the duration the LeaderElector clients should wait
	// between tries of actions.
	retryPeriod time.Duration
}

// Add sets dependencies on i, and adds it to the list of Runnables to start.
func (cm *controllerManager) Add(r Runnable) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Set dependencies on the object
	if err := cm.SetFields(r); err != nil {
		return err
	}

	var shouldStart bool

	// Add the runnable to the leader election or the non-leaderelection list
	if leRunnable, ok := r.(LeaderElectionRunnable); ok && !leRunnable.NeedLeaderElection() {
		shouldStart = cm.started
		cm.nonLeaderElectionRunnables = append(cm.nonLeaderElectionRunnables, r)
	} else {
		shouldStart = cm.startedLeader
		cm.leaderElectionRunnables = append(cm.leaderElectionRunnables, r)
	}

	if shouldStart {
		// If already started, start the controller
		go func() {
			cm.errChan <- r.Start(cm.internalStop)
		}()
	}

	return nil
}

func (cm *controllerManager) SetFields(i interface{}) error {
	if _, err := inject.ConfigInto(cm.config, i); err != nil {
		return err
	}
	if _, err := inject.ClientInto(cm.client, i); err != nil {
		return err
	}
	if _, err := inject.APIReaderInto(cm.apiReader, i); err != nil {
		return err
	}
	if _, err := inject.SchemeInto(cm.scheme, i); err != nil {
		return err
	}
	if _, err := inject.CacheInto(cm.cache, i); err != nil {
		return err
	}
	if _, err := inject.InjectorInto(cm.SetFields, i); err != nil {
		return err
	}
	if _, err := inject.StopChannelInto(cm.internalStop, i); err != nil {
		return err
	}
	if _, err := inject.MapperInto(cm.mapper, i); err != nil {
		return err
	}
	return nil
}

func (cm *controllerManager) GetConfig() *rest.Config {
	return cm.config
}

func (cm *controllerManager) GetClient() client.Client {
	return cm.client
}

func (cm *controllerManager) GetScheme() *runtime.Scheme {
	return cm.scheme
}

func (cm *controllerManager) GetFieldIndexer() client.FieldIndexer {
	return cm.fieldIndexes
}

func (cm *controllerManager) GetCache() cache.Cache {
	return cm.cache
}

func (cm *controllerManager) GetEventRecorderFor(name string) record.EventRecorder {
	return cm.recorderProvider.GetEventRecorderFor(name)
}

func (cm *controllerManager) GetRESTMapper() meta.RESTMapper {
	return cm.mapper
}

func (cm *controllerManager) GetAPIReader() client.Reader {
	return cm.apiReader
}

func (cm *controllerManager) GetWebhookServer() *webhook.Server {
	if cm.webhookServer == nil {
		cm.webhookServer = &webhook.Server{
			Port: cm.port,
			Host: cm.host,
		}
		if err := cm.Add(cm.webhookServer); err != nil {
			panic("unable to add webhookServer to the controller manager")
		}
	}
	return cm.webhookServer
}

func (cm *controllerManager) serveMetrics(stop <-chan struct{}) {
	var metricsPath = "/metrics"
	handler := promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{
		ErrorHandling: promhttp.HTTPErrorOnError,
	})
	// TODO(JoelSpeed): Use existing Kubernetes machinery for serving metrics
	mux := http.NewServeMux()
	mux.Handle(metricsPath, handler)
	server := http.Server{
		Handler: mux,
	}
	// Run the server
	go func() {
		log.Info("starting metrics server", "path", metricsPath)
		if err := server.Serve(cm.metricsListener); err != nil && err != http.ErrServerClosed {
			cm.errChan <- err
		}
	}()

	// Shutdown the server when stop is closed
	select {
	case <-stop:
		if err := server.Shutdown(context.Background()); err != nil {
			cm.errChan <- err
		}
	}
}

func (cm *controllerManager) Start(stop <-chan struct{}) error {
	// join the passed-in stop channel as an upstream feeding into cm.internalStopper
	defer close(cm.internalStopper)

	// Metrics should be served whether the controller is leader or not.
	// (If we don't serve metrics for non-leaders, prometheus will still scrape
	// the pod but will get a connection refused)
	if cm.metricsListener != nil {
		go cm.serveMetrics(cm.internalStop)
	}

	go cm.startNonLeaderElectionRunnables()

	if cm.resourceLock != nil {
		err := cm.startLeaderElection()
		if err != nil {
			return err
		}
	} else {
		go cm.startLeaderElectionRunnables()
	}

	select {
	case <-stop:
		// We are done
		return nil
	case err := <-cm.errChan:
		// Error starting a controller
		return err
	}
}

func (cm *controllerManager) startNonLeaderElectionRunnables() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.waitForCache()

	// Start the non-leaderelection Runnables after the cache has synced
	for _, c := range cm.nonLeaderElectionRunnables {
		// Controllers block, but we want to return an error if any have an error starting.
		// Write any Start errors to a channel so we can return them
		ctrl := c
		go func() {
			cm.errChan <- ctrl.Start(cm.internalStop)
		}()
	}
}

func (cm *controllerManager) startLeaderElectionRunnables() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.waitForCache()

	// Start the leader election Runnables after the cache has synced
	for _, c := range cm.leaderElectionRunnables {
		// Controllers block, but we want to return an error if any have an error starting.
		// Write any Start errors to a channel so we can return them
		ctrl := c
		go func() {
			cm.errChan <- ctrl.Start(cm.internalStop)
		}()
	}

	cm.startedLeader = true
}

func (cm *controllerManager) waitForCache() {
	if cm.started {
		return
	}

	// Start the Cache. Allow the function to start the cache to be mocked out for testing
	if cm.startCache == nil {
		cm.startCache = cm.cache.Start
	}
	go func() {
		if err := cm.startCache(cm.internalStop); err != nil {
			cm.errChan <- err
		}
	}()

	// Wait for the caches to sync.
	// TODO(community): Check the return value and write a test
	cm.cache.WaitForCacheSync(cm.internalStop)
	cm.started = true
}

func (cm *controllerManager) startLeaderElection() (err error) {
	l, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:          cm.resourceLock,
		LeaseDuration: cm.leaseDuration,
		RenewDeadline: cm.renewDeadline,
		RetryPeriod:   cm.retryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(_ context.Context) {
				cm.startLeaderElectionRunnables()
			},
			OnStoppedLeading: func() {
				// Most implementations of leader election log.Fatal() here.
				// Since Start is wrapped in log.Fatal when called, we can just return
				// an error here which will cause the program to exit.
				cm.errChan <- fmt.Errorf("leader election lost")
			},
		},
	})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-cm.internalStop:
			cancel()
		case <-ctx.Done():
		}
	}()

	// Start the leader elector process
	go l.Run(ctx)
	return nil
}
