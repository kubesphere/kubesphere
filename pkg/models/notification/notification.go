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
	v2 "kubesphere.io/kubesphere/pkg/apis/notification/v2"
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
	Update(user, resource string, obj runtime.Object) (runtime.Object, error)

	ListSecret(query *query.Query) (*api.ListResult, error)
	GetSecret(name string) (interface{}, error)
	CreateOrUpdateSecret(obj *corev1.Secret) (*corev1.Secret, error)
	DeleteSecret(name string) error

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

	// If user is nil, it will list all global object.
	if user == "" {
		appendGlobalLabel(resource, q)
	} else {
		// If the user is not nil, only return the object belong to this user.
		appendTenantLabel(user, q)
	}

	return o.resourceGetter.List(resource, "", q)
}

// Get the specified object, if you want to get a global object, the user must be nil.
// If you want to get a tenant object, the user must equal to the tenant specified in labels of the object.
func (o *operator) Get(user, resource, name string) (runtime.Object, error) {
	obj, err := o.resourceGetter.Get(resource, "", name)
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

	if err := authorizer(user, obj); err != nil {
		return nil, err
	}

	switch resource {
	case v2.ResourcesPluralDingTalkConfig:
		return o.ksClient.NotificationV2().DingTalkConfigs().Create(context.Background(), obj.(*v2.DingTalkConfig), v1.CreateOptions{})
	case v2.ResourcesPluralDingTalkReceiver:
		return o.ksClient.NotificationV2().DingTalkReceivers().Create(context.Background(), obj.(*v2.DingTalkReceiver), v1.CreateOptions{})
	case v2.ResourcesPluralEmailConfig:
		return o.ksClient.NotificationV2().EmailConfigs().Create(context.Background(), obj.(*v2.EmailConfig), v1.CreateOptions{})
	case v2.ResourcesPluralEmailReceiver:
		return o.ksClient.NotificationV2().EmailReceivers().Create(context.Background(), obj.(*v2.EmailReceiver), v1.CreateOptions{})
	case v2.ResourcesPluralSlackConfig:
		return o.ksClient.NotificationV2().SlackConfigs().Create(context.Background(), obj.(*v2.SlackConfig), v1.CreateOptions{})
	case v2.ResourcesPluralSlackReceiver:
		return o.ksClient.NotificationV2().SlackReceivers().Create(context.Background(), obj.(*v2.SlackReceiver), v1.CreateOptions{})
	case v2.ResourcesPluralWebhookConfig:
		return o.ksClient.NotificationV2().WebhookConfigs().Create(context.Background(), obj.(*v2.WebhookConfig), v1.CreateOptions{})
	case v2.ResourcesPluralWebhookReceiver:
		return o.ksClient.NotificationV2().WebhookReceivers().Create(context.Background(), obj.(*v2.WebhookReceiver), v1.CreateOptions{})
	case v2.ResourcesPluralWechatConfig:
		return o.ksClient.NotificationV2().WechatConfigs().Create(context.Background(), obj.(*v2.WechatConfig), v1.CreateOptions{})
	case v2.ResourcesPluralWechatReceiver:
		return o.ksClient.NotificationV2().WechatReceivers().Create(context.Background(), obj.(*v2.WechatReceiver), v1.CreateOptions{})
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
	case v2.ResourcesPluralDingTalkConfig:
		return o.ksClient.NotificationV2().DingTalkConfigs().Delete(context.Background(), name, v1.DeleteOptions{})
	case v2.ResourcesPluralDingTalkReceiver:
		return o.ksClient.NotificationV2().DingTalkReceivers().Delete(context.Background(), name, v1.DeleteOptions{})
	case v2.ResourcesPluralEmailConfig:
		return o.ksClient.NotificationV2().EmailConfigs().Delete(context.Background(), name, v1.DeleteOptions{})
	case v2.ResourcesPluralEmailReceiver:
		return o.ksClient.NotificationV2().EmailReceivers().Delete(context.Background(), name, v1.DeleteOptions{})
	case v2.ResourcesPluralSlackConfig:
		return o.ksClient.NotificationV2().SlackConfigs().Delete(context.Background(), name, v1.DeleteOptions{})
	case v2.ResourcesPluralSlackReceiver:
		return o.ksClient.NotificationV2().SlackReceivers().Delete(context.Background(), name, v1.DeleteOptions{})
	case v2.ResourcesPluralWebhookConfig:
		return o.ksClient.NotificationV2().WebhookConfigs().Delete(context.Background(), name, v1.DeleteOptions{})
	case v2.ResourcesPluralWebhookReceiver:
		return o.ksClient.NotificationV2().WebhookReceivers().Delete(context.Background(), name, v1.DeleteOptions{})
	case v2.ResourcesPluralWechatConfig:
		return o.ksClient.NotificationV2().WechatConfigs().Delete(context.Background(), name, v1.DeleteOptions{})
	case v2.ResourcesPluralWechatReceiver:
		return o.ksClient.NotificationV2().WechatReceivers().Delete(context.Background(), name, v1.DeleteOptions{})
	default:
		return errors.NewInternalError(nil)
	}
}

