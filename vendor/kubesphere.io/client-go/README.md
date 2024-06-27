# KubeSphere client-go Project

The KubeSphere client-go Project is a rest-client of go libraries for communicating with the KubeSphere API Server.

# How to use it

1. Import client-go packages:
```golang
import (
	"kubesphere.io/client-go/rest"
	"kubesphere.io/client-go/client"
	"kubesphere.io/client-go/client/generic"
)
```
2. Create a generic client instance:
```golang
    var client client.Client
	config := &rest.Config{
		Host:     "127.0.0.1:9090",
		Username: "admin",
		Password: "P@88w0rd",
	}
	client = generic.NewForConfigOrDie(config, client.Options{Scheme: f.Scheme})
```
> generic.NewForConfigOrDie returns a client.Client that reads and writes from/to an KubeSphere API server. 

> It's only compatible with Kubernetes-like API objects.

3. KubeSphere API server provided a proxy to Kubernetes API Server. The client can read and write those Kubernetes native objects with the client directly.

```golang
	deploy := &appsv1.Deployment{}
	client.Get(context.TODO(), client.ObjectKey{Namespace: "kubesphere-system", Name: "ks-apiserver"}, deploy)
```

4. URLOptions and WorkspaceOptions can be provided to read and write Kubernetes likely Object that provided by KubeSphere API.
```golang
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ks-test",
			Labels: map[string]string{
				constants.WorkspaceLabelKey: "Workspace",
			},
		},
	}

	opts := &client.URLOptions{
		Group:   "tenant.kubesphere.io",
		Version: "v1alpha2",
	}

	err := f.GenericClient(f.BaseName).Create(context.TODO(), ns, opts, &client.WorkspaceOptions{Name: "Workspace"})
```

The KubeSphere API Architecture can be found at https://kubesphere.io/docs/reference/api-docs/

# Where does it come from?

client-go is synced from https://github.com/kubesphere/kubesphere/blob/master/staging/src/kubesphere.io/client-go. Code changes are made in that location, merged into `kubesphere.io/client-go` and later synced here.