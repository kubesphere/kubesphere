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
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/util/sets"
	kubeadmapiext "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1alpha1"
	"k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/validation"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/preflight"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	"k8s.io/kubernetes/pkg/util/initsystem"
	utilsexec "k8s.io/utils/exec"
)

var (
	crictlSandboxesParamsFormat = "%s -r %s sandboxes --quiet | xargs -r"
	crictlStopParamsFormat      = "%s -r %s stops %s"
	crictlRemoveParamsFormat    = "%s -r %s rms %s"
)

// NewCmdReset returns the "kubeadm reset" command
func NewCmdReset(in io.Reader, out io.Writer) *cobra.Command {
	var skipPreFlight bool
	var certsDir string
	var criSocketPath string
	var ignorePreflightErrors []string
	var forceReset bool

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Run this to revert any changes made to this host by 'kubeadm init' or 'kubeadm join'.",
		Run: func(cmd *cobra.Command, args []string) {
			ignorePreflightErrorsSet, err := validation.ValidateIgnorePreflightErrors(ignorePreflightErrors, skipPreFlight)
			kubeadmutil.CheckErr(err)

			r, err := NewReset(in, ignorePreflightErrorsSet, forceReset, certsDir, criSocketPath)
			kubeadmutil.CheckErr(err)
			kubeadmutil.CheckErr(r.Run(out))
		},
	}

	cmd.PersistentFlags().StringSliceVar(
		&ignorePreflightErrors, "ignore-preflight-errors", ignorePreflightErrors,
		"A list of checks whose errors will be shown as warnings. Example: 'IsPrivilegedUser,Swap'. Value 'all' ignores errors from all checks.",
	)
	cmd.PersistentFlags().BoolVar(
		&skipPreFlight, "skip-preflight-checks", false,
		"Skip preflight checks which normally run before modifying the system.",
	)
	cmd.PersistentFlags().MarkDeprecated("skip-preflight-checks", "it is now equivalent to --ignore-preflight-errors=all")

	cmd.PersistentFlags().StringVar(
		&certsDir, "cert-dir", kubeadmapiext.DefaultCertificatesDir,
		"The path to the directory where the certificates are stored. If specified, clean this directory.",
	)

	cmd.PersistentFlags().StringVar(
		&criSocketPath, "cri-socket", "/var/run/dockershim.sock",
		"The path to the CRI socket to use with crictl when cleaning up containers.",
	)

	cmd.PersistentFlags().BoolVar(
		&forceReset, "force", false,
		"Reset the node without prompting for confirmation.",
	)

	return cmd
}

// Reset defines struct used for kubeadm reset command
type Reset struct {
	certsDir      string
	criSocketPath string
}

// NewReset instantiate Reset struct
func NewReset(in io.Reader, ignorePreflightErrors sets.String, forceReset bool, certsDir, criSocketPath string) (*Reset, error) {
	if !forceReset {
		fmt.Println("[reset] WARNING: changes made to this host by 'kubeadm init' or 'kubeadm join' will be reverted.")
		fmt.Print("[reset] are you sure you want to proceed? [y/N]: ")
		s := bufio.NewScanner(in)
		s.Scan()
		if err := s.Err(); err != nil {
			return nil, err
		}
		if strings.ToLower(s.Text()) != "y" {
			return nil, errors.New("Aborted reset operation")
		}
	}

	glog.Infoln("[preflight] running pre-flight checks")
	if err := preflight.RunRootCheckOnly(ignorePreflightErrors); err != nil {
		return nil, err
	}

	return &Reset{
		certsDir:      certsDir,
		criSocketPath: criSocketPath,
	}, nil
}

// Run reverts any changes made to this host by "kubeadm init" or "kubeadm join".
func (r *Reset) Run(out io.Writer) error {

	// Try to stop the kubelet service
	glog.V(1).Infof("[reset] getting init system")
	initSystem, err := initsystem.GetInitSystem()
	if err != nil {
		glog.Warningln("[reset] the kubelet service could not be stopped by kubeadm. Unable to detect a supported init system!")
		glog.Warningln("[reset] please ensure kubelet is stopped manually")
	} else {
		glog.Infoln("[reset] stopping the kubelet service")
		if err := initSystem.ServiceStop("kubelet"); err != nil {
			glog.Warningf("[reset] the kubelet service could not be stopped by kubeadm: [%v]\n", err)
			glog.Warningln("[reset] please ensure kubelet is stopped manually")
		}
	}

	// Try to unmount mounted directories under /var/lib/kubelet in order to be able to remove the /var/lib/kubelet directory later
	glog.Infof("[reset] unmounting mounted directories in %q\n", "/var/lib/kubelet")
	umountDirsCmd := "awk '$2 ~ path {print $2}' path=/var/lib/kubelet /proc/mounts | xargs -r umount"

	glog.V(1).Infof("[reset] executing command %q", umountDirsCmd)
	umountOutputBytes, err := exec.Command("sh", "-c", umountDirsCmd).Output()
	if err != nil {
		glog.Errorf("[reset] failed to unmount mounted directories in /var/lib/kubelet: %s\n", string(umountOutputBytes))
	}

	glog.Infoln("[reset] removing kubernetes-managed containers")
	dockerCheck := preflight.ServiceCheck{Service: "docker", CheckIfActive: true}
	execer := utilsexec.New()

	reset(execer, dockerCheck, r.criSocketPath)

	dirsToClean := []string{"/var/lib/kubelet", "/etc/cni/net.d", "/var/lib/dockershim", "/var/run/kubernetes"}

	// Only clear etcd data when the etcd manifest is found. In case it is not found, we must assume that the user
	// provided external etcd endpoints. In that case, it is his own responsibility to reset etcd
	etcdManifestPath := filepath.Join(kubeadmconstants.KubernetesDir, kubeadmconstants.ManifestsSubDirName, "etcd.yaml")
	glog.V(1).Infof("[reset] checking for etcd manifest")
	if _, err := os.Stat(etcdManifestPath); err == nil {
		glog.V(1).Infof("Found one at %s", etcdManifestPath)
		dirsToClean = append(dirsToClean, "/var/lib/etcd")
	} else {
		glog.Infof("[reset] no etcd manifest found in %q. Assuming external etcd\n", etcdManifestPath)
	}

	// Then clean contents from the stateful kubelet, etcd and cni directories
	glog.Infof("[reset] deleting contents of stateful directories: %v\n", dirsToClean)
	for _, dir := range dirsToClean {
		glog.V(1).Infof("[reset] deleting content of %s", dir)
		cleanDir(dir)
	}

	// Remove contents from the config and pki directories
	glog.V(1).Infoln("[reset] removing contents from the config and pki directories")
	if r.certsDir != kubeadmapiext.DefaultCertificatesDir {
		glog.Warningf("[reset] WARNING: cleaning a non-default certificates directory: %q\n", r.certsDir)
	}
	resetConfigDir(kubeadmconstants.KubernetesDir, r.certsDir)

	return nil
}

