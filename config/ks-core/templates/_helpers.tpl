{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "ks-core.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "ks-core.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "ks-core.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "ks-core.labels" -}}
helm.sh/chart: {{ include "ks-core.chart" . }}
{{ include "ks-core.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "ks-core.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ks-core.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "ks-core.serviceAccountName" -}}
{{- default "kubesphere" (.Values.serviceAccount).name }}
{{- end }}

{{/*
Create the name of the secret of sa token.
*/}}
{{- define "ks-core.serviceAccountTokenName" -}}
{{-  printf "%s-%s" ( include "ks-core.serviceAccountName" . )   "sa-token" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "portal.url" -}}
{{- if and .Values.portal.https (.Values.portal.https).port }}
{{- if eq (int .Values.portal.https.port) 443 }}
{{- printf "https://%s" .Values.portal.hostname }}
{{- else }}
{{- printf "https://%s:%d" .Values.portal.hostname (int .Values.portal.https.port) }}
{{- end }}
{{- else }}
{{- if eq (int .Values.portal.http.port) 80 }}
{{- printf "http://%s" .Values.portal.hostname }}
{{- else }}
{{- printf "http://%s:%d" .Values.portal.hostname (int .Values.portal.http.port) }}
{{- end }}
{{- end }}
{{- end }}

{{- define "jwtSecret" -}}
{{- if eq .Values.authentication.issuer.jwtSecret "" }}
{{- with lookup "v1" "ConfigMap" (printf "%s" .Release.Namespace) "kubesphere-config" }}
{{- with (fromYaml (index .data "kubesphere.yaml")) }}
{{- if and .authentication (.authentication).jwtSecret }}
{{- .authentication.jwtSecret }}
{{- else if and .authentication (.authentication).issuer ((.authentication).issuer).jwtSecret }}
{{- .authentication.issuer.jwtSecret }}
{{- else }}
{{- $.Values.authentication.issuer.jwtSecret | default (randAlphaNum 32 ) }}
{{- end }}
{{- else }}
{{- $.Values.authentication.issuer.jwtSecret | default (randAlphaNum 32 ) }}
{{- end }}
{{- else }}
{{- $.Values.authentication.issuer.jwtSecret | default (randAlphaNum 32 ) }}
{{- end }}
{{- else }}
{{- .Values.authentication.issuer.jwtSecret }}
{{- end }}
{{- end }}

{{- define "role" -}}
{{- if eq .Values.role "" }}
{{- with lookup "v1" "ConfigMap" (printf "%s" .Release.Namespace) "kubesphere-config" }}
{{- with (fromYaml (index .data "kubesphere.yaml")) }}
{{- if and .multicluster (.multicluster).clusterRole }}
{{- if eq .multicluster.clusterRole "none" }}
{{- "host" }}
{{- else }}
{{- .multicluster.clusterRole }}
{{- end }}
{{- else }}
{{- $.Values.role | default "host" }}
{{- end }}
{{- else }}
{{- $.Values.role | default "host" }}
{{- end }}
{{- else }}
{{- $.Values.role | default "host" }}
{{- end }}
{{- else }}
{{- .Values.role }}
{{- end }}
{{- end }}

{{- define "hostClusterName" -}}
{{- if eq .Values.hostClusterName "" }}
{{- with lookup "v1" "ConfigMap" (printf "%s" .Release.Namespace) "kubesphere-config" }}
{{- with (fromYaml (index .data "kubesphere.yaml")) }}
{{- if and .multicluster (.multicluster).hostClusterName }}
{{- .multicluster.hostClusterName }}
{{- else }}
{{- $.Values.hostClusterName | default "host" }}
{{- end }}
{{- else }}
{{- $.Values.hostClusterName | default "host" }}
{{- end }}
{{- else }}
{{- $.Values.hostClusterName | default "host" }}
{{- end }}
{{- else }}
{{- .Values.hostClusterName }}
{{- end }}
{{- end }}

{{- define "validateHostClusterName" -}}
{{- $name := . -}}
{{- $pattern := "^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$" -}}
{{- if not (regexMatch $pattern $name) -}}
{{- fail (printf "Invalid hostClusterName '%s': a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character." $name) -}}
{{- else -}}
{{- $name -}}
{{- end -}}
{{- end }}

{{/*
Returns user's password or use default
*/}}
{{- define "getOrDefaultPass" }}
{{- if not .Values.authentication.adminPassword -}}
{{- printf "$2a$10$zcHepmzfKPoxCVCYZr5K7ORPZZ/ySe9p/7IUb/8u./xHrnSX2LOCO" -}}
{{- else -}}
{{- printf "%s" .Values.authentication.adminPassword -}}
{{- end -}}
{{- end }}

{{/*
Returns user's password or use default. Used by NOTES.txt
*/}}
{{- define "printOrDefaultPass" }}
{{- if not .Values.authentication.adminPassword -}}
{{- printf "P@88w0rd" -}}
{{- else -}}
{{- printf "%s" .Values.authentication.adminPassword -}}
{{- end -}}
{{- end }}

{{- define "getNodeAddress" -}}
{{- $address := "127.0.0.1"}}
{{- $found := false }}
{{- with $nodes := lookup "v1" "Node" "" "" }}
{{- range $nodeKey, $node := $nodes.items }}
{{- if (hasKey $node.metadata.labels "node-role.kubernetes.io/control-plane") }}
{{- range $k, $v := $node.status.addresses }}
{{- if and (eq $v.type "InternalIP") (not $found) }}
{{- $address = $v.address }}
{{- $found = true }}
{{- end }}
{{- end }}
{{- end }}
{{- end }}
{{- else }}
{{- end }}
{{- printf "%s" $address }}
{{- end }}
