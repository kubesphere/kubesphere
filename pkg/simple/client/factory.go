package client

import (
	"errors"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/devops/jenkins"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/kubesphere"
	"kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"kubesphere.io/kubesphere/pkg/simple/client/logging/elasticsearch"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring/prometheus"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"kubesphere.io/kubesphere/pkg/simple/client/sonarqube"
	"sync"
)

var ErrClientSetNotEnabled = errors.New("client set not enabled")

type ClientSetOptions struct {
	mySQLOptions        *mysql.Options
	redisOptions        *cache.Options
	kubernetesOptions   *k8s.KubernetesOptions
	devopsOptions       *jenkins.Options
	sonarqubeOptions    *sonarqube.Options
	ldapOptions         *ldap.Options
	s3Options           *s3.Options
	openPitrixOptions   *openpitrix.Options
	prometheusOptions   *prometheus.Options
	kubesphereOptions   *kubesphere.Options
	elasticsearhOptions *elasticsearch.Options
}

func NewClientSetOptions() *ClientSetOptions {
	return &ClientSetOptions{
		mySQLOptions:        mysql.NewMySQLOptions(),
		redisOptions:        cache.NewRedisOptions(),
		kubernetesOptions:   k8s.NewKubernetesOptions(),
		ldapOptions:         ldap.NewOptions(),
		devopsOptions:       jenkins.NewDevopsOptions(),
		sonarqubeOptions:    sonarqube.NewSonarQubeOptions(),
		s3Options:           s3.NewS3Options(),
		openPitrixOptions:   openpitrix.NewOptions(),
		prometheusOptions:   prometheus.NewPrometheusOptions(),
		kubesphereOptions:   kubesphere.NewKubeSphereOptions(),
		elasticsearhOptions: elasticsearch.NewElasticSearchOptions(),
	}
}

func (c *ClientSetOptions) SetMySQLOptions(options *mysql.Options) *ClientSetOptions {
	c.mySQLOptions = options
	return c
}

func (c *ClientSetOptions) SetRedisOptions(options *cache.Options) *ClientSetOptions {
	c.redisOptions = options
	return c
}

func (c *ClientSetOptions) SetKubernetesOptions(options *k8s.KubernetesOptions) *ClientSetOptions {
	c.kubernetesOptions = options
	return c
}

func (c *ClientSetOptions) SetDevopsOptions(options *jenkins.Options) *ClientSetOptions {
	c.devopsOptions = options
	return c
}

func (c *ClientSetOptions) SetLdapOptions(options *ldap.Options) *ClientSetOptions {
	c.ldapOptions = options
	return c
}

func (c *ClientSetOptions) SetS3Options(options *s3.Options) *ClientSetOptions {
	c.s3Options = options
	return c
}

func (c *ClientSetOptions) SetOpenPitrixOptions(options *openpitrix.Options) *ClientSetOptions {
	c.openPitrixOptions = options
	return c
}

func (c *ClientSetOptions) SetPrometheusOptions(options *prometheus.Options) *ClientSetOptions {
	c.prometheusOptions = options
	return c
}

func (c *ClientSetOptions) SetSonarQubeOptions(options *sonarqube.Options) *ClientSetOptions {
	c.sonarqubeOptions = options
	return c
}

func (c *ClientSetOptions) SetKubeSphereOptions(options *kubesphere.Options) *ClientSetOptions {
	c.kubesphereOptions = options
	return c
}

func (c *ClientSetOptions) SetElasticSearchOptions(options *elasticsearch.Options) *ClientSetOptions {
	c.elasticsearhOptions = options
	return c
}

