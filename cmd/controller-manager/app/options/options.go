package options

import (
	"flag"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/leaderelection"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	kubesphereconfig "kubesphere.io/kubesphere/pkg/server/config"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"kubesphere.io/kubesphere/pkg/simple/client/s2is3"
	"strings"
	"time"
)

type KubeSphereControllerManagerOptions struct {
	KubernetesOptions *k8s.KubernetesOptions
	DevopsOptions     *devops.DevopsOptions
	S3Options         *s2is3.S3Options
	OpenPitrixOptions *openpitrix.OpenPitrixOptions

	LeaderElection *leaderelection.LeaderElectionConfig
}

func NewKubeSphereControllerManagerOptions() *KubeSphereControllerManagerOptions {
	s := &KubeSphereControllerManagerOptions{
		KubernetesOptions: k8s.NewKubernetesOptions(),
		DevopsOptions:     devops.NewDevopsOptions(),
		S3Options:         s2is3.NewS3Options(),
		OpenPitrixOptions: openpitrix.NewOpenPitrixOptions(),
		LeaderElection: &leaderelection.LeaderElectionConfig{
			LeaseDuration: 30 * time.Second,
			RenewDeadline: 15 * time.Second,
			RetryPeriod:   5 * time.Second,
		},
	}

	return s
}

func (s *KubeSphereControllerManagerOptions) ApplyTo(conf *kubesphereconfig.Config) {
	s.S3Options.ApplyTo(conf.S3Options)
	s.KubernetesOptions.ApplyTo(conf.KubernetesOptions)
	s.DevopsOptions.ApplyTo(conf.DevopsOptions)
	s.OpenPitrixOptions.ApplyTo(conf.OpenPitrixOptions)
}

func (s *KubeSphereControllerManagerOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}

	s.KubernetesOptions.AddFlags(fss.FlagSet("kubernetes"))
	s.DevopsOptions.AddFlags(fss.FlagSet("devops"))
	s.S3Options.AddFlags(fss.FlagSet("s3"))
	s.OpenPitrixOptions.AddFlags(fss.FlagSet("openpitrix"))

	fs := fss.FlagSet("leaderelection")
	s.bindLeaderElectionFlags(s.LeaderElection, fs)

	kfs := fss.FlagSet("klog")
	local := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		kfs.AddGoFlag(fl)
	})

	return fss
}

func (s *KubeSphereControllerManagerOptions) Validate() []error {
	var errs []error
	errs = append(errs, s.DevopsOptions.Validate()...)
	errs = append(errs, s.KubernetesOptions.Validate()...)
	errs = append(errs, s.S3Options.Validate()...)
	errs = append(errs, s.OpenPitrixOptions.Validate()...)
	return errs
}

func (s *KubeSphereControllerManagerOptions) bindLeaderElectionFlags(l *leaderelection.LeaderElectionConfig, fs *pflag.FlagSet) {
	fs.DurationVar(&l.LeaseDuration, "leader-elect-lease-duration", l.LeaseDuration, ""+
		"The duration that non-leader candidates will wait after observing a leadership "+
		"renewal until attempting to acquire leadership of a led but unrenewed leader "+
		"slot. This is effectively the maximum duration that a leader can be stopped "+
		"before it is replaced by another candidate. This is only applicable if leader "+
		"election is enabled.")
	fs.DurationVar(&l.RenewDeadline, "leader-elect-renew-deadline", l.RenewDeadline, ""+
		"The interval between attempts by the acting master to renew a leadership slot "+
		"before it stops leading. This must be less than or equal to the lease duration. "+
		"This is only applicable if leader election is enabled.")
	fs.DurationVar(&l.RetryPeriod, "leader-elect-retry-period", l.RetryPeriod, ""+
		"The duration the clients should wait between attempting acquisition and renewal "+
		"of a leadership. This is only applicable if leader election is enabled.")
}
