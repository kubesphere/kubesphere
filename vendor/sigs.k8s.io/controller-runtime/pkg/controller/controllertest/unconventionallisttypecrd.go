package controllertest

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ runtime.Object = &UnconventionalListType{}
var _ runtime.Object = &UnconventionalListTypeList{}

// UnconventionalListType is used to test CRDs with List types that
// have a slice of pointers rather than a slice of literals.
type UnconventionalListType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              string `json:"spec,omitempty"`
}

// DeepCopyObject implements runtime.Object
// Handwritten for simplicity.
func (u *UnconventionalListType) DeepCopyObject() runtime.Object {
	return u.DeepCopy()
}

// DeepCopy implements *UnconventionalListType
// Handwritten for simplicity.
func (u *UnconventionalListType) DeepCopy() *UnconventionalListType {
	return &UnconventionalListType{
		TypeMeta:   u.TypeMeta,
		ObjectMeta: *u.ObjectMeta.DeepCopy(),
		Spec:       u.Spec,
	}
}

// UnconventionalListTypeList is used to test CRDs with List types that
// have a slice of pointers rather than a slice of literals.
type UnconventionalListTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []*UnconventionalListType `json:"items"`
}

// DeepCopyObject implements runtime.Object
// Handwritten for simplicity.
func (u *UnconventionalListTypeList) DeepCopyObject() runtime.Object {
	return u.DeepCopy()
}

// DeepCopy implements *UnconventionalListTypeListt
// Handwritten for simplicity.
func (u *UnconventionalListTypeList) DeepCopy() *UnconventionalListTypeList {
	out := &UnconventionalListTypeList{
		TypeMeta: u.TypeMeta,
		ListMeta: *u.ListMeta.DeepCopy(),
	}
	for _, item := range u.Items {
		out.Items = append(out.Items, item.DeepCopy())
	}
	return out
}
