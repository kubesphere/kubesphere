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

package v1alpha1

import (
	"k8s.io/api/auditregistration/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Receiver config which received the audit alert
type Receiver struct {
	// Receiver name
	// +optional
	ReceicerName string `json:"name,omitempty" protobuf:"bytes,8,opt,name=name"`
	// Receiver type, alertmanager or webhook
	// +optional
	ReceiverType string `json:"type,omitempty" protobuf:"bytes,8,opt,name=type"`
	// ClientConfig holds the connection parameters for the webhook
	// +optional
	ReceiverConfig v1alpha1.WebhookClientConfig `json:"config,omitempty" protobuf:"bytes,8,opt,name=config"`
}

type AuditSinkPolicy struct {
	ArchivingRuleSelector *metav1.LabelSelector `json:"archivingRuleSelector,omitempty" protobuf:"bytes,8,opt,name=archivingRuleSelector"`
	AlertingRuleSelector  *metav1.LabelSelector `json:"alertingRuleSelector,omitempty" protobuf:"bytes,8,opt,name=alertingRuleSelector"`
}

type DynamicAuditConfig struct {
	// Throttle holds the options for throttling the webhook
	// +optional
	Throttle *v1alpha1.WebhookThrottleConfig `json:"throttle,omitempty" protobuf:"bytes,18,opt,name=throttle"`
	// Policy defines the policy for selecting which events should be sent to the webhook
	// +optional
	Policy *v1alpha1.Policy `json:"policy,omitempty" protobuf:"bytes,18,opt,name=policy"`
}

// WebhookSpec defines the desired state of Webhook
type WebhookSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Number of desired pods. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1.
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`
	// The webhook docker image name.
	// +optional
	Image string `json:"image,omitempty" protobuf:"bytes,2,opt,name=image"`
	// Image pull policy.
	// One of Always, Never, IfNotPresent.
	// Defaults to Always if :latest tag is specified, or IfNotPresent otherwise.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/containers/images#updating-images
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty" protobuf:"bytes,14,opt,name=imagePullPolicy,casttype=PullPolicy"`
	// ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
	// If specified, these secrets will be passed to individual puller implementations for them to use. For example,
	// in the case of docker, only DockerConfig type secrets are honored.
	// More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,15,rep,name=imagePullSecrets"`
	// Arguments to the entrypoint..
	// It will be appended to the args and replace the default value.
	// +optional
	Args []string `json:"args,omitempty" protobuf:"bytes,3,rep,name=args"`
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty" protobuf:"bytes,7,rep,name=nodeSelector"`
	// If specified, the pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty" protobuf:"bytes,18,opt,name=affinity"`
	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`
	// Compute Resources required by this container.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
	//  Receiver contains the information to make a connection with the alertmanager
	// +optional
	Receivers []Receiver `json:"receivers,omitempty" protobuf:"bytes,8,opt,name=receivers"`

	// AuditSinkPolicy is a rule selector, only the rule matched this selector will be taked effect.
	// +optional
	*AuditSinkPolicy `json:"auditSinkPolicy,omitempty" protobuf:"bytes,8,opt,name=auditSinkPolicy"`
	// Rule priority, DEBUG < INFO < WARNING
	//Audit events will be stored only when the priority of the audit rule
	// matching the audit event is greater than this.
	Priority string `json:"priority,omitempty" protobuf:"bytes,8,opt,name=priority"`
	// Audit type, static or dynamic.
	AuditType string `json:"auditType,omitempty" protobuf:"bytes,8,opt,name=auditType"`
	// The Level that all requests are recorded at.
	// available options: None, Metadata, Request, RequestResponse
	// default: Metadata
	// +optional
	AuditLevel v1alpha1.Level `json:"auditLevel" protobuf:"bytes,1,opt,name=auditLevel"`
	// K8s auditing is enabled or not.
	K8sAuditingEnabled bool `json:"k8sAuditingEnabled,omitempty" protobuf:"bytes,8,opt,name=priority"`
}

// WebhookStatus defines the observed state of Webhook
type WebhookStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +kubebuilder:object:root=true

// Webhook is the Schema for the webhooks API
type Webhook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebhookSpec   `json:"spec,omitempty"`
	Status WebhookStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WebhookList contains a list of Webhook
type WebhookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Webhook `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Webhook{}, &WebhookList{})
}
