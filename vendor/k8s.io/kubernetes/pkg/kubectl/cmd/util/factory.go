/*
Copyright 2014 The Kubernetes Authors.

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

package util

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	scaleclient "k8s.io/client-go/scale"
	"k8s.io/client-go/tools/clientcmd"
	api "k8s.io/kubernetes/pkg/apis/core"
	apiv1 "k8s.io/kubernetes/pkg/apis/core/v1"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	coreclient "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/core/internalversion"
	"k8s.io/kubernetes/pkg/kubectl"
	"k8s.io/kubernetes/pkg/kubectl/categories"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/openapi"
	"k8s.io/kubernetes/pkg/kubectl/plugins"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/kubectl/validation"
	"k8s.io/kubernetes/pkg/printers"
)

const (
	FlagMatchBinaryVersion = "match-server-version"
)

var (
	FlagHTTPCacheDir = "cache-dir"
)

// Factory provides abstractions that allow the Kubectl command to be extended across multiple types
// of resources and different API sets.
// The rings are here for a reason. In order for composers to be able to provide alternative factory implementations
// they need to provide low level pieces of *certain* functions so that when the factory calls back into itself
// it uses the custom version of the function. Rather than try to enumerate everything that someone would want to override
// we split the factory into rings, where each ring can depend on methods in an earlier ring, but cannot depend
// upon peer methods in its own ring.
// TODO: make the functions interfaces
// TODO: pass the various interfaces on the factory directly into the command constructors (so the
// commands are decoupled from the factory).
type Factory interface {
	ClientAccessFactory
	ObjectMappingFactory
	BuilderFactory
}

type DiscoveryClientFactory interface {
	// Returns a discovery client
	DiscoveryClient() (discovery.CachedDiscoveryInterface, error)

	// BindFlags adds any discovery flags that are common to all kubectl sub commands.
	BindFlags(flags *pflag.FlagSet)
}

// ClientAccessFactory holds the first level of factory methods.
// Generally provides discovery, negotiation, and no-dep calls.
// TODO The polymorphic calls probably deserve their own interface.
type ClientAccessFactory interface {
	// Returns a discovery client
	DiscoveryClient() (discovery.CachedDiscoveryInterface, error)

	// ClientSet gives you back an internal, generated clientset
	ClientSet() (internalclientset.Interface, error)

	// DynamicClient returns a dynamic client ready for use
	DynamicClient() (dynamic.DynamicInterface, error)

	// KubernetesClientSet gives you back an external clientset
	KubernetesClientSet() (*kubernetes.Clientset, error)

	// Returns a RESTClient for accessing Kubernetes resources or an error.
	RESTClient() (*restclient.RESTClient, error)
	// Returns a client.Config for accessing the Kubernetes server.
	ClientConfig() (*restclient.Config, error)
	// BareClientConfig returns a client.Config that has NOT been negotiated. It's
	// just directions to the server. People use this to build RESTMappers on top of
	BareClientConfig() (*restclient.Config, error)

	// UpdatePodSpecForObject will call the provided function on the pod spec this object supports,
	// return false if no pod spec is supported, or return an error.
	UpdatePodSpecForObject(obj runtime.Object, fn func(*v1.PodSpec) error) (bool, error)

	// MapBasedSelectorForObject returns the map-based selector associated with the provided object. If a
	// new set-based selector is provided, an error is returned if the selector cannot be converted to a
	// map-based selector
	MapBasedSelectorForObject(object runtime.Object) (string, error)
	// PortsForObject returns the ports associated with the provided object
	PortsForObject(object runtime.Object) ([]string, error)
	// ProtocolsForObject returns the <port, protocol> mapping associated with the provided object
	ProtocolsForObject(object runtime.Object) (map[string]string, error)
	// LabelsForObject returns the labels associated with the provided object
	LabelsForObject(object runtime.Object) (map[string]string, error)

	// Command will stringify and return all environment arguments ie. a command run by a client
	// using the factory.
	Command(cmd *cobra.Command, showSecrets bool) string
	// BindFlags adds any flags that are common to all kubectl sub commands.
	BindFlags(flags *pflag.FlagSet)
	// BindExternalFlags adds any flags defined by external projects (not part of pflags)
	BindExternalFlags(flags *pflag.FlagSet)

	// SuggestedPodTemplateResources returns a list of resource types that declare a pod template
	SuggestedPodTemplateResources() []schema.GroupResource

	// Pauser marks the object in the info as paused. Currently supported only for Deployments.
	// Returns the patched object in bytes and any error that occurred during the encoding or
	// in case the object is already paused.
	Pauser(info *resource.Info) ([]byte, error)
	// Resumer resumes a paused object inside the info. Currently supported only for Deployments.
	// Returns the patched object in bytes and any error that occurred during the encoding or
	// in case the object is already resumed.
	Resumer(info *resource.Info) ([]byte, error)

	// ResolveImage resolves the image names. For kubernetes this function is just
	// passthrough but it allows to perform more sophisticated image name resolving for
	// third-party vendors.
	ResolveImage(imageName string) (string, error)

	// Returns the default namespace to use in cases where no
	// other namespace is specified and whether the namespace was
	// overridden.
	DefaultNamespace() (string, bool, error)
	// Generators returns the generators for the provided command
	Generators(cmdName string) map[string]kubectl.Generator
	// Check whether the kind of resources could be exposed
	CanBeExposed(kind schema.GroupKind) error
	// Check whether the kind of resources could be autoscaled
	CanBeAutoscaled(kind schema.GroupKind) error

	// EditorEnvs returns a group of environment variables that the edit command
	// can range over in order to determine if the user has specified an editor
	// of their choice.
	EditorEnvs() []string
}

// ObjectMappingFactory holds the second level of factory methods. These functions depend upon ClientAccessFactory methods.
// Generally they provide object typing and functions that build requests based on the negotiated clients.
type ObjectMappingFactory interface {
	// Returns interfaces for dealing with arbitrary runtime.Objects.
	RESTMapper() (meta.RESTMapper, error)
	// Returns interface for expanding categories like `all`.
	CategoryExpander() categories.CategoryExpander
	// Returns a RESTClient for working with the specified RESTMapping or an error. This is intended
	// for working with arbitrary resources and is not guaranteed to point to a Kubernetes APIServer.
	ClientForMapping(mapping *meta.RESTMapping) (resource.RESTClient, error)
	// Returns a RESTClient for working with Unstructured objects.
	UnstructuredClientForMapping(mapping *meta.RESTMapping) (resource.RESTClient, error)
	// Returns a Describer for displaying the specified RESTMapping type or an error.
	Describer(mapping *meta.RESTMapping) (printers.Describer, error)

	// LogsForObject returns a request for the logs associated with the provided object
	LogsForObject(object, options runtime.Object, timeout time.Duration) (*restclient.Request, error)
	// Returns a HistoryViewer for viewing change history
	HistoryViewer(mapping *meta.RESTMapping) (kubectl.HistoryViewer, error)
	// Returns a Rollbacker for changing the rollback version of the specified RESTMapping type or an error
	Rollbacker(mapping *meta.RESTMapping) (kubectl.Rollbacker, error)
	// Returns a StatusViewer for printing rollout status.
	StatusViewer(mapping *meta.RESTMapping) (kubectl.StatusViewer, error)

	// AttachablePodForObject returns the pod to which to attach given an object.
	AttachablePodForObject(object runtime.Object, timeout time.Duration) (*api.Pod, error)

	// ApproximatePodTemplateForObject returns a pod template object for the provided source.
	// It may return both an error and a object. It attempt to return the best possible template
	// available at the current time.
	ApproximatePodTemplateForObject(runtime.Object) (*api.PodTemplateSpec, error)

	// Returns a schema that can validate objects stored on disk.
	Validator(validate bool) (validation.Schema, error)
	// OpenAPISchema returns the schema openapi schema definition
	OpenAPISchema() (openapi.Resources, error)
}

// BuilderFactory holds the third level of factory methods. These functions depend upon ObjectMappingFactory and ClientAccessFactory methods.
// Generally they depend upon client mapper functions
type BuilderFactory interface {
	// NewBuilder returns an object that assists in loading objects from both disk and the server
	// and which implements the common patterns for CLI interactions with generic resources.
	NewBuilder() *resource.Builder
	// PluginLoader provides the implementation to be used to load cli plugins.
	PluginLoader() plugins.PluginLoader
	// PluginRunner provides the implementation to be used to run cli plugins.
	PluginRunner() plugins.PluginRunner
	// Returns a Scaler for changing the size of the specified RESTMapping type or an error
	Scaler() (kubectl.Scaler, error)
	// ScaleClient gives you back scale getter
	ScaleClient() (scaleclient.ScalesGetter, error)
	// Returns a Reaper for gracefully shutting down resources.
	Reaper(mapping *meta.RESTMapping) (kubectl.Reaper, error)
}

type factory struct {
	ClientAccessFactory
	ObjectMappingFactory
	BuilderFactory
}

// NewFactory creates a factory with the default Kubernetes resources defined
// if optionalClientConfig is nil, then flags will be bound to a new clientcmd.ClientConfig.
// if optionalClientConfig is not nil, then this factory will make use of it.
func NewFactory(optionalClientConfig clientcmd.ClientConfig) Factory {
	clientAccessFactory := NewClientAccessFactory(optionalClientConfig)
	objectMappingFactory := NewObjectMappingFactory(clientAccessFactory)
	builderFactory := NewBuilderFactory(clientAccessFactory, objectMappingFactory)

	return &factory{
		ClientAccessFactory:  clientAccessFactory,
		ObjectMappingFactory: objectMappingFactory,
		BuilderFactory:       builderFactory,
	}
}

// GetFirstPod returns a pod matching the namespace and label selector
// and the number of all pods that match the label selector.
func GetFirstPod(client coreclient.PodsGetter, namespace string, selector string, timeout time.Duration, sortBy func([]*v1.Pod) sort.Interface) (*api.Pod, int, error) {
	options := metav1.ListOptions{LabelSelector: selector}

	podList, err := client.Pods(namespace).List(options)
	if err != nil {
		return nil, 0, err
	}
	pods := []*v1.Pod{}
	for i := range podList.Items {
		pod := podList.Items[i]
		externalPod := &v1.Pod{}
		apiv1.Convert_core_Pod_To_v1_Pod(&pod, externalPod, nil)
		pods = append(pods, externalPod)
	}
	if len(pods) > 0 {
		sort.Sort(sortBy(pods))
		internalPod := &api.Pod{}
		apiv1.Convert_v1_Pod_To_core_Pod(pods[0], internalPod, nil)
		return internalPod, len(podList.Items), nil
	}

	// Watch until we observe a pod
	options.ResourceVersion = podList.ResourceVersion
	w, err := client.Pods(namespace).Watch(options)
	if err != nil {
		return nil, 0, err
	}
	defer w.Stop()

	condition := func(event watch.Event) (bool, error) {
		return event.Type == watch.Added || event.Type == watch.Modified, nil
	}
	event, err := watch.Until(timeout, w, condition)
	if err != nil {
		return nil, 0, err
	}
	pod, ok := event.Object.(*api.Pod)
	if !ok {
		return nil, 0, fmt.Errorf("%#v is not a pod event", event)
	}
	return pod, 1, nil
}

func makePortsString(ports []api.ServicePort, useNodePort bool) string {
	pieces := make([]string, len(ports))
	for ix := range ports {
		var port int32
		if useNodePort {
			port = ports[ix].NodePort
		} else {
			port = ports[ix].Port
		}
		pieces[ix] = fmt.Sprintf("%s:%d", strings.ToLower(string(ports[ix].Protocol)), port)
	}
	return strings.Join(pieces, ",")
}

func getPorts(spec api.PodSpec) []string {
	result := []string{}
	for _, container := range spec.Containers {
		for _, port := range container.Ports {
			result = append(result, strconv.Itoa(int(port.ContainerPort)))
		}
	}
	return result
}

func getProtocols(spec api.PodSpec) map[string]string {
	result := make(map[string]string)
	for _, container := range spec.Containers {
		for _, port := range container.Ports {
			result[strconv.Itoa(int(port.ContainerPort))] = string(port.Protocol)
		}
	}
	return result
}

// Extracts the ports exposed by a service from the given service spec.
func getServicePorts(spec api.ServiceSpec) []string {
	result := []string{}
	for _, servicePort := range spec.Ports {
		result = append(result, strconv.Itoa(int(servicePort.Port)))
	}
	return result
}

// Extracts the protocols exposed by a service from the given service spec.
func getServiceProtocols(spec api.ServiceSpec) map[string]string {
	result := make(map[string]string)
	for _, servicePort := range spec.Ports {
		result[strconv.Itoa(int(servicePort.Port))] = string(servicePort.Protocol)
	}
	return result
}
