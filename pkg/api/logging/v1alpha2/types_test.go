package v1alpha2

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/stretchr/testify/assert"
)

func TestParseQueryParameter(t *testing.T) {
	// default operation -- query
	defaultParam := "namespaces=default,my-ns,my-test&namespace_query=default,my&workloads=my-wl,demo-wl&workload_query=wl&pods=my-po,demo-po&pod_query=po&containers=my-cont,demo-cont&container_query=cont&log_query=ERR&start_time=1136214245&end_time=1136214245&from=0&sort=desc"
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost/tenant.kubesphere.io/v2alpha1/logs?%s", defaultParam), nil)
	if err != nil {
		t.Fatal(err)
	}

	expected := Query{
		Operation:       OperationQuery,
		NamespaceFilter: "default,my-ns,my-test",
		NamespaceSearch: "default,my",
		WorkloadFilter:  "my-wl,demo-wl",
		WorkloadSearch:  "wl",
		PodFilter:       "my-po,demo-po",
		PodSearch:       "po",
		ContainerFilter: "my-cont,demo-cont",
		ContainerSearch: "cont",
		LogSearch:       "ERR",

		StartTime: time.Unix(1136214245, 0),
		EndTime:   time.Unix(1136214245, 0),

		Sort: OrderDescending,
		From: int64(0),
		Size: int64(10),
	}

	request := restful.NewRequest(req)
	actual, err := ParseQueryParameter(request)
	assert.NoError(t, err)
	assert.Equal(t, &expected, actual)

	// histogram operation
	queryParamInterval := "operation=histogram&interval=15m&" + defaultParam
	reqInterval, err := http.NewRequest("GET", fmt.Sprintf("http://localhost/tenant.kubesphere.io/v2alpha1/logs?%s", queryParamInterval), nil)
	if err != nil {
		t.Fatal(err)
	}

	expected = Query{
		Operation:       OperationHistogram,
		NamespaceFilter: "default,my-ns,my-test",
		NamespaceSearch: "default,my",
		WorkloadFilter:  "my-wl,demo-wl",
		WorkloadSearch:  "wl",
		PodFilter:       "my-po,demo-po",
		PodSearch:       "po",
		ContainerFilter: "my-cont,demo-cont",
		ContainerSearch: "cont",
		LogSearch:       "ERR",

		StartTime: time.Unix(1136214245, 0),
		EndTime:   time.Unix(1136214245, 0),

		Interval: DefaultInterval,
		From:     int64(0),
		Size:     int64(0),
	}

	requestInterval := restful.NewRequest(reqInterval)
	actual, err = ParseQueryParameter(requestInterval)
	assert.NoError(t, err)
	assert.Equal(t, &expected, actual)
}
