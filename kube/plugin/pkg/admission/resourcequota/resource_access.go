/*
Copyright 2016 The Kubernetes Authors.

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

package resourcequota

import (
	corev1 "k8s.io/api/core/v1"
)

// QuotaAccessor abstracts the get/set logic from the rest of the Evaluator.  This could be a test stub, a straight passthrough,
// or most commonly a series of deconflicting caches.
type QuotaAccessor interface {
	// UpdateQuotaStatus is called to persist final status.  This method should write to persistent storage.
	// An error indicates that write didn't complete successfully.
	UpdateQuotaStatus(newQuota *corev1.ResourceQuota) error

	// GetQuotas gets all possible quotas for a given namespace
	GetQuotas(namespace string) ([]corev1.ResourceQuota, error)
}
