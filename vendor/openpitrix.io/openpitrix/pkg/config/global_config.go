// Copyright 2017 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package config

import (
	"fmt"
	"regexp"
	"strings"

	"openpitrix.io/openpitrix/pkg/constants"
	"openpitrix.io/openpitrix/pkg/logger"
	"openpitrix.io/openpitrix/pkg/util/yamlutil"
)

type GlobalConfig struct {
	App             AppServiceConfig                  `json:"app"`
	Repo            RepoServiceConfig                 `json:"repo"`
	Cluster         ClusterServiceConfig              `json:"cluster"`
	RuntimeProvider map[string]*RuntimeProviderConfig `json:"runtime_provider"`
	Pilot           PilotServiceConfig                `json:"pilot"`
	Job             JobServiceConfig                  `json:"job"`
	Task            TaskServiceConfig                 `json:"task"`
	BasicCfg        BasicConfig                       `json:"basic_config"`
	InstallModule   InstallModuleConfig               `json:"install_module"`
}

type AppServiceConfig struct {
	DefaultDraftStatus bool `json:"default_draft_status"`
}

type RepoServiceConfig struct {
	Cron          string `json:"cron"`
	MaxRepoEvents int32  `json:"max_repo_events"`
}

type ClusterServiceConfig struct {
	FrontgateConf       string `json:"frontgate_conf"`
	FrontgateAutoDelete bool   `json:"frontgate_auto_delete"`
	FrontgateAutoUpdate bool   `json:"frontgate_auto_update"`
	RegistryMirror      string `json:"registry_mirror"`
}

type PilotServiceConfig struct {
	Ip   string `json:"ip"`
	Port int32  `json:"port"`
}

type JobServiceConfig struct {
	MaxWorkingJobs int32 `json:"max_working_jobs"`
}

type TaskServiceConfig struct {
	MaxWorkingTasks int32 `json:"max_working_tasks"`
}

type BasicConfig struct {
	PlatformName string `json:"platform_name"`
	PlatformUrl  string `json:"platform_url"`
}

type RuntimeProviderConfig struct {
	ApiServer       string                 `json:"api_server"`
	Zone            string                 `json:"zone"`
	ImageId         string                 `json:"image_id"`
	ImageUrl        string                 `json:"image_url"`
	ImageName       string                 `json:"image_name"`
	FrontgateConf   string                 `json:"frontgate_conf"`
	ProviderType    string                 `json:"provider_type"`
	Host            string                 `json:"host"`
	Port            int                    `json:"port"`
	Enable          bool                   `json:"enable"`
	AdvancedOptions map[string]interface{} `json:"advanced_options"`
}

type InstallModuleConfig struct {
	Iam          bool `json:"iam"`
	Notification bool `json:"notification"`
}

func (r *RuntimeProviderConfig) GetPort() int {
	if r.Port > 0 {
		return r.Port
	} else {
		return constants.RuntimeProviderManagerPort
	}
}

func (r *RuntimeProviderConfig) GetHost(provider string) string {
	if len(r.Host) > 0 {
		return r.Host
	} else {
		return constants.ProviderPrefix + provider
	}
}

func (r *RuntimeProviderConfig) GetEnable() bool {
	return r.Enable
}

func (g *GlobalConfig) GetAppDefaultStatus() string {
	if g.App.DefaultDraftStatus {
		return constants.StatusDraft
	}
	return constants.StatusActive
}

func (g *GlobalConfig) GetRuntimeImageIdAndUrl(apiServer, zone string) (*RuntimeProviderConfig, error) {
	if strings.HasPrefix(apiServer, "https://") {
		apiServer = strings.Split(apiServer, "https://")[1]
	}

	for _, imageConfig := range g.RuntimeProvider {
		if imageConfig.ApiServer == apiServer && imageConfig.Zone == zone {
			return imageConfig, nil
		}
	}
	for _, imageConfig := range g.RuntimeProvider {
		if imageConfig.ApiServer == apiServer && imageConfig.Zone == ".*" {
			return imageConfig, nil
		}
	}
	for _, imageConfig := range g.RuntimeProvider {
		matched, _ := regexp.MatchString(imageConfig.ApiServer, apiServer)

		if matched && imageConfig.Zone == zone {
			return imageConfig, nil
		}
	}
	for _, imageConfig := range g.RuntimeProvider {
		matched, _ := regexp.MatchString(imageConfig.ApiServer, apiServer)

		if matched && imageConfig.Zone == ".*" {
			return imageConfig, nil
		}
	}

	logger.Error(nil, "No such runtime image with api server [%s] zone [%s]. ", apiServer, zone)
	return nil, fmt.Errorf("no such runtime image with api server [%s] zone [%s]. ", apiServer, zone)
}

func (g *GlobalConfig) RegisterRuntimeProviderConfig(provider, config string) error {
	runtimeProviderConfig, err := ParseRuntimeProviderConfig([]byte(config))
	if err != nil {
		return err
	}

	if len(g.RuntimeProvider) == 0 {
		g.RuntimeProvider = make(map[string]*RuntimeProviderConfig)
	}
	_, ok := g.RuntimeProvider[provider]
	if !ok {
		g.RuntimeProvider[provider] = runtimeProviderConfig
	} else {
		oldEnable := g.RuntimeProvider[provider].Enable
		g.RuntimeProvider[provider] = runtimeProviderConfig
		g.RuntimeProvider[provider].Enable = oldEnable
	}
	return nil
}

func ParseGlobalConfig(data []byte) (GlobalConfig, error) {
	var globalConfig GlobalConfig
	err := yamlutil.Decode(data, &globalConfig)
	if err != nil {
		return globalConfig, err
	}
	return globalConfig, nil
}

func ParseRuntimeProviderConfig(data []byte) (*RuntimeProviderConfig, error) {
	var runtimeProviderConfig *RuntimeProviderConfig
	err := yamlutil.Decode(data, &runtimeProviderConfig)
	if err != nil {
		return runtimeProviderConfig, err
	}
	return runtimeProviderConfig, nil
}

func DecodeInitConfig() GlobalConfig {
	globalConfig, err := ParseGlobalConfig([]byte(InitialGlobalConfig))
	if err != nil {
		fmt.Print("InitialGlobalConfig is invalid, please fix it")
		panic(err)
	}
	return globalConfig
}

func EncodeGlobalConfig(conf GlobalConfig) string {
	out, err := yamlutil.Encode(conf)
	if err != nil {
		panic(err)
	}
	return string(out)
}

func init() {
	DecodeInitConfig()
}
