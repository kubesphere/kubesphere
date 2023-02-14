// This is a generated file. Do not edit directly.
// Ensure you've carefully read
// https://git.k8s.io/community/contributors/devel/sig-architecture/vendor.md
// Run hack/pin-dependency.sh to change pinned dependency versions.
// Run hack/update-vendor.sh to update go.mod files and the vendor directory.

module kubesphere.io/kubesphere

go 1.19

require (
	code.cloudfoundry.org/bytefmt v0.0.0-20190710193110-1eb035ffe2b6
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/PuerkitoBio/goquery v1.5.0
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/aws/aws-sdk-go v1.44.187
	github.com/beevik/etree v1.1.0
	github.com/containernetworking/cni v1.1.2
	github.com/coreos/go-oidc v2.1.0+incompatible
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/distribution v2.8.1+incompatible
	github.com/docker/docker v20.10.23+incompatible
	github.com/elastic/go-elasticsearch/v5 v5.6.1
	github.com/elastic/go-elasticsearch/v6 v6.8.2
	github.com/elastic/go-elasticsearch/v7 v7.3.0
	github.com/emicklei/go-restful v2.16.0+incompatible
	github.com/emicklei/go-restful-openapi v1.4.1
	github.com/evanphx/json-patch v5.6.0+incompatible
	github.com/fatih/structs v1.1.0
	github.com/form3tech-oss/jwt-go v3.2.3+incompatible
	github.com/fsnotify/fsnotify v1.6.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-ldap/ldap v3.0.3+incompatible
	github.com/go-logr/logr v1.2.3
	github.com/go-openapi/loads v0.21.2
	github.com/go-openapi/spec v0.20.7
	github.com/go-openapi/strfmt v0.21.3
	github.com/go-openapi/validate v0.22.0
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/golang/example v0.0.0-20170904185048-46695d81d1fa
	github.com/google/go-cmp v0.5.9
	github.com/google/go-containerregistry v0.5.1
	github.com/google/gops v0.3.23
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.5.0
	github.com/hashicorp/golang-lru v0.6.0
	github.com/json-iterator/go v1.1.12
	github.com/jszwec/csvutil v1.5.0
	github.com/kubernetes-csi/external-snapshotter/client/v4 v4.2.0
	github.com/kubesphere/pvc-autoresizer v0.3.1
	github.com/kubesphere/sonargo v0.0.2
	github.com/kubesphere/storageclass-accessor v0.2.2
	github.com/mitchellh/mapstructure v1.5.0
	github.com/oliveagle/jsonpath v0.0.0-20180606110733-2e52cf6e6852
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.26.0
	github.com/open-policy-agent/opa v0.49.0
	github.com/opencontainers/go-digest v1.0.0
	github.com/opensearch-project/opensearch-go v1.1.0
	github.com/opensearch-project/opensearch-go/v2 v2.0.0
	github.com/operator-framework/helm-operator-plugins v0.0.11
	github.com/pkg/errors v0.9.1
	github.com/projectcalico/kube-controllers v3.8.8+incompatible
	github.com/projectcalico/libcalico-go v1.7.2-0.20191014160346-2382c6cdd056
	github.com/prometheus-community/prom-label-proxy v0.6.0
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.63.0
	github.com/prometheus-operator/prometheus-operator/pkg/client v0.63.0
	github.com/prometheus/client_golang v1.14.0
	github.com/prometheus/common v0.39.0
	github.com/prometheus/prometheus v1.8.2-0.20200907175821-8219b442c864
	github.com/sony/sonyflake v0.0.0-20181109022403-6d5bd6181009
	github.com/speps/go-hashids v2.0.0+incompatible
	github.com/spf13/cobra v1.6.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.13.0
	github.com/stretchr/testify v1.8.1
	golang.org/x/crypto v0.5.0
	golang.org/x/oauth2 v0.4.0
	google.golang.org/grpc v1.52.3
	gopkg.in/cas.v2 v2.2.0
	gopkg.in/square/go-jose.v2 v2.5.1
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	gotest.tools v2.2.0+incompatible
	helm.sh/helm/v3 v3.10.3
	istio.io/api v0.0.0-20201113182140-d4b7e3fc2b44
	istio.io/client-go v0.0.0-20201113183938-0734e976e785
	k8s.io/api v0.26.1
	k8s.io/apiextensions-apiserver v0.26.1
	k8s.io/apimachinery v0.26.1
	k8s.io/apiserver v0.26.1
	k8s.io/cli-runtime v0.26.1
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.26.1
	k8s.io/component-base v0.26.1
	k8s.io/klog/v2 v2.90.0
	k8s.io/kube-openapi v0.0.0-20230202010329-39b3636cbaa3
	k8s.io/kubectl v0.26.1
	k8s.io/metrics v0.26.1
	k8s.io/utils v0.0.0-20230202215443-34013725500c
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
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/BurntSushi/toml v1.1.0 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/sprig/v3 v3.2.2 // indirect
	github.com/Masterminds/squirrel v1.5.3 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/OneOfOne/xxhash v1.2.8 // indirect
	github.com/agnivade/levenshtein v1.1.1 // indirect
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/andybalholm/cascadia v1.0.0 // indirect
	github.com/antlr/antlr4/runtime/Go/antlr v1.4.10 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cenkalti/backoff/v4 v4.2.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/containerd/containerd v1.6.16 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.4.0 // indirect
	github.com/cyphar/filepath-securejoin v0.2.3 // indirect
	github.com/deckarep/golang-set v0.0.0-00010101000000-000000000000 // indirect
	github.com/dennwc/varint v1.0.0 // indirect
	github.com/docker/cli v20.10.22+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.7.0 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/edsrzf/mmap-go v1.1.0 // indirect
	github.com/efficientgo/tools/core v0.0.0-20220225185207-fe763185946b // indirect
	github.com/emicklei/go-restful/v3 v3.10.1 // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
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
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.0.1 // indirect
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
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.11.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jmoiron/sqlx v1.3.5 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20190725054713-01f96b0aa0cd // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/lib/pq v1.10.6 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/magiconair/properties v1.8.5 // indirect
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
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/moby/term v0.0.0-20221205130635-1aeaba878587 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc2 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/operator-framework/operator-lib v0.11.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/pquerna/cachecontrol v0.1.0 // indirect
	github.com/prometheus/alertmanager v0.25.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common/sigv4 v0.1.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/rainycape/unidecode v0.0.0-20150907023854-cb7f23ec59be // indirect
	github.com/rcrowley/go-metrics v0.0.0-20200313005456-10cdbea86bc0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/rubenv/sql-migrate v1.1.2 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/src-d/gcfg v1.4.0 // indirect
	github.com/stoewer/go-strcase v1.2.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/tchap/go-patricia/v2 v2.3.1 // indirect
	github.com/xanzy/ssh-agent v0.2.1 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/xlab/treeprint v1.1.0 // indirect
	github.com/yashtewari/glob-intersection v0.1.0 // indirect
	go.etcd.io/etcd/api/v3 v3.5.5 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.5 // indirect
	go.etcd.io/etcd/client/v3 v3.5.5 // indirect
	go.mongodb.org/mongo-driver v1.11.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.35.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.37.0 // indirect
	go.opentelemetry.io/otel v1.11.2 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.11.2 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.11.2 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.11.2 // indirect
	go.opentelemetry.io/otel/metric v0.34.0 // indirect
	go.opentelemetry.io/otel/sdk v1.11.2 // indirect
	go.opentelemetry.io/otel/trace v1.11.2 // indirect
	go.opentelemetry.io/proto/otlp v0.19.0 // indirect
	go.starlark.net v0.0.0-20200306205701-8dd3e2ee1dd5 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/goleak v1.2.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/exp v0.0.0-20230124195608-d38c7dcee874 // indirect
	golang.org/x/net v0.5.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.4.0 // indirect
	golang.org/x/term v0.4.0 // indirect
	golang.org/x/text v0.6.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.5.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230124163310-31e0e69b6fc2 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/asn1-ber.v1 v1.0.0-20181015200546-f715ec2f112d // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.66.6 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.2 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	istio.io/gogo-genproto v0.0.0-20201113182723-5b8563d8a012 // indirect
	k8s.io/gengo v0.0.0-20220902162205-c0856e24416d // indirect
	k8s.io/kms v0.26.1 // indirect
	oras.land/oras-go v1.2.2 // indirect
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.0.35 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)

