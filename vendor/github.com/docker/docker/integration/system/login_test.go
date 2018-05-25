package system // import "github.com/docker/docker/integration/system"

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/integration/internal/requirement"
	"github.com/docker/docker/internal/test/request"
	"github.com/gotestyourself/gotestyourself/assert"
	is "github.com/gotestyourself/gotestyourself/assert/cmp"
	"github.com/gotestyourself/gotestyourself/skip"
)

// Test case for GitHub 22244
func TestLoginFailsWithBadCredentials(t *testing.T) {
	skip.IfCondition(t, !requirement.HasHubConnectivity(t))

	client := request.NewAPIClient(t)

	config := types.AuthConfig{
		Username: "no-user",
		Password: "no-password",
	}
	_, err := client.RegistryLogin(context.Background(), config)
	expected := "Error response from daemon: Get https://registry-1.docker.io/v2/: unauthorized: incorrect username or password"
	assert.Check(t, is.Error(err, expected))
}
