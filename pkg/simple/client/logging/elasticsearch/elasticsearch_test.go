package elasticsearch

import (
	"github.com/google/go-cmp/cmp"
	"kubesphere.io/kubesphere/pkg/simple/client/logging"
	v5 "kubesphere.io/kubesphere/pkg/simple/client/logging/elasticsearch/versions/v5"
	v6 "kubesphere.io/kubesphere/pkg/simple/client/logging/elasticsearch/versions/v6"
	v7 "kubesphere.io/kubesphere/pkg/simple/client/logging/elasticsearch/versions/v7"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func MockElasticsearchService(pattern string, fakeResp string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(pattern, func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(fakeResp))
	})
	return httptest.NewServer(mux)
}

func TestDetectVersionMajor(t *testing.T) {
	var tests = []struct {
		description   string
		fakeResp      string
		expected      string
		expectedError bool
	}{
		{
			description: "detect es 6.x version number",
			fakeResp: `{
  "name" : "elasticsearch-logging-data-0",
  "cluster_name" : "elasticsearch",
  "cluster_uuid" : "uLm0838MSd60T1XEh5P2Qg",
  "version" : {
    "number" : "6.7.0",
    "build_flavor" : "oss",
    "build_type" : "docker",
    "build_hash" : "8453f77",
    "build_date" : "2019-03-21T15:32:29.844721Z",
    "build_snapshot" : false,
    "lucene_version" : "7.7.0",
    "minimum_wire_compatibility_version" : "5.6.0",
    "minimum_index_compatibility_version" : "5.0.0"
  },
  "tagline" : "You Know, for Search"
}`,
			expected:      ElasticV6,
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			es := MockElasticsearchService("/", test.fakeResp)
			defer es.Close()

			v, err := detectVersionMajor(es.URL)
			if err == nil && test.expectedError {
				t.Fatalf("expected error while got nothing")
			} else if err != nil && !test.expectedError {
				t.Fatal(err)
			}

			if v != test.expected {
				t.Fatalf("expected get version %s, but got %s", test.expected, v)
			}
		})
	}
}

