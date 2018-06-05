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

package dockershim

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/blang/semver"
	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"

	runtimeapi "k8s.io/kubernetes/pkg/kubelet/apis/cri/runtime/v1alpha2"
)

func TestModifyContainerConfig(t *testing.T) {
	var uid int64 = 123
	var username = "testuser"
	var gid int64 = 423

	cases := []struct {
		name     string
		sc       *runtimeapi.LinuxContainerSecurityContext
		expected *dockercontainer.Config
		isErr    bool
	}{
		{
			name: "container.SecurityContext.RunAsUser set",
			sc: &runtimeapi.LinuxContainerSecurityContext{
				RunAsUser: &runtimeapi.Int64Value{Value: uid},
			},
			expected: &dockercontainer.Config{
				User: strconv.FormatInt(uid, 10),
			},
			isErr: false,
		},
		{
			name: "container.SecurityContext.RunAsUsername set",
			sc: &runtimeapi.LinuxContainerSecurityContext{
				RunAsUsername: username,
			},
			expected: &dockercontainer.Config{
				User: username,
			},
			isErr: false,
		},
		{
			name:     "no RunAsUser value set",
			sc:       &runtimeapi.LinuxContainerSecurityContext{},
			expected: &dockercontainer.Config{},
			isErr:    false,
		},
		{
			name: "RunAsUser value set, RunAsGroup set",
			sc: &runtimeapi.LinuxContainerSecurityContext{
				RunAsUser:  &runtimeapi.Int64Value{Value: uid},
				RunAsGroup: &runtimeapi.Int64Value{Value: gid},
			},
			expected: &dockercontainer.Config{
				User: "123:423",
			},
			isErr: false,
		},
		{
			name: "RunAsUsername value set, RunAsGroup set",
			sc: &runtimeapi.LinuxContainerSecurityContext{
				RunAsUsername: username,
				RunAsGroup:    &runtimeapi.Int64Value{Value: gid},
			},
			expected: &dockercontainer.Config{
				User: "testuser:423",
			},
			isErr: false,
		},
		{
			name: "RunAsUser/RunAsUsername not set, RunAsGroup set",
			sc: &runtimeapi.LinuxContainerSecurityContext{
				RunAsGroup: &runtimeapi.Int64Value{Value: gid},
			},
			isErr: true,
		},
	}

	for _, tc := range cases {
		dockerCfg := &dockercontainer.Config{}
		err := modifyContainerConfig(tc.sc, dockerCfg)
		if tc.isErr {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, tc.expected, dockerCfg, "[Test case %q]", tc.name)
		}
	}
}

func TestModifyHostConfig(t *testing.T) {
	setNetworkHC := &dockercontainer.HostConfig{}
	setPrivSC := &runtimeapi.LinuxContainerSecurityContext{}
	setPrivSC.Privileged = true
	setPrivHC := &dockercontainer.HostConfig{
		Privileged: true,
	}
	setCapsHC := &dockercontainer.HostConfig{
		CapAdd:  []string{"addCapA", "addCapB"},
		CapDrop: []string{"dropCapA", "dropCapB"},
	}
	setSELinuxHC := &dockercontainer.HostConfig{
		SecurityOpt: []string{
			fmt.Sprintf("%s:%s", selinuxLabelUser('='), "user"),
			fmt.Sprintf("%s:%s", selinuxLabelRole('='), "role"),
			fmt.Sprintf("%s:%s", selinuxLabelType('='), "type"),
			fmt.Sprintf("%s:%s", selinuxLabelLevel('='), "level"),
		},
	}

	cases := []struct {
		name     string
		sc       *runtimeapi.LinuxContainerSecurityContext
		expected *dockercontainer.HostConfig
	}{
		{
			name:     "fully set container.SecurityContext",
			sc:       fullValidSecurityContext(),
			expected: fullValidHostConfig(),
		},
		{
			name:     "empty container.SecurityContext",
			sc:       &runtimeapi.LinuxContainerSecurityContext{},
			expected: setNetworkHC,
		},
		{
			name:     "container.SecurityContext.Privileged",
			sc:       setPrivSC,
			expected: setPrivHC,
		},
		{
			name: "container.SecurityContext.Capabilities",
			sc: &runtimeapi.LinuxContainerSecurityContext{
				Capabilities: inputCapabilities(),
			},
			expected: setCapsHC,
		},
		{
			name: "container.SecurityContext.SELinuxOptions",
			sc: &runtimeapi.LinuxContainerSecurityContext{
				SelinuxOptions: inputSELinuxOptions(),
			},
			expected: setSELinuxHC,
		},
	}

	for _, tc := range cases {
		dockerCfg := &dockercontainer.HostConfig{}
		modifyHostConfig(tc.sc, dockerCfg, '=')
		assert.Equal(t, tc.expected, dockerCfg, "[Test case %q]", tc.name)
	}
}

