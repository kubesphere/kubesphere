{{/* vim: set filetype=mustache: */}}

{{- define "ingress-controller.yaml" }}
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: ks-router
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: kubesphere
          component: ks-router
          tier: backend
      template:
        metadata:
          labels:
            app: kubesphere
            component: ks-router
            tier: backend
          annotations:
            prometheus.io/port: '10254'
            prometheus.io/scrape: 'true'
        spec:
          serviceAccountName: kubesphere-router-serviceaccount
          containers:
            - name: nginx-ingress-controller
              image: {{ .Values.image.nginx_ingress_controller_repo }}:{{ .Values.image.nginx_ingress_controller_tag | default .Chart.AppVersion}}
              args:
                - /nginx-ingress-controller
                - --default-backend-service=$(POD_NAMESPACE)/default-http-backend
                - --annotations-prefix=nginx.ingress.kubernetes.io
                - --update-status
                - --update-status-on-shutdown
              env:
                - name: POD_NAME
                  valueFrom:
                    fieldRef:
                      fieldPath: metadata.name
                - name: POD_NAMESPACE
                  valueFrom:
                    fieldRef:
                      fieldPath: metadata.namespace
              ports:
              - name: http
                containerPort: 80
              - name: https
                containerPort: 443
              livenessProbe:
                failureThreshold: 3
                httpGet:
                  path: /healthz
                  port: 10254
                  scheme: HTTP
                initialDelaySeconds: 10
                periodSeconds: 10
                successThreshold: 1
                timeoutSeconds: 1
              readinessProbe:
                failureThreshold: 3
                httpGet:
                  path: /healthz
                  port: 10254
                  scheme: HTTP
                periodSeconds: 10
                successThreshold: 1
                timeoutSeconds: 1
              securityContext:
                runAsNonRoot: false
{{- end }}

{{- define "ingress-controller-svc.yaml" }}
    apiVersion: v1
    kind: Service
    metadata:
      name: kubesphere-router-gateway
      labels:
        app: kubesphere
        component: ks-router
        tier: backend
    spec:
      selector:
        app: kubesphere
        component: ks-router
        tier: backend
      type: LoadBalancer
      ports:
        - name: http
          protocol: TCP
          port: 80
          targetPort: 80
        - name: https
          protocol: TCP
          port: 443
          targetPort: 443
{{- end }}