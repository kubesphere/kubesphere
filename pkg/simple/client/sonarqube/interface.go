package sonarqube

import (
	sonargo "github.com/kubesphere/sonargo/sonar"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
)

type SonarInterface interface {
	GetSonarResultsByTaskIds(taskId ...string) ([]*SonarStatus, error)
}

type sonarQube struct {
	client *sonargo.Client
}

const (
	SonarAnalysisActionClass = "hudson.plugins.sonar.action.SonarAnalysisAction"
	SonarMetricKeys          = "alert_status,quality_gate_details,bugs,new_bugs,reliability_rating,new_reliability_rating,vulnerabilities,new_vulnerabilities,security_rating,new_security_rating,code_smells,new_code_smells,sqale_rating,new_maintainability_rating,sqale_index,new_technical_debt,coverage,new_coverage,new_lines_to_cover,tests,duplicated_lines_density,new_duplicated_lines_density,duplicated_blocks,ncloc,ncloc_language_distribution,projects,new_lines"
	SonarAdditionalFields    = "metrics,periods"
)

type SonarStatus struct {
	Measures      *sonargo.MeasuresComponentObject `json:"measures,omitempty"`
	Issues        *sonargo.IssuesSearchObject      `json:"issues,omitempty"`
	GeneralAction *devops.GeneralAction            `json:"generalAction,omitempty"`
	Task          *sonargo.CeTaskObject            `json:"task,omitempty"`
}

func (s *sonarQube) GetSonarResultsByTaskIds(taskIds ...string) ([]*SonarStatus, error) {
	sonarStatuses := make([]*SonarStatus, 0)
	for _, taskId := range taskIds {
		sonarStatus := &SonarStatus{}
		taskOptions := &sonargo.CeTaskOption{
			Id: taskId,
		}
		ceTask, _, err := s.client.Ce.Task(taskOptions)
		if err != nil {
			klog.Errorf("get sonar task error [%+v]", err)
			continue
		}
		sonarStatus.Task = ceTask
		measuresComponentOption := &sonargo.MeasuresComponentOption{
			Component:        ceTask.Task.ComponentKey,
			AdditionalFields: SonarAdditionalFields,
			MetricKeys:       SonarMetricKeys,
		}
		measures, _, err := s.client.Measures.Component(measuresComponentOption)
		if err != nil {
			klog.Errorf("get sonar task error [%+v]", err)
			continue
		}
		sonarStatus.Measures = measures

		issuesSearchOption := &sonargo.IssuesSearchOption{
			AdditionalFields: "_all",
			ComponentKeys:    ceTask.Task.ComponentKey,
			Resolved:         "false",
			Ps:               "10",
			S:                "FILE_LINE",
			Facets:           "severities,types",
		}
		issuesSearch, _, err := s.client.Issues.Search(issuesSearchOption)
		sonarStatus.Issues = issuesSearch
		sonarStatuses = append(sonarStatuses, sonarStatus)
	}
	return sonarStatuses, nil
}
