/*
Copyright 2020 The KubeSphere Authors.

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

package cluster

import (
	"context"
	"fmt"
	"net/http"

	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
)

type ValidatingHandler struct {
	Client  client.Client
	decoder *admission.Decoder
}

var _ admission.DecoderInjector = &ValidatingHandler{}

// InjectDecoder injects the decoder into a ValidatingHandler.
func (h *ValidatingHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

// Handle handles admission requests.
func (h *ValidatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	if req.Operation != v1.Update {
		return admission.Allowed("")
	}

	newCluster := &clusterv1alpha1.Cluster{}
	if err := h.decoder.DecodeRaw(req.Object, newCluster); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	oldCluster := &clusterv1alpha1.Cluster{}
	if err := h.decoder.DecodeRaw(req.OldObject, oldCluster); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// The cluster created for the first time has no status information
	if oldCluster.Status.UID == "" {
		return admission.Allowed("")
	}

	clusterConfig, err := clientcmd.RESTConfigFromKubeConfig(newCluster.Spec.Connection.KubeConfig)
	if err != nil {
		return admission.Denied(fmt.Sprintf("failed to load cluster config for %s: %s", newCluster.Name, err))
	}
	clusterClient, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return admission.Denied(err.Error())
	}
	kubeSystem, err := clusterClient.CoreV1().Namespaces().Get(ctx, metav1.NamespaceSystem, metav1.GetOptions{})
	if err != nil {
		return admission.Denied(err.Error())
	}

	if oldCluster.Status.UID != kubeSystem.UID {
		return admission.Denied("this kubeconfig corresponds to a different cluster than the previous one, you need to make sure that kubeconfig is not from another cluster")
	}
	return admission.Allowed("")
}
