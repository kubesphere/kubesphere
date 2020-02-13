package models

import (
	osappsv1 "github.com/openshift/api/apps/v1"
	appsv1 "k8s.io/api/apps/v1"
	batch_v1 "k8s.io/api/batch/v1"
	batch_v1beta1 "k8s.io/api/batch/v1beta1"
	"k8s.io/api/core/v1"

	"github.com/kiali/kiali/config"
	"github.com/kiali/kiali/prometheus"
)

type WorkloadList struct {
	// Namespace where the workloads live in
	// required: true
	// example: bookinfo
	Namespace Namespace `json:"namespace"`

	// Workloads for a given namespace
	// required: true
	Workloads []WorkloadListItem `json:"workloads"`
}

// WorkloadListItem has the necessary information to display the console workload list
type WorkloadListItem struct {
	// Name of the workload
	// required: true
	// example: reviews-v1
	Name string `json:"name"`

	// Type of the workload
	// required: true
	// example: deployment
	Type string `json:"type"`

	// Creation timestamp (in RFC3339 format)
	// required: true
	// example: 2018-07-31T12:24:17Z
	CreatedAt string `json:"createdAt"`

	// Kubernetes ResourceVersion
	// required: true
	// example: 192892127
	ResourceVersion string `json:"resourceVersion"`

	// Define if Pods related to this Workload has an IstioSidecar deployed
	// required: true
	// example: true
	IstioSidecar bool `json:"istioSidecar"`

	// Workload labels
	Labels map[string]string `json:"labels"`

	// Define if Pods related to this Workload has the label App
	// required: true
	// example: true
	AppLabel bool `json:"appLabel"`

	// Define if Pods related to this Workload has the label Version
	// required: true
	// example: true
	VersionLabel bool `json:"versionLabel"`

	// Number of current workload pods
	// required: true
	// example: 1
	PodCount int `json:"podCount"`
}

type WorkloadOverviews []*WorkloadListItem

// Workload has the details of a workload
type Workload struct {
	WorkloadListItem

	// Number of desired replicas
	// required: true
	// example: 2
	Replicas int32 `json:"replicas"`

	// Number of available replicas
	// required: true
	// example: 1
	AvailableReplicas int32 `json:"availableReplicas"`

	// Number of unavailable replicas
	// required: true
	// example: 1
	UnavailableReplicas int32 `json:"unavailableReplicas"`

	// Pods bound to the workload
	Pods Pods `json:"pods"`

	// Services that match workload selector
	Services Services `json:"services"`

	DestinationServices []DestinationService `json:"destinationServices"`

	// Runtimes and associated dashboards
	Runtimes []Runtime `json:"runtimes"`
}

// DestinationService holds service identifiers used for workload dependencies
type DestinationService struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type Workloads []*Workload

func (workload *WorkloadListItem) ParseWorkload(w *Workload) {
	conf := config.Get()
	workload.Name = w.Name
	workload.Type = w.Type
	workload.CreatedAt = w.CreatedAt
	workload.ResourceVersion = w.ResourceVersion
	workload.IstioSidecar = w.Pods.HasIstioSideCar()
	workload.Labels = w.Labels
	workload.PodCount = len(w.Pods)

	/** Check the labels app and version required by Istio in template Pods*/
	_, workload.AppLabel = w.Labels[conf.IstioLabels.AppLabelName]
	_, workload.VersionLabel = w.Labels[conf.IstioLabels.VersionLabelName]
}

func (workload *Workload) ParseDeployment(d *appsv1.Deployment) {
	conf := config.Get()
	workload.Name = d.Name
	workload.Type = "Deployment"
	workload.Labels = d.Spec.Template.Labels

	/** Check the labels app and version required by Istio in template Pods*/
	_, workload.AppLabel = workload.Labels[conf.IstioLabels.AppLabelName]
	_, workload.VersionLabel = workload.Labels[conf.IstioLabels.VersionLabelName]

	workload.CreatedAt = formatTime(d.CreationTimestamp.Time)
	workload.ResourceVersion = d.ResourceVersion
	workload.Replicas = d.Status.Replicas
	workload.AvailableReplicas = d.Status.AvailableReplicas
	// Deployments/ReplicaSets have a different parameters to indicate unavailable
	// calculate "desired" - "current" sounds reasonable on this context
	workload.UnavailableReplicas = workload.Replicas - workload.AvailableReplicas
}

