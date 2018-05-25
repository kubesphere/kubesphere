package daemon // import "github.com/docker/docker/daemon"

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/container"
	"github.com/docker/docker/daemon/network"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/image"
	"github.com/docker/docker/opts"
	"github.com/docker/docker/pkg/signal"
	"github.com/docker/docker/pkg/system"
	"github.com/docker/docker/pkg/truncindex"
	"github.com/docker/docker/runconfig"
	volumemounts "github.com/docker/docker/volume/mounts"
	"github.com/docker/go-connections/nat"
	"github.com/opencontainers/selinux/go-selinux/label"
	"github.com/pkg/errors"
)

// GetContainer looks for a container using the provided information, which could be
// one of the following inputs from the caller:
//  - A full container ID, which will exact match a container in daemon's list
//  - A container name, which will only exact match via the GetByName() function
//  - A partial container ID prefix (e.g. short ID) of any length that is
//    unique enough to only return a single container object
//  If none of these searches succeed, an error is returned
func (daemon *Daemon) GetContainer(prefixOrName string) (*container.Container, error) {
	if len(prefixOrName) == 0 {
		return nil, errors.WithStack(invalidIdentifier(prefixOrName))
	}

	if containerByID := daemon.containers.Get(prefixOrName); containerByID != nil {
		// prefix is an exact match to a full container ID
		return containerByID, nil
	}

	// GetByName will match only an exact name provided; we ignore errors
	if containerByName, _ := daemon.GetByName(prefixOrName); containerByName != nil {
		// prefix is an exact match to a full container Name
		return containerByName, nil
	}

	containerID, indexError := daemon.idIndex.Get(prefixOrName)
	if indexError != nil {
		// When truncindex defines an error type, use that instead
		if indexError == truncindex.ErrNotExist {
			return nil, containerNotFound(prefixOrName)
		}
		return nil, errdefs.System(indexError)
	}
	return daemon.containers.Get(containerID), nil
}

// checkContainer make sure the specified container validates the specified conditions
func (daemon *Daemon) checkContainer(container *container.Container, conditions ...func(*container.Container) error) error {
	for _, condition := range conditions {
		if err := condition(container); err != nil {
			return err
		}
	}
	return nil
}

// Exists returns a true if a container of the specified ID or name exists,
// false otherwise.
func (daemon *Daemon) Exists(id string) bool {
	c, _ := daemon.GetContainer(id)
	return c != nil
}

// IsPaused returns a bool indicating if the specified container is paused.
func (daemon *Daemon) IsPaused(id string) bool {
	c, _ := daemon.GetContainer(id)
	return c.State.IsPaused()
}

func (daemon *Daemon) containerRoot(id string) string {
	return filepath.Join(daemon.repository, id)
}

// Load reads the contents of a container from disk
// This is typically done at startup.
func (daemon *Daemon) load(id string) (*container.Container, error) {
	container := daemon.newBaseContainer(id)

	if err := container.FromDisk(); err != nil {
		return nil, err
	}
	if err := label.ReserveLabel(container.ProcessLabel); err != nil {
		return nil, err
	}

	if container.ID != id {
		return container, fmt.Errorf("Container %s is stored at %s", container.ID, id)
	}

	return container, nil
}

// Register makes a container object usable by the daemon as <container.ID>
func (daemon *Daemon) Register(c *container.Container) error {
	// Attach to stdout and stderr
	if c.Config.OpenStdin {
		c.StreamConfig.NewInputPipes()
	} else {
		c.StreamConfig.NewNopInputPipe()
	}

	// once in the memory store it is visible to other goroutines
	// grab a Lock until it has been checkpointed to avoid races
	c.Lock()
	defer c.Unlock()

	daemon.containers.Add(c.ID, c)
	daemon.idIndex.Add(c.ID)
	return c.CheckpointTo(daemon.containersReplica)
}

