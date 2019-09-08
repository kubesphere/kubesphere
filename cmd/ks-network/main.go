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
