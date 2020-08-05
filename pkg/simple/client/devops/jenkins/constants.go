/*
Copyright 2020 KubeSphere Authors

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
