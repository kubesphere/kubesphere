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

package devopscredential

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
	"kubesphere.io/kubesphere/pkg/constants"
	devopsClient "kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"reflect"
	"strings"
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

	secretLister corev1lister.SecretLister
	secretSynced cache.InformerSynced

	namespaceLister corev1lister.NamespaceLister
	namespaceSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface

	workerLoopPeriod time.Duration

	devopsClient devopsClient.Interface
}

func NewController(client clientset.Interface,
	devopsClinet devopsClient.Interface,
	namespaceInformer corev1informer.NamespaceInformer,
	secretInformer corev1informer.SecretInformer) *Controller {

	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(func(format string, args ...interface{}) {
		klog.Info(fmt.Sprintf(format, args))
	})
	broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: client.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "devopscredential-controller"})

	v := &Controller{
		client:           client,
		devopsClient:     devopsClinet,
		workqueue:        workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "devopscredential"),
		secretLister:     secretInformer.Lister(),
		secretSynced:     secretInformer.Informer().HasSynced,
		namespaceLister:  namespaceInformer.Lister(),
		namespaceSynced:  namespaceInformer.Informer().HasSynced,
		workerLoopPeriod: time.Second,
	}

	v.eventBroadcaster = broadcaster
	v.eventRecorder = recorder

	secretInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			secret, ok := obj.(*v1.Secret)
			if ok && strings.HasPrefix(string(secret.Type), devopsv1alpha3.DevOpsCredentialPrefix) {
				v.enqueueSecret(obj)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			old, ook := oldObj.(*v1.Secret)
			new, nok := newObj.(*v1.Secret)
			if ook && nok && old.ResourceVersion == new.ResourceVersion {
				return
			}
			if ook && nok && strings.HasPrefix(string(new.Type), devopsv1alpha3.DevOpsCredentialPrefix) {
				v.enqueueSecret(newObj)
			}
		},
		DeleteFunc: func(obj interface{}) {
			secret, ok := obj.(*v1.Secret)
			if ok && strings.HasPrefix(string(secret.Type), devopsv1alpha3.DevOpsCredentialPrefix) {
				v.enqueueSecret(obj)
			}
		},
	})
	return v
}

// enqueueSecret takes a Foo resource and converts it into a namespace/name
// string which is then put onto the work workqueue. This method should *not* be
// passed resources of any type other than DevOpsProject.
func (c *Controller) enqueueSecret(obj interface{}) {
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

	klog.Info("starting devopscredential controller")
	defer klog.Info("shutting down  devopscredential controller")

	if !cache.WaitForCacheSync(stopCh, c.secretSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.worker, c.workerLoopPeriod, stopCh)
	}

	<-stopCh
	return nil
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the secret resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	nsName, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.Error(err, fmt.Sprintf("could not split copySecret meta %s ", key))
		return nil
	}
	namespace, err := c.namespaceLister.Get(nsName)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Info(fmt.Sprintf("namespace '%s' in work queue no longer exists ", key))
			return nil
		}
		klog.Error(err, fmt.Sprintf("could not get namespace %s ", key))
		return err
	}
	if !isDevOpsProjectAdminNamespace(namespace) {
		err := fmt.Errorf("cound not create credential in normal namespaces %s", namespace.Name)
		klog.Warning(err)
		return err
	}

	secret, err := c.secretLister.Secrets(nsName).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Info(fmt.Sprintf("secret '%s' in work queue no longer exists ", key))
			return nil
		}
		klog.Error(err, fmt.Sprintf("could not get secret %s ", key))
		return err
	}

	copySecret := secret.DeepCopy()
	// DeletionTimestamp.IsZero() means copySecret has not been deleted.
	if secret.ObjectMeta.DeletionTimestamp.IsZero() {
		// https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#finalizers
		if !sliceutil.HasString(secret.ObjectMeta.Finalizers, devopsv1alpha3.CredentialFinalizerName) {
			copySecret.ObjectMeta.Finalizers = append(copySecret.ObjectMeta.Finalizers, devopsv1alpha3.CredentialFinalizerName)
		}
		// Check secret config exists, otherwise we will create it.
		// if secret exists, update config
		_, err := c.devopsClient.CreateCredentialInProject(nsName, copySecret)
		if err != nil {
			if _, ok := copySecret.Annotations[devopsv1alpha3.CredentialAutoSyncAnnoKey]; ok {
				_, err := c.devopsClient.UpdateCredentialInProject(nsName, copySecret)
				if err != nil {
					klog.V(8).Info(err, fmt.Sprintf("failed to update secret %s ", key))
					return err
				}
			}
		}

	} else {
		// Finalizers processing logic
		if sliceutil.HasString(copySecret.ObjectMeta.Finalizers, devopsv1alpha3.CredentialFinalizerName) {
			if _, err := c.devopsClient.DeleteCredentialInProject(nsName, secret.Name); err != nil {
				klog.V(8).Info(err, fmt.Sprintf("failed to delete secret %s in devops", key))
				return err
			}
			copySecret.ObjectMeta.Finalizers = sliceutil.RemoveString(copySecret.ObjectMeta.Finalizers, func(item string) bool {
				return item == devopsv1alpha3.CredentialFinalizerName
			})

		}
	}
	if !reflect.DeepEqual(secret, copySecret) {
		_, err = c.client.CoreV1().Secrets(nsName).Update(copySecret)
		if err != nil {
			klog.V(8).Info(err, fmt.Sprintf("failed to update secret %s ", key))
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
