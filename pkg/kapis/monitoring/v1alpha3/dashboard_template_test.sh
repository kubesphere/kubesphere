#! /bin/bash

curl -d '{"grafanaDashboardName":"test2","grafanaDashboardUrl":"https://grafana.com/api/dashboards/7362/revisions/5/download"}' -H "Content-Type: application/json" localhost:9090/kapis/monitoring.kubesphere.io/v1alpha3/dashboard/template