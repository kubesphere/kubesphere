package auditing

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"k8s.io/klog"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const (
	WaitTimeout = time.Second
	WebhookURL  = "https://kube-auditing-webhook-svc.kubesphere-logging-system.svc:443"
)

type Backend struct {
	channelCapacity int
	semCh           chan interface{}
	cache           chan *EventList
	client          http.Client
	sendTimeout     time.Duration
	waitTimeout     time.Duration
}

func NewBackend(channelCapacity int, cache chan *EventList, sendTimeout time.Duration) *Backend {

	b := Backend{
		semCh:           make(chan interface{}, channelCapacity),
		channelCapacity: channelCapacity,
		waitTimeout:     WaitTimeout,
		cache:           cache,
		sendTimeout:     sendTimeout,
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

	// Stop when receiver signal Interrupt.
	stopCh := b.SetupSignalHandler()

	for {

		var event *EventList
		select {
		case event = <-b.cache:
			if event == nil {
				break
			}
		case <-stopCh:
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

			response, err := b.client.Post(WebhookURL, "application/json", bytes.NewBuffer(bs))
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

func (b *Backend) SetupSignalHandler() (stopCh <-chan struct{}) {

	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, []os.Signal{os.Interrupt}...)
	go func() {
		<-c
		close(stop)
	}()

	return stop
}
