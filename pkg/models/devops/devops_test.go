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

	httpReq, _ := http.NewRequest(http.MethodGet, baseUrl+"devops/project/pipelines/pipeline/runs/run/nodesdetail/?limit=10000", nil)

	nodesDetails, err := devopsOperator.GetNodesDetail("project", "pipeline", "run", httpReq)
	if err != nil || nodesDetails == nil {
		t.Fatalf("should not get error %+v", err)
	}

	for _, v := range nodesDetails{
		if v.ID != v.Steps[0].ID {
			t.Fatalf("Node id %s and step od %s should equal", v.ID, v.Steps[0].ID)
		}
	}
}

func TestGetBranchNodesDetail(t *testing.T) {
	fakeDevops := fake.NewFakeDevops()

	devopsOperator := NewDevopsOperator(fakeDevops)

	httpReq, _ := http.NewRequest(http.MethodGet, baseUrl+"devops/project/pipelines/pipeline/branches/branch/runs/run/nodesdetail/?limit=10000", nil)

	nodesDetails, err := devopsOperator.GetBranchNodesDetail("project", "pipeline","branch", "run", httpReq)
	if err != nil || nodesDetails == nil {
		t.Fatalf("should not get error %+v", err)
	}

	for _, v := range nodesDetails{
		if v.ID != v.Steps[0].ID {
			t.Fatalf("Node id %s and step od %s should equal", v.ID, v.Steps[0].ID)
		}
	}
}