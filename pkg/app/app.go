/*
Copyright 2018 The KubeSphere Authors.

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
	"crypto/tls"
	"fmt"

	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"net"
	"net/http"

	_ "kubesphere.io/kubesphere/pkg/apis/v1alpha"
	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/controllers"
	"kubesphere.io/kubesphere/pkg/options"
)

type kubeSphereServer struct {
	insecureBindAddress net.IP
	bindAddress         net.IP
	insecurePort        int
	port                int
	certFile            string
	keyFile             string
	container           *restful.Container
}

func newKubeSphereServer(options *options.ServerRunOptions) *kubeSphereServer {

	s := kubeSphereServer{
		insecureBindAddress: options.GetInsecureBindAddress(),
		bindAddress:         options.GetBindAddress(),
		insecurePort:        options.GetInsecurePort(),
		port:                options.GetPort(),
		certFile:            options.GetCertFile(),
		keyFile:             options.GetKeyFile(),
	}

	return &s
}

func preCheck() error {
	k8sClient := client.NewK8sClient()
	nsList, err := k8sClient.CoreV1().Namespaces().List(meta_v1.ListOptions{})
	if err != nil {
		return err
	}
	for _, ns := range nsList.Items {
		if ns.Name == constants.KubeSphereControlNameSpace {
			return nil
		}
	}
	namespace := v1.Namespace{ObjectMeta: meta_v1.ObjectMeta{Name: constants.KubeSphereControlNameSpace}}
	_, err = k8sClient.CoreV1().Namespaces().Create(&namespace)
	return err
}
func (server *kubeSphereServer) run() {
	err := preCheck()
	if err != nil {
		glog.Error(err)
		return
	}

	go controllers.Run()

	if len(server.certFile) > 0 && len(server.keyFile) > 0 {
		servingCert, err := tls.LoadX509KeyPair(server.certFile, server.keyFile)
		if err != nil {
			glog.Error(err)
			return
		}

		secureAddr := fmt.Sprintf("%s:%d", server.bindAddress, server.port)
		glog.Infof("Serving securely on addr: %s", secureAddr)

		httpServer := &http.Server{
			Addr:      secureAddr,
			Handler:   restful.DefaultContainer,
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{servingCert}},
		}

		go func() { glog.Fatal(httpServer.ListenAndServeTLS("", "")) }()

	} else {
		insecureAddr := fmt.Sprintf("%s:%d", server.insecureBindAddress, server.insecurePort)
		glog.Infof("Serving insecurely on addr: %s", insecureAddr)

		go func() { glog.Fatal(http.ListenAndServe(insecureAddr, nil)) }()
	}

	select {}
}

func Run() {
	server := newKubeSphereServer(options.ServerOptions)

	server.run()
}
