// Copyright 2022 The KubeSphere Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"

	"k8s.io/apimachinery/pkg/types"

	"github.com/emicklei/go-restful"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"kubesphere.io/api/notification/v2beta1"
	"kubesphere.io/api/notification/v2beta2"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	"kubesphere.io/kubesphere/pkg/simple/client/notification"
)

const (
	Secret              = "secrets"
	ConfigMap           = "configmaps"
	VerificationAPIPath = "/api/v2/verify"

	V2beta1 = "v2beta1"
	V2beta2 = "v2beta2"
)

type Operator interface {
	ListV2beta1(user, resource, subresource string, query *query.Query) (*api.ListResult, error)
	GetV2beta1(user, resource, name, subresource string) (runtime.Object, error)
	CreateV2beta1(user, resource string, obj runtime.Object) (runtime.Object, error)
	DeleteV2beta1(user, resource, name string) error
	UpdateV2beta1(user, resource, name string, obj runtime.Object) (runtime.Object, error)

	List(user, resource, subresource string, query *query.Query) (*api.ListResult, error)
	Get(user, resource, name, subresource string) (runtime.Object, error)
	Create(user, resource string, obj runtime.Object) (runtime.Object, error)
	Delete(user, resource, name string) error
	Update(user, resource, name string, obj runtime.Object) (runtime.Object, error)
	Patch(user, resource, name string, data []byte) (runtime.Object, error)

	Verify(request *restful.Request, response *restful.Response)

	GetObject(resource, version string) runtime.Object
	IsKnownResource(resource, version, subresource string) bool
}

type operator struct {
	k8sClient      kubernetes.Interface
	ksClient       kubesphere.Interface
	informers      informers.InformerFactory
	resourceGetter *resource.ResourceGetter
	options        *notification.Options
}

type Data struct {
	Config   v2beta2.Config   `json:"config"`
	Receiver v2beta2.Receiver `json:"receiver"`
}

type Result struct {
	Code    int    `json:"Status"`
	Message string `json:"Message"`
}

func NewOperator(
	informers informers.InformerFactory,
	k8sClient kubernetes.Interface,
	ksClient kubesphere.Interface,
	options *notification.Options) Operator {

	return &operator{
		informers:      informers,
		k8sClient:      k8sClient,
		ksClient:       ksClient,
		resourceGetter: resource.NewResourceGetter(informers, nil),
		options:        options,
	}
}

// ListV2beta1 list objects of version v2beta1. Only global objects will be returned if the user is nil.
// If the user is not nil, only tenant objects whose tenant label matches the user will be returned.
func (o *operator) ListV2beta1(user, resource, subresource string, q *query.Query) (*api.ListResult, error) {
	return o.list(user, resource, V2beta1, subresource, q)
}

// List objects. Only global objects will be returned if the user is nil.
// If the user is not nil, only tenant objects whose tenant label matches the user will be returned.
func (o *operator) List(user, resource, subresource string, q *query.Query) (*api.ListResult, error) {
	return o.list(user, resource, V2beta2, subresource, q)
}

func (o *operator) list(user, resource, version, subresource string, q *query.Query) (*api.ListResult, error) {

	if user != "" {
		if resource == v2beta2.ResourcesPluralRouter ||
			resource == v2beta2.ResourcesPluralNotificationManager {
			return nil, errors.NewForbidden(v2beta2.Resource(resource), "",
				fmt.Errorf("tenant can not list %s", resource))
		}
	}

	q.LabelSelector = o.generateLabelSelector(q, user, resource, version)

	ns := ""
	if resource == Secret || resource == ConfigMap {
		ns = constants.NotificationSecretNamespace
	}

	res, err := o.resourceGetter.List(resource, ns, q)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if subresource == "" ||
		(resource != v2beta2.ResourcesPluralConfig && resource != v2beta2.ResourcesPluralReceiver) {
		return res, nil
	}

	results := &api.ListResult{}
	for _, item := range res.Items {
		obj := clean(item, resource, subresource)
		if obj != nil {
			if version == V2beta1 {
				obj = convert(obj)
			}
			results.Items = append(results.Items, obj)
		}
	}
	results.TotalItems = len(results.Items)

	return results, nil
}