func reset(execer utilsexec.Interface, dockerCheck preflight.Checker, criSocketPath string) {
	crictlPath, err := execer.LookPath("crictl")
	if err == nil {
		resetWithCrictl(execer, dockerCheck, criSocketPath, crictlPath)
	} else {
		resetWithDocker(execer, dockerCheck)
	}
}

func resetWithDocker(execer utilsexec.Interface, dockerCheck preflight.Checker) {
	if _, errors := dockerCheck.Check(); len(errors) == 0 {
		if err := execer.Command("sh", "-c", "docker ps -a --filter name=k8s_ -q | xargs -r docker rm --force --volumes").Run(); err != nil {
			glog.Errorln("[reset] Failed to stop the running containers")
		}
	} else {
		glog.Infoln("[reset] docker doesn't seem to be running. Skipping the removal of running Kubernetes containers")
	}
}

func resetWithCrictl(execer utilsexec.Interface, dockerCheck preflight.Checker, criSocketPath, crictlPath string) {
	if criSocketPath != "" {
		glog.Infof("[reset] cleaning up running containers using crictl with socket %s\n", criSocketPath)
		glog.V(1).Infoln("[reset] listing running pods using crictl")

		listcmd := fmt.Sprintf(crictlSandboxesParamsFormat, crictlPath, criSocketPath)
		glog.V(1).Infof("[reset] executing comand %q", listcmd)
		output, err := execer.Command(listcmd).CombinedOutput()
		if err != nil {
			glog.Infoln("[reset] failed to list running pods using crictl. Trying using docker instead")
			resetWithDocker(execer, dockerCheck)
			return
		}
		sandboxes := strings.Split(string(output), " ")
		glog.V(1).Infoln("[reset] stopping and removing running containers using crictl")
		for _, s := range sandboxes {
			stopcmd := fmt.Sprintf(crictlStopParamsFormat, crictlPath, criSocketPath, s)
			glog.V(1).Infof("[reset] executing command %q", stopcmd)
			if err := execer.Command(stopcmd).Run(); err != nil {
				glog.Infoln("[reset] failed to stop the running containers using crictl. Trying using docker instead")
				resetWithDocker(execer, dockerCheck)
				return
			}
			removecmd := fmt.Sprintf(crictlRemoveParamsFormat, crictlPath, criSocketPath, s)
			glog.V(1).Infof("[reset] executing command %q", removecmd)
			if err := execer.Command(removecmd).Run(); err != nil {
				glog.Infoln("[reset] failed to remove the running containers using crictl. Trying using docker instead")
				resetWithDocker(execer, dockerCheck)
				return
			}
		}
	} else {
		glog.Infoln("[reset] CRI socket path not provided for crictl. Trying docker instead")
		resetWithDocker(execer, dockerCheck)
	}
}

// cleanDir removes everything in a directory, but not the directory itself
func cleanDir(filePath string) error {
	// If the directory doesn't even exist there's nothing to do, and we do
	// not consider this an error
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	d, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		if err = os.RemoveAll(filepath.Join(filePath, name)); err != nil {
			return err
		}
	}
	return nil
}

// resetConfigDir is used to cleanup the files kubeadm writes in /etc/kubernetes/.
func resetConfigDir(configPathDir, pkiPathDir string) {
	dirsToClean := []string{
		filepath.Join(configPathDir, kubeadmconstants.ManifestsSubDirName),
		pkiPathDir,
	}
	glog.Infof("[reset] deleting contents of config directories: %v\n", dirsToClean)
	for _, dir := range dirsToClean {
		if err := cleanDir(dir); err != nil {
			glog.Errorf("[reset] failed to remove directory: %q [%v]\n", dir, err)
		}
	}

	filesToClean := []string{
		filepath.Join(configPathDir, kubeadmconstants.AdminKubeConfigFileName),
		filepath.Join(configPathDir, kubeadmconstants.KubeletKubeConfigFileName),
		filepath.Join(configPathDir, kubeadmconstants.KubeletBootstrapKubeConfigFileName),
		filepath.Join(configPathDir, kubeadmconstants.ControllerManagerKubeConfigFileName),
		filepath.Join(configPathDir, kubeadmconstants.SchedulerKubeConfigFileName),
	}
	glog.Infof("[reset] deleting files: %v\n", filesToClean)
	for _, path := range filesToClean {
		if err := os.RemoveAll(path); err != nil {
			glog.Errorf("[reset] failed to remove file: %q [%v]\n", path, err)
		}
	}
}
