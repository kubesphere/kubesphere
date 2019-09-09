package v6

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v6"
	"io/ioutil"
)

type Elastic struct {
	Client *elasticsearch.Client
	index  string
}

func New(address string, index string) Elastic {

	client, _ := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{address},
	})

	return Elastic{Client: client, index: index}
}

func (e Elastic) Search(body []byte) ([]byte, error) {

	response, err := e.Client.Search(
		e.Client.Search.WithContext(context.Background()),
		e.Client.Search.WithIndex(fmt.Sprintf("%s*", e.index)),
		e.Client.Search.WithBody(bytes.NewBuffer(body)),
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
	f, _ := v.(float64)
	return int64(f)
}
