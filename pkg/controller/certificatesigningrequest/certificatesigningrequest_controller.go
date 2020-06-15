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

package certificatesigningrequest

import (
	"fmt"
	certificatesv1beta1 "k8s.io/api/certificates/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	certificatesinformers "k8s.io/client-go/informers/certificates/v1beta1"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	certificateslisters "k8s.io/client-go/listers/certificates/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/kubeconfig"
	"time"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is csrSynced
	successSynced = "Synced"
	// is csrSynced successfully
	messageResourceSynced = "CertificateSigningRequest csrSynced successfully"
	controllerName        = "csr-controller"
)

type Controller struct {
	k8sclient   kubernetes.Interface
	csrInformer certificatesinformers.CertificateSigningRequestInformer
	csrLister   certificateslisters.CertificateSigningRequestLister
	csrSynced   cache.InformerSynced
	cmSynced    cache.InformerSynced
	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder           record.EventRecorder
	kubeconfigOperator kubeconfig.Interface
}

func NewController(k8sClient kubernetes.Interface, csrInformer certificatesinformers.CertificateSigningRequestInformer,
	configMapInformer corev1informers.ConfigMapInformer, config *rest.Config) *Controller {
	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.

	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: k8sClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerName})
	ctl := &Controller{
		k8sclient:          k8sClient,
		csrInformer:        csrInformer,
		csrLister:          csrInformer.Lister(),
		csrSynced:          csrInformer.Informer().HasSynced,
		cmSynced:           configMapInformer.Informer().HasSynced,
		kubeconfigOperator: kubeconfig.NewOperator(k8sClient, configMapInformer, config),
		workqueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "CertificateSigningRequest"),
		recorder:           recorder,
	}
	klog.Info("Setting up event handlers")
	csrInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctl.enqueueCertificateSigningRequest,
		UpdateFunc: func(old, new interface{}) {
			ctl.enqueueCertificateSigningRequest(new)
		},
		DeleteFunc: ctl.enqueueCertificateSigningRequest,
	})
	return ctl
}

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the csrInformer factories to begin populating the csrInformer caches
	klog.Info("Starting CSR controller")

	// Wait for the caches to be csrSynced before starting workers
	klog.Info("Waiting for csrInformer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.csrSynced, c.cmSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process Foo resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")
	return nil
}

func (c *Controller) enqueueCertificateSigningRequest(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the csrInformer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the reconcile, passing it the namespace/name string of the
		// Foo resource to be csrSynced.
		if err := c.reconcile(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully csrSynced %s:%s", "key", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (c *Controller) reconcile(key string) error {

	// Get the CertificateSigningRequest with this name
	csr, err := c.csrLister.Get(key)
	if err != nil {
		// The user may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("csr '%s' in work queue no longer exists", key))
			return nil
		}
		klog.Error(err)
		return err
	}

	// csr create by kubesphere auto approve
	if username := csr.Labels[constants.UsernameLabelKey]; username != "" {
		err = c.Approve(csr)
		if err != nil {
			klog.Error(err)
			return err
		}
		// certificate data is not empty
		if len(csr.Status.Certificate) > 0 {
			err = c.UpdateKubeconfig(csr)
			if err != nil {
				klog.Error(err)
				return err
			}
			// release
			err := c.k8sclient.CertificatesV1beta1().CertificateSigningRequests().Delete(csr.Name, metav1.NewDeleteOptions(0))
			if err != nil {
				klog.Error(err)
				return err
			}
		}
	}

	c.recorder.Event(csr, corev1.EventTypeNormal, successSynced, messageResourceSynced)
	return nil
}

func (c *Controller) Start(stopCh <-chan struct{}) error {
	return c.Run(4, stopCh)
}

func (c *Controller) Approve(csr *certificatesv1beta1.CertificateSigningRequest) error {
	// is approved
	if len(csr.Status.Certificate) > 0 {
		return nil
	}
	csr.Status = certificatesv1beta1.CertificateSigningRequestStatus{
		Conditions: []certificatesv1beta1.CertificateSigningRequestCondition{{
			Type:    "Approved",
			Reason:  "KubeSphereApprove",
			Message: "This CSR was approved by KubeSphere",
			LastUpdateTime: metav1.Time{
				Time: time.Now(),
			},
		}},
	}

	// approve csr
	csr, err := c.k8sclient.CertificatesV1beta1().CertificateSigningRequests().UpdateApproval(csr)

	if err != nil {
		klog.Errorln(err)
		return err
	}

	return nil
}

func (c *Controller) UpdateKubeconfig(csr *certificatesv1beta1.CertificateSigningRequest) error {
	username := csr.Labels[constants.UsernameLabelKey]

	err := c.kubeconfigOperator.UpdateKubeconfig(username, csr.Status.Certificate)

	if err != nil {
		klog.Error(err)
	}

	return err
}
