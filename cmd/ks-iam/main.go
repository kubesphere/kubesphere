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
package main

import (
	goflag "flag"
	"fmt"
	"log"
	"net/http"

	"github.com/golang/glog"

	"kubesphere.io/kubesphere/pkg/apibuilder"
	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/version"

	"github.com/emicklei/go-restful"
	"github.com/spf13/cobra"

	iam "kubesphere.io/kubesphere/pkg/apis/iam/install"
	"kubesphere.io/kubesphere/pkg/filter"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/signals"
)

var (
	port     int
	certFile string
	keyFile  string
	cmd      = &cobra.Command{
		Use:   "ks-iam",
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

			apis := apibuilder.APIBuilder{iam.AddToContainer}
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
