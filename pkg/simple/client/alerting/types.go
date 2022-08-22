package alerting

import (
	"time"
)

type RuleGroup struct {
	Name           string          `json:"name"`
	File           string          `json:"file"`
	Rules          []*AlertingRule `json:"rules"`
	Interval       float64         `json:"interval"`
	EvaluationTime float64         `json:"evaluationTime"`
	LastEvaluation *time.Time      `json:"lastEvaluation"`
}

type AlertingRule struct {
	// State can be "pending", "firing", "inactive".
	State       string            `json:"state"`
	Name        string            `json:"name"`
	Query       string            `json:"query"`
	Duration    float64           `json:"duration"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Alerts      []*Alert          `json:"alerts"`
	// Health can be "ok", "err", "unknown".
	Health         string     `json:"health"`
	LastError      string     `json:"lastError,omitempty"`
	EvaluationTime *float64   `json:"evaluationTime"`
	LastEvaluation *time.Time `json:"lastEvaluation"`
	// Type of an alertingRule is always "alerting".
	Type string `json:"type"`
}

type Alert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	State       string            `json:"state"`
	ActiveAt    *time.Time        `json:"activeAt,omitempty"`
	Value       string            `json:"value"`
}
