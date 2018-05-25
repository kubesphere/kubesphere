package network // import "github.com/docker/docker/integration/network"

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/versions"
	"github.com/docker/docker/internal/test/request"
	"github.com/gotestyourself/gotestyourself/assert"
	is "github.com/gotestyourself/gotestyourself/assert/cmp"
	"github.com/gotestyourself/gotestyourself/skip"
)

func containsNetwork(nws []types.NetworkResource, nw types.NetworkCreateResponse) bool {
	for _, n := range nws {
		if n.ID == nw.ID {
			return true
		}
	}
	return false
}

// createAmbiguousNetworks creates three networks, of which the second network
// uses a prefix of the first network's ID as name. The third network uses the
// first network's ID as name.
//
// After successful creation, properties of all three networks is returned
func createAmbiguousNetworks(t *testing.T) (types.NetworkCreateResponse, types.NetworkCreateResponse, types.NetworkCreateResponse) {
	client := request.NewAPIClient(t)
	ctx := context.Background()

	testNet, err := client.NetworkCreate(ctx, "testNet", types.NetworkCreate{})
	assert.NilError(t, err)
	idPrefixNet, err := client.NetworkCreate(ctx, testNet.ID[:12], types.NetworkCreate{})
	assert.NilError(t, err)
	fullIDNet, err := client.NetworkCreate(ctx, testNet.ID, types.NetworkCreate{})
	assert.NilError(t, err)

	nws, err := client.NetworkList(ctx, types.NetworkListOptions{})
	assert.NilError(t, err)

	assert.Check(t, is.Equal(true, containsNetwork(nws, testNet)), "failed to create network testNet")
	assert.Check(t, is.Equal(true, containsNetwork(nws, idPrefixNet)), "failed to create network idPrefixNet")
	assert.Check(t, is.Equal(true, containsNetwork(nws, fullIDNet)), "failed to create network fullIDNet")
	return testNet, idPrefixNet, fullIDNet
}

// TestDockerNetworkDeletePreferID tests that if a network with a name
// equal to another network's ID exists, the Network with the given
// ID is removed, and not the network with the given name.
func TestDockerNetworkDeletePreferID(t *testing.T) {
	skip.If(t, versions.LessThan(testEnv.DaemonAPIVersion(), "1.34"), "broken in earlier versions")
	defer setupTest(t)()
	client := request.NewAPIClient(t)
	ctx := context.Background()
	testNet, idPrefixNet, fullIDNet := createAmbiguousNetworks(t)

	// Delete the network using a prefix of the first network's ID as name.
	// This should the network name with the id-prefix, not the original network.
	err := client.NetworkRemove(ctx, testNet.ID[:12])
	assert.NilError(t, err)

	// Delete the network using networkID. This should remove the original
	// network, not the network with the name equal to the networkID
	err = client.NetworkRemove(ctx, testNet.ID)
	assert.NilError(t, err)

	// networks "testNet" and "idPrefixNet" should be removed, but "fullIDNet" should still exist
	nws, err := client.NetworkList(ctx, types.NetworkListOptions{})
	assert.NilError(t, err)
	assert.Check(t, is.Equal(false, containsNetwork(nws, testNet)), "Network testNet not removed")
	assert.Check(t, is.Equal(false, containsNetwork(nws, idPrefixNet)), "Network idPrefixNet not removed")
	assert.Check(t, is.Equal(true, containsNetwork(nws, fullIDNet)), "Network fullIDNet not found")
}
