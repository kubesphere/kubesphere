/*
Copyright 2020 KubeSphere Authors

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

package registries

import (
	"github.com/opencontainers/go-digest"
	"time"
)

// ImageBlobInfo describes the info of an image.
type ImageDetails struct {
	// Status is the status of the image search, such as "failed","succeeded".
	Status        string         `json:"status,omitempty" description:"Status is the status of the image search, such as \"succeeded\"."`
	Message       string         `json:"message,omitempty" description:"Status message."`
	ImageManifest *ImageManifest `json:"imageManifest,omitempty" description:"Retrieve the manifest from the registry identified. Reference: https://docs.docker.com/registry/spec/api/#manifest"`
	ImageBlob     *ImageBlob     `json:"imageBlob,omitempty" description:"Retrieve the blob from the registry identified. Reference: https://docs.docker.com/registry/spec/api/#blob"`
	ImageTag      string         `json:"imageTag,omitempty" description:"image tag."`
	Registry      string         `json:"registry,omitempty" description:"registry domain."`
}

type ImageBlob struct {
	Architecture    string          `json:"architecture,omitempty" description:"The architecture field specifies the CPU architecture, for example amd64 or ppc64le."`
	Config          Config          `json:"config,omitempty" description:"The config field references a configuration object for a container."`
	Container       string          `json:"container,omitempty" description:"Container id."`
	ContainerConfig ContainerConfig `json:"container_config,omitempty" description:"The config data of container."`
	Created         time.Time       `json:"created,omitempty" description:"Create time."`
	DockerVersion   string          `json:"docker_version,omitempty" description:"docker version."`
	History         []History       `json:"history,omitempty" description:"The data of history update."`
	Os              string          `json:"os,omitempty" description:"Operating system."`
	Rootfs          Rootfs          `json:"rootfs omitempty" description:"Root filesystem."`
}

type Labels struct {
	Maintainer string `json:"maintainer" description:""`
}
type Config struct {
	HostName     string                 `json:"Hostname,omitempty" description:"A string value containing the hostname to use for the container."`
	DomainName   string                 `json:"Domainname,omitempty" description:"A string value containing the domain name to use for the container."`
	User         string                 `json:"User,omitempty" description:"A string value specifying the user inside the container."`
	AttachStdin  bool                   `json:"AttachStdin,omitempty" description:"Boolean value, attaches to stdin."`
	AttachStdout bool                   `json:"AttachStdout,omitempty" description:"Boolean value, attaches to stdout."`
	AttachStderr bool                   `json:"AttachStderr,omitempty" description:"Boolean value, attaches to stderr."`
	ExposedPorts map[string]interface{} `json:"ExposedPorts,omitempty" description:"An object mapping ports to an empty object in the form of: \"ExposedPorts\": { \"<port>/<tcp|udp>: {}\" }"`
	Tty          bool                   `json:"Tty,omitempty" description:"Boolean value, Attach standard streams to a tty, including stdin if it is not closed."`
	OpenStdin    bool                   `json:"OpenStdin,omitempty" description:"Boolean value, opens stdin"`
	StdinOnce    bool                   `json:"StdinOnce,omitempty" description:"Boolean value, close stdin after the 1 attached client disconnects."`
	Env          []string               `json:"Env,omitempty" description:"A list of environment variables in the form of [\"VAR=value\", ...]"`
	Cmd          []string               `json:"Cmd,omitempty" description:"Command to run specified as a string or an array of strings."`
	ArgsEscaped  bool                   `json:"ArgsEscaped,omitempty" description:"Command is already escaped (Windows only)"`
	Image        string                 `json:"Image,omitempty" description:"A string specifying the image name to use for the container."`
	Volumes      interface{}            `json:"Volumes,omitempty" description:"An object mapping mount point paths (strings) inside the container to empty objects."`
	WorkingDir   string                 `json:"WorkingDir,omitempty" description:"A string specifying the working directory for commands to run in."`
	Entrypoint   interface{}            `json:"Entrypoint,omitempty" description:"The entry point set for the container as a string or an array of strings."`
	OnBuild      interface{}            `json:"OnBuild,omitempty" description:"ONBUILD metadata that were defined in the image's Dockerfile."`
	Labels       Labels                 `json:"Labels,omitempty" description:"The map of labels to a container."`
	StopSignal   string                 `json:"StopSignal,omitempty" description:"Signal to stop a container as a string or unsigned integer."`
}
type ContainerConfig struct {
	HostName     string                 `json:"Hostname,omitempty" description:"A string value containing the hostname to use for the container."`
	DomainName   string                 `json:"Domainname,omitempty" description:"A string value containing the domain name to use for the container."`
	User         string                 `json:"User,omitempty" description:"A string value specifying the user inside the container."`
	AttachStdin  bool                   `json:"AttachStdin,omitempty" description:"Boolean value, attaches to stdin."`
	AttachStdout bool                   `json:"AttachStdout,omitempty" description:"Boolean value, attaches to stdout."`
	AttachStderr bool                   `json:"AttachStderr,omitempty" description:"Boolean value, attaches to stderr."`
	ExposedPorts map[string]interface{} `json:"ExposedPorts,omitempty" description:"An object mapping ports to an empty object in the form of: \"ExposedPorts\": { \"<port>/<tcp|udp>: {}\" }"`
	Tty          bool                   `json:"Tty,omitempty" description:"Boolean value, Attach standard streams to a tty, including stdin if it is not closed."`
	OpenStdin    bool                   `json:"OpenStdin,omitempty" description:"Boolean value, opens stdin"`
	StdinOnce    bool                   `json:"StdinOnce,omitempty" description:"Boolean value, close stdin after the 1 attached client disconnects."`
	Env          []string               `json:"Env,omitempty" description:"A list of environment variables in the form of [\"VAR=value\", ...]"`
	Cmd          []string               `json:"Cmd,omitempty" description:"Command to run specified as a string or an array of strings."`
	ArgsEscaped  bool                   `json:"ArgsEscaped,omitempty" description:"Command is already escaped (Windows only)"`
	Image        string                 `json:"Image,omitempty" description:"A string specifying the image name to use for the container."`
	Volumes      interface{}            `json:"Volumes,omitempty" description:"An object mapping mount point paths (strings) inside the container to empty objects."`
	WorkingDir   string                 `json:"WorkingDir,omitempty" description:"A string specifying the working directory for commands to run in."`
	EntryPoint   interface{}            `json:"Entrypoint,omitempty" description:"The entry point set for the container as a string or an array of strings."`
	OnBuild      interface{}            `json:"OnBuild,omitempty" description:"ONBUILD metadata that were defined in the image's Dockerfile."`
	Labels       Labels                 `json:"Labels,omitempty" description:"The map of labels to a container."`
	StopSignal   string                 `json:"StopSignal,omitempty" description:"Signal to stop a container as a string or unsigned integer."`
}
type History struct {
	Created    time.Time `json:"created,omitempty" description:"Created time."`
	CreatedBy  string    `json:"created_by,omitempty" description:"Created command."`
	EmptyLayer bool      `json:"empty_layer,omitempty" description:"Layer empty or not."`
}
type Rootfs struct {
	Type    string   `json:"type,omitempty" description:"Root filesystem type, always \"layers\" "`
	DiffIds []string `json:"diff_ids,omitempty" description:"Contain ids of layer list"`
}

type ImageManifest struct {
	SchemaVersion  int            `json:"schemaVersion,omitempty" description:"This field specifies the image manifest schema version as an integer."`
	MediaType      string         `json:"mediaType,omitempty" description:"The MIME type of the manifest."`
	ManifestConfig ManifestConfig `json:"config,omitempty" description:"The config field references a configuration object for a container."`
	Layers         []Layers       `json:"layers,omitempty" description:"Fields of an item in the layers list."`
}
type ManifestConfig struct {
	MediaType string        `json:"mediaType,omitempty" description:"The MIME type of the image."`
	Size      int           `json:"size,omitempty" description:"The size in bytes of the image."`
	Digest    digest.Digest `json:"digest,omitempty" description:"The digest of the content, as defined by the Registry V2 HTTP API Specificiation. Reference https://docs.docker.com/registry/spec/api/#digest-parameter"`
}
type Layers struct {
	MediaType string `json:"mediaType,omitempty" description:"The MIME type of the layer."`
	Size      int    `json:"size,omitempty" description:"The size in bytes of the layer."`
	Digest    string `json:"digest,omitempty" description:"The digest of the content, as defined by the Registry V2 HTTP API Specificiation. Reference https://docs.docker.com/registry/spec/api/#digest-parameter"`
}
