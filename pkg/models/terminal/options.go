/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package terminal

import "github.com/spf13/pflag"

type Options struct {
	KubectlOptions   KubectlOptions   `json:"kubectl" yaml:"kubectl" mapstructure:"kubectl"`
	NodeShellOptions NodeShellOptions `json:"node" yaml:"node" mapstructure:"node"`
	UploadFileLimit  string           `json:"uploadFileLimit" yaml:"uploadFileLimit"`
}

type KubectlOptions struct {
	// Image defines the Pod image used by the kubectl web terminal.
	Image string `json:"image,omitempty" yaml:"image,omitempty"`
}

type NodeShellOptions struct {
	// Image defines the Pod image used by the node terminal.
	Image   string `json:"image,omitempty" yaml:"image,omitempty"`
	Timeout int    `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

func NewOptions() *Options {
	return &Options{
		KubectlOptions: KubectlOptions{
			Image: "kubesphere/kubectl:v1.27.4",
		},
		NodeShellOptions: NodeShellOptions{
			Image:   "alpine:3.15",
			Timeout: 600,
		},
		UploadFileLimit: "100Mi",
	}
}

func (s *Options) Validate() []error {
	var errs []error
	return errs
}

func (s *Options) ApplyTo(options *Options) {

}

func (s *Options) AddFlags(fs *pflag.FlagSet, c *Options) {

}
