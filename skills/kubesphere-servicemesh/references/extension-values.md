## Extension Values

| Name | Meaning | Default | Range |
| --- | --- | --- | --- |
| backend.istio.meshConfig.defaultConfig.tracing.sampling | Tracing Sampling Rate | 1.0 | 1-100 |
| backend.kiali.prometheus_url | Prometheus URL | http://prometheus-k8s.kubesphere-monitoring-system.svc:9090 | |
| backend.jaeger.storage.options.es.server-urls | OpenSearch/ES URL | https://opensearch-cluster-data.kubesphere-logging-system.svc:9200 | |
| backend.jaeger.storage.options.es.username | OpenSearch/ES Username | admin | |
| backend.jaeger.storage.options.es.password | OpenSearch/ES Password | admin | |
| backend.jaeger.storage.options.secretName | OpenSearch/ES Secret Name | | |