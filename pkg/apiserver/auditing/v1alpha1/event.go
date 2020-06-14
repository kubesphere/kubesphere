package v1alpha1

import "k8s.io/apiserver/pkg/apis/audit"

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
