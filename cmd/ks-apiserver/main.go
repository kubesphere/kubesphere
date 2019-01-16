package main

import (
	goflag "flag"
	"fmt"
	"log"
	"net/http"

	"github.com/golang/glog"

	"kubesphere.io/kubesphere/pkg/apibuilder"
	metrics "kubesphere.io/kubesphere/pkg/apis/metrics/install"
	monitoring "kubesphere.io/kubesphere/pkg/apis/monitoring/install"
	operations "kubesphere.io/kubesphere/pkg/apis/operations/install"
	resources "kubesphere.io/kubesphere/pkg/apis/resources/install"
	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/version"

	"github.com/emicklei/go-restful"
	"github.com/spf13/cobra"

	"kubesphere.io/kubesphere/pkg/filter"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/signals"
)

var (
	port     int
	certFile string
	keyFile  string
	cmd      = &cobra.Command{
		Use:   "ks-apiserver",
		Short: "",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {

			version.PrintAndExitIfRequested()

			v, err := client.K8sClient().ServerVersion()

			if err != nil {
				glog.Fatalln(err)
			}

			glog.Infoln("kubernetes server version", v)

			stopChan := signals.SetupSignalHandler()
			informers.SharedInformerFactory().Start(stopChan)
			informers.SharedInformerFactory().WaitForCacheSync(stopChan)
			log.Println("resources sync success")

			container := restful.NewContainer()
			container.Filter(filter.Logging)

			apis := apibuilder.APIBuilder{metrics.AddToContainer, resources.AddToContainer, operations.AddToContainer, monitoring.AddToContainer}
			apis.AddToContainer(container)

			log.Printf("Server listening on %d.", port)

			if certFile != "" && keyFile != "" {
				return http.ListenAndServeTLS(fmt.Sprintf(":%d", port), certFile, keyFile, container)
			} else {
				return http.ListenAndServe(fmt.Sprintf(":%d", port), container)
			}
		},
	}
)

func init() {
	cmd.Flags().AddGoFlagSet(goflag.CommandLine)
	cmd.PersistentFlags().IntVarP(&port, "insecure-port", "p", 9090, "server port")
	cmd.PersistentFlags().StringVarP(&certFile, "tls-cert-file", "", "", "TLS cert")
	cmd.PersistentFlags().StringVarP(&keyFile, "tls-key-file", "", "", "TLS key")
}

func main() {
	goflag.Parse()
	glog.CopyStandardLogTo("INFO")
	if err := cmd.Execute(); err != nil {
		log.Fatalf("error: %v\n", err)
	}
}