// ClientSet provide best of effort service to initialize clients,
// but there is no guarantee to return a valid client instance,
// so do validity check before use
type ClientSet struct {
	csoptions *ClientSetOptions
	stopCh    <-chan struct{}

	mySQLClient *mysql.Client

	k8sClient           k8s.Client
	ldapClient          ldap.Client
	devopsClient        *jenkins.Client
	sonarQubeClient     *sonarqube.Client
	redisClient         cache.Interface
	s3Client            s3.Interface
	prometheusClient    monitoring.Interface
	openpitrixClient    openpitrix.Client
	kubesphereClient    *kubesphere.Client
	elasticSearchClient *elasticsearch.Elasticsearch
}

var mutex sync.Mutex

// global clientsets instance
var sharedClientSet *ClientSet

func ClientSets() *ClientSet {
	return sharedClientSet
}

func NewClientSetFactory(c *ClientSetOptions, stopCh <-chan struct{}) *ClientSet {
	sharedClientSet = &ClientSet{csoptions: c, stopCh: stopCh}

	if c.kubernetesOptions != nil {
		sharedClientSet.k8sClient = k8s.NewKubernetesClientOrDie(c.kubernetesOptions)
	}

	if c.kubesphereOptions != nil {
		sharedClientSet.kubesphereClient = kubesphere.NewKubeSphereClient(c.kubesphereOptions)
	}

	return sharedClientSet
}

// lazy creating
func (cs *ClientSet) MySQL() (*mysql.Database, error) {
	var err error

	if cs.csoptions.mySQLOptions == nil || cs.csoptions.mySQLOptions.Host == "" {
		return nil, ErrClientSetNotEnabled
	}

	if cs.mySQLClient != nil {
		return cs.mySQLClient.Database(), nil
	} else {
		mutex.Lock()
		defer mutex.Unlock()
		if cs.mySQLClient == nil {
			cs.mySQLClient, err = mysql.NewMySQLClient(cs.csoptions.mySQLOptions, cs.stopCh)
			if err != nil {
				return nil, err
			}
		}

		return cs.mySQLClient.Database(), nil
	}
}

func (cs *ClientSet) Cache() (cache.Interface, error) {
	var err error

	if cs.csoptions.redisOptions == nil || cs.csoptions.redisOptions.RedisURL == "" {
		return nil, ErrClientSetNotEnabled
	}

	if cs.redisClient != nil {
		return cs.redisClient, nil
	} else {
		mutex.Lock()
		defer mutex.Unlock()
		if cs.redisClient == nil {
			cs.redisClient, err = cache.NewRedisClient(cs.csoptions.redisOptions, cs.stopCh)
			if err != nil {
				return nil, err
			}
		}

		return cs.redisClient, nil
	}
}

func (cs *ClientSet) Devops() (devops.Interface, error) {
	//var err error
	//
	//if cs.csoptions.devopsOptions == nil || cs.csoptions.devopsOptions.Host == "" {
	//	return nil, ErrClientSetNotEnabled
	//}
	//
	//if cs.devopsClient != nil {
	//	return cs.devopsClient, nil
	//} else {
	//	mutex.Lock()
	//	defer mutex.Unlock()
	//
	//	if cs.devopsClient == nil {
	//		cs.devopsClient, err = jenkins.NewDevopsClient(cs.csoptions.devopsOptions)
	//		if err != nil {
	//			return nil, err
	//		}
	//	}
	//	return cs.devopsClient, nil
	//}
	return nil, nil
}

func (cs *ClientSet) SonarQube() (*sonarqube.Client, error) {
	var err error

	if cs.csoptions.sonarqubeOptions == nil || cs.csoptions.sonarqubeOptions.Host == "" {
		return nil, ErrClientSetNotEnabled
	}

	if cs.sonarQubeClient != nil {
		return cs.sonarQubeClient, nil
	} else {
		mutex.Lock()
		defer mutex.Unlock()

		if cs.sonarQubeClient == nil {
			cs.sonarQubeClient, err = sonarqube.NewSonarQubeClient(cs.csoptions.sonarqubeOptions)
			if err != nil {
				return nil, err
			}
		}
		return cs.sonarQubeClient, nil
	}
}

