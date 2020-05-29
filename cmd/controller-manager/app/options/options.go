package options

import (
	"flag"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/leaderelection"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	kubesphereconfig "kubesphere.io/kubesphere/pkg/apiserver/config"
	"kubesphere.io/kubesphere/pkg/simple/client/devops/jenkins"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/multicluster"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"strings"
	"time"
)

type KubeSphereControllerManagerOptions struct {
	KubernetesOptions   *k8s.KubernetesOptions
	DevopsOptions       *jenkins.Options
	S3Options           *s3.Options
	OpenPitrixOptions   *openpitrix.Options
	MultiClusterOptions *multicluster.Options
	LeaderElect         bool
	LeaderElection      *leaderelection.LeaderElectionConfig
}

func NewKubeSphereControllerManagerOptions() *KubeSphereControllerManagerOptions {
	s := &KubeSphereControllerManagerOptions{
		KubernetesOptions:   k8s.NewKubernetesOptions(),
		DevopsOptions:       jenkins.NewDevopsOptions(),
		S3Options:           s3.NewS3Options(),
		OpenPitrixOptions:   openpitrix.NewOptions(),
		MultiClusterOptions: multicluster.NewOptions(),
		LeaderElection: &leaderelection.LeaderElectionConfig{
			LeaseDuration: 30 * time.Second,
			RenewDeadline: 15 * time.Second,
			RetryPeriod:   5 * time.Second,
		},
		LeaderElect: false,
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

	s.KubernetesOptions.AddFlags(fss.FlagSet("kubernetes"), s.KubernetesOptions)
	s.DevopsOptions.AddFlags(fss.FlagSet("devops"), s.DevopsOptions)
	s.S3Options.AddFlags(fss.FlagSet("s3"), s.S3Options)
	s.OpenPitrixOptions.AddFlags(fss.FlagSet("openpitrix"), s.OpenPitrixOptions)
	s.MultiClusterOptions.AddFlags(fss.FlagSet("multicluster"), s.MultiClusterOptions)

	fs := fss.FlagSet("leaderelection")
	s.bindLeaderElectionFlags(s.LeaderElection, fs)

	fs.BoolVar(&s.LeaderElect, "leader-elect", s.LeaderElect, ""+
		"Whether to enable leader election. This field should be enabled when controller manager"+
		"deployed with multiple replicas.")

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
