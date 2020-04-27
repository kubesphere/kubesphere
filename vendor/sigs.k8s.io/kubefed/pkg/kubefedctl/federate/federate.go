/*
Copyright 2019 The Kubernetes Authors.

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

package federate

import (
	"context"
	"io"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/klog"

	"sigs.k8s.io/kubefed/pkg/apis/core/typeconfig"
	fedv1b1 "sigs.k8s.io/kubefed/pkg/apis/core/v1beta1"
	genericclient "sigs.k8s.io/kubefed/pkg/client/generic"
	ctlutil "sigs.k8s.io/kubefed/pkg/controller/util"
	"sigs.k8s.io/kubefed/pkg/kubefedctl/enable"
	"sigs.k8s.io/kubefed/pkg/kubefedctl/options"
	"sigs.k8s.io/kubefed/pkg/kubefedctl/util"
)

const (
	createResourceRetryTimeout  = 10 * time.Second
	createResourceRetryInterval = 1 * time.Second
)

var (
	// Controller created resources should always be skipped while federating content
	controllerCreatedAPIResourceNames = []string{
		"endpoints",
		"events",
		"events.events.k8s.io",
		"propagatedversions.core.kubefed.io",
	}

	federate_long = `
		Federate creates a federated resource from a kubernetes resource.
		The target resource must exist in the cluster hosting the kubefed
		control plane. If the federated resource needs to be created in the
		API, the control plane must have a FederatedTypeConfig for the type
		of the kubernetes resource. If using with flag '-o yaml', it is not
		necessary for the FederatedTypeConfig to exist (or even for the
		kubefed API to be installed in the cluster).

		Current context is assumed to be a Kubernetes cluster hosting
		the kubefed control plane. Please use the --host-cluster-context
		flag otherwise.`

	federate_example = `
		# Federate resource named "my-cm" in namespace "my-ns" of kubernetes type "configmaps" (identified by short name "cm")
		kubefedctl federate cm "my-cm" -n "my-ns" --host-cluster-context=cluster1`
)

type federateResource struct {
	options.GlobalSubcommandOptions
	typeName             string
	resourceName         string
	resourceNamespace    string
	output               string
	outputYAML           bool
	enableType           bool
	federateContents     bool
	filename             string
	skipAPIResourceNames []string
}

func (j *federateResource) Bind(flags *pflag.FlagSet) {
	flags.StringVarP(&j.resourceNamespace, "namespace", "n", "", "The namespace of the resource to federate.")
	flags.StringVarP(&j.output, "output", "o", "", "If provided, the resource that would be created in the API by the command is instead output to stdout in the provided format.  Valid format is ['yaml'].")
	flags.BoolVarP(&j.enableType, "enable-type", "t", false, "If true, attempt to enable federation of the API type of the resource before creating the federated resource.")
	flags.BoolVarP(&j.federateContents, "contents", "c", false, "Applicable only to namespaces. If provided, the command will federate all resources within the namespace after federating the namespace.")
	flags.StringVarP(&j.filename, "filename", "f", "", "If specified, the provided yaml file will be used as the input for target resources to federate. This mode will only emit federated resource yaml to standard output. Other flag options if provided will be ignored.")
	flags.StringSliceVarP(&j.skipAPIResourceNames, "skip-api-resources", "s", []string{}, "Comma separated names of the api resources to skip when federating contents in a namespace. Name could be short name "+
		"(e.g. 'deploy), kind (e.g. 'deployment'), plural name (e.g. 'deployments'), group qualified plural name (e.g. 'deployments.apps') or group name itself (e.g. 'apps') to skip the whole group.")
}

// Complete ensures that options are valid.
func (j *federateResource) Complete(args []string) error {
	if j.output == "yaml" {
		j.outputYAML = true
	} else if len(j.output) > 0 {
		return errors.Errorf("Invalid value for --output: %s", j.output)
	}

	if len(j.filename) > 0 {
		if len(args) > 0 {
			return errors.Errorf("Flag '--filename' does not take any args. Got args: %v", args)
		}
		return nil
	}

	if len(args) == 0 {
		return errors.New("TYPE-NAME is required")
	}
	j.typeName = args[0]

	if len(args) == 1 {
		return errors.New("RESOURCE-NAME is required")
	}
	j.resourceName = args[1]

	if j.enableType && j.outputYAML {
		return errors.New("Flag '--enable-type' cannot be used with '--output [yaml]'")
	}

	return nil
}

// NewCmdFederateResource defines the `federate` command that federates a
// Kubernetes resource of the given kubernetes type.
func NewCmdFederateResource(cmdOut io.Writer, config util.FedConfig) *cobra.Command {
	opts := &federateResource{}

	cmd := &cobra.Command{
		Use:     "federate TYPE-NAME RESOURCE-NAME",
		Short:   "Federate creates a federated resource from a kubernetes resource",
		Long:    federate_long,
		Example: federate_example,
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
	opts.Bind(flags)

	return cmd
}

// Run is the implementation of the `federate resource` command.
func (j *federateResource) Run(cmdOut io.Writer, config util.FedConfig) error {

	if len(j.resourceNamespace) == 0 {
		var err error
		j.resourceNamespace, err = util.GetNamespace(j.HostClusterContext, j.Kubeconfig, config)
		if err != nil {
			return err
		}
	}

	hostConfig, err := config.HostConfig(j.HostClusterContext, j.Kubeconfig)
	if err != nil {
		return errors.Wrap(err, "Failed to get host cluster config")
	}

	if len(j.filename) > 0 {
		resources, err := DecodeUnstructuredFromFile(j.filename)
		if err != nil {
			return errors.Wrapf(err, "Failed to load yaml from file %q", j.filename)
		}
		federatedResources, err := FederateResources(resources)
		if err != nil {
			return err
		}

		err = WriteUnstructuredObjsToYaml(federatedResources, cmdOut)
		if err != nil {
			return errors.Wrap(err, "Failed to write federated resources to YAML")
		}
		return nil

	}

	qualifiedResourceName := ctlutil.QualifiedName{
		Namespace: j.resourceNamespace,
		Name:      j.resourceName,
	}
	artifacts, err := GetFederateArtifacts(hostConfig, j.typeName, j.KubeFedNamespace, qualifiedResourceName, j.enableType, j.outputYAML)
	if err != nil {
		return err
	}
	artifactsList := []*FederateArtifacts{}
	artifactsList = append(artifactsList, artifacts)

	kind := artifacts.typeConfig.GetTargetType().Kind
	if kind != ctlutil.NamespaceKind && j.federateContents {
		return errors.New("Flag '--contents' can only be used with type 'namespaces'.")
	}

	if kind == ctlutil.NamespaceKind && j.federateContents {
		containedArtifactsList, err := GetContainedArtifactsList(hostConfig, j.resourceName, j.KubeFedNamespace, j.skipAPIResourceNames, j.enableType, j.outputYAML)
		if err != nil {
			return err
		}
		artifactsList = append(artifactsList, containedArtifactsList...)
	}

	if j.outputYAML {
		for _, artifacts := range artifactsList {
			err := WriteUnstructuredObjsToYaml(artifacts.federatedResources, cmdOut)
			if err != nil {
				return errors.Wrap(err, "Failed to write federated resource to YAML")
			}
		}
		return nil
	}

	return CreateResources(cmdOut, hostConfig, artifactsList, j.KubeFedNamespace, j.enableType, j.DryRun)
}

func FederateResources(resources []*unstructured.Unstructured) ([]*unstructured.Unstructured, error) {
	var federatedResources []*unstructured.Unstructured
	for _, targetResource := range resources {
		// A Group, a Version and a Kind is sufficient for API Resource definition.
		gvk := targetResource.GroupVersionKind()

		// Construct an API Resource from above info.
		// TODO(irfanurrehman) Should we depend on the lookup from the
		// API Server instead, for some specific scenario?
		plural, singular := apimeta.UnsafeGuessKindToResource(gvk)
		apiResource := metav1.APIResource{
			Name:         plural.Resource,
			SingularName: singular.Resource,
			Group:        gvk.Group,
			Version:      gvk.Version,
			Kind:         gvk.Kind,
		}
		apiResource.Namespaced = targetResource.GetNamespace() == ""

		qualifiedName := ctlutil.NewQualifiedName(targetResource)
		typeConfig := enable.GenerateTypeConfigForTarget(apiResource, enable.NewEnableTypeDirective())
		federatedResource, err := FederatedResourceFromTargetResource(typeConfig, targetResource)
		if err != nil {
			return nil, errors.Wrapf(err, "Error getting %s from %s %q", typeConfig.GetFederatedType().Kind, typeConfig.GetTargetType().Kind, qualifiedName)
		}

		federatedResources = append(federatedResources, federatedResource)
	}

	return federatedResources, nil
}

type FederateArtifacts struct {
	// Identifies if typeConfig for this type is installed
	typeConfigInstalled bool

	// Identifies the type
	typeConfig typeconfig.Interface
	// List of federated resources of this type
	federatedResources []*unstructured.Unstructured
}

func GetFederateArtifacts(hostConfig *rest.Config, typeName, kubefedNamespace string, qualifiedName ctlutil.QualifiedName, enableType, outputYAML bool) (*FederateArtifacts, error) {
	// Lookup kubernetes API availability
	apiResource, err := enable.LookupAPIResource(hostConfig, typeName, "")
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to find target API resource %s", typeName)
	}
	klog.V(2).Infof("API Resource for %s found", typeName)

	typeConfigInstalled, typeConfig, err := getTypeConfig(hostConfig, *apiResource, kubefedNamespace, enableType, outputYAML)
	if err != nil {
		return nil, err
	}

	targetResource, err := getTargetResource(hostConfig, typeConfig, qualifiedName)
	if err != nil {
		return nil, err
	}

	federatedResource, err := FederatedResourceFromTargetResource(typeConfig, targetResource)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting %s from %s %q", typeConfig.GetFederatedType().Kind, typeConfig.GetTargetType().Kind, qualifiedName)
	}

	var federatedResources []*unstructured.Unstructured
	federatedResources = append(federatedResources, federatedResource)
	return &FederateArtifacts{
		typeConfigInstalled: typeConfigInstalled,
		typeConfig:          typeConfig,
		federatedResources:  federatedResources,
	}, nil
}

func getTypeConfig(hostConfig *rest.Config, apiResource metav1.APIResource, kubefedNamespace string, enableType, outputYAML bool) (bool, typeconfig.Interface, error) {
	resolvedTypeName := typeconfig.GroupQualifiedName(apiResource)
	installedTypeConfig, err := getInstalledTypeConfig(hostConfig, resolvedTypeName, kubefedNamespace)
	if err == nil {
		return true, installedTypeConfig, nil
	}
	notFound := apierrors.IsNotFound(err)
	if notFound && !outputYAML && !enableType {
		return false, nil, errors.Errorf("%v. Consider using '--enable-type' to optionally enable type while federating the resource", err)
	}

	generatedTypeConfig := enable.GenerateTypeConfigForTarget(apiResource, enable.NewEnableTypeDirective())
	if notFound && enableType { // We have already generated typeConfig to additionally enable type
		return false, generatedTypeConfig, nil
	}
	if outputYAML { // Output as yaml does not bother what error happened while accessing typeConfig
		klog.V(1).Infof("Falling back to a generated type config due to lookup failure: %v", err)
		return false, generatedTypeConfig, nil
	}
	return false, nil, err
}

func getInstalledTypeConfig(hostConfig *rest.Config, typeName, kubefedNamespace string) (typeconfig.Interface, error) {
	client, err := genericclient.New(hostConfig)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get generic client")
	}

	concreteTypeConfig := &fedv1b1.FederatedTypeConfig{}
	err = client.Get(context.TODO(), concreteTypeConfig, kubefedNamespace, typeName)
	if err != nil {
		return nil, err
	}
	return concreteTypeConfig, nil
}

func getTargetResource(hostConfig *rest.Config, typeConfig typeconfig.Interface, qualifiedName ctlutil.QualifiedName) (*unstructured.Unstructured, error) {
	targetAPIResource := typeConfig.GetTargetType()
	targetClient, err := ctlutil.NewResourceClient(hostConfig, &targetAPIResource)
	if err != nil {
		return nil, errors.Wrapf(err, "Error creating client for %s", targetAPIResource.Kind)
	}

	kind := targetAPIResource.Kind
	resource, err := targetClient.Resources(qualifiedName.Namespace).Get(qualifiedName.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "Error retrieving target %s %q", kind, qualifiedName)
	}

	klog.V(2).Infof("Target %s %q found", kind, qualifiedName)
	return resource, nil
}

func FederatedResourceFromTargetResource(typeConfig typeconfig.Interface, resource *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	fedAPIResource := typeConfig.GetFederatedType()
	targetResource := resource.DeepCopy()

	targetKind := typeConfig.GetTargetType().Kind

	// Special handling is needed for some controller set fields.
	switch targetKind {
	case ctlutil.NamespaceKind:
		{
			unstructured.RemoveNestedField(targetResource.Object, "spec", "finalizers")
		}
	case ctlutil.ServiceAccountKind:
		{
			unstructured.RemoveNestedField(targetResource.Object, ctlutil.SecretsField)
		}
	case ctlutil.ServiceKind:
		{
			var targetPorts []interface{}
			targetPorts, ok, err := unstructured.NestedSlice(targetResource.Object, "spec", "ports")
			if err != nil {
				return nil, err
			}
			if ok {
				for index := range targetPorts {
					port := targetPorts[index].(map[string]interface{})
					delete(port, "nodePort")
					targetPorts[index] = port
				}
				err := unstructured.SetNestedSlice(targetResource.Object, targetPorts, "spec", "ports")
				if err != nil {
					return nil, err
				}
			}
			unstructured.RemoveNestedField(targetResource.Object, "spec", "clusterIP")
		}
	}

	qualifiedName := ctlutil.NewQualifiedName(targetResource)
	resourceNamespace := getNamespace(typeConfig, qualifiedName)
	fedResource := &unstructured.Unstructured{}
	SetBasicMetaFields(fedResource, fedAPIResource, qualifiedName.Name, resourceNamespace, "")

	// Warn if annotations are present in case the intention is to
	// define annotations in the template of the federated resource.
	annotations, _, err := unstructured.NestedMap(targetResource.Object, "metadata", "annotations")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to retrieve metadata.annotations")
	}
	if len(annotations) > 0 {
		klog.Warningf("Annotations defined for %s %q will not appear in the template of the federated resource: %v", targetKind, qualifiedName, annotations)
	}

	if err := RemoveUnwantedFields(targetResource); err != nil {
		return nil, err
	}

	err = unstructured.SetNestedField(fedResource.Object, targetResource.Object, ctlutil.SpecField, ctlutil.TemplateField)
	if err != nil {
		return nil, err
	}
	err = unstructured.SetNestedStringMap(fedResource.Object, map[string]string{}, ctlutil.SpecField, ctlutil.PlacementField, ctlutil.ClusterSelectorField, ctlutil.MatchLabelsField)
	if err != nil {
		return nil, err
	}

	return fedResource, err
}

func getNamespace(typeConfig typeconfig.Interface, qualifiedName ctlutil.QualifiedName) string {
	if typeConfig.GetTargetType().Kind == ctlutil.NamespaceKind {
		return qualifiedName.Name
	}
	return qualifiedName.Namespace
}

func CreateResources(cmdOut io.Writer, hostConfig *rest.Config, artifactsList []*FederateArtifacts, namespace string, enableType, dryRun bool) error {
	for _, artifacts := range artifactsList {
		if enableType && !artifacts.typeConfigInstalled {
			enableTypeDirective := enable.NewEnableTypeDirective()
			enableTypeDirective.Name = artifacts.typeConfig.GetObjectMeta().Name
			typeResources, err := enable.GetResources(hostConfig, enableTypeDirective)
			if err != nil {
				return err
			}
			err = enable.CreateResources(cmdOut, hostConfig, typeResources, namespace, dryRun)
			if err != nil {
				return err
			}
		}

		err := CreateFederatedResources(hostConfig, artifacts.typeConfig, artifacts.federatedResources, dryRun)
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateFederatedResources(hostConfig *rest.Config, typeConfig typeconfig.Interface, federatedResources []*unstructured.Unstructured, dryRun bool) error {
	for _, federatedResource := range federatedResources {
		err := CreateFederatedResource(hostConfig, typeConfig, federatedResource, dryRun)
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateFederatedResource(hostConfig *rest.Config, typeConfig typeconfig.Interface, federatedResource *unstructured.Unstructured, dryRun bool) error {
	if typeConfig.GetTargetType().Kind == ctlutil.NamespaceKind {
		// TODO: irfanurrehman: Can a target namespace be federated into another namespace?
		klog.Infof("Resource to federate is a namespace. Given namespace will itself be the container for the federated namespace")
	}

	fedAPIResource := typeConfig.GetFederatedType()
	fedKind := fedAPIResource.Kind
	fedClient, err := ctlutil.NewResourceClient(hostConfig, &fedAPIResource)
	if err != nil {
		return errors.Wrapf(err, "Error creating client for %s", fedKind)
	}

	qualifiedFedName := ctlutil.NewQualifiedName(federatedResource)
	if !dryRun {
		// It might take a little while for the federated type to appear if the
		// same is being enabled while or immediately before federating the resource.
		err = wait.PollImmediate(createResourceRetryInterval, createResourceRetryTimeout, func() (bool, error) {
			_, err := fedClient.Resources(federatedResource.GetNamespace()).Create(federatedResource, metav1.CreateOptions{})
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			if err != nil {
				return false, err
			}
			return true, nil
		})
		if err != nil {
			return errors.Wrapf(err, "Error creating federated resource %q", qualifiedFedName)
		}
	}

	klog.Infof("Successfully created %s %q from %s", fedKind, qualifiedFedName, typeConfig.GetTargetType().Kind)
	return nil
}

func GetContainedArtifactsList(hostConfig *rest.Config, containerNamespace, kubefedNamespace string, skipAPIResourceNames []string, enableType, outputYAML bool) ([]*FederateArtifacts, error) {
	targetResourcesList, err := getResourcesInNamespace(hostConfig, containerNamespace, skipAPIResourceNames)
	if err != nil {
		return nil, err
	}

	artifactsList := []*FederateArtifacts{}
	for _, targetResources := range targetResourcesList {
		apiResource := targetResources.apiResource
		typeConfigInstalled, typeConfig, err := getTypeConfig(hostConfig, apiResource, kubefedNamespace, enableType, outputYAML)
		if err != nil {
			return nil, err
		}
		var federatedResources []*unstructured.Unstructured
		for _, targetResource := range targetResources.resources {
			federatedResource, err := FederatedResourceFromTargetResource(typeConfig, targetResource)
			if err != nil {
				return nil, err
			}

			federatedResources = append(federatedResources, federatedResource)
		}
		federateArtifacts := FederateArtifacts{
			typeConfigInstalled: typeConfigInstalled,
			typeConfig:          typeConfig,
			federatedResources:  federatedResources,
		}
		artifactsList = append(artifactsList, &federateArtifacts)
	}

	return artifactsList, nil
}

func WriteUnstructuredObjsToYaml(unstructuredObjs []*unstructured.Unstructured, w io.Writer) error {
	for _, unstructuredObj := range unstructuredObjs {
		if _, err := w.Write([]byte("---\n")); err != nil {
			return errors.Wrap(err, "Error encoding object to yaml")
		}
		err := util.WriteUnstructuredToYaml(unstructuredObj, w)
		if err != nil {
			return err
		}
	}

	return nil
}
