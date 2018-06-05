/*
Copyright 2014 The Kubernetes Authors.

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

// Package app implements a server that runs a set of active
// components.  This includes replication controllers, service endpoints and
// nodes.
//
package app

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	cacheddiscovery "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/informers"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	certutil "k8s.io/client-go/util/cert"
	genericcontrollermanager "k8s.io/kubernetes/cmd/controller-manager/app"
	"k8s.io/kubernetes/cmd/kube-controller-manager/app/config"
	"k8s.io/kubernetes/cmd/kube-controller-manager/app/options"
	"k8s.io/kubernetes/pkg/apis/componentconfig"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/controller"
	serviceaccountcontroller "k8s.io/kubernetes/pkg/controller/serviceaccount"
	"k8s.io/kubernetes/pkg/serviceaccount"
	"k8s.io/kubernetes/pkg/util/configz"
	utilflag "k8s.io/kubernetes/pkg/util/flag"
	"k8s.io/kubernetes/pkg/version"
	"k8s.io/kubernetes/pkg/version/verflag"
)

const (
	// Jitter used when starting controller managers
	ControllerStartJitter = 1.0
)

type ControllerLoopMode int

const (
	IncludeCloudLoops ControllerLoopMode = iota
	ExternalLoops
)

// NewControllerManagerCommand creates a *cobra.Command object with default parameters
func NewControllerManagerCommand() *cobra.Command {
	s := options.NewKubeControllerManagerOptions()
	cmd := &cobra.Command{
		Use: "kube-controller-manager",
		Long: `The Kubernetes controller manager is a daemon that embeds
the core control loops shipped with Kubernetes. In applications of robotics and
automation, a control loop is a non-terminating loop that regulates the state of
the system. In Kubernetes, a controller is a control loop that watches the shared
state of the cluster through the apiserver and makes changes attempting to move the
current state towards the desired state. Examples of controllers that ship with
Kubernetes today are the replication controller, endpoints controller, namespace
controller, and serviceaccounts controller.`,
		Run: func(cmd *cobra.Command, args []string) {
			verflag.PrintAndExitIfRequested()
			utilflag.PrintFlags(cmd.Flags())

			c, err := s.Config(KnownControllers(), ControllersDisabledByDefault.List())
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}

			if err := Run(c.Complete()); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}
	s.AddFlags(cmd.Flags(), KnownControllers(), ControllersDisabledByDefault.List())

	return cmd
}

// ResyncPeriod returns a function which generates a duration each time it is
// invoked; this is so that multiple controllers don't get into lock-step and all
// hammer the apiserver with list requests simultaneously.
func ResyncPeriod(c *config.CompletedConfig) func() time.Duration {
	return func() time.Duration {
		factor := rand.Float64() + 1
		return time.Duration(float64(c.Generic.ComponentConfig.GenericComponent.MinResyncPeriod.Nanoseconds()) * factor)
	}
}

// Run runs the KubeControllerManagerOptions.  This should never exit.
func Run(c *config.CompletedConfig) error {
	// To help debugging, immediately log version
	glog.Infof("Version: %+v", version.Get())

	if cfgz, err := configz.New("componentconfig"); err == nil {
		cfgz.Set(c.Generic.ComponentConfig)
	} else {
		glog.Errorf("unable to register configz: %c", err)
	}

	// Start the controller manager HTTP server
	stopCh := make(chan struct{})
	if c.Generic.SecureServing != nil {
		handler := genericcontrollermanager.NewBaseHandler(&c.Generic)
		handler = genericcontrollermanager.BuildHandlerChain(handler, &c.Generic)
		if err := c.Generic.SecureServing.Serve(handler, 0, stopCh); err != nil {
			return err
		}
	}
	if c.Generic.InsecureServing != nil {
		handler := genericcontrollermanager.NewBaseHandler(&c.Generic)
		handler = genericcontrollermanager.BuildHandlerChain(handler, &c.Generic)
		if err := c.Generic.InsecureServing.Serve(handler, 0, stopCh); err != nil {
			return err
		}
	}

	run := func(stop <-chan struct{}) {
		rootClientBuilder := controller.SimpleControllerClientBuilder{
			ClientConfig: c.Generic.Kubeconfig,
		}
		var clientBuilder controller.ControllerClientBuilder
		if c.Generic.ComponentConfig.KubeCloudShared.UseServiceAccountCredentials {
			if len(c.Generic.ComponentConfig.KubeCloudShared.ServiceAccountKeyFile) == 0 {
				// It'c possible another controller process is creating the tokens for us.
				// If one isn't, we'll timeout and exit when our client builder is unable to create the tokens.
				glog.Warningf("--use-service-account-credentials was specified without providing a --service-account-private-key-file")
			}
			clientBuilder = controller.SAControllerClientBuilder{
				ClientConfig:         restclient.AnonymousClientConfig(c.Generic.Kubeconfig),
				CoreClient:           c.Generic.Client.CoreV1(),
				AuthenticationClient: c.Generic.Client.AuthenticationV1(),
				Namespace:            "kube-system",
			}
		} else {
			clientBuilder = rootClientBuilder
		}
		ctx, err := CreateControllerContext(c, rootClientBuilder, clientBuilder, stop)
		if err != nil {
			glog.Fatalf("error building controller context: %v", err)
		}
		saTokenControllerInitFunc := serviceAccountTokenControllerStarter{rootClientBuilder: rootClientBuilder}.startServiceAccountTokenController

		if err := StartControllers(ctx, saTokenControllerInitFunc, NewControllerInitializers(ctx.LoopMode)); err != nil {
			glog.Fatalf("error starting controllers: %v", err)
		}

		ctx.InformerFactory.Start(ctx.Stop)
		close(ctx.InformersStarted)

		select {}
	}

	if !c.Generic.ComponentConfig.GenericComponent.LeaderElection.LeaderElect {
		run(wait.NeverStop)
		panic("unreachable")
	}

	id, err := os.Hostname()
	if err != nil {
		return err
	}

	// add a uniquifier so that two processes on the same host don't accidentally both become active
	id = id + "_" + string(uuid.NewUUID())
	rl, err := resourcelock.New(c.Generic.ComponentConfig.GenericComponent.LeaderElection.ResourceLock,
		"kube-system",
		"kube-controller-manager",
		c.Generic.LeaderElectionClient.CoreV1(),
		resourcelock.ResourceLockConfig{
			Identity:      id,
			EventRecorder: c.Generic.EventRecorder,
		})
	if err != nil {
		glog.Fatalf("error creating lock: %v", err)
	}

	leaderelection.RunOrDie(leaderelection.LeaderElectionConfig{
		Lock:          rl,
		LeaseDuration: c.Generic.ComponentConfig.GenericComponent.LeaderElection.LeaseDuration.Duration,
		RenewDeadline: c.Generic.ComponentConfig.GenericComponent.LeaderElection.RenewDeadline.Duration,
		RetryPeriod:   c.Generic.ComponentConfig.GenericComponent.LeaderElection.RetryPeriod.Duration,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: run,
			OnStoppedLeading: func() {
				glog.Fatalf("leaderelection lost")
			},
		},
	})
	panic("unreachable")
}

type ControllerContext struct {
	// ClientBuilder will provide a client for this controller to use
	ClientBuilder controller.ControllerClientBuilder

	// InformerFactory gives access to informers for the controller.
	InformerFactory informers.SharedInformerFactory

	// ComponentConfig provides access to init options for a given controller
	ComponentConfig componentconfig.KubeControllerManagerConfiguration

	// DeferredDiscoveryRESTMapper is a RESTMapper that will defer
	// initialization of the RESTMapper until the first mapping is
	// requested.
	RESTMapper *discovery.DeferredDiscoveryRESTMapper

	// AvailableResources is a map listing currently available resources
	AvailableResources map[schema.GroupVersionResource]bool

	// Cloud is the cloud provider interface for the controllers to use.
	// It must be initialized and ready to use.
	Cloud cloudprovider.Interface

	// Control for which control loops to be run
	// IncludeCloudLoops is for a kube-controller-manager running all loops
	// ExternalLoops is for a kube-controller-manager running with a cloud-controller-manager
	LoopMode ControllerLoopMode

	// Stop is the stop channel
	Stop <-chan struct{}

	// InformersStarted is closed after all of the controllers have been initialized and are running.  After this point it is safe,
	// for an individual controller to start the shared informers. Before it is closed, they should not.
	InformersStarted chan struct{}

	// ResyncPeriod generates a duration each time it is invoked; this is so that
	// multiple controllers don't get into lock-step and all hammer the apiserver
	// with list requests simultaneously.
	ResyncPeriod func() time.Duration
}

func (c ControllerContext) IsControllerEnabled(name string) bool {
	return IsControllerEnabled(name, ControllersDisabledByDefault, c.ComponentConfig.Controllers...)
}

func IsControllerEnabled(name string, disabledByDefaultControllers sets.String, controllers ...string) bool {
	hasStar := false
	for _, ctrl := range controllers {
		if ctrl == name {
			return true
		}
		if ctrl == "-"+name {
			return false
		}
		if ctrl == "*" {
			hasStar = true
		}
	}
	// if we get here, there was no explicit choice
	if !hasStar {
		// nothing on by default
		return false
	}
	if disabledByDefaultControllers.Has(name) {
		return false
	}

	return true
}

// InitFunc is used to launch a particular controller.  It may run additional "should I activate checks".
// Any error returned will cause the controller process to `Fatal`
// The bool indicates whether the controller was enabled.
type InitFunc func(ctx ControllerContext) (bool, error)

func KnownControllers() []string {
	ret := sets.StringKeySet(NewControllerInitializers(IncludeCloudLoops))

	// add "special" controllers that aren't initialized normally.  These controllers cannot be initialized
	// using a normal function.  The only known special case is the SA token controller which *must* be started
	// first to ensure that the SA tokens for future controllers will exist.  Think very carefully before adding
	// to this list.
	ret.Insert(
		saTokenControllerName,
	)

	return ret.List()
}

var ControllersDisabledByDefault = sets.NewString(
	"bootstrapsigner",
	"tokencleaner",
)

const (
	saTokenControllerName = "serviceaccount-token"
)

// NewControllerInitializers is a public map of named controller groups (you can start more than one in an init func)
// paired to their InitFunc.  This allows for structured downstream composition and subdivision.
func NewControllerInitializers(loopMode ControllerLoopMode) map[string]InitFunc {
	controllers := map[string]InitFunc{}
	controllers["endpoint"] = startEndpointController
	controllers["replicationcontroller"] = startReplicationController
	controllers["podgc"] = startPodGCController
	controllers["resourcequota"] = startResourceQuotaController
	controllers["namespace"] = startNamespaceController
	controllers["serviceaccount"] = startServiceAccountController
	controllers["garbagecollector"] = startGarbageCollectorController
	controllers["daemonset"] = startDaemonSetController
	controllers["job"] = startJobController
	controllers["deployment"] = startDeploymentController
	controllers["replicaset"] = startReplicaSetController
	controllers["horizontalpodautoscaling"] = startHPAController
	controllers["disruption"] = startDisruptionController
	controllers["statefulset"] = startStatefulSetController
	controllers["cronjob"] = startCronJobController
	controllers["csrsigning"] = startCSRSigningController
	controllers["csrapproving"] = startCSRApprovingController
	controllers["csrcleaner"] = startCSRCleanerController
	controllers["ttl"] = startTTLController
	controllers["bootstrapsigner"] = startBootstrapSignerController
	controllers["tokencleaner"] = startTokenCleanerController
	if loopMode == IncludeCloudLoops {
		controllers["service"] = startServiceController
		controllers["nodeipam"] = startNodeIpamController
		controllers["route"] = startRouteController
		// TODO: volume controller into the IncludeCloudLoops only set.
		// TODO: Separate cluster in cloud check from node lifecycle controller.
	}
	controllers["nodelifecycle"] = startNodeLifecycleController
	controllers["persistentvolume-binder"] = startPersistentVolumeBinderController
	controllers["attachdetach"] = startAttachDetachController
	controllers["persistentvolume-expander"] = startVolumeExpandController
	controllers["clusterrole-aggregation"] = startClusterRoleAggregrationController
	controllers["pvc-protection"] = startPVCProtectionController
	controllers["pv-protection"] = startPVProtectionController

	return controllers
}

// TODO: In general, any controller checking this needs to be dynamic so
//  users don't have to restart their controller manager if they change the apiserver.
// Until we get there, the structure here needs to be exposed for the construction of a proper ControllerContext.
func GetAvailableResources(clientBuilder controller.ControllerClientBuilder) (map[schema.GroupVersionResource]bool, error) {
	client := clientBuilder.ClientOrDie("controller-discovery")
	discoveryClient := client.Discovery()
	resourceMap, err := discoveryClient.ServerResources()
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to get all supported resources from server: %v", err))
	}
	if len(resourceMap) == 0 {
		return nil, fmt.Errorf("unable to get any supported resources from server")
	}

	allResources := map[schema.GroupVersionResource]bool{}
	for _, apiResourceList := range resourceMap {
		version, err := schema.ParseGroupVersion(apiResourceList.GroupVersion)
		if err != nil {
			return nil, err
		}
		for _, apiResource := range apiResourceList.APIResources {
			allResources[version.WithResource(apiResource.Name)] = true
		}
	}

	return allResources, nil
}

// CreateControllerContext creates a context struct containing references to resources needed by the
// controllers such as the cloud provider and clientBuilder. rootClientBuilder is only used for
// the shared-informers client and token controller.
func CreateControllerContext(s *config.CompletedConfig, rootClientBuilder, clientBuilder controller.ControllerClientBuilder, stop <-chan struct{}) (ControllerContext, error) {
	versionedClient := rootClientBuilder.ClientOrDie("shared-informers")
	sharedInformers := informers.NewSharedInformerFactory(versionedClient, ResyncPeriod(s)())

	// If apiserver is not running we should wait for some time and fail only then. This is particularly
	// important when we start apiserver and controller manager at the same time.
	if err := genericcontrollermanager.WaitForAPIServer(versionedClient, 10*time.Second); err != nil {
		return ControllerContext{}, fmt.Errorf("failed to wait for apiserver being healthy: %v", err)
	}

	// Use a discovery client capable of being refreshed.
	discoveryClient := rootClientBuilder.ClientOrDie("controller-discovery")
	cachedClient := cacheddiscovery.NewMemCacheClient(discoveryClient.Discovery())
	restMapper := discovery.NewDeferredDiscoveryRESTMapper(cachedClient)
	go wait.Until(func() {
		restMapper.Reset()
	}, 30*time.Second, stop)

	availableResources, err := GetAvailableResources(rootClientBuilder)
	if err != nil {
		return ControllerContext{}, err
	}

	cloud, loopMode, err := createCloudProvider(s.Generic.ComponentConfig.CloudProvider.Name, s.Generic.ComponentConfig.ExternalCloudVolumePlugin,
		s.Generic.ComponentConfig.CloudProvider.CloudConfigFile, s.Generic.ComponentConfig.KubeCloudShared.AllowUntaggedCloud, sharedInformers)
	if err != nil {
		return ControllerContext{}, err
	}

	ctx := ControllerContext{
		ClientBuilder:      clientBuilder,
		InformerFactory:    sharedInformers,
		ComponentConfig:    s.Generic.ComponentConfig,
		RESTMapper:         restMapper,
		AvailableResources: availableResources,
		Cloud:              cloud,
		LoopMode:           loopMode,
		Stop:               stop,
		InformersStarted:   make(chan struct{}),
		ResyncPeriod:       ResyncPeriod(s),
	}
	return ctx, nil
}

func StartControllers(ctx ControllerContext, startSATokenController InitFunc, controllers map[string]InitFunc) error {
	// Always start the SA token controller first using a full-power client, since it needs to mint tokens for the rest
	// If this fails, just return here and fail since other controllers won't be able to get credentials.
	if _, err := startSATokenController(ctx); err != nil {
		return err
	}

	// Initialize the cloud provider with a reference to the clientBuilder only after token controller
	// has started in case the cloud provider uses the client builder.
	if ctx.Cloud != nil {
		ctx.Cloud.Initialize(ctx.ClientBuilder)
	}

	for controllerName, initFn := range controllers {
		if !ctx.IsControllerEnabled(controllerName) {
			glog.Warningf("%q is disabled", controllerName)
			continue
		}

		time.Sleep(wait.Jitter(ctx.ComponentConfig.GenericComponent.ControllerStartInterval.Duration, ControllerStartJitter))

		glog.V(1).Infof("Starting %q", controllerName)
		started, err := initFn(ctx)
		if err != nil {
			glog.Errorf("Error starting %q", controllerName)
			return err
		}
		if !started {
			glog.Warningf("Skipping %q", controllerName)
			continue
		}
		glog.Infof("Started %q", controllerName)
	}

	return nil
}

// serviceAccountTokenControllerStarter is special because it must run first to set up permissions for other controllers.
// It cannot use the "normal" client builder, so it tracks its own. It must also avoid being included in the "normal"
// init map so that it can always run first.
type serviceAccountTokenControllerStarter struct {
	rootClientBuilder controller.ControllerClientBuilder
}

func (c serviceAccountTokenControllerStarter) startServiceAccountTokenController(ctx ControllerContext) (bool, error) {
	if !ctx.IsControllerEnabled(saTokenControllerName) {
		glog.Warningf("%q is disabled", saTokenControllerName)
		return false, nil
	}

	if len(ctx.ComponentConfig.KubeCloudShared.ServiceAccountKeyFile) == 0 {
		glog.Warningf("%q is disabled because there is no private key", saTokenControllerName)
		return false, nil
	}
	privateKey, err := certutil.PrivateKeyFromFile(ctx.ComponentConfig.KubeCloudShared.ServiceAccountKeyFile)
	if err != nil {
		return true, fmt.Errorf("error reading key for service account token controller: %v", err)
	}

	var rootCA []byte
	if ctx.ComponentConfig.SAController.RootCAFile != "" {
		rootCA, err = ioutil.ReadFile(ctx.ComponentConfig.SAController.RootCAFile)
		if err != nil {
			return true, fmt.Errorf("error reading root-ca-file at %s: %v", ctx.ComponentConfig.SAController.RootCAFile, err)
		}
		if _, err := certutil.ParseCertsPEM(rootCA); err != nil {
			return true, fmt.Errorf("error parsing root-ca-file at %s: %v", ctx.ComponentConfig.SAController.RootCAFile, err)
		}
	} else {
		rootCA = c.rootClientBuilder.ConfigOrDie("tokens-controller").CAData
	}

	controller, err := serviceaccountcontroller.NewTokensController(
		ctx.InformerFactory.Core().V1().ServiceAccounts(),
		ctx.InformerFactory.Core().V1().Secrets(),
		c.rootClientBuilder.ClientOrDie("tokens-controller"),
		serviceaccountcontroller.TokensControllerOptions{
			TokenGenerator: serviceaccount.JWTTokenGenerator(serviceaccount.LegacyIssuer, privateKey),
			RootCA:         rootCA,
		},
	)
	if err != nil {
		return true, fmt.Errorf("error creating Tokens controller: %v", err)
	}
	go controller.Run(int(ctx.ComponentConfig.SAController.ConcurrentSATokenSyncs), ctx.Stop)

	// start the first set of informers now so that other controllers can start
	ctx.InformerFactory.Start(ctx.Stop)

	return true, nil
}
