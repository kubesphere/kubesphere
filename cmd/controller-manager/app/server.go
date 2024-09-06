/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package app

import (
	"context"
	"fmt"

	"github.com/google/gops/agent"
	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"kubesphere.io/kubesphere/cmd/controller-manager/app/options"
	"kubesphere.io/kubesphere/pkg/config"
	"kubesphere.io/kubesphere/pkg/controller"
	"kubesphere.io/kubesphere/pkg/controller/application"
	"kubesphere.io/kubesphere/pkg/controller/certificatesigningrequest"
	"kubesphere.io/kubesphere/pkg/controller/cluster"
	"kubesphere.io/kubesphere/pkg/controller/clusterlabel"
	"kubesphere.io/kubesphere/pkg/controller/clusterrole"
	"kubesphere.io/kubesphere/pkg/controller/clusterrolebinding"
	ksconfig "kubesphere.io/kubesphere/pkg/controller/config"
	"kubesphere.io/kubesphere/pkg/controller/conversion"
	"kubesphere.io/kubesphere/pkg/controller/core"
	"kubesphere.io/kubesphere/pkg/controller/extension"
	"kubesphere.io/kubesphere/pkg/controller/globalrole"
	"kubesphere.io/kubesphere/pkg/controller/globalrolebinding"
	"kubesphere.io/kubesphere/pkg/controller/job"
	"kubesphere.io/kubesphere/pkg/controller/k8sapplication"
	"kubesphere.io/kubesphere/pkg/controller/ksserviceaccount"
	"kubesphere.io/kubesphere/pkg/controller/kubeconfig"
	"kubesphere.io/kubesphere/pkg/controller/kubectl"
	"kubesphere.io/kubesphere/pkg/controller/loginrecord"
	"kubesphere.io/kubesphere/pkg/controller/namespace"
	"kubesphere.io/kubesphere/pkg/controller/quota"
	"kubesphere.io/kubesphere/pkg/controller/role"
	"kubesphere.io/kubesphere/pkg/controller/rolebinding"
	"kubesphere.io/kubesphere/pkg/controller/roletemplate"
	"kubesphere.io/kubesphere/pkg/controller/secret"
	"kubesphere.io/kubesphere/pkg/controller/serviceaccount"
	"kubesphere.io/kubesphere/pkg/controller/storageclass"
	"kubesphere.io/kubesphere/pkg/controller/telemetry"
	"kubesphere.io/kubesphere/pkg/controller/user"
	"kubesphere.io/kubesphere/pkg/controller/workspace"
	"kubesphere.io/kubesphere/pkg/controller/workspacerole"
	"kubesphere.io/kubesphere/pkg/controller/workspacerolebinding"
	"kubesphere.io/kubesphere/pkg/controller/workspacetemplate"
	"kubesphere.io/kubesphere/pkg/utils/term"
	"kubesphere.io/kubesphere/pkg/version"
)

