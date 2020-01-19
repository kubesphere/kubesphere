package elasticsearch

import (
	"bytes"
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"io"
	"kubesphere.io/kubesphere/pkg/simple/client/logging"
	v5 "kubesphere.io/kubesphere/pkg/simple/client/logging/elasticsearch/versions/v5"
	v6 "kubesphere.io/kubesphere/pkg/simple/client/logging/elasticsearch/versions/v6"
	v7 "kubesphere.io/kubesphere/pkg/simple/client/logging/elasticsearch/versions/v7"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
	"strings"
)

const (
	ElasticV5 = "5"
	ElasticV6 = "6"
	ElasticV7 = "7"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Elasticsearch implement logging interface
type Elasticsearch struct {
	c client
}

// versioned es client interface
type client interface {
	// Perform Search API
	Search(body []byte) ([]byte, error)
	Scroll(id string) ([]byte, error)
	ClearScroll(id string)
	GetTotalHitCount(v interface{}) int64
}

func NewElasticsearch(options *Options) (*Elasticsearch, error) {
	var version, index string
	es := &Elasticsearch{}

	if options.Version == "" {
		var err error
		version, err = detectVersionMajor(options.Host)
		if err != nil {
			return nil, err
		}
	} else {
		version = options.Version
	}

	if options.IndexPrefix != "" {
		index = options.IndexPrefix
	} else {
		index = "logstash"
	}

	switch version {
	case ElasticV5:
		es.c = v5.New(options.Host, index)
	case ElasticV6:
		es.c = v6.New(options.Host, index)
	case ElasticV7:
		es.c = v7.New(options.Host, index)
	default:
		return nil, fmt.Errorf("unsupported elasticsearch version %s", version)
	}

	return es, nil
}

func (es *Elasticsearch) ES() *client {
	return &es.c
}

func detectVersionMajor(host string) (string, error) {
	// Info APIs are backward compatible with versions of v5.x, v6.x and v7.x
	es := v6.New(host, "")
	res, err := es.Client.Info(
		es.Client.Info.WithContext(context.Background()),
	)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	var b map[string]interface{}
	if err = json.NewDecoder(res.Body).Decode(&b); err != nil {
		return "", err
	}
	if res.IsError() {
		// Print the response status and error information.
		e, _ := b["error"].(map[string]interface{})
		return "", fmt.Errorf("[%s] type: %v, reason: %v", res.Status(), e["type"], e["reason"])
	}

	// get the major version
	version, _ := b["version"].(map[string]interface{})
	number, _ := version["number"].(string)
	if number == "" {
		return "", fmt.Errorf("failed to detect elastic version number")
	}

	v := strings.Split(number, ".")[0]
	return v, nil
}

func (es Elasticsearch) GetCurrentStats(sf logging.SearchFilter) (logging.Statistics, error) {
	body, err := newBodyBuilder().
		mainBool(sf).
		cardinalityAggregation().
		bytes()
	if err != nil {
		return logging.Statistics{}, err
	}

	b, err := es.c.Search(body)
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

func (es Elasticsearch) CountLogsByInterval(sf logging.SearchFilter, interval string) (logging.Histogram, error) {
	body, err := newBodyBuilder().
		mainBool(sf).
		dateHistogramAggregation(interval).
		bytes()
	if err != nil {
		return logging.Histogram{}, err
	}

	b, err := es.c.Search(body)
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

func (es Elasticsearch) SearchLogs(sf logging.SearchFilter, f, s int64, o string) (logging.Logs, error) {
	body, err := newBodyBuilder().
		mainBool(sf).
		from(f).
		size(s).
		sort(o).
		bytes()
	if err != nil {
		return logging.Logs{}, err
	}

	b, err := es.c.Search(body)
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

func (es Elasticsearch) ExportLogs(sf logging.SearchFilter, w io.Writer) error {
	var id string
	var from int64 = 0
	var size int64 = 1000

	res, err := es.SearchLogs(sf, from, size, "desc")
	defer es.ClearScroll(id)
	if err != nil {
		return err
	}

	if res.Records == nil || len(res.Records) == 0 {
		return nil
	}

	// limit to retrieve max 100k records
	for i := 0; i < 100; i++ {
		res, id, err = es.scroll(id)
		if err != nil {
			return err
		}

		if res.Records == nil || len(res.Records) == 0 {
			return nil
		}

		output := new(bytes.Buffer)
		for _, r := range res.Records {
			output.WriteString(fmt.Sprintf(`%s`, stringutils.StripAnsi(r.Log)))
		}
		_, err = io.Copy(w, output)
		if err != nil {
			return err
		}
	}
	return nil
}

func (es *Elasticsearch) scroll(id string) (logging.Logs, string, error) {
	b, err := es.c.Scroll(id)
	if err != nil {
		return logging.Logs{}, id, err
	}

	res, err := parseResponse(b)
	if err != nil {
		return logging.Logs{}, id, err
	}

	var l logging.Logs
	for _, hit := range res.AllHits {
		l.Records = append(l.Records, logging.Record{
			Log: hit.Log,
		})
	}
	return l, res.ScrollId, nil
}

func (es *Elasticsearch) ClearScroll(id string) {
	if id != "" {
		es.c.ClearScroll(id)
	}
}
