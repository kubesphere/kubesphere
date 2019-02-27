package main

import (
	goflag "flag"
	"fmt"
	"kubesphere.io/kubesphere/pkg/apibuilder"
	metrics "kubesphere.io/kubesphere/pkg/apis/metrics/install"
	operations "kubesphere.io/kubesphere/pkg/apis/operations/install"
	resources "kubesphere.io/kubesphere/pkg/apis/resources/install"
	"log"
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
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

			stopChan := signals.SetupSignalHandler()
			informers.SharedInformerFactory().Start(stopChan)
			informers.SharedInformerFactory().WaitForCacheSync(stopChan)
			log.Println("resources sync success")

			container := restful.NewContainer()
			container.Filter(filter.Logging)

			apis := make(apibuilder.APIBuilder, 0)
			apis = append(apis, metrics.Install, resources.Install, operations.Install)
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
	glog.CopyStandardLogTo("INFO")
	cmd.Flags().AddGoFlagSet(goflag.CommandLine)
	cmd.PersistentFlags().IntVarP(&port, "insecure-port", "p", 9090, "server port")
	cmd.PersistentFlags().StringVarP(&certFile, "tls-cert-file", "", "", "TLS cert")
	cmd.PersistentFlags().StringVarP(&keyFile, "tls-key-file", "", "", "TLS key")
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("error: %v\n", err)
	}
}
