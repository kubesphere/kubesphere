#! /bin/bash

curl -d '{"grafanaDashboardUrl":"https://grafana.com/api/dashboards/7362/revisions/5/download", "description":"this is a test clusterdashboard."}' -H "Content-Type: application/json" localhost:9090/kapis/monitoring.kubesphere.io/v1alpha3/clusterdashboards/test1/template
curl -d '{"grafanaDashboardUrl":"https://grafana.com/api/dashboards/7362/revisions/5/download", "description":"this is a test dashboard."}' -H "Content-Type: application/json" localhost:9090/kapis/monitoring.kubesphere.io/v1alpha3/namespaces/default/dashboards/test2/template