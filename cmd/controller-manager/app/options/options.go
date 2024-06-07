/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package options

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/leaderelection"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"kubesphere.io/kubesphere/pkg/config"
	"kubesphere.io/kubesphere/pkg/controller"
	"kubesphere.io/kubesphere/pkg/controller/options"
	"kubesphere.io/kubesphere/pkg/scheme"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/utils/clusterclient"
)

type ControllerManagerOptions struct {
	options.Options
	LeaderElect    bool
	LeaderElection *leaderelection.LeaderElectionConfig
	WebhookCertDir string
	// ControllerGates is the list of controller gates to enable or disable controller.
	// '*' means "all enabled by default controllers"
	// 'foo' means "enable 'foo'"
	// '-foo' means "disable 'foo'"
	// first item for a particular name wins.
	//     e.g. '-foo,foo' means "disable foo", 'foo,-foo' means "enable foo"
	// * has the lowest priority.
	//     e.g. *,-foo, means "disable 'foo'"
	ControllerGates []string

	DebugMode bool
}

func NewControllerManagerOptions() *ControllerManagerOptions {
	return &ControllerManagerOptions{
		LeaderElection: &leaderelection.LeaderElectionConfig{
			LeaseDuration: 30 * time.Second,
			RenewDeadline: 15 * time.Second,
			RetryPeriod:   5 * time.Second,
		},
		LeaderElect:     false,
		WebhookCertDir:  "",
		ControllerGates: []string{"*"},
	}
}

func (s *ControllerManagerOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}

	s.KubernetesOptions.AddFlags(fss.FlagSet("kubernetes"), s.KubernetesOptions)
	s.AuthenticationOptions.AddFlags(fss.FlagSet("authentication"), s.AuthenticationOptions)
	s.MultiClusterOptions.AddFlags(fss.FlagSet("multicluster"), s.MultiClusterOptions)
	fs := fss.FlagSet("leaderelection")
	s.bindLeaderElectionFlags(s.LeaderElection, fs)

	fs.BoolVar(&s.LeaderElect, "leader-elect", s.LeaderElect, ""+
		"Whether to enable leader election. This field should be enabled when controller manager"+
		"deployed with multiple replicas.")

	fs.StringVar(&s.WebhookCertDir, "webhook-cert-dir", s.WebhookCertDir, ""+
		"Certificate directory used to setup webhooks, need tls.crt and tls.key placed inside."+
		"if not set, webhook server would look up the server key and certificate in"+
		"{TempDir}/k8s-webhook-server/serving-certs")

	gfs := fss.FlagSet("generic")
	gfs.StringSliceVar(&s.ControllerGates, "controllers", []string{"*"}, fmt.Sprintf(""+
		"A list of controllers to enable. '*' enables all on-by-default controllers, 'foo' enables the controller "+
		"named 'foo', '-foo' disables the controller named 'foo'.\nAll controllers: %s",
		strings.Join(controller.Controllers.Keys(), ", ")))

	gfs.BoolVar(&s.DebugMode, "debug", false, "Don't enable this if you don't know what it means.")

	kfs := fss.FlagSet("klog")
	local := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		kfs.AddGoFlag(fl)
	})

	return fss
}

// Validate Options and Genetic Options
func (s *ControllerManagerOptions) Validate() []error {
	var errs []error
	errs = append(errs, s.KubernetesOptions.Validate()...)
	errs = append(errs, s.MultiClusterOptions.Validate()...)
	errs = append(errs, s.ComposedAppOptions.Validate()...)

	// genetic option: controllers, check all selectors are valid
	allControllersNameSet := sets.KeySet(controller.Controllers)
	for _, selector := range s.ControllerGates {
		if selector == "*" {
			continue
		}
		selector = strings.TrimPrefix(selector, "-")
		if !allControllersNameSet.Has(selector) {
			errs = append(errs, fmt.Errorf("%q is not in the list of known controllers", selector))
		}
	}
	return errs
}

