package v7

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"io/ioutil"
)

type Elastic struct {
	client *elasticsearch.Client
	index  string
}

func New(address string, index string) Elastic {

	client, _ := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{address},
	})

	return Elastic{client: client, index: index}
}

func (e Elastic) Search(body []byte) ([]byte, error) {

	response, err := e.client.Search(
		e.client.Search.WithContext(context.Background()),
		e.client.Search.WithIndex(fmt.Sprintf("%s*", e.index)),
		e.client.Search.WithTrackTotalHits(true),
		e.client.Search.WithBody(bytes.NewBuffer(body)),
	)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(response.Body).Decode(&e); err != nil {
			return nil, err
		} else {
			// Print the response status and error information.
			e, _ := e["error"].(map[string]interface{})
			return nil, fmt.Errorf("[%s] %s: %s", response.Status(), e["type"], e["reason"])
		}
	}

	return ioutil.ReadAll(response.Body)
}

func (e Elastic) GetTotalHitCount(v interface{}) int64 {
	m, _ := v.(map[string]interface{})
	f, _ := m["value"].(float64)
	return int64(f)
}
