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

package pipeline

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	corev1informer "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	corev1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	devopsv1alpha3 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha3"
	kubesphereclient "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	devopsinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/devops/v1alpha3"
	devopslisters "kubesphere.io/kubesphere/pkg/client/listers/devops/v1alpha3"
	"kubesphere.io/kubesphere/pkg/constants"
	devopsClient "kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"reflect"
	"time"
)

/**
  DevOps project controller is used to maintain the state of the DevOps project.
*/

type Controller struct {
	client           clientset.Interface
	kubesphereClient kubesphereclient.Interface

	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder

	devOpsProjectLister devopslisters.PipelineLister
	pipelineSynced      cache.InformerSynced

	namespaceLister corev1lister.NamespaceLister
	namespaceSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface

	workerLoopPeriod time.Duration

	devopsClient devopsClient.Interface
}

func NewController(client clientset.Interface,
	kubesphereClient kubesphereclient.Interface,
	devopsClinet devopsClient.Interface,
	namespaceInformer corev1informer.NamespaceInformer,
	devopsInformer devopsinformers.PipelineInformer) *Controller {

	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(func(format string, args ...interface{}) {
		klog.Info(fmt.Sprintf(format, args))
	})
	broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: client.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "pipeline-controller"})

	v := &Controller{
		client:              client,
		devopsClient:        devopsClinet,
		kubesphereClient:    kubesphereClient,
		workqueue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pipeline"),
		devOpsProjectLister: devopsInformer.Lister(),
		pipelineSynced:      devopsInformer.Informer().HasSynced,
		namespaceLister:     namespaceInformer.Lister(),
		namespaceSynced:     namespaceInformer.Informer().HasSynced,
		workerLoopPeriod:    time.Second,
	}

	v.eventBroadcaster = broadcaster
	v.eventRecorder = recorder

	devopsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: v.enqueuePipeline,
		UpdateFunc: func(oldObj, newObj interface{}) {
			old := oldObj.(*devopsv1alpha3.Pipeline)
			new := newObj.(*devopsv1alpha3.Pipeline)
			if old.ResourceVersion == new.ResourceVersion {
				return
			}
			v.enqueuePipeline(newObj)
		},
		DeleteFunc: v.enqueuePipeline,
	})
	return v
}

// enqueuePipeline takes a Foo resource and converts it into a namespace/name
// string which is then put onto the work workqueue. This method should *not* be
// passed resources of any type other than DevOpsProject.
func (c *Controller) enqueuePipeline(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool

		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		if err := c.syncHandler(key); err != nil {
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		c.workqueue.Forget(obj)
		klog.V(5).Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		klog.Error(err, "could not reconcile devopsProject")
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) worker() {

	for c.processNextWorkItem() {
	}
}

func (c *Controller) Start(stopCh <-chan struct{}) error {
	return c.Run(1, stopCh)
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Info("starting pipeline controller")
	defer klog.Info("shutting down  pipeline controller")

	if !cache.WaitForCacheSync(stopCh, c.pipelineSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.worker, c.workerLoopPeriod, stopCh)
	}

	<-stopCh
	return nil
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the pipeline resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	nsName, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.Error(err, fmt.Sprintf("could not split copyPipeline meta %s ", key))
		return nil
	}
	namespace, err := c.namespaceLister.Get(nsName)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Info(fmt.Sprintf("namespace '%s' in work queue no longer exists ", key))
			return nil
		}
		klog.V(8).Info(err, fmt.Sprintf("could not get namespace %s ", key))
		return err
	}
	if !isDevOpsProjectAdminNamespace(namespace) {
		err := fmt.Errorf("cound not create copyPipeline in normal namespaces %s", namespace.Name)
		klog.Warning(err)
		return err
	}

	pipeline, err := c.devOpsProjectLister.Pipelines(nsName).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.V(8).Info(fmt.Sprintf("copyPipeline '%s' in work queue no longer exists ", key))
			return nil
		}
		klog.Error(err, fmt.Sprintf("could not get copyPipeline %s ", key))
		return err
	}

	copyPipeline := pipeline.DeepCopy()
	// DeletionTimestamp.IsZero() means copyPipeline has not been deleted.
	if copyPipeline.ObjectMeta.DeletionTimestamp.IsZero() {
		// https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#finalizers
		if !sliceutil.HasString(copyPipeline.ObjectMeta.Finalizers, devopsv1alpha3.PipelineFinalizerName) {
			copyPipeline.ObjectMeta.Finalizers = append(copyPipeline.ObjectMeta.Finalizers, devopsv1alpha3.PipelineFinalizerName)
		}

		// Check pipeline config exists, otherwise we will create it.
		// if pipeline exists, check & update config
		jenkinsPipeline, err := c.devopsClient.GetProjectPipelineConfig(nsName, pipeline.Name)
		if err == nil {
			if !reflect.DeepEqual(jenkinsPipeline.Spec, copyPipeline.Spec) {
				_, err := c.devopsClient.UpdateProjectPipeline(nsName, copyPipeline)
				if err != nil {
					klog.V(8).Info(err, fmt.Sprintf("failed to update pipeline config %s ", key))
					return err
				}
			}
		} else {
			_, err := c.devopsClient.CreateProjectPipeline(nsName, copyPipeline)
			if err != nil {
				klog.V(8).Info(err, fmt.Sprintf("failed to create copyPipeline %s ", key))
				return err
			}
		}

	} else {
		// Finalizers processing logic
		if sliceutil.HasString(copyPipeline.ObjectMeta.Finalizers, devopsv1alpha3.PipelineFinalizerName) {
			if _, err := c.devopsClient.DeleteProjectPipeline(nsName, pipeline.Name); err != nil {
				klog.V(8).Info(err, fmt.Sprintf("failed to delete pipeline %s in devops", key))
			}
			copyPipeline.ObjectMeta.Finalizers = sliceutil.RemoveString(copyPipeline.ObjectMeta.Finalizers, func(item string) bool {
				return item == devopsv1alpha3.PipelineFinalizerName
			})

		}
	}
	if !reflect.DeepEqual(pipeline, copyPipeline) {
		_, err = c.kubesphereClient.DevopsV1alpha3().Pipelines(nsName).Update(copyPipeline)
		if err != nil {
			klog.V(8).Info(err, fmt.Sprintf("failed to update pipeline %s ", key))
			return err
		}
	}

	return nil
}

func isDevOpsProjectAdminNamespace(namespace *v1.Namespace) bool {
	_, ok := namespace.Labels[constants.DevOpsProjectLabelKey]

	return ok && k8sutil.IsControlledBy(namespace.OwnerReferences,
		devopsv1alpha3.ResourceKindDevOpsProject, "")

}
