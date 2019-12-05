# Connect to remote service with telepresence

Telepresence is an open source tool that lets you run a single service locally, while connecting that service to a remote Kubernetes cluster.

We can use telepresence to help us running KubeSphere apiserver locally.


## 1. Install telepresence

You can read the [official installation documentation](https://www.telepresence.io/reference/install.html) to install telepresence.

## 2. Run telepresence

Open your command line and run the command `telepresence`, telepresence will help you to enter a bash connected to a remote Kubernetes cluster.

Test telepresence with KubeSphere apigateway:

```bash

@kubernetes-admin@cluster.local|bash-3.2$ curl http://ks-apigateway.kubesphere-system

401 Unauthorized
```

## 3. Run your module in bash

Now your module can easily connect to remote services.

```bash
go run cmd/ks-apiserver/apiserver.go
```
## 4. Further more

You can use telepresence to replace services in the cluster for debugging. For more information, please refer to the [official documentation](https://www.telepresence.io/discussion/overview).
