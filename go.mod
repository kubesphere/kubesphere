// This is a generated file. Do not edit directly.
// Ensure you've carefully read
// https://git.k8s.io/community/contributors/devel/sig-architecture/vendor.md
// Run hack/pin-dependency.sh to change pinned dependency versions.
// Run hack/update-vendor.sh to update go.mod files and the vendor directory.

module kubesphere.io/kubesphere

go 1.19

require (
	code.cloudfoundry.org/bytefmt v0.0.0-20190710193110-1eb035ffe2b6
	github.com/Masterminds/semver/v3 v3.2.0
	github.com/PuerkitoBio/goquery v1.5.0
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/aws/aws-sdk-go v1.44.187
	github.com/beevik/etree v1.1.0
	github.com/containernetworking/cni v1.1.2
	github.com/coreos/go-oidc v2.1.0+incompatible
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/distribution v2.8.2+incompatible
	github.com/docker/docker v24.0.9+incompatible
	github.com/elastic/go-elasticsearch/v5 v5.6.1
	github.com/elastic/go-elasticsearch/v6 v6.8.2
	github.com/elastic/go-elasticsearch/v7 v7.3.0
	github.com/emicklei/go-restful-openapi/v2 v2.9.1
	github.com/emicklei/go-restful/v3 v3.11.0
	github.com/evanphx/json-patch v5.6.0+incompatible
	github.com/fatih/structs v1.1.0
	github.com/fsnotify/fsnotify v1.6.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-ldap/ldap v3.0.3+incompatible
	github.com/go-logr/logr v1.4.1
	github.com/go-openapi/loads v0.21.2
	github.com/go-openapi/spec v0.20.7
	github.com/go-openapi/strfmt v0.21.3
	github.com/go-openapi/validate v0.22.0
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/golang-jwt/jwt/v4 v4.2.0
	github.com/golang/example v0.0.0-20170904185048-46695d81d1fa
	github.com/google/go-cmp v0.6.0
	github.com/google/go-containerregistry v0.14.0
	github.com/google/gops v0.3.23
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.5.0
	github.com/hashicorp/golang-lru v0.6.0
	github.com/json-iterator/go v1.1.12
	github.com/jszwec/csvutil v1.5.0
	github.com/kubernetes-csi/external-snapshotter/client/v4 v4.2.0
	github.com/kubesphere/pvc-autoresizer v0.3.1
	github.com/kubesphere/sonargo v0.0.2
	github.com/kubesphere/storageclass-accessor v0.2.4-0.20230919084454-2f39c69db301
	github.com/mitchellh/mapstructure v1.5.0
	github.com/moby/locker v1.0.1
	github.com/moby/term v0.0.0-20221205130635-1aeaba878587
	github.com/oliveagle/jsonpath v0.0.0-20180606110733-2e52cf6e6852
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.27.10
	github.com/open-policy-agent/opa v0.49.0
	github.com/opencontainers/go-digest v1.0.0
	github.com/opensearch-project/opensearch-go v1.1.0
	github.com/opensearch-project/opensearch-go/v2 v2.0.0
	github.com/operator-framework/helm-operator-plugins v0.0.11
	github.com/pkg/errors v0.9.1
	github.com/projectcalico/api v0.0.0
	github.com/projectcalico/calico v0.0.0-20230227071013-a73515ddc939
	github.com/prometheus-community/prom-label-proxy v0.6.0
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.63.0
	github.com/prometheus-operator/prometheus-operator/pkg/client v0.63.0
	github.com/prometheus/client_golang v1.14.0
	github.com/prometheus/common v0.39.0
	github.com/prometheus/prometheus v0.42.0
	github.com/sony/sonyflake v0.0.0-20181109022403-6d5bd6181009
	github.com/speps/go-hashids v2.0.0+incompatible
	github.com/spf13/cobra v1.7.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.13.0
	github.com/stretchr/testify v1.8.4
	golang.org/x/crypto v0.21.0
	golang.org/x/oauth2 v0.10.0
	google.golang.org/grpc v1.58.3
	gopkg.in/cas.v2 v2.2.0
	gopkg.in/square/go-jose.v2 v2.5.1
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	gotest.tools v2.2.0+incompatible
	helm.sh/helm/v3 v3.11.1
	istio.io/api v0.0.0-20201113182140-d4b7e3fc2b44
	istio.io/client-go v0.0.0-20201113183938-0734e976e785
	k8s.io/api v0.26.2
	k8s.io/apiextensions-apiserver v0.26.1
	k8s.io/apimachinery v0.26.2
	k8s.io/apiserver v0.26.2
	k8s.io/cli-runtime v0.26.1
	k8s.io/client-go v0.26.2
	k8s.io/code-generator v0.26.1
	k8s.io/component-base v0.26.2
	k8s.io/klog/v2 v2.90.1
	k8s.io/kube-openapi v0.0.0-20230224204730-66828de6f33b
	k8s.io/kubectl v0.26.1
	k8s.io/metrics v0.26.1
	k8s.io/utils v0.0.0-20230220204549-a5ecb0141aa5
	kubesphere.io/api v0.0.0
	kubesphere.io/client-go v0.0.0
	kubesphere.io/monitoring-dashboard v0.2.2
	sigs.k8s.io/application v0.8.4-0.20201016185654-c8e2959e57a0
	sigs.k8s.io/controller-runtime v0.14.4
	sigs.k8s.io/controller-tools v0.11.1
	sigs.k8s.io/kubefed v0.0.0-20230207032540-cdda80892665
	sigs.k8s.io/kustomize/api v0.12.1
	sigs.k8s.io/kustomize/kyaml v0.13.9
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20230811130428-ced1acdcaa24 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/sprig/v3 v3.2.3 // indirect
	github.com/Masterminds/squirrel v1.5.3 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/Microsoft/hcsshim v0.11.4 // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/OneOfOne/xxhash v1.2.8 // indirect
	github.com/agnivade/levenshtein v1.1.1 // indirect
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/andybalholm/cascadia v1.0.0 // indirect
	github.com/antlr/antlr4/runtime/Go/antlr v1.4.10 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/containerd/containerd v1.7.13 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/cyphar/filepath-securejoin v0.2.4 // indirect
	github.com/deckarep/golang-set v1.8.0 // indirect
	github.com/dennwc/varint v1.0.0 // indirect
	github.com/docker/cli v24.0.6+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.7.0 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/edsrzf/mmap-go v1.1.0 // indirect
	github.com/efficientgo/tools/core v0.0.0-20220225185207-fe763185946b // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/go-errors/errors v1.0.1 // indirect
	github.com/go-gorp/gorp/v3 v3.0.2 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/analysis v0.21.4 // indirect
	github.com/go-openapi/errors v0.20.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/runtime v0.25.0 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/go-resty/resty/v2 v2.5.0 // indirect
	github.com/go-task/slim-sprig v0.0.0-20210107165309-348f09dbbbc0 // indirect
	github.com/gobuffalo/flect v0.3.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/glog v1.1.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/cel-go v0.12.6 // indirect
	github.com/google/gnostic v0.6.9 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/gosimple/slug v1.1.1 // indirect
	github.com/gosuri/uitable v0.0.4 // indirect
	github.com/grafana-tools/sdk v0.0.0-20210625151406-43693eb2f02b // indirect
	github.com/grafana/regexp v0.0.0-20221122212121-6b5c0a4cb7fd // indirect
	github.com/gregjones/httpcache v0.0.0-20181110185634-c63ab54fda8f // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.16.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/huandu/xstrings v1.3.3 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jmoiron/sqlx v1.3.5 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/compress v1.16.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/lib/pq v1.10.7 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/metalmatze/signal v0.0.0-20210307161603-1c9aa721a97a // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc5 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/operator-framework/operator-lib v0.11.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/pquerna/cachecontrol v0.1.0 // indirect
	github.com/prometheus/alertmanager v0.25.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common/sigv4 v0.1.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/rainycape/unidecode v0.0.0-20150907023854-cb7f23ec59be // indirect
	github.com/rcrowley/go-metrics v0.0.0-20200313005456-10cdbea86bc0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/rubenv/sql-migrate v1.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/afero v1.9.2 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/src-d/gcfg v1.4.0 // indirect
	github.com/stoewer/go-strcase v1.2.0 // indirect
	github.com/tchap/go-patricia/v2 v2.3.1 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/xlab/treeprint v1.1.0 // indirect
	github.com/yashtewari/glob-intersection v0.1.0 // indirect
	go.etcd.io/etcd/api/v3 v3.5.7 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.7 // indirect
	go.etcd.io/etcd/client/v3 v3.5.7 // indirect
	go.mongodb.org/mongo-driver v1.11.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.45.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.45.0 // indirect
	go.opentelemetry.io/otel v1.23.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.19.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.19.0 // indirect
	go.opentelemetry.io/otel/metric v1.23.0 // indirect
	go.opentelemetry.io/otel/sdk v1.23.0 // indirect
	go.opentelemetry.io/otel/trace v1.23.0 // indirect
	go.opentelemetry.io/proto/otlp v1.0.0 // indirect
	go.starlark.net v0.0.0-20200306205701-8dd3e2ee1dd5 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/goleak v1.2.1 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/exp v0.0.0-20230124195608-d38c7dcee874 // indirect
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/term v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.13.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230711160842-782d3b101e98 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/asn1-ber.v1 v1.0.0-20181015200546-f715ec2f112d // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.2 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	istio.io/gogo-genproto v0.0.0-20201113182723-5b8563d8a012 // indirect
	k8s.io/gengo v0.0.0-20220902162205-c0856e24416d // indirect
	k8s.io/kms v0.26.1 // indirect
	oras.land/oras-go v1.2.4 // indirect
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.0.35 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)

