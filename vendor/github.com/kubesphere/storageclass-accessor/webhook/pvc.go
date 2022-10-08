package webhook

import (
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type ReqInfo struct {
	resource         string
	name             string
	namespace        string
	operator         string
	storageClassName string
}

var reviewResponse = &admissionv1.AdmissionResponse{
	Allowed: true,
	Result:  &metav1.Status{},
}

func AdmitPVC(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	klog.Info("admitting pvc")

	if ar.Request.Operation != admissionv1.Create {
		return reviewResponse
	}

	raw := ar.Request.Object.Raw

	var newPVC *corev1.PersistentVolumeClaim

	deserializer := codecs.UniversalDeserializer()
	pvc := &corev1.PersistentVolumeClaim{}
	obj, _, err := deserializer.Decode(raw, nil, pvc)
	if err != nil {
		klog.Error(err)
		return toV1AdmissionResponse(err)
	}
	var ok bool
	newPVC, ok = obj.(*corev1.PersistentVolumeClaim)
	if !ok {
		klog.Error("obj can't exchange to pvc object")
		return toV1AdmissionResponse(err)
	}

	reqPVC := ReqInfo{
		resource:         "persistentVolumeClaim",
		name:             newPVC.Name,
		namespace:        newPVC.Namespace,
		operator:         string(ar.Request.Operation),
		storageClassName: *newPVC.Spec.StorageClassName,
	}
	return DecidePVCV1(reqPVC)
}

func DecidePVCV1(pvc ReqInfo) *admissionv1.AdmissionResponse {

	accessors, err := getAccessors(pvc.storageClassName)

	if err != nil {
		klog.Error("get accessor failed, err:", err)
		return toV1AdmissionResponse(err)
	} else if len(accessors) == 0 {
		klog.Info("Not Found accessor for the storageClass:", pvc.storageClassName)
		return reviewResponse
	}

	for _, accessor := range accessors {
		if err = validateNameSpace(pvc, accessor); err != nil {
			return toV1AdmissionResponse(err)
		}
		if err = validateWorkSpace(pvc, accessor); err != nil {
			return toV1AdmissionResponse(err)
		}
	}
	return reviewResponse
}
