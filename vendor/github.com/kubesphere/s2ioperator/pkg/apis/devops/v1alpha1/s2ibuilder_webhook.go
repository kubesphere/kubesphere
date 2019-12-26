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
	"encoding/json"
	"fmt"
	"github.com/kubesphere/s2ioperator/pkg/errors"
	"github.com/kubesphere/s2ioperator/pkg/util/reflectutils"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	errorutil "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"strings"
)

const (
	DefaultRevisionId = "master"
	DefaultTag        = "latest"
)

// log is for logging in this package.
var s2ibuilderlog = logf.Log.WithName("s2ibuilder-resource")

func (r *S2iBuilder) SetupWebhookWithManager(mgr ctrl.Manager) error {
	kclient = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-devops-kubesphere-io-v1alpha1-s2ibuilder,mutating=true,failurePolicy=fail,groups=devops.kubesphere.io,resources=s2ibuilders,verbs=create;update,versions=v1alpha1,name=s2ibuilder.kb.io

var _ webhook.Defaulter = &S2iBuilder{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *S2iBuilder) Default() {
	s2ibuilderlog.Info("default", "name", r.Name)

	if r.Spec.Config.RevisionId == "" {
		r.Spec.Config.RevisionId = DefaultRevisionId
	}

	if r.Spec.Config.Tag == "" {
		r.Spec.Config.Tag = DefaultTag
	}

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-devops-kubesphere-io-v1alpha1-s2ibuilder,mutating=false,failurePolicy=fail,groups=devops.kubesphere.io,resources=s2ibuilders,versions=v1alpha1,name=vs2ibuilder.kb.io

var _ webhook.Validator = &S2iBuilder{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *S2iBuilder) ValidateCreate() error {
	s2ibuilderlog.Info("validate create", "name", r.Name)

	fromTemplate := false
	if r.Spec.FromTemplate != nil {
		t := &S2iBuilderTemplate{}
		err := kclient.Get(context.TODO(), types.NamespacedName{Name: r.Spec.FromTemplate.Name}, t)
		if err != nil {
			if k8serror.IsNotFound(err) {
				return fmt.Errorf("Template not found, pls check the template name  [%s] or create a template", r.Spec.FromTemplate.Name)
			}
			return err
		}

		errs := validateParameter(r.Spec.FromTemplate.Parameters, t.Spec.Parameters)
		if len(errs) != 0 {
			return errorutil.NewAggregate(errs)
		}
		var BaseImages []string
		for _, ImageInfo := range t.Spec.ContainerInfo {
			BaseImages = append(BaseImages, ImageInfo.BuilderImage)
		}
		if r.Spec.FromTemplate.BuilderImage != "" {
			if !reflectutils.Contains(r.Spec.FromTemplate.BuilderImage, BaseImages) {
				return fmt.Errorf("builder's baseImage [%s] not in builder baseImages [%v]",
					r.Spec.FromTemplate.BuilderImage, BaseImages)
			}
		}
		fromTemplate = true
	}
	if anno, ok := r.Annotations[AutoScaleAnnotations]; ok {
		if err := validatingS2iBuilderAutoScale(anno); err != nil {
			return errors.NewFieldInvalidValueWithReason(AutoScaleAnnotations, err.Error())
		}
	}
	if errs := validateConfig(r.Spec.Config, fromTemplate); len(errs) == 0 {
		return nil
	} else {
		return errorutil.NewAggregate(errs)
	}
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *S2iBuilder) ValidateUpdate(old runtime.Object) error {
	s2ibuilderlog.Info("validate update", "name", r.Name)

	s2ibuilderlog.Info("validate create", "name", r.Name)

	fromTemplate := false
	if r.Spec.FromTemplate != nil {
		t := &S2iBuilderTemplate{}
		err := kclient.Get(context.TODO(), types.NamespacedName{Name: r.Spec.FromTemplate.Name}, t)
		if err != nil {
			if k8serror.IsNotFound(err) {
				return fmt.Errorf("Template not found, pls check the template name  [%s] or create a template", r.Spec.FromTemplate.Name)
			}
			return err
		}

		errs := validateParameter(r.Spec.FromTemplate.Parameters, t.Spec.Parameters)
		if len(errs) != 0 {
			return errorutil.NewAggregate(errs)
		}
		var BaseImages []string
		for _, ImageInfo := range t.Spec.ContainerInfo {
			BaseImages = append(BaseImages, ImageInfo.BuilderImage)
		}
		if r.Spec.FromTemplate.BuilderImage != "" {
			if !reflectutils.Contains(r.Spec.FromTemplate.BuilderImage, BaseImages) {
				return fmt.Errorf("builder's baseImage [%s] not in builder baseImages [%v]",
					r.Spec.FromTemplate.BuilderImage, BaseImages)
			}
		}
		fromTemplate = true
	}
	if anno, ok := r.Annotations[AutoScaleAnnotations]; ok {
		if err := validatingS2iBuilderAutoScale(anno); err != nil {
			return errors.NewFieldInvalidValueWithReason(AutoScaleAnnotations, err.Error())
		}
	}
	if errs := validateConfig(r.Spec.Config, fromTemplate); len(errs) == 0 {
		return nil
	} else {
		return errorutil.NewAggregate(errs)
	}
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *S2iBuilder) ValidateDelete() error {
	s2ibuilderlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

// ValidateConfig returns a list of error from validation.
func validateConfig(config *S2iConfig, fromTemplate bool) []error {
	allErrs := make([]error, 0)
	if !config.IsBinaryURL && len(config.SourceURL) == 0 {
		allErrs = append(allErrs, errors.NewFieldRequired("sourceUrl"))
	}
	if !fromTemplate && len(config.BuilderImage) == 0 {
		allErrs = append(allErrs, errors.NewFieldRequired("builderImage"))
	}
	switch config.BuilderPullPolicy {
	case PullNever, PullAlways, PullIfNotPresent:
	default:
		allErrs = append(allErrs, errors.NewFieldInvalidValue("builderPullPolicy"))
	}
	if config.DockerNetworkMode != "" && !validateDockerNetworkMode(config.DockerNetworkMode) {
		allErrs = append(allErrs, errors.NewFieldInvalidValue("dockerNetworkMode"))
	}
	if config.Labels != nil {
		for k := range config.Labels {
			if len(k) == 0 {
				allErrs = append(allErrs, errors.NewFieldInvalidValue("labels"))
			}
		}
	}
	if config.BuilderImage != "" {
		if err := validateDockerReference(config.BuilderImage); err != nil {
			allErrs = append(allErrs, errors.NewFieldInvalidValueWithReason("builderImage", err.Error()))
		}
	}
	if config.RuntimeAuthentication != nil {
		if config.RuntimeAuthentication.SecretRef == nil {
			if config.RuntimeAuthentication.Username == "" && config.RuntimeAuthentication.Password == "" {
				allErrs = append(allErrs, errors.NewFieldRequired("RuntimeAuthentication username|password / secretRef"))
			}
		}
	}
	if config.IncrementalAuthentication != nil {
		if config.IncrementalAuthentication.SecretRef == nil {
			if config.IncrementalAuthentication.Username == "" && config.IncrementalAuthentication.Password == "" {
				allErrs = append(allErrs, errors.NewFieldRequired("IncrementalAuthentication username|password / secretRef"))
			}
		}
	}
	if config.PullAuthentication != nil {
		if config.PullAuthentication.SecretRef == nil {
			if config.PullAuthentication.Username == "" && config.PullAuthentication.Password == "" {
				allErrs = append(allErrs, errors.NewFieldRequired("PullAuthentication username|password / secretRef"))
			}
		}
	}
	if config.PushAuthentication != nil {
		if config.PushAuthentication.SecretRef == nil {
			if config.PushAuthentication.Username == "" && config.PushAuthentication.Password == "" {
				allErrs = append(allErrs, errors.NewFieldRequired("PushAuthentication username|password / secretRef"))
			}
		}
	}
	return allErrs
}

// validateDockerNetworkMode checks wether the network mode conforms to the docker remote API specification (v1.19)
// Supported values are: bridge, host, container:<name|id>, and netns:/proc/<pid>/ns/net
func validateDockerNetworkMode(mode DockerNetworkMode) bool {
	switch mode {
	case DockerNetworkModeBridge, DockerNetworkModeHost:
		return true
	}
	if strings.HasPrefix(string(mode), DockerNetworkModeContainerPrefix) {
		return true
	}
	if strings.HasPrefix(string(mode), DockerNetworkModeNetworkNamespacePrefix) {
		return true
	}
	return false
}

func validateParameter(user, tmpt []Parameter) []error {
	findParameter := func(name string, ps []Parameter) int {
		for index, v := range ps {
			if v.Key == name {
				return index
			}
		}
		return -1
	}
	allErrs := make([]error, 0)
	for _, v := range tmpt {
		index := findParameter(v.Key, user)
		if v.Required && (index == -1 || user[index].Value == "") {
			allErrs = append(allErrs, errors.NewFieldRequired("Parameter:"+v.Key))
		}
	}
	return allErrs
}

func validatingS2iBuilderAutoScale(anno string) error {

	s2iAutoScale := make([]S2iAutoScale, 0)
	if err := json.Unmarshal([]byte(anno), &s2iAutoScale); err != nil {
		return err
	}
	for _, scale := range s2iAutoScale {
		switch scale.Kind {
		case KindStatefulSet:
			return nil
		case KindDeployment:
			return nil
		default:
			return fmt.Errorf("unsupport workload type [%s], name [%s]", scale.Kind, scale.Name)
		}
	}
	return nil
}
