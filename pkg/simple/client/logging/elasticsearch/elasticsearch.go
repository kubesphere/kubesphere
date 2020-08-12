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

package elasticsearch

import (
	"bytes"
	"context"
	"fmt"
	"github.com/json-iterator/go"
	"io"
	"kubesphere.io/kubesphere/pkg/simple/client/logging"
	"kubesphere.io/kubesphere/pkg/simple/client/logging/elasticsearch/versions/v5"
	"kubesphere.io/kubesphere/pkg/simple/client/logging/elasticsearch/versions/v6"
	"kubesphere.io/kubesphere/pkg/simple/client/logging/elasticsearch/versions/v7"
	"kubesphere.io/kubesphere/pkg/utils/esutil"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
	"strings"
	"sync"
)

const (
	ElasticV5 = "5"
	ElasticV6 = "6"
	ElasticV7 = "7"
)

// Elasticsearch implement logging interface
type Elasticsearch struct {
	host    string
	version string
	index   string

	c   client
	mux sync.Mutex
}

// versioned es client interface
type client interface {
	Search(indices string, body []byte, scroll bool) ([]byte, error)
	Scroll(id string) ([]byte, error)
	ClearScroll(id string)
	GetTotalHitCount(v interface{}) int64
}

func NewElasticsearch(options *Options) (*Elasticsearch, error) {
	var err error
	es := &Elasticsearch{
		host:    options.Host,
		version: options.Version,
		index:   options.IndexPrefix,
	}

	switch es.version {
	case ElasticV5:
		es.c, err = v5.New(es.host, es.index)
	case ElasticV6:
		es.c, err = v6.New(es.host, es.index)
	case ElasticV7:
		es.c, err = v7.New(es.host, es.index)
	case "":
		es.c = nil
	default:
		return nil, fmt.Errorf("unsupported elasticsearch version %s", es.version)
	}

	return es, err
}

