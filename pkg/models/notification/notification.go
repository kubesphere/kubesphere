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
	v2alpha1 "kubesphere.io/kubesphere/pkg/apis/notification/v2alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
)

type Operator interface {
	List(user, resource string, query *query.Query) (*api.ListResult, error)
	Get(user, resource, name string) (runtime.Object, error)
	Create(user, resource string, obj runtime.Object) (runtime.Object, error)
	Delete(user, resource, name string) error
	Update(user, resource, name string, obj runtime.Object) (runtime.Object, error)

	GetObject(resource string) runtime.Object
	IsKnownResource(resource string) bool
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
func (o *operator) List(user, resource string, q *query.Query) (*api.ListResult, error) {

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

	return o.resourceGetter.List(resource, constants.NotificationSecretNamespace, q)
}

// Get the specified object, if you want to get a global object, the user must be nil.
// If you want to get a tenant object, the user must equal to the tenant specified in labels of the object.
func (o *operator) Get(user, resource, name string) (runtime.Object, error) {
	obj, err := o.resourceGetter.Get(resource, constants.NotificationSecretNamespace, name)
	if err != nil {
		return nil, err
	}

	if err := authorizer(user, obj); err != nil {
		return nil, err
	}

	return obj, nil
}

// Create an object. A global object will be created if the user is nil.
// A tenant object will be created if the user is not nil.
func (o *operator) Create(user, resource string, obj runtime.Object) (runtime.Object, error) {

	if err := appendLabel(user, obj); err != nil {
		return nil, err
	}

	switch resource {
	case v2alpha1.ResourcesPluralDingTalkConfig:
		return o.ksClient.NotificationV2alpha1().DingTalkConfigs().Create(context.Background(), obj.(*v2alpha1.DingTalkConfig), v1.CreateOptions{})
	case v2alpha1.ResourcesPluralDingTalkReceiver:
		return o.ksClient.NotificationV2alpha1().DingTalkReceivers().Create(context.Background(), obj.(*v2alpha1.DingTalkReceiver), v1.CreateOptions{})
	case v2alpha1.ResourcesPluralEmailConfig:
		return o.ksClient.NotificationV2alpha1().EmailConfigs().Create(context.Background(), obj.(*v2alpha1.EmailConfig), v1.CreateOptions{})
	case v2alpha1.ResourcesPluralEmailReceiver:
		return o.ksClient.NotificationV2alpha1().EmailReceivers().Create(context.Background(), obj.(*v2alpha1.EmailReceiver), v1.CreateOptions{})
	case v2alpha1.ResourcesPluralSlackConfig:
		return o.ksClient.NotificationV2alpha1().SlackConfigs().Create(context.Background(), obj.(*v2alpha1.SlackConfig), v1.CreateOptions{})
	case v2alpha1.ResourcesPluralSlackReceiver:
		return o.ksClient.NotificationV2alpha1().SlackReceivers().Create(context.Background(), obj.(*v2alpha1.SlackReceiver), v1.CreateOptions{})
	case v2alpha1.ResourcesPluralWebhookConfig:
		return o.ksClient.NotificationV2alpha1().WebhookConfigs().Create(context.Background(), obj.(*v2alpha1.WebhookConfig), v1.CreateOptions{})
	case v2alpha1.ResourcesPluralWebhookReceiver:
		return o.ksClient.NotificationV2alpha1().WebhookReceivers().Create(context.Background(), obj.(*v2alpha1.WebhookReceiver), v1.CreateOptions{})
	case v2alpha1.ResourcesPluralWechatConfig:
		return o.ksClient.NotificationV2alpha1().WechatConfigs().Create(context.Background(), obj.(*v2alpha1.WechatConfig), v1.CreateOptions{})
	case v2alpha1.ResourcesPluralWechatReceiver:
		return o.ksClient.NotificationV2alpha1().WechatReceivers().Create(context.Background(), obj.(*v2alpha1.WechatReceiver), v1.CreateOptions{})
	case "secrets":
		return o.k8sClient.CoreV1().Secrets(constants.NotificationSecretNamespace).Create(context.Background(), obj.(*corev1.Secret), v1.CreateOptions{})
	default:
		return nil, errors.NewInternalError(nil)
	}
}

// Delete an object. A global object will be deleted if the user is nil.
// If the user is not nil, a tenant object whose tenant label matches the user will be deleted.
func (o *operator) Delete(user, resource, name string) error {

	if obj, err := o.Get(user, resource, name); err != nil {
		return err
	} else {
		if err := authorizer(user, obj); err != nil {
			return err
		}
	}

	switch resource {
	case v2alpha1.ResourcesPluralDingTalkConfig:
		return o.ksClient.NotificationV2alpha1().DingTalkConfigs().Delete(context.Background(), name, v1.DeleteOptions{})
	case v2alpha1.ResourcesPluralDingTalkReceiver:
		return o.ksClient.NotificationV2alpha1().DingTalkReceivers().Delete(context.Background(), name, v1.DeleteOptions{})
	case v2alpha1.ResourcesPluralEmailConfig:
		return o.ksClient.NotificationV2alpha1().EmailConfigs().Delete(context.Background(), name, v1.DeleteOptions{})
	case v2alpha1.ResourcesPluralEmailReceiver:
		return o.ksClient.NotificationV2alpha1().EmailReceivers().Delete(context.Background(), name, v1.DeleteOptions{})
	case v2alpha1.ResourcesPluralSlackConfig:
		return o.ksClient.NotificationV2alpha1().SlackConfigs().Delete(context.Background(), name, v1.DeleteOptions{})
	case v2alpha1.ResourcesPluralSlackReceiver:
		return o.ksClient.NotificationV2alpha1().SlackReceivers().Delete(context.Background(), name, v1.DeleteOptions{})
	case v2alpha1.ResourcesPluralWebhookConfig:
		return o.ksClient.NotificationV2alpha1().WebhookConfigs().Delete(context.Background(), name, v1.DeleteOptions{})
	case v2alpha1.ResourcesPluralWebhookReceiver:
		return o.ksClient.NotificationV2alpha1().WebhookReceivers().Delete(context.Background(), name, v1.DeleteOptions{})
	case v2alpha1.ResourcesPluralWechatConfig:
		return o.ksClient.NotificationV2alpha1().WechatConfigs().Delete(context.Background(), name, v1.DeleteOptions{})
	case v2alpha1.ResourcesPluralWechatReceiver:
		return o.ksClient.NotificationV2alpha1().WechatReceivers().Delete(context.Background(), name, v1.DeleteOptions{})
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

	if old, err := o.Get(user, resource, name); err != nil {
		return nil, err
	} else {
		if err := authorizer(user, old); err != nil {
			return nil, err
		}
	}

	switch resource {
	case v2alpha1.ResourcesPluralDingTalkConfig:
		return o.ksClient.NotificationV2alpha1().DingTalkConfigs().Update(context.Background(), obj.(*v2alpha1.DingTalkConfig), v1.UpdateOptions{})
	case v2alpha1.ResourcesPluralDingTalkReceiver:
		return o.ksClient.NotificationV2alpha1().DingTalkReceivers().Update(context.Background(), obj.(*v2alpha1.DingTalkReceiver), v1.UpdateOptions{})
	case v2alpha1.ResourcesPluralEmailConfig:
		return o.ksClient.NotificationV2alpha1().EmailConfigs().Update(context.Background(), obj.(*v2alpha1.EmailConfig), v1.UpdateOptions{})
	case v2alpha1.ResourcesPluralEmailReceiver:
		return o.ksClient.NotificationV2alpha1().EmailReceivers().Update(context.Background(), obj.(*v2alpha1.EmailReceiver), v1.UpdateOptions{})
	case v2alpha1.ResourcesPluralSlackConfig:
		return o.ksClient.NotificationV2alpha1().SlackConfigs().Update(context.Background(), obj.(*v2alpha1.SlackConfig), v1.UpdateOptions{})
	case v2alpha1.ResourcesPluralSlackReceiver:
		return o.ksClient.NotificationV2alpha1().SlackReceivers().Update(context.Background(), obj.(*v2alpha1.SlackReceiver), v1.UpdateOptions{})
	case v2alpha1.ResourcesPluralWebhookConfig:
		return o.ksClient.NotificationV2alpha1().WebhookConfigs().Update(context.Background(), obj.(*v2alpha1.WebhookConfig), v1.UpdateOptions{})
	case v2alpha1.ResourcesPluralWebhookReceiver:
		return o.ksClient.NotificationV2alpha1().WebhookReceivers().Update(context.Background(), obj.(*v2alpha1.WebhookReceiver), v1.UpdateOptions{})
	case v2alpha1.ResourcesPluralWechatConfig:
		return o.ksClient.NotificationV2alpha1().WechatConfigs().Update(context.Background(), obj.(*v2alpha1.WechatConfig), v1.UpdateOptions{})
	case v2alpha1.ResourcesPluralWechatReceiver:
		return o.ksClient.NotificationV2alpha1().WechatReceivers().Update(context.Background(), obj.(*v2alpha1.WechatReceiver), v1.UpdateOptions{})
	case "secrets":
		return o.k8sClient.CoreV1().Secrets(constants.NotificationSecretNamespace).Update(context.Background(), obj.(*corev1.Secret), v1.UpdateOptions{})
	default:
		return nil, errors.NewInternalError(nil)
	}
}

func (o *operator) GetObject(resource string) runtime.Object {

	switch resource {
	case v2alpha1.ResourcesPluralDingTalkConfig:
		return &v2alpha1.DingTalkConfig{}
	case v2alpha1.ResourcesPluralDingTalkReceiver:
		return &v2alpha1.DingTalkReceiver{}
	case v2alpha1.ResourcesPluralEmailConfig:
		return &v2alpha1.EmailConfig{}
	case v2alpha1.ResourcesPluralEmailReceiver:
		return &v2alpha1.EmailReceiver{}
	case v2alpha1.ResourcesPluralSlackConfig:
		return &v2alpha1.SlackConfig{}
	case v2alpha1.ResourcesPluralSlackReceiver:
		return &v2alpha1.SlackReceiver{}
	case v2alpha1.ResourcesPluralWebhookConfig:
		return &v2alpha1.WebhookConfig{}
	case v2alpha1.ResourcesPluralWebhookReceiver:
		return &v2alpha1.WebhookReceiver{}
	case v2alpha1.ResourcesPluralWechatConfig:
		return &v2alpha1.WechatConfig{}
	case v2alpha1.ResourcesPluralWechatReceiver:
		return &v2alpha1.WechatReceiver{}
	case "secrets":
		return &corev1.Secret{}
	default:
		return nil
	}
}

func (o *operator) IsKnownResource(resource string) bool {

	if obj := o.GetObject(resource); obj == nil {
		return false
	}

	return true
}

// Does the user has permission to access this object.
func authorizer(user string, obj runtime.Object) error {
	// If the user is not nil, it must equal to the tenant specified in labels of the object.
	if user != "" && !isOwner(user, obj) {
		return errors.NewForbidden(v2alpha1.Resource(obj.GetObjectKind().GroupVersionKind().GroupKind().Kind), "",
			fmt.Errorf("user '%s' is not the owner of object", user))
	}

	// If the user is nil, the object must be a global object.
	if user == "" && !isGlobal(obj) {
		return errors.NewForbidden(v2alpha1.Resource(obj.GetObjectKind().GroupVersionKind().GroupKind().Kind), "",
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
	case *v2alpha1.DingTalkConfig, *v2alpha1.EmailConfig, *v2alpha1.SlackConfig, *v2alpha1.WebhookConfig, *v2alpha1.WechatConfig:
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
