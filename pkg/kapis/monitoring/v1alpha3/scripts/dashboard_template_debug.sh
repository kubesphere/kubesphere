#! /bin/bash
# Copyright 2022 The KubeSphere Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

curl -d '{"grafanaDashboardUrl":"https://grafana.com/api/dashboards/7362/revisions/5/download", "description":"this is a test clusterdashboard."}' -H "Content-Type: application/json" localhost:9090/kapis/monitoring.kubesphere.io/v1alpha3/clusterdashboards/test1/template
curl -d '{"grafanaDashboardUrl":"https://grafana.com/api/dashboards/7362/revisions/5/download", "description":"this is a test dashboard."}' -H "Content-Type: application/json" localhost:9090/kapis/monitoring.kubesphere.io/v1alpha3/namespaces/default/dashboards/test2/template