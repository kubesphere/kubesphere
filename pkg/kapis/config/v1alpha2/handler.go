/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha2

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/emicklei/go-restful/v3"
	jsonpatch "github.com/evanphx/json-patch/v5"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	"kubesphere.io/kubesphere/pkg/apiserver/options"
	"kubesphere.io/kubesphere/pkg/apiserver/rest"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/server/errors"
)

const (
	themeConfigurationName             = "platform-configuration-theme"
	GenericPlatformConfigurationKind   = "GenericPlatformConfiguration"
	ClusterConnectionConfigurationKind = "ClusterConnectionConfiguration"
)

type GenericPlatformConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Data runtime.RawExtension `json:"data,omitempty"`
}

func NewHandler(config *options.Options, client client.Client) rest.Handler {
	return &handler{
		config:                         config,
		client:                         client,
		oauthClientConfigurationGetter: oauth.NewOAuthClientGetter(client),
	}
}

func NewFakeHandler() rest.Handler {
	return &handler{}
}

type handler struct {
	config                         *options.Options
	client                         client.Client
	oauthClientConfigurationGetter oauth.ClientGetter
}

type ThemeConfiguration map[string]string

func (h *handler) updateThemeConfiguration(req *restful.Request, resp *restful.Response) {
	platformInformation := ThemeConfiguration{}
	if err := req.ReadEntity(&platformInformation); err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      themeConfigurationName,
			Namespace: constants.KubeSphereNamespace,
		},
	}
	_, err := ctrl.CreateOrUpdate(req.Request.Context(), h.client, &configMap, func() error {
		configMap.Data = platformInformation
		return nil
	})
	if err != nil {
		api.HandleInternalError(resp, req, err)
		return
	}
	_ = resp.WriteEntity(platformInformation)
}

func (h *handler) getThemeConfiguration(req *restful.Request, resp *restful.Response) {
	var configMap corev1.ConfigMap
	themeConfiguration := ThemeConfiguration{}
	configName := types.NamespacedName{Namespace: constants.KubeSphereNamespace, Name: themeConfigurationName}
	if err := h.client.Get(req.Request.Context(), configName, &configMap); err != nil {
		if apierrors.IsNotFound(err) {
			_ = resp.WriteEntity(themeConfiguration)
			return
		}
		api.HandleInternalError(resp, req, err)
		return
	}
	_ = resp.WriteEntity(configMap.Data)
}

type OAuthConfiguration struct {
	Issuer            *oauth.IssuerOptions              `json:"issuer"`
	IdentityProviders []*identityprovider.Configuration `json:"identityProviders"`
	Clients           []*oauth.Client                   `json:"clients"`
}

func (h *handler) getOAuthConfiguration(req *restful.Request, resp *restful.Response) {
	providers := identityprovider.SharedIdentityProviderController.ListConfigurations()
	clients, err := h.oauthClientConfigurationGetter.ListOAuthClients(req.Request.Context())
	if err != nil {
		api.HandleInternalError(resp, req, fmt.Errorf("failed to list OAuth clients: %s", err))
		return
	}
	configuration := OAuthConfiguration{
		Issuer:            h.config.AuthenticationOptions.Issuer,
		IdentityProviders: providers,
		Clients:           clients,
	}
	_ = resp.WriteEntity(configuration)
}

type ClusterConnectionConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Data runtime.RawExtension `json:"data,omitempty"`
}

