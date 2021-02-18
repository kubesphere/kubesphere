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

package v1alpha1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/emicklei/go-restful"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/cli-runtime/pkg/printers"
	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apis/cluster/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/config"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	clusterlister "kubesphere.io/kubesphere/pkg/client/listers/cluster/v1alpha1"
	"kubesphere.io/kubesphere/pkg/version"
)

const (
	defaultAgentImage    = "kubesphere/tower:v1.0"
	defaultTimeout       = 10 * time.Second
	KubesphereNamespace  = "kubesphere-system"
	KubeSphereConfigName = "kubesphere-config"
	KubeSphereApiServer  = "ks-apiserver"
)

var errClusterConnectionIsNotProxy = fmt.Errorf("cluster is not using proxy connection")

type handler struct {
	serviceLister   v1.ServiceLister
	clusterLister   clusterlister.ClusterLister
	configMapLister v1.ConfigMapLister

	proxyService string
	proxyAddress string
	agentImage   string
	yamlPrinter  *printers.YAMLPrinter
}

func newHandler(k8sInformers k8sinformers.SharedInformerFactory, ksInformers externalversions.SharedInformerFactory, proxyService, proxyAddress, agentImage string) *handler {

	if len(agentImage) == 0 {
		agentImage = defaultAgentImage
	}

	return &handler{
		serviceLister:   k8sInformers.Core().V1().Services().Lister(),
		clusterLister:   ksInformers.Cluster().V1alpha1().Clusters().Lister(),
		configMapLister: k8sInformers.Core().V1().ConfigMaps().Lister(),

		proxyService: proxyService,
		proxyAddress: proxyAddress,
		agentImage:   agentImage,
		yamlPrinter:  &printers.YAMLPrinter{},
	}
}

// generateAgentDeployment will return a deployment yaml for proxy connection type cluster
// ProxyPublishAddress takes high precedence over proxyPublishService, use proxyPublishService ingress
// address only when proxyPublishAddress is not provided.
func (h *handler) generateAgentDeployment(request *restful.Request, response *restful.Response) {
	clusterName := request.PathParameter("cluster")

	cluster, err := h.clusterLister.Get(clusterName)
	if err != nil {
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		} else {
			api.HandleInternalError(response, request, err)
			return
		}
	}

	if cluster.Spec.Connection.Type != v1alpha1.ConnectionTypeProxy {
		api.HandleNotFound(response, request, fmt.Errorf("cluster %s is not using proxy connection", cluster.Name))
		return
	}

	// use service ingress address
	if len(h.proxyAddress) == 0 {
		err = h.populateProxyAddress()
		if err != nil {
			api.HandleNotFound(response, request, err)
			return
		}
	}

	var buf bytes.Buffer

	err = h.generateDefaultDeployment(cluster, &buf)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	response.Write(buf.Bytes())
}

//
func (h *handler) populateProxyAddress() error {
	if len(h.proxyService) == 0 {
		return fmt.Errorf("neither proxy address nor proxy service provided")
	}
	namespace := "kubesphere-system"
	parts := strings.Split(h.proxyService, ".")
	if len(parts) > 1 && len(parts[1]) != 0 {
		namespace = parts[1]
	}

	service, err := h.serviceLister.Services(namespace).Get(parts[0])
	if err != nil {
		return fmt.Errorf("service %s not found in namespace %s", parts[0], namespace)
	}

	if len(service.Spec.Ports) == 0 {
		return fmt.Errorf("there are no ports in proxy service %s spec", h.proxyService)
	}

	port := service.Spec.Ports[0].Port

	var serviceAddress string
	for _, ingress := range service.Status.LoadBalancer.Ingress {
		if len(ingress.Hostname) != 0 {
			serviceAddress = fmt.Sprintf("http://%s:%d", ingress.Hostname, port)
		}

		if len(ingress.IP) != 0 {
			serviceAddress = fmt.Sprintf("http://%s:%d", ingress.IP, port)
		}
	}

	if len(serviceAddress) == 0 {
		return fmt.Errorf("cannot generate agent deployment yaml for member cluster "+
			" because %s service has no public address, please check %s status, or set address "+
			" mannually in ClusterConfiguration", h.proxyService, h.proxyService)
	}

	h.proxyAddress = serviceAddress
	return nil
}

