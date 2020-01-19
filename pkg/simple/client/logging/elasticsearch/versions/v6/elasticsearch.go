package v6

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v6"
	"github.com/elastic/go-elasticsearch/v6/esapi"
	"io/ioutil"
	"k8s.io/klog"
	"time"
)

type Elastic struct {
	Client *elasticsearch.Client
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

	return &Elastic{Client: client, index: index}
}

func (e *Elastic) Search(body []byte, scrollTimeout time.Duration) ([]byte, error) {
	response, err := e.Client.Search(
		e.Client.Search.WithContext(context.Background()),
		e.Client.Search.WithIndex(fmt.Sprintf("%s*", e.index)),
		e.Client.Search.WithBody(bytes.NewBuffer(body)),
		e.Client.Search.WithScroll(scrollTimeout))
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
	response, err := e.Client.Scroll(
		e.Client.Scroll.WithContext(context.Background()),
		e.Client.Scroll.WithScrollID(scrollId),
		e.Client.Scroll.WithScroll(scrollTimeout))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.IsError() {
		return nil, parseError(response)
	}

	return ioutil.ReadAll(response.Body)
}

func (e *Elastic) ClearScroll(scrollId string) {
	response, _ := e.Client.ClearScroll(
		e.Client.ClearScroll.WithContext(context.Background()),
		e.Client.ClearScroll.WithScrollID(scrollId))
	defer response.Body.Close()
}

func (e *Elastic) GetTotalHitCount(v interface{}) int64 {
	f, _ := v.(float64)
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
