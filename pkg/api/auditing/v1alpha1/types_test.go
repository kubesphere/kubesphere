package v1alpha1

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
)

func TestParseQueryParameter(t *testing.T) {
	queryParam := "operation=query&workspace_filter=my-ws,demo-ws&workspace_search=my,demo&objectref_namespace_filter=my-ns,my-test&objectref_namespace_search=my&objectref_name_filter=my-ref,demo-ref&objectref_name_search=ref&source_ip_search=192.168.&user_filter=user1&user_search=my,demo&group_search=my,demo&start_time=1136214245&end_time=1136214245&from=0&sort=desc"
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost/tenant.kubesphere.io/v2alpha1/auditing/events?%s", queryParam), nil)
	if err != nil {
		t.Fatal(err)
	}

	expected := Query{
		Operation:                "query",
		WorkspaceFilter:          "my-ws,demo-ws",
		WorkspaceSearch:          "my,demo",
		ObjectRefNamespaceFilter: "my-ns,my-test",
		ObjectRefNamespaceSearch: "my",
		ObjectRefNameFilter:      "my-ref,demo-ref",
		ObjectRefNameSearch:      "ref",
		UserFilter:               "user1",
		UserSearch:               "my,demo",
		GroupSearch:              "my,demo",
		SourceIpSearch:           "192.168.",

		StartTime: time.Unix(1136214245, 0),
		EndTime:   time.Unix(1136214245, 0),

		Interval: "15m",
		Sort:     "desc",
		From:     int64(0),
		Size:     int64(10),
	}

	request := restful.NewRequest(req)
	actual, err := ParseQueryParameter(request)
	assert.NoError(t, err)
	assert.Equal(t, &expected, actual)
}
