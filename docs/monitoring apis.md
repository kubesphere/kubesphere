## KubeSphere Monitoring HTTP APIs

## APIs format

### Instant query format

`http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/<resources>?metrics_name=<metrics_name>&time=<timestamp>&timeout=<timeout>`

- time = < rfc3339 | unix_timestamp >: Evaluation timestamp. Optional.
- timeout = < duration >: Evaluation timeout, default 30 seconds. Optional.


### Range query format

`http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/<resources>?metrics_name=<metrics_name>&start=<start_timestamp>&end=<end_timestamp>&step=<step>&timeout=<timeout>`

- start = < rfc3339 | unix_timestamp >: start timestamp. Required.
- end = < rfc3339 | unix_timestamp >: end timestamp. Required.
- step = < step >: query step between start and end. Required.
- timeout = < duration >: Evaluation timeout, default 30 seconds. Optional.

> the "resources" parameter are the specific monitoring targets for both instant query and range query, can be: cluster, node, namespace, pod, container.
metrics_name is currently refers to cpu memory utilisation, this parameter may be a little difference amoung different monitor targets, Required

## Cluster Metrics

- query or range_query:

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/cluster?metrics_name=<metrics_name>&time=1536819000'

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/cluster?metrics_name=<metrics_name>&start=1536819000&end=1536819080&step=15s'

- **metrics_name = cluster_cpu_utilisation | cluster_memory_utilisation
**
- kubeSphere Endpoint and Response example:

		'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/cluster?metrics_name=cluster_cpu_utilisation&start=1536819000&end=1536819080&step=15s'

	```json
	{
		"status": "success",
		"data": {
			"resultType": "matrix",
			"result": [{
					"metric": {
						"__name__": ":node_cpu_utilisation:avg1m"
					},
					"values": [[1536819000, "0.07647222222165306"], [1536819015, "0.07647222222165306"], [1536819030, "0.075083333333411"], [1536819045, "0.075083333333411"], [1536819060, "0.08219444444597102"], [1536819075, "0.08219444444597102"]]
				}
			]
		}
	}

	```

## Node Metrics

- query a specific node metrics:

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/nodes/<node_id>?metrics_name=<metrics_name>&time=1536819000'

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/nodes/<node_id>?metrics_name=<metrics_name>&start=1536819000&end=1536819080&step=15s'

