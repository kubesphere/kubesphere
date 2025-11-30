# Istio ConfigMap Grayscale Release

## Overview

This feature implements Istio grayscale release support for ConfigMap resources in KubeSphere, addressing GitHub Issue #3904. It enables gradual traffic shifting and canary deployments for configuration changes with minimal risk and maximum control.

## Features

### Traffic Splitting Strategies

1. **Weighted Splitting**: Distribute traffic based on percentage (0-100%)
2. **Header-Based Routing**: Route traffic based on HTTP headers for targeted users
3. **Cookie-Based Routing**: Route traffic based on cookies for A/B testing

### Rollback Capabilities

1. **Immediate Rollback**: Instant rollback to original configuration
2. **Gradual Rollback**: Step-by-step rollback with configurable delays

### Version Management

1. **Automatic Versioning**: Generate version hashes for ConfigMaps
2. **Status Tracking**: Monitor rollout progress and current state
3. **Validation**: Ensure ConfigMaps exist before rollout

## Implementation

### Core Components

- `ConfigMapGrayscaleManager`: Main manager for grayscale releases
- `ConfigMapGrayscaleSpec`: Configuration specification
- `ConfigMapGrayscaleStatus`: Status and state tracking
- `GrayscaleStrategy`: Traffic splitting strategy definition

### API Types

```go
type ConfigMapGrayscaleSpec struct {
    ServiceName       string            `json:"serviceName"`
    OriginalConfigMap string            `json:"originalConfigMap"`
    CanaryConfigMap   string            `json:"canaryConfigMap"`
    CanaryWeight      int32             `json:"canaryWeight"`
    Strategy          GrayscaleStrategy `json:"strategy"`
    Namespace         string            `json:"namespace"`
}

type GrayscaleStrategy struct {
    Type         string   `json:"type"`
    HeaderName   string   `json:"headerName,omitempty"`
    HeaderValues []string `json:"headerValues,omitempty"`
    CookieName   string   `json:"cookieName,omitempty"`
    CookieValue  string   `json:"cookieValue,omitempty"`
}
```

## Usage Examples

### Weighted Traffic Splitting

```go
spec := &ConfigMapGrayscaleSpec{
    ServiceName:       "my-app-service",
    OriginalConfigMap: "app-config-v1",
    CanaryConfigMap:   "app-config-v2",
    CanaryWeight:      20,
    Strategy: GrayscaleStrategy{
        Type: "weighted",
    },
    Namespace: "default",
}

manager := NewConfigMapGrayscaleManager(k8sClient, istioClient)
status, err := manager.CreateGrayscaleRelease(context.Background(), spec)
```

### Header-Based Routing

```go
spec := &ConfigMapGrayscaleSpec{
    ServiceName:       "my-app-service",
    OriginalConfigMap: "app-config-v1",
    CanaryConfigMap:   "app-config-v2",
    Strategy: GrayscaleStrategy{
        Type:        "header",
        HeaderName:  "x-feature-flags",
        HeaderValues: []string{"beta", "test"},
    },
    Namespace: "default",
}
```

### Cookie-Based Routing

```go
spec := &ConfigMapGrayscaleSpec{
    ServiceName:       "my-app-service",
    OriginalConfigMap: "app-config-v1",
    CanaryConfigMap:   "app-config-v2",
    Strategy: GrayscaleStrategy{
        Type:       "cookie",
        CookieName: "user_segment",
        CookieValue: "beta_testers",
    },
    Namespace: "default",
}
```

## Operations

### Create Grayscale Release

```go
status, err := manager.CreateGrayscaleRelease(ctx, spec)
if err != nil {
    log.Fatalf("Failed to create grayscale release: %v", err)
}
```

### Update Traffic Weight

```go
err := manager.UpdateTrafficWeight(ctx, "my-grayscale", 50)
if err != nil {
    log.Fatalf("Failed to update traffic weight: %v", err)
}
```

### Rollback

```go
// Immediate rollback
err := manager.Rollback(ctx, "my-grayscale", "immediate")

// Gradual rollback
err := manager.Rollback(ctx, "my-grayscale", "gradual")
```

## VirtualService Configuration

The feature automatically generates Istio VirtualService configurations based on the selected strategy:

### Weighted VirtualService

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: my-service-grayscale
spec:
  hosts:
  - my-service
  http:
  - route:
    - destination:
        host: my-service-v1.0.1234567890
        subset: original
      weight: 80
    - destination:
        host: my-service-v1.0.1234567891
        subset: canary
      weight: 20
```

### Header-Based VirtualService

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
spec:
  hosts:
  - my-service
  http:
  - match:
    - headers:
        x-feature-flags:
          exact: beta
    route:
    - destination:
        host: my-service-v1.0.1234567891
        subset: canary
      weight: 100
  - route:
    - destination:
        host: my-service-v1.0.1234567890
        subset: original
      weight: 100
```

