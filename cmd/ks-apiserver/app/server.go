package app

import (
	"fmt"
	"github.com/emicklei/go-restful-openapi"
	"github.com/spf13/cobra"
	"kubesphere.io/kubesphere/cmd/ks-apiserver/app/options"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/filter"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/signals"
	"log"
	"net/http"
)

func NewAPIServerCommand() *cobra.Command {
	s := options.NewServerRunOptions()

	cmd := &cobra.Command{
		Use: "ks-apiserver",
		Long: `The KubeSphere API server validates and configures data
for the api objects. The API Server services REST operations and provides the frontend to the
cluster's shared state through which all other components interact.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			//s.AddFlags(cmd.Flags())

			return Run(s)
		},
	}

	s.AddFlags(cmd.Flags())

	return cmd
}

func Run(s *options.ServerRunOptions) error {

	var err error

	stopChan := signals.SetupSignalHandler()
	informers.SharedInformerFactory().Start(stopChan)
	informers.SharedInformerFactory().WaitForCacheSync(stopChan)
	log.Println("resources sync success")

	container := runtime.Container
	container.Filter(filter.Logging)

	if len(s.KubeConfig) > 0 {
		client.KubeConfigFile = s.KubeConfig
	}

	if s.ApiDoc {
		config := restfulspec.Config{
			WebServices: container.RegisteredWebServices(),
			APIPath:     "/apidoc.json",
		}
		container.Add(restfulspec.NewOpenAPIService(config))
	}

	log.Printf("Server listening on %d.", s.InsecurePort)

	if s.InsecurePort != 0 {
		err = http.ListenAndServe(fmt.Sprintf("%s:%d", s.BindAddress, s.InsecurePort), container)
	}

	if s.SecurePort != 0 && len(s.TlsCertFile) > 0 && len(s.TlsPrivateKey) > 0 {
		err = http.ListenAndServeTLS(fmt.Sprintf("%s:%d", s.BindAddress, s.SecurePort), s.TlsCertFile, s.TlsPrivateKey, container)
	}

	return err
}
