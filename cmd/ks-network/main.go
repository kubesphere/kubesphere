package main

import (
	"flag"
	"time"

	"github.com/projectcalico/libcalico-go/lib/apiconfig"
	"github.com/projectcalico/libcalico-go/lib/clientv3"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	ksinformer "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/controller/network/nsnetworkpolicy"
	"kubesphere.io/kubesphere/pkg/controller/network/provider"
)

const (
	certPath = "/calicocerts"
)

var npProviderFlag string

func init() {
	flag.StringVar(&npProviderFlag, "np-provider", "calico", "specify the network policy provider, k8s or calico")
}
func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()
	klog.V(1).Info("Preparing kubernetes client")
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	k8sClientset := kubernetes.NewForConfigOrDie(config)
	ksClientset := versioned.NewForConfigOrDie(config)
	informer := ksinformer.NewSharedInformerFactory(ksClientset, time.Minute*10)
	klog.V(1).Info("Kubernetes client initialized successfully")
	var npProvider provider.NsNetworkPolicyProvider
	if npProviderFlag == "calico" {
		klog.V(1).Info("Preparing calico client")
		config := apiconfig.NewCalicoAPIConfig()
		config.Spec.EtcdEndpoints = "https://127.0.0.1:2379"
		config.Spec.EtcdKeyFile = certPath + "/etcd-key"
		config.Spec.EtcdCertFile = certPath + "/etcd-cert"
		config.Spec.EtcdCACertFile = certPath + "/etcd-ca"
		config.Spec.DatastoreType = apiconfig.EtcdV3
		client, err := clientv3.New(*config)
		if err != nil {
			klog.Fatal("Failed to initialize calico client", err)
		}
		npProvider = provider.NewCalicoNetworkProvider(client.NetworkPolicies())
		klog.V(1).Info("Calico client initialized successfully")
	}
	//TODO: support no-calico cni
	c := nsnetworkpolicy.NewController(k8sClientset, ksClientset, informer.Network().V1alpha1().NamespaceNetworkPolicies(), npProvider)
	stop := make(chan struct{})
	klog.V(1).Infof("Starting controller")
	go informer.Network().V1alpha1().NamespaceNetworkPolicies().Informer().Run(stop)
	if err := c.Run(1, stop); err != nil {
		klog.Fatal(err)
	}
}
