package openpitrix

import (
	compbasemetrics "k8s.io/component-base/metrics"
	"kubesphere.io/kubesphere/pkg/utils/metrics"
)

var (
	appTemplateCreationCounter = compbasemetrics.NewCounterVec(
		&compbasemetrics.CounterOpts{
			Name:           "application_template_creation",
			Help:           "Counter of application template creation broken out for each workspace, name and create state",
			StabilityLevel: compbasemetrics.ALPHA,
		},
		[]string{"workspace", "name", "state"},
	)
)

func init() {
	Register()
}

func Register() {
	metrics.MustRegister(appTemplateCreationCounter)
}
