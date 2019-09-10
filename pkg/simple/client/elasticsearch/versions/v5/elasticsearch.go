package v5

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v5"
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
			return nil, fmt.Errorf("[%s] %s: %s",
				response.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}

	return ioutil.ReadAll(response.Body)
}

func (e Elastic) GetTotalHitCount(v interface{}) int64 {
	return int64(v.(float64))
}
