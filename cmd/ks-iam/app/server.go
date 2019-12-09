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
	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/cmd/ks-iam/app/options"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/kapis"
	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/server"
	apiserverconfig "kubesphere.io/kubesphere/pkg/server/config"
	"kubesphere.io/kubesphere/pkg/server/filter"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"kubesphere.io/kubesphere/pkg/utils/jwtutil"
	"kubesphere.io/kubesphere/pkg/utils/signals"
	"kubesphere.io/kubesphere/pkg/utils/term"
	"net/http"
)

func NewAPIServerCommand() *cobra.Command {
	s := options.NewServerRunOptions()

	cmd := &cobra.Command{
		Use: "ks-iam",
		Long: `The KubeSphere account server validates and configures data
for the api objects. The API Server services REST operations and provides the frontend to the
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

func Run(s *options.ServerRunOptions, stopChan <-chan struct{}) error {
	csop := client.NewClientSetOptions()
	csop.SetKubernetesOptions(s.KubernetesOptions).
		SetLdapOptions(s.LdapOptions).
		SetRedisOptions(s.RedisOptions).
		SetMySQLOptions(s.MySQLOptions)

	client.NewClientSetFactory(csop, stopChan)

	waitForResourceSync(stopChan)

	err := iam.Init(s.AdminEmail, s.AdminPassword, s.AuthRateLimit, s.TokenIdleTimeout, s.EnableMultiLogin, s.GenerateKubeConfig)

	jwtutil.Setup(s.JWTSecret)

	if err != nil {
		return err
	}

	container := runtime.Container
	container.Filter(filter.Logging)
	container.DoNotRecover(false)
	container.RecoverHandler(server.LogStackOnRecover)

	kapis.InstallAuthorizationAPIs(container)

	if s.GenericServerRunOptions.InsecurePort != 0 {
		klog.Infof("Server listening on %s:%d ", s.GenericServerRunOptions.BindAddress, s.GenericServerRunOptions.InsecurePort)
		err = http.ListenAndServe(fmt.Sprintf("%s:%d", s.GenericServerRunOptions.BindAddress, s.GenericServerRunOptions.InsecurePort), container)
	}

	if s.GenericServerRunOptions.SecurePort != 0 && len(s.GenericServerRunOptions.TlsCertFile) > 0 && len(s.GenericServerRunOptions.TlsPrivateKey) > 0 {
		klog.Infof("Server listening on %s:%d", s.GenericServerRunOptions.BindAddress, s.GenericServerRunOptions.SecurePort)
		err = http.ListenAndServeTLS(fmt.Sprintf("%s:%d", s.GenericServerRunOptions.BindAddress, s.GenericServerRunOptions.SecurePort), s.GenericServerRunOptions.TlsCertFile, s.GenericServerRunOptions.TlsPrivateKey, container)
	}

	return err
}

func Complete(s *options.ServerRunOptions) error {
	conf := apiserverconfig.Get()

	conf.Apply(&apiserverconfig.Config{
		KubernetesOptions: s.KubernetesOptions,
		LdapOptions:       s.LdapOptions,
		RedisOptions:      s.RedisOptions,
		MySQLOptions:      s.MySQLOptions,
	})

	s.KubernetesOptions = conf.KubernetesOptions
	s.LdapOptions = conf.LdapOptions
	s.RedisOptions = conf.RedisOptions
	s.MySQLOptions = conf.MySQLOptions

	return nil
}

func waitForResourceSync(stopCh <-chan struct{}) {

	informerFactory := informers.SharedInformerFactory()
	informerFactory.Rbac().V1().Roles().Lister()
	informerFactory.Rbac().V1().RoleBindings().Lister()
	informerFactory.Rbac().V1().ClusterRoles().Lister()
	informerFactory.Rbac().V1().ClusterRoleBindings().Lister()

	informerFactory.Core().V1().Namespaces().Lister()

	informerFactory.Start(stopCh)
	informerFactory.WaitForCacheSync(stopCh)

	ksInformerFactory := informers.KsSharedInformerFactory()
	ksInformerFactory.Tenant().V1alpha1().Workspaces().Lister()

	ksInformerFactory.Start(stopCh)
	ksInformerFactory.WaitForCacheSync(stopCh)
}