func TestModifyHostConfigWithGroups(t *testing.T) {
	supplementalGroupsSC := &runtimeapi.LinuxContainerSecurityContext{}
	supplementalGroupsSC.SupplementalGroups = []int64{2222}
	supplementalGroupHC := &dockercontainer.HostConfig{}
	supplementalGroupHC.GroupAdd = []string{"2222"}

	testCases := []struct {
		name            string
		securityContext *runtimeapi.LinuxContainerSecurityContext
		expected        *dockercontainer.HostConfig
	}{
		{
			name:            "nil",
			securityContext: nil,
			expected:        &dockercontainer.HostConfig{},
		},
		{
			name:            "SupplementalGroup",
			securityContext: supplementalGroupsSC,
			expected:        supplementalGroupHC,
		},
	}

	for _, tc := range testCases {
		dockerCfg := &dockercontainer.HostConfig{}
		modifyHostConfig(tc.securityContext, dockerCfg, '=')
		assert.Equal(t, tc.expected, dockerCfg, "[Test case %q]", tc.name)
	}
}

func TestModifyHostConfigAndNamespaceOptionsForContainer(t *testing.T) {
	priv := true
	sandboxID := "sandbox"
	sandboxNSMode := fmt.Sprintf("container:%v", sandboxID)
	setPrivSC := &runtimeapi.LinuxContainerSecurityContext{}
	setPrivSC.Privileged = priv
	setPrivHC := &dockercontainer.HostConfig{
		Privileged:  true,
		IpcMode:     dockercontainer.IpcMode(sandboxNSMode),
		NetworkMode: dockercontainer.NetworkMode(sandboxNSMode),
		PidMode:     dockercontainer.PidMode(sandboxNSMode),
	}
	setCapsHC := &dockercontainer.HostConfig{
		CapAdd:      []string{"addCapA", "addCapB"},
		CapDrop:     []string{"dropCapA", "dropCapB"},
		IpcMode:     dockercontainer.IpcMode(sandboxNSMode),
		NetworkMode: dockercontainer.NetworkMode(sandboxNSMode),
		PidMode:     dockercontainer.PidMode(sandboxNSMode),
	}
	setSELinuxHC := &dockercontainer.HostConfig{
		SecurityOpt: []string{
			fmt.Sprintf("%s:%s", selinuxLabelUser('='), "user"),
			fmt.Sprintf("%s:%s", selinuxLabelRole('='), "role"),
			fmt.Sprintf("%s:%s", selinuxLabelType('='), "type"),
			fmt.Sprintf("%s:%s", selinuxLabelLevel('='), "level"),
		},
		IpcMode:     dockercontainer.IpcMode(sandboxNSMode),
		NetworkMode: dockercontainer.NetworkMode(sandboxNSMode),
		PidMode:     dockercontainer.PidMode(sandboxNSMode),
	}

	cases := []struct {
		name     string
		sc       *runtimeapi.LinuxContainerSecurityContext
		expected *dockercontainer.HostConfig
	}{
		{
			name:     "container.SecurityContext.Privileged",
			sc:       setPrivSC,
			expected: setPrivHC,
		},
		{
			name: "container.SecurityContext.Capabilities",
			sc: &runtimeapi.LinuxContainerSecurityContext{
				Capabilities: inputCapabilities(),
			},
			expected: setCapsHC,
		},
		{
			name: "container.SecurityContext.SELinuxOptions",
			sc: &runtimeapi.LinuxContainerSecurityContext{
				SelinuxOptions: inputSELinuxOptions(),
			},
			expected: setSELinuxHC,
		},
	}

	for _, tc := range cases {
		dockerCfg := &dockercontainer.HostConfig{}
		modifyHostConfig(tc.sc, dockerCfg, '=')
		modifyContainerNamespaceOptions(tc.sc.GetNamespaceOptions(), sandboxID, dockerCfg)
		assert.Equal(t, tc.expected, dockerCfg, "[Test case %q]", tc.name)
	}
}

