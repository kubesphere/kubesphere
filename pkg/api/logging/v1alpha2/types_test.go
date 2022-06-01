package v1alpha2

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
)

func TestParseQueryParameter(t *testing.T) {
	// default operation -- query
	queryParam := "start_time=1136214245&end_time=1136214245&from=0"
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost/tenant.kubesphere.io/v2alpha1/logs?%s", queryParam), nil)
	if err != nil {
		t.Fatal(err)
	}

	request := restful.NewRequest(req)
	_, err = ParseQueryParameter(request)
	assert.NoError(t, err)

	// interval operation
	queryParamInterval := "operation=interval&start_time=1136214245&end_time=1136214245&from=0"
	reqInterval, err := http.NewRequest("GET", fmt.Sprintf("http://localhost/tenant.kubesphere.io/v2alpha1/logs?%s", queryParamInterval), nil)
	if err != nil {
		t.Fatal(err)
	}

	requestInterval := restful.NewRequest(reqInterval)
	_, err = ParseQueryParameter(requestInterval)
	assert.NoError(t, err)
}
