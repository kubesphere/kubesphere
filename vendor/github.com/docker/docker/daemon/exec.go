package daemon // import "github.com/docker/docker/daemon"

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/container"
	"github.com/docker/docker/container/stream"
	"github.com/docker/docker/daemon/exec"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/pools"
	"github.com/docker/docker/pkg/signal"
	"github.com/docker/docker/pkg/term"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Seconds to wait after sending TERM before trying KILL
const termProcessTimeout = 10

func (d *Daemon) registerExecCommand(container *container.Container, config *exec.Config) {
	// Storing execs in container in order to kill them gracefully whenever the container is stopped or removed.
	container.ExecCommands.Add(config.ID, config)
	// Storing execs in daemon for easy access via Engine API.
	d.execCommands.Add(config.ID, config)
}

// ExecExists looks up the exec instance and returns a bool if it exists or not.
// It will also return the error produced by `getConfig`
func (d *Daemon) ExecExists(name string) (bool, error) {
	if _, err := d.getExecConfig(name); err != nil {
		return false, err
	}
	return true, nil
}

// getExecConfig looks up the exec instance by name. If the container associated
// with the exec instance is stopped or paused, it will return an error.
func (d *Daemon) getExecConfig(name string) (*exec.Config, error) {
	ec := d.execCommands.Get(name)
	if ec == nil {
		return nil, errExecNotFound(name)
	}

	// If the exec is found but its container is not in the daemon's list of
	// containers then it must have been deleted, in which case instead of
	// saying the container isn't running, we should return a 404 so that
	// the user sees the same error now that they will after the
	// 5 minute clean-up loop is run which erases old/dead execs.
	container := d.containers.Get(ec.ContainerID)
	if container == nil {
		return nil, containerNotFound(name)
	}
	if !container.IsRunning() {
		return nil, fmt.Errorf("Container %s is not running: %s", container.ID, container.State.String())
	}
	if container.IsPaused() {
		return nil, errExecPaused(container.ID)
	}
	if container.IsRestarting() {
		return nil, errContainerIsRestarting(container.ID)
	}
	return ec, nil
}

func (d *Daemon) unregisterExecCommand(container *container.Container, execConfig *exec.Config) {
	container.ExecCommands.Delete(execConfig.ID, execConfig.Pid)
	d.execCommands.Delete(execConfig.ID, execConfig.Pid)
}

func (d *Daemon) getActiveContainer(name string) (*container.Container, error) {
	container, err := d.GetContainer(name)
	if err != nil {
		return nil, err
	}

	if !container.IsRunning() {
		return nil, errNotRunning(container.ID)
	}
	if container.IsPaused() {
		return nil, errExecPaused(name)
	}
	if container.IsRestarting() {
		return nil, errContainerIsRestarting(container.ID)
	}
	return container, nil
}

// ContainerExecCreate sets up an exec in a running container.
func (d *Daemon) ContainerExecCreate(name string, config *types.ExecConfig) (string, error) {
	cntr, err := d.getActiveContainer(name)
	if err != nil {
		return "", err
	}

	cmd := strslice.StrSlice(config.Cmd)
	entrypoint, args := d.getEntrypointAndArgs(strslice.StrSlice{}, cmd)

	keys := []byte{}
	if config.DetachKeys != "" {
		keys, err = term.ToBytes(config.DetachKeys)
		if err != nil {
			err = fmt.Errorf("Invalid escape keys (%s) provided", config.DetachKeys)
			return "", err
		}
	}

	execConfig := exec.NewConfig()
	execConfig.OpenStdin = config.AttachStdin
	execConfig.OpenStdout = config.AttachStdout
	execConfig.OpenStderr = config.AttachStderr
	execConfig.ContainerID = cntr.ID
	execConfig.DetachKeys = keys
	execConfig.Entrypoint = entrypoint
	execConfig.Args = args
	execConfig.Tty = config.Tty
	execConfig.Privileged = config.Privileged
	execConfig.User = config.User
	execConfig.WorkingDir = config.WorkingDir

	linkedEnv, err := d.setupLinkedContainers(cntr)
	if err != nil {
		return "", err
	}
	execConfig.Env = container.ReplaceOrAppendEnvValues(cntr.CreateDaemonEnvironment(config.Tty, linkedEnv), config.Env)
	if len(execConfig.User) == 0 {
		execConfig.User = cntr.Config.User
	}
	if len(execConfig.WorkingDir) == 0 {
		execConfig.WorkingDir = cntr.Config.WorkingDir
	}

	d.registerExecCommand(cntr, execConfig)

	attributes := map[string]string{
		"execID": execConfig.ID,
	}
	d.LogContainerEventWithAttributes(cntr, "exec_create: "+execConfig.Entrypoint+" "+strings.Join(execConfig.Args, " "), attributes)

	return execConfig.ID, nil
}

