/*
Copyright 2020 The KubeSphere Authors.

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

package group

import (
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	iam1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	fedv1beta1types "kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	iamv1alpha2informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/iam/v1alpha2"
	fedv1beta1informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/types/v1beta1"
	iamv1alpha1listers "kubesphere.io/kubesphere/pkg/client/listers/iam/v1alpha2"
	fedv1beta1lister "kubesphere.io/kubesphere/pkg/client/listers/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/controller/utils/controller"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	successSynced         = "Synced"
	messageResourceSynced = "Group synced successfully"
	controllerName        = "group-controller"
	finalizer             = "finalizers.kubesphere.io/groups"
)

type Controller struct {
	controller.BaseController
	scheme               *runtime.Scheme
	k8sClient            kubernetes.Interface
	ksClient             kubesphere.Interface
	groupInformer        iamv1alpha2informers.GroupInformer
	groupLister          iamv1alpha1listers.GroupLister
	recorder             record.EventRecorder
	federatedGroupLister fedv1beta1lister.FederatedGroupLister
	multiClusterEnabled  bool
}

// NewController creates Group Controller instance
func NewController(k8sClient kubernetes.Interface, ksClient kubesphere.Interface, groupInformer iamv1alpha2informers.GroupInformer,
	federatedGroupInformer fedv1beta1informers.FederatedGroupInformer,
	multiClusterEnabled bool) *Controller {

	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: k8sClient.CoreV1().Events("")})
	ctl := &Controller{
		BaseController: controller.BaseController{
			Workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Group"),
			Synced:    []cache.InformerSynced{groupInformer.Informer().HasSynced},
			Name:      controllerName,
		},
		recorder:            eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerName}),
		k8sClient:           k8sClient,
		ksClient:            ksClient,
		groupInformer:       groupInformer,
		groupLister:         groupInformer.Lister(),
		multiClusterEnabled: multiClusterEnabled,
	}

	if ctl.multiClusterEnabled {
		ctl.federatedGroupLister = federatedGroupInformer.Lister()
		ctl.Synced = append(ctl.Synced, federatedGroupInformer.Informer().HasSynced)
	}

	ctl.Handler = ctl.reconcile

	klog.Info("Setting up event handlers")
	groupInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctl.Enqueue,
		UpdateFunc: func(old, new interface{}) {
			ctl.Enqueue(new)
		},
		DeleteFunc: ctl.Enqueue,
	})
	return ctl
}

func (c *Controller) Start(stopCh <-chan struct{}) error {
	return c.Run(1, stopCh)
}

// reconcile handles Group informer events, clear up related reource when group is being deleted.
func (c *Controller) reconcile(key string) error {

	group, err := c.groupLister.Get(key)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("group '%s' in work queue no longer exists", key))
			return nil
		}
		klog.Error(err)
		return err
	}
	if group.ObjectMeta.DeletionTimestamp.IsZero() {
		var g *iam1alpha2.Group
		if !sliceutil.HasString(group.Finalizers, finalizer) {
			g = group.DeepCopy()
			g.ObjectMeta.Finalizers = append(g.ObjectMeta.Finalizers, finalizer)
		}

		if c.multiClusterEnabled {
			// Ensure not controlled by Kubefed
			if group.Labels == nil || group.Labels[constants.KubefedManagedLabel] != "false" {
				if g == nil {
					g = group.DeepCopy()
				}
				if g.Labels == nil {
					g.Labels = make(map[string]string, 0)
				}
				g.Labels[constants.KubefedManagedLabel] = "false"
			}
		}

		if g != nil {
			if _, err = c.ksClient.IamV1alpha2().Groups().Update(g); err != nil {
				return err
			}
			// Skip reconcile when group is updated.
			return nil
		}
	} else {
		// The object is being deleted
		if sliceutil.HasString(group.ObjectMeta.Finalizers, finalizer) {
			if err = c.deleteGroupBindings(group); err != nil {
				klog.Error(err)
				return err
			}

			if err = c.deleteRoleBindings(group); err != nil {
				klog.Error(err)
				return err
			}

			group.Finalizers = sliceutil.RemoveString(group.ObjectMeta.Finalizers, func(item string) bool {
				return item == finalizer
			})

			if group, err = c.ksClient.IamV1alpha2().Groups().Update(group); err != nil {
				return err
			}
		}
		return nil
	}

	// synchronization through kubefed-controller when multi cluster is enabled
	if c.multiClusterEnabled {
		if err = c.multiClusterSync(group); err != nil {
			klog.Error(err)
			return err
		}
	}

	c.recorder.Event(group, corev1.EventTypeNormal, successSynced, messageResourceSynced)
	return nil
}

func (c *Controller) deleteGroupBindings(group *iam1alpha2.Group) error {

	// Groupbindings that created by kubesphere will be deleted directly.
	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{iam1alpha2.GroupReferenceLabel: group.Name}).String(),
	}
	deleteOptions := metav1.NewDeleteOptions(0)

	if err := c.ksClient.IamV1alpha2().GroupBindings().
		DeleteCollection(deleteOptions, listOptions); err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

// remove all RoleBindings.
func (c *Controller) deleteRoleBindings(group *iam1alpha2.Group) error {
	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{iam1alpha2.GroupReferenceLabel: group.Name}).String(),
	}
	deleteOptions := metav1.NewDeleteOptions(0)

	if err := c.ksClient.IamV1alpha2().WorkspaceRoleBindings().
		DeleteCollection(deleteOptions, listOptions); err != nil {
		klog.Error(err)
		return err
	}

	if err := c.k8sClient.RbacV1().ClusterRoleBindings().
		DeleteCollection(deleteOptions, listOptions); err != nil {
		klog.Error(err)
		return err
	}

	if result, err := c.k8sClient.CoreV1().Namespaces().List(metav1.ListOptions{}); err != nil {
		klog.Error(err)
		return err
	} else {
		for _, namespace := range result.Items {
			if err = c.k8sClient.RbacV1().RoleBindings(namespace.Name).DeleteCollection(deleteOptions, listOptions); err != nil {
				klog.Error(err)
				return err
			}
		}
	}

	return nil
}

func (c *Controller) multiClusterSync(group *iam1alpha2.Group) error {

	obj, err := c.federatedGroupLister.Get(group.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			return c.createFederatedGroup(group)
		}
		klog.Error(err)
		return err
	}

	if !reflect.DeepEqual(obj.Spec.Template.Labels, group.Labels) {

		obj.Spec.Template.Labels = group.Labels

		if _, err = c.ksClient.TypesV1beta1().FederatedGroups().Update(obj); err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) createFederatedGroup(group *iam1alpha2.Group) error {
	federatedGroup := &fedv1beta1types.FederatedGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name: group.Name,
		},
		Spec: fedv1beta1types.FederatedGroupSpec{
			Template: fedv1beta1types.GroupTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Labels: group.Labels,
				},
				Spec: group.Spec,
			},
			Placement: fedv1beta1types.GenericPlacementFields{
				ClusterSelector: &metav1.LabelSelector{},
			},
		},
	}

	// must bind group lifecycle
	err := controllerutil.SetControllerReference(group, federatedGroup, scheme.Scheme)
	if err != nil {
		return err
	}
	if _, err = c.ksClient.TypesV1beta1().FederatedGroups().Create(federatedGroup); err != nil {
		return err
	}
	return nil
}
