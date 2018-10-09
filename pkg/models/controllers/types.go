/*
Copyright 2018 The KubeSphere Authors.

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

package controllers

import (
	"time"

	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/jinzhu/gorm"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	appV1 "k8s.io/client-go/listers/apps/v1"
	batchv1 "k8s.io/client-go/listers/batch/v1"
	batchv1beta1 "k8s.io/client-go/listers/batch/v1beta1"
	coreV1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/listers/extensions/v1beta1"
	rbacV1 "k8s.io/client-go/listers/rbac/v1"
	storageV1 "k8s.io/client-go/listers/storage/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	resyncCircle = 600
	Stopped      = "stopped"
	PvcPending   = "pending"
	Running      = "running"
	Updating     = "updating"
	Failed       = "failed"
	Unfinished   = "unfinished"
	Completed    = "completed"
	Pause        = "pause"
	Warning      = "warning"
	Error        = "error"
	DisplayName  = "displayName"

	Pods                  = "pods"
	Deployments           = "deployments"
	Daemonsets            = "daemonsets"
	Statefulsets          = "statefulsets"
	Namespaces            = "namespaces"
	Ingresses             = "ingresses"
	PersistentVolumeClaim = "persistent-volume-claims"
	Roles                 = "roles"
	ClusterRoles          = "cluster-roles"
	Services              = "services"
	StorageClasses        = "storage-classes"
	Applications          = "applications"
	Jobs                  = "jobs"
	Cronjobs              = "cronjobs"
	Nodes                 = "nodes"
	Replicasets           = "replicasets"
	ControllerRevisions   = "controllerrevisions"
	ConfigMaps            = "configmaps"
	Secrets               = "secrets"
)

type MapString struct {
	Values map[string]string `json:"values" gorm:"type:TEXT"`
}

func (annotation *MapString) Scan(val interface{}) error {
	switch val := val.(type) {
	case string:
		return json.Unmarshal([]byte(val), annotation)
	case []byte:
		return json.Unmarshal(val, annotation)
	default:
		return errors.New("not support")
	}
	return nil
}

func (annotation MapString) Value() (driver.Value, error) {
	bytes, err := json.Marshal(annotation)
	return string(bytes), err
}

type Taints struct {
	Values []v1.Taint `json:"values" gorm:"type:TEXT"`
}

func (taints *Taints) Scan(val interface{}) error {
	switch val := val.(type) {
	case string:
		return json.Unmarshal([]byte(val), taints)
	case []byte:
		return json.Unmarshal(val, taints)
	default:
		return errors.New("not support")
	}
	return nil
}

func (taints Taints) Value() (driver.Value, error) {
	bytes, err := json.Marshal(taints)
	return string(bytes), err
}

type Deployment struct {
	Name        string `gorm:"primary_key" json:"name"`
	DisplayName string `json:"displayName,omitempty" gorm:"column:displayName"`
	Namespace   string `gorm:"primary_key" json:"namespace"`
	App         string `json:"app,omitempty"`

	Available  int32     `json:"available"`
	Desire     int32     `json:"desire"`
	Status     string    `json:"status"`
	Labels     MapString `json:"labels"`
	Annotation MapString `json:"annotations"`
	UpdateTime time.Time `gorm:"column:updateTime" json:"updateTime,omitempty"`
}

type Statefulset struct {
	Name        string `gorm:"primary_key" json:"name,omitempty"`
	DisplayName string `json:"displayName,omitempty" gorm:"column:displayName"`
	Namespace   string `gorm:"primary_key" json:"namespace,omitempty"`
	App         string `json:"app,omitempty"`

	Available  int32     `json:"available"`
	Desire     int32     `json:"desire"`
	Status     string    `json:"status"`
	Annotation MapString `json:"annotations"`
	Labels     MapString `json:"labels"`
	CreateTime time.Time `gorm:"column:createTime" json:"createTime,omitempty"`
}

type Daemonset struct {
	Name        string `gorm:"primary_key" json:"name,omitempty"`
	DisplayName string `json:"displayName,omitempty" gorm:"column:displayName"`
	Namespace   string `gorm:"primary_key" json:"namespace,omitempty"`
	App         string `json:"app,omitempty"`

	Available    int32     `json:"available"`
	Desire       int32     `json:"desire"`
	Status       string    `json:"status"`
	NodeSelector string    `json:"nodeSelector, omitempty"`
	Annotation   MapString `json:"annotations"`
	Labels       MapString `json:"labels"`
	CreateTime   time.Time `gorm:"column:createTime" json:"createTime,omitempty"`
}

type Service struct {
	Name        string `gorm:"primary_key" json:"name"`
	DisplayName string `json:"displayName,omitempty" gorm:"column:displayName"`
	Namespace   string `gorm:"primary_key" json:"namespace"`
	ServiceType string `gorm:"column:type" json:"type,omitempty"`

	App        string `json:"app,omitempty"`
	VirtualIp  string `gorm:"column:virtualIp" json:"virtualIp,omitempty"`
	ExternalIp string `gorm:"column:externalIp" json:"externalIp,omitempty"`

	Ports      string    `json:"ports,omitempty"`
	NodePorts  string    `gorm:"column:nodePorts" json:"nodePorts,omitempty"`
	Annotation MapString `json:"annotations"`
	Labels     MapString `json:"labels"`
	CreateTime time.Time `gorm:"column:createTime" json:"createTime,omitempty"`
}

type Pvc struct {
	Name             string    `gorm:"primary_key" json:"name"`
	DisplayName      string    `json:"displayName,omitempty" gorm:"column:displayName"`
	Namespace        string    `gorm:"primary_key" json:"namespace"`
	Status           string    `json:"status,omitempty"`
	Capacity         string    `json:"capacity,omitempty"`
	AccessMode       string    `gorm:"column:accessMode" json:"accessMode,omitempty"`
	Annotation       MapString `json:"annotations"`
	Labels           MapString `json:"labels"`
	CreateTime       time.Time `gorm:"column:createTime" json:"createTime,omitempty"`
	StorageClassName string    `gorm:"column:storage_class" json:"storage_class,omitempty"`
	InUse            bool      `gorm:"column:inUse" json:"inUse"`
}

type ingressRule struct {
	Host    string `json:"host"`
	Path    string `json:"path"`
	Service string `json:"service"`
	Port    int32  `json:"port"`
}

type Ingress struct {
	Name           string    `gorm:"primary_key" json:"name"`
	DisplayName    string    `json:"displayName,omitempty" gorm:"column:displayName"`
	Namespace      string    `gorm:"primary_key" json:"namespace"`
	Ip             string    `json:"ip,omitempty"`
	Rules          string    `gorm:"type:text" json:"rules, omitempty"`
	TlsTermination string    `gorm:"column:tlsTermination" json:"tlsTermination,omitempty"`
	Annotation     MapString `json:"annotations"`
	Labels         MapString `json:"labels"`
	CreateTime     time.Time `gorm:"column:createTime" json:"createTime,omitempty"`
}

type Pod struct {
	Name         string     `gorm:"primary_key" json:"name"`
	Namespace    string     `gorm:"primary_key" json:"namespace"`
	Status       string     `json:"status,omitempty"`
	Node         string     `json:"node,omitempty"`
	NodeIp       string     `gorm:"column:nodeIp" json:"nodeIp,omitempty"`
	PodIp        string     `gorm:"column:podIp" json:"podIp,omitempty"`
	Containers   Containers `gorm:"type:text" json:"containers,omitempty"`
	Annotation   MapString  `json:"annotations"`
	Labels       MapString  `json:"labels"`
	RestartCount int        `json:"restartCount"`
	CreateTime   time.Time  `gorm:"column:createTime" json:"createTime,omitempty"`
}

type Container struct {
	Name      string                  `json:"name"`
	Ready     bool                    `json:"ready,omitempty"`
	Image     string                  `json:"image"`
	Resources v1.ResourceRequirements `json:"resources"`
	Ports     []v1.ContainerPort      `json:"ports"`
}
type Containers []Container

func (containers *Containers) Scan(val interface{}) error {
	switch val := val.(type) {
	case string:
		return json.Unmarshal([]byte(val), containers)
	case []byte:
		return json.Unmarshal(val, containers)
	default:
		return errors.New("not support")
	}
	return nil
}

func (containers Containers) Value() (driver.Value, error) {
	bytes, err := json.Marshal(containers)
	return string(bytes), err
}

type Role struct {
	Name        string    `gorm:"primary_key" json:"name"`
	DisplayName string    `json:"displayName,omitempty" gorm:"column:displayName"`
	Namespace   string    `gorm:"primary_key" json:"namespace"`
	Annotation  MapString `json:"annotations"`
	CreateTime  time.Time `gorm:"column:createTime" json:"createTime,omitempty"`
}

type ClusterRole struct {
	Name        string    `gorm:"primary_key" json:"name"`
	DisplayName string    `json:"displayName,omitempty" gorm:"column:displayName"`
	Annotation  MapString `json:"annotations"`
	CreateTime  time.Time `gorm:"column:createTime" json:"createTime,omitempty"`
}

type Namespace struct {
	Name        string `gorm:"primary_key" json:"name"`
	DisplayName string `json:"displayName,omitempty" gorm:"column:displayName"`
	Creator     string `json:"creator,omitempty"`
	Status      string `json:"status"`

	Descrition string          `json:"description,omitempty"`
	Annotation MapString       `json:"annotations"`
	CreateTime time.Time       `gorm:"column:createTime" json:"createTime,omitempty"`
	Usage      v1.ResourceList `gorm:"-" json:"usage,omitempty"`
}

type StorageClass struct {
	Name        string    `gorm:"primary_key" json:"name"`
	DisplayName string    `json:"displayName,omitempty" gorm:"column:displayName"`
	Creator     string    `json:"creator,omitempty"`
	Annotation  MapString `json:"annotations"`
	CreateTime  time.Time `gorm:"column:createTime" json:"createTime,omitempty"`
	IsDefault   bool      `json:"default"`
	Count       int       `json:"count"`
	Provisioner string    `json:"provisioner"`
}

type JobRevisions map[int]JobStatus

type JobStatus struct {
	Status         string    `json:"status"`
	Reasons        []string  `json:"reasons"`
	Messages       []string  `json:"messages"`
	Succeed        int32     `json:"succeed"`
	DesirePodNum   int32     `json:"desire"`
	Failed         int32     `json:"failed"`
	StartTime      time.Time `json:"start-time"`
	CompletionTime time.Time `json:"completion-time"`
}

type Job struct {
	Name        string `gorm:"primary_key" json:"name,omitempty"`
	DisplayName string `json:"displayName,omitempty" gorm:"column:displayName"`
	Namespace   string `gorm:"primary_key" json:"namespace,omitempty"`

	Completed  int32     `json:"completed"`
	Desire     int32     `json:"desire"`
	Status     string    `json:"status"`
	Annotation MapString `json:"annotations"`
	Labels     MapString `json:"labels"`
	CreateTime time.Time `gorm:"column:createTime" json:"createTime,omitempty"`
	UpdateTime time.Time `gorm:"column:updateTime" json:"updateTime,omitempty"`
}

type CronJob struct {
	Name        string `gorm:"primary_key" json:"name,omitempty"`
	DisplayName string `json:"displayName,omitempty" gorm:"column:displayName"`
	Namespace   string `gorm:"primary_key" json:"namespace,omitempty"`

	Active           int        `json:"active"`
	Schedule         string     `json:"schedule"`
	Status           string     `json:"status"`
	Annotation       MapString  `json:"annotations"`
	Labels           MapString  `json:"labels"`
	LastScheduleTime *time.Time `gorm:"column:lastScheduleTime" json:"lastScheduleTime,omitempty"`
}

type Node struct {
	Name        string    `gorm:"primary_key" json:"name,omitempty"`
	DisplayName string    `json:"displayName,omitempty" gorm:"column:displayName"`
	Ip          string    `json:"ip"`
	Status      string    `json:"status"`
	Annotation  MapString `json:"annotations"`
	Labels      MapString `json:"labels"`
	Taints      Taints    `json:"taints"`
	Msg         string    `json:"msg"`
	Role        string    `json:"role"`
	CreateTime  time.Time `gorm:"column:createTime" json:"createTime,omitempty"`
}

type ConfigMap struct {
	Name        string    `gorm:"primary_key" json:"name"`
	Namespace   string    `gorm:"primary_key" json:"namespace"`
	DisplayName string    `json:"displayName,omitempty" gorm:"column:displayName"`
	CreateTime  time.Time `gorm:"column:createTime" json:"createTime,omitempty"`
	Annotation  MapString `json:"annotations"`
	Entries     string    `gorm:"type:text" json:"entries"`
}

type Secret struct {
	Name        string    `gorm:"primary_key" json:"name"`
	Namespace   string    `gorm:"primary_key" json:"namespace"`
	DisplayName string    `json:"displayName,omitempty" gorm:"column:displayName"`
	CreateTime  time.Time `gorm:"column:createTime" json:"createTime,omitempty"`
	Annotation  MapString `json:"annotations"`
	Entries     int       `json:"entries"`
	Type        string    `json:"type"`
}

type Paging struct {
	Limit, Offset, Page int
}

type Controller interface {
	chanStop() chan struct{}
	chanAlive() chan struct{}
	CountWithConditions(condition string) int
	total() int
	initListerAndInformer()
	sync(stopChan chan struct{})
	Name() string
	CloseDB()
	Lister() interface{}
	ListWithConditions(condition string, paging *Paging, order string) (int, interface{}, error)
}

type CommonAttribute struct {
	K8sClient *kubernetes.Clientset
	Name      string
	DB        *gorm.DB
	stopChan  chan struct{}
	aliveChan chan struct{}
}

func (ca *CommonAttribute) chanStop() chan struct{} {

	return ca.stopChan
}

func (ca *CommonAttribute) chanAlive() chan struct{} {

	return ca.aliveChan
}

func (ca *CommonAttribute) CloseDB() {

	ca.DB.Close()
}

type DeploymentCtl struct {
	CommonAttribute
	lister   appV1.DeploymentLister
	informer cache.SharedIndexInformer
}

type StatefulsetCtl struct {
	CommonAttribute
	lister   appV1.StatefulSetLister
	informer cache.SharedIndexInformer
}

type DaemonsetCtl struct {
	CommonAttribute
	lister   appV1.DaemonSetLister
	informer cache.SharedIndexInformer
}

type ServiceCtl struct {
	CommonAttribute
	lister   coreV1.ServiceLister
	informer cache.SharedIndexInformer
}

type PvcCtl struct {
	CommonAttribute
	lister   coreV1.PersistentVolumeClaimLister
	informer cache.SharedIndexInformer
}

type PodCtl struct {
	CommonAttribute
	lister   coreV1.PodLister
	informer cache.SharedIndexInformer
}

type IngressCtl struct {
	lister   v1beta1.IngressLister
	informer cache.SharedIndexInformer
	CommonAttribute
}

type NamespaceCtl struct {
	CommonAttribute
	lister   coreV1.NamespaceLister
	informer cache.SharedIndexInformer
}

type StorageClassCtl struct {
	lister   storageV1.StorageClassLister
	informer cache.SharedIndexInformer
	CommonAttribute
}

type RoleCtl struct {
	lister   rbacV1.RoleLister
	informer cache.SharedIndexInformer
	CommonAttribute
}

type ClusterRoleCtl struct {
	lister   rbacV1.ClusterRoleLister
	informer cache.SharedIndexInformer
	CommonAttribute
}

type JobCtl struct {
	lister   batchv1.JobLister
	informer cache.SharedIndexInformer
	CommonAttribute
}

type CronJobCtl struct {
	lister   batchv1beta1.CronJobLister
	informer cache.SharedIndexInformer
	CommonAttribute
}

type NodeCtl struct {
	lister   coreV1.NodeLister
	informer cache.SharedIndexInformer
	CommonAttribute
}

type ReplicaSetCtl struct {
	lister   appV1.ReplicaSetLister
	informer cache.SharedIndexInformer
	CommonAttribute
}

type ControllerRevisionCtl struct {
	lister   appV1.ControllerRevisionLister
	informer cache.SharedIndexInformer
	CommonAttribute
}

type ConfigMapCtl struct {
	lister   coreV1.ConfigMapLister
	informer cache.SharedIndexInformer
	CommonAttribute
}

type SecretCtl struct {
	lister   coreV1.SecretLister
	informer cache.SharedIndexInformer
	CommonAttribute
}
