/*
Copyright 2020 KubeSphere Authors

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

package application

import (
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	informersv1 "k8s.io/client-go/informers/apps/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	listersv1 "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	log "k8s.io/klog"
	servicemeshinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/servicemesh/v1alpha2"
	servicemeshlisters "kubesphere.io/kubesphere/pkg/client/listers/servicemesh/v1alpha2"
	"kubesphere.io/kubesphere/pkg/controller/virtualservice/util"
	applicationclient "sigs.k8s.io/application/pkg/client/clientset/versioned"
	applicationinformers "sigs.k8s.io/application/pkg/client/informers/externalversions/app/v1beta1"
	applicationlister "sigs.k8s.io/application/pkg/client/listers/app/v1beta1"
)

const (
	// maxRetries is the number of times a service will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the
	// sequence of delays between successive queuings of a service.
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s, 5.1s, 10.2s, 20.4s, 41s, 82s
	maxRetries = 15
)

type ApplicationController struct {
	client clientset.Interface

	applicationClient applicationclient.Interface

	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder

	applicationLister applicationlister.ApplicationLister
	applicationSynced cache.InformerSynced

	serviceLister corelisters.ServiceLister
	serviceSynced cache.InformerSynced

	deploymentLister listersv1.DeploymentLister
	deploymentSynced cache.InformerSynced

	statefulSetLister listersv1.StatefulSetLister
	statefulSetSynced cache.InformerSynced

	strategyLister servicemeshlisters.StrategyLister
	strategySynced cache.InformerSynced

	servicePolicyLister servicemeshlisters.ServicePolicyLister
	servicePolicySynced cache.InformerSynced

	queue workqueue.RateLimitingInterface

	workerLoopPeriod time.Duration
}

func NewApplicationController(serviceInformer coreinformers.ServiceInformer,
	deploymentInformer informersv1.DeploymentInformer,
	statefulSetInformer informersv1.StatefulSetInformer,
	strategyInformer servicemeshinformers.StrategyInformer,
	servicePolicyInformer servicemeshinformers.ServicePolicyInformer,
	applicationInformer applicationinformers.ApplicationInformer,
	client clientset.Interface,
	applicationClient applicationclient.Interface) *ApplicationController {

	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(func(format string, args ...interface{}) {
		log.Info(fmt.Sprintf(format, args))
	})
	broadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: client.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "application-controller"})

	v := &ApplicationController{
		client:            client,
		applicationClient: applicationClient,
		queue:             workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "application"),
		workerLoopPeriod:  time.Second,
	}

	v.deploymentLister = deploymentInformer.Lister()
	v.deploymentSynced = deploymentInformer.Informer().HasSynced

	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    v.enqueueObject,
		DeleteFunc: v.enqueueObject,
	})

	v.statefulSetLister = statefulSetInformer.Lister()
	v.statefulSetSynced = statefulSetInformer.Informer().HasSynced

	statefulSetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    v.enqueueObject,
		DeleteFunc: v.enqueueObject,
	})

	v.serviceLister = serviceInformer.Lister()
	v.serviceSynced = serviceInformer.Informer().HasSynced

	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    v.enqueueObject,
		DeleteFunc: v.enqueueObject,
	})

	v.strategyLister = strategyInformer.Lister()
	v.strategySynced = strategyInformer.Informer().HasSynced

	strategyInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    v.enqueueObject,
		DeleteFunc: v.enqueueObject,
	})

	v.servicePolicyLister = servicePolicyInformer.Lister()
	v.servicePolicySynced = servicePolicyInformer.Informer().HasSynced

	servicePolicyInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    v.enqueueObject,
		DeleteFunc: v.enqueueObject,
	})

	v.applicationLister = applicationInformer.Lister()
	v.applicationSynced = applicationInformer.Informer().HasSynced

	v.eventBroadcaster = broadcaster
	v.eventRecorder = recorder

	return v

}

func (v *ApplicationController) Start(stopCh <-chan struct{}) error {
	return v.Run(2, stopCh)
}

func (v *ApplicationController) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer v.queue.ShutDown()

	log.Info("starting application controller")
	defer log.Info("shutting down application controller")

	if !cache.WaitForCacheSync(stopCh, v.deploymentSynced, v.statefulSetSynced, v.serviceSynced, v.strategySynced, v.servicePolicySynced, v.applicationSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < workers; i++ {
		go wait.Until(v.worker, v.workerLoopPeriod, stopCh)
	}
	<-stopCh
	return nil
}

func (v *ApplicationController) worker() {

	for v.processNextWorkItem() {
	}
}

func (v *ApplicationController) processNextWorkItem() bool {
	eKey, quit := v.queue.Get()
	if quit {
		return false
	}

	defer v.queue.Done(eKey)

	err := v.syncApplication(eKey.(string))
	v.handleErr(err, eKey)

	return true
}

func (v *ApplicationController) syncApplication(key string) error {
	startTime := time.Now()
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Error(err, "not a valid controller key", "key", key)
		return err
	}

	defer func() {
		log.V(4).Info("Finished updating application.", "namespace", namespace, "name", name, "duration", time.Since(startTime))
	}()

	application, err := v.applicationLister.Applications(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			// application has been deleted
			return nil
		}
		log.Error(err, "get application failed")
	}

	annotations := application.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations["kubesphere.io/last-updated"] = time.Now().String()
	application.SetAnnotations(annotations)

	_, err = v.applicationClient.AppV1beta1().Applications(namespace).Update(application)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(4).Info("application has been deleted during update")
			return nil
		}
		log.Error(err, "failed to update application", "namespace", namespace, "name", name)
		return err
	}

	return nil
}

func (v *ApplicationController) enqueueObject(obj interface{}) {
	var resource = obj.(metav1.Object)

	if resource.GetLabels() == nil || !util.IsApplicationComponent(resource.GetLabels()) {
		return
	}

	applicationName := util.GetApplictionName(resource.GetLabels())

	if len(applicationName) > 0 {
		key := resource.GetNamespace() + "/" + applicationName
		v.queue.Add(key)
	}
}

func (v *ApplicationController) handleErr(err error, key interface{}) {
	if err == nil {
		v.queue.Forget(key)
		return
	}

	if v.queue.NumRequeues(key) < maxRetries {
		log.V(2).Info("Error syncing virtualservice for service retrying.", "key", key, "error", err)
		v.queue.AddRateLimited(key)
		return
	}

	log.V(4).Info("Dropping service out of the queue.", "key", key, "error", err)
	v.queue.Forget(key)
	utilruntime.HandleError(err)
}
