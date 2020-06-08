package auditing

import (
	"bytes"
	"github.com/google/go-cmp/cmp"
	"k8s.io/apiserver/pkg/apis/audit"
	"net/http"
	"testing"
)

func TestGetPathParameter(t *testing.T) {
	uri := "/kapis/devops.kubesphere.io/v1alpha3/devops/{devops}/pipelines/{pipeline}"
	url := "/kapis/devops.kubesphere.io/v1alpha3/devops/devproject/pipelines/pl"

	p := getPathParameter(uri, url, "devops")

	if diff := cmp.Diff("devproject", p); len(diff) != 0 {
		t.Errorf("%T differ (-got, +expected), %s", "value", diff)
	}
}

func TestGetName(t *testing.T) {

	c := &Config{
		URI:         "/kapis/devops.kubesphere.io/v1alpha3/devops/{devops}/pipelines/{pipeline}",
		Method:      "Get",
		Resource:    "devops",
		Subresource: "pipeline",
		NamPath: []string{
			"parameter.devops",
			"parameter.pipeline",
		},
	}

	a := auditing{}
	req, _ := http.NewRequest(http.MethodGet, "/kapis/devops.kubesphere.io/v1alpha3/devops/devproject/pipelines/pl", nil)
	name := a.getName(c, nil, req)

	if diff := cmp.Diff("devproject.pl", name); len(diff) != 0 {
		t.Errorf("%T differ (-got, +expected), %s", "name", diff)
	}
}

func TestGetBody(t *testing.T) {

	c := &Config{
		URI:         "/kapis/tenant.kubesphere.io/v1alpha2/workspaces",
		Method:      "Post",
		Resource:    "devops",
		Subresource: "pipeline",
		NamPath: []string{
			"body.metadata.name",
		},
	}

	body := "{   \"apiVersion\": \"tenant.kubesphere.io/v1alpha1\",   \"kind\": \"Workspace\",   \"metadata\": {     \"annotations\": {       \"kubesphere.io/alias-name\": \"\",       \"kubesphere.io/creator\": \"admin\"     },     \"name\": \"auditing\"   },   \"spec\": {     \"manager\": \"admin\"   } }"
	req, _ := http.NewRequest(http.MethodGet, "/kapis/tenant.kubesphere.io/v1alpha2/workspaces", bytes.NewBufferString(body))

	a := auditing{}
	bs := a.getBody(audit.LevelMetadata, c, req)

	if diff := cmp.Diff(body, string(bs)); len(diff) != 0 {
		t.Errorf("%T differ (-got, +expected), %s", "body", diff)
	}
}
