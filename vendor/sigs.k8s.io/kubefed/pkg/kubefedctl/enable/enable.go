/*
Copyright 2018 The Kubernetes Authors.

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

package enable

import (
	"context"
	"fmt"
	"io"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	apiextv1b1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextv1b1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	pkgruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/klog"

	"sigs.k8s.io/kubefed/pkg/apis/core/typeconfig"
	fedv1b1 "sigs.k8s.io/kubefed/pkg/apis/core/v1beta1"
	genericclient "sigs.k8s.io/kubefed/pkg/client/generic"
	ctlutil "sigs.k8s.io/kubefed/pkg/controller/util"
	"sigs.k8s.io/kubefed/pkg/kubefedctl/options"
	"sigs.k8s.io/kubefed/pkg/kubefedctl/util"
)

const (
	federatedGroupUsage = "The name of the API group to use for the generated federated type."
	targetVersionUsage  = "Optional, the API version of the target type."
)

var (
	enable_long = `
		Enables a Kubernetes API type (including a CRD) to be propagated
		to clusters registered with a KubeFed control plane.  A CRD for
		the federated type will be generated and a FederatedTypeConfig will
		be created to configure a sync controller.

		Current context is assumed to be a Kubernetes cluster hosting
		the kubefed control plane. Please use the
		--host-cluster-context flag otherwise.`

	enable_example = `
		# Enable federation of Deployments
		kubefedctl enable deployments.apps --host-cluster-context=cluster1

		# Enable federation of Deployments identified by name specified in
		# deployment.yaml
		kubefedctl enable -f deployment.yaml`
)

type enableType struct {
	options.GlobalSubcommandOptions
	options.CommonEnableOptions
	enableTypeOptions
}

type enableTypeOptions struct {
	federatedVersion    string
	output              string
	outputYAML          bool
	filename            string
	enableTypeDirective *EnableTypeDirective
}

// Bind adds the join specific arguments to the flagset passed in as an
// argument.
func (o *enableTypeOptions) Bind(flags *pflag.FlagSet) {
	flags.StringVar(&o.federatedVersion, "federated-version", options.DefaultFederatedVersion, "The API version to use for the generated federated type.")
	flags.StringVarP(&o.output, "output", "o", "", "If provided, the resources that would be created in the API by the command are instead output to stdout in the provided format.  Valid values are ['yaml'].")
	flags.StringVarP(&o.filename, "filename", "f", "", "If provided, the command will be configured from the provided yaml file.  Only --output will be accepted from the command line")
}

// NewCmdTypeEnable defines the `enable` command that
// enables federation of a Kubernetes API type.
func NewCmdTypeEnable(cmdOut io.Writer, config util.FedConfig) *cobra.Command {
	opts := &enableType{}

	cmd := &cobra.Command{
		Use:     "enable (NAME | -f FILENAME)",
		Short:   "Enables propagation of a Kubernetes API type",
		Long:    enable_long,
		Example: enable_example,
		Run: func(cmd *cobra.Command, args []string) {
			err := opts.Complete(args)
			if err != nil {
				klog.Fatalf("Error: %v", err)
			}

			err = opts.Run(cmdOut, config)
			if err != nil {
				klog.Fatalf("Error: %v", err)
			}
		},
	}

	flags := cmd.Flags()
	opts.GlobalSubcommandBind(flags)
	opts.CommonSubcommandBind(flags, federatedGroupUsage, targetVersionUsage)
	opts.Bind(flags)

	return cmd
}

// Complete ensures that options are valid and marshals them if necessary.
func (j *enableType) Complete(args []string) error {
	j.enableTypeDirective = NewEnableTypeDirective()
	fd := j.enableTypeDirective

	if j.output == "yaml" {
		j.outputYAML = true
	} else if len(j.output) > 0 {
		return errors.Errorf("Invalid value for --output: %s", j.output)
	}

	if len(j.filename) > 0 {
		err := DecodeYAMLFromFile(j.filename, fd)
		if err != nil {
			return errors.Wrapf(err, "Failed to load yaml from file %q", j.filename)
		}
		return nil
	}

	if err := j.SetName(args); err != nil {
		return err
	}

	fd.Name = j.TargetName

	if len(j.TargetVersion) > 0 {
		fd.Spec.TargetVersion = j.TargetVersion
	}
	if len(j.FederatedGroup) > 0 {
		fd.Spec.FederatedGroup = j.FederatedGroup
	}
	if len(j.federatedVersion) > 0 {
		fd.Spec.FederatedVersion = j.federatedVersion
	}

	return nil
}

// Run is the implementation of the `enable` command.
func (j *enableType) Run(cmdOut io.Writer, config util.FedConfig) error {
	hostConfig, err := config.HostConfig(j.HostClusterContext, j.Kubeconfig)
	if err != nil {
		return errors.Wrap(err, "Failed to get host cluster config")
	}

	resources, err := GetResources(hostConfig, j.enableTypeDirective)
	if err != nil {
		return err
	}

	if j.outputYAML {
		concreteTypeConfig := resources.TypeConfig.(*fedv1b1.FederatedTypeConfig)
		objects := []pkgruntime.Object{concreteTypeConfig, resources.CRD}
		err := writeObjectsToYAML(objects, cmdOut)
		if err != nil {
			return errors.Wrap(err, "Failed to write objects to YAML")
		}
		// -o yaml implies dry run
		return nil
	}

	return CreateResources(cmdOut, hostConfig, resources, j.KubeFedNamespace, j.DryRun)
}

type typeResources struct {
	TypeConfig typeconfig.Interface
	CRD        *apiextv1b1.CustomResourceDefinition
}

func GetResources(config *rest.Config, enableTypeDirective *EnableTypeDirective) (*typeResources, error) {
	apiResource, err := LookupAPIResource(config, enableTypeDirective.Name, enableTypeDirective.Spec.TargetVersion)
	if err != nil {
		return nil, err
	}
	klog.V(2).Infof("Found type %q", resourceKey(*apiResource))

	typeConfig := GenerateTypeConfigForTarget(*apiResource, enableTypeDirective)

	accessor, err := newSchemaAccessor(config, *apiResource)
	if err != nil {
		return nil, errors.Wrap(err, "Error initializing validation schema accessor")
	}

	shortNames := []string{}
	for _, shortName := range apiResource.ShortNames {
		shortNames = append(shortNames, fmt.Sprintf("f%s", shortName))
	}

	crd := federatedTypeCRD(typeConfig, accessor, shortNames)

	return &typeResources{
		TypeConfig: typeConfig,
		CRD:        crd,
	}, nil
}

// TODO(marun) Allow updates to the configuration for a type that has
// already been enabled for kubefed.  This would likely involve
// updating the version of the target type and the validation of the schema.
func CreateResources(cmdOut io.Writer, config *rest.Config, resources *typeResources, namespace string, dryRun bool) error {
	write := func(data string) {
		if cmdOut != nil {
			if _, err := cmdOut.Write([]byte(data)); err != nil {
				klog.Fatalf("Unexpected err: %v\n", err)
			}
		}
	}

	hostClientset, err := util.HostClientset(config)
	if err != nil {
		return errors.Wrap(err, "Failed to create host clientset")
	}
	_, err = hostClientset.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return errors.Wrapf(err, "KubeFed system namespace %q does not exist", namespace)
	} else if err != nil {
		return errors.Wrapf(err, "Error attempting to determine whether KubeFed system namespace %q exists", namespace)
	}

	client, err := genericclient.New(config)
	if err != nil {
		return errors.Wrap(err, "Failed to get kubefed clientset")
	}

	concreteTypeConfig := resources.TypeConfig.(*fedv1b1.FederatedTypeConfig)
	existingTypeConfig := &fedv1b1.FederatedTypeConfig{}
	err = client.Get(context.TODO(), existingTypeConfig, namespace, concreteTypeConfig.Name)
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrapf(err, "Error retrieving FederatedTypeConfig %q", concreteTypeConfig.Name)
	}
	if err == nil {
		fedType := existingTypeConfig.GetFederatedType()
		target := existingTypeConfig.GetTargetType()
		concreteType := concreteTypeConfig.GetFederatedType()
		if fedType.Name != concreteType.Name || fedType.Version != concreteType.Version || fedType.Group != concreteType.Group {
			return errors.Errorf("Federation is already enabled for %q with federated type %q. Changing the federated type to %q is not supported.",
				qualifiedAPIResourceName(target),
				qualifiedAPIResourceName(fedType),
				qualifiedAPIResourceName(concreteType))
		}
	}

	crdClient, err := apiextv1b1client.NewForConfig(config)
	if err != nil {
		return errors.Wrap(err, "Failed to create crd clientset")
	}

	existingCRD, err := crdClient.CustomResourceDefinitions().Get(resources.CRD.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		if !dryRun {
			_, err = crdClient.CustomResourceDefinitions().Create(resources.CRD)
			if err != nil {
				return errors.Wrapf(err, "Error creating CRD %q", resources.CRD.Name)
			}
		}
		write(fmt.Sprintf("customresourcedefinition.apiextensions.k8s.io/%s created\n", resources.CRD.Name))
	} else if err != nil {
		return errors.Wrapf(err, "Error getting CRD %q", resources.CRD.Name)
	} else {
		ftcs := &fedv1b1.FederatedTypeConfigList{}
		err := client.List(context.TODO(), ftcs, namespace)
		if err != nil {
			return errors.Wrap(err, "Error getting FederatedTypeConfig list")
		}

		for _, ftc := range ftcs.Items {
			targetAPI := concreteTypeConfig.Spec.TargetType
			existingAPI := ftc.Spec.TargetType
			if IsEquivalentAPI(&existingAPI, &targetAPI) {
				existingName := qualifiedAPIResourceName(ftc.GetTargetType())
				name := qualifiedAPIResourceName(concreteTypeConfig.GetTargetType())
				qualifiedFTCName := ctlutil.QualifiedName{
					Namespace: ftc.Namespace,
					Name:      ftc.Name,
				}

				return errors.Errorf("Failed to enable %q. Federation of this type is already enabled for equivalent type %q by FederatedTypeConfig %q",
					name, existingName, qualifiedFTCName)
			}

			if concreteTypeConfig.Name == ftc.Name {
				continue
			}

			fedType := ftc.Spec.FederatedType
			name := typeconfig.GroupQualifiedName(metav1.APIResource{Name: fedType.PluralName, Group: fedType.Group})
			if name == existingCRD.Name {
				return errors.Errorf("Failed to enable federation of %q due to the FederatedTypeConfig for %q already referencing a federated type CRD named %q. If these target types are distinct despite sharing the same kind, specifying a non-default --federated-group should allow %q to be enabled.",
					concreteTypeConfig.Name, ftc.Name, name, concreteTypeConfig.Name)
			}
		}

		existingCRD.Spec = resources.CRD.Spec
		if !dryRun {
			_, err = crdClient.CustomResourceDefinitions().Update(existingCRD)
			if err != nil {
				return errors.Wrapf(err, "Error updating CRD %q", resources.CRD.Name)
			}
		}
		write(fmt.Sprintf("customresourcedefinition.apiextensions.k8s.io/%s updated\n", resources.CRD.Name))
	}

	concreteTypeConfig.Namespace = namespace
	err = client.Get(context.TODO(), existingTypeConfig, namespace, concreteTypeConfig.Name)
	createdOrUpdated := "created"
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Wrapf(err, "Error retrieving FederatedTypeConfig %q", concreteTypeConfig.Name)
		}
		if !dryRun {
			err = client.Create(context.TODO(), concreteTypeConfig)
			if err != nil {
				return errors.Wrapf(err, "Error creating FederatedTypeConfig %q", concreteTypeConfig.Name)
			}
		}
	} else {
		existingTypeConfig.Spec = concreteTypeConfig.Spec
		if !dryRun {
			err = client.Update(context.TODO(), existingTypeConfig)
			if err != nil {
				return errors.Wrapf(err, "Error updating FederatedTypeConfig %q", concreteTypeConfig.Name)
			}
		}
		createdOrUpdated = "updated"
	}
	write(fmt.Sprintf("federatedtypeconfig.core.kubefed.io/%s %s in namespace %s\n",
		concreteTypeConfig.Name, createdOrUpdated, namespace))
	return nil
}

func GenerateTypeConfigForTarget(apiResource metav1.APIResource, enableTypeDirective *EnableTypeDirective) typeconfig.Interface {
	spec := enableTypeDirective.Spec
	kind := apiResource.Kind
	pluralName := apiResource.Name
	typeConfig := &fedv1b1.FederatedTypeConfig{
		// Explicitly including TypeMeta will ensure it will be
		// serialized properly to yaml.
		TypeMeta: metav1.TypeMeta{
			Kind:       "FederatedTypeConfig",
			APIVersion: "core.kubefed.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: typeconfig.GroupQualifiedName(apiResource),
		},
		Spec: fedv1b1.FederatedTypeConfigSpec{
			TargetType: fedv1b1.APIResource{
				Version: apiResource.Version,
				Kind:    kind,
				Scope:   NamespacedToScope(apiResource),
			},
			Propagation: fedv1b1.PropagationEnabled,
			FederatedType: fedv1b1.APIResource{
				Group:      spec.FederatedGroup,
				Version:    spec.FederatedVersion,
				Kind:       fmt.Sprintf("Federated%s", kind),
				PluralName: fmt.Sprintf("federated%s", pluralName),
				Scope:      FederatedNamespacedToScope(apiResource),
			},
		},
	}

	// Set defaults that would normally be set by the api
	fedv1b1.SetFederatedTypeConfigDefaults(typeConfig)
	return typeConfig
}

func qualifiedAPIResourceName(resource metav1.APIResource) string {
	if resource.Group == "" {
		return fmt.Sprintf("%s/%s", resource.Name, resource.Version)
	}
	return fmt.Sprintf("%s.%s/%s", resource.Name, resource.Group, resource.Version)
}

func federatedTypeCRD(typeConfig typeconfig.Interface, accessor schemaAccessor, shortNames []string) *apiextv1b1.CustomResourceDefinition {
	templateSchema := accessor.templateSchema()
	schema := federatedTypeValidationSchema(templateSchema)
	return CrdForAPIResource(typeConfig.GetFederatedType(), schema, shortNames)
}

func writeObjectsToYAML(objects []pkgruntime.Object, w io.Writer) error {
	for _, obj := range objects {
		if _, err := w.Write([]byte("---\n")); err != nil {
			return errors.Wrap(err, "Error encoding object to yaml")
		}

		if err := writeObjectToYAML(obj, w); err != nil {
			return errors.Wrap(err, "Error encoding object to yaml")
		}
	}
	return nil
}

func writeObjectToYAML(obj pkgruntime.Object, w io.Writer) error {
	json, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(obj)
	if err != nil {
		return err
	}

	unstructuredObj := &unstructured.Unstructured{}
	if _, _, err := unstructured.UnstructuredJSONScheme.Decode(json, nil, unstructuredObj); err != nil {
		return err
	}

	return util.WriteUnstructuredToYaml(unstructuredObj, w)
}