func (o *operator) generateLabelSelector(q *query.Query, user, resource, version string) string {

	if resource == v2beta2.ResourcesPluralNotificationManager {
		return q.LabelSelector
	}

	labelSelector := q.LabelSelector
	if len(labelSelector) > 0 {
		labelSelector = q.LabelSelector + ","
	}

	filter := ""
	// If user is nil, it will list all global object.
	if user == "" {
		if isConfig(o.GetObject(resource, version)) {
			filter = "type=default"
		} else {
			filter = "type=global"
		}
	} else {
		// If the user is not nil, only return the object belong to this user.
		filter = "type=tenant,user=" + user
	}

	labelSelector = labelSelector + filter
	return labelSelector
}

// GetV2beta1 get the specified object of version v2beta1, if you want to get a global object, the user must be nil.
// If you want to get a tenant object, the user must equal to the tenant specified in labels of the object.
func (o *operator) GetV2beta1(user, resource, name, subresource string) (runtime.Object, error) {
	return o.get(user, resource, V2beta1, name, subresource)
}

// Get the specified object, if you want to get a global object, the user must be nil.
// If you want to get a tenant object, the user must equal to the tenant specified in labels of the object.
func (o *operator) Get(user, resource, name, subresource string) (runtime.Object, error) {
	return o.get(user, resource, V2beta2, name, subresource)
}

func (o *operator) get(user, resource, version, name, subresource string) (runtime.Object, error) {

	if user != "" {
		if resource == v2beta2.ResourcesPluralRouter ||
			resource == v2beta2.ResourcesPluralNotificationManager {
			return nil, errors.NewForbidden(v2beta2.Resource(resource), "",
				fmt.Errorf("tenant can not get %s", resource))
		}
	}

	ns := ""
	if resource == Secret || resource == ConfigMap {
		ns = constants.NotificationSecretNamespace
	}

	obj, err := o.resourceGetter.Get(resource, ns, name)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if err := authorizer(user, obj); err != nil {
		klog.Error(err)
		return nil, err
	}

	if subresource == "" ||
		(resource != v2beta2.ResourcesPluralConfig && resource != v2beta2.ResourcesPluralReceiver) {
		return obj, nil
	}

	res := clean(obj, resource, subresource)
	if res == nil {
		return nil, errors.NewNotFound(v2beta1.Resource(obj.GetObjectKind().GroupVersionKind().GroupKind().Kind), name)
	}

	if version == V2beta1 {
		res = convert(res)
	}

	return res, nil
}

// CreateV2beta1 an object of version v2beta1. A global object will be created if the user is nil.
// A tenant object will be created if the user is not nil.
func (o *operator) CreateV2beta1(user, resource string, obj runtime.Object) (runtime.Object, error) {
	return o.create(user, resource, V2beta1, obj)
}

// Create an object. A global object will be created if the user is nil.
// A tenant object will be created if the user is not nil.
func (o *operator) Create(user, resource string, obj runtime.Object) (runtime.Object, error) {
	return o.create(user, resource, V2beta2, obj)
}

