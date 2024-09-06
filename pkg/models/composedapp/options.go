/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package composedapp

import (
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/labels"
)

type Options struct {
	// KubeSphere is using sigs.k8s.io/application as fundamental object to implement Application Management.
	// There are other projects also built on sigs.k8s.io/application, when KubeSphere installed along side
	// them, conflicts happen. So we leave an option to only reconcile applications  matched with the given
	// selector. Default will reconcile all applications.
	//    For example
	//      "kubesphere.io/creator=" means reconcile applications with this label key
	//      "!kubesphere.io/creator" means exclude applications with this key
	AppSelector string `json:"appSelector,omitempty" yaml:"appSelector,omitempty" mapstructure:"appSelector,omitempty"`
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) Validate() []error {
	var err []error
	if _, validateErr := labels.Parse(o.AppSelector); validateErr != nil {
		err = append(err, validateErr)
	}
	return err
}

func (o *Options) AddFlags(fs *pflag.FlagSet, s *Options) {
	fs.StringVar(&o.AppSelector, "app-selector", s.AppSelector, "Selector to filter k8s applications to reconcile")
}
