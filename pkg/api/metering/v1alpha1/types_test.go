package v1alpha1

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"

	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
)

func TestParseQueryParameter(t *testing.T) {
	queryParam := "level=LevelCluster&operation=query&time=1559347600&start=1559347200&end=1561939200&step=10m&sort_metric=meter_workspace_cpu_usage&sort_type=desc&page=1&limit=5&metrics_filter=meter_workspace_cpu_usage|meter_workspace_memory_usage&resources_filter=cpu&workspace=my-ws&namespace=my-ns&node=my-node&kind=deployment&workload=my-wl&pod=my-pod&applications=nignx&services=my-svc&storageclass=nfs&pvc_filter=my-pvc&cluster=my-cluster"
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost/tenant.kubesphere.io/v2alpha1/metering?%s", queryParam), nil)
	if err != nil {
		t.Fatal(err)
	}

	expected := Query{
		Level:            monitoring.Level(1),
		Operation:        "query",
		LabelSelector:    "",
		Time:             "1559347600",
		Start:            "1559347200",
		End:              "1561939200",
		Step:             "10m",
		Target:           "meter_workspace_cpu_usage",
		Order:            "desc",
		Page:             "1",
		Limit:            "5",
		MetricFilter:     "meter_workspace_cpu_usage|meter_workspace_memory_usage",
		ResourceFilter:   "cpu",
		NodeName:         "my-node",
		WorkspaceName:    "my-ws",
		NamespaceName:    "my-ns",
		WorkloadKind:     "deployment",
		WorkloadName:     "my-wl",
		PodName:          "my-pod",
		Applications:     "nignx",
		Services:         "my-svc",
		StorageClassName: "nfs",
		PVCFilter:        "my-pvc",
		Cluster:          "my-cluster",
	}

	request := restful.NewRequest(req)
	actual := ParseQueryParameter(request)
	assert.Equal(t, &expected, actual)
}