func (o *operator) create(user, resource, version string, obj runtime.Object) (runtime.Object, error) {

	if user != "" {
		if resource == v2beta2.ResourcesPluralRouter ||
			resource == v2beta2.ResourcesPluralNotificationManager {
			return nil, errors.NewForbidden(v2beta2.Resource(resource), "",
				fmt.Errorf("tenant can not create %s", resource))
		}
	}

	if err := appendLabel(user, resource, obj); err != nil {
		return nil, err
	}

	switch resource {
	case v2beta2.ResourcesPluralNotificationManager:
		return o.ksClient.NotificationV2beta2().NotificationManagers().Create(context.Background(), obj.(*v2beta2.NotificationManager), v1.CreateOptions{})
	case v2beta2.ResourcesPluralConfig:
		if version == V2beta1 {
			return o.ksClient.NotificationV2beta1().Configs().Create(context.Background(), obj.(*v2beta1.Config), v1.CreateOptions{})
		} else {
			return o.ksClient.NotificationV2beta2().Configs().Create(context.Background(), obj.(*v2beta2.Config), v1.CreateOptions{})
		}
	case v2beta2.ResourcesPluralReceiver:
		if version == V2beta1 {
			return o.ksClient.NotificationV2beta1().Receivers().Create(context.Background(), obj.(*v2beta1.Receiver), v1.CreateOptions{})
		} else {
			return o.ksClient.NotificationV2beta2().Receivers().Create(context.Background(), obj.(*v2beta2.Receiver), v1.CreateOptions{})
		}
	case v2beta2.ResourcesPluralRouter:
		return o.ksClient.NotificationV2beta2().Routers().Create(context.Background(), obj.(*v2beta2.Router), v1.CreateOptions{})
	case v2beta2.ResourcesPluralSilence:
		return o.ksClient.NotificationV2beta2().Silences().Create(context.Background(), obj.(*v2beta2.Silence), v1.CreateOptions{})
	case Secret:
		return o.k8sClient.CoreV1().Secrets(constants.NotificationSecretNamespace).Create(context.Background(), obj.(*corev1.Secret), v1.CreateOptions{})
	case ConfigMap:
		return o.k8sClient.CoreV1().ConfigMaps(constants.NotificationSecretNamespace).Create(context.Background(), obj.(*corev1.ConfigMap), v1.CreateOptions{})
	default:
		return nil, errors.NewInternalError(nil)
	}
}

// DeleteV2beta1 an object of version v2beta1. A global object will be deleted if the user is nil.
// If the user is not nil, a tenant object whose tenant label matches the user will be deleted.
func (o *operator) DeleteV2beta1(user, resource, name string) error {
	return o.delete(user, resource, name)
}

// Delete an object. A global object will be deleted if the user is nil.
// If the user is not nil, a tenant object whose tenant label matches the user will be deleted.
func (o *operator) Delete(user, resource, name string) error {
	return o.delete(user, resource, name)
}

func (o *operator) delete(user, resource, name string) error {

	if user != "" {
		if resource == v2beta2.ResourcesPluralRouter ||
			resource == v2beta2.ResourcesPluralNotificationManager {
			return errors.NewForbidden(v2beta2.Resource(resource), "",
				fmt.Errorf("tenant can not delete %s", resource))
		}
	}

	if _, err := o.Get(user, resource, name, ""); err != nil {
		klog.Error(err)
		return err
	}

	switch resource {
	case v2beta2.ResourcesPluralNotificationManager:
		return o.ksClient.NotificationV2beta2().NotificationManagers().Delete(context.Background(), name, v1.DeleteOptions{})
	case v2beta2.ResourcesPluralConfig:
		return o.ksClient.NotificationV2beta2().Configs().Delete(context.Background(), name, v1.DeleteOptions{})
	case v2beta2.ResourcesPluralReceiver:
		return o.ksClient.NotificationV2beta2().Receivers().Delete(context.Background(), name, v1.DeleteOptions{})
	case v2beta2.ResourcesPluralRouter:
		return o.ksClient.NotificationV2beta2().Routers().Delete(context.Background(), name, v1.DeleteOptions{})
	case v2beta2.ResourcesPluralSilence:
		return o.ksClient.NotificationV2beta2().Silences().Delete(context.Background(), name, v1.DeleteOptions{})
	case Secret:
		return o.k8sClient.CoreV1().Secrets(constants.NotificationSecretNamespace).Delete(context.Background(), name, v1.DeleteOptions{})
	case ConfigMap:
		return o.k8sClient.CoreV1().ConfigMaps(constants.NotificationSecretNamespace).Delete(context.Background(), name, v1.DeleteOptions{})
	default:
		return errors.NewInternalError(nil)
	}
}

// UpdateV2beta1 an object of version v2beta1, only a global object will be updated if the user is nil.
// If the user is not nil, a tenant object whose tenant label matches the user will be updated.
func (o *operator) UpdateV2beta1(user, resource, name string, obj runtime.Object) (runtime.Object, error) {
	return o.update(user, resource, V2beta1, name, obj)
}

