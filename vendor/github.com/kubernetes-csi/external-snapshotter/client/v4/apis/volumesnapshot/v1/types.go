/*
Copyright 2020 The Kubernetes Authors.

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

// +kubebuilder:object:generate=true
package v1

import (
	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeSnapshot is a user's request for either creating a point-in-time
// snapshot of a persistent volume, or binding to a pre-existing snapshot.
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ReadyToUse",type=boolean,JSONPath=`.status.readyToUse`,description="Indicates if the snapshot is ready to be used to restore a volume."
// +kubebuilder:printcolumn:name="SourcePVC",type=string,JSONPath=`.spec.source.persistentVolumeClaimName`,description="If a new snapshot needs to be created, this contains the name of the source PVC from which this snapshot was (or will be) created."
// +kubebuilder:printcolumn:name="SourceSnapshotContent",type=string,JSONPath=`.spec.source.volumeSnapshotContentName`,description="If a snapshot already exists, this contains the name of the existing VolumeSnapshotContent object representing the existing snapshot."
// +kubebuilder:printcolumn:name="RestoreSize",type=string,JSONPath=`.status.restoreSize`,description="Represents the minimum size of volume required to rehydrate from this snapshot."
// +kubebuilder:printcolumn:name="SnapshotClass",type=string,JSONPath=`.spec.volumeSnapshotClassName`,description="The name of the VolumeSnapshotClass requested by the VolumeSnapshot."
// +kubebuilder:printcolumn:name="SnapshotContent",type=string,JSONPath=`.status.boundVolumeSnapshotContentName`,description="Name of the VolumeSnapshotContent object to which the VolumeSnapshot object intends to bind to. Please note that verification of binding actually requires checking both VolumeSnapshot and VolumeSnapshotContent to ensure both are pointing at each other. Binding MUST be verified prior to usage of this object."
// +kubebuilder:printcolumn:name="CreationTime",type=date,JSONPath=`.status.creationTime`,description="Timestamp when the point-in-time snapshot was taken by the underlying storage system."
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type VolumeSnapshot struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// spec defines the desired characteristics of a snapshot requested by a user.
	// More info: https://kubernetes.io/docs/concepts/storage/volume-snapshots#volumesnapshots
	// Required.
	Spec VolumeSnapshotSpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`

	// status represents the current information of a snapshot.
	// Consumers must verify binding between VolumeSnapshot and
	// VolumeSnapshotContent objects is successful (by validating that both
	// VolumeSnapshot and VolumeSnapshotContent point at each other) before
	// using this object.
	// +optional
	Status *VolumeSnapshotStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// VolumeSnapshotList is a list of VolumeSnapshot objects
type VolumeSnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// List of VolumeSnapshots
	Items []VolumeSnapshot `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// VolumeSnapshotSpec describes the common attributes of a volume snapshot.
type VolumeSnapshotSpec struct {
	// source specifies where a snapshot will be created from.
	// This field is immutable after creation.
	// Required.
	Source VolumeSnapshotSource `json:"source" protobuf:"bytes,1,opt,name=source"`

	// VolumeSnapshotClassName is the name of the VolumeSnapshotClass
	// requested by the VolumeSnapshot.
	// VolumeSnapshotClassName may be left nil to indicate that the default
	// SnapshotClass should be used.
	// A given cluster may have multiple default Volume SnapshotClasses: one
	// default per CSI Driver. If a VolumeSnapshot does not specify a SnapshotClass,
	// VolumeSnapshotSource will be checked to figure out what the associated
	// CSI Driver is, and the default VolumeSnapshotClass associated with that
	// CSI Driver will be used. If more than one VolumeSnapshotClass exist for
	// a given CSI Driver and more than one have been marked as default,
	// CreateSnapshot will fail and generate an event.
	// Empty string is not allowed for this field.
	// +optional
	VolumeSnapshotClassName *string `json:"volumeSnapshotClassName,omitempty" protobuf:"bytes,2,opt,name=volumeSnapshotClassName"`
}

// VolumeSnapshotSource specifies whether the underlying snapshot should be
// dynamically taken upon creation or if a pre-existing VolumeSnapshotContent
// object should be used.
// Exactly one of its members must be set.
// Members in VolumeSnapshotSource are immutable.
type VolumeSnapshotSource struct {
	// persistentVolumeClaimName specifies the name of the PersistentVolumeClaim
	// object representing the volume from which a snapshot should be created.
	// This PVC is assumed to be in the same namespace as the VolumeSnapshot
	// object.
	// This field should be set if the snapshot does not exists, and needs to be
	// created.
	// This field is immutable.
	// +optional
	PersistentVolumeClaimName *string `json:"persistentVolumeClaimName,omitempty" protobuf:"bytes,1,opt,name=persistentVolumeClaimName"`

	// volumeSnapshotContentName specifies the name of a pre-existing VolumeSnapshotContent
	// object representing an existing volume snapshot.
	// This field should be set if the snapshot already exists and only needs a representation in Kubernetes.
	// This field is immutable.
	// +optional
	VolumeSnapshotContentName *string `json:"volumeSnapshotContentName,omitempty" protobuf:"bytes,2,opt,name=volumeSnapshotContentName"`
}

// VolumeSnapshotStatus is the status of the VolumeSnapshot
// Note that CreationTime, RestoreSize, ReadyToUse, and Error are in both
// VolumeSnapshotStatus and VolumeSnapshotContentStatus. Fields in VolumeSnapshotStatus
// are updated based on fields in VolumeSnapshotContentStatus. They are eventual
// consistency. These fields are duplicate in both objects due to the following reasons:
// - Fields in VolumeSnapshotContentStatus can be used for filtering when importing a
//   volumesnapshot.
// - VolumsnapshotStatus is used by end users because they cannot see VolumeSnapshotContent.
// - CSI snapshotter sidecar is light weight as it only watches VolumeSnapshotContent
//   object, not VolumeSnapshot object.
type VolumeSnapshotStatus struct {
	// boundVolumeSnapshotContentName is the name of the VolumeSnapshotContent
	// object to which this VolumeSnapshot object intends to bind to.
	// If not specified, it indicates that the VolumeSnapshot object has not been
	// successfully bound to a VolumeSnapshotContent object yet.
	// NOTE: To avoid possible security issues, consumers must verify binding between
	// VolumeSnapshot and VolumeSnapshotContent objects is successful (by validating that
	// both VolumeSnapshot and VolumeSnapshotContent point at each other) before using
	// this object.
	// +optional
	BoundVolumeSnapshotContentName *string `json:"boundVolumeSnapshotContentName,omitempty" protobuf:"bytes,1,opt,name=boundVolumeSnapshotContentName"`

	// creationTime is the timestamp when the point-in-time snapshot is taken
	// by the underlying storage system.
	// In dynamic snapshot creation case, this field will be filled in by the
	// snapshot controller with the "creation_time" value returned from CSI
	// "CreateSnapshot" gRPC call.
	// For a pre-existing snapshot, this field will be filled with the "creation_time"
	// value returned from the CSI "ListSnapshots" gRPC call if the driver supports it.
	// If not specified, it may indicate that the creation time of the snapshot is unknown.
	// +optional
	CreationTime *metav1.Time `json:"creationTime,omitempty" protobuf:"bytes,2,opt,name=creationTime"`

	// readyToUse indicates if the snapshot is ready to be used to restore a volume.
	// In dynamic snapshot creation case, this field will be filled in by the
	// snapshot controller with the "ready_to_use" value returned from CSI
	// "CreateSnapshot" gRPC call.
	// For a pre-existing snapshot, this field will be filled with the "ready_to_use"
	// value returned from the CSI "ListSnapshots" gRPC call if the driver supports it,
	// otherwise, this field will be set to "True".
	// If not specified, it means the readiness of a snapshot is unknown.
	// +optional
	ReadyToUse *bool `json:"readyToUse,omitempty" protobuf:"varint,3,opt,name=readyToUse"`

	// restoreSize represents the minimum size of volume required to create a volume
	// from this snapshot.
	// In dynamic snapshot creation case, this field will be filled in by the
	// snapshot controller with the "size_bytes" value returned from CSI
	// "CreateSnapshot" gRPC call.
	// For a pre-existing snapshot, this field will be filled with the "size_bytes"
	// value returned from the CSI "ListSnapshots" gRPC call if the driver supports it.
	// When restoring a volume from this snapshot, the size of the volume MUST NOT
	// be smaller than the restoreSize if it is specified, otherwise the restoration will fail.
	// If not specified, it indicates that the size is unknown.
	// +optional
	RestoreSize *resource.Quantity `json:"restoreSize,omitempty" protobuf:"bytes,4,opt,name=restoreSize"`

	// error is the last observed error during snapshot creation, if any.
	// This field could be helpful to upper level controllers(i.e., application controller)
	// to decide whether they should continue on waiting for the snapshot to be created
	// based on the type of error reported.
	// The snapshot controller will keep retrying when an error occurrs during the
	// snapshot creation. Upon success, this error field will be cleared.
	// +optional
	Error *VolumeSnapshotError `json:"error,omitempty" protobuf:"bytes,5,opt,name=error,casttype=VolumeSnapshotError"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeSnapshotClass specifies parameters that a underlying storage system uses when
// creating a volume snapshot. A specific VolumeSnapshotClass is used by specifying its
// name in a VolumeSnapshot object.
// VolumeSnapshotClasses are non-namespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Driver",type=string,JSONPath=`.driver`
// +kubebuilder:printcolumn:name="DeletionPolicy",type=string,JSONPath=`.deletionPolicy`,description="Determines whether a VolumeSnapshotContent created through the VolumeSnapshotClass should be deleted when its bound VolumeSnapshot is deleted."
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type VolumeSnapshotClass struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// driver is the name of the storage driver that handles this VolumeSnapshotClass.
	// Required.
	Driver string `json:"driver" protobuf:"bytes,2,opt,name=driver"`

	// parameters is a key-value map with storage driver specific parameters for creating snapshots.
	// These values are opaque to Kubernetes.
	// +optional
	Parameters map[string]string `json:"parameters,omitempty" protobuf:"bytes,3,rep,name=parameters"`

	// deletionPolicy determines whether a VolumeSnapshotContent created through
	// the VolumeSnapshotClass should be deleted when its bound VolumeSnapshot is deleted.
	// Supported values are "Retain" and "Delete".
	// "Retain" means that the VolumeSnapshotContent and its physical snapshot on underlying storage system are kept.
	// "Delete" means that the VolumeSnapshotContent and its physical snapshot on underlying storage system are deleted.
	// Required.
	DeletionPolicy DeletionPolicy `json:"deletionPolicy" protobuf:"bytes,4,opt,name=deletionPolicy"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeSnapshotClassList is a collection of VolumeSnapshotClasses.
// +kubebuilder:object:root=true
type VolumeSnapshotClassList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// items is the list of VolumeSnapshotClasses
	Items []VolumeSnapshotClass `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeSnapshotContent represents the actual "on-disk" snapshot object in the
// underlying storage system
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ReadyToUse",type=boolean,JSONPath=`.status.readyToUse`,description="Indicates if the snapshot is ready to be used to restore a volume."
// +kubebuilder:printcolumn:name="RestoreSize",type=integer,JSONPath=`.status.restoreSize`,description="Represents the complete size of the snapshot in bytes"
// +kubebuilder:printcolumn:name="DeletionPolicy",type=string,JSONPath=`.spec.deletionPolicy`,description="Determines whether this VolumeSnapshotContent and its physical snapshot on the underlying storage system should be deleted when its bound VolumeSnapshot is deleted."
// +kubebuilder:printcolumn:name="Driver",type=string,JSONPath=`.spec.driver`,description="Name of the CSI driver used to create the physical snapshot on the underlying storage system."
// +kubebuilder:printcolumn:name="VolumeSnapshotClass",type=string,JSONPath=`.spec.volumeSnapshotClassName`,description="Name of the VolumeSnapshotClass to which this snapshot belongs."
// +kubebuilder:printcolumn:name="VolumeSnapshot",type=string,JSONPath=`.spec.volumeSnapshotRef.name`,description="Name of the VolumeSnapshot object to which this VolumeSnapshotContent object is bound."
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type VolumeSnapshotContent struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// spec defines properties of a VolumeSnapshotContent created by the underlying storage system.
	// Required.
	Spec VolumeSnapshotContentSpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`

	// status represents the current information of a snapshot.
	// +optional
	Status *VolumeSnapshotContentStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeSnapshotContentList is a list of VolumeSnapshotContent objects
// +kubebuilder:object:root=true
type VolumeSnapshotContentList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// items is the list of VolumeSnapshotContents
	Items []VolumeSnapshotContent `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// VolumeSnapshotContentSpec is the specification of a VolumeSnapshotContent
type VolumeSnapshotContentSpec struct {
	// volumeSnapshotRef specifies the VolumeSnapshot object to which this
	// VolumeSnapshotContent object is bound.
	// VolumeSnapshot.Spec.VolumeSnapshotContentName field must reference to
	// this VolumeSnapshotContent's name for the bidirectional binding to be valid.
	// For a pre-existing VolumeSnapshotContent object, name and namespace of the
	// VolumeSnapshot object MUST be provided for binding to happen.
	// This field is immutable after creation.
	// Required.
	VolumeSnapshotRef core_v1.ObjectReference `json:"volumeSnapshotRef" protobuf:"bytes,1,opt,name=volumeSnapshotRef"`

	// deletionPolicy determines whether this VolumeSnapshotContent and its physical snapshot on
	// the underlying storage system should be deleted when its bound VolumeSnapshot is deleted.
	// Supported values are "Retain" and "Delete".
	// "Retain" means that the VolumeSnapshotContent and its physical snapshot on underlying storage system are kept.
	// "Delete" means that the VolumeSnapshotContent and its physical snapshot on underlying storage system are deleted.
	// For dynamically provisioned snapshots, this field will automatically be filled in by the
	// CSI snapshotter sidecar with the "DeletionPolicy" field defined in the corresponding
	// VolumeSnapshotClass.
	// For pre-existing snapshots, users MUST specify this field when creating the
	//  VolumeSnapshotContent object.
	// Required.
	DeletionPolicy DeletionPolicy `json:"deletionPolicy" protobuf:"bytes,2,opt,name=deletionPolicy"`

	// driver is the name of the CSI driver used to create the physical snapshot on
	// the underlying storage system.
	// This MUST be the same as the name returned by the CSI GetPluginName() call for
	// that driver.
	// Required.
	Driver string `json:"driver" protobuf:"bytes,3,opt,name=driver"`

	// name of the VolumeSnapshotClass from which this snapshot was (or will be)
	// created.
	// Note that after provisioning, the VolumeSnapshotClass may be deleted or
	// recreated with different set of values, and as such, should not be referenced
	// post-snapshot creation.
	// +optional
	VolumeSnapshotClassName *string `json:"volumeSnapshotClassName,omitempty" protobuf:"bytes,4,opt,name=volumeSnapshotClassName"`

	// source specifies whether the snapshot is (or should be) dynamically provisioned
	// or already exists, and just requires a Kubernetes object representation.
	// This field is immutable after creation.
	// Required.
	Source VolumeSnapshotContentSource `json:"source" protobuf:"bytes,5,opt,name=source"`
}

// VolumeSnapshotContentSource represents the CSI source of a snapshot.
// Exactly one of its members must be set.
// Members in VolumeSnapshotContentSource are immutable.
// TODO(xiangqian): Add a webhook to ensure that VolumeSnapshotContentSource members
// will be immutable once specified.
type VolumeSnapshotContentSource struct {
	// volumeHandle specifies the CSI "volume_id" of the volume from which a snapshot
	// should be dynamically taken from.
	// This field is immutable.
	// +optional
	VolumeHandle *string `json:"volumeHandle,omitempty" protobuf:"bytes,1,opt,name=volumeHandle"`

	// snapshotHandle specifies the CSI "snapshot_id" of a pre-existing snapshot on
	// the underlying storage system for which a Kubernetes object representation
	// was (or should be) created.
	// This field is immutable.
	// +optional
	SnapshotHandle *string `json:"snapshotHandle,omitempty" protobuf:"bytes,2,opt,name=snapshotHandle"`
}

// VolumeSnapshotContentStatus is the status of a VolumeSnapshotContent object
// Note that CreationTime, RestoreSize, ReadyToUse, and Error are in both
// VolumeSnapshotStatus and VolumeSnapshotContentStatus. Fields in VolumeSnapshotStatus
// are updated based on fields in VolumeSnapshotContentStatus. They are eventual
// consistency. These fields are duplicate in both objects due to the following reasons:
// - Fields in VolumeSnapshotContentStatus can be used for filtering when importing a
//   volumesnapshot.
// - VolumsnapshotStatus is used by end users because they cannot see VolumeSnapshotContent.
// - CSI snapshotter sidecar is light weight as it only watches VolumeSnapshotContent
//   object, not VolumeSnapshot object.
type VolumeSnapshotContentStatus struct {
	// snapshotHandle is the CSI "snapshot_id" of a snapshot on the underlying storage system.
	// If not specified, it indicates that dynamic snapshot creation has either failed
	// or it is still in progress.
	// +optional
	SnapshotHandle *string `json:"snapshotHandle,omitempty" protobuf:"bytes,1,opt,name=snapshotHandle"`

	// creationTime is the timestamp when the point-in-time snapshot is taken
	// by the underlying storage system.
	// In dynamic snapshot creation case, this field will be filled in by the
	// CSI snapshotter sidecar with the "creation_time" value returned from CSI
	// "CreateSnapshot" gRPC call.
	// For a pre-existing snapshot, this field will be filled with the "creation_time"
	// value returned from the CSI "ListSnapshots" gRPC call if the driver supports it.
	// If not specified, it indicates the creation time is unknown.
	// The format of this field is a Unix nanoseconds time encoded as an int64.
	// On Unix, the command `date +%s%N` returns the current time in nanoseconds
	// since 1970-01-01 00:00:00 UTC.
	// +optional
	CreationTime *int64 `json:"creationTime,omitempty" protobuf:"varint,2,opt,name=creationTime"`

	// restoreSize represents the complete size of the snapshot in bytes.
	// In dynamic snapshot creation case, this field will be filled in by the
	// CSI snapshotter sidecar with the "size_bytes" value returned from CSI
	// "CreateSnapshot" gRPC call.
	// For a pre-existing snapshot, this field will be filled with the "size_bytes"
	// value returned from the CSI "ListSnapshots" gRPC call if the driver supports it.
	// When restoring a volume from this snapshot, the size of the volume MUST NOT
	// be smaller than the restoreSize if it is specified, otherwise the restoration will fail.
	// If not specified, it indicates that the size is unknown.
	// +kubebuilder:validation:Minimum=0
	// +optional
	RestoreSize *int64 `json:"restoreSize,omitempty" protobuf:"bytes,3,opt,name=restoreSize"`

	// readyToUse indicates if a snapshot is ready to be used to restore a volume.
	// In dynamic snapshot creation case, this field will be filled in by the
	// CSI snapshotter sidecar with the "ready_to_use" value returned from CSI
	// "CreateSnapshot" gRPC call.
	// For a pre-existing snapshot, this field will be filled with the "ready_to_use"
	// value returned from the CSI "ListSnapshots" gRPC call if the driver supports it,
	// otherwise, this field will be set to "True".
	// If not specified, it means the readiness of a snapshot is unknown.
	// +optional.
	ReadyToUse *bool `json:"readyToUse,omitempty" protobuf:"varint,4,opt,name=readyToUse"`

	// error is the last observed error during snapshot creation, if any.
	// Upon success after retry, this error field will be cleared.
	// +optional
	Error *VolumeSnapshotError `json:"error,omitempty" protobuf:"bytes,5,opt,name=error,casttype=VolumeSnapshotError"`
}

// DeletionPolicy describes a policy for end-of-life maintenance of volume snapshot contents
// +kubebuilder:validation:Enum=Delete;Retain
type DeletionPolicy string

const (
	// volumeSnapshotContentDelete means the snapshot will be deleted from the
	// underlying storage system on release from its volume snapshot.
	VolumeSnapshotContentDelete DeletionPolicy = "Delete"

	// volumeSnapshotContentRetain means the snapshot will be left in its current
	// state on release from its volume snapshot.
	VolumeSnapshotContentRetain DeletionPolicy = "Retain"
)

// VolumeSnapshotError describes an error encountered during snapshot creation.
type VolumeSnapshotError struct {
	// time is the timestamp when the error was encountered.
	// +optional
	Time *metav1.Time `json:"time,omitempty" protobuf:"bytes,1,opt,name=time"`

	// message is a string detailing the encountered error during snapshot
	// creation if specified.
	// NOTE: message may be logged, and it should not contain sensitive
	// information.
	// +optional
	Message *string `json:"message,omitempty" protobuf:"bytes,2,opt,name=message"`
}
