package notification

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apis/notification/v2beta1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	"reflect"
)

type Operator interface {
	List(user, resource, subresource string, query *query.Query) (*api.ListResult, error)
	Get(user, resource, name, subresource string) (runtime.Object, error)
	Create(user, resource string, obj runtime.Object) (runtime.Object, error)
	Delete(user, resource, name string) error
	Update(user, resource, name string, obj runtime.Object) (runtime.Object, error)

	GetObject(resource string) runtime.Object
	IsKnownResource(resource, subresource string) bool
}

type operator struct {
	k8sClient      kubernetes.Interface
	ksClient       kubesphere.Interface
	informers      informers.InformerFactory
	resourceGetter *resource.ResourceGetter
}

func NewOperator(
	informers informers.InformerFactory,
	k8sClient kubernetes.Interface,
	ksClient kubesphere.Interface) Operator {

	return &operator{
		informers:      informers,
		k8sClient:      k8sClient,
		ksClient:       ksClient,
		resourceGetter: resource.NewResourceGetter(informers, nil),
	}
}

// List objects. Only global objects will be returned if the user is nil.
// If the user is not nil, only tenant objects whose tenant label matches the user will be returned.
func (o *operator) List(user, resource, subresource string, q *query.Query) (*api.ListResult, error) {

	if len(q.LabelSelector) > 0 {
		q.LabelSelector = q.LabelSelector + ","
	}

	filter := ""
	// If user is nil, it will list all global object.
	if user == "" {
		if isConfig(o.GetObject(resource)) {
			filter = "type=default"
		} else {
			filter = "type=global"
		}
	} else {
		// If the user is not nil, only return the object belong to this user.
		filter = "type=tenant,user=" + user
	}

	q.LabelSelector = q.LabelSelector + filter

	res, err := o.resourceGetter.List(resource, constants.NotificationSecretNamespace, q)
	if err != nil {
		return nil, err
	}

	if subresource == "" || resource == "secrets" {
		return res, nil
	}

	results := &api.ListResult{}
	for _, item := range res.Items {
		obj := clean(item, resource, subresource)
		if obj != nil {
			results.Items = append(results.Items, obj)
		}
	}
	results.TotalItems = len(results.Items)

	return results, nil
}

// Get the specified object, if you want to get a global object, the user must be nil.
// If you want to get a tenant object, the user must equal to the tenant specified in labels of the object.
func (o *operator) Get(user, resource, name, subresource string) (runtime.Object, error) {
	obj, err := o.resourceGetter.Get(resource, constants.NotificationSecretNamespace, name)
	if err != nil {
		return nil, err
	}

	if err := authorizer(user, obj); err != nil {
		return nil, err
	}

	if subresource == "" || resource == "secrets" {
		return obj, nil
	}

	res := clean(obj, resource, subresource)
	if res == nil {
		return nil, errors.NewNotFound(v2beta1.Resource(obj.GetObjectKind().GroupVersionKind().GroupKind().Kind), name)
	}

	return res, nil
}

// Create an object. A global object will be created if the user is nil.
// A tenant object will be created if the user is not nil.
func (o *operator) Create(user, resource string, obj runtime.Object) (runtime.Object, error) {

	if err := appendLabel(user, obj); err != nil {
		return nil, err
	}

	switch resource {
	case v2beta1.ResourcesPluralConfig:
		return o.ksClient.NotificationV2beta1().Configs().Create(context.Background(), obj.(*v2beta1.Config), v1.CreateOptions{})
	case v2beta1.ResourcesPluralReceiver:
		return o.ksClient.NotificationV2beta1().Receivers().Create(context.Background(), obj.(*v2beta1.Receiver), v1.CreateOptions{})
	case "secrets":
		return o.k8sClient.CoreV1().Secrets(constants.NotificationSecretNamespace).Create(context.Background(), obj.(*corev1.Secret), v1.CreateOptions{})
	default:
		return nil, errors.NewInternalError(nil)
	}
}