func (workload *Workload) ParseReplicaSet(r *appsv1.ReplicaSet) {
	conf := config.Get()
	workload.Name = r.Name
	workload.Type = "ReplicaSet"
	workload.Labels = r.Spec.Template.Labels

	/** Check the labels app and version required by Istio in template Pods*/
	_, workload.AppLabel = workload.Labels[conf.IstioLabels.AppLabelName]
	_, workload.VersionLabel = workload.Labels[conf.IstioLabels.VersionLabelName]

	workload.CreatedAt = formatTime(r.CreationTimestamp.Time)
	workload.ResourceVersion = r.ResourceVersion
	workload.Replicas = r.Status.Replicas
	workload.AvailableReplicas = r.Status.AvailableReplicas
	// Deployments/ReplicaSets have a different parameters to indicate unavailable
	// calculate "desired" - "current" sounds reasonable on this context
	workload.UnavailableReplicas = workload.Replicas - workload.AvailableReplicas
}

func (workload *Workload) ParseReplicationController(r *v1.ReplicationController) {
	conf := config.Get()
	workload.Name = r.Name
	workload.Type = "ReplicationController"
	workload.Labels = r.Spec.Template.Labels

	/** Check the labels app and version required by Istio in template Pods*/
	_, workload.AppLabel = workload.Labels[conf.IstioLabels.AppLabelName]
	_, workload.VersionLabel = workload.Labels[conf.IstioLabels.VersionLabelName]

	workload.CreatedAt = formatTime(r.CreationTimestamp.Time)
	workload.ResourceVersion = r.ResourceVersion
	workload.Replicas = r.Status.Replicas
	workload.AvailableReplicas = r.Status.AvailableReplicas
	// Deployments/ReplicaSets have a different parameters to indicate unavailable
	// calculate "desired" - "current" sounds reasonable on this context
	workload.UnavailableReplicas = workload.Replicas - workload.AvailableReplicas
}

func (workload *Workload) ParseDeploymentConfig(dc *osappsv1.DeploymentConfig) {
	workload.Name = dc.Name
	workload.Type = "DeploymentConfig"
	workload.Labels = dc.Spec.Template.Labels
	workload.CreatedAt = formatTime(dc.CreationTimestamp.Time)
	workload.ResourceVersion = dc.ResourceVersion
	workload.Replicas = dc.Status.Replicas
	workload.AvailableReplicas = dc.Status.AvailableReplicas
	// Deployments/ReplicaSets have a different parameters to indicate unavailable
	// calculate "desired" - "current" sounds reasonable on this context
	workload.UnavailableReplicas = workload.Replicas - workload.AvailableReplicas
}

func (workload *Workload) ParseStatefulSet(s *appsv1.StatefulSet) {
	conf := config.Get()
	workload.Name = s.Name
	workload.Type = "StatefulSet"
	workload.Labels = s.Spec.Template.Labels

	/** Check the labels app and version required by Istio in template Pods*/
	_, workload.AppLabel = workload.Labels[conf.IstioLabels.AppLabelName]
	_, workload.VersionLabel = workload.Labels[conf.IstioLabels.VersionLabelName]

	workload.CreatedAt = formatTime(s.CreationTimestamp.Time)
	workload.ResourceVersion = s.ResourceVersion
	workload.Replicas = s.Status.Replicas
	workload.AvailableReplicas = s.Status.ReadyReplicas
	// Deployments/ReplicaSets have a different parameters to indicate unavailable
	// calculate "desired" - "current" sounds reasonable on this context
	workload.UnavailableReplicas = workload.Replicas - workload.AvailableReplicas
}

func (workload *Workload) ParsePod(pod *v1.Pod) {
	conf := config.Get()
	workload.Name = pod.Name
	workload.Type = "Pod"
	workload.Labels = pod.Labels

	/** Check the labels app and version required by Istio in template Pods*/
	_, workload.AppLabel = workload.Labels[conf.IstioLabels.AppLabelName]
	_, workload.VersionLabel = workload.Labels[conf.IstioLabels.VersionLabelName]

	workload.CreatedAt = formatTime(pod.CreationTimestamp.Time)
	workload.ResourceVersion = pod.ResourceVersion

	var podReplicas, podAvailableReplicas int32
	podReplicas = 1
	podAvailableReplicas = 1

	// When a Workload is a single pod we don't have access to any controller replicas
	// On this case we differentiate when pod is terminated with success versus not running
	// Probably it might be more cases to refine here
	if pod.Status.Phase == "Succeed" {
		podReplicas = 0
		podAvailableReplicas = 0
	} else if pod.Status.Phase != "Running" {
		podAvailableReplicas = 0
	}

	workload.Replicas = podReplicas
	workload.AvailableReplicas = podAvailableReplicas
	// Deployments/ReplicaSets have a different parameters to indicate unavailable
	// calculate "desired" - "current" sounds reasonable on this context
	workload.UnavailableReplicas = workload.Replicas - workload.AvailableReplicas
}

