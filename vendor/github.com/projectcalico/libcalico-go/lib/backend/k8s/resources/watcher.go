// Copyright (c) 2016-2018 Tigera, Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/sirupsen/logrus"
	kwatch "k8s.io/apimachinery/pkg/watch"

	"github.com/projectcalico/libcalico-go/lib/backend/api"
	"github.com/projectcalico/libcalico-go/lib/backend/model"
	cerrors "github.com/projectcalico/libcalico-go/lib/errors"
)

const (
	resultsBufSize = 100
)

func newK8sWatcherConverter(
	ctx context.Context,
	name string,
	converter ConvertK8sResourceToKVPair,
	k8sWatch kwatch.Interface,
) api.WatchInterface {
	ctx, cancel := context.WithCancel(ctx)
	wc := &k8sWatcherConverter{
		logCxt:     logrus.WithField("resource", name),
		converter:  converter,
		k8sWatch:   k8sWatch,
		context:    ctx,
		cancel:     cancel,
		resultChan: make(chan api.WatchEvent, resultsBufSize),
	}
	go wc.processK8sEvents()
	return wc
}

type k8sWatcherConverter struct {
	logCxt     *logrus.Entry
	converter  ConvertK8sResourceToKVPair
	k8sWatch   kwatch.Interface
	context    context.Context
	cancel     context.CancelFunc
	resultChan chan api.WatchEvent
	terminated uint32
}

// Stop stops the watcher and releases associated resources.
// This calls through to the context cancel function.
func (crw *k8sWatcherConverter) Stop() {
	crw.cancel()
	crw.k8sWatch.Stop()
}

// ResultChan returns a channel used to receive WatchEvents.
func (crw *k8sWatcherConverter) ResultChan() <-chan api.WatchEvent {
	return crw.resultChan
}

// HasTerminated returns true when the watcher has completed termination processing.
func (crw *k8sWatcherConverter) HasTerminated() bool {
	return atomic.LoadUint32(&crw.terminated) != 0
}

// Loop to process the events stream from the underlying k8s Watcher and convert them to
// backend KVPs.
func (crw *k8sWatcherConverter) processK8sEvents() {
	crw.logCxt.Debug("Kubernetes watcher/converter started")
	defer func() {
		crw.logCxt.Debug("Kubernetes watcher/converter stopped, closing result channel")
		crw.Stop()
		close(crw.resultChan)
		atomic.AddUint32(&crw.terminated, 1)
	}()

	for {
		select {
		case event, ok := <-crw.k8sWatch.ResultChan():
			var e *api.WatchEvent
			if !ok {
				// The channel is closed so send a terminating watcher event indicating the watch was
				// closed by the remote.
				crw.logCxt.Debug("Watcher terminated by remote")
				e = &api.WatchEvent{
					Type: api.WatchError,
					Error: cerrors.ErrorWatchTerminated{
						Err:            fmt.Errorf("terminating error event from Kubernetes watcher: closed by remote"),
						ClosedByRemote: true,
					},
				}
			} else {
				// We have a valid event, so convert it.
				e = crw.convertEvent(event)
				if e == nil {
					crw.logCxt.Debug("Event converted to a no-op")
					continue
				}
			}

			select {
			case crw.resultChan <- *e:
				crw.logCxt.Debug("Kubernetes event converted and sent to backend watcher")

				// If this is an error event, check to see if it's a terminating one (the
				// convertEvent method will decide that).  If so, terminate this watcher.
				if e.Type == api.WatchError {
					crw.logCxt.WithError(e.Error).Debug("Watch event was an error event type")
					if _, ok := e.Error.(cerrors.ErrorWatchTerminated); ok {
						crw.logCxt.Debug("Watch event indicates a terminated watcher")
						return
					}
				}
			case <-crw.context.Done():
				crw.logCxt.Debug("Process watcher done event during watch event in kdd client")
				return
			}
		case <-crw.context.Done(): // user cancel
			crw.logCxt.Debug("Process watcher done event in kdd client")
			return
		}
	}
}

// convertEvent converts a Kubernetes Watch event into the equivalent Calico backend
// client watch event.
func (crw *k8sWatcherConverter) convertEvent(kevent kwatch.Event) *api.WatchEvent {
	var kvp *model.KVPair
	var err error
	if kevent.Type != kwatch.Error {
		k8sRes := kevent.Object.(Resource)
		kvp, err = crw.converter(k8sRes)
		if err != nil {
			crw.logCxt.WithError(err).Warning("Error converting Kubernetes resource to Calico resource")
			return &api.WatchEvent{
				Type:  api.WatchError,
				Error: err,
			}
		}
		if kvp == nil {
			return nil
		}
	}

	switch kevent.Type {
	case kwatch.Error:
		// An error directly from the k8s watcher is a terminating event.
		return &api.WatchEvent{
			Type: api.WatchError,
			Error: cerrors.ErrorWatchTerminated{
				Err: fmt.Errorf("terminating error event from Kubernetes watcher: %v", kevent.Object),
			},
		}
	case kwatch.Deleted:
		return &api.WatchEvent{
			Type: api.WatchDeleted,
			Old:  kvp,
		}
	case kwatch.Added:
		return &api.WatchEvent{
			Type: api.WatchAdded,
			New:  kvp,
		}
	case kwatch.Modified:
		// In KDD we don't have access to the previous settings, so just set the current settings.
		return &api.WatchEvent{
			Type: api.WatchModified,
			New:  kvp,
		}
	default:
		return &api.WatchEvent{
			Type:  api.WatchError,
			Error: fmt.Errorf("unhandled Kubernetes watcher event type: %v", kevent.Type),
		}
	}
}
