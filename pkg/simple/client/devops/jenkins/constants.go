package jenkins

const (
	STATUS_FAIL           = "FAIL"
	STATUS_ERROR          = "ERROR"
	STATUS_ABORTED        = "ABORTED"
	STATUS_REGRESSION     = "REGRESSION"
	STATUS_SUCCESS        = "SUCCESS"
	STATUS_FIXED          = "FIXED"
	STATUS_PASSED         = "PASSED"
	RESULT_STATUS_FAILURE = "FAILURE"
	RESULT_STATUS_FAILED  = "FAILED"
	RESULT_STATUS_SKIPPED = "SKIPPED"
	STR_RE_SPLIT_VIEW     = "(.*)/view/([^/]*)/?"
)

const (
	GLOBAL_ROLE  = "globalRoles"
	PROJECT_ROLE = "projectRoles"
)

var ParameterTypeMap = map[string]string{
	"hudson.model.StringParameterDefinition":   "string",
	"hudson.model.ChoiceParameterDefinition":   "choice",
	"hudson.model.TextParameterDefinition":     "text",
	"hudson.model.BooleanParameterDefinition":  "boolean",
	"hudson.model.FileParameterDefinition":     "file",
	"hudson.model.PasswordParameterDefinition": "password",
}
