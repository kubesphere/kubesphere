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
{{- if .Values.serviceAccount.create }}
{{- default (include "ks-core.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Returns user's password or use default
*/}}
{{- define "getOrDefaultPass" }}
{{- if not .Values.adminPassword -}}
{{- printf "$2a$10$zcHepmzfKPoxCVCYZr5K7ORPZZ/ySe9p/7IUb/8u./xHrnSX2LOCO" -}}
{{- else -}}
{{- printf "%s" .Values.adminPassword -}}
{{- end -}}
{{- end }}

{{/*
Returns user's password or use default. Used by NOTES.txt
*/}}
{{- define "printOrDefaultPass" }}
{{- if not .Values.adminPassword -}}
{{- printf "P@88w0rd" -}}
{{- else -}}
{{- printf "%s" .Values.adminPassword -}}
{{- end -}}
{{- end }}

{{- define "getNodeAddress" -}}
{{- $address := "127.0.0.1"}}
{{- range $index, $node := (lookup "v1" "Node" "" "").items -}}
  {{- range $k, $v := $node.status.addresses }}
    {{- if (eq $v.type "InternalIP") }}
      {{- $address = $v.address }}
      {{- break }}
    {{- end }}
  {{- end }}
  {{- if (ne $address "127.0.0.1") }}
    {{- break }}
  {{- end }}
{{- end }}
{{- printf "%s" $address }}
{{- end }}