package app

import (
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"kubesphere.io/kubesphere/pkg/controller/destinationrule"
	"kubesphere.io/kubesphere/pkg/controller/virtualservice"
	"kubesphere.io/kubesphere/pkg/simple/controller/namespace"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	istioclientset "github.com/knative/pkg/client/clientset/versioned"
	istioinformers "github.com/knative/pkg/client/informers/externalversions"
	servicemeshclientset "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	servicemeshinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
)

const defaultResync = 600 * time.Second

var log = logf.Log.WithName("controller-manager")

func AddControllers(mgr manager.Manager, cfg *rest.Config, stopCh <-chan struct{}) error {

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Error(err, "building kubernetes client failed")
	}

	istioclient, err := istioclientset.NewForConfig(cfg)
	if err != nil {
		log.Error(err, "create istio client failed")
		return err
	}

	informerFactory := informers.NewSharedInformerFactory(kubeClient, defaultResync)
	istioInformer := istioinformers.NewSharedInformerFactory(istioclient, defaultResync)

	servicemeshclient, err := servicemeshclientset.NewForConfig(cfg)
	if err != nil {
		log.Error(err, "create servicemesh client failed")
		return err
	}

	servicemeshinformer := servicemeshinformers.NewSharedInformerFactory(servicemeshclient, defaultResync)

	vsController := virtualservice.NewVirtualServiceController(informerFactory.Core().V1().Services(),
		istioInformer.Networking().V1alpha3().VirtualServices(),
		istioInformer.Networking().V1alpha3().DestinationRules(),
		servicemeshinformer.Servicemesh().V1alpha2().Strategies(),
		kubeClient,
		istioclient,
		servicemeshclient)

	drController := destinationrule.NewDestinationRuleController(informerFactory.Apps().V1().Deployments(),
		istioInformer.Networking().V1alpha3().DestinationRules(),
		informerFactory.Core().V1().Services(),
		servicemeshinformer.Servicemesh().V1alpha2().ServicePolicies(),
		kubeClient,
		istioclient)

	nsController := namespace.NewNamespaceController(kubeClient,
		informerFactory.Core().V1().Namespaces(),
		informerFactory.Rbac().V1().Roles(),
	)

	servicemeshinformer.Start(stopCh)
	istioInformer.Start(stopCh)
	informerFactory.Start(stopCh)

	controllers := map[string]manager.Runnable{
		"virtualservice-controller":  vsController,
		"destinationrule-controller": drController,
		"namespace-controller":       nsController,
	}

	for name, ctrl := range controllers {
		err = mgr.Add(ctrl)
		if err != nil {
			log.Error(err, "add controller to manager failed", "name", name)
			return err
		}
	}

	return nil
}
