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
	"encoding/json"
	"fmt"
	"io"
	"kubesphere.io/kubesphere/pkg/simple/client/es"
	"kubesphere.io/kubesphere/pkg/simple/client/es/query"
	"kubesphere.io/kubesphere/pkg/simple/client/logging"
	"time"

	"kubesphere.io/kubesphere/pkg/utils/stringutils"
)

const (
	podNameMaxLength          = 63
	podNameSuffixLength       = 6  // 5 characters + 1 hyphen
	replicaSetSuffixMaxLength = 11 // max 10 characters + 1 hyphen
)

type Source struct {
	Log        string `json:"log"`
	Time       string `json:"time"`
	Kubernetes `json:"kubernetes"`
}

type Kubernetes struct {
	Namespace string `json:"namespace_name"`
	Pod       string `json:"pod_name"`
	Container string `json:"container_name"`
	Host      string `json:"host"`
}

// Elasticsearch implement logging interface
type client struct {
	c *es.Client
}

func NewClient(options *logging.Options) (logging.Client, error) {

	c := &client{}

	var err error
	c.c, err = es.NewClient(options.Host, options.IndexPrefix, options.Version)
	return c, err
}

func (c *client) GetCurrentStats(sf logging.SearchFilter) (logging.Statistics, error) {
	var err error

	b := query.NewBuilder().
		WithQuery(parseToQueryPart(sf)).
		WithAggregations(query.NewAggregations().
			WithCardinalityAggregation("kubernetes.docker_id.keyword")).
		WithSize(0)

	resp, err := c.c.Search(b, sf.Starttime, sf.Endtime, false)
	if err != nil {
		return logging.Statistics{}, err
	}

	return logging.Statistics{
			Containers: resp.Value,
			Logs:       c.c.GetTotalHitCount(resp.Total),
		},
		nil
}

func (c *client) CountLogsByInterval(sf logging.SearchFilter, interval string) (logging.Histogram, error) {

	b := query.NewBuilder().
		WithQuery(parseToQueryPart(sf)).
		WithAggregations(query.NewAggregations().
			WithDateHistogramAggregation("time", interval)).
		WithSize(0)

	resp, err := c.c.Search(b, sf.Starttime, sf.Endtime, false)
	if err != nil {
		return logging.Histogram{}, err
	}

	h := logging.Histogram{
		Total: c.c.GetTotalHitCount(resp.Total),
	}
	for _, bucket := range resp.Buckets {
		h.Buckets = append(h.Buckets, logging.Bucket{
			Time:  bucket.Key,
			Count: bucket.Count,
		})
	}
	return h, nil
}

func (c *client) SearchLogs(sf logging.SearchFilter, f, s int64, o string) (logging.Logs, error) {

	b := query.NewBuilder().
		WithQuery(parseToQueryPart(sf)).
		WithSort("time", o).
		WithFrom(f).
		WithSize(s)

	resp, err := c.c.Search(b, sf.Starttime, sf.Endtime, false)
	if err != nil {
		return logging.Logs{}, err
	}

	l := logging.Logs{
		Total: c.c.GetTotalHitCount(resp.Total),
	}

	for _, hit := range resp.AllHits {
		s := c.getSource(hit.Source)
		l.Records = append(l.Records, logging.Record{
			Log:       s.Log,
			Time:      s.Time,
			Namespace: s.Namespace,
			Pod:       s.Pod,
			Container: s.Container,
		})
	}
	return l, nil
}

