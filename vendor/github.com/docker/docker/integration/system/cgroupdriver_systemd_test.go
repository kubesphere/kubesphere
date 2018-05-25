package system

import (
	"context"
	"os"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/integration/internal/container"
	"github.com/docker/docker/internal/test/daemon"

	"github.com/gotestyourself/gotestyourself/assert"
)

// hasSystemd checks whether the host was booted with systemd as its init
// system. Stolen from
// https://github.com/coreos/go-systemd/blob/176f85496f4e/util/util.go#L68
func hasSystemd() bool {
	fi, err := os.Lstat("/run/systemd/system")
	if err != nil {
		return false
	}
	return fi.IsDir()
}

// TestCgroupDriverSystemdMemoryLimit checks that container
// memory limit can be set when using systemd cgroupdriver.
//  https://github.com/moby/moby/issues/35123
func TestCgroupDriverSystemdMemoryLimit(t *testing.T) {
	t.Parallel()

	if !hasSystemd() {
		t.Skip("systemd not available")
	}

	d := daemon.New(t)
	client, err := d.NewClient()
	assert.NilError(t, err)
	d.StartWithBusybox(t, "--exec-opt", "native.cgroupdriver=systemd", "--iptables=false")
	defer d.Stop(t)

	const mem = 64 * 1024 * 1024 // 64 MB

	ctx := context.Background()
	ctrID := container.Create(t, ctx, client, func(c *container.TestContainerConfig) {
		c.HostConfig.Resources.Memory = mem
	})
	defer client.ContainerRemove(ctx, ctrID, types.ContainerRemoveOptions{Force: true})

	err = client.ContainerStart(ctx, ctrID, types.ContainerStartOptions{})
	assert.NilError(t, err)

	s, err := client.ContainerInspect(ctx, ctrID)
	assert.NilError(t, err)
	assert.Equal(t, s.HostConfig.Memory, mem)
}