// Update an object, only a global object will be updated if the user is nil.
// If the user is not nil, a tenant object whose tenant label matches the user will be updated.
func (o *operator) Update(user, resource string, obj runtime.Object) (runtime.Object, error) {

	name, err := getName(obj)
	if err != nil {
		return nil, err
	}

	if _, err := o.Get(user, resource, name); err != nil {
		return nil, err
	} else {
		if err := authorizer(user, obj); err != nil {
			return nil, err
		}
	}

	switch resource {
	case v2.ResourcesPluralDingTalkConfig:
		return o.ksClient.NotificationV2().DingTalkConfigs().Update(context.Background(), obj.(*v2.DingTalkConfig), v1.UpdateOptions{})
	case v2.ResourcesPluralDingTalkReceiver:
		return o.ksClient.NotificationV2().DingTalkReceivers().Update(context.Background(), obj.(*v2.DingTalkReceiver), v1.UpdateOptions{})
	case v2.ResourcesPluralEmailConfig:
		return o.ksClient.NotificationV2().EmailConfigs().Update(context.Background(), obj.(*v2.EmailConfig), v1.UpdateOptions{})
	case v2.ResourcesPluralEmailReceiver:
		return o.ksClient.NotificationV2().EmailReceivers().Update(context.Background(), obj.(*v2.EmailReceiver), v1.UpdateOptions{})
	case v2.ResourcesPluralSlackConfig:
		return o.ksClient.NotificationV2().SlackConfigs().Update(context.Background(), obj.(*v2.SlackConfig), v1.UpdateOptions{})
	case v2.ResourcesPluralSlackReceiver:
		return o.ksClient.NotificationV2().SlackReceivers().Update(context.Background(), obj.(*v2.SlackReceiver), v1.UpdateOptions{})
	case v2.ResourcesPluralWebhookConfig:
		return o.ksClient.NotificationV2().WebhookConfigs().Update(context.Background(), obj.(*v2.WebhookConfig), v1.UpdateOptions{})
	case v2.ResourcesPluralWebhookReceiver:
		return o.ksClient.NotificationV2().WebhookReceivers().Update(context.Background(), obj.(*v2.WebhookReceiver), v1.UpdateOptions{})
	case v2.ResourcesPluralWechatConfig:
		return o.ksClient.NotificationV2().WechatConfigs().Update(context.Background(), obj.(*v2.WechatConfig), v1.UpdateOptions{})
	case v2.ResourcesPluralWechatReceiver:
		return o.ksClient.NotificationV2().WechatReceivers().Update(context.Background(), obj.(*v2.WechatReceiver), v1.UpdateOptions{})
	default:
		return nil, errors.NewInternalError(nil)
	}
}

func (o *operator) ListSecret(q *query.Query) (*api.ListResult, error) {

	appendManagedLabel(q)
	return o.resourceGetter.List("secrets", constants.NotificationSecretNamespace, q)
}

func (o *operator) GetSecret(name string) (interface{}, error) {
	obj, err := o.resourceGetter.Get("secrets", constants.NotificationSecretNamespace, name)
	if err != nil {
		return nil, err
	}

	if !isManagedByNotification(obj.(*corev1.Secret)) {
		return nil, errors.NewForbidden(v2.Resource(obj.GetObjectKind().GroupVersionKind().GroupKind().Kind), "",
			fmt.Errorf("secret '%s' is not managed by notification", name))
	}

	return obj, nil
}