// Delete an object. A global object will be deleted if the user is nil.
// If the user is not nil, a tenant object whose tenant label matches the user will be deleted.
func (o *operator) Delete(user, resource, name string) error {

	if obj, err := o.Get(user, resource, name, ""); err != nil {
		return err
	} else {
		if err := authorizer(user, obj); err != nil {
			return err
		}
	}

	switch resource {
	case v2beta1.ResourcesPluralConfig:
		return o.ksClient.NotificationV2beta1().Configs().Delete(context.Background(), name, v1.DeleteOptions{})
	case v2beta1.ResourcesPluralReceiver:
		return o.ksClient.NotificationV2beta1().Receivers().Delete(context.Background(), name, v1.DeleteOptions{})
	case "secrets":
		return o.k8sClient.CoreV1().Secrets(constants.NotificationSecretNamespace).Delete(context.Background(), name, v1.DeleteOptions{})
	default:
		return errors.NewInternalError(nil)
	}
}

// Update an object, only a global object will be updated if the user is nil.
// If the user is not nil, a tenant object whose tenant label matches the user will be updated.
func (o *operator) Update(user, resource, name string, obj runtime.Object) (runtime.Object, error) {

	if err := appendLabel(user, obj); err != nil {
		return nil, err
	}

	if old, err := o.Get(user, resource, name, ""); err != nil {
		return nil, err
	} else {
		if err := authorizer(user, old); err != nil {
			return nil, err
		}
	}

	switch resource {
	case v2beta1.ResourcesPluralConfig:
		return o.ksClient.NotificationV2beta1().Configs().Update(context.Background(), obj.(*v2beta1.Config), v1.UpdateOptions{})
	case v2beta1.ResourcesPluralReceiver:
		return o.ksClient.NotificationV2beta1().Receivers().Update(context.Background(), obj.(*v2beta1.Receiver), v1.UpdateOptions{})
	case "secrets":
		return o.k8sClient.CoreV1().Secrets(constants.NotificationSecretNamespace).Update(context.Background(), obj.(*corev1.Secret), v1.UpdateOptions{})
	default:
		return nil, errors.NewInternalError(nil)
	}
}

func (o *operator) GetObject(resource string) runtime.Object {

	switch resource {
	case v2beta1.ResourcesPluralConfig:
		return &v2beta1.Config{}
	case v2beta1.ResourcesPluralReceiver:
		return &v2beta1.Receiver{}
	case "secrets":
		return &corev1.Secret{}
	default:
		return nil
	}
}

func (o *operator) IsKnownResource(resource, subresource string) bool {

	if obj := o.GetObject(resource); obj == nil {
		return false
	}

	// "" means get all types of the config or receiver.
	if subresource != "dingtalk" &&
		subresource != "email" &&
		subresource != "slack" &&
		subresource != "webhook" &&
		subresource != "wechat" &&
		subresource != "" {
		return false
	}

	return true
}

// Does the user has permission to access this object.
func authorizer(user string, obj runtime.Object) error {
	// If the user is not nil, it must equal to the tenant specified in labels of the object.
	if user != "" && !isOwner(user, obj) {
		return errors.NewForbidden(v2beta1.Resource(obj.GetObjectKind().GroupVersionKind().GroupKind().Kind), "",
			fmt.Errorf("user '%s' is not the owner of object", user))
	}

	// If the user is nil, the object must be a global object.
	if user == "" && !isGlobal(obj) {
		return errors.NewForbidden(v2beta1.Resource(obj.GetObjectKind().GroupVersionKind().GroupKind().Kind), "",
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

	return accessor.GetLabels()["user"] == user
}

func isConfig(obj runtime.Object) bool {
	switch obj.(type) {
	case *v2beta1.Config:
		return true
	default:
		return false
	}
}

// Is the object is a global object.
func isGlobal(obj runtime.Object) bool {
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

func appendLabel(user string, obj runtime.Object) error {

	accessor, err := meta.Accessor(obj)
	if err != nil {
		klog.Errorln(err)
		return err
	}

	labels := accessor.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
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
	if resource == v2beta1.ResourcesPluralConfig {
		config := obj.(*v2beta1.Config)
		newConfig := config.DeepCopy()
		newConfig.Spec = v2beta1.ConfigSpec{}
		switch subresource {
		case "dingtalk":
			newConfig.Spec.DingTalk = config.Spec.DingTalk
		case "email":
			newConfig.Spec.Email = config.Spec.Email
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
	} else {
		receiver := obj.(*v2beta1.Receiver)
		newReceiver := receiver.DeepCopy()
		newReceiver.Spec = v2beta1.ReceiverSpec{}
		switch subresource {
		case "dingtalk":
			newReceiver.Spec.DingTalk = receiver.Spec.DingTalk
		case "email":
			newReceiver.Spec.Email = receiver.Spec.Email
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
	}
}
