{{/*
Return the proper image name
*/}}

{{- define "common.images.image" -}}
{{- $registryName := .global.imageRegistry -}}
{{- $repositoryName := .imageRoot.repository -}}
{{- $separator := ":" -}}
{{- $termination := .global.tag | toString -}}
{{- if .imageRoot.registry }}
    {{- $registryName = .imageRoot.registry -}}
{{- end -}}
{{- if .imageRoot.tag }}
    {{- $termination = .imageRoot.tag | toString -}}
{{- end -}}
{{- if .imageRoot.digest }}
    {{- $separator = "@" -}}
    {{- $termination = .imageRoot.digest | toString -}}
{{- end -}}
{{- printf "%s/%s%s%s" $registryName $repositoryName $separator $termination -}}
{{- end -}}


{{- define "image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.image "global" .Values.global) }}
{{- end -}}

{{- define "haproxy.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.haproxy.image "global" .Values.global) }}
{{- end -}}

{{- define "sysctlImage" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.sysctlImage "global" .Values.global) }}
{{- end -}}

{{- define "exporter.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.exporter.image "global" .Values.global) }}
{{- end -}}
