package devops

import (
	"kubesphere.io/kubesphere/pkg/simple/client/devops/fake"
	"net/http"
	"testing"
)

const baseUrl = "http://127.0.0.1/kapis/devops.kubesphere.io/v1alpha2/"

func TestGetNodesDetail(t *testing.T) {
	fakeDevops := fake.NewFakeDevops()

	devopsOperator := NewDevopsOperator(fakeDevops)

	httpReq, _ := http.NewRequest(http.MethodGet, baseUrl+"devops/project/pipelines/pipeline/branches/brnach/runs/run/nodesdetail/?limit=10000", nil)

	_, _ = devopsOperator.GetNodesDetail("projectName", "pipelineName", "runId", httpReq)
}
