/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"github.com/renstrom/dedent"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	certutil "k8s.io/client-go/util/cert"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmapiext "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1alpha1"
	"k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/validation"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/discovery"
	"k8s.io/kubernetes/cmd/kubeadm/app/features"
	kubeletphase "k8s.io/kubernetes/cmd/kubeadm/app/phases/kubelet"
	"k8s.io/kubernetes/cmd/kubeadm/app/preflight"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	kubeconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	nodeutil "k8s.io/kubernetes/pkg/util/node"
	utilsexec "k8s.io/utils/exec"
)

var (
	joinDoneMsgf = dedent.Dedent(`
		This node has joined the cluster:
		* Certificate signing request was sent to master and a response
		  was received.
		* The Kubelet was informed of the new secure connection details.

		Run 'kubectl get nodes' on the master to see this node join the cluster.
		`)

	joinLongDescription = dedent.Dedent(`
		When joining a kubeadm initialized cluster, we need to establish
		bidirectional trust. This is split into discovery (having the Node
		trust the Kubernetes Master) and TLS bootstrap (having the Kubernetes
		Master trust the Node).

		There are 2 main schemes for discovery. The first is to use a shared
		token along with the IP address of the API server. The second is to
		provide a file - a subset of the standard kubeconfig file. This file
		can be a local file or downloaded via an HTTPS URL. The forms are
		kubeadm join --discovery-token abcdef.1234567890abcdef 1.2.3.4:6443,
		kubeadm join --discovery-file path/to/file.conf, or kubeadm join
		--discovery-file https://url/file.conf. Only one form can be used. If
		the discovery information is loaded from a URL, HTTPS must be used.
		Also, in that case the host installed CA bundle is used to verify
		the connection.

		If you use a shared token for discovery, you should also pass the
		--discovery-token-ca-cert-hash flag to validate the public key of the
		root certificate authority (CA) presented by the Kubernetes Master. The
		value of this flag is specified as "<hash-type>:<hex-encoded-value>",
		where the supported hash type is "sha256". The hash is calculated over
		the bytes of the Subject Public Key Info (SPKI) object (as in RFC7469).
		This value is available in the output of "kubeadm init" or can be
		calcuated using standard tools. The --discovery-token-ca-cert-hash flag
		may be repeated multiple times to allow more than one public key.

		If you cannot know the CA public key hash ahead of time, you can pass
		the --discovery-token-unsafe-skip-ca-verification flag to disable this
		verification. This weakens the kubeadm security model since other nodes
		can potentially impersonate the Kubernetes Master.

		The TLS bootstrap mechanism is also driven via a shared token. This is
		used to temporarily authenticate with the Kubernetes Master to submit a
		certificate signing request (CSR) for a locally created key pair. By
		default, kubeadm will set up the Kubernetes Master to automatically
		approve these signing requests. This token is passed in with the
		--tls-bootstrap-token abcdef.1234567890abcdef flag.

		Often times the same token is used for both parts. In this case, the
		--token flag can be used instead of specifying each token individually.
		`)
)

// NewCmdJoin returns "kubeadm join" command.
func NewCmdJoin(out io.Writer) *cobra.Command {
	cfg := &kubeadmapiext.NodeConfiguration{}
	legacyscheme.Scheme.Default(cfg)

	var skipPreFlight bool
	var cfgPath string
	var featureGatesString string
	var ignorePreflightErrors []string

	cmd := &cobra.Command{
		Use:   "join",
		Short: "Run this on any machine you wish to join an existing cluster",
		Long:  joinLongDescription,
		Run: func(cmd *cobra.Command, args []string) {
			j, err := NewValidJoin(cfg, args, skipPreFlight, cfgPath, featureGatesString, ignorePreflightErrors)
			kubeadmutil.CheckErr(err)
			kubeadmutil.CheckErr(j.Validate(cmd))
			kubeadmutil.CheckErr(j.Run(out))
		},
	}

	AddJoinConfigFlags(cmd.PersistentFlags(), cfg, &featureGatesString)
	AddJoinOtherFlags(cmd.PersistentFlags(), &cfgPath, &skipPreFlight, &ignorePreflightErrors)

	return cmd
}

