## Extension Values

These are configurable values in the InstallPlan's `spec.config` for the Gateway extension.

To check the actual latest default versions from the installed extension:
```bash
kubectl get configmap -n kubesphere-system -l kubesphere.io/extension-ref=gateway -o json | python3 -c "import json,sys; [print(c['metadata']['name']) for c in json.load(sys.stdin)['items']]"
```

| Name | Meaning | Default | Range |
| --- | --- | --- | --- |
| `config.gateway.namespace` | Namespace where gateway components are deployed | `kubesphere-controls-system` | |
| `config.gateway.exposeNodeLabelKey` | Node label key for gateway pod scheduling | `node-role.kubernetes.io/control-plane` | |
| `config.gateway.versionConstraint` | Allowed ingress-nginx chart version range | — | (e.g. `>= 4.3.0, <= 4.13.9`) |
| `config.gateway.logSearchEndpoint` | Endpoint for log search proxy | `http://whizard-telemetry-apiserver.extension-whizard-telemetry.svc:9090` | |
| `config.gateway.valuesOverride.controller.image.registry` | Container image registry for ingress-nginx | `docker.io` | |
| `config.gateway.valuesOverride.controller.image.tag` | ingress-nginx controller image tag | `v1.13.9` | |
| `config.upgradeTool.image.repository` | Upgrade tool image repository | `kubesphere/gateway-upgrade-tool` | |
| `config.upgradeTool.image.tag` | Upgrade tool image tag | `v1.1.6` | |
| `config.upgradeTool.backup.enable` | Enable automatic backup before upgrade | `true` | `true` / `false` |
| `config.upgradeTool.backup.persistentVolumeClaim.storageCapacity` | Backup PVC size | `1Gi` | |
| `config.helmExecutor.image.registry` | Helm executor image registry | `registry.cn-beijing.aliyuncs.com` | |
| `config.helmExecutor.image.tag` | Helm executor image tag | `v1.33.1` | |
| `config.helmExecutor.timeout` | Helm operation timeout | `10m` | |
| `config.helmExecutor.resources.limits.cpu` | Helm executor CPU limit | `500m` | |
| `config.helmExecutor.resources.limits.memory` | Helm executor memory limit | `500Mi` | |

### Example: Customize node scheduling and resources

```yaml
apiVersion: kubesphere.io/v1alpha1
kind: InstallPlan
metadata:
  name: gateway
spec:
  extension:
    name: gateway
    version: "1.1.7"
  enabled: true
  upgradeStrategy: Manual
  clusterScheduling:
    placement:
      clusters:
        - "host"
  config: |
    config:
      gateway:
        namespace: kubesphere-controls-system
        exposeNodeLabelKey: "node-role.kubernetes.io/ingress"
        valuesOverride:
          controller:
            image:
              registry: docker.io
              tag: v1.13.9
            nodeSelector:
              kubernetes.io/os: linux
            tolerations:
              - key: "node-role.kubernetes.io/ingress"
                operator: "Exists"
                effect: "NoSchedule"
            resources:
              requests:
                cpu: 500m
                memory: 512Mi
              limits:
                cpu: 2
                memory: 2Gi
      upgradeTool:
        backup:
          enable: "true"
          persistentVolumeClaim:
            storageCapacity: "5Gi"
      helmExecutor:
        timeout: 15m
        resources:
          limits:
            cpu: "1"
            memory: 1Gi
```

### Example: Minimal override (image registry only)

```yaml
  config: |
    config:
      gateway:
        valuesOverride:
          controller:
            image:
              registry: registry.cn-beijing.aliyuncs.com
```

### Example: Custom log search endpoint

```yaml
  config: |
    config:
      gateway:
        logSearchEndpoint: "http://my-custom-logstack.ns.svc:9090"
```
