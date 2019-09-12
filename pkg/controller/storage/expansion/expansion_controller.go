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

package expansion

import (
	"errors"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsv1informers "k8s.io/client-go/informers/apps/v1"
	corev1informers "k8s.io/client-go/informers/core/v1"
	storagev1informer "k8s.io/client-go/informers/storage/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	listerappsv1 "k8s.io/client-go/listers/apps/v1"
	listercorev1 "k8s.io/client-go/listers/core/v1"
	listerstoragev1 "k8s.io/client-go/listers/storage/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"time"
)

const controllerAgentName = "expansion-controller"

var supportedProvisioner = []string{"disk.csi.qingcloud.com", "csi-qingcloud"}

var retryTime = wait.Backoff{
	Duration: 1 * time.Second,
	Factor:   2,
	Steps:    12,
}

type VolumeExpansionController struct {
	kubeclientset kubernetes.Interface
	pvcLister     listercorev1.PersistentVolumeClaimLister
	pvcSynced     cache.InformerSynced
	classLister   listerstoragev1.StorageClassLister
	classSynced   cache.InformerSynced
	podLister     listercorev1.PodLister
	podSynced     cache.InformerSynced
	deployLister  listerappsv1.DeploymentLister
	deploySynced  cache.InformerSynced
	rsLister      listerappsv1.ReplicaSetLister
	rsSynced      cache.InformerSynced
	stsLister     listerappsv1.StatefulSetLister
	stsSynced     cache.InformerSynced
	workqueue     workqueue.RateLimitingInterface
	recorder      record.EventRecorder
}

// NewController returns a new volume expansion controller
func NewVolumeExpansionController(
	kubeclientset kubernetes.Interface,
	pvcInformer corev1informers.PersistentVolumeClaimInformer,
	classInformer storagev1informer.StorageClassInformer,
	podInformer corev1informers.PodInformer,
	deployInformer appsv1informers.DeploymentInformer,
	rsInformer appsv1informers.ReplicaSetInformer,
	stsInformer appsv1informers.StatefulSetInformer) *VolumeExpansionController {
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	controller := &VolumeExpansionController{
		kubeclientset: kubeclientset,
		pvcLister:     pvcInformer.Lister(),
		pvcSynced:     pvcInformer.Informer().HasSynced,
		classLister:   classInformer.Lister(),
		classSynced:   classInformer.Informer().HasSynced,
		podLister:     podInformer.Lister(),
		podSynced:     podInformer.Informer().HasSynced,
		deployLister:  deployInformer.Lister(),
		deploySynced:  deployInformer.Informer().HasSynced,
		rsLister:      rsInformer.Lister(),
		rsSynced:      rsInformer.Informer().HasSynced,
		stsLister:     stsInformer.Lister(),
		stsSynced:     stsInformer.Informer().HasSynced,
		workqueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "expansion"),
		recorder:      recorder,
	}
	klog.V(2).Info("Setting up event handlers")
	pvcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueuePVC,
		UpdateFunc: func(old, new interface{}) {
			oldPVC, ok := old.(*corev1.PersistentVolumeClaim)
			if !ok {
				return
			}
			oldSize := oldPVC.Spec.Resources.Requests[corev1.ResourceStorage]
			newPVC, ok := new.(*corev1.PersistentVolumeClaim)
			if !ok {
				return
			}
			newSize := newPVC.Spec.Resources.Requests[corev1.ResourceStorage]
			if newSize.Cmp(oldSize) > 0 && newSize.Cmp(newPVC.Status.Capacity[corev1.ResourceStorage]) > 0 {
				controller.handleObject(new)
			}
		},
		DeleteFunc: controller.enqueuePVC,
	})
	return controller
}

func (c *VolumeExpansionController) Start(stopCh <-chan struct{}) error {
	return c.Run(5, stopCh)
}