// Currently, this method works because of serviceaccount/clusterrole/clusterrolebinding already
// created by kubesphere, we don't need to create them again. And it's a little bit inconvenient
// if we want to change the template.
// TODO(jeff): load template from configmap
func (h *handler) generateDefaultDeployment(cluster *v1alpha1.Cluster, w io.Writer) error {

	_, err := url.Parse(h.proxyAddress)
	if err != nil {
		return fmt.Errorf("invalid proxy address %s, should format like http[s]://1.2.3.4:123", h.proxyAddress)
	}

	if cluster.Spec.Connection.Type == v1alpha1.ConnectionTypeDirect {
		return errClusterConnectionIsNotProxy
	}

	agent := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-agent",
			Namespace: "kubesphere-system",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":                       "agent",
					"app.kubernetes.io/part-of": "tower",
				},
			},
			Strategy: appsv1.DeploymentStrategy{},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                       "agent",
						"app.kubernetes.io/part-of": "tower",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "agent",
							Command: []string{
								"/agent",
								fmt.Sprintf("--name=%s", cluster.Name),
								fmt.Sprintf("--token=%s", cluster.Spec.Connection.Token),
								fmt.Sprintf("--proxy-server=%s", h.proxyAddress),
								fmt.Sprintf("--keepalive=10s"),
								fmt.Sprintf("--kubesphere-service=ks-apiserver.kubesphere-system.svc:80"),
								fmt.Sprintf("--kubernetes-service=kubernetes.default.svc:443"),
								fmt.Sprintf("--v=0"),
							},
							Image: h.agentImage,
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("200M"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("100M"),
								},
							},
						},
					},
					ServiceAccountName: "kubesphere",
				},
			},
		},
	}

	return h.yamlPrinter.PrintObj(&agent, w)
}

// ValidateCluster validate cluster kubeconfig and kubesphere apiserver address, check their accessibility
func (h *handler) validateCluster(request *restful.Request, response *restful.Response) {
	var cluster v1alpha1.Cluster

	err := request.ReadEntity(&cluster)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	if cluster.Spec.Connection.Type != v1alpha1.ConnectionTypeDirect {
		api.HandleBadRequest(response, request, fmt.Errorf("cluster connection type MUST be direct"))
		return
	}

	if len(cluster.Spec.Connection.KubeConfig) == 0 {
		api.HandleBadRequest(response, request, fmt.Errorf("cluster kubeconfig MUST NOT be empty"))
		return
	}

	err = h.validateKubeConfig(cluster.Spec.Connection.KubeConfig)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	_, err = validateKubeSphereAPIServer(cluster.Spec.Connection.KubeSphereAPIEndpoint, cluster.Spec.Connection.KubeConfig)
	if err != nil {
		api.HandleBadRequest(response, request, fmt.Errorf("unable validate kubesphere endpoint, %v", err))
		return
	}

	err = h.validateMemberClusterConfiguration(cluster.Spec.Connection.KubeConfig)
	if err != nil {
		api.HandleBadRequest(response, request, fmt.Errorf("failed to validate member cluster configuration, err: %v", err))
	}

	response.WriteHeader(http.StatusOK)
}

// validateKubeConfig takes base64 encoded kubeconfig and check its validity
func (h *handler) validateKubeConfig(kubeconfig []byte) error {
	config, err := loadKubeConfigFromBytes(kubeconfig)
	if err != nil {
		return err
	}

	clusters, err := h.clusterLister.List(labels.Everything())
	if err != nil {
		return err
	}

	// clusters with the exactly same KubernetesAPIEndpoint considered to be one
	// MUST not import the same cluster twice
	for _, cluster := range clusters {
		if len(cluster.Spec.Connection.KubernetesAPIEndpoint) != 0 && cluster.Spec.Connection.KubernetesAPIEndpoint == config.Host {
			return fmt.Errorf("existing cluster %s with the exacty same server address, MUST not import the same cluster twice", cluster.Name)
		}
	}

	config.Timeout = defaultTimeout

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	_, err = clientSet.Discovery().ServerVersion()

	return err
}

func loadKubeConfigFromBytes(kubeconfig []byte) (*rest.Config, error) {
	clientConfig, err := clientcmd.NewClientConfigFromBytes(kubeconfig)
	if err != nil {
		return nil, err
	}

	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	return config, nil
}

