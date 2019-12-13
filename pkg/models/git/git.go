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

package git

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	corev1 "k8s.io/api/core/v1"
	"kubesphere.io/kubesphere/pkg/informers"
)

type AuthInfo struct {
	RemoteUrl string                  `json:"remoteUrl" description:"git server url"`
	SecretRef *corev1.SecretReference `json:"secretRef,omitempty" description:"auth secret reference"`
}

func GitReadVerify(namespace string, authInfo AuthInfo) error {
	username := ""
	password := ""
	if authInfo.SecretRef != nil {
		secret, err := informers.SharedInformerFactory().Core().V1().Secrets().Lister().
			Secrets(authInfo.SecretRef.Namespace).Get(authInfo.SecretRef.Name)
		if err != nil {
			return err
		}
		usernameBytes, ok := secret.Data[corev1.BasicAuthUsernameKey]
		if !ok {
			return fmt.Errorf("could not get username in secret %s", secret.Name)
		}
		passwordBytes, ok := secret.Data[corev1.BasicAuthPasswordKey]
		if !ok {
			return fmt.Errorf("could not get password in secret %s", secret.Name)
		}
		username = string(usernameBytes)
		password = string(passwordBytes)
	}

	return gitReadVerifyWithBasicAuth(string(username), string(password), authInfo.RemoteUrl)
}

func gitReadVerifyWithBasicAuth(username string, password string, remote string) error {
	r, _ := git.Init(memory.NewStorage(), nil)

	// Add a new remote, with the default fetch refspec
	origin, err := r.CreateRemote(&config.RemoteConfig{
		Name: git.DefaultRemoteName,
		URLs: []string{remote},
	})
	if err != nil {
		return err
	}
	_, err = origin.List(&git.ListOptions{Auth: &http.BasicAuth{Username: string(username), Password: string(password)}})
	return err
}
