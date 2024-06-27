package options

import (
	"kubesphere.io/utils/s3"

	"kubesphere.io/kubesphere/pkg/apiserver/auditing"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization"
	"kubesphere.io/kubesphere/pkg/models/terminal"
	"kubesphere.io/kubesphere/pkg/multicluster"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
)

type Options struct {
	MultiClusterOptions   *multicluster.Options   `json:"multicluster"`
	AuthenticationOptions *authentication.Options `json:"-"`
	KubernetesOptions     *k8s.Options            `json:"-"`
	CacheOptions          *cache.Options          `json:"-"`
	AuthorizationOptions  *authorization.Options  `json:"-"`
	AuditingOptions       *auditing.Options       `json:"-"`
	TerminalOptions       *terminal.Options       `json:"-"`
	S3Options             *s3.Options             `json:"-"`
}
