/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package quota

import (
	"context"
	"net/http"
	"sort"
	"sync"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	admissionapi "k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"kubesphere.io/kubesphere/kube/pkg/quota/v1"
	"kubesphere.io/kubesphere/kube/pkg/quota/v1/generic"
	"kubesphere.io/kubesphere/kube/pkg/quota/v1/install"
	"kubesphere.io/kubesphere/kube/plugin/pkg/admission/resourcequota"
	resourcequotaapi "kubesphere.io/kubesphere/kube/plugin/pkg/admission/resourcequota/apis/resourcequota"
	"kubesphere.io/kubesphere/pkg/scheme"
)

const (
	numEvaluatorThreads = 10
)

type ResourceQuotaAdmission struct {
	client client.Client

	decoder *webhook.AdmissionDecoder

	lockFactory LockFactory

	// these are used to create the evaluator
	registry quota.Registry

	init      sync.Once
	evaluator resourcequota.Evaluator
}

const webhookName = "resource-quota-webhook"

func (w *Webhook) Name() string {
	return webhookName
}

var _ kscontroller.Controller = &Webhook{}

type Webhook struct {
}

func (w *Webhook) SetupWithManager(mgr *kscontroller.Manager) error {
	resourceQuotaAdmission := &ResourceQuotaAdmission{
		client:      mgr.GetClient(),
		lockFactory: NewDefaultLockFactory(),
		decoder:     admission.NewDecoder(mgr.GetScheme()),
		registry:    generic.NewRegistry(install.NewQuotaConfigurationForAdmission().Evaluators()),
	}
	mgr.GetWebhookServer().Register("/validate-quota-kubesphere-io-v1alpha2", &webhook.Admission{Handler: resourceQuotaAdmission})
	return nil
}

func (r *ResourceQuotaAdmission) Handle(ctx context.Context, req webhook.AdmissionRequest) webhook.AdmissionResponse {
	// ignore all operations that correspond to sub-resource actions
	if len(req.RequestSubResource) != 0 {
		return webhook.Allowed("")
	}
	// ignore cluster level resources
	if len(req.Namespace) == 0 {
		return webhook.Allowed("")
	}

	r.init.Do(func() {
		resourceQuotaAccessor := newQuotaAccessor(r.client)
		r.evaluator = resourcequota.NewQuotaEvaluator(resourceQuotaAccessor, install.DefaultIgnoredResources(), r.registry, r.lockAquisition, &resourcequotaapi.Configuration{}, numEvaluatorThreads, utilwait.NeverStop)
	})

	attributesRecord, err := convertToAdmissionAttributes(req)
	if err != nil {
		klog.Error(err)
		return webhook.Errored(http.StatusBadRequest, err)
	}

	if err := r.evaluator.Evaluate(attributesRecord); err != nil {
		if errors.IsForbidden(err) {
			klog.Info(err)
			return webhook.Denied(err.Error())
		}
		klog.Error(err)
		return webhook.Errored(http.StatusInternalServerError, err)
	}

	return webhook.Allowed("")
}

type ByName []corev1.ResourceQuota

func (v ByName) Len() int           { return len(v) }
func (v ByName) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v ByName) Less(i, j int) bool { return v[i].Name < v[j].Name }

func (r *ResourceQuotaAdmission) lockAquisition(quotas []corev1.ResourceQuota) func() {
	var locks []sync.Locker

	// acquire the locks in alphabetical order because I'm too lazy to think of something clever
	sort.Sort(ByName(quotas))
	for _, quota := range quotas {
		lock := r.lockFactory.GetLock(string(quota.UID))
		lock.Lock()
		locks = append(locks, lock)
	}

	return func() {
		for i := len(locks) - 1; i >= 0; i-- {
			locks[i].Unlock()
		}
	}
}

func convertToAdmissionAttributes(req admission.Request) (admissionapi.Attributes, error) {
	var err error
	var object runtime.Object
	if len(req.Object.Raw) > 0 {
		object, _, err = scheme.Codecs.UniversalDeserializer().Decode(req.Object.Raw, nil, nil)
		if err != nil {
			return nil, err
		}
	}

	var oldObject runtime.Object
	if len(req.OldObject.Raw) > 0 {
		oldObject, _, err = scheme.Codecs.UniversalDeserializer().Decode(req.OldObject.Raw, nil, nil)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
	}

	var operationOptions runtime.Object
	if len(req.Options.Raw) > 0 {
		operationOptions, _, err = scheme.Codecs.UniversalDeserializer().Decode(req.Options.Raw, nil, nil)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
	}

	extras := map[string][]string{}
	for k, v := range req.UserInfo.Extra {
		extras[k] = v
	}

	attributesRecord := admissionapi.NewAttributesRecord(object,
		oldObject,
		schema.GroupVersionKind{
			Group:   req.RequestKind.Group,
			Version: req.RequestKind.Version,
			Kind:    req.RequestKind.Kind,
		},
		req.Namespace,
		req.Name,
		schema.GroupVersionResource{
			Group:    req.RequestResource.Group,
			Version:  req.RequestResource.Version,
			Resource: req.RequestResource.Resource,
		},
		req.SubResource,
		admissionapi.Operation(req.Operation),
		operationOptions,
		*req.DryRun,
		&user.DefaultInfo{
			Name:   req.UserInfo.Username,
			UID:    req.UserInfo.UID,
			Groups: req.UserInfo.Groups,
			Extra:  extras,
		})
	return attributesRecord, nil
}
