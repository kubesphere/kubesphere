/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package kubeconfig

const (
	AuthModeServiceAccountToken AuthMode = "service-account-token"
	AuthModeClientCertificate   AuthMode = "client-certificate"
	AuthModeOIDCToken           AuthMode = "oidc-token"
	AuthModeWebhookToken        AuthMode = "webhook-token"
)

type AuthMode string

type Options struct {
	AuthMode AuthMode `json:"authMode" yaml:"authMode" mapstructure:"authMode"`
}

func NewOptions() *Options {
	return &Options{AuthMode: AuthModeClientCertificate}
}
