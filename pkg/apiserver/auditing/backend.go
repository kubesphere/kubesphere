/*
Copyright 2020 KubeSphere Authors

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

package auditing

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apiserver/auditing/v1alpha1"
	options "kubesphere.io/kubesphere/pkg/simple/client/auditing/elasticsearch"
	"net/http"
	"time"
)

const (
	WaitTimeout          = time.Second
	SendTimeout          = time.Second * 3
	DefaultGoroutinesNum = 100
	DefaultBatchSize     = 100
	DefaultBatchWait     = time.Second * 3
	WebhookURL           = "https://kube-auditing-webhook-svc.kubesphere-logging-system.svc:443/audit/webhook/event"
)

type Backend struct {
	url          string
	semCh        chan interface{}
	cache        chan *v1alpha1.Event
	client       http.Client
	sendTimeout  time.Duration
	waitTimeout  time.Duration
	maxBatchSize int
	maxBatchWait time.Duration
	stopCh       <-chan struct{}
}

func NewBackend(opts *options.Options, cache chan *v1alpha1.Event, stopCh <-chan struct{}) *Backend {

	b := Backend{
		url:          opts.WebhookUrl,
		waitTimeout:  WaitTimeout,
		cache:        cache,
		sendTimeout:  SendTimeout,
		maxBatchSize: opts.MaxBatchSize,
		maxBatchWait: opts.MaxBatchWait,
		stopCh:       stopCh,
	}

	if len(b.url) == 0 {
		b.url = WebhookURL
	}

	if b.maxBatchWait == 0 {
		b.maxBatchWait = DefaultBatchWait
	}

	if b.maxBatchSize == 0 {
		b.maxBatchSize = DefaultBatchSize
	}

	goroutinesNum := opts.GoroutinesNum
	if goroutinesNum == 0 {
		goroutinesNum = DefaultGoroutinesNum
	}
	b.semCh = make(chan interface{}, goroutinesNum)

	b.client = http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: b.sendTimeout,
	}

	go b.worker()

	return &b
}

func (b *Backend) worker() {

	for {
		events := b.getEvents()
		if events == nil {
			break
		}

		if len(events.Items) == 0 {
			continue
		}

		go b.sendEvents(events)
	}
}

func (b *Backend) getEvents() *v1alpha1.EventList {

	ctx, cancel := context.WithTimeout(context.Background(), b.maxBatchWait)
	defer cancel()

	events := &v1alpha1.EventList{}
	for {
		select {
		case event := <-b.cache:
			if event == nil {
				break
			}
			events.Items = append(events.Items, *event)
			if len(events.Items) >= b.maxBatchSize {
				return events
			}
		case <-ctx.Done():
			return events
		case <-b.stopCh:
			return nil
		}
	}
}

func (b *Backend) sendEvents(events *v1alpha1.EventList) {

	ctx, cancel := context.WithTimeout(context.Background(), b.sendTimeout)
	defer cancel()

	stopCh := make(chan struct{})

	send := func() {
		ctx, cancel := context.WithTimeout(context.Background(), b.waitTimeout)
		defer cancel()

		select {
		case <-ctx.Done():
			klog.Error("get goroutine timeout")
			return
		case b.semCh <- struct{}{}:
		}

		start := time.Now()
		defer func() {
			stopCh <- struct{}{}
			klog.V(8).Infof("send %d auditing logs used %d", len(events.Items), time.Now().Sub(start).Milliseconds())
		}()

		bs, err := b.eventToBytes(events)
		if err != nil {
			klog.V(6).Infof("json marshal error, %s", err)
			return
		}

		klog.V(8).Infof("%s", string(bs))

		response, err := b.client.Post(b.url, "application/json", bytes.NewBuffer(bs))
		if err != nil {
			klog.Errorf("send audit events error, %s", err)
			return
		}

		if response.StatusCode != http.StatusOK {
			klog.Errorf("send audit events error[%d]", response.StatusCode)
			return
		}
	}

	go send()

	defer func() {
		<-b.semCh
	}()

	select {
	case <-ctx.Done():
		klog.Error("send audit events timeout")
	case <-stopCh:
	}
}

func (b *Backend) eventToBytes(event *v1alpha1.EventList) ([]byte, error) {

	bs, err := json.Marshal(event)
	if err != nil {
		// Normally, the serialization failure is caused by the failure of ResponseObject serialization.
		// To ensure the integrity of the auditing event to the greatest extent,
		// it is necessary to delete ResponseObject and and then try to serialize again.
		if event.Items[0].ResponseObject != nil {
			event.Items[0].ResponseObject = nil
			return json.Marshal(event)
		}

		return nil, err
	}

	return bs, err
}
