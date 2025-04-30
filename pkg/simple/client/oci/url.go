/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package oci

import (
	"fmt"

	"oras.land/oras-go/pkg/registry"
)

func buildScheme(plainHTTP bool) string {
	if plainHTTP {
		return "http"
	}
	return "https"
}

func buildRegistryBaseURL(plainHTTP bool, ref registry.Reference) string {
	return fmt.Sprintf("%s://%s/v2/", buildScheme(plainHTTP), ref.Host())
}

func buildRegistryCatalogURL(plainHTTP bool, ref registry.Reference) string {
	return fmt.Sprintf("%s://%s/v2/_catalog", buildScheme(plainHTTP), ref.Host())
}