func (es *Elasticsearch) loadClient() error {
	// Check if Elasticsearch client has been initialized.
	if es.c != nil {
		return nil
	}

	// Create Elasticsearch client.
	es.mux.Lock()
	defer es.mux.Unlock()

	if es.c != nil {
		return nil
	}

	// Detect Elasticsearch server version using Info API.
	// Info API is backward compatible across v5, v6 and v7.
	esv6, err := v6.New(es.host, "")
	if err != nil {
		return err
	}

	res, err := esv6.Client.Info(
		esv6.Client.Info.WithContext(context.Background()),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var b map[string]interface{}
	if err = jsoniter.NewDecoder(res.Body).Decode(&b); err != nil {
		return err
	}
	if res.IsError() {
		// Print the response status and error information.
		e, _ := b["error"].(map[string]interface{})
		return fmt.Errorf("[%s] type: %v, reason: %v", res.Status(), e["type"], e["reason"])
	}

	// get the major version
	version, _ := b["version"].(map[string]interface{})
	number, _ := version["number"].(string)
	if number == "" {
		return fmt.Errorf("failed to detect elastic version number")
	}

	var c client
	v := strings.Split(number, ".")[0]
	switch v {
	case ElasticV5:
		c, err = v5.New(es.host, es.index)
	case ElasticV6:
		c, err = v6.New(es.host, es.index)
	case ElasticV7:
		c, err = v7.New(es.host, es.index)
	default:
		err = fmt.Errorf("unsupported elasticsearch version %s", version)
	}

	if err != nil {
		return err
	}

	es.c = c
	es.version = v
	return nil
}

func (es *Elasticsearch) GetCurrentStats(sf logging.SearchFilter) (logging.Statistics, error) {
	var err error

	err = es.loadClient()
	if err != nil {
		return logging.Statistics{}, err
	}

	body, err := newBodyBuilder().
		mainBool(sf).
		cardinalityAggregation().
		bytes()
	if err != nil {
		return logging.Statistics{}, err
	}

	b, err := es.c.Search(esutil.ResolveIndexNames(es.index, sf.Starttime, sf.Endtime), body, true)
	if err != nil {
		return logging.Statistics{}, err
	}

	res, err := parseResponse(b)
	if err != nil {
		return logging.Statistics{}, err
	}

	return logging.Statistics{
			Containers: res.Value,
			Logs:       es.c.GetTotalHitCount(res.Total),
		},
		nil
}

func (es *Elasticsearch) CountLogsByInterval(sf logging.SearchFilter, interval string) (logging.Histogram, error) {
	var err error

	err = es.loadClient()
	if err != nil {
		return logging.Histogram{}, err
	}

	body, err := newBodyBuilder().
		mainBool(sf).
		dateHistogramAggregation(interval).
		bytes()
	if err != nil {
		return logging.Histogram{}, err
	}

	b, err := es.c.Search(esutil.ResolveIndexNames(es.index, sf.Starttime, sf.Endtime), body, false)
	if err != nil {
		return logging.Histogram{}, err
	}

	res, err := parseResponse(b)
	if err != nil {
		return logging.Histogram{}, err
	}

	var h logging.Histogram
	h.Total = es.c.GetTotalHitCount(res.Total)
	for _, b := range res.Buckets {
		h.Buckets = append(h.Buckets, logging.Bucket{
			Time:  b.Time,
			Count: b.Count,
		})
	}
	return h, nil
}

func (es *Elasticsearch) SearchLogs(sf logging.SearchFilter, f, s int64, o string) (logging.Logs, error) {
	var err error

	err = es.loadClient()
	if err != nil {
		return logging.Logs{}, err
	}

	body, err := newBodyBuilder().
		mainBool(sf).
		from(f).
		size(s).
		sort(o).
		bytes()
	if err != nil {
		return logging.Logs{}, err
	}

	b, err := es.c.Search(esutil.ResolveIndexNames(es.index, sf.Starttime, sf.Endtime), body, false)
	if err != nil {
		return logging.Logs{}, err
	}

	res, err := parseResponse(b)
	if err != nil {
		return logging.Logs{}, err
	}

	var l logging.Logs
	l.Total = es.c.GetTotalHitCount(res.Total)
	for _, hit := range res.AllHits {
		l.Records = append(l.Records, logging.Record{
			Log:       hit.Log,
			Time:      hit.Time,
			Namespace: hit.Namespace,
			Pod:       hit.Pod,
			Container: hit.Container,
		})
	}
	return l, nil
}

func (es *Elasticsearch) ExportLogs(sf logging.SearchFilter, w io.Writer) error {
	var err error
	var id string
	var data []string

	err = es.loadClient()
	if err != nil {
		return err
	}

	// Initial Search
	body, err := newBodyBuilder().
		mainBool(sf).
		from(0).
		size(1000).
		sort("desc").
		bytes()
	if err != nil {
		return err
	}

	b, err := es.c.Search(esutil.ResolveIndexNames(es.index, sf.Starttime, sf.Endtime), body, true)
	defer es.ClearScroll(id)
	if err != nil {
		return err
	}
	res, err := parseResponse(b)
	if err != nil {
		return err
	}

	id = res.ScrollId
	for _, hit := range res.AllHits {
		data = append(data, hit.Log)
	}

	// limit to retrieve max 100k records
	for i := 0; i < 100; i++ {
		if i != 0 {
			data, id, err = es.scroll(id)
			if err != nil {
				return err
			}
		}
		if len(data) == 0 {
			return nil
		}

		output := new(bytes.Buffer)
		for _, l := range data {
			output.WriteString(fmt.Sprintf(`%s`, stringutils.StripAnsi(l)))
		}
		_, err = io.Copy(w, output)
		if err != nil {
			return err
		}
	}
	return nil
}

func (es *Elasticsearch) scroll(id string) ([]string, string, error) {
	b, err := es.c.Scroll(id)
	if err != nil {
		return nil, id, err
	}

	res, err := parseResponse(b)
	if err != nil {
		return nil, id, err
	}

	var data []string
	for _, hit := range res.AllHits {
		data = append(data, hit.Log)
	}
	return data, res.ScrollId, nil
}

func (es *Elasticsearch) ClearScroll(id string) {
	if id != "" {
		es.c.ClearScroll(id)
	}
}
