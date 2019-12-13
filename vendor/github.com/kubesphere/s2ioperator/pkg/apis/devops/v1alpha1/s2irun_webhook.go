/*
Copyright 2019 The Kubesphere Authors.

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

package v1alpha1

import (
	"context"
	"fmt"
	"github.com/kubesphere/s2ioperator/pkg/errors"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var (
	s2irunlog = logf.Log.WithName("s2irun-resource")
	kclient   client.Client
)

func (r *S2iRun) SetupWebhookWithManager(mgr ctrl.Manager) error {
	kclient = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-devops-kubesphere-io-v1alpha1-s2irun,mutating=false,failurePolicy=fail,groups=devops.kubesphere.io,resources=s2iruns,versions=v1alpha1,name=vs2irun.kb.io

var _ webhook.Validator = &S2iRun{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *S2iRun) ValidateCreate() error {
	s2irunlog.Info("validate create", "name", r.Name)

	origin := &S2iRun{}

	builder := &S2iBuilder{}

	err := kclient.Get(context.TODO(), types.NamespacedName{Namespace: r.Namespace, Name: r.Spec.BuilderName}, builder)
	if err != nil && !k8serror.IsNotFound(err) {
		return errors.NewFieldInvalidValueWithReason("no", "could not call k8s api")
	}
	if !k8serror.IsNotFound(err) {
		if r.Spec.NewSourceURL != "" && !builder.Spec.Config.IsBinaryURL {
			return errors.NewFieldInvalidValueWithReason("newSourceURL", "only b2i could set newSourceURL")
		}
	}

	err = kclient.Get(context.TODO(), types.NamespacedName{Namespace: r.Namespace, Name: r.Name}, origin)
	if !k8serror.IsNotFound(err) && origin.Status.RunState != "" && !reflect.DeepEqual(origin.Spec, r.Spec) {
		return errors.NewFieldInvalidValueWithReason("spec", "should not change s2i run spec when job started")
	}

	if r.Spec.NewTag != "" {
		validateImageName := fmt.Sprintf("validate:%s", r.Spec.NewTag)
		if err := validateDockerReference(validateImageName); err != nil {
			return err
		}
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *S2iRun) ValidateUpdate(old runtime.Object) error {
	s2irunlog.Info("validate update", "name", r.Name)

	origin := &S2iRun{}

	builder := &S2iBuilder{}

	err := kclient.Get(context.TODO(), types.NamespacedName{Namespace: r.Namespace, Name: r.Spec.BuilderName}, builder)
	if err != nil && !k8serror.IsNotFound(err) {
		return errors.NewFieldInvalidValueWithReason("no", "could not call k8s api")
	}
	if !k8serror.IsNotFound(err) {
		if r.Spec.NewSourceURL != "" && !builder.Spec.Config.IsBinaryURL {
			return errors.NewFieldInvalidValueWithReason("newSourceURL", "only b2i could set newSourceURL")
		}
	}

	err = kclient.Get(context.TODO(), types.NamespacedName{Namespace: r.Namespace, Name: r.Name}, origin)
	if !k8serror.IsNotFound(err) && origin.Status.RunState != "" && !reflect.DeepEqual(origin.Spec, r.Spec) {
		return errors.NewFieldInvalidValueWithReason("spec", "should not change s2i run spec when job started")
	}

	if r.Spec.NewTag != "" {
		validateImageName := fmt.Sprintf("validate:%s", r.Spec.NewTag)
		if err := validateDockerReference(validateImageName); err != nil {
			return err
		}
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *S2iRun) ValidateDelete() error {
	s2irunlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
