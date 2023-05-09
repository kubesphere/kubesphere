package v1alpha1

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/stretchr/testify/assert"
)

func TestParseQueryParameter(t *testing.T) {
	queryParam := "operation=query&workspace_filter=my-ws,demo-ws&workspace_search=my,demo&involved_object_namespace_filter=my-ns,my-test&involved_object_namespace_search=my&involved_object_name_filter=my-involvedObject,demo-involvedObject&involved_object_name_search=involvedObject&involved_object_kind_filter=involvedObject.kind&reason_filter=reason.filter&reason_search=reason&message_search=message&type_filter=Normal&interval=15m&start_time=1136214245&end_time=1136214245&from=0&sort=desc"
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost/tenant.kubesphere.io/v2alpha1/events?%s", queryParam), nil)
	if err != nil {
		t.Fatal(err)
	}

	expected := Query{
		Operation:                     "query",
		WorkspaceFilter:               "my-ws,demo-ws",
		WorkspaceSearch:               "my,demo",
		InvolvedObjectNamespaceFilter: "my-ns,my-test",
		InvolvedObjectNamespaceSearch: "my",
		InvolvedObjectNameFilter:      "my-involvedObject,demo-involvedObject",
		InvolvedObjectNameSearch:      "involvedObject",
		InvolvedObjectKindFilter:      "involvedObject.kind",
		ReasonFilter:                  "reason.filter",
		ReasonSearch:                  "reason",
		MessageSearch:                 "message",
		TypeFilter:                    "Normal",

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
