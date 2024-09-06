/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha1

import (
	"bytes"
	"fmt"
	"net/url"

	"github.com/emicklei/go-restful/v3"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
)

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
		opts = append(opts, getter.WithInsecureSkipVerifyTLS(true))
		if repo.Spec.BasicAuth != nil {
			opts = append(opts, getter.WithBasicAuth(repo.Spec.BasicAuth.Username, repo.Spec.BasicAuth.Password))
		}
		chartGetter, err = getter.NewOCIGetter(opts...)
	case "http", "https":
		options := make([]getter.Option, 0)
		if chartURL.Scheme == "https" {
			options = append(options, getter.WithInsecureSkipVerifyTLS(true))
		}
		if repo.Spec.BasicAuth != nil {
			options = append(options, getter.WithBasicAuth(repo.Spec.BasicAuth.Username, repo.Spec.BasicAuth.Password))
		}
		chartGetter, err = getter.NewHTTPGetter(options...)
	}
	if err != nil {
		api.HandleInternalError(response, request, fmt.Errorf("failed to create chart getter: %v", err))
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