## Testing

### Unit Tests

The implementation includes comprehensive unit tests covering:

- Grayscale release creation for all strategies
- Traffic weight updates
- Rollback operations
- VirtualService generation
- Error handling and validation

### Running Tests

```bash
# Run all tests
go test ./pkg/apiserver/istio/...

# Run specific test
go test ./pkg/apiserver/istio/ -run TestConfigMapGrayscaleManager_CreateGrayscaleRelease

# Run with coverage
go test -cover ./pkg/apiserver/istio/...
```

### Test Coverage

- Weighted strategy: ✅
- Header strategy: ✅
- Cookie strategy: ✅
- Traffic weight updates: ✅
- Rollback operations: ✅
- Error validation: ✅

## Integration with KubeSphere

### Console Integration

The feature integrates with KubeSphere console to provide:

- Visual traffic distribution dashboard
- Configuration editor with validation
- Real-time metrics and monitoring
- One-click rollback functionality
- Version history and comparison

### API Integration

- RESTful API endpoints for CRUD operations
- WebSocket support for real-time updates
- Authentication and authorization
- Audit logging

## Security Considerations

### RBAC Permissions

Required permissions for ConfigMap grayscale operations:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: configmap-grayscale-operator
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "list", "create", "update", "patch", "delete"]
- apiGroups: ["networking.istio.io"]
  resources: ["virtualservices"]
  verbs: ["get", "list", "create", "update", "patch", "delete"]
```

### Best Practices

1. **ConfigMap Validation**: Always validate ConfigMap syntax before rollout
2. **Gradual Rollout**: Start with low canary weight (5-10%)
3. **Monitoring**: Set up comprehensive monitoring and alerting
4. **Rollback Planning**: Always have rollback strategy defined
5. **Version Management**: Clean up old versions regularly

## Performance Considerations

### VirtualService Updates

- Updates are applied immediately to existing connections
- Consider connection draining during updates
- Monitor proxy reload times

### Resource Usage

- Versioned ConfigMaps consume additional storage
- Implement cleanup policies for old versions
- Monitor storage usage in large deployments

## Troubleshooting

### Common Issues

1. **VirtualService Not Created**
   - Check Istio permissions
   - Verify VirtualService CRD exists
   - Check service name spelling

2. **ConfigMap Not Found**
   - Verify ConfigMap exists in correct namespace
   - Check RBAC permissions
   - Validate ConfigMap names

3. **Traffic Weight Not Applied**
   - Check VirtualService configuration
   - Verify Istio proxy configuration
   - Restart affected pods

### Debug Commands

```bash
# Check VirtualService status
kubectl get virtualservice -l managed-by=kubesphere-grayscale

# Check ConfigMap versions
kubectl get configmaps -l original-configmap=my-config --show-labels

# Check Istio proxy configuration
kubectl exec -it <pod-name> -c istio-proxy -- pilot-agent request GET config_dump
```

## Migration Guide

### From Manual ConfigMap Updates

1. Create versioned ConfigMaps
2. Set up grayscale release configuration
3. Gradually migrate traffic
4. Clean up old configurations

### From Other Grayscale Tools

1. Export existing routing rules
2. Convert to ConfigMapGrayscale format
3. Apply new configuration
4. Remove old resources

## Future Enhancements

### Planned Features

1. **Progressive Rollout**: Step-by-step rollout with conditions
2. **Metrics Integration**: Enhanced monitoring and alerting
3. **Multi-Service Support**: Apply same ConfigMap to multiple services
4. **Automatic Rollback**: Error threshold-based automatic rollback
5. **Performance Optimization**: Reduced VirtualService update overhead

### API Extensions

1. **Rollback Policies**: Configurable rollback behavior
2. **Progressive Steps**: Multi-step rollout configuration
3. **Metrics Collection**: Performance metrics integration
4. **Event Notifications**: Webhook support for status changes

## Contributing

### Development Setup

1. Install Go 1.19+
2. Install Istio client-go libraries
3. Set up Kubernetes cluster with Istio
4. Install KubeSphere development environment

### Code Standards

- Follow Go conventions and best practices
- Write comprehensive unit tests
- Update documentation for API changes
- Ensure backward compatibility

### Testing Requirements

- Unit tests for all new features
- Integration tests for API endpoints
- Performance tests for high-load scenarios
- Security tests for permission validation

## Support

For issues and questions:

1. Check this documentation
2. Review test examples
3. Search existing GitHub issues
4. Create new issue with detailed description
5. Join KubeSphere community discussions

## Changelog

### v1.0.0
- Initial ConfigMap grayscale release support
- Weighted, header-based, and cookie-based routing
- Basic rollback functionality
- Comprehensive test suite
- Documentation and examples

### v1.1.0 (Planned)
- Progressive rollout support
- Enhanced metrics and monitoring
- Multi-service support
- Advanced rollback strategies
