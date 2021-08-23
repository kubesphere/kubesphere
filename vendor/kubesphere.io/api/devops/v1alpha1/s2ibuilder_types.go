/*
Copyright 2020 The KubeSphere Authors.

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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type RunState string

const (
	ResourceKindS2iBuilder     = "S2iBuilder"
	ResourceSingularS2iBuilder = "s2ibuilder"
	ResourcePluralS2iBuilder   = "s2ibuilders"
)

const (
	NotRunning RunState = "Not Running Yet"
	Running    RunState = "Running"
	Successful RunState = "Successful"
	Failed     RunState = "Failed"
	Unknown    RunState = "Unknown"
)
const (
	AutoScaleAnnotations             = "devops.kubesphere.io/autoscale"
	S2iRunLabel                      = "devops.kubesphere.io/s2ir"
	S2irCompletedScaleAnnotations    = "devops.kubesphere.io/completedscale"
	WorkLoadCompletedInitAnnotations = "devops.kubesphere.io/inithasbeencomplted"
	S2iRunDoNotAutoScaleAnnotations  = "devops.kubesphere.io/donotautoscale"
	DescriptionAnnotations           = "desc"
)
const (
	KindDeployment  = "Deployment"
	KindStatefulSet = "StatefulSet"
)

// EnvironmentSpec specifies a single environment variable.
type EnvironmentSpec struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ProxyConfig holds proxy configuration.
type ProxyConfig struct {
	HTTPProxy  string `json:"httpProxy,omitempty"`
	HTTPSProxy string `json:"httpsProxy,omitempty"`
}

// CGroupLimits holds limits used to constrain container resources.
type CGroupLimits struct {
	MemoryLimitBytes int64  `json:"memoryLimitBytes"`
	CPUShares        int64  `json:"cpuShares"`
	CPUPeriod        int64  `json:"cpuPeriod"`
	CPUQuota         int64  `json:"cpuQuota"`
	MemorySwap       int64  `json:"memorySwap"`
	Parent           string `json:"parent"`
}

// VolumeSpec represents a single volume mount point.
type VolumeSpec struct {
	// Source is a reference to the volume source.
	Source string `json:"source,omitempty"`
	// Destination is the path to mount the volume to - absolute or relative.
	Destination string `json:"destination,omitempty"`
	// Keep indicates if the mounted data should be kept in the final image.
	Keep bool `json:"keep,omitempty"`
}

// DockerConfig contains the configuration for a Docker connection.
type DockerConfig struct {
	// Endpoint is the docker network endpoint or socket
	Endpoint string `json:"endPoint"`

	// CertFile is the certificate file path for a TLS connection
	CertFile string `json:"certFile"`

	// KeyFile is the key file path for a TLS connection
	KeyFile string `json:"keyFile"`

	// CAFile is the certificate authority file path for a TLS connection
	CAFile string `json:"caFile"`

	// UseTLS indicates if TLS must be used
	UseTLS bool `json:"useTLS"`

	// TLSVerify indicates if TLS peer must be verified
	TLSVerify bool `json:"tlsVerify"`
}

// AuthConfig is our abstraction of the Registry authorization information for whatever
// docker client we happen to be based on
type AuthConfig struct {
	Username      string                       `json:"username,omitempty"`
	Password      string                       `json:"password,omitempty"`
	Email         string                       `json:"email,omitempty"`
	ServerAddress string                       `json:"serverAddress,omitempty"`
	SecretRef     *corev1.LocalObjectReference `json:"secretRef,omitempty"`
}

// ContainerConfig is the abstraction of the docker client provider (formerly go-dockerclient, now either
// engine-api or kube docker client) container.Config type that is leveraged by s2i or origin
type ContainerConfig struct {
	Labels map[string]string
	Env    []string
}

type PullPolicy string

const (
	// PullAlways means that we always attempt to pull the latest image.
	PullAlways PullPolicy = "always"

	// PullNever means that we never pull an image, but only use a local image.
	PullNever PullPolicy = "never"

	// PullIfNotPresent means that we pull if the image isn't present on disk.
	PullIfNotPresent PullPolicy = "if-not-present"

	// DefaultBuilderPullPolicy specifies the default pull policy to use
	DefaultBuilderPullPolicy = PullIfNotPresent

	// DefaultRuntimeImagePullPolicy specifies the default pull policy to use.
	DefaultRuntimeImagePullPolicy = PullIfNotPresent

	// DefaultPreviousImagePullPolicy specifies policy for pulling the previously
	// build Docker image when doing incremental build
	DefaultPreviousImagePullPolicy = PullIfNotPresent
)

// DockerNetworkMode specifies the network mode setting for the docker container
type DockerNetworkMode string

const (
	// DockerNetworkModeHost places the container in the default (host) network namespace.
	DockerNetworkModeHost DockerNetworkMode = "host"
	// DockerNetworkModeBridge instructs docker to create a network namespace for this container connected to the docker0 bridge via a veth-pair.
	DockerNetworkModeBridge DockerNetworkMode = "bridge"
	// DockerNetworkModeContainerPrefix is the string prefix used by NewDockerNetworkModeContainer.
	DockerNetworkModeContainerPrefix string = "container:"
	// DockerNetworkModeNetworkNamespacePrefix is the string prefix used when sharing a namespace from a CRI-O container.
	DockerNetworkModeNetworkNamespacePrefix string = "netns:"
)

type TriggerSource string

const (
	Default TriggerSource = "Manual"
	Github  TriggerSource = "Github"
	Gitlab  TriggerSource = "Gitlab"
	SVN     TriggerSource = "SVN"
	Others  TriggerSource = "Others"
)

// NewDockerNetworkModeContainer creates a DockerNetworkMode value which instructs docker to place the container in the network namespace of an existing container.
// It can be used, for instance, to place the s2i container in the network namespace of the infrastructure container of a k8s pod.
func NewDockerNetworkModeContainer(id string) DockerNetworkMode {
	return DockerNetworkMode(DockerNetworkModeContainerPrefix + id)
}

// String implements the String() function of pflags.Value so this can be used as
// command line parameter.
// This method is really used just to show the default value when printing help.
// It will not default the configuration.
func (p *PullPolicy) String() string {
	if len(string(*p)) == 0 {
		return string(DefaultBuilderPullPolicy)
	}
	return string(*p)
}

// Type implements the Type() function of pflags.Value interface
func (p *PullPolicy) Type() string {
	return "string"
}

// Set implements the Set() function of pflags.Value interface
// The valid options are "always", "never" or "if-not-present"
func (p *PullPolicy) Set(v string) error {
	switch v {
	case "always":
		*p = PullAlways
	case "never":
		*p = PullNever
	case "if-not-present":
		*p = PullIfNotPresent
	default:
		return fmt.Errorf("invalid value %q, valid values are: always, never or if-not-present", v)
	}
	return nil
}

type S2iConfig struct {
	// DisplayName is a result image display-name label. This defaults to the
	// output image name.
	DisplayName string `json:"displayName,omitempty"`

	// Description is a result image description label. The default is no
	// description.
	Description string `json:"description,omitempty"`

	// BuilderImage describes which image is used for building the result images.
	BuilderImage string `json:"builderImage,omitempty"`

	// BuilderImageVersion provides optional version information about the builder image.
	BuilderImageVersion string `json:"builderImageVersion,omitempty"`

	// BuilderBaseImageVersion provides optional version information about the builder base image.
	BuilderBaseImageVersion string `json:"builderBaseImageVersion,omitempty"`

	// RuntimeImage specifies the image that will be a base for resulting image
	// and will be used for running an application. By default, BuilderImage is
	// used for building and running, but the latter may be overridden.
	RuntimeImage string `json:"runtimeImage,omitempty"`

	//OutputImageName is a result image name without tag, default is latest. tag will append to ImageName in the end
	OutputImageName string `json:"outputImageName,omitempty"`
	// RuntimeImagePullPolicy specifies when to pull a runtime image.
	RuntimeImagePullPolicy PullPolicy `json:"runtimeImagePullPolicy,omitempty"`

	// RuntimeAuthentication holds the authentication information for pulling the
	// runtime Docker images from private repositories.
	RuntimeAuthentication *AuthConfig `json:"runtimeAuthentication,omitempty"`

	// RuntimeArtifacts specifies a list of source/destination pairs that will
	// be copied from builder to a runtime image. Source can be a file or
	// directory. Destination must be a directory. Regardless whether it
	// is an absolute or relative path, it will be placed into image's WORKDIR.
	// Destination also can be empty or equals to ".", in this case it just
	// refers to a root of WORKDIR.
	// In case it's empty, S2I will try to get this list from
	// io.openshift.s2i.assemble-input-files label on a RuntimeImage.
	RuntimeArtifacts []VolumeSpec `json:"runtimeArtifacts,omitempty"`

	// DockerConfig describes how to access host docker daemon.
	DockerConfig *DockerConfig `json:"dockerConfig,omitempty"`

	// PullAuthentication holds the authentication information for pulling the
	// Docker images from private repositories
	PullAuthentication *AuthConfig `json:"pullAuthentication,omitempty"`

	// PullAuthentication holds the authentication information for pulling the
	// Docker images from private repositories
	PushAuthentication *AuthConfig `json:"pushAuthentication,omitempty"`

	// IncrementalAuthentication holds the authentication information for pulling the
	// previous image from private repositories
	IncrementalAuthentication *AuthConfig `json:"incrementalAuthentication,omitempty"`

	// DockerNetworkMode is used to set the docker network setting to --net=container:<id>
	// when the builder is invoked from a container.
	DockerNetworkMode DockerNetworkMode `json:"dockerNetworkMode,omitempty"`

	// PreserveWorkingDir describes if working directory should be left after processing.
	PreserveWorkingDir bool `json:"preserveWorkingDir,omitempty"`

	//ImageName Contains the registry address and reponame, tag should set by field tag alone
	ImageName string `json:"imageName"`
	// Tag is a result image tag name.
	Tag string `json:"tag,omitempty"`

	// BuilderPullPolicy specifies when to pull the builder image
	BuilderPullPolicy PullPolicy `json:"builderPullPolicy,omitempty"`

	// PreviousImagePullPolicy specifies when to pull the previously build image
	// when doing incremental build
	PreviousImagePullPolicy PullPolicy `json:"previousImagePullPolicy,omitempty"`

	// Incremental describes whether to try to perform incremental build.
	Incremental bool `json:"incremental,omitempty"`

	// IncrementalFromTag sets an alternative image tag to look for existing
	// artifacts. Tag is used by default if this is not set.
	IncrementalFromTag string `json:"incrementalFromTag,omitempty"`

	// RemovePreviousImage describes if previous image should be removed after successful build.
	// This applies only to incremental builds.
	RemovePreviousImage bool `json:"removePreviousImage,omitempty"`

	// Environment is a map of environment variables to be passed to the image.
	Environment []EnvironmentSpec `json:"environment,omitempty"`

	// LabelNamespace provides the namespace under which the labels will be generated.
	LabelNamespace string `json:"labelNamespace,omitempty"`

	// CallbackURL is a URL which is called upon successful build to inform about that fact.
	CallbackURL string `json:"callbackUrl,omitempty"`

	// ScriptsURL is a URL describing where to fetch the S2I scripts from during build process.
	// This url can be a reference within the builder image if the scheme is specified as image://
	ScriptsURL string `json:"scriptsUrl,omitempty"`

	// Destination specifies a location where the untar operation will place its artifacts.
	Destination string `json:"destination,omitempty"`

	// WorkingDir describes temporary directory used for downloading sources, scripts and tar operations.
	WorkingDir string `json:"workingDir,omitempty"`

	// WorkingSourceDir describes the subdirectory off of WorkingDir set up during the repo download
	// that is later used as the root for ignore processing
	WorkingSourceDir string `json:"workingSourceDir,omitempty"`

	// LayeredBuild describes if this is build which layered scripts and sources on top of BuilderImage.
	LayeredBuild bool `json:"layeredBuild,omitempty"`

	// Specify a relative directory inside the application repository that should
	// be used as a root directory for the application.
	ContextDir string `json:"contextDir,omitempty"`

	// AssembleUser specifies the user to run the assemble script in container
	AssembleUser string `json:"assembleUser,omitempty"`

	// RunImage will trigger a "docker run ..." invocation of the produced image so the user
	// can see if it operates as he would expect
	RunImage bool `json:"runImage,omitempty"`

	// Usage allows for properly shortcircuiting s2i logic when `s2i usage` is invoked
	Usage bool `json:"usage,omitempty"`

	// Injections specifies a list source/destination folders that are injected to
	// the container that runs assemble.
	// All files we inject will be truncated after the assemble script finishes.
	Injections []VolumeSpec `json:"injections,omitempty"`

	// CGroupLimits describes the cgroups limits that will be applied to any containers
	// run by s2i.
	CGroupLimits *CGroupLimits `json:"cgroupLimits,omitempty"`

	// DropCapabilities contains a list of capabilities to drop when executing containers
	DropCapabilities []string `json:"dropCapabilities,omitempty"`

	// ScriptDownloadProxyConfig optionally specifies the http and https proxy
	// to use when downloading scripts
	ScriptDownloadProxyConfig *ProxyConfig `json:"scriptDownloadProxyConfig,omitempty"`

	// ExcludeRegExp contains a string representation of the regular expression desired for
	// deciding which files to exclude from the tar stream
	ExcludeRegExp string `json:"excludeRegExp,omitempty"`

	// BlockOnBuild prevents s2i from performing a docker build operation
	// if one is necessary to execute ONBUILD commands, or to layer source code into
	// the container for images that don't have a tar binary available, if the
	// image contains ONBUILD commands that would be executed.
	BlockOnBuild bool `json:"blockOnBuild,omitempty"`

	// HasOnBuild will be set to true if the builder image contains ONBUILD instructions
	HasOnBuild bool `json:"hasOnBuild,omitempty"`

	// BuildVolumes specifies a list of volumes to mount to container running the
	// build.
	BuildVolumes []string `json:"buildVolumes,omitempty"`

	// Labels specify labels and their values to be applied to the resulting image. Label keys
	// must have non-zero length. The labels defined here override generated labels in case
	// they have the same name.
	Labels map[string]string `json:"labels,omitempty"`

	// SecurityOpt are passed as options to the docker containers launched by s2i.
	SecurityOpt []string `json:"securityOpt,omitempty"`

	// KeepSymlinks indicates to copy symlinks as symlinks. Default behavior is to follow
	// symlinks and copy files by content.
	KeepSymlinks bool `json:"keepSymlinks,omitempty"`

	// AsDockerfile indicates the path where the Dockerfile should be written instead of building
	// a new image.
	AsDockerfile string `json:"asDockerfile,omitempty"`

	// ImageWorkDir is the default working directory for the builder image.
	ImageWorkDir string `json:"imageWorkDir,omitempty"`

	// ImageScriptsURL is the default location to find the assemble/run scripts for a builder image.
	// This url can be a reference within the builder image if the scheme is specified as image://
	ImageScriptsURL string `json:"imageScriptsUrl,omitempty"`

	// AddHost Add a line to /etc/hosts for test purpose or private use in LAN. Its format is host:IP,multiple hosts can be added  by using multiple --add-host
	AddHost []string `json:"addHost,omitempty"`

	// Export Push the result image to specify image registry in tag
	Export bool `json:"export,omitempty"`

	// SourceURL is  url of the codes such as https://github.com/a/b.git
	SourceURL string `json:"sourceUrl"`

	// IsBinaryURL explain the type of SourceURL.
	// If it is IsBinaryURL, it will download the file directly without using git.
	IsBinaryURL bool `json:"isBinaryURL,omitempty"`

	// GitSecretRef is the BasicAuth Secret of Git Clone
	GitSecretRef *corev1.LocalObjectReference `json:"gitSecretRef,omitempty"`

	// The RevisionId is a branch name or a SHA-1 hash of every important thing about the commit
	RevisionId string `json:"revisionId,omitempty"`

	// The name of taint.
	TaintKey string `json:"taintKey,omitempty"`

	// The key of Node Affinity.
	NodeAffinityKey string `json:"nodeAffinityKey,omitempty"`

	// The values of Node Affinity.
	NodeAffinityValues []string `json:"nodeAffinityValues,omitempty"`

	// Whether output build result to status.
	OutputBuildResult bool `json:"outputBuildResult,omitempty"`

	// Regular expressions, ignoring names that do not match the provided regular expression
	BranchExpression string `json:"branchExpression,omitempty"`

	// SecretCode
	SecretCode string `json:"secretCode,omitempty"`
}

type UserDefineTemplate struct {
	//Name specify a template to use, so many fields in Config can left empty
	Name string `json:"name,omitempty"`
	//Parameters must use with `template`, fill some parameters which template will use
	Parameters []Parameter `json:"parameters,omitempty"`
	//BaseImage specify which version of this template to use
	BuilderImage string `json:"builderImage,omitempty"`
}

// S2iBuilderSpec defines the desired state of S2iBuilder
type S2iBuilderSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Config *S2iConfig `json:"config,omitempty"`
	//FromTemplate define some inputs from user
	FromTemplate *UserDefineTemplate `json:"fromTemplate,omitempty"`
}

// S2iBuilderStatus defines the observed state of S2iBuilder
type S2iBuilderStatus struct {
	//RunCount represent the sum of s2irun of this builder
	RunCount int `json:"runCount"`
	//LastRunState return the state of the newest run of this builder
	LastRunState RunState `json:"lastRunState,omitempty"`
	//LastRunState return the name of the newest run of this builder
	LastRunName *string `json:"lastRunName,omitempty"`
	//LastRunStartTime return the startTime of the newest run of this builder
	LastRunStartTime *metav1.Time `json:"lastRunStartTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// S2iBuilder is the Schema for the s2ibuilders API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="RunCount",type="integer",JSONPath=".status.runCount"
// +kubebuilder:printcolumn:name="LastRunState",type="string",JSONPath=".status.lastRunState"
// +kubebuilder:printcolumn:name="LastRunName",type="string",JSONPath=".status.lastRunName"
// +kubebuilder:printcolumn:name="LastRunStartTime",type="date",JSONPath=".status.lastRunStartTime"
// +kubebuilder:resource:shortName=s2ib
type S2iBuilder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   S2iBuilderSpec   `json:"spec,omitempty"`
	Status S2iBuilderStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// S2iBuilderList contains a list of S2iBuilder
type S2iBuilderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []S2iBuilder `json:"items"`
}

type S2iAutoScale struct {
	Kind         string   `json:"kind"`
	Name         string   `json:"name"`
	InitReplicas *int32   `json:"initReplicas,omitempty"`
	Containers   []string `json:"containers,omitempty"`
}

type DockerConfigJson struct {
	Auths DockerConfigMap `json:"auths"`
}

// DockerConfig represents the config file used by the docker CLI.
// This config that represents the credentials that should be used
// when pulling images from specific image repositories.
type DockerConfigMap map[string]DockerConfigEntry

type DockerConfigEntry struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	Email         string `json:"email"`
	ServerAddress string `json:"serverAddress,omitempty"`
}

func init() {
	SchemeBuilder.Register(&S2iBuilder{}, &S2iBuilderList{})
}
