package container // import "github.com/docker/docker/integration/container"

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/integration/internal/container"
	"github.com/docker/docker/internal/test/request"
	"github.com/docker/docker/internal/testutil"
	"github.com/gotestyourself/gotestyourself/assert"
	is "github.com/gotestyourself/gotestyourself/assert/cmp"
	"github.com/gotestyourself/gotestyourself/fs"
	"github.com/gotestyourself/gotestyourself/poll"
	"github.com/gotestyourself/gotestyourself/skip"
)

func getPrefixAndSlashFromDaemonPlatform() (prefix, slash string) {
	if testEnv.OSType == "windows" {
		return "c:", `\`
	}
	return "", "/"
}

// Test case for #5244: `docker rm` fails if bind dir doesn't exist anymore
func TestRemoveContainerWithRemovedVolume(t *testing.T) {
	skip.If(t, testEnv.IsRemoteDaemon())

	defer setupTest(t)()
	ctx := context.Background()
	client := request.NewAPIClient(t)

	prefix, slash := getPrefixAndSlashFromDaemonPlatform()

	tempDir := fs.NewDir(t, "test-rm-container-with-removed-volume", fs.WithMode(0755))
	defer tempDir.Remove()

	cID := container.Run(t, ctx, client, container.WithCmd("true"), container.WithBind(tempDir.Path(), prefix+slash+"test"))
	poll.WaitOn(t, container.IsInState(ctx, client, cID, "exited"), poll.WithDelay(100*time.Millisecond))

	err := os.RemoveAll(tempDir.Path())
	assert.NilError(t, err)

	err = client.ContainerRemove(ctx, cID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
	})
	assert.NilError(t, err)

	_, _, err = client.ContainerInspectWithRaw(ctx, cID, true)
	testutil.ErrorContains(t, err, "No such container")
}

// Test case for #2099/#2125
func TestRemoveContainerWithVolume(t *testing.T) {
	defer setupTest(t)()
	ctx := context.Background()
	client := request.NewAPIClient(t)

	prefix, slash := getPrefixAndSlashFromDaemonPlatform()

	cID := container.Run(t, ctx, client, container.WithCmd("true"), container.WithVolume(prefix+slash+"srv"))
	poll.WaitOn(t, container.IsInState(ctx, client, cID, "exited"), poll.WithDelay(100*time.Millisecond))

	insp, _, err := client.ContainerInspectWithRaw(ctx, cID, true)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(1, len(insp.Mounts)))
	volName := insp.Mounts[0].Name

	err = client.ContainerRemove(ctx, cID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
	})
	assert.NilError(t, err)

	volumes, err := client.VolumeList(ctx, filters.NewArgs(filters.Arg("name", volName)))
	assert.NilError(t, err)
	assert.Check(t, is.Equal(0, len(volumes.Volumes)))
}

func TestRemoveContainerRunning(t *testing.T) {
	defer setupTest(t)()
	ctx := context.Background()
	client := request.NewAPIClient(t)

	cID := container.Run(t, ctx, client)

	err := client.ContainerRemove(ctx, cID, types.ContainerRemoveOptions{})
	testutil.ErrorContains(t, err, "cannot remove a running container")
}

func TestRemoveContainerForceRemoveRunning(t *testing.T) {
	defer setupTest(t)()
	ctx := context.Background()
	client := request.NewAPIClient(t)

	cID := container.Run(t, ctx, client)

	err := client.ContainerRemove(ctx, cID, types.ContainerRemoveOptions{
		Force: true,
	})
	assert.NilError(t, err)
}

func TestRemoveInvalidContainer(t *testing.T) {
	defer setupTest(t)()
	ctx := context.Background()
	client := request.NewAPIClient(t)

	err := client.ContainerRemove(ctx, "unknown", types.ContainerRemoveOptions{})
	testutil.ErrorContains(t, err, "No such container")
}
