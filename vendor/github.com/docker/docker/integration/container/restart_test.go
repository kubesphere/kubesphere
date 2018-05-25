package container // import "github.com/docker/docker/integration/container"

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/internal/test/daemon"
	"github.com/gotestyourself/gotestyourself/assert"
	"github.com/gotestyourself/gotestyourself/skip"
)

func TestDaemonRestartKillContainers(t *testing.T) {
	skip.If(t, testEnv.IsRemoteDaemon, "cannot start daemon on remote test run")
	type testCase struct {
		desc       string
		config     *container.Config
		hostConfig *container.HostConfig

		xRunning            bool
		xRunningLiveRestore bool
		xStart              bool
	}

	for _, c := range []testCase{
		{
			desc:                "container without restart policy",
			config:              &container.Config{Image: "busybox", Cmd: []string{"top"}},
			xRunningLiveRestore: true,
			xStart:              true,
		},
		{
			desc:                "container with restart=always",
			config:              &container.Config{Image: "busybox", Cmd: []string{"top"}},
			hostConfig:          &container.HostConfig{RestartPolicy: container.RestartPolicy{Name: "always"}},
			xRunning:            true,
			xRunningLiveRestore: true,
			xStart:              true,
		},
		{
			desc:       "container created should not be restarted",
			config:     &container.Config{Image: "busybox", Cmd: []string{"top"}},
			hostConfig: &container.HostConfig{RestartPolicy: container.RestartPolicy{Name: "always"}},
		},
	} {
		for _, liveRestoreEnabled := range []bool{false, true} {
			for fnName, stopDaemon := range map[string]func(*testing.T, *daemon.Daemon){
				"kill-daemon": func(t *testing.T, d *daemon.Daemon) {
					err := d.Kill()
					assert.NilError(t, err)
				},
				"stop-daemon": func(t *testing.T, d *daemon.Daemon) {
					d.Stop(t)
				},
			} {
				t.Run(fmt.Sprintf("live-restore=%v/%s/%s", liveRestoreEnabled, c.desc, fnName), func(t *testing.T) {
					c := c
					liveRestoreEnabled := liveRestoreEnabled
					stopDaemon := stopDaemon

					t.Parallel()

					d := daemon.New(t)
					client, err := d.NewClient()
					assert.NilError(t, err)

					args := []string{"--iptables=false"}
					if liveRestoreEnabled {
						args = append(args, "--live-restore")
					}

					d.StartWithBusybox(t, args...)
					defer d.Stop(t)
					ctx := context.Background()

					resp, err := client.ContainerCreate(ctx, c.config, c.hostConfig, nil, "")
					assert.NilError(t, err)
					defer client.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true})

					if c.xStart {
						err = client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
						assert.NilError(t, err)
					}

					stopDaemon(t, d)
					d.Start(t, args...)

					expected := c.xRunning
					if liveRestoreEnabled {
						expected = c.xRunningLiveRestore
					}

					var running bool
					for i := 0; i < 30; i++ {
						inspect, err := client.ContainerInspect(ctx, resp.ID)
						assert.NilError(t, err)

						running = inspect.State.Running
						if running == expected {
							break
						}
						time.Sleep(2 * time.Second)

					}
					assert.Equal(t, expected, running, "got unexpected running state, expected %v, got: %v", expected, running)
					// TODO(cpuguy83): test pause states... this seems to be rather undefined currently
				})
			}
		}
	}
}
