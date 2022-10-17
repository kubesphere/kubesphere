/*
Copyright 2020 The KubeSphere Authors.

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

package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const PipelineFinalizerName = "pipeline.finalizers.kubesphere.io"

const (
	ResourceKindPipeline      = "Pipeline"
	ResourceSingularPipeline  = "pipeline"
	ResourcePluralPipeline    = "pipelines"
	PipelinePrefix            = "pipeline.devops.kubesphere.io/"
	PipelineSpecHash          = PipelinePrefix + "spechash"
	PipelineSyncStatusAnnoKey = PipelinePrefix + "syncstatus"
	PipelineSyncTimeAnnoKey   = PipelinePrefix + "synctime"
	PipelineSyncMsgAnnoKey    = PipelinePrefix + "syncmsg"
)

// PipelineSpec defines the desired state of Pipeline
type PipelineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Type                string               `json:"type" description:"type of devops pipeline, in scm or no scm"`
	Pipeline            *NoScmPipeline       `json:"pipeline,omitempty" description:"no scm pipeline structs"`
	MultiBranchPipeline *MultiBranchPipeline `json:"multi_branch_pipeline,omitempty" description:"in scm pipeline structs"`
}

// PipelineStatus defines the observed state of Pipeline
type PipelineStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +kubebuilder:object:root=true

// Pipeline is the Schema for the pipelines API
// +k8s:openapi-gen=true
type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelineSpec   `json:"spec,omitempty"`
	Status PipelineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PipelineList contains a list of Pipeline
type PipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pipeline `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Pipeline{}, &PipelineList{})
}

const (
	NoScmPipelineType       = "pipeline"
	MultiBranchPipelineType = "multi-branch-pipeline"
)

const (
	SourceTypeSVN       = "svn"
	SourceTypeGit       = "git"
	SourceTypeSingleSVN = "single_svn"
	SourceTypeGitlab    = "gitlab"
	SourceTypeGithub    = "github"
	SourceTypeBitbucket = "bitbucket_server"
)

type NoScmPipeline struct {
	Name              string             `json:"name" description:"name of pipeline"`
	Description       string             `json:"description,omitempty" description:"description of pipeline"`
	Discarder         *DiscarderProperty `json:"discarder,omitempty" description:"Discarder of pipeline, managing when to drop a pipeline"`
	Parameters        []Parameter        `json:"parameters,omitempty" description:"Parameters define of pipeline,user could pass param when run pipeline"`
	DisableConcurrent bool               `json:"disable_concurrent,omitempty" mapstructure:"disable_concurrent" description:"Whether to prohibit the pipeline from running in parallel"`
	TimerTrigger      *TimerTrigger      `json:"timer_trigger,omitempty" mapstructure:"timer_trigger" description:"Timer to trigger pipeline run"`
	RemoteTrigger     *RemoteTrigger     `json:"remote_trigger,omitempty" mapstructure:"remote_trigger" description:"Remote api define to trigger pipeline run"`
	Jenkinsfile       string             `json:"jenkinsfile,omitempty" description:"Jenkinsfile's content'"`
}

type MultiBranchPipeline struct {
	Name                  string                 `json:"name" description:"name of pipeline"`
	Description           string                 `json:"description,omitempty" description:"description of pipeline"`
	Discarder             *DiscarderProperty     `json:"discarder,omitempty" description:"Discarder of pipeline, managing when to drop a pipeline"`
	TimerTrigger          *TimerTrigger          `json:"timer_trigger,omitempty" mapstructure:"timer_trigger" description:"Timer to trigger pipeline run"`
	SourceType            string                 `json:"source_type" description:"type of scm, such as github/git/svn"`
	GitSource             *GitSource             `json:"git_source,omitempty" description:"git scm define"`
	GitHubSource          *GithubSource          `json:"github_source,omitempty" description:"github scm define"`
	GitlabSource          *GitlabSource          `json:"gitlab_source,omitempty" description:"gitlab scm define"`
	SvnSource             *SvnSource             `json:"svn_source,omitempty" description:"multi branch svn scm define"`
	SingleSvnSource       *SingleSvnSource       `json:"single_svn_source,omitempty" description:"single branch svn scm define"`
	BitbucketServerSource *BitbucketServerSource `json:"bitbucket_server_source,omitempty" description:"bitbucket server scm defile"`
	ScriptPath            string                 `json:"script_path" mapstructure:"script_path" description:"script path in scm"`
	MultiBranchJobTrigger *MultiBranchJobTrigger `json:"multibranch_job_trigger,omitempty" mapstructure:"multibranch_job_trigger" description:"Pipeline tasks that need to be triggered when branch creation/deletion"`
}

type GitSource struct {
	ScmId            string          `json:"scm_id,omitempty" description:"uid of scm"`
	Url              string          `json:"url,omitempty" mapstructure:"url" description:"url of git source"`
	CredentialId     string          `json:"credential_id,omitempty" mapstructure:"credential_id" description:"credential id to access git source"`
	DiscoverBranches bool            `json:"discover_branches,omitempty" mapstructure:"discover_branches" description:"Whether to discover a branch"`
	DiscoverTags     bool            `json:"discover_tags,omitempty" mapstructure:"discover_tags" description:"Discover tags configuration"`
	CloneOption      *GitCloneOption `json:"git_clone_option,omitempty" mapstructure:"git_clone_option" description:"advavced git clone options"`
	RegexFilter      string          `json:"regex_filter,omitempty" mapstructure:"regex_filter" description:"Regex used to match the name of the branch that needs to be run"`
}

// GithubSource and BitbucketServerSource have the same structure, but we don't use one due to crd errors
type GithubSource struct {
	ScmId                string               `json:"scm_id,omitempty" description:"uid of scm"`
	Owner                string               `json:"owner,omitempty" mapstructure:"owner" description:"owner of github repo"`
	Repo                 string               `json:"repo,omitempty" mapstructure:"repo" description:"repo name of github repo"`
	CredentialId         string               `json:"credential_id,omitempty" mapstructure:"credential_id" description:"credential id to access github source"`
	ApiUri               string               `json:"api_uri,omitempty" mapstructure:"api_uri" description:"The api url can specify the location of the github apiserver.For private cloud configuration"`
	DiscoverBranches     int                  `json:"discover_branches,omitempty" mapstructure:"discover_branches" description:"Discover branch configuration"`
	DiscoverPRFromOrigin int                  `json:"discover_pr_from_origin,omitempty" mapstructure:"discover_pr_from_origin" description:"Discover origin PR configuration"`
	DiscoverPRFromForks  *DiscoverPRFromForks `json:"discover_pr_from_forks,omitempty" mapstructure:"discover_pr_from_forks" description:"Discover fork PR configuration"`
	DiscoverTags         bool                 `json:"discover_tags,omitempty" mapstructure:"discover_tags" description:"Discover tag configuration"`
	CloneOption          *GitCloneOption      `json:"git_clone_option,omitempty" mapstructure:"git_clone_option" description:"advavced git clone options"`
	RegexFilter          string               `json:"regex_filter,omitempty" mapstructure:"regex_filter" description:"Regex used to match the name of the branch that needs to be run"`
}

type GitlabSource struct {
	ScmId                string               `json:"scm_id,omitempty" description:"uid of scm"`
	Owner                string               `json:"owner,omitempty" mapstructure:"owner" description:"owner of gitlab repo"`
	Repo                 string               `json:"repo,omitempty" mapstructure:"repo" description:"repo name of gitlab repo"`
	ServerName           string               `json:"server_name,omitempty" description:"the name of gitlab server which was configured in jenkins"`
	CredentialId         string               `json:"credential_id,omitempty" mapstructure:"credential_id" description:"credential id to access gitlab source"`
	ApiUri               string               `json:"api_uri,omitempty" mapstructure:"api_uri" description:"The api url can specify the location of the gitlab apiserver.For private cloud configuration"`
	DiscoverBranches     int                  `json:"discover_branches,omitempty" mapstructure:"discover_branches" description:"Discover branch configuration"`
	DiscoverPRFromOrigin int                  `json:"discover_pr_from_origin,omitempty" mapstructure:"discover_pr_from_origin" description:"Discover origin PR configuration"`
	DiscoverPRFromForks  *DiscoverPRFromForks `json:"discover_pr_from_forks,omitempty" mapstructure:"discover_pr_from_forks" description:"Discover fork PR configuration"`
	DiscoverTags         bool                 `json:"discover_tags,omitempty" mapstructure:"discover_tags" description:"Discover tags configuration"`
	CloneOption          *GitCloneOption      `json:"git_clone_option,omitempty" mapstructure:"git_clone_option" description:"advavced git clone options"`
	RegexFilter          string               `json:"regex_filter,omitempty" mapstructure:"regex_filter" description:"Regex used to match the name of the branch that needs to be run"`
}

type BitbucketServerSource struct {
	ScmId                string               `json:"scm_id,omitempty" description:"uid of scm"`
	Owner                string               `json:"owner,omitempty" mapstructure:"owner" description:"owner of github repo"`
	Repo                 string               `json:"repo,omitempty" mapstructure:"repo" description:"repo name of github repo"`
	CredentialId         string               `json:"credential_id,omitempty" mapstructure:"credential_id" description:"credential id to access github source"`
	ApiUri               string               `json:"api_uri,omitempty" mapstructure:"api_uri" description:"The api url can specify the location of the github apiserver.For private cloud configuration"`
	DiscoverBranches     int                  `json:"discover_branches,omitempty" mapstructure:"discover_branches" description:"Discover branch configuration"`
	DiscoverPRFromOrigin int                  `json:"discover_pr_from_origin,omitempty" mapstructure:"discover_pr_from_origin" description:"Discover origin PR configuration"`
	DiscoverPRFromForks  *DiscoverPRFromForks `json:"discover_pr_from_forks,omitempty" mapstructure:"discover_pr_from_forks" description:"Discover fork PR configuration"`
	DiscoverTags         bool                 `json:"discover_tags,omitempty" mapstructure:"discover_tags" description:"Discover tag configuration"`
	CloneOption          *GitCloneOption      `json:"git_clone_option,omitempty" mapstructure:"git_clone_option" description:"advavced git clone options"`
	RegexFilter          string               `json:"regex_filter,omitempty" mapstructure:"regex_filter" description:"Regex used to match the name of the branch that needs to be run"`
}

type MultiBranchJobTrigger struct {
	CreateActionJobsToTrigger string `json:"create_action_job_to_trigger,omitempty" description:"pipeline name to trigger"`
	DeleteActionJobsToTrigger string `json:"delete_action_job_to_trigger,omitempty" description:"pipeline name to trigger"`
}

type GitCloneOption struct {
	Shallow bool `json:"shallow,omitempty" mapstructure:"shallow" description:"Whether to use git shallow clone"`
	Timeout int  `json:"timeout,omitempty" mapstructure:"timeout" description:"git clone timeout mins"`
	Depth   int  `json:"depth,omitempty" mapstructure:"depth" description:"git clone depth"`
}

type SvnSource struct {
	ScmId        string `json:"scm_id,omitempty" description:"uid of scm"`
	Remote       string `json:"remote,omitempty" description:"remote address url"`
	CredentialId string `json:"credential_id,omitempty" mapstructure:"credential_id" description:"credential id to access svn source"`
	Includes     string `json:"includes,omitempty" description:"branches to run pipeline"`
	Excludes     string `json:"excludes,omitempty" description:"branches do not run pipeline"`
}
type SingleSvnSource struct {
	ScmId        string `json:"scm_id,omitempty" description:"uid of scm"`
	Remote       string `json:"remote,omitempty" description:"remote address url"`
	CredentialId string `json:"credential_id,omitempty" mapstructure:"credential_id" description:"credential id to access svn source"`
}

type DiscoverPRFromForks struct {
	Strategy int `json:"strategy,omitempty" mapstructure:"strategy" description:"github discover strategy"`
	Trust    int `json:"trust,omitempty" mapstructure:"trust" description:"trust user type"`
}

type DiscarderProperty struct {
	DaysToKeep string `json:"days_to_keep,omitempty" mapstructure:"days_to_keep" description:"days to keep pipeline"`
	NumToKeep  string `json:"num_to_keep,omitempty" mapstructure:"num_to_keep" description:"nums to keep pipeline"`
}

type Parameter struct {
	Name         string `json:"name" description:"name of param"`
	DefaultValue string `json:"default_value,omitempty" mapstructure:"default_value" description:"default value of param"`
	Type         string `json:"type" description:"type of param"`
	Description  string `json:"description,omitempty" description:"description of pipeline"`
}

type TimerTrigger struct {
	// user in no scm job
	Cron string `json:"cron,omitempty" description:"jenkins cron script"`

	// use in multi-branch job
	Interval string `json:"interval,omitempty" description:"interval ms"`
}

type RemoteTrigger struct {
	Token string `json:"token,omitempty" description:"remote trigger token"`
}

func init() {
	SchemeBuilder.Register(&Pipeline{}, &PipelineList{})
}
