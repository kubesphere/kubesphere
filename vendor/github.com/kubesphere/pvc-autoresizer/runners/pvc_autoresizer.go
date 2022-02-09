package runners

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/kubesphere/pvc-autoresizer/metrics"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=storage.k8s.io,resources=storageclasses,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create

const resizeEnableIndexKey = ".metadata.annotations[resize.kubesphere.io/enabled]"
const storageClassNameIndexKey = ".spec.storageClassName"
const logLevelWarn = 3

// NewPVCAutoresizer returns a new pvcAutoresizer struct
func NewPVCAutoresizer(mc MetricsClient, c client.Client, log logr.Logger, interval time.Duration, recorder record.EventRecorder) manager.Runnable {

	return &pvcAutoresizer{
		metricsClient: mc,
		client:        c,
		log:           log,
		interval:      interval,
		recorder:      recorder,
	}
}

type pvcAutoresizer struct {
	client        client.Client
	metricsClient MetricsClient
	interval      time.Duration
	log           logr.Logger
	recorder      record.EventRecorder
}

// Start implements manager.Runnable
func (w *pvcAutoresizer) Start(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)

	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			startTime := time.Now()
			err := w.reconcile(ctx)
			metrics.ResizerLoopSecondsTotal.Add(time.Since(startTime).Seconds())
			if err != nil {
				w.log.Error(err, "failed to reconcile")
				return err
			}
		}
	}
}

func isTargetPVC(pvc *corev1.PersistentVolumeClaim, sc *storagev1.StorageClass) (bool, error) {
	quantity, err := pvcStorageLimit(pvc, sc)
	if err != nil {
		return false, fmt.Errorf("invalid storage limit: %w", err)
	}
	if quantity.IsZero() {
		return false, nil
	}
	if pvc.Spec.VolumeMode != nil && *pvc.Spec.VolumeMode != corev1.PersistentVolumeFilesystem {
		return false, nil
	}
	if pvc.Status.Phase != corev1.ClaimBound {
		return false, nil
	}
	return true, nil
}

func (w *pvcAutoresizer) getStorageClassList(ctx context.Context) (*storagev1.StorageClassList, error) {
	var scs storagev1.StorageClassList
	err := w.client.List(ctx, &scs, client.MatchingFields(map[string]string{resizeEnableIndexKey: "true"}))
	if err != nil {
		metrics.KubernetesClientFailTotal.Increment()
		return nil, err
	}
	return &scs, nil
}

func (w *pvcAutoresizer) reconcile(ctx context.Context) error {
	scs, err := w.getStorageClassList(ctx)
	if err != nil {
		w.log.Error(err, "getStorageClassList failed")
		return nil
	}

	vsMap, err := w.metricsClient.GetMetrics(ctx)
	if err != nil {
		w.log.Error(err, "metricsClient.GetMetrics failed")
		return nil
	}

	for _, sc := range scs.Items {
		var pvcs corev1.PersistentVolumeClaimList
		err = w.client.List(ctx, &pvcs, client.MatchingFields(map[string]string{storageClassNameIndexKey: sc.Name}))
		if err != nil {
			metrics.KubernetesClientFailTotal.Increment()
			w.log.Error(err, "list pvc failed")
			return nil
		}
		for _, pvc := range pvcs.Items {
			isTarget, err := isTargetPVC(&pvc, &sc)
			if err != nil {
				metrics.ResizerFailedResizeTotal.Increment()
				w.log.WithValues("namespace", pvc.Namespace, "name", pvc.Name).Error(err, "failed to check target PVC")
				continue
			} else if !isTarget {
				continue
			}
			namespacedName := types.NamespacedName{
				Namespace: pvc.Namespace,
				Name:      pvc.Name,
			}
			if _, ok := vsMap[namespacedName]; !ok {
				continue
			}
			err = w.resize(ctx, &pvc, vsMap[namespacedName], &sc)
			if err != nil {
				metrics.ResizerFailedResizeTotal.Increment()
				w.log.WithValues("namespace", pvc.Namespace, "name", pvc.Name).Error(err, "failed to resize PVC")
			}
		}
	}

	return nil
}

