package main

import (
	goflag "flag"
	"fmt"
	"log"
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
	"github.com/spf13/cobra"

	"kubesphere.io/kubesphere/pkg/apiserver"
	"kubesphere.io/kubesphere/pkg/filter"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/monitoring"
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
			apiserver.AddToContainer(container)
			monitoring.AddToContainer(container)

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
	cmd.PersistentFlags().IntVarP(&port, "port", "p", 9090, "server port")
	cmd.PersistentFlags().StringVarP(&certFile, "certFile", "", "", "TLS cert")
	cmd.PersistentFlags().StringVarP(&keyFile, "keyFile", "", "", "TLS key")
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("error: %v\n", err)
	}
}