func TestGetCurrentStats(t *testing.T) {
	var tests = []struct {
		description   string
		searchFilter  logging.SearchFilter
		fakeVersion   string
		fakeResp      string
		expected      logging.Statistics
		expectedError bool
	}{
		{
			description:  "[es 6.x] run as admin",
			searchFilter: logging.SearchFilter{},
			fakeVersion:  ElasticV6,
			fakeResp: `{
    "took": 171,
    "timed_out": false,
    "_shards": {
        "total": 10,
        "successful": 10,
        "skipped": 0,
        "failed": 0
    },
    "hits": {
        "total": 241222,
        "max_score": 1.0,
        "hits": [
            {
                "_index": "ks-logstash-log-2020.02.28",
                "_type": "flb_type",
                "_id": "Hn1GjXABMO5aQxyNsyxy",
                "_score": 1.0,
                "_source": {
                    "@timestamp": "2020-02-28T19:25:29.015Z",
                    "log": "  value: \"hostpath\"\n",
                    "time": "2020-02-28T19:25:29.015492329Z",
                    "kubernetes": {
                        "pod_name": "openebs-localpv-provisioner-55c66b57b4-jgtjc",
                        "namespace_name": "kube-system",
                        "host": "ks-allinone",
                        "container_name": "openebs-localpv-provisioner",
                        "docker_id": "cac01cd01cc79d8a8903ddbe6fbde9ac7497919a3f33c61861443703a9e08b39",
                        "container_hash": "25d789bcd3d12a4ba50bbb56eed1de33279d04352adbba8fd7e3b7b938aec806"
                    }
                }
            },
            {
                "_index": "ks-logstash-log-2020.02.28",
                "_type": "flb_type",
                "_id": "I31GjXABMO5aQxyNsyxy",
                "_score": 1.0,
                "_source": {
                    "@timestamp": "2020-02-28T19:25:33.103Z",
                    "log": "I0228 19:25:33.102631       1 controller.go:1040] provision \"kubesphere-system/redis-pvc\" class \"local\": trying to save persistentvolume \"pvc-be6d127d-9366-4ea8-b1ce-f30c1b3a447b\"\n",
                    "time": "2020-02-28T19:25:33.103075891Z",
                    "kubernetes": {
                        "pod_name": "openebs-localpv-provisioner-55c66b57b4-jgtjc",
                        "namespace_name": "kube-system",
                        "host": "ks-allinone",
                        "container_name": "openebs-localpv-provisioner",
                        "docker_id": "cac01cd01cc79d8a8903ddbe6fbde9ac7497919a3f33c61861443703a9e08b39",
                        "container_hash": "25d789bcd3d12a4ba50bbb56eed1de33279d04352adbba8fd7e3b7b938aec806"
                    }
                }
            },
            {
                "_index": "ks-logstash-log-2020.02.28",
                "_type": "flb_type",
                "_id": "JX1GjXABMO5aQxyNsyxy",
                "_score": 1.0,
                "_source": {
                    "@timestamp": "2020-02-28T19:25:33.113Z",
                    "log": "I0228 19:25:33.112200       1 controller.go:1088] provision \"kubesphere-system/redis-pvc\" class \"local\": succeeded\n",
                    "time": "2020-02-28T19:25:33.113110332Z",
                    "kubernetes": {
                        "pod_name": "openebs-localpv-provisioner-55c66b57b4-jgtjc",
                        "namespace_name": "kube-system",
                        "host": "ks-allinone",
                        "container_name": "openebs-localpv-provisioner",
                        "docker_id": "cac01cd01cc79d8a8903ddbe6fbde9ac7497919a3f33c61861443703a9e08b39",
                        "container_hash": "25d789bcd3d12a4ba50bbb56eed1de33279d04352adbba8fd7e3b7b938aec806"
                    }
                }
            },
            {
                "_index": "ks-logstash-log-2020.02.28",
                "_type": "flb_type",
                "_id": "Kn1GjXABMO5aQxyNsyxy",
                "_score": 1.0,
                "_source": {
                    "@timestamp": "2020-02-28T19:25:34.168Z",
                    "log": "  value: \"hostpath\"\n",
                    "time": "2020-02-28T19:25:34.168983384Z",
                    "kubernetes": {
                        "pod_name": "openebs-localpv-provisioner-55c66b57b4-jgtjc",
                        "namespace_name": "kube-system",
                        "host": "ks-allinone",
                        "container_name": "openebs-localpv-provisioner",
                        "docker_id": "cac01cd01cc79d8a8903ddbe6fbde9ac7497919a3f33c61861443703a9e08b39",
                        "container_hash": "25d789bcd3d12a4ba50bbb56eed1de33279d04352adbba8fd7e3b7b938aec806"
                    }
                }
            },
            {
                "_index": "ks-logstash-log-2020.02.28",
                "_type": "flb_type",
                "_id": "LH1GjXABMO5aQxyNsyxy",
                "_score": 1.0,
                "_source": {
                    "@timestamp": "2020-02-28T19:25:34.168Z",
                    "log": "  value: \"/var/openebs/local/\"\n",
                    "time": "2020-02-28T19:25:34.168997393Z",
                    "kubernetes": {
                        "pod_name": "openebs-localpv-provisioner-55c66b57b4-jgtjc",
                        "namespace_name": "kube-system",
                        "host": "ks-allinone",
                        "container_name": "openebs-localpv-provisioner",
                        "docker_id": "cac01cd01cc79d8a8903ddbe6fbde9ac7497919a3f33c61861443703a9e08b39",
                        "container_hash": "25d789bcd3d12a4ba50bbb56eed1de33279d04352adbba8fd7e3b7b938aec806"
                    }
                }
            },
            {
                "_index": "ks-logstash-log-2020.02.28",
                "_type": "flb_type",
                "_id": "NX1GjXABMO5aQxyNsyxy",
                "_score": 1.0,
                "_source": {
                    "@timestamp": "2020-02-28T19:25:42.868Z",
                    "log": "I0228 19:25:42.868413       1 config.go:83] SC local has config:- name: StorageType\n",
                    "time": "2020-02-28T19:25:42.868578188Z",
                    "kubernetes": {
                        "pod_name": "openebs-localpv-provisioner-55c66b57b4-jgtjc",
                        "namespace_name": "kube-system",
                        "host": "ks-allinone",
                        "container_name": "openebs-localpv-provisioner",
                        "docker_id": "cac01cd01cc79d8a8903ddbe6fbde9ac7497919a3f33c61861443703a9e08b39",
                        "container_hash": "25d789bcd3d12a4ba50bbb56eed1de33279d04352adbba8fd7e3b7b938aec806"
                    }
                }
            },
            {
                "_index": "ks-logstash-log-2020.02.28",
                "_type": "flb_type",
                "_id": "Q31GjXABMO5aQxyNsyxy",
                "_score": 1.0,
                "_source": {
                    "@timestamp": "2020-02-28T19:26:13.881Z",
                    "log": "- name: BasePath\n",
                    "time": "2020-02-28T19:26:13.881180681Z",
                    "kubernetes": {
                        "pod_name": "openebs-localpv-provisioner-55c66b57b4-jgtjc",
                        "namespace_name": "kube-system",
                        "host": "ks-allinone",
                        "container_name": "openebs-localpv-provisioner",
                        "docker_id": "cac01cd01cc79d8a8903ddbe6fbde9ac7497919a3f33c61861443703a9e08b39",
                        "container_hash": "25d789bcd3d12a4ba50bbb56eed1de33279d04352adbba8fd7e3b7b938aec806"
                    }
                }
            },
            {
                "_index": "ks-logstash-log-2020.02.28",
                "_type": "flb_type",
                "_id": "S31GjXABMO5aQxyNsyxy",
                "_score": 1.0,
                "_source": {
                    "@timestamp": "2020-02-28T19:26:14.597Z",
                    "log": "  value: \"/var/openebs/local/\"\n",
                    "time": "2020-02-28T19:26:14.597702238Z",
                    "kubernetes": {
                        "pod_name": "openebs-localpv-provisioner-55c66b57b4-jgtjc",
                        "namespace_name": "kube-system",
                        "host": "ks-allinone",
                        "container_name": "openebs-localpv-provisioner",
                        "docker_id": "cac01cd01cc79d8a8903ddbe6fbde9ac7497919a3f33c61861443703a9e08b39",
                        "container_hash": "25d789bcd3d12a4ba50bbb56eed1de33279d04352adbba8fd7e3b7b938aec806"
                    }
                }
            },
            {
                "_index": "ks-logstash-log-2020.02.28",
                "_type": "flb_type",
                "_id": "TH1GjXABMO5aQxyNsyxy",
                "_score": 1.0,
                "_source": {
                    "@timestamp": "2020-02-28T19:26:14.597Z",
                    "log": "I0228 19:26:14.597007       1 provisioner_hostpath.go:42] Creating volume pvc-c3b1e67f-00d2-407d-8c45-690bb273c16a at ks-allinone:/var/openebs/local/pvc-c3b1e67f-00d2-407d-8c45-690bb273c16a\n",
                    "time": "2020-02-28T19:26:14.597708432Z",
                    "kubernetes": {
                        "pod_name": "openebs-localpv-provisioner-55c66b57b4-jgtjc",
                        "namespace_name": "kube-system",
                        "host": "ks-allinone",
                        "container_name": "openebs-localpv-provisioner",
                        "docker_id": "cac01cd01cc79d8a8903ddbe6fbde9ac7497919a3f33c61861443703a9e08b39",
                        "container_hash": "25d789bcd3d12a4ba50bbb56eed1de33279d04352adbba8fd7e3b7b938aec806"
                    }
                }
            },
            {
                "_index": "ks-logstash-log-2020.02.28",
                "_type": "flb_type",
                "_id": "UX1GjXABMO5aQxyNsyxy",
                "_score": 1.0,
                "_source": {
                    "@timestamp": "2020-02-28T19:26:15.920Z",
                    "log": "I0228 19:26:15.915071       1 event.go:221] Event(v1.ObjectReference{Kind:\"PersistentVolumeClaim\", Namespace:\"kubesphere-system\", Name:\"mysql-pvc\", UID:\"1e87deb5-eaec-475f-8eb6-8613b3be80a4\", APIVersion:\"v1\", ResourceVersion:\"2397\", FieldPath:\"\"}): type: 'Normal' reason: 'ProvisioningSucceeded' Successfully provisioned volume pvc-1e87deb5-eaec-475f-8eb6-8613b3be80a4\n",
                    "time": "2020-02-28T19:26:15.920650572Z",
                    "kubernetes": {
                        "pod_name": "openebs-localpv-provisioner-55c66b57b4-jgtjc",
                        "namespace_name": "kube-system",
                        "host": "ks-allinone",
                        "container_name": "openebs-localpv-provisioner",
                        "docker_id": "cac01cd01cc79d8a8903ddbe6fbde9ac7497919a3f33c61861443703a9e08b39",
                        "container_hash": "25d789bcd3d12a4ba50bbb56eed1de33279d04352adbba8fd7e3b7b938aec806"
                    }
                }
            }
        ]
    },
    "aggregations": {
        "container_count": {
            "value": 93
        }
    }
}`,
			expected: logging.Statistics{
				Containers: 93,
				Logs:       241222,
			},
			expectedError: false,
		},
		{
			description: "[es 6.x] index not found",
			searchFilter: logging.SearchFilter{
				NamespaceFilter: map[string]time.Time{
					"workspace-1-project-a": time.Unix(1582000000, 0),
					"workspace-1-project-b": time.Unix(1582333333, 0),
				},
			},
			fakeVersion: ElasticV6,
			fakeResp: `{
   "error": {
       "root_cause": [
           {
               "type": "index_not_found_exception",
               "reason": "no such index",
               "resource.type": "index_or_alias",
               "resource.id": "ks-lsdfsdfsdfs",
               "index_uuid": "_na_",
               "index": "ks-lsdfsdfsdfs"
           }
       ],
       "type": "index_not_found_exception",
       "reason": "no such index",
       "resource.type": "index_or_alias",
       "resource.id": "ks-lsdfsdfsdfs",
       "index_uuid": "_na_",
       "index": "ks-lsdfsdfsdfs"
   },
   "status": 404
}`,
			expected: logging.Statistics{
				Containers: 0,
				Logs:       0,
			},
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			es := MockElasticsearchService("/", test.fakeResp)
			defer es.Close()

			clientv5 := Elasticsearch{c: v5.New(es.URL, "ks-logstash-log")}
			clientv6 := Elasticsearch{c: v6.New(es.URL, "ks-logstash-log")}
			clientv7 := Elasticsearch{c: v7.New(es.URL, "ks-logstash-log")}

			var stats logging.Statistics
			var err error
			switch test.fakeVersion {
			case ElasticV5:
				stats, err = clientv5.GetCurrentStats(test.searchFilter)
			case ElasticV6:
				stats, err = clientv6.GetCurrentStats(test.searchFilter)
			case ElasticV7:
				stats, err = clientv7.GetCurrentStats(test.searchFilter)
			}

			if err != nil && !test.expectedError {
				t.Fatal(err)
			} else if diff := cmp.Diff(stats, test.expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", test.expected, diff)
			}
		})
	}
}