func (c *VolumeExpansionController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()
	klog.V(2).Info("Starting expand volume controller")
	klog.V(2).Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.pvcSynced, c.classSynced, c.podSynced, c.deploySynced, c.rsSynced,
		c.stsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.V(2).Info("Starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.V(2).Info("Started workers")
	<-stopCh
	klog.V(2).Info("Shutting down workers")

	return nil
}

func (c *VolumeExpansionController) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *VolumeExpansionController) processNextWorkItem() bool {
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
		klog.V(2).Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler will re-attach PVC on workloads.
// Step 1. Find the workload (deployment or statefulset) mounting PVC.
// If more than one workloads mounts the same PVC, the controller will not
// do anything.
// Step 2. Verify workload types.
// Step 3. Scale down workload.
// Step 4. Retry to check PVC status.
// Step 5. Scale up workload.
func (c *VolumeExpansionController) syncHandler(key string) error {
	klog.V(5).Infof("syncHandler: handle %s", key)
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get PVC source
	pvc, err := c.pvcLister.PersistentVolumeClaims(namespace).Get(name)
	if err != nil {
		// The PVC resource may no longer exist, in which case we stop processing.
		if apierrors.IsNotFound(err) {
			klog.V(4).Infof("PVC '%s' in work queue no longer exists", key)
			utilruntime.HandleError(fmt.Errorf("PVC '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}
	// find workload
	workload, err := c.findWorkload(name, namespace)
	if err != nil {
		return err
	}
	if workload == nil {
		klog.V(4).Infof("Cannot find any Pods mounting PVC %s", key)
		return nil
	}
	klog.V(5).Infof("Find workload %T pvc name %s", workload, pvc.GetName())
	// handle supported workload
	switch workload.(type) {
	case *appsv1.StatefulSet:
		sts := workload.(*appsv1.StatefulSet)
		klog.V(5).Infof("Find StatefulSet %s", sts.GetName())
	case *appsv1.Deployment:
		deploy := workload.(*appsv1.Deployment)
		klog.V(5).Infof("Find Deployment %s", deploy.GetName())
	default:
		klog.Errorf("Unsupported workload type %T", workload)
		return nil
	}
	// Scale workload to 0
	if err = c.scaleDown(workload, pvc.GetNamespace()); err != nil {
		klog.V(2).Infof("scale down PVC %s mounted workloads failed %s", key, err.Error())
		return err
	}
	klog.V(2).Infof("Scale down PVC %s mounted workloads succeed", key)
	// Wait to scale up
	err = retry.RetryOnConflict(retryTime, func() error {
		klog.V(4).Info("waiting for PVC filesystem expansion")
		if !c.isWaitingScaleUp(name, namespace) {
			return apierrors.NewConflict(schema.GroupResource{Resource: "PersistentVolumeClaim"}, key,
				errors.New("waiting for scaling down and expanding disk"))
		}
		return nil
	})
	klog.V(5).Info("after waiting")
	if err != nil {
		klog.Errorf("Waiting timeout, error: %s", err.Error())
	}

	// Scale up
	if err = c.scaleUp(workload, namespace); err != nil {
		klog.V(2).Infof("Scale up PVC %s mounted workloads failed %s", key, err.Error())
		return err
	}
	klog.V(2).Infof("Scale up PVC %s mounted workloads succeed", key)
	return nil
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the PVC resource that 'owns' it. It does this by looking at the
// StorageClass of PVC whether supporting provisioner and allowing volume
// expansion.
// In KS 2.1, the controller only supports disk.csi.qingcloud.com and
// csi-qingcloud as storageclass provisioner.
func (c *VolumeExpansionController) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		klog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	klog.V(4).Infof("Processing object: %s", object.GetName())

	pvc := obj.(*corev1.PersistentVolumeClaim)
	// Check storage class
	// In KS 2.1, we only support disk.csi.qingcloud.com as storageclass provisioner.
	class := c.getStorageClass(pvc)
	klog.V(4).Infof("Get PVC %s SC was %s", pvc.String(), class.String())
	if class == nil {
		return
	}
	if *class.AllowVolumeExpansion == false {
		return
	}
	for _, p := range supportedProvisioner {
		if class.Provisioner == p {
			klog.V(5).Infof("enqueue PVC %s", claimToClaimKey(pvc))
			c.enqueuePVC(obj)
			return
		}
	}
}

func (c *VolumeExpansionController) getStorageClass(pvc *corev1.PersistentVolumeClaim) *storagev1.StorageClass {
	if pvc == nil {
		return nil
	}
	claimClass := getPersistentVolumeClaimClass(pvc)
	if claimClass == "" {
		klog.V(4).Infof("volume expansion is disabled for PVC without StorageClasses: %s",
			claimToClaimKey(pvc))
		return nil
	}
	class, err := c.classLister.Get(claimClass)
	if err != nil {
		klog.V(4).Infof("failed to expand PVC: %s with error: %v", claimToClaimKey(pvc), err)
		return nil
	}
	return class
}

// enqueuePVC takes a PVC resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than PVC.
func (c *VolumeExpansionController) enqueuePVC(obj interface{}) {
	pvc, ok := obj.(*corev1.PersistentVolumeClaim)
	if !ok {
		return
	}
	size := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	statusSize := pvc.Status.Capacity[corev1.ResourceStorage]

	if pvc.Status.Phase == corev1.ClaimBound && size.Cmp(statusSize) > 0 {
		key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(pvc)
		if err != nil {
			utilruntime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", pvc, err))
			return
		}
		c.workqueue.Add(key)
	}
}

// findWorkload returns the pointer of Pod, StatefulSet or Deployment mounting the PVC.
func (c *VolumeExpansionController) findWorkload(pvc, namespace string) (workloadPtr interface{}, err error) {
	podList, err := c.podLister.Pods(namespace).List(labels.Everything())
	klog.V(4).Infof("podlist len %d", len(podList))
	if err != nil {
		return nil, err
	}
	// Which pod mounting PVC
	podPtrListMountPVC := getPodMountPVC(podList, pvc)
	klog.V(4).Infof("Get %d pods mounting PVC", len(podPtrListMountPVC))
	// In KS 2.1, automatic re-attach PVC only support PVC mounted on a single Pod.
	if len(podPtrListMountPVC) != 1 {
		return nil, nil
	}
	// If pod managed by Deployment, StatefulSet, it returns Deployment or StatefulSet.
	// If not, it returns Pod.
	klog.V(4).Info("Find pod parent")
	ownerRef, err := c.findPodParent(podPtrListMountPVC[0])
	if err != nil {
		return nil, err
	}
	if ownerRef == nil {
		// a single pod
		return podPtrListMountPVC[0], nil
	} else {
		klog.V(4).Infof("OwnerRef kind %s", ownerRef.Kind)
		switch ownerRef.Kind {
		case "StatefulSet":
			return c.stsLister.StatefulSets(namespace).Get(ownerRef.Name)
		case "Deployment":
			return c.deployLister.Deployments(namespace).Get(ownerRef.Name)
		default:
			return nil, nil
		}
	}

}

// If the Pod don't controlled by any controller, return nil.
func (c *VolumeExpansionController) findPodParent(pod *corev1.Pod) (*metav1.OwnerReference, error) {
	if pod == nil {
		return nil, nil
	}
	if ownerRef := metav1.GetControllerOf(pod); ownerRef != nil {
		switch ownerRef.Kind {
		case "ReplicaSet":
			// get deploy
			rs, err := c.rsLister.ReplicaSets(pod.GetNamespace()).Get(ownerRef.Name)
			if err != nil {
				return nil, err
			}
			if rsOwnerRef := metav1.GetControllerOf(rs); rsOwnerRef != nil {
				return rsOwnerRef, nil
			} else {
				return ownerRef, nil
			}
		case "StatefulSet":
			return ownerRef, nil
		default:
			return &metav1.OwnerReference{}, nil
		}
	}
	return nil, nil
}

func (c *VolumeExpansionController) scaleDown(workload interface{}, namespace string) error {
	switch workload.(type) {
	case *appsv1.Deployment:
		deploy := workload.(*appsv1.Deployment)
		scale := &autoscalingv1.Scale{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deploy.GetName(),
				Namespace: deploy.GetNamespace(),
			},
			Spec: autoscalingv1.ScaleSpec{
				Replicas: 0,
			},
		}
		_, err := c.kubeclientset.AppsV1().Deployments(namespace).UpdateScale(deploy.GetName(), scale)
		return err
	case *appsv1.StatefulSet:
		sts := workload.(*appsv1.StatefulSet)
		scale := &autoscalingv1.Scale{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sts.GetName(),
				Namespace: sts.GetNamespace(),
			},
			Spec: autoscalingv1.ScaleSpec{
				Replicas: 0,
			},
		}
		_, err := c.kubeclientset.AppsV1().StatefulSets(namespace).UpdateScale(sts.GetName(), scale)
		return err
	default:
		return fmt.Errorf("unsupported type %T", workload)
	}
}

func (c *VolumeExpansionController) scaleUp(workload interface{}, namespace string) error {
	switch workload.(type) {
	case *appsv1.Deployment:
		deploy := workload.(*appsv1.Deployment)
		scale := &autoscalingv1.Scale{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deploy.GetName(),
				Namespace: deploy.GetNamespace(),
			},
			Spec: autoscalingv1.ScaleSpec{
				Replicas: *deploy.Spec.Replicas,
			},
		}
		_, err := c.kubeclientset.AppsV1().Deployments(namespace).UpdateScale(deploy.GetName(), scale)
		return err
	case *appsv1.StatefulSet:
		sts := workload.(*appsv1.StatefulSet)
		scale := &autoscalingv1.Scale{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sts.GetName(),
				Namespace: sts.GetNamespace(),
			},
			Spec: autoscalingv1.ScaleSpec{
				Replicas: *sts.Spec.Replicas,
			},
		}
		_, err := c.kubeclientset.AppsV1().StatefulSets(namespace).UpdateScale(sts.GetName(), scale)
		return err
	default:
		return fmt.Errorf("unsupported type %T", workload)
	}
}

// isWaitingScaleUp tries to check whether PVC is waiting for restart Pod.
func (c *VolumeExpansionController) isWaitingScaleUp(name, namespace string) bool {
	pvc, err := c.pvcLister.PersistentVolumeClaims(namespace).Get(name)
	if err != nil {
		klog.Errorf("Get PVC error")
	}
	if pvc == nil {
		return false
	}
	for _, condition := range pvc.Status.Conditions {
		if condition.Type == corev1.PersistentVolumeClaimFileSystemResizePending {
			return true
		}
	}
	return false
}
