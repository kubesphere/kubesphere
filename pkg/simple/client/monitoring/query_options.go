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

package monitoring

import (
	"fmt"
	"strings"
	"time"
)

type Level int

const (
	LevelCluster = 1 << iota
	LevelNode
	LevelWorkspace
	LevelNamespace
	LevelApplication
	LevelOpenpitrix
	LevelWorkload
	LevelService
	LevelPod
	LevelContainer
	LevelPVC
	LevelComponent
	LevelIngress
)

var MeteringLevelMap = map[string]int{
	"LevelCluster":     LevelCluster,
	"LevelNode":        LevelNode,
	"LevelWorkspace":   LevelWorkspace,
	"LevelNamespace":   LevelNamespace,
	"LevelApplication": LevelApplication,
	"LevelWorkload":    LevelWorkload,
	"LevelService":     LevelService,
	"LevelPod":         LevelPod,
	"LevelContainer":   LevelContainer,
	"LevelPVC":         LevelPVC,
	"LevelComponent":   LevelComponent,
}

type QueryOption interface {
	Apply(*QueryOptions)
}

type Meteroptions struct {
	Start time.Time
	End   time.Time
	Step  time.Duration
}

type QueryOptions struct {
	Level Level

	NamespacedResourcesFilter string
	QueryType                 string
	ResourceFilter            string
	NodeName                  string
	WorkspaceName             string
	NamespaceName             string
	WorkloadKind              string
	WorkloadName              string
	PodName                   string
	ContainerName             string
	StorageClassName          string
	PersistentVolumeClaimName string
	PVCFilter                 string
	ApplicationName           string
	ServiceName               string
	Ingress                   string
	Job                       string
	Duration                  *time.Duration
	MeterOptions              *Meteroptions
}

func NewQueryOptions() *QueryOptions {
	return &QueryOptions{}
}

type ClusterOption struct{}

func (_ ClusterOption) Apply(o *QueryOptions) {
	o.Level = LevelCluster
}

type NodeOption struct {
	ResourceFilter   string
	NodeName         string
	PVCFilter        string
	StorageClassName string
	QueryType        string
}

func (no NodeOption) Apply(o *QueryOptions) {
	o.Level = LevelNode
	o.ResourceFilter = no.ResourceFilter
	o.NodeName = no.NodeName
	o.PVCFilter = no.PVCFilter
	o.StorageClassName = no.StorageClassName
	o.QueryType = no.QueryType
}

type WorkspaceOption struct {
	ResourceFilter   string
	WorkspaceName    string
	PVCFilter        string
	StorageClassName string
}

func (wo WorkspaceOption) Apply(o *QueryOptions) {
	o.Level = LevelWorkspace
	o.ResourceFilter = wo.ResourceFilter
	o.WorkspaceName = wo.WorkspaceName
	o.PVCFilter = wo.PVCFilter
	o.StorageClassName = wo.StorageClassName
}

type NamespaceOption struct {
	ResourceFilter   string
	WorkspaceName    string
	NamespaceName    string
	PVCFilter        string
	StorageClassName string
}

func (no NamespaceOption) Apply(o *QueryOptions) {
	o.Level = LevelNamespace
	o.ResourceFilter = no.ResourceFilter
	o.WorkspaceName = no.WorkspaceName
	o.NamespaceName = no.NamespaceName
	o.PVCFilter = no.PVCFilter
	o.StorageClassName = no.StorageClassName
}

type ApplicationsOption struct {
	NamespaceName    string
	Applications     []string
	StorageClassName string
}

func (aso ApplicationsOption) Apply(o *QueryOptions) {
	// nothing should be done
	//nolint:gosimple
	return
}

type OpenpitrixsOption struct {
	Cluster          string
	NamespaceName    string
	Openpitrixs      []string
	StorageClassName string
}

func (oso OpenpitrixsOption) Apply(o *QueryOptions) {
	// nothing should be done
	//nolint:gosimple
	return
}

// ApplicationsOption & OpenpitrixsOption share the same ApplicationOption struct
type ApplicationOption struct {
	NamespaceName         string
	Application           string
	ApplicationComponents []string
	StorageClassName      string
}

