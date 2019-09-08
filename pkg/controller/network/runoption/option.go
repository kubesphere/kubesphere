package runoption

import (
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

	KubernetesDataStore = "k8s"
	EtcdDataStore       = "etcd"
)

type RunOption struct {
	ProviderName      string
	DataStoreType     string
	EtcdEndpoints     string
	AllowInsecureEtcd bool
}

func (r RunOption) Run() error {
	klog.V(1).Info("Check config")
	if err := r.check(); err != nil {
		return err
	}
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

	if r.ProviderName == "calico" {
		klog.V(1).Info("Preparing calico client")
		config := apiconfig.NewCalicoAPIConfig()
		config.Spec.EtcdEndpoints = r.EtcdEndpoints
		if !r.AllowInsecureEtcd {
			config.Spec.EtcdKeyFile = certPath + "/etcd-key"
			config.Spec.EtcdCertFile = certPath + "/etcd-cert"
			config.Spec.EtcdCACertFile = certPath + "/etcd-ca"
		}
		if r.DataStoreType == KubernetesDataStore {
			config.Spec.DatastoreType = apiconfig.Kubernetes
		} else {
			config.Spec.DatastoreType = apiconfig.EtcdV3
		}
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
	return c.Run(1, stop)
}

func (r RunOption) check() error {
	return nil
}
