package container // import "github.com/docker/docker/integration/container"

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/versions"
	"github.com/docker/docker/integration/internal/container"
	"github.com/docker/docker/internal/test/request"
	"github.com/docker/docker/internal/testutil"
	"github.com/docker/docker/pkg/stringid"
	"github.com/gotestyourself/gotestyourself/assert"
	is "github.com/gotestyourself/gotestyourself/assert/cmp"
	"github.com/gotestyourself/gotestyourself/poll"
	"github.com/gotestyourself/gotestyourself/skip"
)

// This test simulates the scenario mentioned in #31392:
// Having two linked container, renaming the target and bringing a replacement
// and then deleting and recreating the source container linked to the new target.
// This checks that "rename" updates source container correctly and doesn't set it to null.
func TestRenameLinkedContainer(t *testing.T) {
	skip.If(t, versions.LessThan(testEnv.DaemonAPIVersion(), "1.32"), "broken in earlier versions")
	defer setupTest(t)()
	ctx := context.Background()
	client := request.NewAPIClient(t)

	aName := "a0" + t.Name()
	bName := "b0" + t.Name()
	aID := container.Run(t, ctx, client, container.WithName(aName))
	bID := container.Run(t, ctx, client, container.WithName(bName), container.WithLinks(aName))

	err := client.ContainerRename(ctx, aID, "a1"+t.Name())
	assert.NilError(t, err)

	container.Run(t, ctx, client, container.WithName(aName))

	err = client.ContainerRemove(ctx, bID, types.ContainerRemoveOptions{Force: true})
	assert.NilError(t, err)

	bID = container.Run(t, ctx, client, container.WithName(bName), container.WithLinks(aName))

	inspect, err := client.ContainerInspect(ctx, bID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual([]string{"/" + aName + ":/" + bName + "/" + aName}, inspect.HostConfig.Links))
}

func TestRenameStoppedContainer(t *testing.T) {
	defer setupTest(t)()
	ctx := context.Background()
	client := request.NewAPIClient(t)

	oldName := "first_name" + t.Name()
	cID := container.Run(t, ctx, client, container.WithName(oldName), container.WithCmd("sh"))
	poll.WaitOn(t, container.IsInState(ctx, client, cID, "exited"), poll.WithDelay(100*time.Millisecond))

	inspect, err := client.ContainerInspect(ctx, cID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal("/"+oldName, inspect.Name))

	newName := "new_name" + stringid.GenerateNonCryptoID()
	err = client.ContainerRename(ctx, oldName, newName)
	assert.NilError(t, err)

	inspect, err = client.ContainerInspect(ctx, cID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal("/"+newName, inspect.Name))
}

func TestRenameRunningContainerAndReuse(t *testing.T) {
	defer setupTest(t)()
	ctx := context.Background()
	client := request.NewAPIClient(t)

	oldName := "first_name" + t.Name()
	cID := container.Run(t, ctx, client, container.WithName(oldName))
	poll.WaitOn(t, container.IsInState(ctx, client, cID, "running"), poll.WithDelay(100*time.Millisecond))

	newName := "new_name" + stringid.GenerateNonCryptoID()
	err := client.ContainerRename(ctx, oldName, newName)
	assert.NilError(t, err)

	inspect, err := client.ContainerInspect(ctx, cID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal("/"+newName, inspect.Name))

	_, err = client.ContainerInspect(ctx, oldName)
	testutil.ErrorContains(t, err, "No such container: "+oldName)

	cID = container.Run(t, ctx, client, container.WithName(oldName))
	poll.WaitOn(t, container.IsInState(ctx, client, cID, "running"), poll.WithDelay(100*time.Millisecond))

	inspect, err = client.ContainerInspect(ctx, cID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal("/"+oldName, inspect.Name))
}

func TestRenameInvalidName(t *testing.T) {
	defer setupTest(t)()
	ctx := context.Background()
	client := request.NewAPIClient(t)

	oldName := "first_name" + t.Name()
	cID := container.Run(t, ctx, client, container.WithName(oldName))
	poll.WaitOn(t, container.IsInState(ctx, client, cID, "running"), poll.WithDelay(100*time.Millisecond))

	err := client.ContainerRename(ctx, oldName, "new:invalid")
	testutil.ErrorContains(t, err, "Invalid container name")

	inspect, err := client.ContainerInspect(ctx, oldName)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(cID, inspect.ID))
}