// ContainerExecStart starts a previously set up exec instance. The
// std streams are set up.
// If ctx is cancelled, the process is terminated.
func (d *Daemon) ContainerExecStart(ctx context.Context, name string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (err error) {
	var (
		cStdin           io.ReadCloser
		cStdout, cStderr io.Writer
	)

	ec, err := d.getExecConfig(name)
	if err != nil {
		return errExecNotFound(name)
	}

	ec.Lock()
	if ec.ExitCode != nil {
		ec.Unlock()
		err := fmt.Errorf("Error: Exec command %s has already run", ec.ID)
		return errdefs.Conflict(err)
	}

	if ec.Running {
		ec.Unlock()
		return errdefs.Conflict(fmt.Errorf("Error: Exec command %s is already running", ec.ID))
	}
	ec.Running = true
	ec.Unlock()

	c := d.containers.Get(ec.ContainerID)
	logrus.Debugf("starting exec command %s in container %s", ec.ID, c.ID)
	attributes := map[string]string{
		"execID": ec.ID,
	}
	d.LogContainerEventWithAttributes(c, "exec_start: "+ec.Entrypoint+" "+strings.Join(ec.Args, " "), attributes)

	defer func() {
		if err != nil {
			ec.Lock()
			ec.Running = false
			exitCode := 126
			ec.ExitCode = &exitCode
			if err := ec.CloseStreams(); err != nil {
				logrus.Errorf("failed to cleanup exec %s streams: %s", c.ID, err)
			}
			ec.Unlock()
			c.ExecCommands.Delete(ec.ID, ec.Pid)
		}
	}()

	if ec.OpenStdin && stdin != nil {
		r, w := io.Pipe()
		go func() {
			defer w.Close()
			defer logrus.Debug("Closing buffered stdin pipe")
			pools.Copy(w, stdin)
		}()
		cStdin = r
	}
	if ec.OpenStdout {
		cStdout = stdout
	}
	if ec.OpenStderr {
		cStderr = stderr
	}

	if ec.OpenStdin {
		ec.StreamConfig.NewInputPipes()
	} else {
		ec.StreamConfig.NewNopInputPipe()
	}

	p := &specs.Process{
		Args:     append([]string{ec.Entrypoint}, ec.Args...),
		Env:      ec.Env,
		Terminal: ec.Tty,
		Cwd:      ec.WorkingDir,
	}
	if p.Cwd == "" {
		p.Cwd = "/"
	}

	if err := d.execSetPlatformOpt(c, ec, p); err != nil {
		return err
	}

	attachConfig := stream.AttachConfig{
		TTY:        ec.Tty,
		UseStdin:   cStdin != nil,
		UseStdout:  cStdout != nil,
		UseStderr:  cStderr != nil,
		Stdin:      cStdin,
		Stdout:     cStdout,
		Stderr:     cStderr,
		DetachKeys: ec.DetachKeys,
		CloseStdin: true,
	}
	ec.StreamConfig.AttachStreams(&attachConfig)
	attachErr := ec.StreamConfig.CopyStreams(ctx, &attachConfig)

	// Synchronize with libcontainerd event loop
	ec.Lock()
	c.ExecCommands.Lock()
	systemPid, err := d.containerd.Exec(ctx, c.ID, ec.ID, p, cStdin != nil, ec.InitializeStdio)
	if err != nil {
		c.ExecCommands.Unlock()
		ec.Unlock()
		return translateContainerdStartErr(ec.Entrypoint, ec.SetExitCode, err)
	}
	ec.Pid = systemPid
	c.ExecCommands.Unlock()
	ec.Unlock()

	select {
	case <-ctx.Done():
		logrus.Debugf("Sending TERM signal to process %v in container %v", name, c.ID)
		d.containerd.SignalProcess(ctx, c.ID, name, int(signal.SignalMap["TERM"]))
		select {
		case <-time.After(termProcessTimeout * time.Second):
			logrus.Infof("Container %v, process %v failed to exit within %d seconds of signal TERM - using the force", c.ID, name, termProcessTimeout)
			d.containerd.SignalProcess(ctx, c.ID, name, int(signal.SignalMap["KILL"]))
		case <-attachErr:
			// TERM signal worked
		}
		return ctx.Err()
	case err := <-attachErr:
		if err != nil {
			if _, ok := err.(term.EscapeError); !ok {
				return errdefs.System(errors.Wrap(err, "exec attach failed"))
			}
			attributes := map[string]string{
				"execID": ec.ID,
			}
			d.LogContainerEventWithAttributes(c, "exec_detach", attributes)
		}
	}
	return nil
}

// execCommandGC runs a ticker to clean up the daemon references
// of exec configs that are no longer part of the container.
func (d *Daemon) execCommandGC() {
	for range time.Tick(5 * time.Minute) {
		var (
			cleaned          int
			liveExecCommands = d.containerExecIds()
		)
		for id, config := range d.execCommands.Commands() {
			if config.CanRemove {
				cleaned++
				d.execCommands.Delete(id, config.Pid)
			} else {
				if _, exists := liveExecCommands[id]; !exists {
					config.CanRemove = true
				}
			}
		}
		if cleaned > 0 {
			logrus.Debugf("clean %d unused exec commands", cleaned)
		}
	}
}

// containerExecIds returns a list of all the current exec ids that are in use
// and running inside a container.
func (d *Daemon) containerExecIds() map[string]struct{} {
	ids := map[string]struct{}{}
	for _, c := range d.containers.List() {
		for _, id := range c.ExecCommands.List() {
			ids[id] = struct{}{}
		}
	}
	return ids
}