func init() {
	// core
	runtime.Must(controller.Register(&core.ExtensionReconciler{}))
	runtime.Must(controller.Register(&core.CategoryReconciler{}))
	runtime.Must(controller.Register(&core.RepositoryReconciler{}))
	runtime.Must(controller.Register(&core.InstallPlanReconciler{}))
	runtime.Must(controller.Register(&core.InstallPlanWebhook{}))
	// extension
	runtime.Must(controller.Register(&extension.JSBundleWebhook{}))
	runtime.Must(controller.Register(&extension.APIServiceWebhook{}))
	runtime.Must(controller.Register(&extension.ReverseProxyWebhook{}))
	runtime.Must(controller.Register(&extension.ExtensionEntryWebhook{}))
	// rbac
	runtime.Must(controller.Register(&globalrole.Reconciler{}))
	runtime.Must(controller.Register(&globalrolebinding.Reconciler{}))
	runtime.Must(controller.Register(&workspacerole.Reconciler{}))
	runtime.Must(controller.Register(&workspacerolebinding.Reconciler{}))
	runtime.Must(controller.Register(&clusterrole.Reconciler{}))
	runtime.Must(controller.Register(&clusterrolebinding.Reconciler{}))
	runtime.Must(controller.Register(&role.Reconciler{}))
	runtime.Must(controller.Register(&rolebinding.Reconciler{}))
	runtime.Must(controller.Register(&roletemplate.Reconciler{}))
	runtime.Must(controller.Register(&namespace.Reconciler{}))
	// user management
	runtime.Must(controller.Register(&user.Reconciler{}))
	runtime.Must(controller.Register(&user.Webhook{}))
	runtime.Must(controller.Register(&loginrecord.Reconciler{}))
	// multi cluster
	runtime.Must(controller.Register(&cluster.Reconciler{}))
	runtime.Must(controller.Register(&cluster.Webhook{}))
	runtime.Must(controller.Register(&clusterlabel.Reconciler{}))
	// multi tenancy
	runtime.Must(controller.Register(&workspace.Reconciler{}))
	runtime.Must(controller.Register(&workspacetemplate.Reconciler{}))
	// kubesphere service account
	runtime.Must(controller.Register(&ksserviceaccount.Reconciler{}))
	runtime.Must(controller.Register(&ksserviceaccount.Webhook{}))
	runtime.Must(controller.Register(&secret.ServiceAccountSecretReconciler{}))
	// additional capabilities
	runtime.Must(controller.Register(&serviceaccount.Reconciler{}))
	runtime.Must(controller.Register(&job.Reconciler{}))
	runtime.Must(controller.Register(&storageclass.Reconciler{}))
	runtime.Must(controller.Register(&telemetry.Runnable{}))
	runtime.Must(controller.Register(&ksconfig.Webhook{}))
	runtime.Must(controller.Register(&conversion.Webhook{}))
	// kubeconfig
	runtime.Must(controller.Register(&kubeconfig.Reconciler{}))
	runtime.Must(controller.Register(&certificatesigningrequest.Reconciler{}))
	// resource quota
	runtime.Must(controller.Register(&quota.Reconciler{}))
	runtime.Must(controller.Register(&quota.Webhook{}))
	// app store
	runtime.Must(controller.Register(&application.AppReleaseReconciler{}))
	runtime.Must(controller.Register(&application.RepoReconciler{}))
	runtime.Must(controller.Register(&application.AppCategoryReconciler{}))
	runtime.Must(controller.Register(&application.AppVersionReconciler{}))
	// k8s application
	runtime.Must(controller.Register(&k8sapplication.Reconciler{}))
	// kubectl
	runtime.Must(controller.Register(&kubectl.Reconciler{}))
}

func NewControllerManagerCommand() *cobra.Command {
	s := options.NewControllerManagerOptions()
	if conf, err := config.TryLoadFromDisk(); err == nil {
		s.Merge(conf)
	} else {
		klog.Fatalf("Failed to load configuration from disk: %v", err)
	}

	cmd := &cobra.Command{
		Use:  "controller-manager",
		Long: `KubeSphere controller manager is a daemon that embeds the control loops shipped with KubeSphere.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if errs := s.Validate(); len(errs) != 0 {
				return utilerrors.NewAggregate(errs)
			}
			if s.DebugMode {
				// Add agent to report additional information such as the current stack trace, Go version, memory stats, etc.
				// Bind to a random port on address 127.0.0.1
				if err := agent.Listen(agent.Options{}); err != nil {
					klog.Fatalln(err)
				}
			}
			return Run(signals.SetupSignalHandler(), s)
		},
		SilenceUsage: true,
	}

	namedFlagSets := s.Flags()

	for _, f := range namedFlagSets.FlagSets {
		cmd.Flags().AddFlagSet(f)
	}

	usageFmt := "Usage:\n  %s\n"
	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStdout(), namedFlagSets, cols)
	})

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of KubeSphere controller-manager",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(version.Get())
		},
	}

	cmd.AddCommand(versionCmd)
	return cmd
}

func Run(ctx context.Context, s *options.ControllerManagerOptions) error {
	cm, err := s.NewControllerManager()
	if err != nil {
		return fmt.Errorf("failed to create controller manager: %v", err)
	}
	if err := cm.Run(ctx, controller.Controllers); err != nil {
		return fmt.Errorf("failed to run controller manager: %v", err)
	}
	return nil
}
