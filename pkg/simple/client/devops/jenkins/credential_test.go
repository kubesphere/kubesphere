/*
Copyright 2020 KubeSphere Authors

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

package jenkins

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewUsernamePasswordCredential(t *testing.T) {
	username := "test-user"
	password := "password"
	name := "test-secret"
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test",
		},
		Data: map[string][]byte{
			"username": []byte(username),
			"password": []byte(password),
		},
		Type: "credential.devops.kubesphere.io/basic-auth",
	}
	credential := NewUsernamePasswordCredential(secret)
	if credential.StaplerClass != "com.cloudbees.plugins.credentials.impl.UsernamePasswordCredentialsImpl" {
		t.Fatalf("credential's stapler class should be com.cloudbees.plugins.credentials.impl.UsernamePasswordCredentialsImpl"+
			"other than %s ", credential.StaplerClass)
	}
	if credential.Id != name {
		t.Fatalf("credential's id should be %s "+
			"other than %s ", name, credential.Id)
	}
	if credential.Username != username {
		t.Fatalf("credential's username should be %s "+
			"other than %s ", username, credential.Username)
	}
	if credential.Password != password {
		t.Fatalf("credential's password should be %s "+
			"other than %s ", password, credential.Password)
	}
}

func TestNewSshCredential(t *testing.T) {
	username := "test-user"
	passphrase := "passphrase"
	privatekey := "pk"
	name := "test-secret"
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test",
		},
		Data: map[string][]byte{
			"username":    []byte(username),
			"passphrase":  []byte(passphrase),
			"private_key": []byte(privatekey),
		},
		Type: "credential.devops.kubesphere.io/ssh-auth",
	}
	credential := NewSshCredential(secret)
	if credential.StaplerClass != "com.cloudbees.jenkins.plugins.sshcredentials.impl.BasicSSHUserPrivateKey" {
		t.Fatalf("credential's stapler class should be com.cloudbees.jenkins.plugins.sshcredentials.impl.BasicSSHUserPrivateKey"+
			"other than %s ", credential.StaplerClass)
	}
	if credential.Id != name {
		t.Fatalf("credential's id should be %s "+
			"other than %s ", name, credential.Id)
	}
	if credential.Username != username {
		t.Fatalf("credential's username should be %s "+
			"other than %s ", username, credential.Username)
	}
	if credential.Passphrase != passphrase {
		t.Fatalf("credential's passphrase should be %s "+
			"other than %s ", passphrase, credential.Passphrase)
	}
	if credential.KeySource.PrivateKey != privatekey {
		t.Fatalf("credential's privatekey should be %s "+
			"other than %s ", privatekey, credential.KeySource.PrivateKey)
	}
}

func TestNewKubeconfigCredential(t *testing.T) {
	content := []byte("test-content")
	name := "test-secret"
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test",
		},
		Type: "credential.devops.kubesphere.io/kubeconfig",
		Data: map[string][]byte{"content": content},
	}
	credential := NewKubeconfigCredential(secret)
	if credential.StaplerClass != "com.microsoft.jenkins.kubernetes.credentials.KubeconfigCredentials" {
		t.Fatalf("credential's stapler class should be com.microsoft.jenkins.kubernetes.credentials.KubeconfigCredentials"+
			"other than %s ", credential.StaplerClass)
	}
	if credential.Id != name {
		t.Fatalf("credential's id should be %s "+
			"other than %s ", name, credential.Id)
	}
	if credential.KubeconfigSource.Content != string(content) {
		t.Fatalf("credential's content should be %s "+
			"other than %s ", string(content), credential.KubeconfigSource.Content)
	}
}

func TestNewSecretTextCredential(t *testing.T) {
	content := []byte("test-content")
	name := "test-secret"
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test",
		},
		Type: "credential.devops.kubesphere.io/secret-text",
		Data: map[string][]byte{"secret": content},
	}
	credential := NewSecretTextCredential(secret)
	if credential.StaplerClass != "org.jenkinsci.plugins.plaincredentials.impl.StringCredentialsImpl" {
		t.Fatalf("credential's stapler class should be org.jenkinsci.plugins.plaincredentials.impl.StringCredentialsImpl"+
			"other than %s ", credential.StaplerClass)
	}
	if credential.Id != name {
		t.Fatalf("credential's id should be %s "+
			"other than %s ", name, credential.Id)
	}
	if credential.Secret != string(content) {
		t.Fatalf("credential's content should be %s "+
			"other than %s ", string(content), credential.Secret)
	}
}
