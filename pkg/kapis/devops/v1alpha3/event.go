package v1alpha3

import "sync"

var notifier eventNotifier
var once sync.Once

// eventNotifier is the bus for event distribution.
// when you need to add a specific Event handler,
// you need to implement the EventHandler interface and call RegisterEventHandler to register with eventNotifier.
// eventNotifier should be accessed in singleton mode, do not initialize an eventNotifier instance yourself.

type eventNotifier struct {
	lock      sync.RWMutex
	listeners []EventHandler
}

func (e *eventNotifier) onPipelineStarted(event Event) {
	e.lock.RLock()
	defer e.lock.RUnlock()
	for _, handler := range e.listeners {
		handler.OnPipelineStarted(event)
	}
}

func (e *eventNotifier) onPipelineCompleted(event Event) {
	e.lock.RLock()
	defer e.lock.RUnlock()
	for _, handler := range e.listeners {
		handler.OnPipelineCompleted(event)
	}
}

func (e *eventNotifier) onPipelineFinalized(event Event) {
	e.lock.RLock()
	defer e.lock.RUnlock()
	for _, handler := range e.listeners {
		handler.OnPipelineFinalized(event)
	}
}

func (e *eventNotifier) onPipelinePendingReview(event Event) {
	e.lock.RLock()
	defer e.lock.RUnlock()
	for _, handler := range e.listeners {
		handler.OnPipelinePendingReview(event)
	}
}

func (e *eventNotifier) onPipelineReviewProceeded(event Event) {
	e.lock.RLock()
	defer e.lock.RUnlock()
	for _, handler := range e.listeners {
		handler.OnPipelineReviewProceeded(event)
	}
}

func (e *eventNotifier) onPipelineReviewAborted(event Event) {
	e.lock.RLock()
	defer e.lock.RUnlock()
	for _, handler := range e.listeners {
		handler.OnPipelineReviewAborted(event)
	}
}

func GetEventNotifier() *eventNotifier {
	once.Do(func() {
		notifier = eventNotifier{
			lock: sync.RWMutex{},
		}
	})
	return &notifier
}

func RegisterEventHandler(handler EventHandler) {
	notifier := GetEventNotifier()
	notifier.lock.Lock()
	defer notifier.lock.Unlock()
	notifier.listeners = append(notifier.listeners, handler)
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

// onPipelineReviewAborted calls PipelineReviewAborted if it's not nil.
func (r ResourceEventHandlerFuncs) OnPipelineReviewAborted(event Event) {
	if r.PipelineReviewAborted != nil {
		r.PipelineReviewAborted(event)
	}
}
