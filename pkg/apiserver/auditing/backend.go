package auditing

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"k8s.io/klog"
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
	cache           chan *EventList
	client          http.Client
	sendTimeout     time.Duration
	waitTimeout     time.Duration
	stopCh          <-chan struct{}
}

func NewBackend(url string, channelCapacity int, cache chan *EventList, sendTimeout time.Duration, stopCh <-chan struct{}) *Backend {

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

		var event *EventList
		select {
		case event = <-b.cache:
			if event == nil {
				break
			}
		case <-b.stopCh:
			break
		}

		send := func(event *EventList) {
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

			bs, err := json.Marshal(event)
			if err != nil {
				klog.Errorf("json marshal error, %s", err)
				return
			}

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
