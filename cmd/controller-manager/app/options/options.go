package options

import (
	"flag"
	"github.com/spf13/pflag"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiserverconfig "k8s.io/apiserver/pkg/apis/config"
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

	LeaderElection *apiserverconfig.LeaderElectionConfiguration
}

func NewKubeSphereControllerManagerOptions() *KubeSphereControllerManagerOptions {
	s := &KubeSphereControllerManagerOptions{
		KubernetesOptions: k8s.NewKubernetesOptions(),
		DevopsOptions:     devops.NewDevopsOptions(),
		S3Options:         s2is3.NewS3Options(),
		OpenPitrixOptions: openpitrix.NewOpenPitrixOptions(),
		LeaderElection: &apiserverconfig.LeaderElectionConfiguration{
			LeaderElect:   false,
			LeaseDuration: v1.Duration{Duration: 30 * time.Second},
			RenewDeadline: v1.Duration{Duration: 15 * time.Second},
			RetryPeriod:   v1.Duration{Duration: 5 * time.Second},
			ResourceLock:  "ks-controller-manager-leader-election",
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

func (s *KubeSphereControllerManagerOptions) bindLeaderElectionFlags(l *apiserverconfig.LeaderElectionConfiguration, fs *pflag.FlagSet) {
	fs.BoolVar(&l.LeaderElect, "leader-elect", l.LeaderElect, ""+
		"Start a leader election client and gain leadership before "+
		"executing the main loop. Enable this when running replicated "+
		"components for high availability.")
	fs.DurationVar(&l.LeaseDuration.Duration, "leader-elect-lease-duration", l.LeaseDuration.Duration, ""+
		"The duration that non-leader candidates will wait after observing a leadership "+
		"renewal until attempting to acquire leadership of a led but unrenewed leader "+
		"slot. This is effectively the maximum duration that a leader can be stopped "+
		"before it is replaced by another candidate. This is only applicable if leader "+
		"election is enabled.")
	fs.DurationVar(&l.RenewDeadline.Duration, "leader-elect-renew-deadline", l.RenewDeadline.Duration, ""+
		"The interval between attempts by the acting master to renew a leadership slot "+
		"before it stops leading. This must be less than or equal to the lease duration. "+
		"This is only applicable if leader election is enabled.")
	fs.DurationVar(&l.RetryPeriod.Duration, "leader-elect-retry-period", l.RetryPeriod.Duration, ""+
		"The duration the clients should wait between attempting acquisition and renewal "+
		"of a leadership. This is only applicable if leader election is enabled.")
	fs.StringVar(&l.ResourceLock, "leader-elect-resource-lock", l.ResourceLock, ""+
		"The type of resource object that is used for locking during "+
		"leader election. Supported options are `endpoints` (default) and `configmaps`.")
}
