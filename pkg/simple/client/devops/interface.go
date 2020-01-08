package devops

type Interface interface {

	CredentialOperator

	BuildGetter

	PipelineOperator

	ProjectMemberOperator

	ProjectPipelineOperator

}
