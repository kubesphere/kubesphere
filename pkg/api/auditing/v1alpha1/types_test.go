package v1alpha1

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
)

func TestParseQueryParameter(t *testing.T) {
	queryParam := "operation=query&workspace_filter=my-ws,demo-ws&workspace_search=my,demo&start_time=1136214245&end_time=1136214245&from=0"
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost/tenant.kubesphere.io/v2alpha1/auditing/events?%s", queryParam), nil)
	if err != nil {
		t.Fatal(err)
	}

	request := restful.NewRequest(req)
	_, err = ParseQueryParameter(request)
	assert.NoError(t, err)
}
