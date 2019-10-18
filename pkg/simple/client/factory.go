package client

import (
	"fmt"
	goredis "github.com/go-redis/redis"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	esclient "kubesphere.io/kubesphere/pkg/simple/client/elasticsearch"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/kubesphere"
	"kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"kubesphere.io/kubesphere/pkg/simple/client/prometheus"
	"kubesphere.io/kubesphere/pkg/simple/client/redis"
	"kubesphere.io/kubesphere/pkg/simple/client/s2is3"
	"kubesphere.io/kubesphere/pkg/simple/client/sonarqube"
	"sync"
)

type ClientSetNotEnabledError struct {
	err error
}

func (e ClientSetNotEnabledError) Error() string {
	return fmt.Sprintf("client set not enabled: %v", e.err)
}

type ClientSetOptions struct {
	mySQLOptions        *mysql.MySQLOptions
	redisOptions        *redis.RedisOptions
	kubernetesOptions   *k8s.KubernetesOptions
	devopsOptions       *devops.DevopsOptions
	sonarqubeOptions    *sonarqube.SonarQubeOptions
	ldapOptions         *ldap.LdapOptions
	s3Options           *s2is3.S3Options
	openPitrixOptions   *openpitrix.OpenPitrixOptions
	prometheusOptions   *prometheus.PrometheusOptions
	kubesphereOptions   *kubesphere.KubeSphereOptions
	elasticSearhOptions *esclient.ElasticSearchOptions
}

func NewClientSetOptions() *ClientSetOptions {
	return &ClientSetOptions{
		mySQLOptions:        mysql.NewMySQLOptions(),
		redisOptions:        redis.NewRedisOptions(),
		kubernetesOptions:   k8s.NewKubernetesOptions(),
		ldapOptions:         ldap.NewLdapOptions(),
		devopsOptions:       devops.NewDevopsOptions(),
		sonarqubeOptions:    sonarqube.NewSonarQubeOptions(),
		s3Options:           s2is3.NewS3Options(),
		openPitrixOptions:   openpitrix.NewOpenPitrixOptions(),
		prometheusOptions:   prometheus.NewPrometheusOptions(),
		kubesphereOptions:   kubesphere.NewKubeSphereOptions(),
		elasticSearhOptions: esclient.NewElasticSearchOptions(),
	}
}

func (c *ClientSetOptions) SetMySQLOptions(options *mysql.MySQLOptions) *ClientSetOptions {
	c.mySQLOptions = options
	return c
}

func (c *ClientSetOptions) SetRedisOptions(options *redis.RedisOptions) *ClientSetOptions {
	c.redisOptions = options
	return c
}

func (c *ClientSetOptions) SetKubernetesOptions(options *k8s.KubernetesOptions) *ClientSetOptions {
	c.kubernetesOptions = options
	return c
}

func (c *ClientSetOptions) SetDevopsOptions(options *devops.DevopsOptions) *ClientSetOptions {
	c.devopsOptions = options
	return c
}

func (c *ClientSetOptions) SetLdapOptions(options *ldap.LdapOptions) *ClientSetOptions {
	c.ldapOptions = options
	return c
}

func (c *ClientSetOptions) SetS3Options(options *s2is3.S3Options) *ClientSetOptions {
	c.s3Options = options
	return c
}

func (c *ClientSetOptions) SetOpenPitrixOptions(options *openpitrix.OpenPitrixOptions) *ClientSetOptions {
	c.openPitrixOptions = options
	return c
}

func (c *ClientSetOptions) SetPrometheusOptions(options *prometheus.PrometheusOptions) *ClientSetOptions {
	c.prometheusOptions = options
	return c
}

func (c *ClientSetOptions) SetSonarQubeOptions(options *sonarqube.SonarQubeOptions) *ClientSetOptions {
	c.sonarqubeOptions = options
	return c
}

func (c *ClientSetOptions) SetKubeSphereOptions(options *kubesphere.KubeSphereOptions) *ClientSetOptions {
	c.kubesphereOptions = options
	return c
}

func (c *ClientSetOptions) SetElasticSearchOptions(options *esclient.ElasticSearchOptions) *ClientSetOptions {
	c.elasticSearhOptions = options
	return c
}

