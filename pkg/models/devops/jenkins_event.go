package devops

import (
	"fmt"
	"k8s.io/klog"
	v1alpha32 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha3"
)

type EventHandler interface {
	HandleJenkinsEvent(eventType string, event *v1alpha32.JenkinsEvent) error
}

type eventHandler struct {
	notifier *v1alpha32.EventNotifier
}

func NewEventHandler(notifier *v1alpha32.EventNotifier) EventHandler {
	return &eventHandler{notifier: notifier}
}

func (e *eventHandler) HandleJenkinsEvent(eventType string, event *v1alpha32.JenkinsEvent) error {
	switch eventType {
	case v1alpha32.PipelineStarted:
		e.notifier.OnPipelineStarted(event.ToEvent())
	case v1alpha32.PipelineCompleted:
		e.notifier.OnPipelineCompleted(event.ToEvent())
	case v1alpha32.PipelineFinalized:
		e.notifier.OnPipelineFinalized(event.ToEvent())
	case v1alpha32.PipelineReviewAborted:
		e.notifier.OnPipelineReviewAborted(event.ToEvent())
	case v1alpha32.PipelineReviewProceeded:
		e.notifier.OnPipelineReviewProceeded(event.ToEvent())
	case v1alpha32.PipelinePendingReview:
		e.notifier.OnPipelinePendingReview(event.ToEvent())
	default:
		err := fmt.Errorf("unknow event type %s, value %v", eventType, event)
		klog.Warning(err)
		return err
	}
	return nil
}
