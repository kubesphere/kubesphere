/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v2

import (
	"errors"
	"fmt"

	"github.com/emicklei/go-restful/v3"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
)

func (h *appHandler) exampleCr(req *restful.Request, resp *restful.Response) {
	name := req.PathParameter("name")
	crd := v1.CustomResourceDefinition{}
	err := h.client.Get(req.Request.Context(), client.ObjectKey{Name: name}, &crd)
	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}
	cr, err := convertCRDToCR(crd)
	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}
	resp.WriteEntity(cr)
}

func convertCRDToCR(crd v1.CustomResourceDefinition) (dstCr unstructured.Unstructured, err error) {

	cr := unstructured.Unstructured{}
	cr.SetName(fmt.Sprintf("%s-Instance", crd.Spec.Names.Singular))
	cr.SetGroupVersionKind(schema.GroupVersionKind{
		Group: crd.Spec.Group,
		Kind:  crd.Spec.Names.Kind,
	})

	var selectedVersion *v1.CustomResourceDefinitionVersion
	for _, version := range crd.Spec.Versions {
		if version.Served && version.Storage {
			selectedVersion = &version
			break
		}
	}
	if selectedVersion == nil {
		return dstCr, errors.New("no served and storage version found in CRD")
	}
	cr.SetAPIVersion(selectedVersion.Name)

	generateProps(selectedVersion, cr, "spec")
	generateProps(selectedVersion, cr, "status")

	return cr, nil
}

func generateProps(selectedVersion *v1.CustomResourceDefinitionVersion, cr unstructured.Unstructured, name string) {
	data := make(map[string]any)
	specProps := selectedVersion.Schema.OpenAPIV3Schema.Properties[name].Properties
	for key, value := range specProps {
		data[key] = getDefaultValue(value)
	}
	cr.Object[name] = data
}

func getDefaultValue(value v1.JSONSchemaProps) any {
	switch value.Type {
	case "object":
		return parseObject(value.Properties)
	case "integer":
		if value.Minimum != nil {
			return *value.Minimum
		}
		return 0
	case "boolean":
		return false
	case "array":
		if value.Items.Schema != nil {
			return []any{getDefaultValue(*value.Items.Schema)}
		}
		return []any{}
	case "string":
		if len(value.Enum) > 0 {
			return string(value.Enum[0].Raw)
		}
		return ""
	default:
		return nil
	}
}

func parseObject(obj map[string]v1.JSONSchemaProps) map[string]any {
	res := make(map[string]any)
	for key, value := range obj {
		res[key] = getDefaultValue(value)
	}
	return res
}
