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
