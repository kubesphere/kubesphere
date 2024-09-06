/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package multicluster

import (
	"errors"
	"time"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/validation"
)

const (
	DefaultResyncPeriod    = 900 * time.Second
	DefaultHostClusterName = "host"
)

type Options struct {
	// ProxyPublishService is the service name of multicluster component tower.
	//   If this field provided, apiserver going to use the ingress.ip of this service.
	// This field will be used when generating agent deployment yaml for joining clusters.
	ProxyPublishService string `json:"proxyPublishService,omitempty" yaml:"proxyPublishService,omitempty"`

	// ProxyPublishAddress is the public address of tower for all cluster agents.
	//   This field takes precedence over field ProxyPublishService.
	// If both field ProxyPublishService and ProxyPublishAddress are empty, apiserver will
	// return 404 Not Found for all cluster agent yaml requests.
	ProxyPublishAddress string `json:"proxyPublishAddress,omitempty" yaml:"proxyPublishAddress,omitempty"`

	// AgentImage is the image used when generating deployment for all cluster agents.
	AgentImage string `json:"agentImage,omitempty" yaml:"agentImage,omitempty"`

	// ClusterControllerResyncPeriod is the resync period used by cluster controller.
	ClusterControllerResyncPeriod time.Duration `json:"clusterControllerResyncPeriod,omitempty" yaml:"clusterControllerResyncPeriod,omitempty"`

	// HostClusterName is the name of the control plane cluster, default set to host.
	HostClusterName string `json:"hostClusterName,omitempty" yaml:"hostClusterName,omitempty"`

	// ClusterName is the name of the current cluster,
	// this value will be set by the cluster-controller and stored in the kubesphere-config configmap.
	ClusterName string `json:"clusterName,omitempty" yaml:"clusterName,omitempty"`

	// ClusterRole is the role of the current cluster,
	// available values: host, member.
	ClusterRole string `json:"clusterRole,omitempty" yaml:"clusterRole,omitempty"`

	// ChartPath is the path of the helm chart file in the container. It is used to install ks-core in the member cluster.
	// By default, no setting is required.
	// If you need to customize it, you can mount the chart file to the ks-controller-manager Pod and change this value.
	ChartPath string `json:"chartPath,omitempty" yaml:"chartPath,omitempty"`
}

// NewOptions returns a default nil options
func NewOptions() *Options {
	return &Options{
		ProxyPublishAddress:           "",
		ProxyPublishService:           "",
		AgentImage:                    "kubesphere/tower:v1.0",
		ClusterControllerResyncPeriod: DefaultResyncPeriod,
		HostClusterName:               DefaultHostClusterName,
	}
}

func (o *Options) Validate() []error {
	var err []error

	res := validation.IsQualifiedName(o.HostClusterName)
	if len(res) == 0 {
		return err
	}

	err = append(err, errors.New("failed to create the host cluster because of invalid cluster name"))
	for _, str := range res {
		err = append(err, errors.New(str))
	}
	return err
}

func (o *Options) AddFlags(fs *pflag.FlagSet, s *Options) {
	fs.StringVar(&o.ProxyPublishService, "proxy-publish-service", s.ProxyPublishService, ""+
		"Service name of tower. APIServer will use its ingress address as proxy publish address."+
		"For example, tower.kubesphere-system.svc.")

	fs.StringVar(&o.ProxyPublishAddress, "proxy-publish-address", s.ProxyPublishAddress, ""+
		"Public address of tower, APIServer will use this field as proxy publish address. This field "+
		"takes precedence over field proxy-publish-service. For example, http://139.198.121.121:8080.")

	fs.StringVar(&o.AgentImage, "agent-image", s.AgentImage, ""+
		"This field is used when generating deployment yaml for agent.")

	fs.DurationVar(&o.ClusterControllerResyncPeriod, "cluster-controller-resync-period", s.ClusterControllerResyncPeriod,
		"Cluster controller resync period to sync cluster resource. e.g. 2m 5m 10m ... default set to 2m")

	fs.StringVar(&o.HostClusterName, "host-cluster-name", s.HostClusterName, "the name of the control plane"+
		" cluster, default set to host")
}
