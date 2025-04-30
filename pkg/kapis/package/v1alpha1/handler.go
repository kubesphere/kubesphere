/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha1

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"

	"github.com/emicklei/go-restful/v3"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
	"kubesphere.io/utils/helm"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
)

var caTemplate = "{{ .TempDIR }}/repository/{{ .RepositoryName }}/ssl/ca.crt"

type handler struct {
	cache runtimeclient.Reader
}

func (h *handler) ListFiles(request *restful.Request, response *restful.Response) {
	extensionVersion := corev1alpha1.ExtensionVersion{}
	if err := h.cache.Get(request.Request.Context(), types.NamespacedName{Name: request.PathParameter("version")}, &extensionVersion); err != nil {
		api.HandleError(response, request, err)
		return
	}

	if extensionVersion.Spec.ChartDataRef != nil {
		configMap := &corev1.ConfigMap{}
		if err := h.cache.Get(request.Request.Context(), types.NamespacedName{Namespace: extensionVersion.Spec.ChartDataRef.Namespace, Name: extensionVersion.Spec.ChartDataRef.Name}, configMap); err != nil {
			api.HandleInternalError(response, request, err)
			return
		}
		data := configMap.BinaryData[extensionVersion.Spec.ChartDataRef.Key]
		if data == nil {
			response.WriteEntity([]interface{}{})
			return
		}
		files, err := loader.LoadArchiveFiles(bytes.NewReader(data))
		if err != nil {
			api.HandleInternalError(response, request, err)
			return
		}
		response.WriteEntity(files)
		return
	}

	chartURL, err := url.Parse(extensionVersion.Spec.ChartURL)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	repo := &corev1alpha1.Repository{}
	if extensionVersion.Spec.Repository != "" {
		if err := h.cache.Get(request.Request.Context(), types.NamespacedName{Name: extensionVersion.Spec.Repository}, repo); err != nil {
			api.HandleInternalError(response, request, err)
			return
		}
	}

	var chartGetter getter.Getter
	switch chartURL.Scheme {
	case registry.OCIScheme:
		opts := make([]getter.Option, 0)
		if extensionVersion.Spec.Repository != "" {
			opts = append(opts, getter.WithInsecureSkipVerifyTLS(repo.Spec.Insecure))
		}
		if repo.Spec.BasicAuth != nil {
			opts = append(opts, getter.WithBasicAuth(repo.Spec.BasicAuth.Username, repo.Spec.BasicAuth.Password))
		}
		chartGetter, err = getter.NewOCIGetter(opts...)
		if err != nil {
			api.HandleInternalError(response, request, fmt.Errorf("failed to create chart getter: %v", err))
			return
		}
	case "http", "https":
		opts := make([]getter.Option, 0)
		if chartURL.Scheme == "https" && extensionVersion.Spec.Repository != "" {
			opts = append(opts, getter.WithInsecureSkipVerifyTLS(repo.Spec.Insecure))
		}
		if repo.Spec.CABundle != "" {
			tlsConfig, err := helm.NewTLSConfig(repo.Spec.CABundle, repo.Spec.Insecure)
			if err != nil {
				api.HandleInternalError(response, request, err)
				return
			}
			opts = append(opts, getter.WithTransport(&http.Transport{TLSClientConfig: tlsConfig}))
		}
		if repo.Spec.BasicAuth != nil {
			opts = append(opts, getter.WithBasicAuth(repo.Spec.BasicAuth.Username, repo.Spec.BasicAuth.Password))
		}
		chartGetter, err = getter.NewHTTPGetter(opts...)
		if err != nil {
			api.HandleInternalError(response, request, fmt.Errorf("failed to create chart getter: %v", err))
			return
		}
	default:
		api.HandleInternalError(response, request, fmt.Errorf("cannot support chartURL %s, it's schame should be: oci,http,https", extensionVersion.Spec.ChartURL))
		return
	}

	data, err := chartGetter.Get(extensionVersion.Spec.ChartURL)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	files, err := loader.LoadArchiveFiles(data)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	_ = response.WriteEntity(files)
}
