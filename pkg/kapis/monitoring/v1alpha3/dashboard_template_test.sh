#! /bin/bash

curl -d '{"grafanaDashboardUrl":"https://grafana.com/api/dashboards/7362/revisions/5/download"}' -H "Content-Type: application/json" localhost:9090/kapis/monitoring.kubesphere.io/v1alpha3/clusterdashboard/test1/template