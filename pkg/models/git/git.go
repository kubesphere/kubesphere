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
	RemoteUrl string `json:"remoteUrl"`
}

func GitReadVerify(namespace string, name string, authInfo AuthInfo) error {
	secret, err := informers.SharedInformerFactory().Core().V1().Secrets().Lister().Secrets(namespace).Get(name)
	if err != nil {
		return err
	}
	username, ok := secret.Data[corev1.BasicAuthUsernameKey]
	if !ok {
		return fmt.Errorf("could not get username in secret %s", secret.Name)
	}
	password, ok := secret.Data[corev1.BasicAuthPasswordKey]
	if !ok {
		return fmt.Errorf("could not get password in secret %s", secret.Name)
	}

	r, _ := git.Init(memory.NewStorage(), nil)

	// Add a new remote, with the default fetch refspec
	origin, err := r.CreateRemote(&config.RemoteConfig{
		Name: git.DefaultRemoteName,
		URLs: []string{authInfo.RemoteUrl},
	})
	if err != nil {
		return err
	}
	_, err = origin.List(&git.ListOptions{Auth:
		&http.BasicAuth{Username: string(username), Password: string(password)}})
	return err
}