func (daemon *Daemon) newContainer(name string, operatingSystem string, config *containertypes.Config, hostConfig *containertypes.HostConfig, imgID image.ID, managed bool) (*container.Container, error) {
	var (
		id             string
		err            error
		noExplicitName = name == ""
	)
	id, name, err = daemon.generateIDAndName(name)
	if err != nil {
		return nil, err
	}

	if hostConfig.NetworkMode.IsHost() {
		if config.Hostname == "" {
			config.Hostname, err = os.Hostname()
			if err != nil {
				return nil, errdefs.System(err)
			}
		}
	} else {
		daemon.generateHostname(id, config)
	}
	entrypoint, args := daemon.getEntrypointAndArgs(config.Entrypoint, config.Cmd)

	base := daemon.newBaseContainer(id)
	base.Created = time.Now().UTC()
	base.Managed = managed
	base.Path = entrypoint
	base.Args = args //FIXME: de-duplicate from config
	base.Config = config
	base.HostConfig = &containertypes.HostConfig{}
	base.ImageID = imgID
	base.NetworkSettings = &network.Settings{IsAnonymousEndpoint: noExplicitName}
	base.Name = name
	base.Driver = daemon.imageService.GraphDriverForOS(operatingSystem)
	base.OS = operatingSystem
	return base, err
}

// GetByName returns a container given a name.
func (daemon *Daemon) GetByName(name string) (*container.Container, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("No container name supplied")
	}
	fullName := name
	if name[0] != '/' {
		fullName = "/" + name
	}
	id, err := daemon.containersReplica.Snapshot().GetID(fullName)
	if err != nil {
		return nil, fmt.Errorf("Could not find entity for %s", name)
	}
	e := daemon.containers.Get(id)
	if e == nil {
		return nil, fmt.Errorf("Could not find container for entity id %s", id)
	}
	return e, nil
}

// newBaseContainer creates a new container with its initial
// configuration based on the root storage from the daemon.
func (daemon *Daemon) newBaseContainer(id string) *container.Container {
	return container.NewBaseContainer(id, daemon.containerRoot(id))
}

func (daemon *Daemon) getEntrypointAndArgs(configEntrypoint strslice.StrSlice, configCmd strslice.StrSlice) (string, []string) {
	if len(configEntrypoint) != 0 {
		return configEntrypoint[0], append(configEntrypoint[1:], configCmd...)
	}
	return configCmd[0], configCmd[1:]
}

func (daemon *Daemon) generateHostname(id string, config *containertypes.Config) {
	// Generate default hostname
	if config.Hostname == "" {
		config.Hostname = id[:12]
	}
}

func (daemon *Daemon) setSecurityOptions(container *container.Container, hostConfig *containertypes.HostConfig) error {
	container.Lock()
	defer container.Unlock()
	return daemon.parseSecurityOpt(container, hostConfig)
}

func (daemon *Daemon) setHostConfig(container *container.Container, hostConfig *containertypes.HostConfig) error {
	// Do not lock while creating volumes since this could be calling out to external plugins
	// Don't want to block other actions, like `docker ps` because we're waiting on an external plugin
	if err := daemon.registerMountPoints(container, hostConfig); err != nil {
		return err
	}

	container.Lock()
	defer container.Unlock()

	// Register any links from the host config before starting the container
	if err := daemon.registerLinks(container, hostConfig); err != nil {
		return err
	}

	runconfig.SetDefaultNetModeIfBlank(hostConfig)
	container.HostConfig = hostConfig
	return container.CheckpointTo(daemon.containersReplica)
}

