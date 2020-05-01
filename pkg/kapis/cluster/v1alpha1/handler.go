package v1alpha1

import (
	"bytes"
	"fmt"
	"github.com/emicklei/go-restful"
	"io"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/listers/core/v1"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apis/cluster/v1alpha1"
	clusterlister "kubesphere.io/kubesphere/pkg/client/listers/cluster/v1alpha1"
	"strings"

	"k8s.io/cli-runtime/pkg/printers"
)

const (
	defaultAgentImage = "kubesphere/tower:v1.0"
)

var ErrClusterConnectionIsNotProxy = fmt.Errorf("cluster is not using proxy connection")

type handler struct {
	serviceLister v1.ServiceLister
	clusterLister clusterlister.ClusterLister
	proxyService  string
	proxyAddress  string
	agentImage    string
	yamlPrinter   *printers.YAMLPrinter
}

func NewHandler(serviceLister v1.ServiceLister, clusterLister clusterlister.ClusterLister, proxyService, proxyAddress, agentImage string) *handler {

	if len(agentImage) == 0 {
		agentImage = defaultAgentImage
	}

	return &handler{
		serviceLister: serviceLister,
		clusterLister: clusterLister,
		proxyService:  proxyService,
		proxyAddress:  proxyAddress,
		agentImage:    agentImage,
		yamlPrinter:   &printers.YAMLPrinter{},
	}
}

func (h *handler) GenerateAgentDeployment(request *restful.Request, response *restful.Response) {
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
		return err
	}

	if len(service.Spec.Ports) == 0 {
		return fmt.Errorf("there are no ports in proxy service spec")
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
		return fmt.Errorf("service ingress is empty")
	}

	h.proxyAddress = serviceAddress
	return nil
}

func (h *handler) generateDefaultDeployment(cluster *v1alpha1.Cluster, w io.Writer) error {

	if cluster.Spec.Connection.Type == v1alpha1.ConnectionTypeDirect {
		return ErrClusterConnectionIsNotProxy
	}

	agent := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-agent",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "agent",
							Command: []string{
								"/agent",
								fmt.Sprintf("--name=%s", cluster.Name),
								fmt.Sprintf("--token=%s", cluster.Spec.Connection.Token),
								fmt.Sprintf("--proxy-server=%s", h.proxyAddress),
								fmt.Sprintf("--kubesphere-service=ks-apiserver.kubesphere-system.svc:80"),
								fmt.Sprintf("--kubernetes-service=kubernetes.default.svc:443"),
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