func (h *handler) listClusterConnectionConfiguration(req *restful.Request, resp *restful.Response) {
	secretList := &corev1.SecretList{}
	if err := h.client.List(req.Request.Context(), secretList, client.InNamespace(constants.KubeSphereNamespace)); err != nil {
		api.HandleError(resp, req, err)
		return
	}

	lists := []ClusterConnectionConfiguration{}
	for _, s := range secretList.Items {
		if s.Type == constants.SecretTypeClusterConnectionConfig {
			config := ClusterConnectionConfiguration{
				TypeMeta: metav1.TypeMeta{
					Kind:       ClusterConnectionConfigurationKind,
					APIVersion: APIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					UID:               s.UID,
					Name:              s.Name,
					ResourceVersion:   s.ResourceVersion,
					CreationTimestamp: s.CreationTimestamp,
				},
			}
			_ = yaml.Unmarshal(s.Data[constants.ClusterConnectionConfigFileName], &config.Data)
			lists = append(lists, config)
		}
	}

	sort.Slice(lists, func(i, j int) bool {
		return lists[i].Name < lists[j].Name
	})
	_ = resp.WriteEntity(lists)
}

func (h *handler) getClusterConnectionConfiguration(req *restful.Request, resp *restful.Response) {
	configName := req.PathParameter("config")
	if len(validation.IsDNS1123Label(configName)) > 0 {
		api.HandleNotFound(resp, req, fmt.Errorf("platform config %s not found", configName))
		return
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.KubeSphereNamespace,
			Name:      configName,
		},
	}
	if err := h.client.Get(req.Request.Context(), client.ObjectKeyFromObject(secret), secret); err != nil {
		if apierrors.IsNotFound(err) {
			api.HandleNotFound(resp, req, fmt.Errorf("cluster connection config %s not found", configName))
			return
		}
		api.HandleError(resp, req, err)
		return
	}

	config := ClusterConnectionConfiguration{
		TypeMeta: metav1.TypeMeta{
			Kind:       ClusterConnectionConfigurationKind,
			APIVersion: APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               secret.UID,
			Name:              secret.Name,
			ResourceVersion:   secret.ResourceVersion,
			CreationTimestamp: secret.CreationTimestamp,
		},
	}
	_ = yaml.Unmarshal(secret.Data[constants.ClusterConnectionConfigFileName], &config.Data)

	_ = resp.WriteEntity(config)
}

func (h *handler) getConfigz(_ *restful.Request, response *restful.Response) {
	_ = response.WriteAsJson(h.config)
}

func (h *handler) getPlatformConfiguration(request *restful.Request, response *restful.Response) {
	configName := request.PathParameter("config")
	if len(validation.IsDNS1123Label(configName)) > 0 {
		api.HandleNotFound(response, request, fmt.Errorf("platform config %s not found", configName))
		return
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.KubeSphereNamespace,
			Name:      fmt.Sprintf(constants.GenericPlatformConfigNameFmt, configName),
		},
	}
	if err := h.client.Get(request.Request.Context(), client.ObjectKeyFromObject(secret), secret); err != nil {
		if apierrors.IsNotFound(err) {
			api.HandleNotFound(response, request, fmt.Errorf("platform config %s not found", configName))
			return
		}
		api.HandleError(response, request, err)
		return
	}

	config := &GenericPlatformConfiguration{}
	config.ConvertFromSecret(secret)

	_ = response.WriteEntity(config)
}

func (h *handler) createPlatformConfiguration(request *restful.Request, response *restful.Response) {
	config := &GenericPlatformConfiguration{}
	if err := request.ReadEntity(config); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	if config.Kind != GenericPlatformConfigurationKind {
		api.HandleBadRequest(response, request, fmt.Errorf("invalid kind: %s", config.Kind))
		return
	}

	if config.APIVersion != APIVersion {
		api.HandleBadRequest(response, request, fmt.Errorf("invalid apiVersion: %s", config.APIVersion))
		return
	}

	if len(validation.IsDNS1123Label(config.Name)) > 0 {
		api.HandleBadRequest(response, request, fmt.Errorf("invalid config name: %s", config.Name))
		return
	}
	secret := config.ConvertToSecret()
	if err := h.client.Create(request.Request.Context(), secret); err != nil {
		api.HandleError(response, request, err)
		return
	}

	config.ConvertFromSecret(secret)

	_ = response.WriteEntity(config)
}

