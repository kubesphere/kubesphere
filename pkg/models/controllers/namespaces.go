/*
Copyright 2018 The KubeSphere Authors.

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

package controllers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/kubernetes/pkg/util/slice"

	"k8s.io/client-go/informers"

	"k8s.io/kubernetes/pkg/apis/core"

	utilversion "k8s.io/kubernetes/pkg/util/version"

	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/options"
)

const (
	provider                     = "kubernetes"
	admin                        = "admin"
	operator                     = "operator"
	viewer                       = "viewer"
	kubectlNamespace             = constants.KubeSphereControlNamespace
	kubectlConfigKey             = "config"
	openPitrixRuntimeAnnotateKey = "openpitrix_runtime"
	creatorAnnotateKey           = "creator"
	initTimeAnnotateKey          = "kubesphere.io/init-time"
	workspaceLabelKey            = "kubesphere.io/workspace"
)

var adminRules = []rbac.PolicyRule{{Verbs: []string{"*"}, APIGroups: []string{"*"}, Resources: []string{"*"}}}
var editorRules = []rbac.PolicyRule{{Verbs: []string{"*"}, APIGroups: []string{"", "apps", "extensions", "batch", "kubesphere.io", "account.kubesphere.io"}, Resources: []string{"*"}}}
var viewerRules = []rbac.PolicyRule{{Verbs: []string{"list", "get", "watch"}, APIGroups: []string{"", "apps", "extensions", "batch", "kubesphere.io", "account.kubesphere.io"}, Resources: []string{"*"}}}

type runTime struct {
	RuntimeId         string `json:"runtime_id"`
	RuntimeUrl        string `json:"runtime_url"`
	Name              string `json:"name"`
	Provider          string `json:"provider"`
	Zone              string `json:"zone"`
	RuntimeCredential string `json:"runtime_credential"`
}

type DeleteRunTime struct {
	RuntimeId []string `json:"runtime_id"`
}

func (ctl *NamespaceCtl) getKubeConfig(user string) (string, error) {
	k8sClient := client.NewK8sClient()
	configmap, err := k8sClient.CoreV1().ConfigMaps(kubectlNamespace).Get(user, metaV1.GetOptions{})
	if err != nil {
		glog.Errorln(err)
		return "", err
	}
	return configmap.Data[kubectlConfigKey], nil
}

func (ctl *NamespaceCtl) deleteOpRuntime(item v1.Namespace) {

	var runtimeId string
	if item.Annotations == nil {
		runtimeId = ""
	} else {
		runtimeId = item.Annotations[openPitrixRuntimeAnnotateKey]
	}

	if len(runtimeId) == 0 {
		return
	}

	url := options.ServerOptions.GetOpAddress() + "/v1/runtimes"

	var deleteRuntime = DeleteRunTime{RuntimeId: []string{runtimeId}}

	body, err := json.Marshal(deleteRuntime)
	if err != nil {
		glog.Error("runtime release failed:", item.Name, runtimeId, err)
		return
	}

	glog.Info("runtime release succeeded:", item.Name, runtimeId)

	// todo: if delete failed, what's to be done?
	makeHttpRequest("DELETE", url, string(body))
}

func (ctl *NamespaceCtl) createOpRuntime(namespace string) ([]byte, error) {
	zone := namespace
	name := namespace
	kubeConfig, err := ctl.getKubeConfig("admin")
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	url := options.ServerOptions.GetOpAddress() + "/v1/runtimes"

	option := runTime{Name: name, Provider: provider, RuntimeCredential: kubeConfig, Zone: zone}
	body, err := json.Marshal(option)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	return makeHttpRequest("POST", url, string(body))
}

func (ctl *NamespaceCtl) updateSystemRoleBindings(namespace *v1.Namespace) error {

	workspace := ""

	if namespace.Labels != nil {
		workspace = namespace.Labels[workspaceLabelKey]
	}

	adminBinding, err := ctl.K8sClient.RbacV1().RoleBindings(namespace.Name).Get(admin, metaV1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			adminBinding = new(rbac.RoleBinding)
			adminBinding.Name = admin
			adminBinding.Namespace = namespace.Name
			adminBinding.RoleRef = rbac.RoleRef{Kind: "Role", Name: admin}
		} else {
			return err
		}
	}

	adminBinding.Subjects = make([]rbac.Subject, 0)

	if workspace != "" {
		workspaceAdmin, err := ctl.K8sClient.RbacV1().ClusterRoleBindings().Get(fmt.Sprintf("system:%s:%s", workspace, constants.WorkspaceAdmin), metaV1.GetOptions{})
		if err != nil {
			return err
		}
		adminBinding.Subjects = append(adminBinding.Subjects, workspaceAdmin.Subjects...)
	}

	if adminBinding.ResourceVersion == "" {
		_, err = ctl.K8sClient.RbacV1().RoleBindings(namespace.Name).Create(adminBinding)
	} else {
		_, err = ctl.K8sClient.RbacV1().RoleBindings(namespace.Name).Update(adminBinding)
	}

	if err != nil {
		return err
	}

	viewerBinding, err := ctl.K8sClient.RbacV1().RoleBindings(namespace.Name).Get(viewer, metaV1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			viewerBinding = new(rbac.RoleBinding)
			viewerBinding.Name = viewer
			viewerBinding.Namespace = namespace.Name
			viewerBinding.RoleRef = rbac.RoleRef{Kind: "Role", Name: viewer}
		} else {
			return err
		}
	}

	viewerBinding.Subjects = make([]rbac.Subject, 0)

	if workspace != "" {
		workspaceViewer, err := ctl.K8sClient.RbacV1().ClusterRoleBindings().Get(fmt.Sprintf("system:%s:%s", workspace, constants.WorkspaceViewer), metaV1.GetOptions{})
		if err != nil {
			return err
		}
		viewerBinding.Subjects = append(viewerBinding.Subjects, workspaceViewer.Subjects...)
	}

	if viewerBinding.ResourceVersion == "" {
		_, err = ctl.K8sClient.RbacV1().RoleBindings(namespace.Name).Create(viewerBinding)
	} else {
		_, err = ctl.K8sClient.RbacV1().RoleBindings(namespace.Name).Update(viewerBinding)
	}

	if err != nil {
		return err
	}

	return nil
}

func (ctl *NamespaceCtl) createDefaultRoleBinding(namespace *v1.Namespace) error {
	creator := ""
	if namespace.Annotations != nil {
		creator = namespace.Annotations[creatorAnnotateKey]
	}
	// create once
	if creator != "" {
		creatorBindingName := fmt.Sprintf("%s-admin", creator)
		creatorBinding, err := ctl.K8sClient.RbacV1().RoleBindings(namespace.Name).Get(creatorBindingName, metaV1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				creatorBinding = new(rbac.RoleBinding)
				creatorBinding.Name = creatorBindingName
				creatorBinding.Namespace = namespace.Name
				creatorBinding.RoleRef = rbac.RoleRef{Kind: "Role", Name: admin}
			} else {
				return err
			}
		}

		creatorBinding.Subjects = []rbac.Subject{{Kind: rbac.UserKind, Name: creator}}

		if creatorBinding.ResourceVersion == "" {
			_, err = ctl.K8sClient.RbacV1().RoleBindings(namespace.Name).Create(creatorBinding)
		} else {
			_, err = ctl.K8sClient.RbacV1().RoleBindings(namespace.Name).Update(creatorBinding)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (ctl *NamespaceCtl) CreateDefaultRoleAndRoleBinding(namespace *v1.Namespace) error {
	adminRole := &rbac.Role{ObjectMeta: metaV1.ObjectMeta{Name: admin, Namespace: namespace.Name, Annotations: map[string]string{creatorAnnotateKey: "system"}}, Rules: adminRules}
	operatorRole := &rbac.Role{ObjectMeta: metaV1.ObjectMeta{Name: operator, Namespace: namespace.Name, Annotations: map[string]string{creatorAnnotateKey: "system"}}, Rules: editorRules}
	viewerRole := &rbac.Role{ObjectMeta: metaV1.ObjectMeta{Name: viewer, Namespace: namespace.Name, Annotations: map[string]string{creatorAnnotateKey: "system"}}, Rules: viewerRules}

	_, err := ctl.K8sClient.RbacV1().Roles(namespace.Name).Create(adminRole)

	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		if err := ctl.createDefaultRoleBinding(namespace); err != nil {
			glog.Warning("default role binding create failed", namespace.Name)
		}
	}

	_, err = ctl.K8sClient.RbacV1().Roles(namespace.Name).Create(operatorRole)

	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	_, err = ctl.K8sClient.RbacV1().Roles(namespace.Name).Create(viewerRole)

	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func (ctl *NamespaceCtl) createRoleAndRuntime(namespace *v1.Namespace) {
	runtime := ""
	initTime := ""
	if namespace.Annotations != nil {
		runtime = namespace.Annotations[openPitrixRuntimeAnnotateKey]
		initTime = namespace.Annotations[initTimeAnnotateKey]
	}

	componentsNamespaces := []string{constants.KubeSystemNamespace, constants.OpenPitrixNamespace, constants.IstioNamespace, constants.KubeSphereNamespace}

	if runtime == "" && !slice.ContainsString(componentsNamespaces, namespace.Name, nil) {
		_, runtimeCreateError := ctl.createOpRuntime(namespace.Name)
		if runtimeCreateError != nil {
			glog.Error("runtime create error:", runtimeCreateError)
		}
	}

	if initTime == "" {
		err := ctl.CreateDefaultRoleAndRoleBinding(namespace)
		if err == nil {
			err = ctl.updateSystemRoleBindings(namespace)
			if err != nil {
				glog.Error("role binding update error:", err)
			}
		} else {
			glog.Error("default role create error:", err)
		}
		if err == nil {
			pathJson := fmt.Sprintf(`{"metadata":{"annotations":{"%s":"%s"}}}`, initTimeAnnotateKey, time.Now().UTC().Format("2006-01-02T15:04:05Z"))
			_, err = ctl.K8sClient.CoreV1().Namespaces().Patch(namespace.Name, "application/strategic-merge-patch+json", []byte(pathJson))
			if err != nil {
				glog.Error("annotations patch error init failed:", namespace.Name, err)
			}
		}
	}
}

func (ctl *NamespaceCtl) createCephSecretAfterNewNs(item v1.Namespace) {
	// Kubernetes version must < 1.11.0
	verInfo, err := ctl.K8sClient.ServerVersion()
	if err != nil {
		glog.Error("consult k8s server error: ", err)
		return
	}
	if !utilversion.MustParseSemantic(verInfo.String()).LessThan(utilversion.MustParseSemantic("v1.11.0")) {
		glog.Infof("disable Ceph secret controller due to k8s version %s >= v1.11.0", verInfo.String())
		return
	}

	// Create Ceph secret in the new namespace
	newNsName := item.Name
	scList, _ := ctl.K8sClient.StorageV1().StorageClasses().List(metaV1.ListOptions{})
	if scList == nil {
		return
	}
	for _, sc := range scList.Items {
		if sc.Provisioner == rbdPluginName {
			glog.Infof("would create Ceph user secret in storage class %s at namespace %s", sc.GetName(), newNsName)
			if secretName, ok := sc.Parameters[rbdUserSecretNameKey]; ok {
				secret, err := ctl.K8sClient.CoreV1().Secrets(core.NamespaceSystem).Get(secretName, metaV1.GetOptions{})
				if err != nil {
					if errors.IsNotFound(err) {
						glog.Errorf("cannot find secret in namespace %s, error: %s", core.NamespaceSystem, err.Error())
						continue
					}
					glog.Errorf("failed to find secret in namespace %s, error: %s", core.NamespaceSystem, err.Error())
					continue
				}
				glog.Infof("succeed to find secret %s in namespace %s", secret.GetName(), secret.GetNamespace())

				newSecret := &v1.Secret{
					TypeMeta: metaV1.TypeMeta{
						Kind:       secret.Kind,
						APIVersion: secret.APIVersion,
					},
					ObjectMeta: metaV1.ObjectMeta{
						Name:                       secret.GetName(),
						Namespace:                  newNsName,
						Labels:                     secret.GetLabels(),
						Annotations:                secret.GetAnnotations(),
						DeletionGracePeriodSeconds: secret.GetDeletionGracePeriodSeconds(),
						ClusterName:                secret.GetClusterName(),
					},
					Data:       secret.Data,
					StringData: secret.StringData,
					Type:       secret.Type,
				}
				glog.Infof("creating secret %s in namespace %s...", newSecret.GetName(), newSecret.GetNamespace())
				_, err = ctl.K8sClient.CoreV1().Secrets(newSecret.GetNamespace()).Create(newSecret)
				if err != nil {
					glog.Errorf("failed to create secret in namespace %s, error: %v", newSecret.GetNamespace(), err)
					continue
				}
			} else {
				glog.Errorf("failed to find user secret name in storage class %s", sc.GetName())
			}
		}
	}
}

func (ctl *NamespaceCtl) generateObject(item *v1.Namespace) *Namespace {
	var displayName string

	if item.Annotations != nil && len(item.Annotations[DisplayName]) > 0 {
		displayName = item.Annotations[DisplayName]
	}

	name := item.Name
	createTime := item.CreationTimestamp.Time
	status := fmt.Sprintf("%v", item.Status.Phase)

	if createTime.IsZero() {
		createTime = time.Now()
	}

	object := &Namespace{
		Name:        name,
		DisplayName: displayName,
		CreateTime:  createTime,
		Status:      status,
		Annotation:  MapString{item.Annotations},
	}

	return object
}

func (ctl *NamespaceCtl) Name() string {
	return ctl.CommonAttribute.Name
}

func (ctl *NamespaceCtl) sync(stopChan chan struct{}) {
	db := ctl.DB

	if db.HasTable(&Namespace{}) {
		db.DropTable(&Namespace{})
	}

	db = db.CreateTable(&Namespace{})

	ctl.initListerAndInformer()
	//list, err := ctl.lister.List(labels.Everything())
	//if err != nil {
	//	glog.Error(err)
	//	return
	//}

	//for _, item := range list {
	//	obj := ctl.generateObject(item)
	//	db.Create(obj)
	//  ctl.createRoleAndRuntime(item)
	//}

	ctl.informer.Run(stopChan)
}

func (ctl *NamespaceCtl) total() int {
	list, err := ctl.lister.List(labels.Everything())
	if err != nil {
		glog.Errorf("count %s falied, reason:%s", err, ctl.Name())
		return 0
	}
	return len(list)
}

func (ctl *NamespaceCtl) initListerAndInformer() {
	db := ctl.DB

	informerFactory := informers.NewSharedInformerFactory(ctl.K8sClient, time.Second*resyncCircle)

	ctl.lister = informerFactory.Core().V1().Namespaces().Lister()

	informer := informerFactory.Core().V1().Namespaces().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {

			object := obj.(*v1.Namespace)
			mysqlObject := ctl.generateObject(object)
			db.Create(mysqlObject)
			ctl.createRoleAndRuntime(object)
			ctl.createCephSecretAfterNewNs(*object)
		},
		UpdateFunc: func(old, new interface{}) {
			object := new.(*v1.Namespace)
			mysqlObject := ctl.generateObject(object)
			db.Save(mysqlObject)
			ctl.createRoleAndRuntime(object)
		},
		DeleteFunc: func(obj interface{}) {
			var item Namespace
			object := obj.(*v1.Namespace)
			db.Where("name=?", object.Name).Find(&item)
			db.Delete(item)
			ctl.deleteOpRuntime(*object)
		},
	})

	ctl.informer = informer
}

func (ctl *NamespaceCtl) CountWithConditions(conditions string) int {
	var object Namespace

	return countWithConditions(ctl.DB, conditions, &object)
}

func (ctl *NamespaceCtl) ListWithConditions(conditions string, paging *Paging, order string) (int, interface{}, error) {
	var list []Namespace
	var object Namespace
	var total int

	if len(order) == 0 {
		order = "createTime desc"
	}

	listWithConditions(ctl.DB, &total, &object, &list, conditions, paging, order)

	if paging != nil {
		for index := range list {
			usage, err := ctl.GetNamespaceQuota(list[index].Name)
			if err == nil {
				list[index].Usage = usage
			}
		}
	}

	return total, list, nil
}

func getUsage(namespace, resource string) int {
	ctl := ResourceControllers.Controllers[resource]
	return ctl.CountWithConditions(fmt.Sprintf("namespace = '%s' ", namespace))
}

func (ctl *NamespaceCtl) GetNamespaceQuota(namespace string) (v1.ResourceList, error) {

	usage := make(v1.ResourceList)

	resourceList := []string{Daemonsets, Deployments, Ingresses, Roles, Services, Statefulsets, PersistentVolumeClaim, Pods, Jobs, Cronjobs}
	for _, resourceName := range resourceList {
		used := getUsage(namespace, resourceName)
		var quantity resource.Quantity
		quantity.Set(int64(used))
		usage[v1.ResourceName(resourceName)] = quantity
	}

	podCtl := ResourceControllers.Controllers[Pods]
	var quantity resource.Quantity
	used := podCtl.CountWithConditions(fmt.Sprintf("status=\"%s\" And namespace=\"%s\"", "Running", namespace))
	quantity.Set(int64(used))
	usage["runningPods"] = quantity
	return usage, nil
}

func (ctl *NamespaceCtl) Lister() interface{} {

	return ctl.lister
}
