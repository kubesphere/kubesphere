{{/*
Return the proper image name
*/}}
{{- define "ks-apiserver.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.apiserver.image "global" .Values.global) }}
{{- end -}}

{{- define "ks-console.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.console.image "global" .Values.global) }}
{{- end -}}

{{- define "ks-controller-manager.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.controller.image "global" .Values.global) }}
{{- end -}}

{{- define "kubectl.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.kubectl.image "global" .Values.global) }}
{{- end -}}

{{- define "nodeShell.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.nodeShell.image "global" .Values.global) }}
{{- end -}}

{{- define "helm.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.helmExecutor.image "global" .Values.global) }}
{{- end -}}

{{- define "upgrade.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.upgrade.image "global" .Values.global) }}
{{- end -}}

{{- define "redis.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.redis.image "global" .Values.global) }}
{{- end -}}

{{- define "extensions_museum.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.ksExtensionRepository.image "global" .Values.global) }}
{{- end -}}

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

{{/*
Return the proper Docker Image Registry Secret Names
*/}}
{{- define "apiserver.imagePullSecrets" -}}
{{- include "common.images.pullSecrets" (dict "images" (list .Values.apiserver.image) "global" .Values.global) -}}
{{- end -}}

{{- define "console.imagePullSecrets" -}}
{{- include "common.images.pullSecrets" (dict "images" (list .Values.console.image) "global" .Values.global) -}}
{{- end -}}

{{- define "controller.imagePullSecrets" -}}
{{- include "common.images.pullSecrets" (dict "images" (list .Values.controller.image) "global" .Values.global) -}}
{{- end -}}

{{- define "extensions_museum.imagePullSecrets" -}}
{{- include "common.images.pullSecrets" (dict "images" (list .Values.ksExtensionRepository.image) "global" .Values.global) -}}
{{- end -}}

{{- define "common.images.pullSecrets" -}}
  {{- $pullSecrets := list }}

  {{- if .global }}
    {{- range .global.imagePullSecrets -}}
      {{- $pullSecrets = append $pullSecrets . -}}
    {{- end -}}
  {{- end -}}

  {{- range .images -}}
    {{- range .pullSecrets -}}
      {{- $pullSecrets = append $pullSecrets . -}}
    {{- end -}}
  {{- end -}}

  {{- if (not (empty $pullSecrets)) }}
imagePullSecrets:
    {{- range $pullSecrets }}
  - name: {{ . }}
    {{- end }}
  {{- end }}
{{- end -}}
