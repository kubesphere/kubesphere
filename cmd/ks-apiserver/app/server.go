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
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/apiserver/servicemesh/tracing"
	"kubesphere.io/kubesphere/pkg/filter"
	"kubesphere.io/kubesphere/pkg/informers"
	logging "kubesphere.io/kubesphere/pkg/models/log"
	"kubesphere.io/kubesphere/pkg/server"
	apiserverconfig "kubesphere.io/kubesphere/pkg/server/config"
	"kubesphere.io/kubesphere/pkg/simple/client"
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
			err := apiserverconfig.Load()
			if err != nil {
				return err
			}

			err = Complete(s)
			if err != nil {
				return err
			}

			if errs := s.Validate(); len(errs) != 0 {
				return utilerrors.NewAggregate(errs)
			}

			return Run(s, signals.SetupSignalHandler())
		},
	}

	fs := cmd.Flags()
	namedFlagSets := s.Flags()

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

	initializeESClientConfig()

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

func initializeESClientConfig() {

	// List all outputs
	outputs, err := logging.GetFluentbitOutputFromConfigMap()
	if err != nil {
		klog.Errorln(err)
		return
	}

	// Iterate the outputs to get elasticsearch configs
	for _, output := range outputs {
		if configs := logging.ParseEsOutputParams(output.Parameters); configs != nil {
			configs.WriteESConfigs()
			return
		}
	}
}

//
func CreateAPIServer(s *options.ServerRunOptions) error {
	var err error

	container := runtime.Container
	container.DoNotRecover(false)
	container.Filter(filter.Logging)
	container.RecoverHandler(server.LogStackOnRecover)

	// install config api
	apiserverconfig.InstallAPI(container)

	for _, webservice := range container.RegisteredWebServices() {
		for _, route := range webservice.Routes() {
			klog.V(0).Info(route.Method, route.Path)
		}
	}

	if s.GenericServerRunOptions.InsecurePort != 0 {
		err = http.ListenAndServe(fmt.Sprintf("%s:%d", s.GenericServerRunOptions.BindAddress, s.GenericServerRunOptions.InsecurePort), container)
		if err != nil {
			klog.Infof("Server listening on %d.", s.GenericServerRunOptions.InsecurePort)
		}
	}

	if s.GenericServerRunOptions.SecurePort != 0 && len(s.GenericServerRunOptions.TlsCertFile) > 0 && len(s.GenericServerRunOptions.TlsPrivateKey) > 0 {
		klog.Infof("Server listening on %d.", s.GenericServerRunOptions.SecurePort)
		err = http.ListenAndServeTLS(fmt.Sprintf("%s:%d", s.GenericServerRunOptions.BindAddress, s.GenericServerRunOptions.SecurePort), s.GenericServerRunOptions.TlsCertFile, s.GenericServerRunOptions.TlsPrivateKey, container)
	}

	return err
}

func CreateClientSet(conf *apiserverconfig.Config, stopCh <-chan struct{}) error {
	csop := &client.ClientSetOptions{}

	csop.SetDevopsOptions(conf.DevopsOptions).
		SetKubernetesOptions(conf.KubernetesOptions).
		SetMySQLOptions(conf.MySQLOptions)

	client.NewClientSetFactory(csop, stopCh)

	return nil
}

func WaitForResourceSync(stopCh <-chan struct{}) error {

	//apis.AddToScheme(scheme.Scheme)

	informerFactory := informers.SharedInformerFactory()

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

		{Group: "batch", Version: "v1", Resource: "jobs"},
		{Group: "batch", Version: "v1beta1", Resource: "cronjobs"},

		{Group: "extensions", Version: "v1beta1", Resource: "ingresses"},

		{Group: "autoscaling", Version: "v2beta2", Resource: "horizontalpodautoscalers"},
	}

	for _, gvr := range k8sGVRs {
		_, err := informerFactory.ForResource(gvr)
		if err != nil {
			klog.Errorf("cannot create informer for %s", gvr)
			return err
		}
	}

	informerFactory.Start(stopCh)
	informerFactory.WaitForCacheSync(stopCh)

	s2iInformerFactory := informers.S2iSharedInformerFactory()

	s2iGVRs := []schema.GroupVersionResource{
		{Group: "devops.kubesphere.io", Version: "v1alpha1", Resource: "s2ibuildertemplates"},
		{Group: "devops.kubesphere.io", Version: "v1alpha1", Resource: "s2iruns"},
		{Group: "devops.kubesphere.io", Version: "v1alpha1", Resource: "s2ibuilders"},
	}

	for _, gvr := range s2iGVRs {
		_, err := s2iInformerFactory.ForResource(gvr)
		if err != nil {
			return err
		}
	}

	s2iInformerFactory.Start(stopCh)
	s2iInformerFactory.WaitForCacheSync(stopCh)

	ksInformerFactory := informers.KsSharedInformerFactory()

	ksGVRs := []schema.GroupVersionResource{
		{Group: "tenant.kubesphere.io", Version: "v1alpha1", Resource: "workspaces"},
		{Group: "devops.kubesphere.io", Version: "v1alpha1", Resource: "s2ibinaries"},

		{Group: "servicemesh.kubesphere.io", Version: "v1alpha2", Resource: "strategies"},
		{Group: "servicemesh.kubesphere.io", Version: "v1alpha2", Resource: "servicepolicies"},
	}

	for _, gvr := range ksGVRs {
		_, err := ksInformerFactory.ForResource(gvr)
		if err != nil {
			return err
		}
	}

	ksInformerFactory.Start(stopCh)
	ksInformerFactory.WaitForCacheSync(stopCh)

	return nil

}

// apply server run options to configuration
func Complete(s *options.ServerRunOptions) error {

	// loading configuration file
	conf := apiserverconfig.Get()

	conf.Apply(&apiserverconfig.Config{
		MySQLOptions:       s.MySQLOptions,
		DevopsOptions:      s.DevopsOptions,
		SonarQubeOptions:   s.SonarQubeOptions,
		KubernetesOptions:  s.KubernetesOptions,
		ServiceMeshOptions: s.ServiceMeshOptions,
		MonitoringOptions:  s.MonitoringOptions,
		LdapOptions:        s.LdapOptions,
		RedisOptions:       s.RedisOptions,
		S3Options:          s.S3Options,
		OpenPitrixOptions:  s.OpenPitrixOptions,
	})

	s = &options.ServerRunOptions{
		GenericServerRunOptions: s.GenericServerRunOptions,
		KubernetesOptions:       conf.KubernetesOptions,
		DevopsOptions:           conf.DevopsOptions,
		SonarQubeOptions:        conf.SonarQubeOptions,
		ServiceMeshOptions:      conf.ServiceMeshOptions,
		MySQLOptions:            conf.MySQLOptions,
		MonitoringOptions:       conf.MonitoringOptions,
		LdapOptions:             conf.LdapOptions,
		RedisOptions:            conf.RedisOptions,
		S3Options:               conf.S3Options,
		OpenPitrixOptions:       conf.OpenPitrixOptions,
	}

	return nil
}
