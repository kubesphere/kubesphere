package alerting

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListRules(t *testing.T) {
	var tests = []struct {
		description string
		fakeCode    int
		fakeResp    string
		expectError bool
	}{{
		description: "list alerting rules from prometheus endpoint",
		expectError: false,
		fakeCode:    200,
		fakeResp: `
{
    "status": "success",
    "data": {
        "groups": [
            {
                "name": "kubernetes-resources",
                "file": "/etc/prometheus/rules/prometheus-k8s-rulefiles-0/kubesphere-monitoring-system-prometheus-k8s-rules.yaml",
                "rules": [
                    {
                        "state": "firing",
                        "name": "KubeCPUOvercommit",
                        "query": "sum(namespace:kube_pod_container_resource_requests_cpu_cores:sum) / sum(kube_node_status_allocatable{resource='cpu'}) > (count(kube_node_status_allocatable{resource='cpu'}) - 1) / count(kube_node_status_allocatable{resource='cpu'})",
                        "duration": 300,
                        "labels": {
                            "severity": "warning"
                        },
                        "annotations": {
                            "message": "Cluster has overcommitted CPU resource requests for Pods and cannot tolerate node failure.",
                            "runbook_url": "https://github.com/kubernetes-monitoring/kubernetes-mixin/tree/master/runbook.md#alert-name-kubecpuovercommit"
                        },
                        "alerts": [
                            {
                                "labels": {
                                    "alertname": "KubeCPUOvercommit",
                                    "severity": "warning"
                                },
                                "annotations": {
                                    "message": "Cluster has overcommitted CPU resource requests for Pods and cannot tolerate node failure.",
                                    "runbook_url": "https://github.com/ kubernetes-monitoring/kubernetes-mixin/tree/master/runbook.md#alert-name-kubecpuovercommit"
                                },
                                "state": "firing",
                                "activeAt": "2020-09-22T06:18:47.55260138Z",
                                "value": "4.405e-01"
                            }
                        ],
                        "health": "ok",
                        "evaluationTime": 0.000894038,
                        "lastEvaluation": "2020-09-22T08:57:17.566233983Z",
                        "type": "alerting"
                    }
                ]
            }
        ]
    }
}
`,
	}}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			mock := MockService(epRules, test.fakeCode, test.fakeResp)
			defer mock.Close()
			c, e := NewRuleClient(&Options{PrometheusEndpoint: mock.URL})
			if e != nil {
				t.Fatal(e)
			}
			rgs, e := c.PrometheusRules(context.TODO())
			if test.expectError {
			} else {
				if e != nil {
					t.Fatal(e)
				} else if len(rgs) == 1 && len(rgs[0].Rules) == 1 {

				} else {
					t.Fatalf("expect %d group and %d rule but got %d group and %d rule", 1, 1, len(rgs), len(rgs[0].Rules))
				}
			}

		})
	}
}

func MockService(pattern string, fakeCode int, fakeResp string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(pattern, func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(fakeCode)
		res.Write([]byte(fakeResp))
	})
	return httptest.NewServer(mux)
}
