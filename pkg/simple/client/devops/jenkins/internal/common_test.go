package internal

import "testing"

func TestCommonSituation(t *testing.T) {
	// make sure these functions do not panic
	// I add these test cases because it's possible that users just do give the git source
	AppendGitlabSourceToEtree(nil, nil)
	AppendGithubSourceToEtree(nil, nil)
	AppendBitbucketServerSourceToEtree(nil, nil)
	AppendGitSourceToEtree(nil, nil)
	AppendSingleSvnSourceToEtree(nil, nil)
	AppendSvnSourceToEtree(nil, nil)
}
