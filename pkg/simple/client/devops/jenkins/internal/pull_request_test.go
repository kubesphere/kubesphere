package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPRDiscoverTrust(t *testing.T) {
	assert.Equal(t, PRDiscoverTrust(1).String(), "TrustMembers")
	assert.Equal(t, PRDiscoverTrust(2).String(), "TrustEveryone")
	assert.Equal(t, PRDiscoverTrust(3).String(), "TrustPermission")
	assert.Equal(t, PRDiscoverTrust(4).String(), "TrustNobody")
	assert.Equal(t, PRDiscoverTrust(-1).IsValid(), false)
	assert.Equal(t, PRDiscoverTrust(1).Value(), 1)

	assert.Equal(t, PRDiscoverTrust(1).ParseFromString("TrustMembers"), PRDiscoverTrustMember)
	assert.Equal(t, PRDiscoverTrust(1).ParseFromString("TrustEveryone"), PRDiscoverTrustEveryone)
	assert.Equal(t, PRDiscoverTrust(1).ParseFromString("TrustPermission"), PRDiscoverTrustPermission)
	assert.Equal(t, PRDiscoverTrust(1).ParseFromString("TrustNobody"), PRDiscoverTrustNobody)
	assert.Equal(t, PRDiscoverTrust(1).ParseFromString("fake").IsValid(), false)

	// GitHub
	assert.Equal(t, GitHubPRDiscoverTrust(1).String(), "TrustContributors")
	assert.Equal(t, GitHubPRDiscoverTrust(2).String(), PRDiscoverTrust(2).String())
	assert.Equal(t, GitHubPRDiscoverTrust(1).Value(), 1)
	assert.Equal(t, GitHubPRDiscoverTrust(1).ParseFromString("TrustContributors"), GitHubPRDiscoverTrustContributors)
	assert.Equal(t, GitHubPRDiscoverTrust(1).ParseFromString("TrustEveryone").String(), "TrustEveryone")
	assert.Equal(t, GitHubPRDiscoverTrust(1).ParseFromString("fake").IsValid(), false)

	// Bithucket
	assert.Equal(t, BitbucketPRDiscoverTrust(1).String(), "TrustEveryone")
	assert.Equal(t, BitbucketPRDiscoverTrust(2).String(), "TrustTeamForks")
	assert.Equal(t, BitbucketPRDiscoverTrust(3).String(), "TrustNobody")
	assert.Equal(t, BitbucketPRDiscoverTrust(3).Value(), 3)
	assert.Equal(t, BitbucketPRDiscoverTrust(-1).String(), "TrustEveryone")
	assert.Equal(t, BitbucketPRDiscoverTrust(1).ParseFromString("TrustEveryone"), BitbucketPRDiscoverTrustEveryone)
	assert.Equal(t, BitbucketPRDiscoverTrust(1).ParseFromString("TrustTeamForks"), BitbucketPRDiscoverTrustTeamForks)
	assert.Equal(t, BitbucketPRDiscoverTrust(1).ParseFromString("TrustNobody"), BitbucketPRDiscoverTrustNobody)
	assert.Equal(t, BitbucketPRDiscoverTrust(1).ParseFromString("fake"), BitbucketPRDiscoverTrustEveryone)
	assert.Equal(t, BitbucketPRDiscoverTrust(1).ParseFromString("TrustNobody").IsValid(), true)
}