func (ao ApplicationOption) Apply(o *QueryOptions) {
	o.Level = LevelApplication
	o.NamespaceName = ao.NamespaceName
	o.ApplicationName = ao.Application
	o.StorageClassName = ao.StorageClassName

	app_components := strings.Join(ao.ApplicationComponents[:], "|")

	if len(app_components) > 0 {
		o.ResourceFilter = fmt.Sprintf(`namespace="%s", workload=~"%s"`, o.NamespaceName, app_components)
	} else {
		o.ResourceFilter = fmt.Sprintf(`namespace="%s", workload=~"%s"`, o.NamespaceName, ".*")
	}
}

type WorkloadOption struct {
	ResourceFilter string
	NamespaceName  string
	WorkloadKind   string
}

func (wo WorkloadOption) Apply(o *QueryOptions) {
	o.Level = LevelWorkload
	o.ResourceFilter = wo.ResourceFilter
	o.NamespaceName = wo.NamespaceName
	o.WorkloadKind = wo.WorkloadKind
}

type ServicesOption struct {
	NamespaceName string
	Services      []string
}

func (sso ServicesOption) Apply(o *QueryOptions) {
	// nothing should be done
	//nolint:gosimple
	return
}

type ServiceOption struct {
	ResourceFilter string
	NamespaceName  string
	ServiceName    string
	PodNames       []string
}

func (so ServiceOption) Apply(o *QueryOptions) {
	o.Level = LevelService
	o.NamespaceName = so.NamespaceName
	o.ServiceName = so.ServiceName

	pod_names := strings.Join(so.PodNames, "|")

	if len(pod_names) > 0 {
		o.ResourceFilter = fmt.Sprintf(`pod=~"%s", namespace="%s"`, pod_names, o.NamespaceName)
	} else {
		o.ResourceFilter = fmt.Sprintf(`pod=~"%s", namespace="%s"`, ".*", o.NamespaceName)
	}
}

type PodOption struct {
	NamespacedResourcesFilter string
	ResourceFilter            string
	NodeName                  string
	NamespaceName             string
	WorkloadKind              string
	WorkloadName              string
	PodName                   string
}

func (po PodOption) Apply(o *QueryOptions) {
	o.Level = LevelPod
	o.NamespacedResourcesFilter = po.NamespacedResourcesFilter
	o.ResourceFilter = po.ResourceFilter
	o.NodeName = po.NodeName
	o.NamespaceName = po.NamespaceName
	o.WorkloadKind = po.WorkloadKind
	o.WorkloadName = po.WorkloadName
	o.PodName = po.PodName
}

type ContainerOption struct {
	ResourceFilter string
	NamespaceName  string
	PodName        string
	ContainerName  string
}

func (co ContainerOption) Apply(o *QueryOptions) {
	o.Level = LevelContainer
	o.ResourceFilter = co.ResourceFilter
	o.NamespaceName = co.NamespaceName
	o.PodName = co.PodName
	o.ContainerName = co.ContainerName
}

type PVCOption struct {
	ResourceFilter            string
	NamespaceName             string
	StorageClassName          string
	PersistentVolumeClaimName string
}

func (po PVCOption) Apply(o *QueryOptions) {
	o.Level = LevelPVC
	o.ResourceFilter = po.ResourceFilter
	o.NamespaceName = po.NamespaceName
	o.StorageClassName = po.StorageClassName
	o.PersistentVolumeClaimName = po.PersistentVolumeClaimName

	// for meter
	o.PVCFilter = po.PersistentVolumeClaimName
}

type IngressOption struct {
	ResourceFilter string
	NamespaceName  string
	Ingress        string
	Job            string
	Pod            string
	Duration       *time.Duration
}

func (no IngressOption) Apply(o *QueryOptions) {
	o.Level = LevelIngress
	o.ResourceFilter = no.ResourceFilter
	o.NamespaceName = no.NamespaceName
	o.Ingress = no.Ingress
	o.Job = no.Job
	o.PodName = no.Pod
	o.Duration = no.Duration
}

type ComponentOption struct{}

func (_ ComponentOption) Apply(o *QueryOptions) {
	o.Level = LevelComponent
}

type MeterOption struct {
	Start time.Time
	End   time.Time
	Step  time.Duration
}

func (mo MeterOption) Apply(o *QueryOptions) {
	o.MeterOptions = &Meteroptions{
		Start: mo.Start,
		End:   mo.End,
		Step:  mo.Step,
	}
}