// Update an object, only a global object will be updated if the user is nil.
// If the user is not nil, a tenant object whose tenant label matches the user will be updated.
func (o *operator) Update(user, resource, name string, obj runtime.Object) (runtime.Object, error) {
	return o.update(user, resource, V2beta2, name, obj)
}

func (o *operator) update(user, resource, version, name string, obj runtime.Object) (runtime.Object, error) {

	if user != "" {
		if resource == v2beta2.ResourcesPluralRouter ||
			resource == v2beta2.ResourcesPluralNotificationManager {
			return nil, errors.NewForbidden(v2beta2.Resource(resource), "",
				fmt.Errorf("tenant can not update %s", resource))
		}
	}

	accessor, err := meta.Accessor(obj)
	if err != nil {
		return nil, err
	}

	if accessor.GetName() != name {
		return nil, fmt.Errorf("incorrcet parameter, resource name is not equal to the name in body")
	}

	_, err = o.Get(user, resource, name, "")
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if err := appendLabel(user, resource, obj); err != nil {
		return nil, err
	}

	switch resource {
	case v2beta2.ResourcesPluralNotificationManager:
		return o.ksClient.NotificationV2beta2().NotificationManagers().Update(context.Background(), obj.(*v2beta2.NotificationManager), v1.UpdateOptions{})
	case v2beta2.ResourcesPluralConfig:
		if version == V2beta1 {
			return o.ksClient.NotificationV2beta1().Configs().Update(context.Background(), obj.(*v2beta1.Config), v1.UpdateOptions{})
		} else {
			return o.ksClient.NotificationV2beta2().Configs().Update(context.Background(), obj.(*v2beta2.Config), v1.UpdateOptions{})
		}
	case v2beta2.ResourcesPluralReceiver:
		if version == V2beta1 {
			return o.ksClient.NotificationV2beta1().Receivers().Update(context.Background(), obj.(*v2beta1.Receiver), v1.UpdateOptions{})
		} else {
			return o.ksClient.NotificationV2beta2().Receivers().Update(context.Background(), obj.(*v2beta2.Receiver), v1.UpdateOptions{})
		}
	case v2beta2.ResourcesPluralRouter:
		return o.ksClient.NotificationV2beta2().Routers().Update(context.Background(), obj.(*v2beta2.Router), v1.UpdateOptions{})
	case v2beta2.ResourcesPluralSilence:
		return o.ksClient.NotificationV2beta2().Silences().Update(context.Background(), obj.(*v2beta2.Silence), v1.UpdateOptions{})
	case Secret:
		return o.k8sClient.CoreV1().Secrets(constants.NotificationSecretNamespace).Update(context.Background(), obj.(*corev1.Secret), v1.UpdateOptions{})
	case ConfigMap:
		return o.k8sClient.CoreV1().ConfigMaps(constants.NotificationSecretNamespace).Update(context.Background(), obj.(*corev1.ConfigMap), v1.UpdateOptions{})
	default:
		return nil, errors.NewInternalError(nil)
	}
}

// Patch an object, only a global object will be patched if the user is nil.
// If the user is not nil, a tenant object whose tenant label matches the user will be patched.
func (o *operator) Patch(user, resource, name string, data []byte) (runtime.Object, error) {

	if user != "" {
		if resource == v2beta2.ResourcesPluralRouter ||
			resource == v2beta2.ResourcesPluralNotificationManager {
			return nil, errors.NewForbidden(v2beta2.Resource(resource), "",
				fmt.Errorf("tenant can not update %s", resource))
		}
	}

	_, err := o.Get(user, resource, name, "")
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	switch resource {
	case v2beta2.ResourcesPluralNotificationManager:
		return o.ksClient.NotificationV2beta2().NotificationManagers().Patch(context.Background(), name, types.MergePatchType, data, v1.PatchOptions{})
	case v2beta2.ResourcesPluralConfig:
		return o.ksClient.NotificationV2beta2().Configs().Patch(context.Background(), name, types.MergePatchType, data, v1.PatchOptions{})
	case v2beta2.ResourcesPluralReceiver:
		return o.ksClient.NotificationV2beta2().Receivers().Patch(context.Background(), name, types.MergePatchType, data, v1.PatchOptions{})
	case v2beta2.ResourcesPluralRouter:
		return o.ksClient.NotificationV2beta2().Routers().Patch(context.Background(), name, types.MergePatchType, data, v1.PatchOptions{})
	case v2beta2.ResourcesPluralSilence:
		return o.ksClient.NotificationV2beta2().Silences().Patch(context.Background(), name, types.MergePatchType, data, v1.PatchOptions{})
	case Secret:
		return o.k8sClient.CoreV1().Secrets(constants.NotificationSecretNamespace).Patch(context.Background(), name, types.MergePatchType, data, v1.PatchOptions{})
	case ConfigMap:
		return o.k8sClient.CoreV1().ConfigMaps(constants.NotificationSecretNamespace).Patch(context.Background(), name, types.MergePatchType, data, v1.PatchOptions{})
	default:
		return nil, errors.NewInternalError(nil)
	}
}

