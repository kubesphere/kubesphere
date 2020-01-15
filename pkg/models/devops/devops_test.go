package devops

import (
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/devops/fake"
	"net/http"
	"testing"
)

const baseUrl = "http://127.0.0.1/kapis/devops.kubesphere.io/v1alpha2/"

func TestGetNodesDetail(t *testing.T) {
	PipelineRunNodes := []devops.PipelineRunNodes{
		{ID: "0", Result: "true", DisplayName: "fakeBranchPipelineRunNode1"},
		{ID: "1", Result: "true", DisplayName: "fakeBranchPipelineRunNode2"},
		{ID: "2", Result: "true", DisplayName: "fakeBranchPipelineRunNode3"},
	}

	NodeSteps := []devops.NodeSteps{
		{ID:"1",Result:"true",DisplayName:"fakeBranchNodeStep"},
	}

	//dict := make(map[string][]devops.NodeSteps)
	//dict["1"] = NodeSteps

	fakeDevops := fake.NewFakeDevopsNodesDetail(PipelineRunNodes, NodeSteps)

	devopsOperator := NewDevopsOperator(fakeDevops)

	httpReq, _ := http.NewRequest(http.MethodGet, baseUrl+"devops/project/pipelines/pipeline/runs/run/nodesdetail/?limit=10000", nil)

	nodesDetails, err := devopsOperator.GetNodesDetail("project", "pipeline", "run", httpReq)
	if err != nil || nodesDetails == nil {
		t.Fatalf("should not get error %+v", err)
	}

	for i, v := range nodesDetails {
		if v.ID != string(i) {
			t.Fatalf("Node id %s and step od %s should equal", v.ID, v.Steps[0].ID)
		}
		if v.Steps[0].ID != NodeSteps[0].ID {
			t.Fatalf("Get step id %s but intput is %s.", v.Steps[0].ID, NodeSteps[0].ID)
		}
	}
}

func TestGetBranchNodesDetail(t *testing.T) {
	PipelineRunNodes := []devops.PipelineRunNodes{
		{ID: "1", Result: "true", DisplayName: "fakeBranchPipelineRunNode1"},
		{ID: "2", Result: "true", DisplayName: "fakeBranchPipelineRunNode2"},
		{ID: "3", Result: "true", DisplayName: "fakeBranchPipelineRunNode3"},
	}

	NodeSteps := []devops.NodeSteps{
		{ID:"1",Result:"true",DisplayName:"fakeBranchNodeStep"},
	}

	fakeDevops := fake.NewFakeDevopsNodesDetail(PipelineRunNodes, NodeSteps)

	devopsOperator := NewDevopsOperator(fakeDevops)

	httpReq, _ := http.NewRequest(http.MethodGet, baseUrl+"devops/project/pipelines/pipeline/branchs/branc/runs/run/nodesdetail/?limit=10000", nil)

	nodesDetails, err := devopsOperator.GetBranchNodesDetail("project", "pipeline","branch", "run", httpReq)
	if err != nil || nodesDetails == nil {
		t.Fatalf("should not get error %+v", err)
	}

	for i, v := range nodesDetails {
		if v.ID != string(i) {
			t.Fatalf("Node id %s and step od %s should equal", v.ID, v.Steps[0].ID)
		}
		if v.Steps[0].ID != NodeSteps[0].ID {
			t.Fatalf("Get step id %s but intput is %s.", v.Steps[0].ID, NodeSteps[0].ID)
		}
	}
}