func TestModifySandboxNamespaceOptions(t *testing.T) {
	cases := []struct {
		name     string
		nsOpt    *runtimeapi.NamespaceOption
		expected *dockercontainer.HostConfig
	}{
		{
			name: "Host Network NamespaceOption",
			nsOpt: &runtimeapi.NamespaceOption{
				Network: runtimeapi.NamespaceMode_NODE,
			},
			expected: &dockercontainer.HostConfig{
				NetworkMode: namespaceModeHost,
			},
		},
		{
			name: "Host IPC NamespaceOption",
			nsOpt: &runtimeapi.NamespaceOption{
				Ipc: runtimeapi.NamespaceMode_NODE,
			},
			expected: &dockercontainer.HostConfig{
				IpcMode:     namespaceModeHost,
				NetworkMode: "default",
			},
		},
		{
			name: "Host PID NamespaceOption",
			nsOpt: &runtimeapi.NamespaceOption{
				Pid: runtimeapi.NamespaceMode_NODE,
			},
			expected: &dockercontainer.HostConfig{
				PidMode:     namespaceModeHost,
				NetworkMode: "default",
			},
		},
	}
	for _, tc := range cases {
		dockerCfg := &dockercontainer.HostConfig{}
		modifySandboxNamespaceOptions(tc.nsOpt, dockerCfg, nil)
		assert.Equal(t, tc.expected, dockerCfg, "[Test case %q]", tc.name)
	}
}

func TestModifyContainerNamespaceOptions(t *testing.T) {
	sandboxID := "sandbox"
	sandboxNSMode := fmt.Sprintf("container:%v", sandboxID)
	cases := []struct {
		name     string
		nsOpt    *runtimeapi.NamespaceOption
		expected *dockercontainer.HostConfig
	}{
		{
			name: "Host Network NamespaceOption",
			nsOpt: &runtimeapi.NamespaceOption{
				Network: runtimeapi.NamespaceMode_NODE,
			},
			expected: &dockercontainer.HostConfig{
				NetworkMode: dockercontainer.NetworkMode(sandboxNSMode),
				IpcMode:     dockercontainer.IpcMode(sandboxNSMode),
				UTSMode:     namespaceModeHost,
				PidMode:     dockercontainer.PidMode(sandboxNSMode),
			},
		},
		{
			name: "Host IPC NamespaceOption",
			nsOpt: &runtimeapi.NamespaceOption{
				Ipc: runtimeapi.NamespaceMode_NODE,
			},
			expected: &dockercontainer.HostConfig{
				NetworkMode: dockercontainer.NetworkMode(sandboxNSMode),
				IpcMode:     dockercontainer.IpcMode(sandboxNSMode),
				PidMode:     dockercontainer.PidMode(sandboxNSMode),
			},
		},
		{
			name: "Host PID NamespaceOption",
			nsOpt: &runtimeapi.NamespaceOption{
				Pid: runtimeapi.NamespaceMode_NODE,
			},
			expected: &dockercontainer.HostConfig{
				NetworkMode: dockercontainer.NetworkMode(sandboxNSMode),
				IpcMode:     dockercontainer.IpcMode(sandboxNSMode),
				PidMode:     namespaceModeHost,
			},
		},
	}
	for _, tc := range cases {
		dockerCfg := &dockercontainer.HostConfig{}
		modifyContainerNamespaceOptions(tc.nsOpt, sandboxID, dockerCfg)
		assert.Equal(t, tc.expected, dockerCfg, "[Test case %q]", tc.name)
	}
}

