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
	coreV1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/listers/extensions/v1beta1"
	rbacV1 "k8s.io/client-go/listers/rbac/v1"
	storageV1 "k8s.io/client-go/listers/storage/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	resyncCircle               = 600
	Stopped                    = "stopped"
	PvcPending                 = "Pending"
	Running                    = "running"
	Updating                   = "updating"
	tablePods                  = "pods"
	tableDeployments           = "deployments"
	tableDaemonsets            = "daemonsets"
	tableStatefulsets          = "statefulsets"
	tableNamespaces            = "namespaces"
	tableIngresses             = "ingresses"
	tablePersistentVolumeClaim = "pvcs"
	tableRoles                 = "roles"
	tableClusterRoles          = "cluster_roles"
	tableServices              = "services"
	tableStorageClasses        = "storage_classes"

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
)

type Annotation struct {
	Values map[string]string `gorm:"type:TEXT"`
}

func (annotation *Annotation) Scan(val interface{}) error {
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

func (annotation Annotation) Value() (driver.Value, error) {
	bytes, err := json.Marshal(annotation)
	return string(bytes), err
}

type Deployment struct {
	Name      string `gorm:"primary_key" json:"name"`
	Namespace string `gorm:"primary_key" json:"namespace"`
	App       string `json:"app,omitempty"`

	Available  int32      `json:"available"`
	Desire     int32      `json:"desire"`
	Status     string     `json:"status"`
	Annotation Annotation `json:"annotations"`
	UpdateTime time.Time  `gorm:"column:updateTime" json:"updateTime,omitempty"`
}

func (Deployment) TableName() string {
	return tableDeployments
}

type Statefulset struct {
	Name      string `gorm:"primary_key" json:"name,omitempty"`
	Namespace string `gorm:"primary_key" json:"namespace,omitempty"`
	App       string `json:"app,omitempty"`

	Available  int32      `json:"available"`
	Desire     int32      `json:"desire"`
	Status     string     `json:"status"`
	Annotation Annotation `json:"annotations"`
	CreateTime time.Time  `gorm:"column:createTime" json:"createTime,omitempty"`
}

func (Statefulset) TableName() string {
	return tableStatefulsets
}

type Daemonset struct {
	Name      string `gorm:"primary_key" json:"name,omitempty"`
	Namespace string `gorm:"primary_key" json:"namespace,omitempty"`
	App       string `json:"app,omitempty"`

	Available    int32      `json:"available"`
	Desire       int32      `json:"desire"`
	Status       string     `json:"status"`
	NodeSelector string     `json:"nodeSelector, omitempty"`
	Annotation   Annotation `json:"annotations"`
	CreateTime   time.Time  `gorm:"column:createTime" json:"createTime,omitempty"`
}

func (Daemonset) TableName() string {
	return tableDaemonsets
}

type Service struct {
	Name        string `gorm:"primary_key" json:"name"`
	Namespace   string `gorm:"primary_key" json:"namespace"`
	ServiceType string `json:"type,omitempty"`

	VirtualIp  string `json:"virtualIp,omitempty"`
	ExternalIp string `json:"externalIp,omitempty"`

	Ports      string     `json:"ports,omitempty"`
	NodePorts  string     `json:"nodePorts,omitempty"`
	Annotation Annotation `json:"annotations"`
	CreateTime time.Time  `gorm:"column:createTime" json:"createTime,omitempty"`
}

func (Service) TableName() string {
	return tableServices
}

type Pvc struct {
	Name             string     `gorm:"primary_key" json:"name"`
	Namespace        string     `gorm:"primary_key" json:"namespace"`
	Status           string     `json:"status,omitempty"`
	Capacity         string     `json:"capacity,omitempty"`
	AccessMode       string     `json:"accessMode,omitempty"`
	Annotation       Annotation `json:"annotations"`
	CreateTime       time.Time  `gorm:"column:createTime" json:"createTime,omitempty"`
	StorageClassName string     `gorm:"column:storage_class" json:"storage_class,omitempty"`
	InUse            bool       `gorm:"-" json:"inUse"`
}

func (Pvc) TableName() string {
	return tablePersistentVolumeClaim
}

type ingressRule struct {
	Host    string `json:"host"`
	Path    string `json:"path"`
	Service string `json:"service"`
	Port    int32  `json:"port"`
}

type Ingress struct {
	Name           string     `gorm:"primary_key" json:"name"`
	Namespace      string     `gorm:"primary_key" json:"namespace"`
	Ip             string     `json:"ip,omitempty"`
	Rules          string     `gorm:"type:text" json:"rules, omitempty"`
	TlsTermination string     `json:"tlsTermination,omitempty"`
	Annotation     Annotation `json:"annotations"`
	CreateTime     time.Time  `gorm:"column:createTime" json:"createTime,omitempty"`
}

func (Ingress) TableName() string {
	return tableIngresses
}

type Pod struct {
	Name         string     `gorm:"primary_key" json:"name"`
	Namespace    string     `gorm:"primary_key" json:"namespace"`
	Status       string     `json:"status,omitempty"`
	Node         string     `json:"node,omitempty"`
	NodeIp       string     `json:"nodeIp,omitempty"`
	PodIp        string     `json:"podIp,omitempty"`
	Containers   Containers `gorm:"type:text" json:"containers,omitempty"`
	Annotation   Annotation `json:"annotations"`
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

func (Pod) TableName() string {
	return tablePods
}

type Role struct {
	Name       string     `gorm:"primary_key" json:"name"`
	Namespace  string     `gorm:"primary_key" json:"namespace"`
	Annotation Annotation `json:"annotations"`
	CreateTime time.Time  `gorm:"column:createTime" json:"createTime,omitempty"`
}

func (Role) TableName() string {
	return tableRoles
}

type ClusterRole struct {
	Name       string     `gorm:"primary_key" json:"name"`
	Annotation Annotation `json:"annotations"`
	CreateTime time.Time  `gorm:"column:createTime" json:"createTime,omitempty"`
}

func (ClusterRole) TableName() string {
	return tableClusterRoles
}

type Namespace struct {
	Name    string `gorm:"primary_key" json:"name"`
	Creator string `json:"creator,omitempty"`
	Status  string `json:"status"`

	Descrition string          `json:"description,omitempty"`
	Annotation Annotation      `json:"annotations"`
	CreateTime time.Time       `gorm:"column:createTime" json:"createTime,omitempty"`
	Usaeg      v1.ResourceList `gorm:"-" json:"usage,omitempty"`
}

func (Namespace) TableName() string {
	return tableNamespaces
}

type StorageClass struct {
	Name       string     `gorm:"primary_key" json:"name"`
	Creator    string     `json:"creator,omitempty"`
	Annotation Annotation `json:"annotations"`
	CreateTime time.Time  `gorm:"column:createTime" json:"createTime,omitempty"`
	IsDefault  bool       `json:"default"`
	Count      int        `json:"count"`
}

func (StorageClass) TableName() string {
	return tableStorageClasses
}

type Paging struct {
	Limit, Offset int
}

type Controller interface {
	chanStop() chan struct{}
	chanAlive() chan struct{}
	CountWithConditions(condition string) int
	total() int
	initListerAndInformer()
	sync(stopChan chan struct{})
	Name() string
	ListWithConditions(condition string, paging *Paging) (int, interface{}, error)
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
