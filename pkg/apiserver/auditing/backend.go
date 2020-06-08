package auditing

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"k8s.io/client-go/listers/auditregistration/v1alpha1"
	"k8s.io/klog"
	"net/http"
	"time"
)

const (
	WaitTimeout = time.Second
)

type Backend struct {
	lister          v1alpha1.AuditSinkLister
	channelCapacity int
	semCh           chan interface{}
	cache           chan *EventList
	client          http.Client
	sendTimeout     time.Duration
	waitTimeout     time.Duration
}

func NewBackend(channelCapacity int, cache chan *EventList, sendTimeout time.Duration, lister v1alpha1.AuditSinkLister) *Backend {

	b := Backend{
		semCh:           make(chan interface{}, channelCapacity),
		channelCapacity: channelCapacity,
		waitTimeout:     WaitTimeout,
		cache:           cache,
		sendTimeout:     sendTimeout,
		lister:          lister,
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
		event := <-b.cache
		if event == nil {
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

			response, err := b.client.Post(b.getWebhookUrl(), "application/json", bytes.NewBuffer(bs))
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

func (b *Backend) getWebhookUrl() string {
	as, err := b.lister.Get(DefaultAuditSink)
	if err != nil {
		return ""
	}

	if as.Spec.Webhook.ClientConfig.URL != nil {
		return *as.Spec.Webhook.ClientConfig.URL
	} else if as.Spec.Webhook.ClientConfig.Service != nil {
		s := as.Spec.Webhook.ClientConfig.Service
		url := fmt.Sprintf("https://%s.%s.svc:443", s.Name, s.Namespace)
		if s.Path != nil {
			url = fmt.Sprintf("%s%s", url, *s.Path)
		}
		return url
	} else {
		return ""
	}
}