func (o *operator) GetObject(resource, version string) runtime.Object {

	switch resource {
	case v2beta2.ResourcesPluralNotificationManager:
		return &v2beta2.NotificationManager{}
	case v2beta2.ResourcesPluralConfig:
		if version == V2beta1 {
			return &v2beta1.Config{}
		} else {
			return &v2beta2.Config{}
		}
	case v2beta2.ResourcesPluralReceiver:
		if version == V2beta1 {
			return &v2beta1.Receiver{}
		} else {
			return &v2beta2.Receiver{}
		}
	case v2beta2.ResourcesPluralRouter:
		if version == V2beta1 {
			return nil
		}
		return &v2beta2.Router{}
	case v2beta2.ResourcesPluralSilence:
		if version == V2beta1 {
			return nil
		}
		return &v2beta2.Silence{}
	case Secret:
		return &corev1.Secret{}
	case ConfigMap:
		return &corev1.ConfigMap{}
	default:
		return nil
	}
}

func (o *operator) IsKnownResource(resource, version, subresource string) bool {

	if obj := o.GetObject(resource, version); obj == nil {
		return false
	}

	res := false
	// "" means get all types of the config or receiver.
	if subresource == "dingtalk" ||
		subresource == "email" ||
		subresource == "slack" ||
		subresource == "webhook" ||
		subresource == "wechat" ||
		subresource == "" {
		res = true
	}

	if version == V2beta2 && subresource == "feishu" {
		res = true
	}

	return res
}

func (o *operator) Verify(request *restful.Request, response *restful.Response) {
	if o.options == nil || len(o.options.Endpoint) == 0 {
		_ = response.WriteAsJson(Result{
			http.StatusInternalServerError,
			"Cannot find Notification Manager endpoint",
		})
		return
	}

	reqBody, err := ioutil.ReadAll(request.Request.Body)
	if err != nil {
		klog.Error(err)
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, err)
		return
	}

	data := Data{}
	err = json.Unmarshal(reqBody, &data)
	if err != nil {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, err)
		return
	}

	receiver := data.Receiver
	user := request.PathParameter("user")

	if err := authorizer(user, &receiver); err != nil {
		klog.Error(err)
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, err)
		return
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", o.options.Endpoint, VerificationAPIPath), bytes.NewReader(reqBody))
	if err != nil {
		klog.Error(err)
		_ = response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}
	req.Header = request.Request.Header

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		klog.Error(err)
		_ = response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		klog.Error(err)
		// return 500
		_ = response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		return
	}

	response.AddHeader(restful.HEADER_ContentType, restful.MIME_JSON)
	response.WriteHeader(http.StatusOK)
	_, _ = response.Write(body)
}

// Does the user has permission to access this object.
func authorizer(user string, obj runtime.Object) error {
	// If the user is not nil, it must equal to the tenant specified in labels of the object.
	if user != "" && !isOwner(user, obj) {
		return errors.NewForbidden(v2beta2.Resource(obj.GetObjectKind().GroupVersionKind().GroupKind().Kind), "",
			fmt.Errorf("user '%s' is not the owner of object", user))
	}

	// If the user is nil, the object must be a global object.
	if user == "" && !isGlobal(obj) {
		return errors.NewForbidden(v2beta2.Resource(obj.GetObjectKind().GroupVersionKind().GroupKind().Kind), "",
			fmt.Errorf("object is not a global object"))
	}

	return nil
}

