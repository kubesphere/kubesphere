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
	"github.com/json-iterator/go"
	kconfig "github.com/kiali/kiali/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"kubesphere.io/kubesphere/cmd/ks-apiserver/app/options"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/apiserver/servicemesh/tracing"
	"kubesphere.io/kubesphere/pkg/filter"
	"kubesphere.io/kubesphere/pkg/informers"
	logging "kubesphere.io/kubesphere/pkg/models/log"
	"kubesphere.io/kubesphere/pkg/signals"
	"kubesphere.io/kubesphere/pkg/simple/client/admin_jenkins"
	"kubesphere.io/kubesphere/pkg/simple/client/devops_mysql"
	"log"
	"net/http"
)

var jsonIter = jsoniter.ConfigCompatibleWithStandardLibrary

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

	for _, webservice := range container.RegisteredWebServices() {
		for _, route := range webservice.Routes() {
			log.Println(route.Method, route.Path)
		}
	}

	initializeAdminJenkins()
	initializeDevOpsDatabase()
	initializeESClientConfig()
	initializeServicemeshConfig(s)

	if s.GenericServerRunOptions.InsecurePort != 0 {
		log.Printf("Server listening on %d.", s.GenericServerRunOptions.InsecurePort)
		err = http.ListenAndServe(fmt.Sprintf("%s:%d", s.GenericServerRunOptions.BindAddress, s.GenericServerRunOptions.InsecurePort), container)
	}

	if s.GenericServerRunOptions.SecurePort != 0 && len(s.GenericServerRunOptions.TlsCertFile) > 0 && len(s.GenericServerRunOptions.TlsPrivateKey) > 0 {
		log.Printf("Server listening on %d.", s.GenericServerRunOptions.SecurePort)
		err = http.ListenAndServeTLS(fmt.Sprintf("%s:%d", s.GenericServerRunOptions.BindAddress, s.GenericServerRunOptions.SecurePort), s.GenericServerRunOptions.TlsCertFile, s.GenericServerRunOptions.TlsPrivateKey, container)
	}

	return err
}

func initializeAdminJenkins() {
	admin_jenkins.Client()
}

func initializeDevOpsDatabase() {
	devops_mysql.OpenDatabase()
}

func initializeServicemeshConfig(s *options.ServerRunOptions) {
	// Initialize kiali config
	config := kconfig.NewConfig()

	tracing.JaegerQueryUrl = s.JaegerQueryServiceUrl

	// Exclude system namespaces
	config.API.Namespaces.Exclude = []string{"istio-system", "kubesphere*", "kube*"}
	config.InCluster = true

	// Set default prometheus service url
	config.ExternalServices.PrometheusServiceURL = s.ServicemeshPrometheusServiceUrl
	config.ExternalServices.PrometheusCustomMetricsURL = config.ExternalServices.PrometheusServiceURL

	// Set istio pilot discovery service url
	config.ExternalServices.Istio.UrlServiceVersion = s.IstioPilotServiceURL

	kconfig.Set(config)
}

func initializeESClientConfig() {

	// List all outputs
	outputs, err := logging.GetFluentbitOutputFromConfigMap()
	if err != nil {
		glog.Errorln(err)
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
	informerFactory.Apps().V1().ReplicaSets().Lister()

	informerFactory.Batch().V1().Jobs().Lister()
	informerFactory.Batch().V1beta1().CronJobs().Lister()
	informerFactory.Extensions().V1beta1().Ingresses().Lister()

	informerFactory.Start(stopChan)
	informerFactory.WaitForCacheSync(stopChan)

	s2iInformerFactory := informers.S2iSharedInformerFactory()
	s2iInformerFactory.Devops().V1alpha1().S2iBuilderTemplates().Lister()
	s2iInformerFactory.Devops().V1alpha1().S2iRuns().Lister()
	s2iInformerFactory.Devops().V1alpha1().S2iBuilders().Lister()

	s2iInformerFactory.Start(stopChan)
	s2iInformerFactory.WaitForCacheSync(stopChan)

	ksInformerFactory := informers.KsSharedInformerFactory()
	ksInformerFactory.Tenant().V1alpha1().Workspaces().Lister()

	ksInformerFactory.Start(stopChan)
	ksInformerFactory.WaitForCacheSync(stopChan)

	log.Println("resources sync success")
}
