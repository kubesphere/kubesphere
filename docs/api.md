// ...existing code...

## Service Proxy API

### GET /kapis/proxy/v1alpha1/namespaces/{namespace}/services/{service}/{path}

Proxies a request to a Kubernetes service through ks-apiserver.

**Parameters**:
- `namespace`: The namespace of the service (required).
- `service`: The name of the service (required).
- `path`: The path to access on the service (optional).

**Example**:
```bash
curl -H "Authorization: Bearer $TOKEN" http://ks-apiserver:80/kapis/proxy/v1alpha1/namespaces/default/services/my-service/
```

**Response**:
- Forwards the response from the service.

// ...existing code...

