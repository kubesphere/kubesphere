/*

 Copyright 2019 The KubeSphere Authors.

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
package app

import (
	"fmt"
	kconfig "github.com/kiali/kiali/config"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/cmd/ks-apiserver/app/options"
	"kubesphere.io/kubesphere/pkg/apiserver"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/apiserver/servicemesh/tracing"
	kinformers "kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/kapis"
	"kubesphere.io/kubesphere/pkg/server"
	apiserverconfig "kubesphere.io/kubesphere/pkg/server/config"
	"kubesphere.io/kubesphere/pkg/server/filter"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"kubesphere.io/kubesphere/pkg/utils/signals"
	"kubesphere.io/kubesphere/pkg/utils/term"
	"net/http"
)

func NewAPIServerCommand() *cobra.Command {
	s := options.NewServerRunOptions()

	cmd := &cobra.Command{
		Use: "ks-apiserver",
		Long: `The KubeSphere API server validates and configures data for the api objects. 
The API Server services REST operations and provides the frontend to the
cluster's shared state through which all other components interact.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if errs := s.Validate(); len(errs) != 0 {
				return utilerrors.NewAggregate(errs)
			}

			return Run(s, signals.SetupSignalHandler())
		},
	}

	configOptions := load()

	fs := cmd.Flags()
	namedFlagSets := s.Flags(configOptions)

	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	usageFmt := "Usage:\n  %s\n"
	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStdout(), namedFlagSets, cols)
	})
	return cmd
}

func Run(s *options.ServerRunOptions, stopCh <-chan struct{}) error {

	err := CreateClientSet(apiserverconfig.Get(), stopCh)
	if err != nil {
		return err
	}

	err = WaitForResourceSync(stopCh)
	if err != nil {
		return err
	}

	initializeServicemeshConfig(s)

	err = CreateAPIServer(s)
	if err != nil {
		return err
	}

	return nil
}

func initializeServicemeshConfig(s *options.ServerRunOptions) {
	// Initialize kiali config
	config := kconfig.NewConfig()

	tracing.JaegerQueryUrl = s.ServiceMeshOptions.JaegerQueryHost

	// Exclude system namespaces
	config.API.Namespaces.Exclude = []string{"istio-system", "kubesphere*", "kube*"}
	config.InCluster = true

	// Set default prometheus service url
	config.ExternalServices.PrometheusServiceURL = s.ServiceMeshOptions.ServicemeshPrometheusHost
	config.ExternalServices.PrometheusCustomMetricsURL = config.ExternalServices.PrometheusServiceURL

	// Set istio pilot discovery service url
	config.ExternalServices.Istio.UrlServiceVersion = s.ServiceMeshOptions.IstioPilotHost

	kconfig.Set(config)
}

//
func CreateAPIServer(s *options.ServerRunOptions, stopCh <-chan struct{}) error {
	var err error

	deps := createDeps(s, stopCh)

	apiserver := apiserver.New(deps)

	container := runtime.Container
	container.DoNotRecover(false)
	container.Filter(filter.Logging)
	container.RecoverHandler(server.LogStackOnRecover)

	kapis.InstallAPIs(container)

	// install config api
	apiserverconfig.InstallAPI(container)

	if s.GenericServerRunOptions.InsecurePort != 0 {
		err = http.ListenAndServe(fmt.Sprintf("%s:%d", s.GenericServerRunOptions.BindAddress, s.GenericServerRunOptions.InsecurePort), container)
		if err == nil {
			klog.V(0).Infof("Server listening on insecure port %d.", s.GenericServerRunOptions.InsecurePort)
		}
	}

	if s.GenericServerRunOptions.SecurePort != 0 && len(s.GenericServerRunOptions.TlsCertFile) > 0 && len(s.GenericServerRunOptions.TlsPrivateKey) > 0 {
		err = http.ListenAndServeTLS(fmt.Sprintf("%s:%d", s.GenericServerRunOptions.BindAddress, s.GenericServerRunOptions.SecurePort), s.GenericServerRunOptions.TlsCertFile, s.GenericServerRunOptions.TlsPrivateKey, container)
		if err == nil {
			klog.V(0).Infof("Server listening on secure port %d.", s.GenericServerRunOptions.SecurePort)
		}
	}

	return err
}

func createDeps(s *options.ServerRunOptions, stopCh <-chan struct{}) *apiserver.Dependencies {
	deps := &apiserver.Dependencies{}

	if s.KubernetesOptions == nil || s.KubernetesOptions.KubeConfig == "" {
		klog.Warning("kubeconfig not provided, will use in-cluster config")
	}

	var err error
	deps.KubeClient, err = k8s.NewKubernetesClient(s.KubernetesOptions)
	if err != nil {
		klog.Fatalf("error happened when initializing kubernetes client, %v", err)
	}

	if s.S3Options != nil && s.S3Options.Endpoint != "" {
		deps.S3, err = s3.NewS3Client(s.S3Options)
		if err != nil {
			klog.Fatalf("error initializing s3 client, %v", err)
		}
	}

	if s.OpenPitrixOptions != nil && !s.OpenPitrixOptions.IsEmpty() {
		deps.OpenPitrix, err = openpitrix.NewOpenPitrixClient(s.OpenPitrixOptions)
		if err != nil {
			klog.Fatalf("error happened when initializing openpitrix client, %v", err)
		}
	}

	return deps

}

func WaitForResourceSync(stopCh <-chan struct{}) error {
	klog.V(0).Info("Start cache objects")

	discoveryClient := client.ClientSets().K8s().Discovery()
	apiResourcesList, err := discoveryClient.ServerResources()
	if err != nil {
		return err
	}

	isResourceExists := func(resource schema.GroupVersionResource) bool {
		for _, apiResource := range apiResourcesList {
			if apiResource.GroupVersion == resource.GroupVersion().String() {
				for _, rsc := range apiResource.APIResources {
					if rsc.Name == resource.Resource {
						return true
					}
				}
			}
		}
		return false
	}

	informerFactory := kinformers.NewInformerFactories(client.ClientSets().K8s().Kubernetes(), client.ClientSets().K8s().KubeSphere(), client.ClientSets().K8s().S2i(),
		client.ClientSets().K8s().Application())

	// resources we have to create informer first
	k8sGVRs := []schema.GroupVersionResource{
		{Group: "", Version: "v1", Resource: "namespaces"},
		{Group: "", Version: "v1", Resource: "nodes"},
		{Group: "", Version: "v1", Resource: "resourcequotas"},
		{Group: "", Version: "v1", Resource: "pods"},
		{Group: "", Version: "v1", Resource: "services"},
		{Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
		{Group: "", Version: "v1", Resource: "secrets"},
		{Group: "", Version: "v1", Resource: "configmaps"},

		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"},

		{Group: "apps", Version: "v1", Resource: "deployments"},
		{Group: "apps", Version: "v1", Resource: "daemonsets"},
		{Group: "apps", Version: "v1", Resource: "replicasets"},
		{Group: "apps", Version: "v1", Resource: "statefulsets"},
		{Group: "apps", Version: "v1", Resource: "controllerrevisions"},

		{Group: "storage.k8s.io", Version: "v1", Resource: "storageclasses"},

		{Group: "batch", Version: "v1", Resource: "jobs"},
		{Group: "batch", Version: "v1beta1", Resource: "cronjobs"},

		{Group: "extensions", Version: "v1beta1", Resource: "ingresses"},

		{Group: "autoscaling", Version: "v2beta2", Resource: "horizontalpodautoscalers"},
	}

	for _, gvr := range k8sGVRs {
		if !isResourceExists(gvr) {
			klog.Warningf("resource %s not exists in the cluster", gvr)
		} else {
			_, err := informerFactory.KubernetesSharedInformerFactory().ForResource(gvr)
			if err != nil {
				klog.Errorf("cannot create informer for %s", gvr)
				return err
			}
		}
	}

	informerFactory.KubernetesSharedInformerFactory().Start(stopCh)
	informerFactory.KubernetesSharedInformerFactory().WaitForCacheSync(stopCh)

	s2iInformerFactory := informerFactory.S2iSharedInformerFactory()

	s2iGVRs := []schema.GroupVersionResource{
		{Group: "devops.kubesphere.io", Version: "v1alpha1", Resource: "s2ibuildertemplates"},
		{Group: "devops.kubesphere.io", Version: "v1alpha1", Resource: "s2iruns"},
		{Group: "devops.kubesphere.io", Version: "v1alpha1", Resource: "s2ibuilders"},
	}

	for _, gvr := range s2iGVRs {
		if !isResourceExists(gvr) {
			klog.Warningf("resource %s not exists in the cluster", gvr)
		} else {
			_, err := s2iInformerFactory.ForResource(gvr)
			if err != nil {
				return err
			}
		}
	}

	s2iInformerFactory.Start(stopCh)
	s2iInformerFactory.WaitForCacheSync(stopCh)

	ksInformerFactory := informerFactory.KubeSphereSharedInformerFactory()

	ksGVRs := []schema.GroupVersionResource{
		{Group: "tenant.kubesphere.io", Version: "v1alpha1", Resource: "workspaces"},
		{Group: "devops.kubesphere.io", Version: "v1alpha1", Resource: "s2ibinaries"},

		{Group: "servicemesh.kubesphere.io", Version: "v1alpha2", Resource: "strategies"},
		{Group: "servicemesh.kubesphere.io", Version: "v1alpha2", Resource: "servicepolicies"},
	}

	for _, gvr := range ksGVRs {
		if !isResourceExists(gvr) {
			klog.Warningf("resource %s not exists in the cluster", gvr)
		} else {
			_, err := ksInformerFactory.ForResource(gvr)
			if err != nil {
				return err
			}
		}
	}

	ksInformerFactory.Start(stopCh)
	ksInformerFactory.WaitForCacheSync(stopCh)

	appInformerFactory := informerFactory.ApplicationSharedInformerFactory()

	appGVRs := []schema.GroupVersionResource{
		{Group: "app.k8s.io", Version: "v1beta1", Resource: "applications"},
	}

	for _, gvr := range appGVRs {
		if !isResourceExists(gvr) {
			klog.Warningf("resource %s not exists in the cluster", gvr)
		} else {
			_, err := appInformerFactory.ForResource(gvr)
			if err != nil {
				return err
			}
		}
	}

	appInformerFactory.Start(stopCh)
	appInformerFactory.WaitForCacheSync(stopCh)

	klog.V(0).Info("Finished caching objects")

	return nil

}

// load options from config file
func load() *options.ServerRunOptions {
	conf := apiserverconfig.Get()

	return &options.ServerRunOptions{
		KubernetesOptions:  conf.KubernetesOptions,
		DevopsOptions:      conf.DevopsOptions,
		SonarQubeOptions:   conf.SonarQubeOptions,
		ServiceMeshOptions: conf.ServiceMeshOptions,
		MySQLOptions:       conf.MySQLOptions,
		MonitoringOptions:  conf.MonitoringOptions,
		S3Options:          conf.S3Options,
		OpenPitrixOptions:  conf.OpenPitrixOptions,
		LoggingOptions:     conf.LoggingOptions,
	}
}

func initConfigz() error {
	return nil
}
