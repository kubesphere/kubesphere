package elastic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// Elasticsearch client
type Elasticsearch struct {
	Addr string
}

// request build the url of an API request call
func (client *Elasticsearch) request(index, class string, id int64, request string) string {
	var url string
	if index == "" {
		url = fmt.Sprintf("http://%s/_%s", client.Addr, request)
	} else if class == "" {
		url = fmt.Sprintf("http://%s/%s/_%s", client.Addr, index, request)
	} else if id < 0 {
		url = fmt.Sprintf("http://%s/%s/%s/_%s", client.Addr, index, class, request)
	} else {
		url = fmt.Sprintf("http://%s/%s/%s/%d/_%s", client.Addr, index, class, id, request)
	}
	return url
}

// Execute an HTTP request and parse the response
func (client *Elasticsearch) Execute(method, url, query string, parser Parser) (interface{}, error) {
	var body io.Reader
	if query != "" {
		body = bytes.NewReader([]byte(query))
	}
	// submit the request
	log.Println(method, url, query)
	reader, err := exec(method, url, body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if data, err := ioutil.ReadAll(reader); err == nil {
		// marshal response
		result, err := parser.Parse(data)
		return result, err
	}
	return nil, err
}

// String returns a string representation of the dictionary
func String(obj interface{}) string {
	marshaled, err := json.Marshal(obj)
	if err != nil {
		log.Println(err)
	}
	return string(marshaled)
}

// urlString Construct a url
func urlString(prefix string, params map[string]string) string {
	url := prefix
	if len(params) > 0 {
		if strings.Contains(url, "?") {
			// if there is already a key/value pair in url
			if url[len(url)-1] != byte('?') && len(params) > 0 {
				url += "&"
			}
		} else {
			url += "?"
		}
		for name, value := range params {
			url += name
			if value != "" {
				url += "=" + value
			}
			url += "&"
		}
		url = url[:len(url)-1]
	}
	return url
}

// Execute a REST request
func exec(method, url string, body io.Reader) (io.Reader, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
