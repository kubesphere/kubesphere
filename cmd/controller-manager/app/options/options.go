package options

import (
	"flag"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	kubesphereconfig "kubesphere.io/kubesphere/pkg/server/config"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/s2is3"
	"strings"
)

type KubeSphereControllerManagerOptions struct {
	KubernetesOptions *k8s.KubernetesOptions
	DevopsOptions     *devops.DevopsOptions
	S3Options         *s2is3.S3Options
}

func NewKubeSphereControllerManagerOptions() *KubeSphereControllerManagerOptions {
	s := &KubeSphereControllerManagerOptions{
		KubernetesOptions: k8s.NewKubernetesOptions(),
		DevopsOptions:     devops.NewDevopsOptions(),
		S3Options:         s2is3.NewS3Options(),
	}

	return s
}

func (s *KubeSphereControllerManagerOptions) ApplyTo(conf *kubesphereconfig.Config) {
	s.S3Options.ApplyTo(conf.S3Options)
	s.KubernetesOptions.ApplyTo(conf.KubernetesOptions)
	s.DevopsOptions.ApplyTo(conf.DevopsOptions)

}

func (s *KubeSphereControllerManagerOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}

	s.KubernetesOptions.AddFlags(fss.FlagSet("kubernetes"))
	s.DevopsOptions.AddFlags(fss.FlagSet("devops"))
	s.S3Options.AddFlags(fss.FlagSet("s3"))

	fs := fss.FlagSet("klog")
	local := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		fs.AddGoFlag(fl)
	})

	return fss
}

func (s *KubeSphereControllerManagerOptions) Validate() []error {
	var errs []error

	errs = append(errs, s.DevopsOptions.Validate()...)
	errs = append(errs, s.KubernetesOptions.Validate()...)
	errs = append(errs, s.S3Options.Validate()...)

	return errs
}