func (s *ControllerManagerOptions) bindLeaderElectionFlags(l *leaderelection.LeaderElectionConfig, fs *pflag.FlagSet) {
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

// Merge new config without validation
// When misconfigured, the app should just crash directly
func (s *ControllerManagerOptions) Merge(conf *config.Config) {
	if conf == nil {
		return
	}
	if conf.KubernetesOptions != nil {
		s.KubernetesOptions = conf.KubernetesOptions
	}
	if conf.AuthenticationOptions != nil {
		s.AuthenticationOptions = conf.AuthenticationOptions
	}
	if conf.MultiClusterOptions != nil {
		s.MultiClusterOptions = conf.MultiClusterOptions
	}
	if conf.TerminalOptions != nil {
		s.TerminalOptions = conf.TerminalOptions
	}
	if conf.TelemetryOptions != nil {
		s.TelemetryOptions = conf.TelemetryOptions
	}
	if conf.HelmExecutorOptions != nil {
		s.HelmExecutorOptions = conf.HelmExecutorOptions
	}
	if conf.ExtensionOptions != nil {
		s.ExtensionOptions = conf.ExtensionOptions
	}
	if conf.KubeSphereOptions != nil {
		s.KubeSphereOptions = conf.KubeSphereOptions
	}
	if conf.ComposedAppOptions != nil {
		s.ComposedAppOptions = conf.ComposedAppOptions
	}
	if conf.S3Options != nil {
		s.S3Options = conf.S3Options
	}
}

func (s *ControllerManagerOptions) NewControllerManager() (*controller.Manager, error) {
	cm := &controller.Manager{}

	webhookServer := webhook.NewServer(webhook.Options{
		CertDir: s.WebhookCertDir,
		Port:    8443,
	})

	cmOptions := manager.Options{
		Scheme:        scheme.Scheme,
		WebhookServer: webhookServer,
	}

	if s.LeaderElect {
		cmOptions = manager.Options{
			Scheme:                  scheme.Scheme,
			WebhookServer:           webhookServer,
			LeaderElection:          s.LeaderElect,
			LeaderElectionNamespace: "kubesphere-system",
			LeaderElectionID:        "ks-controller-manager-leader-election",
			LeaseDuration:           &s.LeaderElection.LeaseDuration,
			RetryPeriod:             &s.LeaderElection.RetryPeriod,
			RenewDeadline:           &s.LeaderElection.RenewDeadline,
		}
	}

	k8sClient, err := k8s.NewKubernetesClient(s.KubernetesOptions)
	if err != nil {
		return nil, fmt.Errorf("unable to create kubernetes client: %v", err)
	}
	k8sVersionInfo, err := k8sClient.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch k8s version info: %v", err)
	}
	k8sVersion, err := semver.NewVersion(k8sVersionInfo.GitVersion)
	if err != nil {
		return nil, err
	}

	klog.V(0).Info("setting up manager")
	ctrl.SetLogger(klog.NewKlogr())
	// Use 8443 instead of 443 because we need root permission to bind port 443
	mgr, err := manager.New(k8sClient.Config(), cmOptions)
	if err != nil {
		klog.Fatalf("unable to set up overall controller manager: %v", err)
	}

	clusterClient, err := clusterclient.NewClusterClientSet(mgr.GetCache())
	if err != nil {
		return nil, fmt.Errorf("unable to create cluster client: %v", err)
	}

	cm.K8sClient = k8sClient
	cm.ClusterClient = clusterClient
	cm.Options = s.Options
	cm.IsControllerEnabled = s.IsControllerEnabled
	cm.Manager = mgr
	cm.K8sVersion = k8sVersion
	return cm, nil
}

// IsControllerEnabled check if a specified controller enabled or not.
func (s *ControllerManagerOptions) IsControllerEnabled(name string) bool {
	allowedAll := false
	for _, controllerGate := range s.ControllerGates {
		if controllerGate == name {
			return true
		}
		if controllerGate == "-"+name {
			return false
		}
		if controllerGate == "*" {
			allowedAll = true
		}
	}
	return allowedAll
}
