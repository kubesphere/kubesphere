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

package git

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type AuthInfo struct {
	RemoteUrl string                  `json:"remoteUrl" description:"git server url"`
	SecretRef *corev1.SecretReference `json:"secretRef,omitempty" description:"auth secret reference"`
}

type GitVerifier interface {
	VerifyGitCredential(remoteUrl, namespace, secretName string) error
}

type gitVerifier struct {
	cache runtimeclient.Reader
}

func NewGitVerifier(cacheReader runtimeclient.Reader) GitVerifier {
	return &gitVerifier{cache: cacheReader}
}

func (c *gitVerifier) VerifyGitCredential(remoteUrl, namespace, secretName string) error {
	var username, password string

	if len(secretName) > 0 {
		secret := &corev1.Secret{}
		if err := c.cache.Get(context.Background(),
			types.NamespacedName{Namespace: namespace, Name: secretName}, secret); err != nil {
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

	return c.gitReadVerifyWithBasicAuth(username, password, remoteUrl)
}

func (c *gitVerifier) gitReadVerifyWithBasicAuth(username string, password string, remote string) error {
	r, _ := git.Init(memory.NewStorage(), nil)
	// Add a new remote, with the default fetch refspec
	origin, err := r.CreateRemote(&config.RemoteConfig{
		Name: git.DefaultRemoteName,
		URLs: []string{remote},
	})
	if err != nil {
		return err
	}
	_, err = origin.List(&git.ListOptions{Auth: &http.BasicAuth{Username: username, Password: password}})
	return err
}
