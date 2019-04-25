package gojenkins

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