// NewValidJoin validates the command line that are passed to the cobra command
func NewValidJoin(cfg *kubeadmapiext.NodeConfiguration, args []string, skipPreFlight bool, cfgPath, featureGatesString string, ignorePreflightErrors []string) (*Join, error) {
	cfg.DiscoveryTokenAPIServers = args

	var err error
	if cfg.FeatureGates, err = features.NewFeatureGate(&features.InitFeatureGates, featureGatesString); err != nil {
		return nil, err
	}

	legacyscheme.Scheme.Default(cfg)
	internalcfg := &kubeadmapi.NodeConfiguration{}
	legacyscheme.Scheme.Convert(cfg, internalcfg, nil)

	ignorePreflightErrorsSet, err := validation.ValidateIgnorePreflightErrors(ignorePreflightErrors, skipPreFlight)
	if err != nil {
		return nil, err
	}

	return NewJoin(cfgPath, args, internalcfg, ignorePreflightErrorsSet)
}

// AddJoinConfigFlags adds join flags bound to the config to the specified flagset
func AddJoinConfigFlags(flagSet *flag.FlagSet, cfg *kubeadmapiext.NodeConfiguration, featureGatesString *string) {
	flagSet.StringVar(
		&cfg.DiscoveryFile, "discovery-file", "",
		"A file or url from which to load cluster information.")
	flagSet.StringVar(
		&cfg.DiscoveryToken, "discovery-token", "",
		"A token used to validate cluster information fetched from the master.")
	flagSet.StringVar(
		&cfg.NodeName, "node-name", "",
		"Specify the node name.")
	flagSet.StringVar(
		&cfg.TLSBootstrapToken, "tls-bootstrap-token", "",
		"A token used for TLS bootstrapping.")
	flagSet.StringSliceVar(
		&cfg.DiscoveryTokenCACertHashes, "discovery-token-ca-cert-hash", []string{},
		"For token-based discovery, validate that the root CA public key matches this hash (format: \"<type>:<value>\").")
	flagSet.BoolVar(
		&cfg.DiscoveryTokenUnsafeSkipCAVerification, "discovery-token-unsafe-skip-ca-verification", false,
		"For token-based discovery, allow joining without --discovery-token-ca-cert-hash pinning.")
	flagSet.StringVar(
		&cfg.Token, "token", "",
		"Use this token for both discovery-token and tls-bootstrap-token.")
	flagSet.StringVar(
		featureGatesString, "feature-gates", *featureGatesString,
		"A set of key=value pairs that describe feature gates for various features. "+
			"Options are:\n"+strings.Join(features.KnownFeatures(&features.InitFeatureGates), "\n"))
	flagSet.StringVar(
		&cfg.CRISocket, "cri-socket", cfg.CRISocket,
		`Specify the CRI socket to connect to.`,
	)
}

// AddJoinOtherFlags adds join flags that are not bound to a configuration file to the given flagset
func AddJoinOtherFlags(flagSet *flag.FlagSet, cfgPath *string, skipPreFlight *bool, ignorePreflightErrors *[]string) {
	flagSet.StringVar(
		cfgPath, "config", *cfgPath,
		"Path to kubeadm config file.")

	flagSet.StringSliceVar(
		ignorePreflightErrors, "ignore-preflight-errors", *ignorePreflightErrors,
		"A list of checks whose errors will be shown as warnings. Example: 'IsPrivilegedUser,Swap'. Value 'all' ignores errors from all checks.",
	)
	flagSet.BoolVar(
		skipPreFlight, "skip-preflight-checks", false,
		"Skip preflight checks which normally run before modifying the system.",
	)
	flagSet.MarkDeprecated("skip-preflight-checks", "it is now equivalent to --ignore-preflight-errors=all")
}

