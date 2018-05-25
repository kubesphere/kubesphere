package volumes

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/internal/test/daemon"
	"github.com/docker/docker/internal/test/fixtures/plugin"
	"github.com/gotestyourself/gotestyourself/assert"
	"github.com/gotestyourself/gotestyourself/skip"
)

// TestPluginWithDevMounts tests very specific regression caused by mounts ordering
// (sorted in the daemon). See #36698
func TestPluginWithDevMounts(t *testing.T) {
	skip.If(t, testEnv.IsRemoteDaemon, "cannot run daemon when remote daemon")
	t.Parallel()

	d := daemon.New(t)
	d.Start(t, "--iptables=false")
	defer d.Stop(t)

	client, err := d.NewClient()
	assert.Assert(t, err)
	ctx := context.Background()

	testDir, err := ioutil.TempDir("", "test-dir")
	assert.Assert(t, err)
	defer os.RemoveAll(testDir)

	createPlugin(t, client, "test", "dummy", asVolumeDriver, func(c *plugin.Config) {
		root := "/"
		dev := "/dev"
		mounts := []types.PluginMount{
			{Type: "bind", Source: &root, Destination: "/host", Options: []string{"rbind"}},
			{Type: "bind", Source: &dev, Destination: "/dev", Options: []string{"rbind"}},
			{Type: "bind", Source: &testDir, Destination: "/etc/foo", Options: []string{"rbind"}},
		}
		c.PluginConfig.Mounts = append(c.PluginConfig.Mounts, mounts...)
		c.PropagatedMount = "/propagated"
		c.Network = types.PluginConfigNetwork{Type: "host"}
		c.IpcHost = true
	})

	err = client.PluginEnable(ctx, "test", types.PluginEnableOptions{Timeout: 30})
	assert.Assert(t, err)
	defer func() {
		err := client.PluginRemove(ctx, "test", types.PluginRemoveOptions{Force: true})
		assert.Check(t, err)
	}()

	p, _, err := client.PluginInspectWithRaw(ctx, "test")
	assert.Assert(t, err)
	assert.Assert(t, p.Enabled)
}