func (workload *Workload) ParseJob(job *batch_v1.Job) {
	conf := config.Get()
	workload.Name = job.Name
	workload.Type = "Job"
	workload.Labels = job.Labels

	/** Check the labels app and version required by Istio in template Pods*/
	_, workload.AppLabel = workload.Labels[conf.IstioLabels.AppLabelName]
	_, workload.VersionLabel = workload.Labels[conf.IstioLabels.VersionLabelName]

	workload.CreatedAt = formatTime(job.CreationTimestamp.Time)
	workload.ResourceVersion = job.ResourceVersion

	workload.Replicas = job.Status.Active + job.Status.Succeeded + job.Status.Failed
	workload.AvailableReplicas = job.Status.Active + job.Status.Succeeded

	// Deployments/ReplicaSets have a different parameters to indicate unavailable
	// calculate "desired" - "current" sounds reasonable on this context
	workload.UnavailableReplicas = job.Status.Failed
}

func (workload *Workload) ParseCronJob(cnjb *batch_v1beta1.CronJob) {
	conf := config.Get()
	workload.Name = cnjb.Name
	workload.Type = "CronJob"
	workload.Labels = cnjb.Labels

	/** Check the labels app and version required by Istio in template Pods*/
	_, workload.AppLabel = workload.Labels[conf.IstioLabels.AppLabelName]
	_, workload.VersionLabel = workload.Labels[conf.IstioLabels.VersionLabelName]

	workload.CreatedAt = formatTime(cnjb.CreationTimestamp.Time)
	workload.ResourceVersion = cnjb.ResourceVersion

	// We don't have the information of this controller
	// We will infer the number of replicas as the number of pods without succeed state
	// We will infer the number of available as the number of pods with running state
	// If this is not enough, we should try to fetch the controller, it is not doing now to not overload kiali fetching all types of controllers
	var podReplicas, podAvailableReplicas int32
	podReplicas = 0
	podAvailableReplicas = 0
	for _, pod := range workload.Pods {
		if pod.Status != "Succeeded" {
			podReplicas++
		}
		if pod.Status == "Running" {
			podAvailableReplicas++
		}
	}
	workload.Replicas = podReplicas
	workload.AvailableReplicas = podAvailableReplicas
	// Deployments/ReplicaSets have a different parameters to indicate unavailable
	// calculate "desired" - "current" sounds reasonable on this context
	if podReplicas > podAvailableReplicas {
		workload.UnavailableReplicas = workload.Replicas - workload.AvailableReplicas
	} else {
		// On this case a Job may have all pods terminated
		// Then it is not an unhealth condition
		workload.UnavailableReplicas = 0
	}
}

func (workload *Workload) ParsePods(controllerName string, controllerType string, pods []v1.Pod) {
	conf := config.Get()
	workload.Name = controllerName
	workload.Type = controllerType
	// We don't have the information of this controller
	// We will infer the number of replicas as the number of pods without succeed state
	// We will infer the number of available as the number of pods with running state
	// If this is not enough, we should try to fetch the controller, it is not doing now to not overload kiali fetching all types of controllers
	var podReplicas, podAvailableReplicas int32
	podReplicas = 0
	podAvailableReplicas = 0
	for _, pod := range pods {
		if pod.Status.Phase != "Succeeded" {
			podReplicas++
		}
		if pod.Status.Phase == "Running" {
			podAvailableReplicas++
		}
	}
	workload.Replicas = podReplicas
	workload.AvailableReplicas = podAvailableReplicas
	// Deployments/ReplicaSets have a different parameters to indicate unavailable
	// calculate "desired" - "current" sounds reasonable on this context
	if podReplicas > podAvailableReplicas {
		workload.UnavailableReplicas = workload.Replicas - workload.AvailableReplicas
	} else {
		// On this case a Job may have all pods terminated
		// Then it is not an unhealth condition
		workload.UnavailableReplicas = 0
	}
	// We fetch one pod as template for labels
	// There could be corner cases not correct, then we should support more controllers
	if len(pods) > 0 {
		workload.Labels = pods[0].Labels
		workload.CreatedAt = formatTime(pods[0].CreationTimestamp.Time)
		workload.ResourceVersion = pods[0].ResourceVersion
	}

	/** Check the labels app and version required by Istio in template Pods*/
	_, workload.AppLabel = workload.Labels[conf.IstioLabels.AppLabelName]
	_, workload.VersionLabel = workload.Labels[conf.IstioLabels.VersionLabelName]
}

func (workload *Workload) SetPods(pods []v1.Pod) {
	workload.Pods.Parse(pods)
	workload.IstioSidecar = workload.Pods.HasIstioSideCar()
}

func (workload *Workload) SetServices(svcs []v1.Service) {
	workload.Services.Parse(svcs)
}

func (workload *Workload) SetDestinationServices(dss []prometheus.Service) {
	workload.DestinationServices = make([]DestinationService, 0, len(dss))
	for _, service := range dss {
		workload.DestinationServices = append(workload.DestinationServices, DestinationService{
			Name:      service.ServiceName,
			Namespace: service.Namespace,
		})
	}
}