// validateKubeSphereAPIServer uses version api to check the accessibility
// If kubesphere apiserver endpoint is not provided, use kube-apiserver proxy instead
func validateKubeSphereAPIServer(ksEndpoint string, kubeconfig []byte) (*version.Info, error) {
	if len(ksEndpoint) == 0 && len(kubeconfig) == 0 {
		return nil, fmt.Errorf("neither kubesphere api endpoint nor kubeconfig was provided")
	}

	client := http.Client{
		Timeout: defaultTimeout,
	}

	path := fmt.Sprintf("%s/kapis/version", ksEndpoint)

	if len(ksEndpoint) != 0 {
		_, err := url.Parse(ksEndpoint)
		if err != nil {
			return nil, err
		}
	} else {
		config, err := loadKubeConfigFromBytes(kubeconfig)
		if err != nil {
			return nil, err
		}

		transport, err := rest.TransportFor(config)
		if err != nil {
			return nil, err
		}

		client.Transport = transport
		path = fmt.Sprintf("%s/api/v1/namespaces/%s/services/:%s:/proxy/kapis/version", config.Host, KubesphereNamespace, KubeSphereApiServer)
	}

	response, err := client.Get(path)
	if err != nil {
		return nil, err
	}

	responseBytes, _ := ioutil.ReadAll(response.Body)
	responseBody := string(responseBytes)

	response.Body = ioutil.NopCloser(bytes.NewBuffer(responseBytes))

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response: %s , please make sure %s.%s.svc of member cluster is up and running", KubeSphereApiServer, KubesphereNamespace, responseBody)
	}

	ver := version.Info{}
	err = json.NewDecoder(response.Body).Decode(&ver)
	if err != nil {
		return nil, fmt.Errorf("invalid response: %s , please make sure %s.%s.svc of member cluster is up and running", KubeSphereApiServer, KubesphereNamespace, responseBody)
	}

	return &ver, nil
}

// validateMemberClusterConfiguration compares host and member cluster jwt, if they are not same, it changes member
// cluster jwt to host's, then restart member cluster ks-apiserver.
func (h *handler) validateMemberClusterConfiguration(memberKubeconfig []byte) error {
	hConfig, err := h.getHostClusterConfig()
	if err != nil {
		return err
	}

	mConfig, err := h.getMemberClusterConfig(memberKubeconfig)
	if err != nil {
		return err
	}

	if hConfig.AuthenticationOptions.JwtSecret != mConfig.AuthenticationOptions.JwtSecret {
		return fmt.Errorf("hostcluster Jwt is not equal to member cluster jwt, please edit the member cluster cluster config")
	}

	return nil
}

// getMemberClusterConfig returns KubeSphere running config by the given member cluster kubeconfig
func (h *handler) getMemberClusterConfig(kubeconfig []byte) (*config.Config, error) {
	config, err := loadKubeConfigFromBytes(kubeconfig)
	if err != nil {
		return nil, err
	}

	config.Timeout = defaultTimeout

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	memberCm, err := clientSet.CoreV1().ConfigMaps(KubesphereNamespace).Get(context.Background(), KubeSphereConfigName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return getConfigFromCm(memberCm)
}

// getHostClusterConfig returns KubeSphere running config from host cluster ConfigMap
func (h *handler) getHostClusterConfig() (*config.Config, error) {
	hostCm, err := h.configMapLister.ConfigMaps(KubesphereNamespace).Get(KubeSphereConfigName)
	if err != nil {
		return nil, fmt.Errorf("failed to get host cluster %s/configmap/%s, err: %s", KubesphereNamespace, KubeSphereConfigName, err)
	}

	return getConfigFromCm(hostCm)
}

// getConfigFromCm returns KubeSphere ruuning config by the given ConfigMap.
func getConfigFromCm(cm *corev1.ConfigMap) (*config.Config, error) {
	Config := config.New()

	value, ok := cm.Data["kubesphere.yaml"]
	if !ok {
		return nil, fmt.Errorf("failed to get %s/configmap/%s kubesphere.yaml value", KubesphereNamespace, KubeSphereConfigName)
	}

	err := yaml.Unmarshal([]byte(value), Config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal value from %s/configmap/%s. err: %s", KubesphereNamespace, KubeSphereConfigName, err)
	}

	return Config, nil
}