// Test case for GitHub issue 22466
// Docker's service discovery works for named containers so
// ping to a named container should work, and an anonymous
// container without a name does not work with service discovery.
// However, an anonymous could be renamed to a named container.
// This test is to make sure once the container has been renamed,
// the service discovery for the (re)named container works.
func TestRenameAnonymousContainer(t *testing.T) {
	defer setupTest(t)()
	ctx := context.Background()
	client := request.NewAPIClient(t)

	networkName := "network1" + t.Name()
	_, err := client.NetworkCreate(ctx, networkName, types.NetworkCreate{})

	assert.NilError(t, err)
	cID := container.Run(t, ctx, client, func(c *container.TestContainerConfig) {
		c.NetworkingConfig.EndpointsConfig = map[string]*network.EndpointSettings{
			networkName: {},
		}
		c.HostConfig.NetworkMode = containertypes.NetworkMode(networkName)
	})

	container1Name := "container1" + t.Name()
	err = client.ContainerRename(ctx, cID, container1Name)
	assert.NilError(t, err)
	// Stop/Start the container to get registered
	// FIXME(vdemeester) this is a really weird behavior as it fails otherwise
	err = client.ContainerStop(ctx, container1Name, nil)
	assert.NilError(t, err)
	err = client.ContainerStart(ctx, container1Name, types.ContainerStartOptions{})
	assert.NilError(t, err)

	poll.WaitOn(t, container.IsInState(ctx, client, cID, "running"), poll.WithDelay(100*time.Millisecond))

	count := "-c"
	if testEnv.OSType == "windows" {
		count = "-n"
	}
	cID = container.Run(t, ctx, client, func(c *container.TestContainerConfig) {
		c.NetworkingConfig.EndpointsConfig = map[string]*network.EndpointSettings{
			networkName: {},
		}
		c.HostConfig.NetworkMode = containertypes.NetworkMode(networkName)
	}, container.WithCmd("ping", count, "1", container1Name))
	poll.WaitOn(t, container.IsInState(ctx, client, cID, "exited"), poll.WithDelay(100*time.Millisecond))

	inspect, err := client.ContainerInspect(ctx, cID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(0, inspect.State.ExitCode), "container %s exited with the wrong exitcode: %+v", cID, inspect)
}

// TODO: should be a unit test
func TestRenameContainerWithSameName(t *testing.T) {
	defer setupTest(t)()
	ctx := context.Background()
	client := request.NewAPIClient(t)

	oldName := "old" + t.Name()
	cID := container.Run(t, ctx, client, container.WithName(oldName))

	poll.WaitOn(t, container.IsInState(ctx, client, cID, "running"), poll.WithDelay(100*time.Millisecond))
	err := client.ContainerRename(ctx, oldName, oldName)
	testutil.ErrorContains(t, err, "Renaming a container with the same name")
	err = client.ContainerRename(ctx, cID, oldName)
	testutil.ErrorContains(t, err, "Renaming a container with the same name")
}

// Test case for GitHub issue 23973
// When a container is being renamed, the container might
// be linked to another container. In that case, the meta data
// of the linked container should be updated so that the other
// container could still reference to the container that is renamed.
func TestRenameContainerWithLinkedContainer(t *testing.T) {
	skip.If(t, testEnv.IsRemoteDaemon())

	defer setupTest(t)()
	ctx := context.Background()
	client := request.NewAPIClient(t)

	db1Name := "db1" + t.Name()
	db1ID := container.Run(t, ctx, client, container.WithName(db1Name))
	poll.WaitOn(t, container.IsInState(ctx, client, db1ID, "running"), poll.WithDelay(100*time.Millisecond))

	app1Name := "app1" + t.Name()
	app2Name := "app2" + t.Name()
	app1ID := container.Run(t, ctx, client, container.WithName(app1Name), container.WithLinks(db1Name+":/mysql"))
	poll.WaitOn(t, container.IsInState(ctx, client, app1ID, "running"), poll.WithDelay(100*time.Millisecond))

	err := client.ContainerRename(ctx, app1Name, app2Name)
	assert.NilError(t, err)

	inspect, err := client.ContainerInspect(ctx, app2Name+"/mysql")
	assert.NilError(t, err)
	assert.Check(t, is.Equal(db1ID, inspect.ID))
}