// verifyContainerSettings performs validation of the hostconfig and config
// structures.
func (daemon *Daemon) verifyContainerSettings(platform string, hostConfig *containertypes.HostConfig, config *containertypes.Config, update bool) ([]string, error) {
	// First perform verification of settings common across all platforms.
	if config != nil {
		if config.WorkingDir != "" {
			wdInvalid := false
			if runtime.GOOS == platform {
				config.WorkingDir = filepath.FromSlash(config.WorkingDir) // Ensure in platform semantics
				if !system.IsAbs(config.WorkingDir) {
					wdInvalid = true
				}
			} else {
				// LCOW. Force Unix semantics
				config.WorkingDir = strings.Replace(config.WorkingDir, string(os.PathSeparator), "/", -1)
				if !path.IsAbs(config.WorkingDir) {
					wdInvalid = true
				}
			}
			if wdInvalid {
				return nil, fmt.Errorf("the working directory '%s' is invalid, it needs to be an absolute path", config.WorkingDir)
			}
		}

		if len(config.StopSignal) > 0 {
			_, err := signal.ParseSignal(config.StopSignal)
			if err != nil {
				return nil, err
			}
		}

		// Validate if Env contains empty variable or not (e.g., ``, `=foo`)
		for _, env := range config.Env {
			if _, err := opts.ValidateEnv(env); err != nil {
				return nil, err
			}
		}

		// Validate the healthcheck params of Config
		if config.Healthcheck != nil {
			if config.Healthcheck.Interval != 0 && config.Healthcheck.Interval < containertypes.MinimumDuration {
				return nil, errors.Errorf("Interval in Healthcheck cannot be less than %s", containertypes.MinimumDuration)
			}

			if config.Healthcheck.Timeout != 0 && config.Healthcheck.Timeout < containertypes.MinimumDuration {
				return nil, errors.Errorf("Timeout in Healthcheck cannot be less than %s", containertypes.MinimumDuration)
			}

			if config.Healthcheck.Retries < 0 {
				return nil, errors.Errorf("Retries in Healthcheck cannot be negative")
			}

			if config.Healthcheck.StartPeriod != 0 && config.Healthcheck.StartPeriod < containertypes.MinimumDuration {
				return nil, errors.Errorf("StartPeriod in Healthcheck cannot be less than %s", containertypes.MinimumDuration)
			}
		}
	}

	if hostConfig == nil {
		return nil, nil
	}

	if hostConfig.AutoRemove && !hostConfig.RestartPolicy.IsNone() {
		return nil, errors.Errorf("can't create 'AutoRemove' container with restart policy")
	}

	// Validate mounts; check if host directories still exist
	parser := volumemounts.NewParser(platform)
	for _, cfg := range hostConfig.Mounts {
		if err := parser.ValidateMountConfig(&cfg); err != nil {
			return nil, err
		}
	}

	for _, extraHost := range hostConfig.ExtraHosts {
		if _, err := opts.ValidateExtraHost(extraHost); err != nil {
			return nil, err
		}
	}

	for port := range hostConfig.PortBindings {
		_, portStr := nat.SplitProtoPort(string(port))
		if _, err := nat.ParsePort(portStr); err != nil {
			return nil, errors.Errorf("invalid port specification: %q", portStr)
		}
		for _, pb := range hostConfig.PortBindings[port] {
			_, err := nat.NewPort(nat.SplitProtoPort(pb.HostPort))
			if err != nil {
				return nil, errors.Errorf("invalid port specification: %q", pb.HostPort)
			}
		}
	}

	p := hostConfig.RestartPolicy

	switch p.Name {
	case "always", "unless-stopped", "no":
		if p.MaximumRetryCount != 0 {
			return nil, errors.Errorf("maximum retry count cannot be used with restart policy '%s'", p.Name)
		}
	case "on-failure":
		if p.MaximumRetryCount < 0 {
			return nil, errors.Errorf("maximum retry count cannot be negative")
		}
	case "":
		// do nothing
	default:
		return nil, errors.Errorf("invalid restart policy '%s'", p.Name)
	}

	if !hostConfig.Isolation.IsValid() {
		return nil, errors.Errorf("invalid isolation '%s' on %s", hostConfig.Isolation, runtime.GOOS)
	}

	var (
		err      error
		warnings []string
	)
	// Now do platform-specific verification
	if warnings, err = verifyPlatformContainerSettings(daemon, hostConfig, config, update); err != nil {
		return warnings, err
	}
	if hostConfig.NetworkMode.IsHost() && len(hostConfig.PortBindings) > 0 {
		warnings = append(warnings, "Published ports are discarded when using host network mode")
	}
	return warnings, err
}
