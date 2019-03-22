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
	goflag "flag"
	"fmt"
	"github.com/golang/glog"
	kconfig "github.com/kiali/kiali/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"kubesphere.io/kubesphere/cmd/ks-apiserver/app/options"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/filter"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/signals"
	"kubesphere.io/kubesphere/pkg/simple/client/prometheus"
	"log"
	"net/http"
	"net/url"
)

func NewAPIServerCommand() *cobra.Command {
	s := options.NewServerRunOptions()

	cmd := &cobra.Command{
		Use: "ks-apiserver",
		Long: `The KubeSphere API server validates and configures data
for the api objects. The API Server services REST operations and provides the frontend to the
cluster's shared state through which all other components interact.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(s)
		},
	}

	s.AddFlags(cmd.Flags())
	cmd.Flags().AddGoFlagSet(goflag.CommandLine)
	glog.CopyStandardLogTo("INFO")

	return cmd
}

func Run(s *options.ServerRunOptions) error {

	pflag.VisitAll(func(flag *pflag.Flag) {
		log.Printf("FLAG: --%s=%q", flag.Name, flag.Value)
	})

	var err error

	waitForResourceSync()

	container := runtime.Container
	container.Filter(filter.Logging)

	log.Printf("Server listening on %d.", s.GenericServerRunOptions.InsecurePort)

	for _, webservice := range container.RegisteredWebServices() {
		for _, route := range webservice.Routes() {
			log.Printf(route.Path)
		}
	}

	initializeKialiConfig(s)

	if s.GenericServerRunOptions.InsecurePort != 0 {
		err = http.ListenAndServe(fmt.Sprintf("%s:%d", s.GenericServerRunOptions.BindAddress, s.GenericServerRunOptions.InsecurePort), container)
	}

	if s.GenericServerRunOptions.SecurePort != 0 && len(s.GenericServerRunOptions.TlsCertFile) > 0 && len(s.GenericServerRunOptions.TlsPrivateKey) > 0 {
		err = http.ListenAndServeTLS(fmt.Sprintf("%s:%d", s.GenericServerRunOptions.BindAddress, s.GenericServerRunOptions.SecurePort), s.GenericServerRunOptions.TlsCertFile, s.GenericServerRunOptions.TlsPrivateKey, container)
	}

	return err
}

func initializeKialiConfig(s *options.ServerRunOptions) {
	// Initialize kiali config
	config := kconfig.NewConfig()

	// Exclude system namespaces
	config.API.Namespaces.Exclude = []string{"istio-system", "kubesphere*", "kube*"}
	config.InCluster = true

	// Set default prometheus service url
	config.ExternalServices.PrometheusServiceURL = "http://prometheus.kubesphere-monitoring-system.svc:9090"

	// ugly hack to get prometheus service url
	if pflag.Parsed() && pflag.Lookup("prometheus-endpoint") != nil {
		// Set prometheus
		endpoint, err := url.Parse(prometheus.PrometheusAPIEndpoint)
		if err != nil {
			config.ExternalServices.PrometheusServiceURL = endpoint.Path
		}
	}

	config.ExternalServices.PrometheusCustomMetricsURL = config.ExternalServices.PrometheusServiceURL

	// Set istio pilot discovery service url
	config.ExternalServices.Istio.UrlServiceVersion = s.IstioPilotServiceURL

	kconfig.Set(config)
}

func waitForResourceSync() {
	stopChan := signals.SetupSignalHandler()

	informerFactory := informers.SharedInformerFactory()
	informerFactory.Rbac().V1().Roles().Lister()
	informerFactory.Rbac().V1().RoleBindings().Lister()
	informerFactory.Rbac().V1().ClusterRoles().Lister()
	informerFactory.Rbac().V1().ClusterRoleBindings().Lister()

	informerFactory.Storage().V1().StorageClasses().Lister()

	informerFactory.Core().V1().Namespaces().Lister()
	informerFactory.Core().V1().Nodes().Lister()
	informerFactory.Core().V1().ResourceQuotas().Lister()
	informerFactory.Core().V1().Pods().Lister()
	informerFactory.Core().V1().Services().Lister()
	informerFactory.Core().V1().PersistentVolumeClaims().Lister()
	informerFactory.Core().V1().Secrets().Lister()
	informerFactory.Core().V1().ConfigMaps().Lister()

	informerFactory.Apps().V1().ControllerRevisions().Lister()
	informerFactory.Apps().V1().StatefulSets().Lister()
	informerFactory.Apps().V1().Deployments().Lister()
	informerFactory.Apps().V1().DaemonSets().Lister()

	informerFactory.Batch().V1().Jobs().Lister()
	informerFactory.Batch().V1beta1().CronJobs().Lister()

	informerFactory.Start(stopChan)
	informerFactory.WaitForCacheSync(stopChan)
	log.Println("resources sync success")
}
