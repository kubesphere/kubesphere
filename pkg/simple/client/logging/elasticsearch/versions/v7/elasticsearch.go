package v7

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"io/ioutil"
	"k8s.io/klog"
	"time"
)

type Elastic struct {
	client *elasticsearch.Client
	index  string
}

func New(address string, index string) *Elastic {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{address},
	})
	if err != nil {
		klog.Error(err)
		return nil
	}

	return &Elastic{client: client, index: index}
}

func (e *Elastic) Search(body []byte, scrollTimeout time.Duration) ([]byte, error) {
	response, err := e.client.Search(
		e.client.Search.WithContext(context.Background()),
		e.client.Search.WithIndex(fmt.Sprintf("%s*", e.index)),
		e.client.Search.WithTrackTotalHits(true),
		e.client.Search.WithBody(bytes.NewBuffer(body)),
		e.client.Search.WithScroll(scrollTimeout))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.IsError() {
		return nil, parseError(response)
	}

	return ioutil.ReadAll(response.Body)
}

func (e *Elastic) Scroll(scrollId string, scrollTimeout time.Duration) ([]byte, error) {
	response, err := e.client.Scroll(
		e.client.Scroll.WithContext(context.Background()),
		e.client.Scroll.WithScrollID(scrollId),
		e.client.Scroll.WithScroll(scrollTimeout))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.IsError() {
		return nil, parseError(response)
	}

	b, err := ioutil.ReadAll(response.Body)
	return b, err
}

func (e *Elastic) ClearScroll(scrollId string) {
	response, _ := e.client.ClearScroll(
		e.client.ClearScroll.WithContext(context.Background()),
		e.client.ClearScroll.WithScrollID(scrollId))
	defer response.Body.Close()
}

func (e *Elastic) GetTotalHitCount(v interface{}) int64 {
	m, _ := v.(map[string]interface{})
	f, _ := m["value"].(float64)
	return int64(f)
}

func parseError(response *esapi.Response) error {
	var e map[string]interface{}
	if err := json.NewDecoder(response.Body).Decode(&e); err != nil {
		return err
	} else {
		// Print the response status and error information.
		e, _ := e["error"].(map[string]interface{})
		return fmt.Errorf("type: %v, reason: %v", e["type"], e["reason"])
	}
}
