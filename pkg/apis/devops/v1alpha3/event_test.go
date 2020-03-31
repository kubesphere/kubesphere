package v1alpha3

import (
	"reflect"
	"testing"
)

type recorder struct {
	store []Event
}

func (r *recorder) record(event Event) {
	r.store = append(r.store, event)
}

func TestRegisterEventHandler(t *testing.T) {
	r := &recorder{}
	notifier := NewEventNotifier()
	notifier.RegisterEventHandler(ResourceEventHandlerFuncs{
		PipelineStartedFunc: func(event Event) {
			r.record(event)
		},
	})
	expects := []Event{
		{Timestamp: 1},
		{Timestamp: 2},
	}
	for _, expect := range expects {
		notifier.OnPipelineStarted(expect)
	}
	if len(r.store) != 2 {
		t.Fatalf("expect store length %v, but got %v", len(expects), len(r.store))
	}
	for i, expect := range expects {
		if !reflect.DeepEqual(expect, r.store[i]) {
			t.Fatalf("expect \n %v \n, but got \n %v", expect, r.store[i])
		}
	}

}