// Is the user equal to the tenant specified in the object labels.
func isOwner(user string, obj interface{}) bool {

	accessor, err := meta.Accessor(obj)
	if err != nil {
		klog.Errorln(err)
		return false
	}

	return accessor.GetLabels()["user"] == user && accessor.GetLabels()["type"] == "tenant"
}

func isConfig(obj runtime.Object) bool {
	switch obj.(type) {
	case *v2beta1.Config, *v2beta2.Config:
		return true
	default:
		return false
	}
}

// Is the object is a global object.
func isGlobal(obj runtime.Object) bool {

	if _, ok := obj.(*v2beta2.NotificationManager); ok {
		return true
	}

	accessor, err := meta.Accessor(obj)
	if err != nil {
		klog.Errorln(err)
		return false
	}

	if isConfig(obj) {
		return accessor.GetLabels()["type"] == "default"
	} else {
		return accessor.GetLabels()["type"] == "global"
	}
}

func appendLabel(user, resource string, obj runtime.Object) error {

	accessor, err := meta.Accessor(obj)
	if err != nil {
		klog.Errorln(err)
		return err
	}

	labels := accessor.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	if resource == Secret || resource == ConfigMap {
		labels[constants.NotificationManagedLabel] = "true"
	}

	if user == "" {
		if isConfig(obj) {
			labels["type"] = "default"
		} else {
			labels["type"] = "global"
		}
	} else {
		labels["type"] = "tenant"
		labels["user"] = user
	}

	accessor.SetLabels(labels)
	return nil
}

func clean(obj interface{}, resource, subresource string) runtime.Object {
	if resource == v2beta2.ResourcesPluralConfig {
		config := obj.(*v2beta2.Config)
		newConfig := config.DeepCopy()
		newConfig.Spec = v2beta2.ConfigSpec{}
		switch subresource {
		case "dingtalk":
			newConfig.Spec.DingTalk = config.Spec.DingTalk
		case "email":
			newConfig.Spec.Email = config.Spec.Email
		case "feishu":
			newConfig.Spec.Feishu = config.Spec.Feishu
		case "slack":
			newConfig.Spec.Slack = config.Spec.Slack
		case "webhook":
			newConfig.Spec.Webhook = config.Spec.Webhook
		case "wechat":
			newConfig.Spec.Wechat = config.Spec.Wechat
		default:
			return nil
		}

		if reflect.ValueOf(newConfig.Spec).IsZero() {
			return nil
		}

		return newConfig
	} else if resource == v2beta2.ResourcesPluralReceiver {
		receiver := obj.(*v2beta2.Receiver)
		newReceiver := receiver.DeepCopy()
		newReceiver.Spec = v2beta2.ReceiverSpec{}
		switch subresource {
		case "dingtalk":
			newReceiver.Spec.DingTalk = receiver.Spec.DingTalk
		case "email":
			newReceiver.Spec.Email = receiver.Spec.Email
		case "feishu":
			newReceiver.Spec.Feishu = receiver.Spec.Feishu
		case "slack":
			newReceiver.Spec.Slack = receiver.Spec.Slack
		case "webhook":
			newReceiver.Spec.Webhook = receiver.Spec.Webhook
		case "wechat":
			newReceiver.Spec.Wechat = receiver.Spec.Wechat
		default:
			return nil
		}

		if reflect.ValueOf(newReceiver.Spec).IsZero() {
			return nil
		}

		return newReceiver
	} else {
		return obj.(runtime.Object)
	}
}

func convert(obj runtime.Object) runtime.Object {
	switch obj := obj.(type) {
	case *v2beta2.Config:
		dst := &v2beta1.Config{}
		_ = obj.ConvertTo(dst)
		return dst
	case *v2beta2.Receiver:
		dst := &v2beta1.Receiver{}
		_ = obj.ConvertTo(dst)
		return dst
	default:
		return obj
	}
}