func (c *client) ExportLogs(sf logging.SearchFilter, w io.Writer) error {

	var id string
	var data []string

	b := query.NewBuilder().
		WithQuery(parseToQueryPart(sf)).
		WithSort("time", "desc").
		WithFrom(0).
		WithSize(1000)

	resp, err := c.c.Search(b, sf.Starttime, sf.Endtime, true)
	if err != nil {
		return err
	}

	defer c.c.ClearScroll(id)

	id = resp.ScrollId
	for _, hit := range resp.AllHits {
		data = append(data, c.getSource(hit.Source).Log)
	}

	// limit to retrieve max 100k records
	for i := 0; i < 100; i++ {
		if i != 0 {
			data, id, err = c.scroll(id)
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

func (c *client) scroll(id string) ([]string, string, error) {
	resp, err := c.c.Scroll(id)
	if err != nil {
		return nil, id, err
	}

	var data []string
	for _, hit := range resp.AllHits {
		data = append(data, c.getSource(hit.Source).Log)
	}
	return data, resp.ScrollId, nil
}

func (c *client) getSource(val interface{}) Source {

	s := Source{}

	bs, err := json.Marshal(val)
	if err != nil {
		return s
	}

	err = json.Unmarshal(bs, &s)
	if err != nil {
		return s
	}

	return s
}

func parseToQueryPart(sf logging.SearchFilter) *query.Query {

	var mini int32 = 1
	b := query.NewBool()

	bi := query.NewBool().WithMinimumShouldMatch(mini)
	for ns, t := range sf.NamespaceFilter {
		ct := time.Time{}
		if t != nil {
			ct = *t
		}

		bi.AppendShould(query.NewBool().
			AppendFilter(query.NewMatchPhrase("kubernetes.namespace_name.keyword", ns)).
			AppendFilter(query.NewRange("time").WithGTE(ct)))
	}
	b.AppendFilter(bi)

	if sf.WorkloadFilter != nil {
		bi := query.NewBool().WithMinimumShouldMatch(mini)
		for _, wk := range sf.WorkloadFilter {
			bi.AppendShould(query.NewRegex("kubernetes.pod_name.keyword", podNameRegex(wk)))
		}

		b.AppendFilter(bi)
	}

	b.AppendFilter(query.NewBool().
		AppendMultiShould(query.NewMultiMatchPhrase("kubernetes.pod_name.keyword", sf.PodFilter)).
		WithMinimumShouldMatch(mini))

	b.AppendFilter(query.NewBool().
		AppendMultiShould(query.NewMultiMatchPhrase("kubernetes.container_name.keyword", sf.ContainerFilter)).
		WithMinimumShouldMatch(mini))

	// fuzzy matching
	b.AppendFilter(query.NewBool().
		AppendMultiShould(query.NewMultiMatchPhrasePrefix("kubernetes.pod_name", sf.WorkloadSearch)).
		WithMinimumShouldMatch(mini))

	b.AppendFilter(query.NewBool().
		AppendMultiShould(query.NewMultiMatchPhrasePrefix("kubernetes.pod_name", sf.PodSearch)).
		WithMinimumShouldMatch(mini))

	b.AppendFilter(query.NewBool().
		AppendMultiShould(query.NewMultiMatchPhrasePrefix("kubernetes.container_name", sf.ContainerSearch)).
		WithMinimumShouldMatch(mini))

	b.AppendFilter(query.NewBool().
		AppendMultiShould(query.NewMultiMatchPhrasePrefix("log", sf.LogSearch)).
		WithMinimumShouldMatch(mini))

	r := query.NewRange("time")
	if !sf.Starttime.IsZero() {
		r.WithGTE(sf.Starttime)
	}
	if !sf.Endtime.IsZero() {
		r.WithLTE(sf.Endtime)
	}

	b.AppendFilter(r)

	return query.NewQuery().WithBool(b)
}

func podNameRegex(workloadName string) string {
	var regex string
	if len(workloadName) <= podNameMaxLength-replicaSetSuffixMaxLength-podNameSuffixLength {
		// match deployment pods, eg. <deploy>-579dfbcddd-24znw
		// replicaset rand string is limited to vowels
		// https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/apimachinery/pkg/util/rand/rand.go#L83
		regex += workloadName + "-[bcdfghjklmnpqrstvwxz2456789]{1,10}-[a-z0-9]{5}|"
		// match statefulset pods, eg. <sts>-0
		regex += workloadName + "-[0-9]+|"
		// match pods of daemonset or job, eg. <ds>-29tdk, <job>-5xqvl
		regex += workloadName + "-[a-z0-9]{5}"
	} else if len(workloadName) <= podNameMaxLength-podNameSuffixLength {
		replicaSetSuffixLength := podNameMaxLength - podNameSuffixLength - len(workloadName)
		regex += fmt.Sprintf("%s%d%s", workloadName+"-[bcdfghjklmnpqrstvwxz2456789]{", replicaSetSuffixLength, "}[a-z0-9]{5}|")
		regex += workloadName + "-[0-9]+|"
		regex += workloadName + "-[a-z0-9]{5}"
	} else {
		// Rand suffix may overwrites the workload name if the name is too long
		// This won't happen for StatefulSet because long name will cause ReplicaSet fails during StatefulSet creation.
		regex += workloadName[:podNameMaxLength-podNameSuffixLength+1] + "[a-z0-9]{5}|"
		regex += workloadName + "-[0-9]+"
	}
	return regex
}
