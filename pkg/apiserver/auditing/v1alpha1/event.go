package v1alpha1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/apis/audit"
)

type Event struct {
	// Devops project
	Devops string
	// The workspace which this audit event happened
	Workspace string
	// The cluster which this audit event happened
	Cluster string
	// Message send to user.
	Message string

	audit.Event
}

type EventList struct {
	Items []Event
}

type Object struct {
	v1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
}
