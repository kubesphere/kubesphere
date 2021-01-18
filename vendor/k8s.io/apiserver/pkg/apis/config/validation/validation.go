/*
Copyright 2019 The Kubernetes Authors.

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

// Package validation validates EncryptionConfiguration.
package validation

import (
	"encoding/base64"
	"fmt"
	"net/url"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/apis/config"
)

const (
	moreThanOneElementErr    = "more than one provider specified in a single element, should split into different list elements"
	keyLenErrFmt             = "secret is not of the expected length, got %d, expected one of %v"
	unsupportedSchemeErrFmt  = "unsupported scheme %q for KMS provider, only unix is supported"
	atLeastOneRequiredErrFmt = "at least one %s is required"
	mandatoryFieldErrFmt     = "%s is a mandatory field for a %s"
	base64EncodingErr        = "secrets must be base64 encoded"
	zeroOrNegativeErrFmt     = "%s should be a positive value"
	nonZeroErrFmt            = "%s should be a positive value, or negative to disable"
	encryptionConfigNilErr   = "EncryptionConfiguration can't be nil"
)

var (
	aesKeySizes = []int{16, 24, 32}
	// See https://golang.org/pkg/crypto/aes/#NewCipher for details on supported key sizes for AES.
	secretBoxKeySizes = []int{32}
	// See https://godoc.org/golang.org/x/crypto/nacl/secretbox#Open for details on the supported key sizes for Secretbox.
	root = field.NewPath("resources")
)

// ValidateEncryptionConfiguration validates a v1.EncryptionConfiguration.
func ValidateEncryptionConfiguration(c *config.EncryptionConfiguration) field.ErrorList {
	allErrs := field.ErrorList{}

	if c == nil {
		allErrs = append(allErrs, field.Required(root, "EncryptionConfiguration can't be nil"))
		return allErrs
	}

	if len(c.Resources) == 0 {
		allErrs = append(allErrs, field.Required(root, fmt.Sprintf(atLeastOneRequiredErrFmt, root)))
		return allErrs
	}

	for i, conf := range c.Resources {
		r := root.Index(i).Child("resources")
		p := root.Index(i).Child("providers")

		if len(conf.Resources) == 0 {
			allErrs = append(allErrs, field.Required(r, fmt.Sprintf(atLeastOneRequiredErrFmt, r)))
		}

		if len(conf.Providers) == 0 {
			allErrs = append(allErrs, field.Required(p, fmt.Sprintf(atLeastOneRequiredErrFmt, p)))
		}

		for j, provider := range conf.Providers {
			path := p.Index(j)
			allErrs = append(allErrs, validateSingleProvider(provider, path)...)

			switch {
			case provider.KMS != nil:
				allErrs = append(allErrs, validateKMSConfiguration(provider.KMS, path.Child("kms"))...)
			case provider.AESGCM != nil:
				allErrs = append(allErrs, validateKeys(provider.AESGCM.Keys, path.Child("aesgcm").Child("keys"), aesKeySizes)...)
			case provider.AESCBC != nil:
				allErrs = append(allErrs, validateKeys(provider.AESCBC.Keys, path.Child("aescbc").Child("keys"), aesKeySizes)...)
			case provider.Secretbox != nil:
				allErrs = append(allErrs, validateKeys(provider.Secretbox.Keys, path.Child("secretbox").Child("keys"), secretBoxKeySizes)...)
			}
		}
	}

	return allErrs
}

func validateSingleProvider(provider config.ProviderConfiguration, filedPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	found := 0

	if provider.KMS != nil {
		found++
	}
	if provider.AESGCM != nil {
		found++
	}
	if provider.AESCBC != nil {
		found++
	}
	if provider.Secretbox != nil {
		found++
	}
	if provider.Identity != nil {
		found++
	}

	if found == 0 {
		return append(allErrs, field.Invalid(filedPath, provider, "provider does not contain any of the expected providers: KMS, AESGCM, AESCBC, Secretbox, Identity"))
	}

	if found > 1 {
		return append(allErrs, field.Invalid(filedPath, provider, moreThanOneElementErr))
	}

	return allErrs
}

func validateKeys(keys []config.Key, fieldPath *field.Path, expectedLen []int) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(keys) == 0 {
		allErrs = append(allErrs, field.Required(fieldPath, fmt.Sprintf(atLeastOneRequiredErrFmt, "keys")))
		return allErrs
	}

	for i, key := range keys {
		allErrs = append(allErrs, validateKey(key, fieldPath.Index(i), expectedLen)...)
	}

	return allErrs
}

func validateKey(key config.Key, fieldPath *field.Path, expectedLen []int) field.ErrorList {
	allErrs := field.ErrorList{}

	if key.Name == "" {
		allErrs = append(allErrs, field.Required(fieldPath.Child("name"), fmt.Sprintf(mandatoryFieldErrFmt, "name", "key")))
	}

	if key.Secret == "" {
		allErrs = append(allErrs, field.Required(fieldPath.Child("secret"), fmt.Sprintf(mandatoryFieldErrFmt, "secret", "key")))
		return allErrs
	}

	secret, err := base64.StdEncoding.DecodeString(key.Secret)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(fieldPath.Child("secret"), "REDACTED", base64EncodingErr))
		return allErrs
	}

	lenMatched := false
	for _, l := range expectedLen {
		if len(secret) == l {
			lenMatched = true
			break
		}
	}

	if !lenMatched {
		allErrs = append(allErrs, field.Invalid(fieldPath.Child("secret"), "REDACTED", fmt.Sprintf(keyLenErrFmt, len(secret), expectedLen)))
	}

	return allErrs
}

func validateKMSConfiguration(c *config.KMSConfiguration, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if c.Name == "" {
		allErrs = append(allErrs, field.Required(fieldPath.Child("name"), fmt.Sprintf(mandatoryFieldErrFmt, "name", "provider")))
	}
	allErrs = append(allErrs, validateKMSTimeout(c, fieldPath.Child("timeout"))...)
	allErrs = append(allErrs, validateKMSEndpoint(c, fieldPath.Child("endpoint"))...)
	allErrs = append(allErrs, validateKMSCacheSize(c, fieldPath.Child("cachesize"))...)
	return allErrs
}

func validateKMSCacheSize(c *config.KMSConfiguration, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if *c.CacheSize == 0 {
		allErrs = append(allErrs, field.Invalid(fieldPath, *c.CacheSize, fmt.Sprintf(nonZeroErrFmt, "cachesize")))
	}

	return allErrs
}

func validateKMSTimeout(c *config.KMSConfiguration, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if c.Timeout.Duration <= 0 {
		allErrs = append(allErrs, field.Invalid(fieldPath, c.Timeout, fmt.Sprintf(zeroOrNegativeErrFmt, "timeout")))
	}

	return allErrs
}

func validateKMSEndpoint(c *config.KMSConfiguration, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if len(c.Endpoint) == 0 {
		return append(allErrs, field.Invalid(fieldPath, "", fmt.Sprintf(mandatoryFieldErrFmt, "endpoint", "kms")))
	}

	u, err := url.Parse(c.Endpoint)
	if err != nil {
		return append(allErrs, field.Invalid(fieldPath, c.Endpoint, fmt.Sprintf("invalid endpoint for kms provider, error: %v", err)))
	}

	if u.Scheme != "unix" {
		return append(allErrs, field.Invalid(fieldPath, c.Endpoint, fmt.Sprintf(unsupportedSchemeErrFmt, u.Scheme)))
	}

	return allErrs
}
