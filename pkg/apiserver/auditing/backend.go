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
	"net/http"
	"time"
)

const (
	WaitTimeout = time.Second
	WebhookURL  = "https://kube-auditing-webhook-svc.kubesphere-logging-system.svc:443/audit/webhook/event"
)

type Backend struct {
	url             string
	channelCapacity int
	semCh           chan interface{}
	cache           chan *v1alpha1.EventList
	client          http.Client
	sendTimeout     time.Duration
	waitTimeout     time.Duration
	stopCh          <-chan struct{}
}

func NewBackend(url string, channelCapacity int, cache chan *v1alpha1.EventList, sendTimeout time.Duration, stopCh <-chan struct{}) *Backend {

	b := Backend{
		url:             url,
		semCh:           make(chan interface{}, channelCapacity),
		channelCapacity: channelCapacity,
		waitTimeout:     WaitTimeout,
		cache:           cache,
		sendTimeout:     sendTimeout,
		stopCh:          stopCh,
	}

	if len(b.url) == 0 {
		b.url = WebhookURL
	}

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

		var event *v1alpha1.EventList
		select {
		case event = <-b.cache:
			if event == nil {
				break
			}
		case <-b.stopCh:
			break
		}

		send := func(event *v1alpha1.EventList) {
			ctx, cancel := context.WithTimeout(context.Background(), b.waitTimeout)
			defer cancel()

			select {
			case <-ctx.Done():
				klog.Errorf("get goroutine for audit(%s) timeout", event.Items[0].AuditID)
				return
			case b.semCh <- struct{}{}:
			}

			defer func() {
				<-b.semCh
			}()

			bs, err := b.eventToBytes(event)
			if err != nil {
				klog.V(6).Infof("json marshal error, %s", err)
				return
			}

			klog.V(8).Infof("%s", string(bs))

			response, err := b.client.Post(b.url, "application/json", bytes.NewBuffer(bs))
			if err != nil {
				klog.Errorf("send audit event[%s] error, %s", event.Items[0].AuditID, err)
				return
			}

			if response.StatusCode != http.StatusOK {
				klog.Errorf("send audit event[%s] error[%d]", event.Items[0].AuditID, response.StatusCode)
				return
			}
		}

		go send(event)
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