replace (
	code.cloudfoundry.org/bytefmt => code.cloudfoundry.org/bytefmt v0.0.0-20190710193110-1eb035ffe2b6
	github.com/Azure/go-ansiterm => github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1
	github.com/BurntSushi/toml => github.com/BurntSushi/toml v1.2.1
	github.com/MakeNowJust/heredoc => github.com/MakeNowJust/heredoc v1.0.0
	github.com/Masterminds/goutils => github.com/Masterminds/goutils v1.1.1
	github.com/Masterminds/semver/v3 => github.com/Masterminds/semver/v3 v3.2.0
	github.com/Masterminds/sprig/v3 => github.com/Masterminds/sprig/v3 v3.2.3
	github.com/Masterminds/squirrel => github.com/Masterminds/squirrel v1.5.3
	github.com/Microsoft/go-winio => github.com/Microsoft/go-winio v0.6.1
	github.com/OneOfOne/xxhash => github.com/OneOfOne/xxhash v1.2.8
	github.com/PuerkitoBio/goquery => github.com/PuerkitoBio/goquery v1.5.0
	github.com/agnivade/levenshtein => github.com/agnivade/levenshtein v1.1.1
	github.com/alecthomas/units => github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137
	github.com/andybalholm/cascadia => github.com/andybalholm/cascadia v1.0.0
	github.com/asaskevich/govalidator => github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/aws/aws-sdk-go => github.com/aws/aws-sdk-go v1.44.187
	github.com/beevik/etree => github.com/beevik/etree v1.1.0
	github.com/beorn7/perks => github.com/beorn7/perks v1.0.1
	github.com/blang/semver/v4 => github.com/blang/semver/v4 v4.0.0
	github.com/cespare/xxhash/v2 => github.com/cespare/xxhash/v2 v2.2.0
	github.com/chai2010/gettext-go => github.com/chai2010/gettext-go v1.0.2
	github.com/containerd/containerd => github.com/containerd/containerd v1.7.13
	github.com/containernetworking/cni => github.com/containernetworking/cni v1.1.2
	github.com/coreos/go-oidc => github.com/coreos/go-oidc v2.1.0+incompatible
	github.com/cyphar/filepath-securejoin => github.com/cyphar/filepath-securejoin v0.2.3
	github.com/davecgh/go-spew => github.com/davecgh/go-spew v1.1.1
	github.com/docker/cli => github.com/docker/cli v20.10.22+incompatible
	github.com/docker/distribution => github.com/docker/distribution v2.8.2+incompatible
	github.com/docker/docker => github.com/docker/docker v24.0.9+incompatible
	github.com/docker/docker-credential-helpers => github.com/docker/docker-credential-helpers v0.7.0
	github.com/docker/go-connections => github.com/docker/go-connections v0.4.0
	github.com/docker/go-metrics => github.com/docker/go-metrics v0.0.1
	github.com/docker/go-units => github.com/docker/go-units v0.5.0
	github.com/edsrzf/mmap-go => github.com/edsrzf/mmap-go v1.1.0
	github.com/elastic/go-elasticsearch/v5 => github.com/elastic/go-elasticsearch/v5 v5.6.1
	github.com/elastic/go-elasticsearch/v6 => github.com/elastic/go-elasticsearch/v6 v6.8.2
	github.com/elastic/go-elasticsearch/v7 => github.com/elastic/go-elasticsearch/v7 v7.3.0
	github.com/emicklei/go-restful-openapi/v2 => github.com/emicklei/go-restful-openapi/v2 v2.9.2-0.20230507070325-d6acc08e570c
	github.com/emicklei/go-restful/v3 => github.com/emicklei/go-restful/v3 v3.11.0
	github.com/emirpasic/gods => github.com/emirpasic/gods v1.12.0
	github.com/evanphx/json-patch => github.com/evanphx/json-patch v5.6.0+incompatible
	github.com/evanphx/json-patch/v5 => github.com/evanphx/json-patch/v5 v5.6.0
	github.com/exponent-io/jsonpath => github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d
	github.com/fatih/color => github.com/fatih/color v1.13.0
	github.com/fatih/structs => github.com/fatih/structs v1.1.0
	github.com/fsnotify/fsnotify => github.com/fsnotify/fsnotify v1.6.0
	github.com/ghodss/yaml => github.com/ghodss/yaml v1.0.0
	github.com/go-errors/errors => github.com/go-errors/errors v1.0.1
	github.com/go-gorp/gorp/v3 => github.com/go-gorp/gorp/v3 v3.0.2
	github.com/go-kit/log => github.com/go-kit/log v0.2.1
	github.com/go-ldap/ldap => github.com/go-ldap/ldap v3.0.3+incompatible
	github.com/go-logfmt/logfmt => github.com/go-logfmt/logfmt v0.5.1
	github.com/go-logr/logr => github.com/go-logr/logr v1.2.3
	github.com/go-logr/stdr => github.com/go-logr/stdr v1.2.2
	github.com/go-openapi/analysis => github.com/go-openapi/analysis v0.21.4
	github.com/go-openapi/errors => github.com/go-openapi/errors v0.20.3
	github.com/go-openapi/jsonpointer => github.com/go-openapi/jsonpointer v0.19.6
	github.com/go-openapi/jsonreference => github.com/go-openapi/jsonreference v0.20.2
	github.com/go-openapi/loads => github.com/go-openapi/loads v0.21.2
	github.com/go-openapi/runtime => github.com/go-openapi/runtime v0.25.0
	github.com/go-openapi/spec => github.com/go-openapi/spec v0.20.7
	github.com/go-openapi/strfmt => github.com/go-openapi/strfmt v0.21.3
	github.com/go-openapi/swag => github.com/go-openapi/swag v0.22.3
	github.com/go-openapi/validate => github.com/go-openapi/validate v0.22.0
	github.com/go-redis/redis => github.com/go-redis/redis v6.15.2+incompatible
	github.com/go-task/slim-sprig => github.com/go-task/slim-sprig v0.0.0-20210107165309-348f09dbbbc0
	github.com/gobuffalo/flect => github.com/gobuffalo/flect v0.3.0
	github.com/gobwas/glob => github.com/gobwas/glob v0.2.3
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
	github.com/golang-jwt/jwt/v4 => github.com/golang-jwt/jwt/v4 v4.2.0
	github.com/golang/example => github.com/golang/example v0.0.0-20170904185048-46695d81d1fa
	github.com/golang/glog => github.com/golang/glog v1.0.0
	github.com/golang/groupcache => github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da
	github.com/golang/protobuf => github.com/golang/protobuf v1.5.2
	github.com/golang/snappy => github.com/golang/snappy v0.0.4
	github.com/google/btree => github.com/google/btree v1.1.2
	github.com/google/gnostic => github.com/google/gnostic v0.6.9
	github.com/google/go-cmp => github.com/google/go-cmp v0.5.9
	github.com/google/go-containerregistry => github.com/google/go-containerregistry v0.5.1
	github.com/google/go-querystring => github.com/google/go-querystring v1.1.0
	github.com/google/gofuzz => github.com/google/gofuzz v1.2.0
	github.com/google/gops => github.com/google/gops v0.3.23
	github.com/google/shlex => github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/google/uuid => github.com/google/uuid v1.3.0
	github.com/gorilla/mux => github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket => github.com/gorilla/websocket v1.5.0
	github.com/gosimple/slug => github.com/gosimple/slug v1.1.1
	github.com/gosuri/uitable => github.com/gosuri/uitable v0.0.4
	github.com/grafana-tools/sdk => github.com/grafana-tools/sdk v0.0.0-20210625151406-43693eb2f02b
	github.com/gregjones/httpcache => github.com/gregjones/httpcache v0.0.0-20181110185634-c63ab54fda8f
	github.com/hashicorp/golang-lru => github.com/hashicorp/golang-lru v0.6.0
	github.com/hashicorp/hcl => github.com/hashicorp/hcl v1.0.0
	github.com/huandu/xstrings => github.com/huandu/xstrings v1.3.3
	github.com/imdario/mergo => github.com/imdario/mergo v0.3.12
	github.com/inconshreveable/mousetrap => github.com/inconshreveable/mousetrap v1.0.1
	github.com/jbenet/go-context => github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99
	github.com/jmespath/go-jmespath => github.com/jmespath/go-jmespath v0.4.0
	github.com/jmoiron/sqlx => github.com/jmoiron/sqlx v1.3.5
	github.com/josharian/intern => github.com/josharian/intern v1.0.0
	github.com/jpillora/backoff => github.com/jpillora/backoff v1.0.0
	github.com/json-iterator/go => github.com/json-iterator/go v1.1.12
	github.com/jszwec/csvutil => github.com/jszwec/csvutil v1.5.0
	github.com/kevinburke/ssh_config => github.com/kevinburke/ssh_config v0.0.0-20190725054713-01f96b0aa0cd
	github.com/klauspost/compress => github.com/klauspost/compress v1.13.6
	github.com/kubernetes-csi/external-snapshotter/client/v4 => github.com/kubernetes-csi/external-snapshotter/client/v4 v4.2.0
	github.com/kubesphere/pvc-autoresizer => github.com/kubesphere/pvc-autoresizer v0.3.0
	github.com/kubesphere/sonargo => github.com/kubesphere/sonargo v0.0.2
	github.com/kubesphere/storageclass-accessor => github.com/kubesphere/storageclass-accessor v0.2.4-0.20230919084454-2f39c69db301
	github.com/lann/builder => github.com/lann/builder v0.0.0-20180802200727-47ae307949d0
	github.com/lann/ps => github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0
	github.com/lib/pq => github.com/lib/pq v1.10.7
	github.com/liggitt/tabwriter => github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de
	github.com/magiconair/properties => github.com/magiconair/properties v1.8.6
	github.com/mailru/easyjson => github.com/mailru/easyjson v0.7.7
	github.com/mattn/go-colorable => github.com/mattn/go-colorable v0.1.12
	github.com/mattn/go-isatty => github.com/mattn/go-isatty v0.0.14
	github.com/mattn/go-runewidth => github.com/mattn/go-runewidth v0.0.9
	github.com/matttproud/golang_protobuf_extensions => github.com/matttproud/golang_protobuf_extensions v1.0.4
	github.com/mitchellh/copystructure => github.com/mitchellh/copystructure v1.2.0
	github.com/mitchellh/go-homedir => github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-wordwrap => github.com/mitchellh/go-wordwrap v1.0.0
	github.com/mitchellh/mapstructure => github.com/mitchellh/mapstructure v1.5.0
	github.com/mitchellh/reflectwalk => github.com/mitchellh/reflectwalk v1.0.2
	github.com/moby/locker => github.com/moby/locker v1.0.1
	github.com/moby/spdystream => github.com/moby/spdystream v0.2.0
	github.com/moby/term => github.com/moby/term v0.0.0-20221205130635-1aeaba878587
	github.com/modern-go/concurrent => github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd
	github.com/modern-go/reflect2 => github.com/modern-go/reflect2 v1.0.2
	github.com/monochromegane/go-gitignore => github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00
	github.com/morikuni/aec => github.com/morikuni/aec v1.0.0
	github.com/munnerz/goautoneg => github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822
	github.com/mwitkow/go-conntrack => github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f
	github.com/mxk/go-flowrate => github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f
	github.com/nxadm/tail => github.com/nxadm/tail v1.4.8
	github.com/oklog/ulid => github.com/oklog/ulid v1.3.1
	github.com/oliveagle/jsonpath => github.com/oliveagle/jsonpath v0.0.0-20180606110733-2e52cf6e6852
	github.com/onsi/ginkgo => github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega => github.com/onsi/gomega v1.27.1
	github.com/open-policy-agent/opa => github.com/open-policy-agent/opa v0.49.0
	github.com/opencontainers/go-digest => github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec => github.com/opencontainers/image-spec v1.1.0-rc2
	github.com/opensearch-project/opensearch-go => github.com/opensearch-project/opensearch-go v1.1.0
	github.com/opensearch-project/opensearch-go/v2 => github.com/opensearch-project/opensearch-go/v2 v2.0.0
	github.com/opentracing/opentracing-go => github.com/opentracing/opentracing-go v1.2.0
	github.com/operator-framework/helm-operator-plugins => github.com/operator-framework/helm-operator-plugins v0.0.11
	github.com/operator-framework/operator-lib => github.com/operator-framework/operator-lib v0.11.0
	github.com/patrickmn/go-cache => github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pelletier/go-toml => github.com/pelletier/go-toml v1.9.5
	github.com/peterbourgon/diskv => github.com/peterbourgon/diskv v2.0.1+incompatible
	github.com/pkg/errors => github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib => github.com/pmezard/go-difflib v1.0.0
	github.com/pquerna/cachecontrol => github.com/pquerna/cachecontrol v0.1.0
	github.com/projectcalico/api => github.com/kubesphere/calico/api v0.0.0-20230227071013-a73515ddc939 // v3.25.0
	github.com/projectcalico/calico => github.com/kubesphere/calico v0.0.0-20230227071013-a73515ddc939 // v3.25.0
	github.com/prometheus-community/prom-label-proxy => github.com/prometheus-community/prom-label-proxy v0.6.0
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring => github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.63.0
	github.com/prometheus-operator/prometheus-operator/pkg/client => github.com/prometheus-operator/prometheus-operator/pkg/client v0.63.0
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.14.0
	github.com/prometheus/common => github.com/prometheus/common v0.39.0
	github.com/prometheus/procfs => github.com/prometheus/procfs v0.8.0
	github.com/prometheus/prometheus => github.com/prometheus/prometheus v0.42.0
	github.com/rainycape/unidecode => github.com/rainycape/unidecode v0.0.0-20150907023854-cb7f23ec59be
	github.com/rcrowley/go-metrics => github.com/rcrowley/go-metrics v0.0.0-20200313005456-10cdbea86bc0
	github.com/rubenv/sql-migrate => github.com/rubenv/sql-migrate v1.2.0
	github.com/russross/blackfriday/v2 => github.com/russross/blackfriday/v2 v2.1.0
	github.com/sergi/go-diff => github.com/sergi/go-diff v1.1.0
	github.com/shopspring/decimal => github.com/shopspring/decimal v1.2.0
	github.com/sirupsen/logrus => github.com/sirupsen/logrus v1.9.0
	github.com/sony/sonyflake => github.com/sony/sonyflake v0.0.0-20181109022403-6d5bd6181009
	github.com/speps/go-hashids => github.com/speps/go-hashids v2.0.0+incompatible
	github.com/spf13/afero => github.com/spf13/afero v1.9.2
	github.com/spf13/cast => github.com/spf13/cast v1.5.0
	github.com/spf13/cobra => github.com/spf13/cobra v1.6.1
	github.com/spf13/jwalterweatherman => github.com/spf13/jwalterweatherman v1.1.0
	github.com/spf13/pflag => github.com/spf13/pflag v1.0.5
	github.com/spf13/viper => github.com/spf13/viper v1.4.0
	github.com/src-d/gcfg => github.com/src-d/gcfg v1.4.0
	github.com/stretchr/testify => github.com/stretchr/testify v1.8.1
	github.com/tchap/go-patricia/v2 => github.com/tchap/go-patricia/v2 v2.3.1
	github.com/xanzy/ssh-agent => github.com/xanzy/ssh-agent v0.2.1
	github.com/xeipuuv/gojsonpointer => github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb
	github.com/xeipuuv/gojsonreference => github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415
	github.com/xeipuuv/gojsonschema => github.com/xeipuuv/gojsonschema v1.2.0
	github.com/xlab/treeprint => github.com/xlab/treeprint v1.1.0
	github.com/yashtewari/glob-intersection => github.com/yashtewari/glob-intersection v0.1.0
	go.mongodb.org/mongo-driver => go.mongodb.org/mongo-driver v1.11.0
	go.opentelemetry.io/otel => go.opentelemetry.io/otel v1.23.0
	go.opentelemetry.io/otel/trace => go.opentelemetry.io/otel/trace v1.23.0
	go.starlark.net => go.starlark.net v0.0.0-20200306205701-8dd3e2ee1dd5
	go.uber.org/atomic => go.uber.org/atomic v1.10.0
	go.uber.org/goleak => go.uber.org/goleak v1.2.0
	golang.org/x/crypto => golang.org/x/crypto v0.5.0
	golang.org/x/net => golang.org/x/net v0.23.0
	golang.org/x/oauth2 => golang.org/x/oauth2 v0.4.0
	golang.org/x/sync => golang.org/x/sync v0.1.0
	golang.org/x/sys => golang.org/x/sys v0.5.0
	golang.org/x/term => golang.org/x/term v0.5.0
	golang.org/x/text => golang.org/x/text v0.7.0
	golang.org/x/time => golang.org/x/time v0.3.0
	golang.org/x/tools => golang.org/x/tools v0.6.0
	gomodules.xyz/jsonpatch/v2 => gomodules.xyz/jsonpatch/v2 v2.2.0
	google.golang.org/appengine => google.golang.org/appengine v1.6.7
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20230124163310-31e0e69b6fc2
	google.golang.org/grpc => google.golang.org/grpc v1.56.3
	google.golang.org/protobuf => google.golang.org/protobuf v1.33.0
	gopkg.in/asn1-ber.v1 => gopkg.in/asn1-ber.v1 v1.0.0-20181015200546-f715ec2f112d
	gopkg.in/cas.v2 => gopkg.in/cas.v2 v2.2.0
	gopkg.in/inf.v0 => gopkg.in/inf.v0 v0.9.1
	gopkg.in/square/go-jose.v2 => gopkg.in/square/go-jose.v2 v2.5.1
	gopkg.in/src-d/go-billy.v4 => gopkg.in/src-d/go-billy.v4 v4.3.2
	gopkg.in/src-d/go-git.v4 => gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/tomb.v1 => gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7
	gopkg.in/warnings.v0 => gopkg.in/warnings.v0 v0.1.2
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 => gopkg.in/yaml.v3 v3.0.1
	gotest.tools => gotest.tools v2.2.0+incompatible
	helm.sh/helm/v3 => helm.sh/helm/v3 v3.11.1
	istio.io/api => istio.io/api v0.0.0-20201113182140-d4b7e3fc2b44
	istio.io/client-go => istio.io/client-go v0.0.0-20201113183938-0734e976e785
	istio.io/gogo-genproto => istio.io/gogo-genproto v0.0.0-20201113182723-5b8563d8a012
	k8s.io/api => k8s.io/api v0.26.1
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.26.1
	k8s.io/apimachinery => k8s.io/apimachinery v0.26.1
	k8s.io/apiserver => k8s.io/apiserver v0.26.1
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.26.1
	k8s.io/client-go => k8s.io/client-go v0.26.1
	k8s.io/code-generator => k8s.io/code-generator v0.26.1
	k8s.io/component-base => k8s.io/component-base v0.26.1
	k8s.io/gengo => k8s.io/gengo v0.0.0-20220902162205-c0856e24416d
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.90.0
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20230224204730-66828de6f33b
	k8s.io/kubectl => k8s.io/kubectl v0.26.1
	k8s.io/metrics => k8s.io/metrics v0.26.1
	k8s.io/utils => k8s.io/utils v0.0.0-20230202215443-34013725500c
	kubesphere.io/api => ./staging/src/kubesphere.io/api
	kubesphere.io/client-go => ./staging/src/kubesphere.io/client-go
	kubesphere.io/monitoring-dashboard => kubesphere.io/monitoring-dashboard v0.2.2
	kubesphere.io/utils => ./staging/src/kubesphere.io/utils
	oras.land/oras-go => oras.land/oras-go v1.2.4
	sigs.k8s.io/application => sigs.k8s.io/application v0.8.4-0.20201016185654-c8e2959e57a0
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.14.4
	sigs.k8s.io/controller-tools => sigs.k8s.io/controller-tools v0.11.1
	sigs.k8s.io/json => sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd
	sigs.k8s.io/kubefed => github.com/kubesphere/kubefed v0.0.0-20230207032540-cdda80892665
	sigs.k8s.io/kustomize/api => sigs.k8s.io/kustomize/api v0.12.1
	sigs.k8s.io/kustomize/kyaml => sigs.k8s.io/kustomize/kyaml v0.13.9
	sigs.k8s.io/structured-merge-diff/v4 => sigs.k8s.io/structured-merge-diff/v4 v4.2.3
	sigs.k8s.io/yaml => sigs.k8s.io/yaml v1.3.0
)