func (cs *ClientSet) Ldap() (ldap.Client, error) {
	var err error

	if cs.csoptions.ldapOptions == nil || cs.csoptions.ldapOptions.Host == "" {
		return nil, ErrClientSetNotEnabled
	}

	if cs.ldapClient != nil {
		return cs.ldapClient, nil
	} else {
		mutex.Lock()
		defer mutex.Unlock()

		if cs.ldapClient == nil {
			cs.ldapClient, err = ldap.NewClient(cs.csoptions.ldapOptions, cs.stopCh)
			if err != nil {
				return nil, err
			}
		}
		return cs.ldapClient, nil
	}
}

// since kubernetes client is required, we will
// create it on setup
func (cs *ClientSet) K8s() k8s.Client {
	return cs.k8sClient
}

func (cs *ClientSet) S3() (s3.Interface, error) {
	var err error

	if cs.csoptions.s3Options == nil || cs.csoptions.s3Options.Endpoint == "" {
		return nil, ErrClientSetNotEnabled
	}

	if cs.s3Client != nil {
		return cs.s3Client, nil
	} else {
		mutex.Lock()
		defer mutex.Unlock()

		if cs.s3Client == nil {
			cs.s3Client, err = s3.NewS3Client(cs.csoptions.s3Options)
			if err != nil {
				return nil, err
			}
		}
		return cs.s3Client, nil
	}
}

func (cs *ClientSet) OpenPitrix() (openpitrix.Client, error) {
	var err error

	if cs.csoptions.openPitrixOptions == nil ||
		cs.csoptions.openPitrixOptions.RepoManagerEndpoint == "" ||
		cs.csoptions.openPitrixOptions.RuntimeManagerEndpoint == "" ||
		cs.csoptions.openPitrixOptions.ClusterManagerEndpoint == "" ||
		cs.csoptions.openPitrixOptions.AppManagerEndpoint == "" ||
		cs.csoptions.openPitrixOptions.AttachmentManagerEndpoint == "" ||
		cs.csoptions.openPitrixOptions.RepoIndexerEndpoint == "" ||
		cs.csoptions.openPitrixOptions.CategoryManagerEndpoint == "" {
		return nil, ErrClientSetNotEnabled
	}

	if cs.openpitrixClient != nil {
		return cs.openpitrixClient, nil
	} else {
		cs.openpitrixClient, err = openpitrix.NewClient(cs.csoptions.openPitrixOptions)
		if err != nil {
			return nil, err
		}

		return cs.openpitrixClient, nil
	}
}

func (cs *ClientSet) MonitoringClient() (monitoring.Interface, error) {
	if cs.csoptions.prometheusOptions == nil || cs.csoptions.prometheusOptions.Endpoint == "" {
		return nil, ErrClientSetNotEnabled
	}

	if cs.prometheusClient != nil {
		return cs.prometheusClient, nil
	} else {
		mutex.Lock()
		defer mutex.Unlock()

		if cs.prometheusClient == nil {
			cs.prometheusClient = prometheus.NewPrometheus(cs.csoptions.prometheusOptions)
		}
		return cs.prometheusClient, nil
	}
}

func (cs *ClientSet) KubeSphere() *kubesphere.Client {
	return cs.kubesphereClient
}

func (cs *ClientSet) ElasticSearch() (*elasticsearch.Elasticsearch, error) {
	var err error

	if cs.csoptions.elasticsearhOptions == nil || cs.csoptions.elasticsearhOptions.Host == "" {
		return nil, ErrClientSetNotEnabled
	}

	if cs.elasticSearchClient != nil {
		return cs.elasticSearchClient, nil
	} else {
		mutex.Lock()
		defer mutex.Unlock()

		if cs.elasticSearchClient == nil {
			cs.elasticSearchClient, err = elasticsearch.NewElasticsearch(cs.csoptions.elasticsearhOptions)
			if err != nil {
				return nil, err
			}
		}

		return cs.elasticSearchClient, nil
	}
}
