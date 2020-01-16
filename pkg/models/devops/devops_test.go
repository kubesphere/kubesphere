package devops

import (
	"kubesphere.io/kubesphere/pkg/simple/client/devops/fake"
	"net/http"
	"testing"
)

const baseUrl = "http://127.0.0.1/kapis/devops.kubesphere.io/v1alpha2/"

func TestGetNodesDetail(t *testing.T) {
	fakeData := make(map[string][]byte)
	PipelineRunNodes := `[
  {
    "displayName": "Deploy to Kubernetes",
    "id": "1",
    "result": "SUCCESS"
  },
  {
    "displayName": "Deploy to Kubernetes",
    "id": "2",
    "result": "SUCCESS"
  },
  {
    "displayName": "Deploy to Kubernetes",
    "id": "3",
    "result": "SUCCESS"
  }
]`

	NodeSteps := `[
  {
    "displayName": "Deploy to Kubernetes",
    "id": "21",
    "result": "SUCCESS"
  }
]`

	fakeData["NodeSteps"] = []byte(NodeSteps)
	fakeData["PipelineRunNodes"] = []byte(PipelineRunNodes)

	devopsOperator := NewDevopsOperator(fake.NewFakeDevops(fakeData))

	httpReq, _ := http.NewRequest(http.MethodGet, baseUrl+"devops/project/pipelines/pipeline/runs/run/nodesdetail/?limit=10000", nil)

	nodesDetails, err := devopsOperator.GetNodesDetail("project", "pipeline", "run", httpReq)
	if err != nil || nodesDetails == nil {
		t.Fatalf("should not get error %+v", err)
	}

	for _, v := range nodesDetails {
		if v.Steps[0].ID == "" {
			t.Fatalf("Can not get any step.")
		}
	}
}

func TestGetBranchNodesDetail(t *testing.T) {
	fakeData := make(map[string][]byte)

	BranchPipelineRunNodes := `[
  {
    "displayName": "Deploy to Kubernetes",
    "id": "1",
    "result": "SUCCESS"
  },
  {
    "displayName": "Deploy to Kubernetes",
    "id": "2",
    "result": "SUCCESS"
  },
  {
    "displayName": "Deploy to Kubernetes",
    "id": "3",
    "result": "SUCCESS"
  }
]`

	BranchNodeSteps := `[
  {
    "displayName": "Deploy to Kubernetes",
    "id": "21",
    "result": "SUCCESS"
  }
]`

	fakeData["BranchNodeSteps"] = []byte(BranchNodeSteps)
	fakeData["BranchPipelineRunNodes"] = []byte(BranchPipelineRunNodes)

	devopsOperator := NewDevopsOperator(fake.NewFakeDevops(fakeData))

	httpReq, _ := http.NewRequest(http.MethodGet, baseUrl+"devops/project/pipelines/pipeline/branchs/branc/runs/run/nodesdetail/?limit=10000", nil)

	nodesDetails, err := devopsOperator.GetBranchNodesDetail("project", "pipeline", "branch", "run", httpReq)
	if err != nil || nodesDetails == nil {
		t.Fatalf("should not get error %+v", err)
	}

	for _, v := range nodesDetails {
		if v.Steps[0].ID == "" {
			t.Fatalf("Can not get any step.")
		}
	}
}
