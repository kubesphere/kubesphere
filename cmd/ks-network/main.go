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
	"flag"

	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/controller/network/runoption"
)

var opt runoption.RunOption

func init() {
	flag.StringVar(&opt.ProviderName, "np-provider", "calico", "specify the network policy provider, k8s or calico")
	flag.BoolVar(&opt.AllowInsecureEtcd, "allow-insecure-etcd", false, "specify allow connect to etcd using insecure http")
	flag.StringVar(&opt.DataStoreType, "datastore-type", "k8s", "specify the datastore type of calico")
	//TODO add more flags
}

func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()
	klog.V(1).Info("Preparing kubernetes client")
	klog.Fatal(opt.Run())
}