func TestModifyContainerNamespacePIDOverride(t *testing.T) {
	cases := []struct {
		name            string
		disable         bool
		version         *semver.Version
		input, expected dockercontainer.PidMode
	}{
		{
			name:     "mode:CONTAINER docker:NEW flag:UNSET",
			disable:  true,
			version:  &semver.Version{Major: 1, Minor: 26},
			input:    "",
			expected: "",
		},
		{
			name:     "mode:CONTAINER docker:NEW flag:SET",
			disable:  false,
			version:  &semver.Version{Major: 1, Minor: 26},
			input:    "",
			expected: "container:sandbox",
		},
		{
			name:     "mode:CONTAINER docker:OLD flag:UNSET",
			disable:  true,
			version:  &semver.Version{Major: 1, Minor: 25},
			input:    "",
			expected: "",
		},
		{
			name:     "mode:CONTAINER docker:OLD flag:SET",
			disable:  false,
			version:  &semver.Version{Major: 1, Minor: 25},
			input:    "",
			expected: "",
		},
		{
			name:     "mode:HOST docker:NEW flag:UNSET",
			disable:  true,
			version:  &semver.Version{Major: 1, Minor: 26},
			input:    "host",
			expected: "host",
		},
		{
			name:     "mode:HOST docker:NEW flag:SET",
			disable:  false,
			version:  &semver.Version{Major: 1, Minor: 26},
			input:    "host",
			expected: "host",
		},
		{
			name:     "mode:HOST docker:OLD flag:UNSET",
			disable:  true,
			version:  &semver.Version{Major: 1, Minor: 25},
			input:    "host",
			expected: "host",
		},
		{
			name:     "mode:HOST docker:OLD flag:SET",
			disable:  false,
			version:  &semver.Version{Major: 1, Minor: 25},
			input:    "host",
			expected: "host",
		},
		{
			name:     "mode:POD docker:NEW flag:UNSET",
			disable:  true,
			version:  &semver.Version{Major: 1, Minor: 26},
			input:    "container:sandbox",
			expected: "container:sandbox",
		},
		{
			name:     "mode:POD docker:NEW flag:SET",
			disable:  false,
			version:  &semver.Version{Major: 1, Minor: 26},
			input:    "container:sandbox",
			expected: "container:sandbox",
		},
		{
			name:     "mode:POD docker:OLD flag:UNSET",
			disable:  true,
			version:  &semver.Version{Major: 1, Minor: 25},
			input:    "container:sandbox",
			expected: "",
		},
		{
			name:     "mode:POD docker:OLD flag:SET",
			disable:  false,
			version:  &semver.Version{Major: 1, Minor: 25},
			input:    "container:sandbox",
			expected: "",
		},
	}
	for _, tc := range cases {
		dockerCfg := &dockercontainer.HostConfig{PidMode: tc.input}
		modifyContainerPIDNamespaceOverrides(tc.disable, tc.version, dockerCfg, "sandbox")
		assert.Equal(t, tc.expected, dockerCfg.PidMode, "[Test case %q]", tc.name)
	}
}

func fullValidSecurityContext() *runtimeapi.LinuxContainerSecurityContext {
	return &runtimeapi.LinuxContainerSecurityContext{
		Privileged:     true,
		Capabilities:   inputCapabilities(),
		SelinuxOptions: inputSELinuxOptions(),
	}
}

func inputCapabilities() *runtimeapi.Capability {
	return &runtimeapi.Capability{
		AddCapabilities:  []string{"addCapA", "addCapB"},
		DropCapabilities: []string{"dropCapA", "dropCapB"},
	}
}

func inputSELinuxOptions() *runtimeapi.SELinuxOption {
	user := "user"
	role := "role"
	stype := "type"
	level := "level"

	return &runtimeapi.SELinuxOption{
		User:  user,
		Role:  role,
		Type:  stype,
		Level: level,
	}
}

func fullValidHostConfig() *dockercontainer.HostConfig {
	return &dockercontainer.HostConfig{
		Privileged: true,
		CapAdd:     []string{"addCapA", "addCapB"},
		CapDrop:    []string{"dropCapA", "dropCapB"},
		SecurityOpt: []string{
			fmt.Sprintf("%s:%s", selinuxLabelUser('='), "user"),
			fmt.Sprintf("%s:%s", selinuxLabelRole('='), "role"),
			fmt.Sprintf("%s:%s", selinuxLabelType('='), "type"),
			fmt.Sprintf("%s:%s", selinuxLabelLevel('='), "level"),
		},
	}
}