// Join defines struct used by kubeadm join command
type Join struct {
	cfg *kubeadmapi.NodeConfiguration
}

// NewJoin instantiates Join struct with given arguments
func NewJoin(cfgPath string, args []string, cfg *kubeadmapi.NodeConfiguration, ignorePreflightErrors sets.String) (*Join, error) {

	if cfg.NodeName == "" {
		glog.V(1).Infoln("[join] found NodeName empty")
		glog.V(1).Infoln("[join] considered OS hostname as NodeName")
		cfg.NodeName = nodeutil.GetHostname("")
	}

	if cfgPath != "" {
		glog.V(1).Infoln("[join] reading configuration from", cfgPath)
		b, err := ioutil.ReadFile(cfgPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read config from %q [%v]", cfgPath, err)
		}
		glog.V(1).Infoln("[join] decoding configuration information")
		if err := runtime.DecodeInto(legacyscheme.Codecs.UniversalDecoder(), b, cfg); err != nil {
			return nil, fmt.Errorf("unable to decode config from %q [%v]", cfgPath, err)
		}
	}

	glog.Infoln("[preflight] running pre-flight checks")

	// Then continue with the others...
	glog.V(1).Infoln("[preflight] running various checks on all nodes")
	if err := preflight.RunJoinNodeChecks(utilsexec.New(), cfg, ignorePreflightErrors); err != nil {
		return nil, err
	}

	// Try to start the kubelet service in case it's inactive
	glog.V(1).Infoln("[preflight] starting kubelet service if it's inactive")
	preflight.TryStartKubelet(ignorePreflightErrors)

	return &Join{cfg: cfg}, nil
}

// Validate validates mixed arguments passed to cobra.Command
func (j *Join) Validate(cmd *cobra.Command) error {
	if err := validation.ValidateMixedArguments(cmd.PersistentFlags()); err != nil {
		return err
	}
	return validation.ValidateNodeConfiguration(j.cfg).ToAggregate()
}

// Run executes worker node provisioning and tries to join an existing cluster.
func (j *Join) Run(out io.Writer) error {
	glog.V(1).Infoln("[join] retrieving KubeConfig objects")
	cfg, err := discovery.For(j.cfg)
	if err != nil {
		return err
	}

	kubeconfigFile := filepath.Join(kubeadmconstants.KubernetesDir, kubeadmconstants.KubeletBootstrapKubeConfigFileName)

	// Write the bootstrap kubelet config file or the TLS-Boostrapped kubelet config file down to disk
	glog.V(1).Infoln("[join] writing bootstrap kubelet config file at", kubeconfigFile)
	if err := kubeconfigutil.WriteToDisk(kubeconfigFile, cfg); err != nil {
		return fmt.Errorf("couldn't save bootstrap-kubelet.conf to disk: %v", err)
	}

	// Write the ca certificate to disk so kubelet can use it for authentication
	cluster := cfg.Contexts[cfg.CurrentContext].Cluster
	err = certutil.WriteCert(j.cfg.CACertPath, cfg.Clusters[cluster].CertificateAuthorityData)
	if err != nil {
		return fmt.Errorf("couldn't save the CA certificate to disk: %v", err)
	}

	// NOTE: flag "--dynamic-config-dir" should be specified in /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
	glog.V(1).Infoln("[join] consuming base kubelet configuration")
	if features.Enabled(j.cfg.FeatureGates, features.DynamicKubeletConfig) {
		if err := kubeletphase.ConsumeBaseKubeletConfiguration(j.cfg.NodeName); err != nil {
			return fmt.Errorf("error consuming base kubelet configuration: %v", err)
		}
	}

	fmt.Fprintf(out, joinDoneMsgf)
	return nil
}
