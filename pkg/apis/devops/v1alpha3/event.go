package v1alpha3

import (
	"sync"
)

// EventNotifier is a special EventHandler that is responsible for sending events to registered Handlers
// when you need to add a specific Event handler,
// you need to implement the EventHandler interface and call RegisterEventHandler to register with EventNotifier.
type EventNotifier struct {
	lock      sync.RWMutex
	listeners []EventHandler
}

func NewEventNotifier() *EventNotifier {
	return &EventNotifier{
		lock: sync.RWMutex{},
	}
}

func (e *EventNotifier) OnPipelineStarted(event Event) {
	e.lock.RLock()
	defer e.lock.RUnlock()
	for _, handler := range e.listeners {
		handler.OnPipelineStarted(event)
	}
}

func (e *EventNotifier) OnPipelineCompleted(event Event) {
	e.lock.RLock()
	defer e.lock.RUnlock()
	for _, handler := range e.listeners {
		handler.OnPipelineCompleted(event)
	}
}

func (e *EventNotifier) OnPipelineFinalized(event Event) {
	e.lock.RLock()
	defer e.lock.RUnlock()
	for _, handler := range e.listeners {
		handler.OnPipelineFinalized(event)
	}
}

func (e *EventNotifier) OnPipelinePendingReview(event Event) {
	e.lock.RLock()
	defer e.lock.RUnlock()
	for _, handler := range e.listeners {
		handler.OnPipelinePendingReview(event)
	}
}

func (e *EventNotifier) OnPipelineReviewProceeded(event Event) {
	e.lock.RLock()
	defer e.lock.RUnlock()
	for _, handler := range e.listeners {
		handler.OnPipelineReviewProceeded(event)
	}
}

func (e *EventNotifier) OnPipelineReviewAborted(event Event) {
	e.lock.RLock()
	defer e.lock.RUnlock()
	for _, handler := range e.listeners {
		handler.OnPipelineReviewAborted(event)
	}
}

func (e *EventNotifier) RegisterEventHandler(handler EventHandler) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.listeners = append(e.listeners, handler)
}

type EventHandler interface {
	OnPipelineStarted(event Event)
	OnPipelineCompleted(event Event)
	OnPipelineFinalized(event Event)
	OnPipelinePendingReview(event Event)
	OnPipelineReviewProceeded(event Event)
	OnPipelineReviewAborted(event Event)
}

type ResourceEventHandlerFuncs struct {
	PipelineStartedFunc     func(event Event)
	PipelineCompletedFunc   func(event Event)
	PipelineFinalizedFunc   func(event Event)
	PipelinePendingReview   func(event Event)
	PipelineReviewProceeded func(event Event)
	PipelineReviewAborted   func(event Event)
}

// OnPipelineStarted calls PipelineStartedFunc if it's not nil.
func (r ResourceEventHandlerFuncs) OnPipelineStarted(event Event) {
	if r.PipelineStartedFunc != nil {
		r.PipelineStartedFunc(event)
	}
}

// OnPipelineCompleted calls PipelineCompletedFunc if it's not nil.
func (r ResourceEventHandlerFuncs) OnPipelineCompleted(event Event) {
	if r.PipelineCompletedFunc != nil {
		r.PipelineCompletedFunc(event)
	}
}

// OnPipelineFinalized calls PipelineFinalizedFunc if it's not nil.
func (r ResourceEventHandlerFuncs) OnPipelineFinalized(event Event) {
	if r.PipelineFinalizedFunc != nil {
		r.PipelineFinalizedFunc(event)
	}
}

// OnPipelinePendingReview calls PipelinePendingReview if it's not nil.
func (r ResourceEventHandlerFuncs) OnPipelinePendingReview(event Event) {
	if r.PipelinePendingReview != nil {
		r.PipelinePendingReview(event)
	}
}

// PipelinePendingReview calls PipelineReviewProceeded if it's not nil.
func (r ResourceEventHandlerFuncs) OnPipelineReviewProceeded(event Event) {
	if r.PipelineReviewProceeded != nil {
		r.PipelineReviewProceeded(event)
	}
}

// OnPipelineReviewAborted calls PipelineReviewAborted if it's not nil.
func (r ResourceEventHandlerFuncs) OnPipelineReviewAborted(event Event) {
	if r.PipelineReviewAborted != nil {
		r.PipelineReviewAborted(event)
	}
}