func (w *pvcAutoresizer) resize(ctx context.Context, pvc *corev1.PersistentVolumeClaim, vs *VolumeStats, sc *storagev1.StorageClass) error {
	log := w.log.WithName("resize").WithValues("namespace", pvc.Namespace, "name", pvc.Name)

	var resizeThreshold string
	if annotation, ok := pvc.Annotations[ResizeThresholdAnnotation]; ok && annotation != "" {
		resizeThreshold = annotation
	} else {
		resizeThreshold = sc.Annotations[ResizeThresholdAnnotation]
	}
	threshold, err := convertSizeInBytes(resizeThreshold, vs.CapacityBytes, DefaultThreshold)
	if err != nil {
		log.V(logLevelWarn).Info("failed to convert threshold annotation", "error", err.Error())
		// lint:ignore nilerr ignores this because invalid annotations should be allowed.
		return nil
	}

	inodesThreshold, err := convertSize(pvc.Annotations[ResizeInodesThresholdAnnotation], vs.CapacityInodeSize, DefaultInodesThreshold)
	if err != nil {
		log.V(logLevelWarn).Info("failed to convert threshold annotation", "error", err.Error())
		// lint:ignore nilerr ignores this because invalid annotations should be allowed.
		return nil
	}

	curReq := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	var resizeIncrease string
	if annotation, ok := pvc.Annotations[ResizeIncreaseAnnotation]; ok && annotation != "" {
		resizeIncrease = annotation
	} else {
		resizeIncrease = sc.Annotations[ResizeIncreaseAnnotation]
	}
	increase, err := convertSizeInBytes(resizeIncrease, curReq.Value(), DefaultIncrease)
	if err != nil {
		log.V(logLevelWarn).Info("failed to convert increase annotation", "error", err.Error())
		// lint:ignore nilerr ignores this because invalid annotations should be allowed.
		return nil
	}

	preCap, exist := pvc.Annotations[PreviousCapacityBytesAnnotation]
	if exist {
		preCapInt64, err := strconv.ParseInt(preCap, 10, 64)
		if err != nil {
			log.V(logLevelWarn).Info("failed to parse pre_cap_bytes annotation", "error", err.Error())
			// lint:ignore nilerr ignores this because invalid annotations should be allowed.
			return nil
		}
		if preCapInt64 == vs.CapacityBytes {
			log.Info("waiting for resizing...", "capacity", vs.CapacityBytes)
			return nil
		}
	}
	limitRes, err := pvcStorageLimit(pvc, sc)
	if err != nil {
		log.Error(err, "fetching storage limit failed")
		return err
	}
	if curReq.Cmp(limitRes) == 0 {
		log.Info("volume storage limit reached")
		metrics.ResizerLimitReachedTotal.Increment()
		return nil
	}

	if threshold > vs.AvailableBytes || inodesThreshold > vs.AvailableInodeSize {
		if pvc.Annotations == nil {
			pvc.Annotations = make(map[string]string)
		}
		newReqBytes := int64(math.Ceil(float64(curReq.Value()+increase)/(1<<30))) << 30
		newReq := resource.NewQuantity(newReqBytes, resource.BinarySI)
		if newReq.Cmp(limitRes) > 0 {
			newReq = &limitRes
		}

		pvc.Spec.Resources.Requests[corev1.ResourceStorage] = *newReq
		pvc.Annotations[PreviousCapacityBytesAnnotation] = strconv.FormatInt(vs.CapacityBytes, 10)
		err = w.client.Update(ctx, pvc)
		if err != nil {
			metrics.KubernetesClientFailTotal.Increment()
			return err
		}
		log.Info("resize started",
			"from", curReq.Value(),
			"to", newReq.Value(),
			"threshold", threshold,
			"available", vs.AvailableBytes,
			"inodesThreshold", inodesThreshold,
			"inodesAvailable", vs.AvailableInodeSize,
		)
		w.recorder.Eventf(pvc, corev1.EventTypeNormal, "Resized", "PVC volume is resized to %s", newReq.String())
		metrics.ResizerSuccessResizeTotal.Increment()
	}

	return nil
}

func indexByResizeEnableAnnotation(obj client.Object) []string {
	sc := obj.(*storagev1.StorageClass)
	if val, ok := sc.Annotations[AutoResizeEnabledKey]; ok {
		return []string{val}
	}

	return []string{}
}

func indexByStorageClassName(obj client.Object) []string {
	pvc := obj.(*corev1.PersistentVolumeClaim)
	scName := pvc.Spec.StorageClassName
	if scName == nil {
		return []string{}
	}
	return []string{*scName}
}

// SetupIndexer setup indices for PVC auto resizer
func SetupIndexer(mgr ctrl.Manager, skipAnnotationCheck bool) error {
	idxFunc := indexByResizeEnableAnnotation
	if skipAnnotationCheck {
		idxFunc = func(_ client.Object) []string { return []string{"true"} }
	}
	err := mgr.GetFieldIndexer().IndexField(context.Background(), &storagev1.StorageClass{}, resizeEnableIndexKey, idxFunc)
	if err != nil {
		return err
	}

	err = mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.PersistentVolumeClaim{}, storageClassNameIndexKey, indexByStorageClassName)
	if err != nil {
		return err
	}

	return nil
}

func convertSizeInBytes(valStr string, capacity int64, defaultVal string) (int64, error) {
	if len(valStr) == 0 {
		valStr = defaultVal
	}
	if strings.HasSuffix(valStr, "%") {
		return calcSize(valStr, capacity)
	}

	quantity, err := resource.ParseQuantity(valStr)
	if err != nil {
		return 0, err
	}
	val := quantity.Value()
	if val <= 0 {
		return 0, fmt.Errorf("annotation value should be positive: %s", valStr)
	}
	return val, nil
}

func convertSize(valStr string, capacity int64, defaultVal string) (int64, error) {
	if len(valStr) == 0 {
		valStr = defaultVal
	}
	if strings.HasSuffix(valStr, "%") {
		return calcSize(valStr, capacity)
	}
	return 0, fmt.Errorf("annotation value should be in percent notation: %s", valStr)
}

func calcSize(valStr string, capacity int64) (int64, error) {
	rate, err := strconv.ParseFloat(strings.TrimRight(valStr, "%"), 64)
	if err != nil {
		return 0, err
	}
	if rate < 0 || rate > 100 {
		return 0, fmt.Errorf("annotation value should between 0 and 100: %s", valStr)
	}

	res := int64(float64(capacity) * rate / 100.0)
	return res, nil
}

func pvcStorageLimit(pvc *corev1.PersistentVolumeClaim, sc *storagev1.StorageClass) (resource.Quantity, error) {
	// storage limit on the annotation has precedence
	if annotation, ok := sc.Annotations[StorageLimitAnnotation]; ok && annotation != "" {
		return resource.ParseQuantity(annotation)
	} else if annotation, ok := pvc.Annotations[StorageLimitAnnotation]; ok && annotation != "" {
		return resource.ParseQuantity(annotation)
	}

	// Storage() returns 0 valued Quantity if Limits does not set
	return *pvc.Spec.Resources.Limits.Storage(), nil
}