// ClientSet provide best of effort service to initialize clients,
// but there is no guarantee to return a valid client instance,
// so do validity check before use
type ClientSet struct {
	csoptions *ClientSetOptions
	stopCh    <-chan struct{}

	mySQLClient *mysql.MySQLClient

	k8sClient           *k8s.KubernetesClient
	ldapClient          *ldap.LdapClient
	devopsClient        *devops.DevopsClient
	sonarQubeClient     *sonarqube.SonarQubeClient
	redisClient         *redis.RedisClient
	s3Client            *s2is3.S3Client
	prometheusClient    *prometheus.PrometheusClient
	openpitrixClient    *openpitrix.OpenPitrixClient
	kubesphereClient    *kubesphere.KubeSphereClient
	elasticSearchClient *esclient.ElasticSearchClient
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
		return nil, ClientSetNotEnabledError{}
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

func (cs *ClientSet) Redis() (*goredis.Client, error) {
	var err error

	if cs.csoptions.redisOptions == nil || cs.csoptions.redisOptions.RedisURL == "" {
		return nil, ClientSetNotEnabledError{}
	}

	if cs.redisClient != nil {
		return cs.redisClient.Redis(), nil
	} else {
		mutex.Lock()
		defer mutex.Unlock()
		if cs.redisClient == nil {
			cs.redisClient, err = redis.NewRedisClient(cs.csoptions.redisOptions, cs.stopCh)
			if err != nil {
				return nil, err
			}
		}

		return cs.redisClient.Redis(), nil
	}
}

func (cs *ClientSet) Devops() (*devops.DevopsClient, error) {
	var err error

	if cs.csoptions.devopsOptions == nil || cs.csoptions.devopsOptions.Host == "" {
		return nil, ClientSetNotEnabledError{}
	}

	if cs.devopsClient != nil {
		return cs.devopsClient, nil
	} else {
		mutex.Lock()
		defer mutex.Unlock()

		if cs.devopsClient == nil {
			cs.devopsClient, err = devops.NewDevopsClient(cs.csoptions.devopsOptions)
			if err != nil {
				return nil, err
			}
		}
		return cs.devopsClient, nil
	}
}

func (cs *ClientSet) SonarQube() (*sonarqube.SonarQubeClient, error) {
	var err error

	if cs.csoptions.sonarqubeOptions == nil || cs.csoptions.sonarqubeOptions.Host == "" {
		return nil, ClientSetNotEnabledError{}
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

func (cs *ClientSet) Ldap() (*ldap.LdapClient, error) {
	var err error

	if cs.csoptions.ldapOptions == nil || cs.csoptions.ldapOptions.Host == "" {
		return nil, ClientSetNotEnabledError{}
	}

	if cs.ldapClient != nil {
		return cs.ldapClient, nil
	} else {
		mutex.Lock()
		defer mutex.Unlock()

		if cs.ldapClient == nil {
			cs.ldapClient, err = ldap.NewLdapClient(cs.csoptions.ldapOptions, cs.stopCh)
			if err != nil {
				return nil, err
			}
		}
		return cs.ldapClient, nil
	}
}

// since kubernetes client is required, we will
// create it on setup
func (cs *ClientSet) K8s() *k8s.KubernetesClient {
	return cs.k8sClient
}

func (cs *ClientSet) S3() (*s2is3.S3Client, error) {
	var err error

	if cs.csoptions.s3Options == nil || cs.csoptions.s3Options.Endpoint == "" {
		return nil, ClientSetNotEnabledError{}
	}

	if cs.s3Client != nil {
		return cs.s3Client, nil
	} else {
		mutex.Lock()
		defer mutex.Unlock()

		if cs.s3Client == nil {
			cs.s3Client, err = s2is3.NewS3Client(cs.csoptions.s3Options)
			if err != nil {
				return nil, err
			}
		}
		return cs.s3Client, nil
	}
}

func (cs *ClientSet) OpenPitrix() (*openpitrix.OpenPitrixClient, error) {
	var err error

	if cs.csoptions.openPitrixOptions == nil ||
		cs.csoptions.openPitrixOptions.RepoManagerEndpoint == "" ||
		cs.csoptions.openPitrixOptions.RuntimeManagerEndpoint == "" ||
		cs.csoptions.openPitrixOptions.ClusterManagerEndpoint == "" ||
		cs.csoptions.openPitrixOptions.AppManagerEndpoint == "" ||
		cs.csoptions.openPitrixOptions.AttachmentManagerEndpoint == "" ||
		cs.csoptions.openPitrixOptions.RepoIndexerEndpoint == "" ||
		cs.csoptions.openPitrixOptions.CategoryManagerEndpoint == "" {
		return nil, ClientSetNotEnabledError{}
	}

	if cs.openpitrixClient != nil {
		return cs.openpitrixClient, nil
	} else {
		cs.openpitrixClient, err = openpitrix.NewOpenPitrixClient(cs.csoptions.openPitrixOptions)
		if err != nil {
			return nil, err
		}

		return cs.openpitrixClient, nil
	}
}

func (cs *ClientSet) Prometheus() (*prometheus.PrometheusClient, error) {
	var err error

	if cs.csoptions.prometheusOptions == nil || cs.csoptions.prometheusOptions.Endpoint == "" {
		return nil, ClientSetNotEnabledError{}
	}

	if cs.prometheusClient != nil {
		return cs.prometheusClient, nil
	} else {
		mutex.Lock()
		defer mutex.Unlock()

		if cs.prometheusClient == nil {
			cs.prometheusClient, err = prometheus.NewPrometheusClient(cs.csoptions.prometheusOptions)
			if err != nil {
				return nil, err
			}
		}
		return cs.prometheusClient, nil
	}
}

func (cs *ClientSet) KubeSphere() *kubesphere.KubeSphereClient {
	return cs.kubesphereClient
}

func (cs *ClientSet) ElasticSearch() (*esclient.ElasticSearchClient, error) {
	var err error

	if cs.csoptions.elasticSearhOptions == nil || cs.csoptions.elasticSearhOptions.Host == "" {
		return nil, ClientSetNotEnabledError{}
	}

	if cs.elasticSearchClient != nil {
		return cs.elasticSearchClient, nil
	} else {
		mutex.Lock()
		defer mutex.Unlock()

		if cs.elasticSearchClient == nil {
			cs.elasticSearchClient, err = esclient.NewLoggingClient(cs.csoptions.elasticSearhOptions)
			if err != nil {
				return nil, err
			}
		}

		return cs.elasticSearchClient, nil
	}
}
