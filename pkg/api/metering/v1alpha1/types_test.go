package v1alpha1

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
)

func TestParseQueryParameter(t *testing.T) {
	queryParam := "/tenant.kubesphere.io/v2alpha1/metering?level=LevelCluster"
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost?%s", queryParam), nil)
	if err != nil {
		t.Fatal(err)
	}

	request := restful.NewRequest(req)
	ret := ParseQueryParameter(request)
	assert.NotNil(t, ret)
}
