/*

 Copyright 2019 The KubeSphere Authors.

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
package registries

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	log "k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/informers"
)

const (
	loginSuccess  = "Login Succeeded"
	StatusFailed  = "failed"
	StatusSuccess = "succeeded"
)

type DockerConfigJson struct {
	Auths DockerConfigMap `json:"auths"`
}

// DockerConfig represents the config file used by the docker CLI.
// This config that represents the credentials that should be used
// when pulling images from specific image repositories.
type DockerConfigMap map[string]DockerConfigEntry

type DockerConfigEntry struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	Email         string `json:"email"`
	ServerAddress string `json:"serverAddress,omitempty"`
}

func RegistryVerify(authInfo AuthInfo) error {
	auth := base64.StdEncoding.EncodeToString([]byte(authInfo.Username + ":" + authInfo.Password))
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())

	if err != nil {
		klog.Error(err)
	}

	config := types.AuthConfig{
		Username:      authInfo.Username,
		Password:      authInfo.Password,
		Auth:          auth,
		ServerAddress: authInfo.ServerHost,
	}

	resp, err := cli.RegistryLogin(ctx, config)
	cli.Close()

	if err != nil {
		return err
	}

	if resp.Status == loginSuccess {
		return nil
	} else {
		return fmt.Errorf(resp.Status)
	}
}

func GetEntryBySecret(namespace, secretName string) (dockerConfigEntry *DockerConfigEntry, err error) {
	if namespace == "" || secretName == "" {
		return &DockerConfigEntry{}, nil
	}
	secret, err := informers.SharedInformerFactory().Core().V1().Secrets().Lister().Secrets(namespace).Get(secretName)
	if err != nil {
		log.Errorf("%+v", err)
		return nil, err
	}
	entry, err := getDockerEntryFromDockerSecret(secret)
	if err != nil {
		log.Errorf("%+v", err)
		return nil, err
	}
	return entry, nil
}

func getDockerEntryFromDockerSecret(instance *corev1.Secret) (dockerConfigEntry *DockerConfigEntry, err error) {

	if instance.Type != corev1.SecretTypeDockerConfigJson {
		return nil, fmt.Errorf("secret %s in ns %s type should be %s",
			instance.Name, instance.Namespace, corev1.SecretTypeDockerConfigJson)
	}
	dockerConfigBytes, ok := instance.Data[corev1.DockerConfigJsonKey]
	if !ok {
		return nil, fmt.Errorf("could not get data %s", corev1.DockerConfigJsonKey)
	}
	dockerConfig := &DockerConfigJson{}
	err = json.Unmarshal(dockerConfigBytes, dockerConfig)
	if err != nil {
		return nil, err
	}
	if len(dockerConfig.Auths) == 0 {
		return nil, fmt.Errorf("docker config auth len should not be 0")
	}
	for registryAddress, dockerConfigEntry := range dockerConfig.Auths {
		dockerConfigEntry.ServerAddress = registryAddress
		return &dockerConfigEntry, nil
	}
	return nil, nil
}
