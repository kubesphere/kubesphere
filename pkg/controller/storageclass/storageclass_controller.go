/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package storageclass

import (
	"context"
	"reflect"
	"strconv"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	annotationAllowSnapshot = "storageclass.kubesphere.io/allow-snapshot"
	annotationAllowClone    = "storageclass.kubesphere.io/allow-clone"
	controllerName          = "storageclass-capability"
	pvcCountAnnotation      = "kubesphere.io/pvc-count"
)

var _ kscontroller.Controller = &Reconciler{}
var _ reconcile.Reconciler = &Reconciler{}

// This controller is responsible to watch StorageClass and CSIDriver.
// And then update StorageClass CRD resource object to the newest status.

type Reconciler struct {
	client.Client
}

func (r *Reconciler) Name() string {
	return controllerName
}

func (r *Reconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	r.Client = mgr.GetClient()

	return builder.ControllerManagedBy(mgr).
		For(&storagev1.StorageClass{},
			builder.WithPredicates(
				predicate.ResourceVersionChangedPredicate{},
			),
		).
		Watches(
			&corev1.PersistentVolumeClaim{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				pvc := obj.(*corev1.PersistentVolumeClaim)
				var storageClassName string
				if pvc.Spec.StorageClassName != nil {
					storageClassName = *pvc.Spec.StorageClassName
				} else if pvc.Annotations[corev1.BetaStorageClassAnnotation] != "" {
					storageClassName = pvc.Annotations[corev1.BetaStorageClassAnnotation]
				}
				return []reconcile.Request{{NamespacedName: types.NamespacedName{Name: storageClassName}}}
			}),
			builder.WithPredicates(predicate.Funcs{
				GenericFunc: func(genericEvent event.GenericEvent) bool {
					return false
				},
				CreateFunc: func(event event.CreateEvent) bool {
					return true
				},
				UpdateFunc: func(updateEvent event.UpdateEvent) bool {
					return false
				},
				DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
					return true
				},
			}),
		).
		Watches(
			&storagev1.CSIDriver{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				storageClassList := &storagev1.StorageClassList{}
				if err := r.List(context.Background(), storageClassList); err != nil {
					klog.Errorf("list StorageClass failed: %v", err)
					return nil
				}
				csiDriver := obj.(*storagev1.CSIDriver)
				requests := make([]reconcile.Request, 0)
				for _, storageClass := range storageClassList.Items {
					if storageClass.Provisioner == csiDriver.Name {
						requests = append(requests, reconcile.Request{
							NamespacedName: types.NamespacedName{
								Name: storageClass.Name,
							},
						})
					}
				}
				return requests
			}),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		WithOptions(controller.Options{MaxConcurrentReconciles: 2}).
		Named(controllerName).Complete(r)
}

// When creating a new storage class, the controller will create a new storage capability object.
// When updating storage class, the controller will update or create the storage capability object.
// When deleting storage class, the controller will delete storage capability object.

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	storageClass := &storagev1.StorageClass{}
	if err := r.Get(ctx, req.NamespacedName, storageClass); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// Cloning and volumeSnapshot support only available for CSI drivers.
	isCSIStorage := r.hasCSIDriver(ctx, storageClass)
	// Annotate storageClass
	storageClassUpdated := storageClass.DeepCopy()
	if isCSIStorage {
		r.updateSnapshotAnnotation(storageClassUpdated, isCSIStorage)
		r.updateCloneVolumeAnnotation(storageClassUpdated, isCSIStorage)
	} else {
		r.removeAnnotations(storageClassUpdated)
	}

	pvcCount, err := r.countPersistentVolumeClaims(ctx, storageClass)
	if err != nil {
		return ctrl.Result{}, err
	}
	if storageClassUpdated.Annotations == nil {
		storageClassUpdated.Annotations = make(map[string]string)
	}
	storageClassUpdated.Annotations[pvcCountAnnotation] = strconv.Itoa(pvcCount)
	if !reflect.DeepEqual(storageClass, storageClassUpdated) {
		return ctrl.Result{}, r.Update(ctx, storageClassUpdated)
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) hasCSIDriver(ctx context.Context, storageClass *storagev1.StorageClass) bool {
	driver := storageClass.Provisioner
	if driver != "" {
		if err := r.Get(ctx, client.ObjectKey{Name: driver}, &storagev1.CSIDriver{}); err != nil {
			return false
		}
		return true
	}
	return false
}

func (r *Reconciler) updateSnapshotAnnotation(storageClass *storagev1.StorageClass, snapshotAllow bool) {
	if storageClass.Annotations == nil {
		storageClass.Annotations = make(map[string]string)
	}
	if _, err := strconv.ParseBool(storageClass.Annotations[annotationAllowSnapshot]); err != nil {
		storageClass.Annotations[annotationAllowSnapshot] = strconv.FormatBool(snapshotAllow)
	}
}

func (r *Reconciler) updateCloneVolumeAnnotation(storageClass *storagev1.StorageClass, cloneAllow bool) {
	if storageClass.Annotations == nil {
		storageClass.Annotations = make(map[string]string)
	}
	if _, err := strconv.ParseBool(storageClass.Annotations[annotationAllowClone]); err != nil {
		storageClass.Annotations[annotationAllowClone] = strconv.FormatBool(cloneAllow)
	}
}

func (r *Reconciler) removeAnnotations(storageClass *storagev1.StorageClass) {
	delete(storageClass.Annotations, annotationAllowClone)
	delete(storageClass.Annotations, annotationAllowSnapshot)
}

func (r *Reconciler) countPersistentVolumeClaims(ctx context.Context, storageClass *storagev1.StorageClass) (int, error) {
	pvcs := &corev1.PersistentVolumeClaimList{}
	if err := r.List(ctx, pvcs); err != nil {
		return 0, err
	}
	var count int
	for _, pvc := range pvcs.Items {
		if (pvc.Spec.StorageClassName != nil && *pvc.Spec.StorageClassName == storageClass.Name) ||
			(pvc.Annotations != nil && pvc.Annotations[corev1.BetaStorageClassAnnotation] == storageClass.Name) {
			count++
		}
	}
	return count, nil
}
