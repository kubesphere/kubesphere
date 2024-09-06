/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package openapi

import (
	"github.com/go-openapi/spec"

	"kubesphere.io/kubesphere/pkg/api"
	ksVersion "kubesphere.io/kubesphere/pkg/version"
)

func EnrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "KubeSphere API",
			Description: "KubeSphere Enterprise OpenAPI",
			Version:     ksVersion.Get().GitVersion,
			Contact: &spec.ContactInfo{
				ContactInfoProps: spec.ContactInfoProps{
					Name:  "KubeSphere",
					URL:   "https://kubesphere.com.cn",
					Email: "support@kubesphere.cloud",
				},
			},
		},
	}
	// setup security definitions
	swo.SecurityDefinitions = map[string]*spec.SecurityScheme{
		"BearerToken": {SecuritySchemeProps: spec.SecuritySchemeProps{
			Type:        "apiKey",
			Name:        "Authorization",
			In:          "header",
			Description: "Bearer Token Authentication",
		}},
	}
	swo.Security = []map[string][]string{{"BearerToken": []string{}}}
	swo.Tags = []spec.Tag{

		{
			TagProps: spec.TagProps{
				Name: api.TagAuthentication,
			},
		},
		{
			TagProps: spec.TagProps{
				Name: api.TagMultiCluster,
			},
		},
		{
			TagProps: spec.TagProps{
				Name: api.TagIdentityManagement,
			},
		},
		{
			TagProps: spec.TagProps{
				Name: api.TagAccessManagement,
			},
		},
		{
			TagProps: spec.TagProps{
				Name: api.TagClusterResources,
			},
		},
		{
			TagProps: spec.TagProps{
				Name: api.TagNamespacedResources,
			},
		},
		{
			TagProps: spec.TagProps{
				Name: api.TagComponentStatus,
			},
		},
		{
			TagProps: spec.TagProps{
				Name: api.TagUserRelatedResources,
			},
		},
		{
			TagProps: spec.TagProps{
				Name: api.TagTerminal,
			},
		},
		{
			TagProps: spec.TagProps{
				Name: api.TagNonResourceAPI,
			},
		},
	}
}