func (h *handler) patchPlatformConfiguration(request *restful.Request, response *restful.Response) {
	var raw map[string]json.RawMessage
	if err := request.ReadEntity(&raw); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	configName := request.PathParameter("config")
	if len(validation.IsDNS1123Label(configName)) > 0 {
		api.HandleNotFound(response, request, fmt.Errorf("platform config %s not found", configName))
		return
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.KubeSphereNamespace,
			Name:      fmt.Sprintf(constants.GenericPlatformConfigNameFmt, configName),
		},
	}

	if err := h.client.Get(request.Request.Context(), client.ObjectKeyFromObject(secret), secret); err != nil {
		if apierrors.IsNotFound(err) {
			api.HandleNotFound(response, request, fmt.Errorf("platform config %s not found", configName))
			return
		}
		api.HandleError(response, request, err)
		return
	}

	config := &GenericPlatformConfiguration{}
	config.ConvertFromSecret(secret)

	original, err := config.Data.MarshalJSON()
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	patchData := raw["data"]
	if len(patchData) > 0 {
		var modifiedData []byte
		modifiedData, err = jsonpatch.MergePatch(original, patchData)
		if err != nil {
			api.HandleBadRequest(response, request, err)
			return
		}

		if err = json.Unmarshal(modifiedData, &config.Data); err != nil {
			api.HandleBadRequest(response, request, err)
			return
		}

		secret = config.ConvertToSecret()
		if err := h.client.Update(request.Request.Context(), secret); err != nil {
			api.HandleError(response, request, err)
			return
		}
	}

	config.ConvertFromSecret(secret)
	_ = response.WriteEntity(config)
}

func (h *handler) updatePlatformConfiguration(request *restful.Request, response *restful.Response) {
	config := &GenericPlatformConfiguration{}
	if err := request.ReadEntity(config); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	secret := config.ConvertToSecret()

	if err := h.client.Update(request.Request.Context(), secret); err != nil {
		api.HandleError(response, request, err)
		return
	}

	config.ConvertFromSecret(secret)

	_ = response.WriteEntity(config)
}

func (h *handler) deletePlatformConfiguration(request *restful.Request, response *restful.Response) {
	configName := request.PathParameter("config")
	if len(validation.IsDNS1123Label(configName)) > 0 {
		api.HandleNotFound(response, request, fmt.Errorf("platform config %s not found", configName))
		return
	}

	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.KubeSphereNamespace,
			Name:      fmt.Sprintf(constants.GenericPlatformConfigNameFmt, configName),
		},
	}

	if err := h.client.Delete(request.Request.Context(), &secret); err != nil {
		api.HandleError(response, request, err)
		return
	}

	_ = response.WriteEntity(errors.None)
}

func (c *GenericPlatformConfiguration) ConvertToSecret() *corev1.Secret {
	yamlData, err := yaml.Marshal(c.Data)
	if err != nil {
		klog.Warningf("Failed to marshal platform configuration: %v", err)
		return nil
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.KubeSphereNamespace,
			Name:      fmt.Sprintf(constants.GenericPlatformConfigNameFmt, c.Name),
		},
		Type: constants.SecretTypeGenericPlatformConfig,
		Data: map[string][]byte{
			constants.GenericPlatformConfigFileName: yamlData,
		},
	}
	return secret
}

func (c *GenericPlatformConfiguration) ConvertFromSecret(secret *corev1.Secret) {
	c.Name = strings.TrimPrefix(secret.Name, fmt.Sprintf(constants.GenericPlatformConfigNameFmt, ""))
	c.CreationTimestamp = secret.CreationTimestamp
	c.ResourceVersion = secret.ResourceVersion
	c.UID = secret.UID
	c.Kind = GenericPlatformConfigurationKind
	c.APIVersion = APIVersion
	_ = yaml.Unmarshal(secret.Data[constants.GenericPlatformConfigFileName], &c.Data)
}
