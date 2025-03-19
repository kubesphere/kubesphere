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
