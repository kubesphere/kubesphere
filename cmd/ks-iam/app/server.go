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
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"kubesphere.io/kubesphere/cmd/ks-iam/app/options"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/filter"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/signals"
	"log"
	"net/http"
)

func NewAPIServerCommand() *cobra.Command {
	s := options.NewServerRunOptions()

	cmd := &cobra.Command{
		Use: "ks-iam",
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

	err = iam.DatabaseInit()

	if err != nil {
		return err
	}

	waitForResourceSync()

	container := runtime.Container
	container.Filter(filter.Logging)

	if s.GenericServerRunOptions.InsecurePort != 0 {
		err = http.ListenAndServe(fmt.Sprintf("%s:%d", s.GenericServerRunOptions.BindAddress, s.GenericServerRunOptions.InsecurePort), container)
	}

	if s.GenericServerRunOptions.SecurePort != 0 && len(s.GenericServerRunOptions.TlsCertFile) > 0 && len(s.GenericServerRunOptions.TlsPrivateKey) > 0 {
		err = http.ListenAndServeTLS(fmt.Sprintf("%s:%d", s.GenericServerRunOptions.BindAddress, s.GenericServerRunOptions.SecurePort), s.GenericServerRunOptions.TlsCertFile, s.GenericServerRunOptions.TlsPrivateKey, container)
	}

	return err
}

func waitForResourceSync() {
	stopChan := signals.SetupSignalHandler()

	informerFactory := informers.SharedInformerFactory()
	informerFactory.Rbac().V1().Roles().Lister()
	informerFactory.Rbac().V1().RoleBindings().Lister()
	informerFactory.Rbac().V1().ClusterRoles().Lister()
	informerFactory.Rbac().V1().ClusterRoleBindings().Lister()

	informerFactory.Core().V1().Namespaces().Lister()

	informerFactory.Start(stopChan)
	informerFactory.WaitForCacheSync(stopChan)
	log.Println("resources sync success")
}
