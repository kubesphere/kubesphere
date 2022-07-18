package v1alpha2

import (
	"fmt"
	"regexp"

	"github.com/emicklei/go-restful"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/informers"
)

const (
	defaultDeployName = "openelb-manager"
	defaultNamespace  = "openelb-system"
	openelbCmd        = defaultDeployName
)

type handler struct {
	informer informers.InformerFactory
}

func (h handler) detectOpenELB(request *restful.Request, response *restful.Response) {
	// check crd
	if !h.checkCRDInstall(request, response) {
		klog.Info("The crd of openelb is not installed or incomplete installation")
		response.WriteEntity(false)
		return
	}

	//  get deployment in openelb-system/all namespace
	deploy := h.getDeploy(request, response)
	if deploy == nil {
		klog.Info("There is no openelb deployment")
		response.WriteEntity(false)
		return
	}

	// check openelb deploy
	if !h.checkDeploy(deploy) {
		klog.Info("Check openelb deploy and it is not openelb deployment")
		response.WriteEntity(false)
		return
	}
	response.WriteEntity(true)
}

func (h handler) matchDeploy(name string) bool {
	if name == defaultDeployName {
		return true
	}

	r1 := regexp.MustCompile("openelb-(.*)-manager")
	r2 := regexp.MustCompile("(.+)-openelb-manager")

	return r1.MatchString(name) || r2.MatchString(name)
}

func (h handler) checkCRDInstall(request *restful.Request, response *restful.Response) bool {
	resourceList := []string{
		"eips.network.kubesphere.io",
		"bgppeers.network.kubesphere.io",
		"bgpconfs.network.kubesphere.io",
	}
	count := len(resourceList)

	crds, err := h.informer.ApiExtensionSharedInformerFactory().Apiextensions().V1().CustomResourceDefinitions().Lister().List(labels.Everything())
	if err != nil {
		klog.Error(err)
		api.HandleInternalError(response, request, err)
		return false
	}

	for _, crd := range crds {
		for _, r := range resourceList {
			if crd.Name == r {
				count--
			}
		}
	}

	return count == 0
}

func (h handler) getDeploy(request *restful.Request, response *restful.Response) (deploy *v1.Deployment) {
	label := map[string]string{
		"app":           "openelb-manager",
		"control-plane": "openelb-manager",
	}

	// openelb-system
	deployments, err := h.informer.KubernetesSharedInformerFactory().Apps().V1().
		Deployments().Lister().Deployments(defaultNamespace).List(labels.SelectorFromSet(label))
	if err != nil {
		klog.Error(err)
		api.HandleInternalError(response, request, err)
		return nil
	}
	if len(deployments) == 0 {
		klog.V(1).Info(fmt.Sprintf("no openelb deployment in %s namespace, may installed in other namespace", defaultNamespace))
	}
	for _, d := range deployments {
		if h.matchDeploy(d.Name) {
			return d
		}
	}

	// all namespace
	deployments, err = h.informer.KubernetesSharedInformerFactory().Apps().V1().
		Deployments().Lister().List(labels.SelectorFromSet(label))
	if err != nil {
		klog.Error(err)
		api.HandleInternalError(response, request, err)
		return nil
	}
	for _, d := range deployments {
		if h.matchDeploy(d.Name) {
			return d
		}
	}

	return nil
}

func (h handler) checkDeploy(deploy *v1.Deployment) bool {
	if deploy == nil || len(deploy.Spec.Template.Spec.Containers) == 0 ||
		len(deploy.Spec.Template.Spec.Containers[0].Command) == 0 {
		return false
	}

	if deploy.Spec.Template.Spec.Containers[0].Command[0] != openelbCmd {
		return false
	}

	return true
}
