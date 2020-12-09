package internal

type PRDiscoverTrust int

const (
	PRDiscoverTrustMember     PRDiscoverTrust = 1
	PRDiscoverTrustEveryone   PRDiscoverTrust = 2
	PRDiscoverTrustPermission PRDiscoverTrust = 3
	PRDiscoverTrustNobody     PRDiscoverTrust = 4
	PRDiscoverUnknown         PRDiscoverTrust = -1
)

func (p PRDiscoverTrust) Value() int {
	return int(p)
}

func (p PRDiscoverTrust) String() string {
	switch p {
	case PRDiscoverTrustMember:
		return "TrustMembers"
	case PRDiscoverTrustEveryone:
		return "TrustEveryone"
	case PRDiscoverTrustPermission:
		return "TrustPermission"
	case PRDiscoverTrustNobody:
		return "TrustNobody"
	}
	return ""
}

func (p PRDiscoverTrust) ParseFromString(prTrust string) PRDiscoverTrust {
	switch prTrust {
	case "TrustMembers":
		return PRDiscoverTrustMember
	case "TrustEveryone":
		return PRDiscoverTrustEveryone
	case "TrustPermission":
		return PRDiscoverTrustPermission
	case "TrustNobody":
		return PRDiscoverTrustNobody
	default:
		return PRDiscoverUnknown
	}
}

// GitHub
type GitHubPRDiscoverTrust int

const (
	GitHubPRDiscoverTrustContributors GitHubPRDiscoverTrust = 1
)

func (p GitHubPRDiscoverTrust) Value() int {
	return int(p)
}

func (p PRDiscoverTrust) IsValid() bool {
	return p.String() != ""
}

func (p GitHubPRDiscoverTrust) String() string {
	switch p {
	case GitHubPRDiscoverTrustContributors:
		return "TrustContributors"
	default:
		return PRDiscoverTrust(p).String()
	}
}

func (p GitHubPRDiscoverTrust) ParseFromString(prTrust string) GitHubPRDiscoverTrust {
	switch prTrust {
	case "TrustContributors":
		return GitHubPRDiscoverTrustContributors
	default:
		return GitHubPRDiscoverTrust(PRDiscoverTrust(p).ParseFromString(prTrust))
	}
}

func (p GitHubPRDiscoverTrust) IsValid() bool {
	return PRDiscoverTrust(p).IsValid()
}

// Bitbucket
type BitbucketPRDiscoverTrust int

const (
	BitbucketPRDiscoverTrustEveryone  BitbucketPRDiscoverTrust = 1
	BitbucketPRDiscoverTrustTeamForks BitbucketPRDiscoverTrust = 2
	BitbucketPRDiscoverTrustNobody    BitbucketPRDiscoverTrust = 3
)

func (p BitbucketPRDiscoverTrust) Value() int {
	return int(p)
}

func (p BitbucketPRDiscoverTrust) IsValid() bool {
	return p.String() != ""
}

func (p BitbucketPRDiscoverTrust) String() string {
	switch p {
	default:
		fallthrough
	case BitbucketPRDiscoverTrustEveryone:
		return "TrustEveryone"
	case BitbucketPRDiscoverTrustTeamForks:
		return "TrustTeamForks"
	case BitbucketPRDiscoverTrustNobody:
		return "TrustNobody"
	}
}

func (p BitbucketPRDiscoverTrust) ParseFromString(prTrust string) BitbucketPRDiscoverTrust {
	switch prTrust {
	default:
		fallthrough
	case "TrustEveryone":
		return BitbucketPRDiscoverTrustEveryone
	case "TrustTeamForks":
		return BitbucketPRDiscoverTrustTeamForks
	case "TrustNobody":
		return BitbucketPRDiscoverTrustNobody
	}
}