replace (
	cloud.google.com/go => cloud.google.com/go v0.105.0
	cloud.google.com/go/accessapproval => cloud.google.com/go/accessapproval v1.5.0
	cloud.google.com/go/accesscontextmanager => cloud.google.com/go/accesscontextmanager v1.4.0
	cloud.google.com/go/aiplatform => cloud.google.com/go/aiplatform v1.24.0
	cloud.google.com/go/analytics => cloud.google.com/go/analytics v0.12.0
	cloud.google.com/go/apigateway => cloud.google.com/go/apigateway v1.4.0
	cloud.google.com/go/apigeeconnect => cloud.google.com/go/apigeeconnect v1.4.0
	cloud.google.com/go/appengine => cloud.google.com/go/appengine v1.5.0
	cloud.google.com/go/area120 => cloud.google.com/go/area120 v0.6.0
	cloud.google.com/go/artifactregistry => cloud.google.com/go/artifactregistry v1.9.0
	cloud.google.com/go/asset => cloud.google.com/go/asset v1.10.0
	cloud.google.com/go/assuredworkloads => cloud.google.com/go/assuredworkloads v1.9.0
	cloud.google.com/go/automl => cloud.google.com/go/automl v1.8.0
	cloud.google.com/go/baremetalsolution => cloud.google.com/go/baremetalsolution v0.4.0
	cloud.google.com/go/batch => cloud.google.com/go/batch v0.4.0
	cloud.google.com/go/beyondcorp => cloud.google.com/go/beyondcorp v0.3.0
	cloud.google.com/go/bigquery => cloud.google.com/go/bigquery v1.43.0
	cloud.google.com/go/billing => cloud.google.com/go/billing v1.7.0
	cloud.google.com/go/binaryauthorization => cloud.google.com/go/binaryauthorization v1.4.0
	cloud.google.com/go/certificatemanager => cloud.google.com/go/certificatemanager v1.4.0
	cloud.google.com/go/channel => cloud.google.com/go/channel v1.9.0
	cloud.google.com/go/cloudbuild => cloud.google.com/go/cloudbuild v1.4.0
	cloud.google.com/go/clouddms => cloud.google.com/go/clouddms v1.4.0
	cloud.google.com/go/cloudtasks => cloud.google.com/go/cloudtasks v1.8.0
	cloud.google.com/go/compute => cloud.google.com/go/compute v1.12.1
	cloud.google.com/go/compute/metadata => cloud.google.com/go/compute/metadata v0.2.1
	cloud.google.com/go/contactcenterinsights => cloud.google.com/go/contactcenterinsights v1.4.0
	cloud.google.com/go/container => cloud.google.com/go/container v1.7.0
	cloud.google.com/go/containeranalysis => cloud.google.com/go/containeranalysis v0.6.0
	cloud.google.com/go/datacatalog => cloud.google.com/go/datacatalog v1.8.0
	cloud.google.com/go/dataflow => cloud.google.com/go/dataflow v0.7.0
	cloud.google.com/go/dataform => cloud.google.com/go/dataform v0.5.0
	cloud.google.com/go/datafusion => cloud.google.com/go/datafusion v1.5.0
	cloud.google.com/go/datalabeling => cloud.google.com/go/datalabeling v0.6.0
	cloud.google.com/go/dataplex => cloud.google.com/go/dataplex v1.4.0
	cloud.google.com/go/dataproc => cloud.google.com/go/dataproc v1.8.0
	cloud.google.com/go/dataqna => cloud.google.com/go/dataqna v0.6.0
	cloud.google.com/go/datastream => cloud.google.com/go/datastream v1.5.0
	cloud.google.com/go/deploy => cloud.google.com/go/deploy v1.5.0
	cloud.google.com/go/dialogflow => cloud.google.com/go/dialogflow v1.19.0
	cloud.google.com/go/dlp => cloud.google.com/go/dlp v1.7.0
	cloud.google.com/go/documentai => cloud.google.com/go/documentai v1.10.0
	cloud.google.com/go/domains => cloud.google.com/go/domains v0.7.0
	cloud.google.com/go/edgecontainer => cloud.google.com/go/edgecontainer v0.2.0
	cloud.google.com/go/essentialcontacts => cloud.google.com/go/essentialcontacts v1.4.0
	cloud.google.com/go/eventarc => cloud.google.com/go/eventarc v1.8.0
	cloud.google.com/go/filestore => cloud.google.com/go/filestore v1.4.0
	cloud.google.com/go/firestore => cloud.google.com/go/firestore v1.1.0
	cloud.google.com/go/functions => cloud.google.com/go/functions v1.9.0
	cloud.google.com/go/gaming => cloud.google.com/go/gaming v1.8.0
	cloud.google.com/go/gkebackup => cloud.google.com/go/gkebackup v0.3.0
	cloud.google.com/go/gkeconnect => cloud.google.com/go/gkeconnect v0.6.0
	cloud.google.com/go/gkehub => cloud.google.com/go/gkehub v0.10.0
	cloud.google.com/go/gkemulticloud => cloud.google.com/go/gkemulticloud v0.4.0
	cloud.google.com/go/grafeas => cloud.google.com/go/grafeas v0.2.0
	cloud.google.com/go/gsuiteaddons => cloud.google.com/go/gsuiteaddons v1.4.0
	cloud.google.com/go/iam => cloud.google.com/go/iam v0.7.0
	cloud.google.com/go/iap => cloud.google.com/go/iap v1.5.0
	cloud.google.com/go/ids => cloud.google.com/go/ids v1.2.0
	cloud.google.com/go/iot => cloud.google.com/go/iot v1.4.0
	cloud.google.com/go/kms => cloud.google.com/go/kms v1.6.0
	cloud.google.com/go/language => cloud.google.com/go/language v1.8.0
	cloud.google.com/go/lifesciences => cloud.google.com/go/lifesciences v0.6.0
	cloud.google.com/go/longrunning => cloud.google.com/go/longrunning v0.3.0
	cloud.google.com/go/managedidentities => cloud.google.com/go/managedidentities v1.4.0
	cloud.google.com/go/mediatranslation => cloud.google.com/go/mediatranslation v0.6.0
	cloud.google.com/go/memcache => cloud.google.com/go/memcache v1.7.0
	cloud.google.com/go/metastore => cloud.google.com/go/metastore v1.8.0
	cloud.google.com/go/monitoring => cloud.google.com/go/monitoring v1.8.0
	cloud.google.com/go/networkconnectivity => cloud.google.com/go/networkconnectivity v1.7.0
	cloud.google.com/go/networkmanagement => cloud.google.com/go/networkmanagement v1.5.0
	cloud.google.com/go/networksecurity => cloud.google.com/go/networksecurity v0.6.0
	cloud.google.com/go/notebooks => cloud.google.com/go/notebooks v1.5.0
	cloud.google.com/go/optimization => cloud.google.com/go/optimization v1.2.0
	cloud.google.com/go/orchestration => cloud.google.com/go/orchestration v1.4.0
	cloud.google.com/go/orgpolicy => cloud.google.com/go/orgpolicy v1.5.0
	cloud.google.com/go/osconfig => cloud.google.com/go/osconfig v1.10.0
	cloud.google.com/go/oslogin => cloud.google.com/go/oslogin v1.7.0
	cloud.google.com/go/phishingprotection => cloud.google.com/go/phishingprotection v0.6.0
	cloud.google.com/go/policytroubleshooter => cloud.google.com/go/policytroubleshooter v1.4.0
	cloud.google.com/go/privatecatalog => cloud.google.com/go/privatecatalog v0.6.0
	cloud.google.com/go/recaptchaenterprise => cloud.google.com/go/recaptchaenterprise v1.3.1
	cloud.google.com/go/recaptchaenterprise/v2 => cloud.google.com/go/recaptchaenterprise/v2 v2.5.0
	cloud.google.com/go/recommendationengine => cloud.google.com/go/recommendationengine v0.6.0
	cloud.google.com/go/recommender => cloud.google.com/go/recommender v1.8.0
	cloud.google.com/go/redis => cloud.google.com/go/redis v1.10.0
	cloud.google.com/go/resourcemanager => cloud.google.com/go/resourcemanager v1.4.0
	cloud.google.com/go/resourcesettings => cloud.google.com/go/resourcesettings v1.4.0
	cloud.google.com/go/retail => cloud.google.com/go/retail v1.11.0
	cloud.google.com/go/run => cloud.google.com/go/run v0.3.0
	cloud.google.com/go/scheduler => cloud.google.com/go/scheduler v1.7.0
	cloud.google.com/go/secretmanager => cloud.google.com/go/secretmanager v1.9.0
	cloud.google.com/go/security => cloud.google.com/go/security v1.10.0
	cloud.google.com/go/securitycenter => cloud.google.com/go/securitycenter v1.16.0
	cloud.google.com/go/servicecontrol => cloud.google.com/go/servicecontrol v1.5.0
	cloud.google.com/go/servicedirectory => cloud.google.com/go/servicedirectory v1.7.0
	cloud.google.com/go/servicemanagement => cloud.google.com/go/servicemanagement v1.5.0
	cloud.google.com/go/serviceusage => cloud.google.com/go/serviceusage v1.4.0
	cloud.google.com/go/shell => cloud.google.com/go/shell v1.4.0
	cloud.google.com/go/speech => cloud.google.com/go/speech v1.9.0
	cloud.google.com/go/storage => cloud.google.com/go/storage v1.27.0
	cloud.google.com/go/storagetransfer => cloud.google.com/go/storagetransfer v1.6.0
	cloud.google.com/go/talent => cloud.google.com/go/talent v1.4.0
	cloud.google.com/go/texttospeech => cloud.google.com/go/texttospeech v1.5.0
	cloud.google.com/go/tpu => cloud.google.com/go/tpu v1.4.0
	cloud.google.com/go/trace => cloud.google.com/go/trace v1.4.0
	cloud.google.com/go/translate => cloud.google.com/go/translate v1.4.0
	cloud.google.com/go/video => cloud.google.com/go/video v1.9.0
	cloud.google.com/go/videointelligence => cloud.google.com/go/videointelligence v1.9.0
	cloud.google.com/go/vision => cloud.google.com/go/vision v1.2.0
	cloud.google.com/go/vision/v2 => cloud.google.com/go/vision/v2 v2.5.0
	cloud.google.com/go/vmmigration => cloud.google.com/go/vmmigration v1.3.0
	cloud.google.com/go/vpcaccess => cloud.google.com/go/vpcaccess v1.5.0
	cloud.google.com/go/webrisk => cloud.google.com/go/webrisk v1.7.0
	cloud.google.com/go/websecurityscanner => cloud.google.com/go/websecurityscanner v1.4.0
	cloud.google.com/go/workflows => cloud.google.com/go/workflows v1.9.0
	code.cloudfoundry.org/bytefmt => code.cloudfoundry.org/bytefmt v0.0.0-20190710193110-1eb035ffe2b6
	github.com/AdaLogics/go-fuzz-headers => github.com/AdaLogics/go-fuzz-headers v0.0.0-20210715213245-6c3934b029d8
	github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v45.1.0+incompatible
	github.com/Azure/go-ansiterm => github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v14.2.0+incompatible
	github.com/Azure/go-autorest/autorest => github.com/Azure/go-autorest/autorest v0.11.27
	github.com/Azure/go-autorest/autorest/adal => github.com/Azure/go-autorest/autorest/adal v0.9.20
	github.com/Azure/go-autorest/autorest/date => github.com/Azure/go-autorest/autorest/date v0.3.0
	github.com/Azure/go-autorest/autorest/mocks => github.com/Azure/go-autorest/autorest/mocks v0.4.2
	github.com/Azure/go-autorest/autorest/to => github.com/Azure/go-autorest/autorest/to v0.3.0
	github.com/Azure/go-autorest/autorest/validation => github.com/Azure/go-autorest/autorest/validation v0.2.0
	github.com/Azure/go-autorest/logger => github.com/Azure/go-autorest/logger v0.2.1
	github.com/Azure/go-autorest/tracing => github.com/Azure/go-autorest/tracing v0.6.0
	github.com/BurntSushi/toml => github.com/BurntSushi/toml v1.1.0
	github.com/DATA-DOG/go-sqlmock => github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/DataDog/datadog-go => github.com/DataDog/datadog-go v3.2.0+incompatible
	github.com/Knetic/govaluate => github.com/Knetic/govaluate v3.0.1-0.20171022003610-9aa49832a739+incompatible
	github.com/MakeNowJust/heredoc => github.com/MakeNowJust/heredoc v1.0.0
	github.com/Masterminds/goutils => github.com/Masterminds/goutils v1.1.1
	github.com/Masterminds/semver => github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/semver/v3 => github.com/Masterminds/semver/v3 v3.1.1
	github.com/Masterminds/sprig => github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/Masterminds/sprig/v3 => github.com/Masterminds/sprig/v3 v3.2.2
	github.com/Masterminds/squirrel => github.com/Masterminds/squirrel v1.5.3
	github.com/Masterminds/vcs => github.com/Masterminds/vcs v1.13.3
	github.com/Microsoft/go-winio => github.com/Microsoft/go-winio v0.5.2
	github.com/Microsoft/hcsshim => github.com/Microsoft/hcsshim v0.9.6
	github.com/NYTimes/gziphandler => github.com/NYTimes/gziphandler v1.1.1
	github.com/OneOfOne/xxhash => github.com/OneOfOne/xxhash v1.2.8
	github.com/PuerkitoBio/goquery => github.com/PuerkitoBio/goquery v1.5.0
	github.com/PuerkitoBio/purell => github.com/PuerkitoBio/purell v1.1.1
	github.com/PuerkitoBio/urlesc => github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578
	github.com/Shopify/logrus-bugsnag => github.com/Shopify/logrus-bugsnag v0.0.0-20171204204709-577dee27f20d
	github.com/Shopify/sarama => github.com/Shopify/sarama v1.19.0
	github.com/Shopify/toxiproxy => github.com/Shopify/toxiproxy v2.1.4+incompatible
	github.com/StackExchange/wmi => github.com/StackExchange/wmi v1.2.1
	github.com/VividCortex/gohistogram => github.com/VividCortex/gohistogram v1.0.0
	github.com/afex/hystrix-go => github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/agnivade/levenshtein => github.com/agnivade/levenshtein v1.1.1
	github.com/ajstarks/svgo => github.com/ajstarks/svgo v0.0.0-20180226025133-644b8db467af
	github.com/alcortesm/tgz => github.com/alcortesm/tgz v0.0.0-20161220082320-9c5fe88206d7
	github.com/alecthomas/units => github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d
	github.com/alessio/shellescape => github.com/alessio/shellescape v1.2.2
	github.com/alexflint/go-filemutex => github.com/alexflint/go-filemutex v1.1.0
	github.com/andreyvit/diff => github.com/andreyvit/diff v0.0.0-20170406064948-c7f18ee00883
	github.com/andybalholm/cascadia => github.com/andybalholm/cascadia v1.0.0
	github.com/anmitsu/go-shlex => github.com/anmitsu/go-shlex v0.0.0-20161002113705-648efa622239
	github.com/antihax/optional => github.com/antihax/optional v1.0.0
	github.com/antlr/antlr4/runtime/Go/antlr => github.com/antlr/antlr4/runtime/Go/antlr v1.4.10
	github.com/apache/thrift => github.com/apache/thrift v0.13.0
	github.com/arbovm/levenshtein => github.com/arbovm/levenshtein v0.0.0-20160628152529-48b4e1c0c4d0
	github.com/armon/circbuf => github.com/armon/circbuf v0.0.0-20150827004946-bbbad097214e
	github.com/armon/consul-api => github.com/armon/consul-api v0.0.0-20180202201655-eb2c6b5be1b6
	github.com/armon/go-metrics => github.com/armon/go-metrics v0.3.3
	github.com/armon/go-radix => github.com/armon/go-radix v1.0.0
	github.com/armon/go-socks5 => github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5
	github.com/aryann/difflib => github.com/aryann/difflib v0.0.0-20170710044230-e206f873d14a
	github.com/asaskevich/govalidator => github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535
	github.com/aws/aws-lambda-go => github.com/aws/aws-lambda-go v1.13.3
	github.com/aws/aws-sdk-go => github.com/aws/aws-sdk-go v1.43.16
	github.com/aws/aws-sdk-go-v2 => github.com/aws/aws-sdk-go-v2 v0.18.0
	github.com/beevik/etree => github.com/beevik/etree v1.1.0
	github.com/benbjohnson/clock => github.com/benbjohnson/clock v1.1.0
	github.com/beorn7/perks => github.com/beorn7/perks v1.0.1
	github.com/bgentry/speakeasy => github.com/bgentry/speakeasy v0.1.0
	github.com/bketelsen/crypt => github.com/bketelsen/crypt v0.0.4
	github.com/blang/semver => github.com/blang/semver v3.5.1+incompatible
	github.com/blang/semver/v4 => github.com/blang/semver/v4 v4.0.0
	github.com/bshuster-repo/logrus-logstash-hook => github.com/bshuster-repo/logrus-logstash-hook v1.0.0
	github.com/buger/jsonparser => github.com/buger/jsonparser v1.1.1
	github.com/bugsnag/bugsnag-go => github.com/bugsnag/bugsnag-go v1.5.3
	github.com/bugsnag/osext => github.com/bugsnag/osext v0.0.0-20130617224835-0dd3f918b21b
	github.com/bugsnag/panicwrap => github.com/bugsnag/panicwrap v1.3.4
	github.com/bytecodealliance/wasmtime-go/v3 => github.com/bytecodealliance/wasmtime-go/v3 v3.0.2
	github.com/casbin/casbin/v2 => github.com/casbin/casbin/v2 v2.1.2
	github.com/cenkalti/backoff => github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cenkalti/backoff/v4 => github.com/cenkalti/backoff/v4 v4.1.3
	github.com/census-instrumentation/opencensus-proto => github.com/census-instrumentation/opencensus-proto v0.2.1
	github.com/certifi/gocertifi => github.com/certifi/gocertifi v0.0.0-20200922220541-2c3bb06c6054
	github.com/cespare/xxhash => github.com/cespare/xxhash v1.1.0
	github.com/cespare/xxhash/v2 => github.com/cespare/xxhash/v2 v2.1.2
	github.com/chai2010/gettext-go => github.com/chai2010/gettext-go v1.0.2
	github.com/checkpoint-restore/go-criu/v5 => github.com/checkpoint-restore/go-criu/v5 v5.3.0
	github.com/chromedp/cdproto => github.com/chromedp/cdproto v0.0.0-20210122124816-7a656c010d57
	github.com/chromedp/chromedp => github.com/chromedp/chromedp v0.6.5
	github.com/chromedp/sysutil => github.com/chromedp/sysutil v1.0.0
	github.com/chzyer/logex => github.com/chzyer/logex v1.1.10
	github.com/chzyer/readline => github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e
	github.com/chzyer/test => github.com/chzyer/test v0.0.0-20180213035817-a1ea475d72b1
	github.com/cilium/ebpf => github.com/cilium/ebpf v0.7.0
	github.com/circonus-labs/circonus-gometrics => github.com/circonus-labs/circonus-gometrics v2.3.1+incompatible
	github.com/circonus-labs/circonusllhist => github.com/circonus-labs/circonusllhist v0.1.3
	github.com/clbanning/x2j => github.com/clbanning/x2j v0.0.0-20191024224557-825249438eec
	github.com/cloudflare/cfssl => github.com/cloudflare/cfssl v1.5.0
	github.com/cncf/udpa/go => github.com/cncf/udpa/go v0.0.0-20210930031921-04548b0d99d4
	github.com/cncf/xds/go => github.com/cncf/xds/go v0.0.0-20211011173535-cb28da3451f1
	github.com/cockroachdb/datadriven => github.com/cockroachdb/datadriven v0.0.0-20200714090401-bf6692d28da5
	github.com/cockroachdb/errors => github.com/cockroachdb/errors v1.2.4
	github.com/cockroachdb/logtags => github.com/cockroachdb/logtags v0.0.0-20190617123548-eb05cc24525f
	github.com/codahale/hdrhistogram => github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd
	github.com/containerd/aufs => github.com/containerd/aufs v1.0.0
	github.com/containerd/btrfs => github.com/containerd/btrfs v1.0.0
	github.com/containerd/cgroups => github.com/containerd/cgroups v1.0.4
	github.com/containerd/console => github.com/containerd/console v1.0.3
	github.com/containerd/containerd => github.com/containerd/containerd v1.6.16
	github.com/containerd/continuity => github.com/containerd/continuity v0.3.0
	github.com/containerd/fifo => github.com/containerd/fifo v1.0.0
	github.com/containerd/go-cni => github.com/containerd/go-cni v1.1.6
	github.com/containerd/go-runc => github.com/containerd/go-runc v1.0.0
	github.com/containerd/imgcrypt => github.com/containerd/imgcrypt v1.1.4
	github.com/containerd/nri => github.com/containerd/nri v0.1.0
	github.com/containerd/stargz-snapshotter/estargz => github.com/containerd/stargz-snapshotter/estargz v0.4.1
	github.com/containerd/ttrpc => github.com/containerd/ttrpc v1.1.0
	github.com/containerd/typeurl => github.com/containerd/typeurl v1.0.2
	github.com/containerd/zfs => github.com/containerd/zfs v1.0.0
	github.com/containernetworking/cni => github.com/containernetworking/cni v1.1.2
	github.com/containernetworking/plugins => github.com/containernetworking/plugins v1.1.1
	github.com/containers/ocicrypt => github.com/containers/ocicrypt v1.1.3
	github.com/coreos/bbolt => github.com/coreos/bbolt v1.3.3
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.17+incompatible
	github.com/coreos/go-etcd => github.com/coreos/go-etcd v2.0.0+incompatible
	github.com/coreos/go-iptables => github.com/coreos/go-iptables v0.6.0
	github.com/coreos/go-oidc => github.com/coreos/go-oidc v2.1.0+incompatible
	github.com/coreos/go-semver => github.com/coreos/go-semver v0.3.0
	github.com/coreos/go-systemd => github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf
	github.com/coreos/go-systemd/v22 => github.com/coreos/go-systemd/v22 v22.3.2
	github.com/coreos/pkg => github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f
	github.com/cpuguy83/go-md2man => github.com/cpuguy83/go-md2man v1.0.10
	github.com/cpuguy83/go-md2man/v2 => github.com/cpuguy83/go-md2man/v2 v2.0.2
	github.com/creack/pty => github.com/creack/pty v1.1.18
	github.com/cyphar/filepath-securejoin => github.com/cyphar/filepath-securejoin v0.2.3
	github.com/d2g/dhcp4 => github.com/d2g/dhcp4 v0.0.0-20170904100407-a1d1b6c41b1c
	github.com/d2g/dhcp4client => github.com/d2g/dhcp4client v1.0.0
	github.com/d2g/dhcp4server => github.com/d2g/dhcp4server v0.0.0-20181031114812-7d4a0a7f59a5
	github.com/danieljoos/wincred => github.com/danieljoos/wincred v1.1.2
	github.com/davecgh/go-spew => github.com/davecgh/go-spew v1.1.1
	github.com/daviddengcn/go-colortext => github.com/daviddengcn/go-colortext v1.0.0
	github.com/deckarep/golang-set => github.com/deckarep/golang-set v1.7.1
	github.com/denisenkom/go-mssqldb => github.com/denisenkom/go-mssqldb v0.9.0
	github.com/dgraph-io/badger/v3 => github.com/dgraph-io/badger/v3 v3.2103.5
	github.com/dgraph-io/ristretto => github.com/dgraph-io/ristretto v0.1.1
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dgryski/go-farm => github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13
	github.com/dgryski/go-sip13 => github.com/dgryski/go-sip13 v0.0.0-20190329191031-25c5027a8c7b
	github.com/dgryski/trifles => github.com/dgryski/trifles v0.0.0-20200323201526-dd97f9abfb48
	github.com/digitalocean/godo => github.com/digitalocean/godo v1.42.1
	github.com/distribution/distribution/v3 => github.com/distribution/distribution/v3 v3.0.0-20221208165359-362910506bc2
	github.com/docker/cli => github.com/docker/cli v20.10.22+incompatible
	github.com/docker/distribution => github.com/docker/distribution v2.8.1+incompatible
	github.com/docker/docker => github.com/docker/docker v20.10.21+incompatible
	github.com/docker/docker-credential-helpers => github.com/docker/docker-credential-helpers v0.7.0
	github.com/docker/go-connections => github.com/docker/go-connections v0.4.0
	github.com/docker/go-events => github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c
	github.com/docker/go-metrics => github.com/docker/go-metrics v0.0.1
	github.com/docker/go-units => github.com/docker/go-units v0.4.0
	github.com/docker/libtrust => github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7
	github.com/docker/spdystream => github.com/docker/spdystream v0.0.0-20160310174837-449fdfce4d96
	github.com/docopt/docopt-go => github.com/docopt/docopt-go v0.0.0-20180111231733-ee0de3bc6815
	github.com/dustin/go-humanize => github.com/dustin/go-humanize v1.0.0
	github.com/eapache/go-resiliency => github.com/eapache/go-resiliency v1.1.0
	github.com/eapache/go-xerial-snappy => github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21
	github.com/eapache/queue => github.com/eapache/queue v1.1.0
	github.com/edsrzf/mmap-go => github.com/edsrzf/mmap-go v1.0.0
	github.com/elastic/go-elasticsearch/v5 => github.com/elastic/go-elasticsearch/v5 v5.6.1
	github.com/elastic/go-elasticsearch/v6 => github.com/elastic/go-elasticsearch/v6 v6.8.2
	github.com/elastic/go-elasticsearch/v7 => github.com/elastic/go-elasticsearch/v7 v7.3.0
	github.com/elazarl/goproxy => github.com/elazarl/goproxy v0.0.0-20200315184450-1f3cb6622dad
	github.com/elazarl/goproxy/ext => github.com/elazarl/goproxy/ext v0.0.0-20190711103511-473e67f1d7d2
	github.com/emicklei/go-restful => github.com/emicklei/go-restful v2.16.0+incompatible
	github.com/emicklei/go-restful-openapi => github.com/emicklei/go-restful-openapi v1.4.1
	github.com/emicklei/go-restful/v3 => github.com/emicklei/go-restful/v3 v3.9.0
	github.com/emirpasic/gods => github.com/emirpasic/gods v1.12.0
	github.com/envoyproxy/go-control-plane => github.com/envoyproxy/go-control-plane v0.10.2-0.20220325020618-49ff273808a1
	github.com/envoyproxy/protoc-gen-validate => github.com/envoyproxy/protoc-gen-validate v0.1.0
	github.com/evanphx/json-patch => github.com/evanphx/json-patch v5.6.0+incompatible
	github.com/evanphx/json-patch/v5 => github.com/evanphx/json-patch/v5 v5.6.0
	github.com/exponent-io/jsonpath => github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d
	github.com/fatih/camelcase => github.com/fatih/camelcase v1.0.0
	github.com/fatih/color => github.com/fatih/color v1.13.0
	github.com/fatih/structs => github.com/fatih/structs v1.1.0
	github.com/felixge/httpsnoop => github.com/felixge/httpsnoop v1.0.3
	github.com/flynn/go-shlex => github.com/flynn/go-shlex v0.0.0-20150515145356-3f9db97f8568
	github.com/fogleman/gg => github.com/fogleman/gg v1.2.1-0.20190220221249-0403632d5b90
	github.com/form3tech-oss/jwt-go => github.com/form3tech-oss/jwt-go v3.2.3+incompatible
	github.com/fortytw2/leaktest => github.com/fortytw2/leaktest v1.3.0
	github.com/foxcpp/go-mockdns => github.com/foxcpp/go-mockdns v0.0.0-20210729171921-fb145fc6f897
	github.com/franela/goblin => github.com/franela/goblin v0.0.0-20200105215937-c9ffbefa60db
	github.com/franela/goreq => github.com/franela/goreq v0.0.0-20171204163338-bcd34c9993f8
	github.com/frankban/quicktest => github.com/frankban/quicktest v1.11.3
	github.com/fsnotify/fsnotify => github.com/fsnotify/fsnotify v1.6.0
	github.com/fvbommel/sortorder => github.com/fvbommel/sortorder v1.0.1
	github.com/getsentry/raven-go => github.com/getsentry/raven-go v0.2.0
	github.com/ghodss/yaml => github.com/ghodss/yaml v1.0.0
	github.com/gliderlabs/ssh => github.com/gliderlabs/ssh v0.2.2
	github.com/go-errors/errors => github.com/go-errors/errors v1.0.1
	github.com/go-gorp/gorp/v3 => github.com/go-gorp/gorp/v3 v3.0.2
	github.com/go-ini/ini => github.com/go-ini/ini v1.67.0
	github.com/go-kit/kit => github.com/go-kit/kit v0.10.0
	github.com/go-kit/log => github.com/go-kit/log v0.2.0
	github.com/go-ldap/ldap => github.com/go-ldap/ldap v3.0.3+incompatible
	github.com/go-logfmt/logfmt => github.com/go-logfmt/logfmt v0.5.1
	github.com/go-logr/logr => github.com/go-logr/logr v1.2.3
	github.com/go-logr/stdr => github.com/go-logr/stdr v1.2.2
	github.com/go-logr/zapr => github.com/go-logr/zapr v1.2.3
	github.com/go-ole/go-ole => github.com/go-ole/go-ole v1.2.6-0.20210915003542-8b1f7f90f6b1
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
	github.com/go-playground/locales => github.com/go-playground/locales v0.12.1
	github.com/go-playground/universal-translator => github.com/go-playground/universal-translator v0.0.0-20170327191703-71201497bace
	github.com/go-redis/redis => github.com/go-redis/redis v6.15.2+incompatible
	github.com/go-resty/resty/v2 => github.com/go-resty/resty/v2 v2.5.0
	github.com/go-sql-driver/mysql => github.com/go-sql-driver/mysql v1.6.0
	github.com/go-stack/stack => github.com/go-stack/stack v1.8.0
	github.com/go-task/slim-sprig => github.com/go-task/slim-sprig v0.0.0-20210107165309-348f09dbbbc0
	github.com/gobuffalo/flect => github.com/gobuffalo/flect v0.3.0
	github.com/gobuffalo/logger => github.com/gobuffalo/logger v1.0.6
	github.com/gobuffalo/packd => github.com/gobuffalo/packd v1.0.1
	github.com/gobuffalo/packr/v2 => github.com/gobuffalo/packr/v2 v2.8.3
	github.com/gobwas/glob => github.com/gobwas/glob v0.2.3
	github.com/gobwas/httphead => github.com/gobwas/httphead v0.1.0
	github.com/gobwas/pool => github.com/gobwas/pool v0.2.1
	github.com/gobwas/ws => github.com/gobwas/ws v1.0.4
	github.com/godbus/dbus/v5 => github.com/godbus/dbus/v5 v5.0.6
	github.com/godror/godror => github.com/godror/godror v0.24.2
	github.com/gofrs/flock => github.com/gofrs/flock v0.8.1
	github.com/gofrs/uuid => github.com/gofrs/uuid v4.0.0+incompatible
	github.com/gogo/googleapis => github.com/gogo/googleapis v1.4.0
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
	github.com/golang-jwt/jwt/v4 => github.com/golang-jwt/jwt/v4 v4.2.0
	github.com/golang-sql/civil => github.com/golang-sql/civil v0.0.0-20190719163853-cb61b32ac6fe
	github.com/golang/example => github.com/golang/example v0.0.0-20170904185048-46695d81d1fa
	github.com/golang/freetype => github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/golang/glog => github.com/golang/glog v1.0.0
	github.com/golang/groupcache => github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da
	github.com/golang/mock => github.com/golang/mock v1.6.0
	github.com/golang/protobuf => github.com/golang/protobuf v1.5.2
	github.com/golang/snappy => github.com/golang/snappy v0.0.4
	github.com/gomodule/redigo => github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/addlicense => github.com/google/addlicense v0.0.0-20200906110928-a0294312aa76
	github.com/google/btree => github.com/google/btree v1.0.1
	github.com/google/cel-go => github.com/google/cel-go v0.12.6
	github.com/google/certificate-transparency-go => github.com/google/certificate-transparency-go v1.0.21
	github.com/google/flatbuffers => github.com/google/flatbuffers v1.12.1
	github.com/google/gnostic => github.com/google/gnostic v0.5.7-v3refs
	github.com/google/go-cmp => github.com/google/go-cmp v0.5.9
	github.com/google/go-containerregistry => github.com/google/go-containerregistry v0.5.1
	github.com/google/go-querystring => github.com/google/go-querystring v1.0.0
	github.com/google/gofuzz => github.com/google/gofuzz v1.2.0
	github.com/google/gops => github.com/google/gops v0.3.23
	github.com/google/martian/v3 => github.com/google/martian/v3 v3.2.1
	github.com/google/pprof => github.com/google/pprof v0.0.0-20210407192527-94a9f03dee38
	github.com/google/shlex => github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/google/uuid => github.com/google/uuid v1.3.0
	github.com/googleapis/enterprise-certificate-proxy => github.com/googleapis/enterprise-certificate-proxy v0.2.0
	github.com/googleapis/gax-go/v2 => github.com/googleapis/gax-go/v2 v2.6.0
	github.com/gophercloud/gophercloud => github.com/gophercloud/gophercloud v0.12.0
	github.com/gopherjs/gopherjs => github.com/gopherjs/gopherjs v0.0.0-20191106031601-ce3c9ade29de
	github.com/gorilla/context => github.com/gorilla/context v1.1.1
	github.com/gorilla/handlers => github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux => github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket => github.com/gorilla/websocket v1.4.2
	github.com/gosimple/slug => github.com/gosimple/slug v1.1.1
	github.com/gosuri/uitable => github.com/gosuri/uitable v0.0.4
	github.com/grafana-tools/sdk => github.com/grafana-tools/sdk v0.0.0-20210625151406-43693eb2f02b
	github.com/gregjones/httpcache => github.com/gregjones/httpcache v0.0.0-20181110185634-c63ab54fda8f
	github.com/grpc-ecosystem/go-grpc-middleware => github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/go-grpc-prometheus => github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway => github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/grpc-ecosystem/grpc-gateway/v2 => github.com/grpc-ecosystem/grpc-gateway/v2 v2.7.0
	github.com/hashicorp/consul/api => github.com/hashicorp/consul/api v1.6.0
	github.com/hashicorp/consul/sdk => github.com/hashicorp/consul/sdk v0.6.0
	github.com/hashicorp/errwrap => github.com/hashicorp/errwrap v1.1.0
	github.com/hashicorp/go-cleanhttp => github.com/hashicorp/go-cleanhttp v0.5.1
	github.com/hashicorp/go-hclog => github.com/hashicorp/go-hclog v0.12.2
	github.com/hashicorp/go-immutable-radix => github.com/hashicorp/go-immutable-radix v1.2.0
	github.com/hashicorp/go-msgpack => github.com/hashicorp/go-msgpack v0.5.3
	github.com/hashicorp/go-multierror => github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-retryablehttp => github.com/hashicorp/go-retryablehttp v0.5.3
	github.com/hashicorp/go-rootcerts => github.com/hashicorp/go-rootcerts v1.0.2
	github.com/hashicorp/go-sockaddr => github.com/hashicorp/go-sockaddr v1.0.2
	github.com/hashicorp/go-syslog => github.com/hashicorp/go-syslog v1.0.0
	github.com/hashicorp/go-uuid => github.com/hashicorp/go-uuid v1.0.1
	github.com/hashicorp/go-version => github.com/hashicorp/go-version v1.2.0
	github.com/hashicorp/golang-lru => github.com/hashicorp/golang-lru v0.5.4
	github.com/hashicorp/hcl => github.com/hashicorp/hcl v1.0.0
	github.com/hashicorp/logutils => github.com/hashicorp/logutils v1.0.0
	github.com/hashicorp/mdns => github.com/hashicorp/mdns v1.0.1
	github.com/hashicorp/memberlist => github.com/hashicorp/memberlist v0.2.2
	github.com/hashicorp/serf => github.com/hashicorp/serf v0.9.3
	github.com/hetznercloud/hcloud-go => github.com/hetznercloud/hcloud-go v1.21.1
	github.com/huandu/xstrings => github.com/huandu/xstrings v1.3.2
	github.com/hudl/fargo => github.com/hudl/fargo v1.3.0
	github.com/iancoleman/strcase => github.com/iancoleman/strcase v0.2.0
	github.com/ianlancetaylor/demangle => github.com/ianlancetaylor/demangle v0.0.0-20200824232613-28f6c0f3b639
	github.com/imdario/mergo => github.com/imdario/mergo v0.3.12
	github.com/inconshreveable/mousetrap => github.com/inconshreveable/mousetrap v1.0.1
	github.com/influxdata/influxdb1-client => github.com/influxdata/influxdb1-client v0.0.0-20191209144304-8bf82d3c094d
	github.com/intel/goresctrl => github.com/intel/goresctrl v0.2.0
	github.com/jbenet/go-context => github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99
	github.com/jessevdk/go-flags => github.com/jessevdk/go-flags v1.4.0
	github.com/jmespath/go-jmespath => github.com/jmespath/go-jmespath v0.4.0
	github.com/jmespath/go-jmespath/internal/testify => github.com/jmespath/go-jmespath/internal/testify v1.5.1
	github.com/jmoiron/sqlx => github.com/jmoiron/sqlx v1.3.5
	github.com/joefitzgerald/rainbow-reporter => github.com/joefitzgerald/rainbow-reporter v0.1.0
	github.com/jonboulle/clockwork => github.com/jonboulle/clockwork v0.2.2
	github.com/josharian/intern => github.com/josharian/intern v1.0.0
	github.com/jpillora/backoff => github.com/jpillora/backoff v1.0.0
	github.com/json-iterator/go => github.com/json-iterator/go v1.1.12
	github.com/jszwec/csvutil => github.com/jszwec/csvutil v1.5.0
	github.com/jtolds/gls => github.com/jtolds/gls v4.20.0+incompatible
	github.com/julienschmidt/httprouter => github.com/julienschmidt/httprouter v1.3.0
	github.com/jung-kurt/gofpdf => github.com/jung-kurt/gofpdf v1.0.3-0.20190309125859-24315acbbda5
	github.com/kardianos/osext => github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/karrick/godirwalk => github.com/karrick/godirwalk v1.16.1
	github.com/kelseyhightower/envconfig => github.com/kelseyhightower/envconfig v1.4.0
	github.com/kevinburke/ssh_config => github.com/kevinburke/ssh_config v0.0.0-20190725054713-01f96b0aa0cd
	github.com/keybase/go-ps => github.com/keybase/go-ps v0.0.0-20190827175125-91aafc93ba19
	github.com/kisielk/errcheck => github.com/kisielk/errcheck v1.5.0
	github.com/kisielk/gotool => github.com/kisielk/gotool v1.0.0
	github.com/klauspost/compress => github.com/klauspost/compress v1.13.6
	github.com/kortschak/utter => github.com/kortschak/utter v1.0.1
	github.com/kr/fs => github.com/kr/fs v0.1.0
	github.com/kr/pretty => github.com/kr/pretty v0.2.1
	github.com/kr/pty => github.com/kr/pty v1.1.8
	github.com/kr/text => github.com/kr/text v0.2.0
	github.com/kubernetes-csi/external-snapshotter/client/v4 => github.com/kubernetes-csi/external-snapshotter/client/v4 v4.2.0
	github.com/kubesphere/pvc-autoresizer => github.com/kubesphere/pvc-autoresizer v0.3.0
	github.com/kubesphere/sonargo => github.com/kubesphere/sonargo v0.0.2
	github.com/kubesphere/storageclass-accessor => github.com/kubesphere/storageclass-accessor v0.2.2
	github.com/kylelemons/godebug => github.com/kylelemons/godebug v1.1.0
	github.com/lann/builder => github.com/lann/builder v0.0.0-20180802200727-47ae307949d0
	github.com/lann/ps => github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0
	github.com/leodido/go-urn => github.com/leodido/go-urn v0.0.0-20181204092800-a67a23e1c1af
	github.com/lib/pq => github.com/lib/pq v1.10.6
	github.com/liggitt/tabwriter => github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de
	github.com/lightstep/lightstep-tracer-common/golang/gogo => github.com/lightstep/lightstep-tracer-common/golang/gogo v0.0.0-20190605223551-bc2310a04743
	github.com/lightstep/lightstep-tracer-go => github.com/lightstep/lightstep-tracer-go v0.18.1
	github.com/linuxkit/virtsock => github.com/linuxkit/virtsock v0.0.0-20201010232012-f8cee7dfc7a3
	github.com/lithammer/dedent => github.com/lithammer/dedent v1.1.0
	github.com/magiconair/properties => github.com/magiconair/properties v1.8.5
	github.com/mailru/easyjson => github.com/mailru/easyjson v0.7.7
	github.com/markbates/errx => github.com/markbates/errx v1.1.0
	github.com/markbates/oncer => github.com/markbates/oncer v1.0.0
	github.com/markbates/safe => github.com/markbates/safe v1.0.1
	github.com/mattn/go-colorable => github.com/mattn/go-colorable v0.1.12
	github.com/mattn/go-isatty => github.com/mattn/go-isatty v0.0.14
	github.com/mattn/go-oci8 => github.com/mattn/go-oci8 v0.1.1
	github.com/mattn/go-runewidth => github.com/mattn/go-runewidth v0.0.9
	github.com/mattn/go-shellwords => github.com/mattn/go-shellwords v1.0.12
	github.com/mattn/go-sqlite3 => github.com/mattn/go-sqlite3 v1.14.6
	github.com/matttproud/golang_protobuf_extensions => github.com/matttproud/golang_protobuf_extensions v1.0.4
	github.com/maxbrunsfeld/counterfeiter/v6 => github.com/maxbrunsfeld/counterfeiter/v6 v6.2.2
	github.com/miekg/dns => github.com/miekg/dns v1.1.43
	github.com/miekg/pkcs11 => github.com/miekg/pkcs11 v1.1.1
	github.com/mistifyio/go-zfs => github.com/mistifyio/go-zfs v2.1.2-0.20190413222219-f784269be439+incompatible
	github.com/mitchellh/cli => github.com/mitchellh/cli v1.1.2
	github.com/mitchellh/copystructure => github.com/mitchellh/copystructure v1.2.0
	github.com/mitchellh/go-homedir => github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-testing-interface => github.com/mitchellh/go-testing-interface v1.0.0
	github.com/mitchellh/go-wordwrap => github.com/mitchellh/go-wordwrap v1.0.0
	github.com/mitchellh/mapstructure => github.com/mitchellh/mapstructure v1.4.3
	github.com/mitchellh/reflectwalk => github.com/mitchellh/reflectwalk v1.0.2
	github.com/moby/locker => github.com/moby/locker v1.0.1
	github.com/moby/spdystream => github.com/moby/spdystream v0.2.0
	github.com/moby/sys/mountinfo => github.com/moby/sys/mountinfo v0.5.0
	github.com/moby/sys/signal => github.com/moby/sys/signal v0.6.0
	github.com/moby/sys/symlink => github.com/moby/sys/symlink v0.2.0
	github.com/moby/term => github.com/moby/term v0.0.0-20221205130635-1aeaba878587
	github.com/modern-go/concurrent => github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd
	github.com/modern-go/reflect2 => github.com/modern-go/reflect2 v1.0.2
	github.com/monochromegane/go-gitignore => github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00
	github.com/montanaflynn/stats => github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe
	github.com/morikuni/aec => github.com/morikuni/aec v1.0.0
	github.com/mrunalp/fileutils => github.com/mrunalp/fileutils v0.5.0
	github.com/munnerz/goautoneg => github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822
	github.com/mwitkow/go-conntrack => github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f
	github.com/mxk/go-flowrate => github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f
	github.com/nats-io/jwt => github.com/nats-io/jwt v0.3.2
	github.com/nats-io/nats-server/v2 => github.com/nats-io/nats-server/v2 v2.1.2
	github.com/nats-io/nats.go => github.com/nats-io/nats.go v1.9.1
	github.com/nats-io/nkeys => github.com/nats-io/nkeys v0.1.3
	github.com/nats-io/nuid => github.com/nats-io/nuid v1.0.1
	github.com/networkplumbing/go-nft => github.com/networkplumbing/go-nft v0.2.0
	github.com/niemeyer/pretty => github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e
	github.com/nxadm/tail => github.com/nxadm/tail v1.4.8
	github.com/oklog/oklog => github.com/oklog/oklog v0.3.2
	github.com/oklog/run => github.com/oklog/run v1.1.0
	github.com/oklog/ulid => github.com/oklog/ulid v1.3.1
	github.com/olekukonko/tablewriter => github.com/olekukonko/tablewriter v0.0.5
	github.com/oliveagle/jsonpath => github.com/oliveagle/jsonpath v0.0.0-20180606110733-2e52cf6e6852
	github.com/onsi/ginkgo => github.com/onsi/ginkgo v1.16.5
	github.com/onsi/ginkgo/v2 => github.com/onsi/ginkgo/v2 v2.8.0
	github.com/onsi/gomega => github.com/onsi/gomega v1.26.0
	github.com/op/go-logging => github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/open-policy-agent/opa => github.com/open-policy-agent/opa v0.49.0
	github.com/opencontainers/go-digest => github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec => github.com/opencontainers/image-spec v1.1.0-rc2
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.1.2
	github.com/opencontainers/runtime-spec => github.com/opencontainers/runtime-spec v1.0.3-0.20210326190908-1c3f411f0417
	github.com/opencontainers/selinux => github.com/opencontainers/selinux v1.10.1
	github.com/opensearch-project/opensearch-go => github.com/opensearch-project/opensearch-go v1.1.0
	github.com/opensearch-project/opensearch-go/v2 => github.com/opensearch-project/opensearch-go/v2 v2.0.0
	github.com/opentracing-contrib/go-observer => github.com/opentracing-contrib/go-observer v0.0.0-20170622124052-a52f23424492
	github.com/opentracing/basictracer-go => github.com/opentracing/basictracer-go v1.0.0
	github.com/opentracing/opentracing-go => github.com/opentracing/opentracing-go v1.2.0
	github.com/openzipkin-contrib/zipkin-go-opentracing => github.com/openzipkin-contrib/zipkin-go-opentracing v0.4.5
	github.com/openzipkin/zipkin-go => github.com/openzipkin/zipkin-go v0.2.2
	github.com/operator-framework/api => github.com/operator-framework/api v0.15.0
	github.com/operator-framework/helm-operator-plugins => github.com/operator-framework/helm-operator-plugins v0.0.11
	github.com/operator-framework/operator-lib => github.com/operator-framework/operator-lib v0.11.0
	github.com/pact-foundation/pact-go => github.com/pact-foundation/pact-go v1.0.4
	github.com/pascaldekloe/goe => github.com/pascaldekloe/goe v0.1.0
	github.com/patrickmn/go-cache => github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pborman/uuid => github.com/pborman/uuid v1.2.1
	github.com/pelletier/go-buffruneio => github.com/pelletier/go-buffruneio v0.2.0
	github.com/pelletier/go-toml => github.com/pelletier/go-toml v1.9.5
	github.com/performancecopilot/speed => github.com/performancecopilot/speed v3.0.0+incompatible
	github.com/peterbourgon/diskv => github.com/peterbourgon/diskv v2.0.1+incompatible
	github.com/peterh/liner => github.com/peterh/liner v1.0.1-0.20180619022028-8c1271fcf47f
	github.com/phayes/freeport => github.com/phayes/freeport v0.0.0-20220201140144-74d24b5ae9f5
	github.com/philhofer/fwd => github.com/philhofer/fwd v1.0.0
	github.com/pierrec/lz4 => github.com/pierrec/lz4 v2.0.5+incompatible
	github.com/pkg/diff => github.com/pkg/diff v0.0.0-20210226163009-20ebb0f2a09e
	github.com/pkg/errors => github.com/pkg/errors v0.9.1
	github.com/pkg/sftp => github.com/pkg/sftp v1.10.1
	github.com/pmezard/go-difflib => github.com/pmezard/go-difflib v1.0.0
	github.com/posener/complete => github.com/posener/complete v1.2.3
	github.com/poy/onpar => github.com/poy/onpar v0.0.0-20190519213022-ee068f8ea4d1
	github.com/pquerna/cachecontrol => github.com/pquerna/cachecontrol v0.1.0
	github.com/pquerna/ffjson => github.com/pquerna/ffjson v0.0.0-20190813045741-dac163c6c0a9
	github.com/prashantv/gostub => github.com/prashantv/gostub v1.1.0
	github.com/projectcalico/go-json => github.com/projectcalico/go-json v0.0.0-20161128004156-6219dc7339ba
	github.com/projectcalico/go-yaml => github.com/projectcalico/go-yaml v0.0.0-20161201183616-955bc3e451ef
	github.com/projectcalico/go-yaml-wrapper => github.com/projectcalico/go-yaml-wrapper v0.0.0-20161127220527-598e54215bee
	github.com/projectcalico/kube-controllers => github.com/projectcalico/kube-controllers v3.8.8+incompatible
	github.com/projectcalico/libcalico-go => github.com/projectcalico/libcalico-go v1.7.2-0.20191014160346-2382c6cdd056
	github.com/prometheus-community/prom-label-proxy => github.com/prometheus-community/prom-label-proxy v0.6.0
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring => github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.63.0
	github.com/prometheus-operator/prometheus-operator/pkg/client => github.com/prometheus-operator/prometheus-operator/pkg/client v0.63.0
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.14.0
	github.com/prometheus/common => github.com/prometheus/common v0.39.0
	github.com/prometheus/procfs => github.com/prometheus/procfs v0.8.0
	github.com/prometheus/prometheus => github.com/prometheus/prometheus v0.42.0
	github.com/rainycape/unidecode => github.com/rainycape/unidecode v0.0.0-20150907023854-cb7f23ec59be
	github.com/rcrowley/go-metrics => github.com/rcrowley/go-metrics v0.0.0-20200313005456-10cdbea86bc0
	github.com/remyoudompheng/bigfft => github.com/remyoudompheng/bigfft v0.0.0-20170806203942-52369c62f446
	github.com/robfig/cron/v3 => github.com/robfig/cron/v3 v3.0.1
	github.com/rogpeppe/fastuuid => github.com/rogpeppe/fastuuid v1.2.0
	github.com/rogpeppe/go-charset => github.com/rogpeppe/go-charset v0.0.0-20180617210344-2471d30d28b4
	github.com/rogpeppe/go-internal => github.com/rogpeppe/go-internal v1.8.0
	github.com/rs/cors => github.com/rs/cors v1.7.0
	github.com/rubenv/sql-migrate => github.com/rubenv/sql-migrate v1.1.2
	github.com/russross/blackfriday => github.com/russross/blackfriday v1.6.0
	github.com/russross/blackfriday/v2 => github.com/russross/blackfriday/v2 v2.1.0
	github.com/ryanuber/columnize => github.com/ryanuber/columnize v2.1.0+incompatible
	github.com/safchain/ethtool => github.com/safchain/ethtool v0.0.0-20210803160452-9aa261dae9b1
	github.com/samuel/go-zookeeper => github.com/samuel/go-zookeeper v0.0.0-20200724154423-2164a8ac840e
	github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
	github.com/sclevine/spec => github.com/sclevine/spec v1.2.0
	github.com/sean-/seed => github.com/sean-/seed v0.0.0-20170313163322-e2103e2c3529
	github.com/seccomp/libseccomp-golang => github.com/seccomp/libseccomp-golang v0.9.2-0.20210429002308-3879420cc921
	github.com/sergi/go-diff => github.com/sergi/go-diff v1.1.0
	github.com/shirou/gopsutil/v3 => github.com/shirou/gopsutil/v3 v3.21.9
	github.com/shopspring/decimal => github.com/shopspring/decimal v1.2.0
	github.com/shurcooL/httpfs => github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/shurcooL/sanitized_anchor_name => github.com/shurcooL/sanitized_anchor_name v1.0.0
	github.com/shurcooL/vfsgen => github.com/shurcooL/vfsgen v0.0.0-20200627165143-92b8a710ab6c
	github.com/sirupsen/logrus => github.com/sirupsen/logrus v1.9.0
	github.com/smartystreets/assertions => github.com/smartystreets/assertions v1.0.1
	github.com/smartystreets/goconvey => github.com/smartystreets/goconvey v1.6.4
	github.com/soheilhy/cmux => github.com/soheilhy/cmux v0.1.5
	github.com/sony/gobreaker => github.com/sony/gobreaker v0.4.1
	github.com/sony/sonyflake => github.com/sony/sonyflake v0.0.0-20181109022403-6d5bd6181009
	github.com/speps/go-hashids => github.com/speps/go-hashids v2.0.0+incompatible
	github.com/spf13/afero => github.com/spf13/afero v1.6.0
	github.com/spf13/cast => github.com/spf13/cast v1.4.1
	github.com/spf13/cobra => github.com/spf13/cobra v1.6.1
	github.com/spf13/jwalterweatherman => github.com/spf13/jwalterweatherman v1.1.0
	github.com/spf13/pflag => github.com/spf13/pflag v1.0.5
	github.com/spf13/viper => github.com/spf13/viper v1.8.1
	github.com/src-d/gcfg => github.com/src-d/gcfg v1.4.0
	github.com/stefanberger/go-pkcs11uri => github.com/stefanberger/go-pkcs11uri v0.0.0-20201008174630-78d3cae3a980
	github.com/stoewer/go-strcase => github.com/stoewer/go-strcase v1.2.0
	github.com/streadway/amqp => github.com/streadway/amqp v0.0.0-20190827072141-edfb9018d271
	github.com/streadway/handy => github.com/streadway/handy v0.0.0-20190108123426-d5acb3125c2a
	github.com/stretchr/objx => github.com/stretchr/objx v0.5.0
	github.com/stretchr/testify => github.com/stretchr/testify v1.8.1
	github.com/subosito/gotenv => github.com/subosito/gotenv v1.2.0
	github.com/syndtr/gocapability => github.com/syndtr/gocapability v0.0.0-20200815063812-42c35b437635
	github.com/tchap/go-patricia => github.com/tchap/go-patricia v2.2.6+incompatible
	github.com/tchap/go-patricia/v2 => github.com/tchap/go-patricia/v2 v2.3.1
	github.com/tidwall/pretty => github.com/tidwall/pretty v1.0.0
	github.com/tinylib/msgp => github.com/tinylib/msgp v1.1.0
	github.com/tklauser/go-sysconf => github.com/tklauser/go-sysconf v0.3.9
	github.com/tklauser/numcpus => github.com/tklauser/numcpus v0.3.0
	github.com/tmc/grpc-websocket-proxy => github.com/tmc/grpc-websocket-proxy v0.0.0-20201229170055-e5319fda7802
	github.com/tv42/httpunix => github.com/tv42/httpunix v0.0.0-20150427012821-b75d8614f926
	github.com/ugorji/go => github.com/ugorji/go v0.0.0-20171019201919-bdcc60b419d1
	github.com/ugorji/go/codec => github.com/ugorji/go/codec v0.0.0-20181204163529-d75b2dcb6bc8
	github.com/urfave/cli => github.com/urfave/cli v1.22.2
	github.com/vishvananda/netlink => github.com/vishvananda/netlink v1.1.1-0.20210330154013-f5de75959ad5
	github.com/vishvananda/netns => github.com/vishvananda/netns v0.0.0-20210104183010-2eb08e3e575f
	github.com/weppos/publicsuffix-go => github.com/weppos/publicsuffix-go v0.13.0
	github.com/xanzy/ssh-agent => github.com/xanzy/ssh-agent v0.2.1
	github.com/xdg-go/pbkdf2 => github.com/xdg-go/pbkdf2 v1.0.0
	github.com/xdg-go/scram => github.com/xdg-go/scram v1.1.1
	github.com/xdg-go/stringprep => github.com/xdg-go/stringprep v1.0.3
	github.com/xeipuuv/gojsonpointer => github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb
	github.com/xeipuuv/gojsonreference => github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415
	github.com/xeipuuv/gojsonschema => github.com/xeipuuv/gojsonschema v1.2.0
	github.com/xiang90/probing => github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2
	github.com/xlab/treeprint => github.com/xlab/treeprint v1.1.0
	github.com/xordataexchange/crypt => github.com/xordataexchange/crypt v0.0.3-0.20170626215501-b2862e3d0a77
	github.com/yashtewari/glob-intersection => github.com/yashtewari/glob-intersection v0.1.0
	github.com/youmark/pkcs8 => github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d
	github.com/yvasiyarov/go-metrics => github.com/yvasiyarov/go-metrics v0.0.0-20150112132944-c25f46c4b940
	github.com/yvasiyarov/gorelic => github.com/yvasiyarov/gorelic v0.0.7
	github.com/yvasiyarov/newrelic_platform_go => github.com/yvasiyarov/newrelic_platform_go v0.0.0-20160601141957-9c099fbc30e9
	github.com/ziutek/mymysql => github.com/ziutek/mymysql v1.5.4
	github.com/zmap/zcrypto => github.com/zmap/zcrypto v0.0.0-20200911161511-43ff0ea04f21
	github.com/zmap/zlint/v2 => github.com/zmap/zlint/v2 v2.2.1
	go.etcd.io/bbolt => go.etcd.io/bbolt v1.3.6
	go.etcd.io/etcd => go.etcd.io/etcd v0.5.0-alpha.5.0.20200910180754-dd1b699fc489
	go.etcd.io/etcd/api/v3 => go.etcd.io/etcd/api/v3 v3.5.5
	go.etcd.io/etcd/client/pkg/v3 => go.etcd.io/etcd/client/pkg/v3 v3.5.5
	go.etcd.io/etcd/client/v2 => go.etcd.io/etcd/client/v2 v2.305.5
	go.etcd.io/etcd/client/v3 => go.etcd.io/etcd/client/v3 v3.5.5
	go.etcd.io/etcd/pkg/v3 => go.etcd.io/etcd/pkg/v3 v3.5.5
	go.etcd.io/etcd/raft/v3 => go.etcd.io/etcd/raft/v3 v3.5.5
	go.etcd.io/etcd/server/v3 => go.etcd.io/etcd/server/v3 v3.5.5
	go.mongodb.org/mongo-driver => go.mongodb.org/mongo-driver v1.10.4
	go.mozilla.org/pkcs7 => go.mozilla.org/pkcs7 v0.0.0-20200128120323-432b2356ecb1
	go.opencensus.io => go.opencensus.io v0.23.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc => go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.35.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp => go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.35.0
	go.opentelemetry.io/otel => go.opentelemetry.io/otel v1.10.0
	go.opentelemetry.io/otel/exporters/otlp/internal/retry => go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.10.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace => go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.10.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc => go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.10.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp => go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.3.0
	go.opentelemetry.io/otel/metric => go.opentelemetry.io/otel/metric v0.31.0
	go.opentelemetry.io/otel/sdk => go.opentelemetry.io/otel/sdk v1.10.0
	go.opentelemetry.io/otel/trace => go.opentelemetry.io/otel/trace v1.10.0
	go.opentelemetry.io/proto/otlp => go.opentelemetry.io/proto/otlp v0.19.0
	go.starlark.net => go.starlark.net v0.0.0-20200306205701-8dd3e2ee1dd5
	go.uber.org/atomic => go.uber.org/atomic v1.10.0
	go.uber.org/automaxprocs => go.uber.org/automaxprocs v1.5.1
	go.uber.org/goleak => go.uber.org/goleak v1.2.0
	go.uber.org/multierr => go.uber.org/multierr v1.6.0
	go.uber.org/zap => go.uber.org/zap v1.24.0
	golang.org/x/crypto => golang.org/x/crypto v0.5.0
	golang.org/x/image => golang.org/x/image v0.0.0-20180708004352-c73c2afc3b81
	golang.org/x/lint => golang.org/x/lint v0.0.0-20190301231843-5614ed5bae6f
	golang.org/x/mod => golang.org/x/mod v0.4.0
	golang.org/x/net => golang.org/x/net v0.5.0
	golang.org/x/oauth2 => golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a
	golang.org/x/sync => golang.org/x/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/sys => golang.org/x/sys v0.0.0-20220708085239-5a0f0661e09d
	golang.org/x/term => golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1
	golang.org/x/text => golang.org/x/text v0.6.0
	golang.org/x/time => golang.org/x/time v0.0.0-20190308202827-9d24e82272b4
	golang.org/x/tools => golang.org/x/tools v0.0.0-20190710153321-831012c29e42
	golang.org/x/xerrors => golang.org/x/xerrors v0.0.0-20190717185122-a985d3407aa7
	gomodules.xyz/jsonpatch/v2 => gomodules.xyz/jsonpatch/v2 v2.2.0
	gonum.org/v1/gonum => gonum.org/v1/gonum v0.6.0
	gonum.org/v1/netlib => gonum.org/v1/netlib v0.0.0-20190331212654-76723241ea4e
	gonum.org/v1/plot => gonum.org/v1/plot v0.0.0-20190515093506-e2840ee46a6b
	google.golang.org/api => google.golang.org/api v0.102.0
	google.golang.org/appengine => google.golang.org/appengine v1.6.7
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20221118155620-16455021b5e6
	google.golang.org/grpc => google.golang.org/grpc v1.52.3
	google.golang.org/grpc/cmd/protoc-gen-go-grpc => google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/protobuf => google.golang.org/protobuf v1.28.1
	gopkg.in/airbrake/gobrake.v2 => gopkg.in/airbrake/gobrake.v2 v2.0.9
	gopkg.in/alecthomas/kingpin.v2 => gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/asn1-ber.v1 => gopkg.in/asn1-ber.v1 v1.0.0-20181015200546-f715ec2f112d
	gopkg.in/cas.v2 => gopkg.in/cas.v2 v2.2.0
	gopkg.in/check.v1 => gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c
	gopkg.in/cheggaaa/pb.v1 => gopkg.in/cheggaaa/pb.v1 v1.0.25
	gopkg.in/errgo.v2 => gopkg.in/errgo.v2 v2.1.0
	gopkg.in/gcfg.v1 => gopkg.in/gcfg.v1 v1.2.3
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 => gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2
	gopkg.in/go-playground/assert.v1 => gopkg.in/go-playground/assert.v1 v1.2.1
	gopkg.in/go-playground/validator.v9 => gopkg.in/go-playground/validator.v9 v9.27.0
	gopkg.in/inf.v0 => gopkg.in/inf.v0 v0.9.1
	gopkg.in/ini.v1 => gopkg.in/ini.v1 v1.62.0
	gopkg.in/natefinch/lumberjack.v2 => gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/square/go-jose.v2 => gopkg.in/square/go-jose.v2 v2.5.1
	gopkg.in/src-d/go-billy.v4 => gopkg.in/src-d/go-billy.v4 v4.3.2
	gopkg.in/src-d/go-git-fixtures.v3 => gopkg.in/src-d/go-git-fixtures.v3 v3.5.0
	gopkg.in/src-d/go-git.v4 => gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/tchap/go-patricia.v2 => gopkg.in/tchap/go-patricia.v2 v2.2.6
	gopkg.in/tomb.v1 => gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7
	gopkg.in/warnings.v0 => gopkg.in/warnings.v0 v0.1.2
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 => gopkg.in/yaml.v3 v3.0.1
	gotest.tools => gotest.tools v2.2.0+incompatible
	gotest.tools/v3 => gotest.tools/v3 v3.0.3
	helm.sh/helm/v3 => helm.sh/helm/v3 v3.10.3
	honnef.co/go/tools => honnef.co/go/tools v0.0.0-20190523083050-ea95bdfd59fc
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
	k8s.io/component-helpers => k8s.io/component-helpers v0.26.1
	k8s.io/cri-api => k8s.io/cri-api v0.25.0
	k8s.io/gengo => k8s.io/gengo v0.0.0-20220902162205-c0856e24416d
	k8s.io/klog => k8s.io/klog v1.0.0
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.80.1
	k8s.io/kms => k8s.io/kms v0.26.1
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20230202010329-39b3636cbaa3
	k8s.io/kubectl => k8s.io/kubectl v0.26.1
	k8s.io/metrics => k8s.io/metrics v0.26.1
	k8s.io/utils => k8s.io/utils v0.0.0-20221128185143-99ec85e7a448
	kubesphere.io/api => ./staging/src/kubesphere.io/api
	kubesphere.io/client-go => ./staging/src/kubesphere.io/client-go
	kubesphere.io/monitoring-dashboard => kubesphere.io/monitoring-dashboard v0.2.2
	kubesphere.io/utils => ./staging/src/kubesphere.io/utils
	modernc.org/cc => modernc.org/cc v1.0.0
	modernc.org/golex => modernc.org/golex v1.0.0
	modernc.org/mathutil => modernc.org/mathutil v1.0.0
	modernc.org/strutil => modernc.org/strutil v1.0.0
	modernc.org/xc => modernc.org/xc v1.0.0
	oras.land/oras-go => oras.land/oras-go v1.2.2
	rsc.io/goversion => rsc.io/goversion v1.2.0
	rsc.io/pdf => rsc.io/pdf v0.1.1
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client => sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.0.35
	sigs.k8s.io/application => sigs.k8s.io/application v0.8.4-0.20201016185654-c8e2959e57a0
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.14.4
	sigs.k8s.io/controller-tools => sigs.k8s.io/controller-tools v0.11.1
	sigs.k8s.io/json => sigs.k8s.io/json v0.0.0-20220713155537-f223a00ba0e2
	sigs.k8s.io/kind => sigs.k8s.io/kind v0.8.1
	sigs.k8s.io/kubebuilder/v3 => sigs.k8s.io/kubebuilder/v3 v3.0.0-alpha.0.0.20220607134920-f19b01da2468
	sigs.k8s.io/kubefed => github.com/kubesphere/kubefed v0.0.0-20230207032540-cdda80892665
	sigs.k8s.io/kustomize/api => sigs.k8s.io/kustomize/api v0.12.1
	sigs.k8s.io/kustomize/kustomize/v4 => sigs.k8s.io/kustomize/kustomize/v4 v4.5.7
	sigs.k8s.io/kustomize/kyaml => sigs.k8s.io/kustomize/kyaml v0.13.9
	sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v1.0.1-0.20191108220359-b1b620dd3f06
	sigs.k8s.io/structured-merge-diff/v3 => sigs.k8s.io/structured-merge-diff/v3 v3.0.0
	sigs.k8s.io/structured-merge-diff/v4 => sigs.k8s.io/structured-merge-diff/v4 v4.2.3
	sigs.k8s.io/yaml => sigs.k8s.io/yaml v1.3.0
	sourcegraph.com/sourcegraph/appdash => sourcegraph.com/sourcegraph/appdash v0.0.0-20190731080439-ebfcffb1b5c0
)