func (o *operator) CreateOrUpdateSecret(obj *corev1.Secret) (*corev1.Secret, error) {

	obj.Namespace = constants.NotificationSecretNamespace
	if obj.Labels == nil {
		obj.Labels = make(map[string]string)
	}
	obj.Labels[constants.NotificationManagedLabel] = "true"
	if obj.ResourceVersion == "" {
		return o.k8sClient.CoreV1().Secrets(constants.NotificationSecretNamespace).Create(context.Background(), obj, v1.CreateOptions{})
	} else {
		return o.k8sClient.CoreV1().Secrets(constants.NotificationSecretNamespace).Update(context.Background(), obj, v1.UpdateOptions{})
	}
}

func (o *operator) DeleteSecret(name string) error {

	if _, err := o.GetSecret(name); err != nil {
		return err
	}

	return o.k8sClient.CoreV1().Secrets(constants.NotificationSecretNamespace).Delete(context.Background(), name, v1.DeleteOptions{})
}

func (o *operator) GetObject(resource string) runtime.Object {

	switch resource {
	case v2.ResourcesPluralDingTalkConfig:
		return &v2.DingTalkConfig{}
	case v2.ResourcesPluralDingTalkReceiver:
		return &v2.DingTalkReceiver{}
	case v2.ResourcesPluralEmailConfig:
		return &v2.EmailConfig{}
	case v2.ResourcesPluralEmailReceiver:
		return &v2.EmailReceiver{}
	case v2.ResourcesPluralSlackConfig:
		return &v2.SlackConfig{}
	case v2.ResourcesPluralSlackReceiver:
		return &v2.SlackReceiver{}
	case v2.ResourcesPluralWebhookConfig:
		return &v2.WebhookConfig{}
	case v2.ResourcesPluralWebhookReceiver:
		return &v2.WebhookReceiver{}
	case v2.ResourcesPluralWechatConfig:
		return &v2.WechatConfig{}
	case v2.ResourcesPluralWechatReceiver:
		return &v2.WechatReceiver{}
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
		return errors.NewForbidden(v2.Resource(obj.GetObjectKind().GroupVersionKind().GroupKind().Kind), "",
			fmt.Errorf("user '%s' is not the owner of object", user))
	}

	// If the user is nil, the object must be a global object.
	if user == "" && !isGlobal(obj) {
		return errors.NewForbidden(v2.Resource(obj.GetObjectKind().GroupVersionKind().GroupKind().Kind), "",
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
	case *v2.DingTalkConfig, *v2.EmailConfig, *v2.SlackConfig, *v2.WebhookConfig, *v2.WechatConfig:
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

func appendTenantLabel(user string, q *query.Query) {

	if len(q.LabelSelector) > 0 {
		q.LabelSelector = q.LabelSelector + ","
	}
	q.LabelSelector = q.LabelSelector + "type=tenant,user=" + user
}

func appendGlobalLabel(resource string, q *query.Query) {

	if len(q.LabelSelector) > 0 {
		q.LabelSelector = q.LabelSelector + ","
	}

	switch resource {
	case v2.ResourcesPluralDingTalkConfig, v2.ResourcesPluralEmailConfig,
		v2.ResourcesPluralSlackConfig, v2.ResourcesPluralWebhookConfig, v2.ResourcesPluralWechatConfig:
		q.LabelSelector = q.LabelSelector + "type=default"
	case v2.ResourcesPluralDingTalkReceiver, v2.ResourcesPluralEmailReceiver,
		v2.ResourcesPluralSlackReceiver, v2.ResourcesPluralWebhookReceiver, v2.ResourcesPluralWechatReceiver:
		q.LabelSelector = q.LabelSelector + "type=global"
	}
}

func appendManagedLabel(q *query.Query) {

	if len(q.LabelSelector) > 0 {
		q.LabelSelector = q.LabelSelector + ","
	}
	q.LabelSelector = q.LabelSelector + constants.NotificationManagedLabel + "=" + "true"
}

func isManagedByNotification(secret *corev1.Secret) bool {
	return secret.Labels[constants.NotificationManagedLabel] == "true"
}

func getName(obj runtime.Object) (string, error) {

	accessor, err := meta.Accessor(obj)
	if err != nil {
		return "", err
	}

	return accessor.GetName(), nil
}