- query batch nodes metrics by using [re2 expression](https://github.com/google/re2/wiki/Syntax):

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/nodes?metrics_name=<metrics_name>&time=1536819000&nodes_re2=<nodes_re2>'

		 curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/nodes?metrics_name=<metrics_name>&start=1536819000&end=1536819080&step=15s&nodes_re2=<nodes_re2>'

- query all nodes metrics:

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/nodes?metrics_name=<metrics_name>&time=1536819000'

		 curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/nodes?metrics_name=<metrics_name>&start=1536819000&end=1536819080&step=15s'


- **metrics_name = node_cpu_utilisation | node_memory_utilisation | node_memory_available | node_memory_total**


- kubeSphere Endpoint and Response example:

		'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/nodes?start=1536819000&end=1536819080&step=15s&metrics_name=node_cpu_utilisation&nodes_re2=i-3.*'

	```json
	{
		"status": "success",
		"data": {
			"resultType": "matrix",
			"result": [{
					"metric": {
						"__name__": "node:node_cpu_utilisation:avg1m",
						"node": "i-39p7faw6"
					},
					"values": [[1536819000, "0.11183333333659295"], [1536819015, "0.11183333333659295"], [1536819030, "0.09633333332991845"], [1536819045, "0.09633333332991845"], [1536819060, "0.10758333333457515"], [1536819075, "0.10758333333457515"]]
				}, {
					"metric": {
						"__name__": "node:node_cpu_utilisation:avg1m",
						"node": "i-3ranic6j"
					},
					"values": [[1536819000, "0.06074999999642994"], [1536819015, "0.06074999999642994"], [1536819030, "0.06133333333612723"], [1536819045, "0.06133333333612723"], [1536819060, "0.06333333333022884"], [1536819075, "0.06333333333022884"]]
				}
			]
		}
	}
	```

## Namespace Metrics

- query a specific namespace metrics:

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/namespaces/<namespace_name>?time=1536819110&metrics_name=<metrics_name>'

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/namespaces/<namespace_name>?start=1536819110&end=1536819360&step=15s&metrics_name=<metrics_name>'

- query batch namespaces metrics by using [re2 expression](https://github.com/google/re2/wiki/Syntax):

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/namespaces?time=1536819110&namespaces_re2=<namespaces_re2>&metrics_name=<metrics_name>'

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/namespaces?start=1536819110&end=1536819360&step=15s&namespaces_re2=<namespaces_re2>&metrics_name=<metrics_name>'

- query all namespaces metrics:

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/namespaces?time=1536819110&metrics_name=<metrics_name>'

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/namespaces?start=1536819110&end=1536819360&step=15s&metrics_name=<metrics_name>'


- **metrics_name = namespace_cpu_utilisation | namespace_memory_utilisation**

- kubeSphere Endpoint and Response example:

		'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/namespaces?start=1536819310&end=1536819360&step=15s&namespaces_re2=k.*&metrics_name=namespace_memory_utilisation'

	```json
	{
		"status": "success",
		"data": {
			"resultType": "matrix",
			"result": [{
					"metric": {
						"__name__": "namespace:container_memory_usage_bytes:sum",
						"namespace": "kube-system"
					},
					"values": [[1536819310, "10597900288"], [1536819325, "10600857600"], [1536819340, "10600857600"], [1536819355, "10564550656"]]
				}, {
					"metric": {
						"__name__": "namespace:container_memory_usage_bytes:sum",
						"namespace": "kubesphere-controls-system"
					},
					"values": [[1536819310, "12242944"], [1536819325, "12242944"], [1536819340, "12242944"], [1536819355, "12242944"]]
				}, {
					"metric": {
						"__name__": "namespace:container_memory_usage_bytes:sum",
						"namespace": "kubesphere-system"
					},
					"values": [[1536819310, "334721024"], [1536819325, "334721024"], [1536819340, "334721024"], [1536819355, "334721024"]]
				}
			]
		}
	}

	```

## Pod Metrics

### Pods metrics in specific namespace

- query a specific pod metrics:

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/namespaces/<namespace_name>/pods/<pod_name>?time=1536819360&metrics_name=<metrics_name>'

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/namespaces/<namespace_name>/pods/<pod_name>?start=1536819110&end=1536819360&step=15s&metrics_name=<metrics_name>'
		
- query batch pods metrics by using [re2 expression](https://github.com/google/re2/wiki/Syntax):

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/namespaces/<namespace_name>/pods?time=1536819320&pod_re2=<pod_re2>&metrics_name=<metrics_name>'

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/namespaces/<namespace_name>/pods?start=1536819320&end=1536819360&step=15s&pod_re2=<pod_re2>&metrics_name=<metrics_name>'


- query all pods metrics:

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/namespaces/<namespace_name>/pods?time=1536819320&metrics_name=<metrics_name>'

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/namespaces/<namespace_name>/pods?start=1536819320&end=1536819360&step=15s&metrics_name=<metrics_name>'


- **metrics_name = pod_cpu_utilisation | pod_memory_utilisation | pod_memory_utilisation_wo_cache
**

- kubeSphere Endpoint and Response example:

		'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/namespaces/monitoring/pods?start=1536819320&end=1536819360&step=15s&pod_re2=alert.*&metrics_name=pod_memory_utilisation'

	```json
	{
		"status": "success",
		"data": {
			"resultType": "matrix",
			"result": [{
					"metric": {
						"namespace": "monitoring",
						"pod_name": "alertmanager-main-0"
					},
					"values": [[1536819320, "15462400"], [1536819335, "15462400"], [1536819350, "15462400"]]
				}, {
					"metric": {
						"namespace": "monitoring",
						"pod_name": "alertmanager-main-1"
					},
					"values": [[1536819320, "13905920"], [1536819335, "13881344"], [1536819350, "13881344"]]
				}, {
					"metric": {
						"namespace": "monitoring",
						"pod_name": "alertmanager-main-2"
					},
					"values": [[1536819320, "14737408"], [1536819335, "14737408"], [1536819350, "14737408"]]
				}
			]
		}
	}

	```

### Pods metrics in specific node

- query a specific pod metrics:

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/nodes/<node_id>/pods/<pod_name>?time=1536819360&metrics_name=<metrics_name>'

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/nodes/<node_id>/pods/<pod_name>?start=1536819110&end=1536819360&step=15s&metrics_name=<metrics_name>'

- query batch pods metrics by using [re2 expression](https://github.com/google/re2/wiki/Syntax):

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/nodes/<node_id>/pods?time=1536819320&pod_re2=<pod_re2>&metrics_name=<metrics_name>'

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/nodes/<node_id>/pods?start=1536819320&end=1536819360&step=15s&pod_re2=<pod_re2>&metrics_name=<metrics_name>'


- query all pods metrics:

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/nodes/<node_id>/pods?time=1536819320&metrics_name=<metrics_name>'

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/nodes/<node_id>/pods?start=1536819320&end=1536819360&step=15s&metrics_name=<metrics_name>'


- **metrics_name = pod_cpu_utilisation | pod_memory_utilisation | pod_memory_utilisation_wo_cache
**

- kubeSphere Endpoint and Response example1:

		'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/nodes/i-39p7faw6/pods?time=1536819320&metrics_name=pod_cpu_utilisation'

	```json
	{
		"status": "success",
		"data": {
			"resultType": "vector",
			"result": [{
					"metric": {
						"node": "i-39p7faw6",
						"pod": "kube-scheduler-i-39p7faw6"
					},
					"value": [1536819320, "0.01070628620000207"]
				}, {
					"metric": {
						"node": "i-39p7faw6",
						"pod": "calico-kube-controllers-66d47b98c9-rmxw2"
					},
					"value": [1536819320, "0.0009234044999933152"]
				},

				...

				{
					"metric": {
						"node": "i-39p7faw6",
						"pod": "fluent-bit-zsjm8"
					},
					"value": [1536819320, "0.0004903794999942571"]
				}, {
					"metric": {
						"node": "i-39p7faw6",
						"pod": "calico-node-chwtn"
					},
					"value": [1536819320, "0.01458303163343686"]
				}
			]
		}
	}
	```
- kubeSphere Endpoint and Response example2:

		'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/nodes/i-39p7faw6/pods/kube-addon-manager-i-39p7faw6?time=1536819360&metrics_name=pod_memory_utilisation_wo_cache'

	```json
	{
		"status": "success",
		"data": {
			"resultType": "vector",
			"result": [{
					"metric": {
						"node": "i-39p7faw6",
						"pod": "kube-addon-manager-i-39p7faw6"
					},
					"value": [1536819360, "7962624"]
				}
			]
		}
	}
	```

## Container Metrics

- query a specific container metrics:

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/namespaces/<namespace_name>/pods/<pod_name>/containers/<container_name>?time=1536819320&metrics_name=<metrics_name>'

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/namespaces/<namespace_name>/pods/<pod_name>/containers/<container_name>?start=1536819320&end=1536819360&step=15s&metrics_name=<metrics_name>'

- query batch containers metrics by using [re2 expression](https://github.com/google/re2/wiki/Syntax):

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/namespaces/<namespace_name>/pods/<pod_name>/containers?time=1536819320&container_re2=<container_re2>&metrics_name=<metrics_name>'

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/namespaces/<namespace_name>/pods/<pod_name>/containers?start=1536819320&end=1536819360&step=15s&container_re2=<container_re2>&metrics_name=<metrics_name>'

- query all containers metrics:

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/namespaces/<namespace_name>/pods/<pod_name>/containers?time=1536819320&metrics_name=<metrics_name>'

		curl 'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/namespaces/<namespace_name>/pods/<pod_name>/containers?start=1536819320&end=1536819360&step=15s&metrics_name=<metrics_name>'


- **metrics_name = container_cpu_utilisation | container_memory_utilisation_wo_cache | container_memory_utilisation
**

- kubeSphere Endpoint and Response example1:

		'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/namespaces/monitoring/pods/alertmanager-main-0/containers?start=1536819320&end=1536819360&step=15s&metrics_name=container_memory_utilisation_wo_cache'

	```json
	{
		"status": "success",
		"data": {
			"resultType": "matrix",
			"result": [{
					"metric": {
						"container_name": "POD",
						"namespace": "monitoring",
						"pod_name": "alertmanager-main-0"
					},
					"values": [[1536819320, "1245184"], [1536819335, "1245184"], [1536819350, "1245184"]]
				}, {
					"metric": {
						"container_name": "alertmanager",
						"namespace": "monitoring",
						"pod_name": "alertmanager-main-0"
					},
					"values": [[1536819320, "11866112"], [1536819335, "11866112"], [1536819350, "11866112"]]
				}, {
					"metric": {
						"container_name": "config-reloader",
						"namespace": "monitoring",
						"pod_name": "alertmanager-main-0"
					},
					"values": [[1536819320, "2347008"], [1536819335, "2347008"], [1536819350, "2347008"]]
				}
			]
		}
	}
	```

- kubeSphere Endpoint and Response example2:

		'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/namespaces/monitoring/pods/alertmanager-main-0/containers/config-reloader?time=1536819360&metrics_name=container_memory_utilisation'

	```json
	{
		"status": "success",
		"data": {
			"resultType": "vector",
			"result": [{
					"metric": {
						"__name__": "container_memory_usage_bytes",
						"container_name": "config-reloader",
						"endpoint": "https-metrics",
						"id": "/kubepods/burstable/poddfed46b2-af9c-11e8-825d-525444c70bfc/fe2bdc50d5294f85ede27f4629904a3e0beb1ad4037092d01498f2df14d2dbd1",
						"image": "sha256:3129a2ca29d75226dc5657a4629cdd5f38accda7f5b75bc5d5a76f5b4e0e5870",
						"instance": "192.168.0.11:10250",
						"job": "kubelet",
						"name": "k8s_config-reloader_alertmanager-main-0_monitoring_dfed46b2-af9c-11e8-825d-525444c70bfc_0",
						"namespace": "monitoring",
						"pod_name": "alertmanager-main-0",
						"service": "kubelet"
					},
					"value": [1536819360, "2347008"]
				}
			]
		}
	}
	```

- kubeSphere Endpoint and Response example3:

		'http://ks-apiserver.kubesphere-system.svc/api/v1alpha1/monitoring/namespaces/monitoring/pods/alertmanager-main-0/containers/config-reloader?time=1536819360&metrics_name=container_memory_utilisation_wo_cache'

	```json
	{
		"status": "success",
		"data": {
			"resultType": "vector",
			"result": [{
					"metric": {
						"container_name": "config-reloader",
						"namespace": "monitoring",
						"pod_name": "alertmanager-main-0"
					},
					"value": [1536819360, "2347008"]
				}
			]
		}
	}
	```









