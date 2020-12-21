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
	options "kubesphere.io/kubesphere/pkg/simple/client/auditing"
	"net/http"
	"time"
)

const (
	GetSenderTimeout     = time.Second
	SendTimeout          = time.Second * 3
	DefaultSendersNum    = 100
	DefaultBatchSize     = 100
	DefaultBatchInterval = time.Second * 3
	WebhookURL           = "https://kube-auditing-webhook-svc.kubesphere-logging-system.svc:443/audit/webhook/event"
)

type Backend struct {
	url                string
	senderCh           chan interface{}
	cache              chan *v1alpha1.Event
	client             http.Client
	sendTimeout        time.Duration
	getSenderTimeout   time.Duration
	eventBatchSize     int
	eventBatchInterval time.Duration
	stopCh             <-chan struct{}
}

func NewBackend(opts *options.Options, cache chan *v1alpha1.Event, stopCh <-chan struct{}) *Backend {

	b := Backend{
		url:                opts.WebhookUrl,
		getSenderTimeout:   GetSenderTimeout,
		cache:              cache,
		sendTimeout:        SendTimeout,
		eventBatchSize:     opts.EventBatchSize,
		eventBatchInterval: opts.EventBatchInterval,
		stopCh:             stopCh,
	}

	if len(b.url) == 0 {
		b.url = WebhookURL
	}

	if b.eventBatchInterval == 0 {
		b.eventBatchInterval = DefaultBatchInterval
	}

	if b.eventBatchSize == 0 {
		b.eventBatchSize = DefaultBatchSize
	}

	sendersNum := opts.EventSendersNum
	if sendersNum == 0 {
		sendersNum = DefaultSendersNum
	}
	b.senderCh = make(chan interface{}, sendersNum)

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

	ctx, cancel := context.WithTimeout(context.Background(), b.eventBatchInterval)
	defer cancel()

	events := &v1alpha1.EventList{}
	for {
		select {
		case event := <-b.cache:
			if event == nil {
				break
			}
			events.Items = append(events.Items, *event)
			if len(events.Items) >= b.eventBatchSize {
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
		ctx, cancel := context.WithTimeout(context.Background(), b.getSenderTimeout)
		defer cancel()

		select {
		case <-ctx.Done():
			klog.Error("Get auditing event sender timeout")
			return
		case b.senderCh <- struct{}{}:
		}

		start := time.Now()
		defer func() {
			stopCh <- struct{}{}
			klog.V(8).Infof("send %d auditing logs used %d", len(events.Items), time.Now().Sub(start).Milliseconds())
		}()

		bs, err := b.eventToBytes(events)
		if err != nil {
			klog.Errorf("json marshal error, %s", err)
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
		<-b.senderCh
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
