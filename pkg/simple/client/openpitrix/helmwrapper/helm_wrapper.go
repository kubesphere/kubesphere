/*
Copyright 2020 The KubeSphere Authors.

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

package helmwrapper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"k8s.io/klog"
	kpath "k8s.io/utils/path"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
	"os"
	"os/exec"
	"path/filepath"
	"sigs.k8s.io/kustomize/pkg/types"
	"strings"
	"time"
)

const (
	workspaceBase = "/tmp/helm-operator"
)

var (
	UninstallNotFoundFormat = "Error: uninstall: Release not loaded: %s: release: not found"
	StatusNotFoundFormat    = "Error: release: not found"

	kustomizationFile  = "kustomization.yaml"
	postRenderExecFile = "helm-post-render.sh"
	// kustomize cannot read stdio now, so we save helm stdout to file, then kustomize reads that file and build the resources
	kustomizeBuild = `#!/bin/sh
# save helm stdout to file, then kustomize read this file
cat > ./.local-helm-output.yaml
kustomize build
`
)

type HelmRes struct {
	Message string
}

type releaseStatus struct {
	Name string `json:"name,omitempty"`
	Info *Info  `json:"info,omitempty"`
}

// copy from helm
// Info describes release information.
type Info struct {
	// FirstDeployed is when the release was first deployed.
	FirstDeployed time.Time `json:"first_deployed,omitempty"`
	// LastDeployed is when the release was last deployed.
	LastDeployed time.Time `json:"last_deployed,omitempty"`
	// Deleted tracks when this object was deleted.
	Deleted time.Time `json:"deleted"`
	// Description is human-friendly "log entry" about this release.
	Description string `json:"description,omitempty"`
	// Status is the current state of the release
	Status string `json:"status,omitempty"`
	// Contains the rendered templates/NOTES.txt if available
	Notes string `json:"notes,omitempty"`
}

type helmRlsStatus struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Revision   int    `json:"revision"`
	Status     string `json:"status"`
	Chart      string `json:"chart"`
	AppVersion string `json:"app_version"`
}

var _ HelmWrapper = &helmWrapper{}

type HelmWrapper interface {
	Install(chartName, chartData, values string) (HelmRes, error)
	// upgrade a release
	Upgrade(chartName, chartData, values string) (HelmRes, error)
	Uninstall() (HelmRes, error)
	// Get manifests
	Manifest() (string, error)
}

func (c *helmWrapper) Status() (status *releaseStatus, err error) {
	if err = c.ensureWorkspace(); err != nil {
		return nil, err
	}
	defer c.cleanup()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := exec.Cmd{
		Path: c.cmdPath,
		Dir:  c.Workspace(),
		Args: []string{
			c.cmdPath,
			"status",
			fmt.Sprintf("%s", c.ReleaseName),
			"--namespace",
			c.Namespace,
			"--output",
			"json",
		},
		Stderr: stderr,
		Stdout: stdout,
	}

	if c.kubeConfigPath() != "" {
		cmd.Args = append(cmd.Args, "--kubeconfig", c.kubeConfigPath())
	}

	err = cmd.Run()

	if err != nil {
		klog.Errorf("namespace: %s, name: %s, run command failed, stderr: %s, error: %v", c.Namespace, c.ReleaseName, stderr, err)
		return
	} else {
		klog.V(2).Infof("namespace: %s, name: %s, run command success", c.Namespace, c.ReleaseName)
		klog.V(8).Infof("namespace: %s, name: %s, run command success, stdout: %s", c.Namespace, c.ReleaseName, stdout)
	}

	status = &releaseStatus{}
	err = json.Unmarshal(stdout.Bytes(), status)
	if err != nil {
		klog.Errorf("namespace: %s, name: %s, json unmarshal failed, error: %s", c.Namespace, c.ReleaseName, err)
	}

	return
}

func (c *helmWrapper) Workspace() string {
	if c.workspaceSuffix == "" {
		return filepath.Join(c.base, fmt.Sprintf("%s_%s", c.Namespace, c.ReleaseName))
	} else {
		return filepath.Join(c.base, fmt.Sprintf("%s_%s_%s", c.Namespace, c.ReleaseName, c.workspaceSuffix))
	}
}

type helmWrapper struct {
	// KubeConfig string
	Kubeconfig string
	Namespace  string
	// helm release name
	ReleaseName string
	ChartName   string

	// add labels to helm chart
	labels map[string]string
	// add annotations to helm chart
	annotations map[string]string

	// helm cmd path
	cmdPath string
	// base should be /dev/shm on linux
	base            string
	workspaceSuffix string
	dryRun          bool
	mock            bool
}

func (c *helmWrapper) kubeConfigPath() string {
	if len(c.Kubeconfig) == 0 {
		return ""
	}
	return filepath.Join(c.Workspace(), "kube.config")
}

// The dir where chart saved
func (c *helmWrapper) chartDir() string {
	return filepath.Join(c.Workspace(), "chart")
}

func (c *helmWrapper) chartPath() string {
	return filepath.Join(c.chartDir(), fmt.Sprintf("%s.tgz", c.ChartName))
}

func (c *helmWrapper) cleanup() {
	if err := os.RemoveAll(c.Workspace()); err != nil {
		klog.Errorf("remove dir %s faield, error: %s", c.Workspace(), err)
	}
}

func (c *helmWrapper) Set(options ...Option) {
	for _, option := range options {
		option(c)
	}
}

type Option func(*helmWrapper)

func SetDryRun(dryRun bool) Option {
	return func(wrapper *helmWrapper) {
		wrapper.dryRun = dryRun
	}
}

// extra annotations added to all resources in chart
func SetAnnotations(annotations map[string]string) Option {
	return func(wrapper *helmWrapper) {
		wrapper.annotations = annotations
	}
}

func SetMock(mock bool) Option {
	return func(wrapper *helmWrapper) {
		wrapper.mock = mock
	}
}

func NewHelmWrapper(kubeconfig, ns, rls string, options ...Option) *helmWrapper {
	c := &helmWrapper{
		Kubeconfig:      kubeconfig,
		Namespace:       ns,
		ReleaseName:     rls,
		base:            workspaceBase,
		cmdPath:         helmPath,
		workspaceSuffix: idutils.GetUuid36(""),
	}

	for _, option := range options {
		option(c)
	}

	return c
}

func (c *helmWrapper) setupPostRenderEnvironment() error {
	if len(c.labels) == 0 && len(c.annotations) == 0 {
		return nil
	}

	// build the executable file
	postRender, err := os.OpenFile(filepath.Join(c.Workspace(), postRenderExecFile), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	_, err = postRender.WriteString(kustomizeBuild)
	if err != nil {
		return err
	}
	postRender.Close()

	// create kustomization.yaml
	kustomization, err := os.OpenFile(filepath.Join(c.Workspace(), kustomizationFile), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	kustomizationConfig := types.Kustomization{
		Resources:         []string{"./.local-helm-output.yaml"},
		CommonAnnotations: c.annotations, // add extra annotations to output
		CommonLabels:      c.labels,      // add extra labels to output
	}

	err = yaml.NewEncoder(kustomization).Encode(kustomizationConfig)
	if err != nil {
		return err
	}
	kustomization.Close()

	return nil
}

// ensureWorkspace check whether workspace exists or not.
// If not exists, create workspace dir.
func (c *helmWrapper) ensureWorkspace() error {
	if exists, err := kpath.Exists(kpath.CheckFollowSymlink, c.Workspace()); err != nil {
		klog.Errorf("check dir %s failed, error: %s", c.Workspace(), err)
		return err
	} else if !exists {
		err = os.MkdirAll(c.Workspace(), os.ModeDir|os.ModePerm)
		if err != nil {
			klog.Errorf("mkdir %s failed, error: %s", c.Workspace(), err)
			return err
		}
	}

	err := os.MkdirAll(c.chartDir(), os.ModeDir|os.ModePerm)
	if err != nil {
		klog.Errorf("mkdir %s failed, error: %s", c.chartDir(), err)
		return err
	}

	if len(c.Kubeconfig) > 0 {
		kubeFile, err := os.OpenFile(c.kubeConfigPath(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		_, err = kubeFile.WriteString(c.Kubeconfig)
		if err != nil {
			return err
		}
		kubeFile.Close()
	}

	return nil
}

// create chart dir in workspace
// write values.yaml into workspace
func (c *helmWrapper) createChart(chartName, chartData, values string) error {
	c.ChartName = chartName

	// write chart
	f, err := os.Create(c.chartPath())

	if err != nil {
		return err
	}

	_, err = f.Write([]byte(chartData))

	if err != nil {
		return err
	}
	f.Close()

	// write values
	f, err = os.Create(filepath.Join(c.Workspace(), "values.yaml"))
	if err != nil {
		return err
	}

	_, err = f.WriteString(values)
	if err != nil {
		return err
	}

	f.Close()
	return nil
}

// helm uninstall
func (c *helmWrapper) Uninstall() (res HelmRes, err error) {
	start := time.Now()
	defer func() {
		klog.V(2).Infof("run command end, namespace: %s, name: %s elapsed: %v", c.Namespace, c.ReleaseName, time.Now().Sub(start))
	}()

	if err = c.ensureWorkspace(); err != nil {
		return
	}
	defer c.cleanup()

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	cmd := exec.Cmd{
		Path:   c.cmdPath,
		Dir:    c.Workspace(),
		Stdout: stdout,
		Stderr: stderr,
	}

	cmd.Args = make([]string, 0, 10)

	// only for mock
	if c.mock {
		cmd.Path = os.Args[0]
		cmd.Args = []string{os.Args[0], "-test.run=TestHelperProcess", "--"}
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	}

	cmd.Args = append(cmd.Args, c.cmdPath,
		"uninstall",
		c.ReleaseName,
		"--namespace",
		c.Namespace)

	if c.dryRun {
		cmd.Args = append(cmd.Args, "--dry-run")
	}

	if c.kubeConfigPath() != "" {
		cmd.Args = append(cmd.Args, "--kubeconfig", c.kubeConfigPath())
	}

	klog.V(4).Infof("run command: %s", cmd.String())
	err = cmd.Run()

	if err != nil {
		eMsg := strings.TrimSpace(stderr.String())
		if fmt.Sprintf(UninstallNotFoundFormat, c.ReleaseName) == eMsg {
			return res, nil
		}
		klog.Errorf("run command failed, stderr: %s, error: %v", eMsg, err)
		res.Message = eMsg
	} else {
		klog.V(2).Infof("namespace: %s, name: %s, run command success", c.Namespace, c.ReleaseName)
		klog.V(8).Infof("namespace: %s, name: %s, run command success, stdout: %s", c.Namespace, c.ReleaseName, stdout)
	}

	return
}

// helm upgrade
func (c *helmWrapper) Upgrade(chartName, chartData, values string) (res HelmRes, err error) {
	// TODO: check release status first
	if true {
		return c.install(chartName, chartData, values, true)
	} else {
		klog.V(3).Infof("release %s/%s not exists, cannot upgrade it, install a new one", c.Namespace, c.ReleaseName)
		return
	}
}

// helm install
func (c *helmWrapper) Install(chartName, chartData, values string) (res HelmRes, err error) {
	return c.install(chartName, chartData, values, false)
}

func (c *helmWrapper) install(chartName, chartData, values string, upgrade bool) (res HelmRes, err error) {
	if klog.V(2) {
		start := time.Now()
		defer func() {
			klog.V(2).Infof("run command end, namespace: %s, name: %s elapsed: %v", c.Namespace, c.ReleaseName, time.Now().Sub(start))
		}()
	}

	if err = c.ensureWorkspace(); err != nil {
		return
	}
	defer c.cleanup()

	err = c.setupPostRenderEnvironment()
	if err != nil {
		return
	}

	if err = c.createChart(chartName, chartData, values); err != nil {
		return
	}
	klog.V(8).Infof("namespace: %s, name: %s, chart values: %s", c.Namespace, c.ReleaseName, values)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := exec.Cmd{
		Path:   c.cmdPath,
		Dir:    c.Workspace(),
		Stdout: stdout,
		Stderr: stderr,
	}

	cmd.Args = make([]string, 0, 10)

	// only for mock
	if c.mock {
		cmd.Path = os.Args[0]
		cmd.Args = []string{os.Args[0], "-test.run=TestHelperProcess", "--"}
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	}

	cmd.Args = append(cmd.Args, c.cmdPath)
	if upgrade {
		cmd.Args = append(cmd.Args, "upgrade")
	} else {
		cmd.Args = append(cmd.Args, "install")
	}

	cmd.Args = append(cmd.Args, c.ReleaseName, c.chartPath(), "--namespace", c.Namespace)

	if len(values) > 0 {
		cmd.Args = append(cmd.Args, "--values", filepath.Join(c.Workspace(), "values.yaml"))
	}

	if c.dryRun {
		cmd.Args = append(cmd.Args, "--dry-run")
	}

	if c.kubeConfigPath() != "" {
		cmd.Args = append(cmd.Args, "--kubeconfig", c.kubeConfigPath())
	}

	// Post render, add annotations or labels to resources
	if len(c.labels) > 0 || len(c.annotations) > 0 {
		cmd.Args = append(cmd.Args, "--post-renderer", filepath.Join(c.Workspace(), postRenderExecFile))
	}

	if klog.V(8) {
		// output debug info
		cmd.Args = append(cmd.Args, "--debug")
	}

	klog.V(4).Infof("run command: %s", cmd.String())
	err = cmd.Run()

	if err != nil {
		klog.Errorf("namespace: %s, name: %s, run command: %s failed, stderr: %s, error: %v", c.Namespace, c.ReleaseName, cmd.String(), stderr, err)
		res.Message = stderr.String()
	} else {
		klog.V(2).Infof("namespace: %s, name: %s, run command success", c.Namespace, c.ReleaseName)
		klog.V(8).Infof("namespace: %s, name: %s, run command success, stdout: %s", c.Namespace, c.ReleaseName, stdout)
	}

	return
}

func (c *helmWrapper) Manifest() (manifest string, err error) {
	if err = c.ensureWorkspace(); err != nil {
		return "", err
	}
	defer c.cleanup()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := exec.Cmd{
		Path: c.cmdPath,
		Dir:  c.Workspace(),
		Args: []string{
			c.cmdPath,
			"get",
			"manifest",
			c.ReleaseName,
			"--namespace",
			c.Namespace,
		},
		Stderr: stderr,
		Stdout: stdout,
	}

	if c.kubeConfigPath() != "" {
		cmd.Args = append(cmd.Args, "--kubeconfig", c.kubeConfigPath())
	}

	if klog.V(8) {
		// output debug info
		cmd.Args = append(cmd.Args, "--debug")
	}

	klog.V(4).Infof("run command: %s", cmd.String())
	err = cmd.Run()

	if err != nil {
		klog.Errorf("namespace: %s, name: %s, run command failed, stderr: %s, error: %v", c.Namespace, c.ReleaseName, stderr, err)
		return "", err
	} else {
		klog.V(2).Infof("namespace: %s, name: %s, run command success", c.Namespace, c.ReleaseName)
		klog.V(8).Infof("namespace: %s, name: %s, run command success, stdout: %s", c.Namespace, c.ReleaseName, stdout)
	}

	return stdout.String(), nil
}
