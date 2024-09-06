/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package registries

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	dockertypes "github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/emicklei/go-restful/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
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

type RegistryGetter interface {
	VerifyRegistryCredential(credential api.RegistryCredential) error
	GetEntry(namespace, secretName, imageName string, insecure bool) (ImageDetails, error)
}

type registryGetter struct {
	cache runtimeclient.Reader
}

func NewRegistryGetter(cacheReader runtimeclient.Reader) RegistryGetter {
	return &registryGetter{cache: cacheReader}
}

func (c *registryGetter) VerifyRegistryCredential(credential api.RegistryCredential) error {
	auth := base64.StdEncoding.EncodeToString([]byte(credential.Username + ":" + credential.Password))
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())

	if err != nil {
		klog.Error(err)
	}

	config := dockertypes.AuthConfig{
		Username:      credential.Username,
		Password:      credential.Password,
		Auth:          auth,
		ServerAddress: credential.ServerHost,
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

func (c *registryGetter) GetEntry(namespace, secretName, imageName string, insecure bool) (ImageDetails, error) {
	imageDetails, err := c.getEntryBySecret(namespace, secretName, imageName, insecure)
	if imageDetails.Status == StatusFailed {
		imageDetails.Message = err.Error()
	}

	return imageDetails, err
}

func (c *registryGetter) getEntryBySecret(namespace, secretName, imageName string, insecure bool) (ImageDetails, error) {
	failedImageDetails := ImageDetails{
		Status:  StatusFailed,
		Message: "",
	}

	var config *DockerConfigEntry

	if namespace == "" || secretName == "" {
		config = &DockerConfigEntry{}
	} else {
		secret := &corev1.Secret{}
		if err := c.cache.Get(context.Background(),
			types.NamespacedName{Namespace: namespace, Name: secretName}, secret); err != nil {
			return failedImageDetails, err
		}
		var err error
		config, err = getDockerEntryFromDockerSecret(secret)
		if err != nil {
			return failedImageDetails, err
		}
	}

	// default use ssl
	checkSSl := func(serverAddress string) bool {
		if strings.HasPrefix(serverAddress, "http://") {
			return false
		} else {
			return true
		}
	}

	if strings.HasPrefix(imageName, "http") {
		dockerurl, err := ParseDockerURL(imageName)
		if err != nil {
			return failedImageDetails, err
		}
		imageName = dockerurl.StringWithoutScheme()
	}

	// parse image
	image, err := ParseImage(imageName)
	if err != nil {
		return failedImageDetails, err
	}

	useSSL := checkSSl(config.ServerAddress)

	// Create the registry client.
	r, err := CreateRegistryClient(config.Username, config.Password, image.Domain, useSSL, insecure)
	if err != nil {
		return failedImageDetails, err
	}

	digestUrl := r.GetDigestUrl(image)

	// Get token.
	token, err := r.Token(digestUrl)
	if err != nil {
		return failedImageDetails, err
	}

	// Get digest.
	imageManifest, err := r.ImageManifest(image, token)
	if err != nil {
		if serviceError, ok := err.(restful.ServiceError); ok {
			return failedImageDetails, serviceError
		}
		return failedImageDetails, err
	}

	image.Digest = imageManifest.ManifestConfig.Digest

	// Get blob.
	imageBlob, err := r.ImageBlob(image, token)
	if err != nil {
		return failedImageDetails, err
	}

	return ImageDetails{
		Status:        StatusSuccess,
		ImageManifest: imageManifest,
		ImageBlob:     imageBlob,
		ImageTag:      image.Tag,
		Registry:      image.Domain,
	}, nil
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
	for registryAddress, dce := range dockerConfig.Auths {
		dce.ServerAddress = registryAddress
		return &dce, nil
	}
	return nil, nil
}
