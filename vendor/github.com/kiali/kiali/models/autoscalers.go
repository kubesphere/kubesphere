package models

import (
	"k8s.io/api/autoscaling/v1"
)

type Autoscaler struct {
	Name      string            `json:"name"`
	Labels    map[string]string `json:"labels"`
	CreatedAt string            `json:"createdAt"`
	// Spec
	MinReplicas                    int32 `json:"minReplicas"`
	MaxReplicas                    int32 `json:"maxReplicas"`
	TargetCPUUtilizationPercentage int32 `json:"targetCPUUtilizationPercentage"`
	// Status
	ObservedGeneration              int64  `json:"observedGeneration,omitempty"`
	LastScaleTime                   string `json:"lastScaleTime,omitempty"`
	CurrentReplicas                 int32  `json:"currentReplicas"`
	DesiredReplicas                 int32  `json:"desiredReplicas"`
	CurrentCPUUtilizationPercentage int32  `json:"currentCPUUtilizationPercentage,omitempty"`
}

func (autoscaler *Autoscaler) Parse(d *v1.HorizontalPodAutoscaler) {
	autoscaler.Name = d.Name
	autoscaler.Labels = d.Labels
	autoscaler.CreatedAt = formatTime(d.CreationTimestamp.Time)

	// Spec
	autoscaler.MaxReplicas = d.Spec.MaxReplicas

	if d.Spec.MinReplicas != nil {
		autoscaler.MinReplicas = *d.Spec.MinReplicas
	}

	if d.Spec.TargetCPUUtilizationPercentage != nil {
		autoscaler.TargetCPUUtilizationPercentage = *d.Spec.TargetCPUUtilizationPercentage
	}

	// Status
	autoscaler.CurrentReplicas = d.Status.CurrentReplicas
	autoscaler.DesiredReplicas = d.Status.DesiredReplicas

	if d.Status.ObservedGeneration != nil {
		autoscaler.ObservedGeneration = *d.Status.ObservedGeneration
	}

	if d.Status.LastScaleTime != nil {
		autoscaler.LastScaleTime = formatTime((*d.Status.LastScaleTime).Time)
	}

	if d.Status.CurrentCPUUtilizationPercentage != nil {
		autoscaler.CurrentCPUUtilizationPercentage = *d.Status.CurrentCPUUtilizationPercentage
	}
}
