package k8s

import (
	"github.com/spf13/pflag"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
	"os"
)

type KubernetesOptions struct {
	// kubeconfig path, if not specified, will use
	// in cluster way to create clientset
	KubeConfig string `json:"kubeconfig" yaml:"kubeconfig"`

	// kubernetes apiserver public address, used to generate kubeconfig
	// for downloading, default to host defined in kubeconfig
	// +optional
	Master string `json:"master,omitempty" yaml:"master"`

	// kubernetes clientset qps
	// +optional
	QPS float32 `json:"qps,omitemtpy" yaml:"qps"`

	// kubernetes clientset burst
	// +optional
	Burst int `json:"burst,omitempty" yaml:"burst"`
}

// NewKubernetesOptions returns a `zero` instance
func NewKubernetesOptions() *KubernetesOptions {
	return &KubernetesOptions{
		KubeConfig: "",
		QPS:        1e6,
		Burst:      1e6,
	}
}

func (k *KubernetesOptions) Validate() []error {
	errors := []error{}

	if k.KubeConfig != "" {
		if _, err := os.Stat(k.KubeConfig); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func (k *KubernetesOptions) ApplyTo(options *KubernetesOptions) {
	reflectutils.Override(options, k)
}

func (k *KubernetesOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&k.KubeConfig, "kubeconfig", k.KubeConfig, ""+
		"Path for kubernetes kubeconfig file, if left blank, will use "+
		"in cluster way.")

	fs.StringVar(&k.Master, "master", k.Master, ""+
		"Used to generate kubeconfig for downloading, if not specified, will use host in kubeconfig.")
}
