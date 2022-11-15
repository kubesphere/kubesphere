/*
Copyright 2020 The Operator-SDK Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ActionConfigGetter interface {
	ActionConfigFor(obj client.Object) (*action.Configuration, error)
}

func NewActionConfigGetter(cfg *rest.Config, rm meta.RESTMapper, log logr.Logger) ActionConfigGetter {
	return &actionConfigGetter{
		cfg:        cfg,
		restMapper: rm,
		log:        log,
	}
}

var _ ActionConfigGetter = &actionConfigGetter{}

type actionConfigGetter struct {
	cfg        *rest.Config
	restMapper meta.RESTMapper
	log        logr.Logger
}

func (acg *actionConfigGetter) ActionConfigFor(obj client.Object) (*action.Configuration, error) {
	// Create a RESTClientGetter
	rcg := newRESTClientGetter(acg.cfg, acg.restMapper, obj.GetNamespace())

	// Setup the debug log function that Helm will use
	debugLog := func(format string, v ...interface{}) {
		if acg.log.GetSink() != nil {
			acg.log.V(1).Info(fmt.Sprintf(format, v...))
		}
	}

	// Create a client that helm will use to manage release resources.
	// The passed object is used as an owner reference on every
	// object the client creates.
	kc := kube.New(rcg)
	kc.Log = debugLog

	// Create the Kubernetes Secrets client. The passed object is
	// also used as an owner reference in the release secrets
	// created by this client.
	kcs, err := cmdutil.NewFactory(rcg).KubernetesClientSet()
	if err != nil {
		return nil, err
	}

	ownerRef := metav1.NewControllerRef(obj, obj.GetObjectKind().GroupVersionKind())
	d := driver.NewSecrets(&ownerRefSecretClient{
		SecretInterface: kcs.CoreV1().Secrets(obj.GetNamespace()),
		refs:            []metav1.OwnerReference{*ownerRef},
	})

	// Also, use the debug log for the storage driver
	d.Log = debugLog

	// Initialize the storage backend
	s := storage.Init(d)

	return &action.Configuration{
		RESTClientGetter: rcg,
		Releases:         s,
		KubeClient:       kc,
		Log:              debugLog,
	}, nil
}

var _ v1.SecretInterface = &ownerRefSecretClient{}

type ownerRefSecretClient struct {
	v1.SecretInterface
	refs []metav1.OwnerReference
}

func (c *ownerRefSecretClient) Create(ctx context.Context, in *corev1.Secret, opts metav1.CreateOptions) (*corev1.Secret, error) {
	in.OwnerReferences = append(in.OwnerReferences, c.refs...)
	return c.SecretInterface.Create(ctx, in, opts)
}

func (c *ownerRefSecretClient) Update(ctx context.Context, in *corev1.Secret, opts metav1.UpdateOptions) (*corev1.Secret, error) {
	in.OwnerReferences = append(in.OwnerReferences, c.refs...)
	return c.SecretInterface.Update(ctx, in, opts)
}
