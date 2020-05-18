package git

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
)

type AuthInfo struct {
	RemoteUrl string                  `json:"remoteUrl" description:"git server url"`
	SecretRef *corev1.SecretReference `json:"secretRef,omitempty" description:"auth secret reference"`
}

type GitVerifier interface {
	VerifyGitCredential(remoteUrl, namespace, secretName string) error
}

type gitVerifier struct {
	informers informers.SharedInformerFactory
}

func NewGitVerifier(informers informers.SharedInformerFactory) GitVerifier {
	return &gitVerifier{informers: informers}
}

func (c *gitVerifier) VerifyGitCredential(remoteUrl, namespace, secretName string) error {
	var username, password string

	if len(secretName) > 0 {
		secret, err := c.informers.Core().V1().Secrets().Lister().Secrets(namespace).Get(secretName)
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
	_, err = origin.List(&git.ListOptions{Auth: &http.BasicAuth{Username: string(username), Password: string(password)}})
	return err
}
