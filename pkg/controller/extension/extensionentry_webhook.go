/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package extension

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"
)

var _ admission.CustomValidator = &ExtensionEntryWebhook{}
var _ kscontroller.Controller = &ExtensionEntryWebhook{}

func (r *ExtensionEntryWebhook) Name() string {
	return "extensionentry-webhook"
}

type ExtensionEntryWebhook struct {
	client.Client
}

func (r *ExtensionEntryWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return r.validateExtensionEntry(ctx, obj.(*extensionsv1alpha1.ExtensionEntry))
}

func (r *ExtensionEntryWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	return r.validateExtensionEntry(ctx, newObj.(*extensionsv1alpha1.ExtensionEntry))
}

func (r *ExtensionEntryWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (r *ExtensionEntryWebhook) validateExtensionEntry(ctx context.Context, extensionEntry *extensionsv1alpha1.ExtensionEntry) (admission.Warnings, error) {
	entryNameSet := sets.NewString()
	entryLinkSet := sets.NewString()
	for index, entry := range extensionEntry.Spec.Entries {
		entryProps := make(map[string]interface{})
		if err := json.Unmarshal(entry.Raw, &entryProps); err != nil {
			return nil, err
		}
		entryNameVal, ok := entryProps["name"]
		if !ok {
			return nil, fmt.Errorf("ExtensionEntry %s spec.entries[%d].name cannot be empty", extensionEntry.Name, index)
		}
		entryName, ok := entryNameVal.(string)
		if !ok {
			return nil, fmt.Errorf("ExtensionEntry %s spec.entries[%d].name %s must be string", extensionEntry.Name, index, entryName)
		}
		if entryNameSet.Has(entryName) {
			return nil, fmt.Errorf("ExtensionEntry %s spec.entries[%d].name %s is duplicated", extensionEntry.Name, index, entryName)
		}
		entryNameSet.Insert(entryName)

		entryLinkVal, ok := entryProps["link"]
		if !ok {
			continue
		}
		entryLink, ok := entryLinkVal.(string)
		if !ok {
			return nil, fmt.Errorf("ExtensionEntry %s spec.entries[%d].link %s must be string", extensionEntry.Name, index, entryName)
		}
		if entryLinkSet.Has(entryLink) {
			return nil, fmt.Errorf("ExtensionEntry %s spec.entries[%d].link %s is duplicated", extensionEntry.Name, index, entryLink)
		}
		entryLinkSet.Insert(entryLink)
	}

	extensionEntries := &extensionsv1alpha1.ExtensionEntryList{}
	if err := r.Client.List(ctx, extensionEntries, &client.ListOptions{}); err != nil {
		return nil, err
	}
	for _, target := range extensionEntries.Items {
		if target.Name == extensionEntry.Name {
			continue
		}
		for index, targetEntry := range target.Spec.Entries {
			entryProps := make(map[string]interface{})
			if err := json.Unmarshal(targetEntry.Raw, &entryProps); err != nil {
				return nil, err
			}
			entryNameVal, ok := entryProps["name"]
			if !ok {
				return nil, fmt.Errorf("ExtensionEntry %s spec.entries[%d].name cannot be empty", target.Name, index)
			}
			entryName, ok := entryNameVal.(string)
			if !ok {
				return nil, fmt.Errorf("ExtensionEntry %s spec.entries[%d].name %s must be string", extensionEntry.Name, index, entryName)
			}

			if entryNameSet.Has(entryName) {
				return nil, fmt.Errorf("ExtensionEntry %s spec.entries[].name %s is already exists", extensionEntry.Name, entryName)
			}

			entryLinkVal, ok := entryProps["link"]
			if !ok {
				continue
			}
			entryLink, ok := entryLinkVal.(string)
			if !ok {
				return nil, fmt.Errorf("ExtensionEntry %s spec.entries[%d].link %s must be string", extensionEntry.Name, index, entryName)
			}
			if entryLinkSet.Has(entryLink) {
				return nil, fmt.Errorf("ExtensionEntry %s spec.entries[].link %s is already exists", extensionEntry.Name, entryLink)
			}
		}
	}
	return nil, nil
}

func (r *ExtensionEntryWebhook) SetupWithManager(mgr *kscontroller.Manager) error {
	r.Client = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		WithValidator(r).
		For(&extensionsv1alpha1.ExtensionEntry{}).
		Complete()
}
