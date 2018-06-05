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

package remote

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang/glog"

	"k8s.io/kubernetes/test/e2e_node/builder"
	"k8s.io/kubernetes/test/utils"
)

const (
	systemSpecPath = "test/e2e_node/system/specs"
)

// NodeE2ERemote contains the specific functions in the node e2e test suite.
type NodeE2ERemote struct{}

func InitNodeE2ERemote() TestSuite {
	// TODO: Register flags.
	return &NodeE2ERemote{}
}

// SetupTestPackage sets up the test package with binaries k8s required for node e2e tests
func (n *NodeE2ERemote) SetupTestPackage(tardir, systemSpecName string) error {
	// Build the executables
	if err := builder.BuildGo(); err != nil {
		return fmt.Errorf("failed to build the dependencies: %v", err)
	}

	// Make sure we can find the newly built binaries
	buildOutputDir, err := utils.GetK8sBuildOutputDir()
	if err != nil {
		return fmt.Errorf("failed to locate kubernetes build output directory: %v", err)
	}

	rootDir, err := utils.GetK8sRootDir()
	if err != nil {
		return fmt.Errorf("failed to locate kubernetes root directory: %v", err)
	}

	// Copy binaries
	requiredBins := []string{"kubelet", "e2e_node.test", "ginkgo", "mounter"}
	for _, bin := range requiredBins {
		source := filepath.Join(buildOutputDir, bin)
		if _, err := os.Stat(source); err != nil {
			return fmt.Errorf("failed to locate test binary %s: %v", bin, err)
		}
		out, err := exec.Command("cp", source, filepath.Join(tardir, bin)).CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to copy %q: %v Output: %q", bin, err, out)
		}
	}

	if systemSpecName != "" {
		// Copy system spec file
		source := filepath.Join(rootDir, systemSpecPath, systemSpecName+".yaml")
		if _, err := os.Stat(source); err != nil {
			return fmt.Errorf("failed to locate system spec %q: %v", source, err)
		}
		out, err := exec.Command("cp", source, tardir).CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to copy system spec %q: %v, output: %q", source, err, out)
		}
	}

	return nil
}

// dest is relative to the root of the tar
func tarAddFile(tar, source, dest string) error {
	dir := filepath.Dir(dest)
	tardir := filepath.Join(tar, dir)
	tardest := filepath.Join(tar, dest)

	out, err := exec.Command("mkdir", "-p", tardir).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create archive bin subdir %q, was dest for file %q. Err: %v. Output:\n%s", tardir, source, err, out)
	}
	out, err = exec.Command("cp", source, tardest).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to copy file %q to the archive bin subdir %q. Err: %v. Output:\n%s", source, tardir, err, out)
	}
	return nil
}

// prependCOSMounterFlag prepends the flag for setting the GCI mounter path to
// args and returns the result.
func prependCOSMounterFlag(args, host, workspace string) (string, error) {
	glog.V(2).Infof("GCI/COS node and GCI/COS mounter both detected, modifying --experimental-mounter-path accordingly")
	mounterPath := filepath.Join(workspace, "mounter")
	args = fmt.Sprintf("--kubelet-flags=--experimental-mounter-path=%s ", mounterPath) + args
	return args, nil
}

// prependMemcgNotificationFlag prepends the flag for enabling memcg
// notification to args and returns the result.
func prependMemcgNotificationFlag(args string) string {
	return "--kubelet-flags=--experimental-kernel-memcg-notification=true " + args
}

// updateOSSpecificKubeletFlags updates the Kubelet args with OS specific
// settings.
func updateOSSpecificKubeletFlags(args, host, workspace string) (string, error) {
	output, err := SSH(host, "cat", "/etc/os-release")
	if err != nil {
		return "", fmt.Errorf("issue detecting node's OS via node's /etc/os-release. Err: %v, Output:\n%s", err, output)
	}
	switch {
	case strings.Contains(output, "ID=gci"), strings.Contains(output, "ID=cos"):
		args = prependMemcgNotificationFlag(args)
		return prependCOSMounterFlag(args, host, workspace)
	case strings.Contains(output, "ID=ubuntu"):
		return prependMemcgNotificationFlag(args), nil
	}
	return args, nil
}

// RunTest runs test on the node.
func (n *NodeE2ERemote) RunTest(host, workspace, results, imageDesc, junitFilePrefix, testArgs, ginkgoArgs, systemSpecName string, timeout time.Duration) (string, error) {
	// Install the cni plugins and add a basic CNI configuration.
	// TODO(random-liu): Do this in cloud init after we remove containervm test.
	if err := setupCNI(host, workspace); err != nil {
		return "", err
	}

	// Configure iptables firewall rules
	if err := configureFirewall(host); err != nil {
		return "", err
	}

	// Kill any running node processes
	cleanupNodeProcesses(host)

	testArgs, err := updateOSSpecificKubeletFlags(testArgs, host, workspace)
	if err != nil {
		return "", err
	}

	systemSpecFile := ""
	if systemSpecName != "" {
		systemSpecFile = systemSpecName + ".yaml"
	}

	// Run the tests
	glog.V(2).Infof("Starting tests on %q", host)
	cmd := getSSHCommand(" && ",
		fmt.Sprintf("cd %s", workspace),
		fmt.Sprintf("timeout -k 30s %fs ./ginkgo %s ./e2e_node.test -- --system-spec-name=%s --system-spec-file=%s --logtostderr --v 4 --node-name=%s --report-dir=%s --report-prefix=%s --image-description=\"%s\" %s",
			timeout.Seconds(), ginkgoArgs, systemSpecName, systemSpecFile, host, results, junitFilePrefix, imageDesc, testArgs),
	)
	return SSH(host, "sh", "-c", cmd)
}
