package daemon // import "github.com/docker/docker/daemon"

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/daemon/names"
)

var (
	validCheckpointNameChars   = names.RestrictedNameChars
	validCheckpointNamePattern = names.RestrictedNamePattern
)

// getCheckpointDir verifies checkpoint directory for create,remove, list options and checks if checkpoint already exists
func getCheckpointDir(checkDir, checkpointID, ctrName, ctrID, ctrCheckpointDir string, create bool) (string, error) {
	var checkpointDir string
	var err2 error
	if checkDir != "" {
		checkpointDir = checkDir
	} else {
		checkpointDir = ctrCheckpointDir
	}
	checkpointAbsDir := filepath.Join(checkpointDir, checkpointID)
	stat, err := os.Stat(checkpointAbsDir)
	if create {
		switch {
		case err == nil && stat.IsDir():
			err2 = fmt.Errorf("checkpoint with name %s already exists for container %s", checkpointID, ctrName)
		case err != nil && os.IsNotExist(err):
			err2 = os.MkdirAll(checkpointAbsDir, 0700)
		case err != nil:
			err2 = err
		case err == nil:
			err2 = fmt.Errorf("%s exists and is not a directory", checkpointAbsDir)
		}
	} else {
		switch {
		case err != nil:
			err2 = fmt.Errorf("checkpoint %s does not exists for container %s", checkpointID, ctrName)
		case err == nil && stat.IsDir():
			err2 = nil
		case err == nil:
			err2 = fmt.Errorf("%s exists and is not a directory", checkpointAbsDir)
		}
	}
	return checkpointAbsDir, err2
}

// CheckpointCreate checkpoints the process running in a container with CRIU
func (daemon *Daemon) CheckpointCreate(name string, config types.CheckpointCreateOptions) error {
	container, err := daemon.GetContainer(name)
	if err != nil {
		return err
	}

	if !container.IsRunning() {
		return fmt.Errorf("Container %s not running", name)
	}

	if container.Config.Tty {
		return fmt.Errorf("checkpoint not support on containers with tty")
	}

	if !validCheckpointNamePattern.MatchString(config.CheckpointID) {
		return fmt.Errorf("Invalid checkpoint ID (%s), only %s are allowed", config.CheckpointID, validCheckpointNameChars)
	}

	checkpointDir, err := getCheckpointDir(config.CheckpointDir, config.CheckpointID, name, container.ID, container.CheckpointDir(), true)
	if err != nil {
		return fmt.Errorf("cannot checkpoint container %s: %s", name, err)
	}

	err = daemon.containerd.CreateCheckpoint(context.Background(), container.ID, checkpointDir, config.Exit)
	if err != nil {
		os.RemoveAll(checkpointDir)
		return fmt.Errorf("Cannot checkpoint container %s: %s", name, err)
	}

	daemon.LogContainerEvent(container, "checkpoint")

	return nil
}

// CheckpointDelete deletes the specified checkpoint
func (daemon *Daemon) CheckpointDelete(name string, config types.CheckpointDeleteOptions) error {
	container, err := daemon.GetContainer(name)
	if err != nil {
		return err
	}
	checkpointDir, err := getCheckpointDir(config.CheckpointDir, config.CheckpointID, name, container.ID, container.CheckpointDir(), false)
	if err == nil {
		return os.RemoveAll(filepath.Join(checkpointDir, config.CheckpointID))
	}
	return err
}

// CheckpointList lists all checkpoints of the specified container
func (daemon *Daemon) CheckpointList(name string, config types.CheckpointListOptions) ([]types.Checkpoint, error) {
	var out []types.Checkpoint

	container, err := daemon.GetContainer(name)
	if err != nil {
		return nil, err
	}

	checkpointDir, err := getCheckpointDir(config.CheckpointDir, "", name, container.ID, container.CheckpointDir(), false)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(checkpointDir, 0755); err != nil {
		return nil, err
	}

	dirs, err := ioutil.ReadDir(checkpointDir)
	if err != nil {
		return nil, err
	}

	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		path := filepath.Join(checkpointDir, d.Name(), "config.json")
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var cpt types.Checkpoint
		if err := json.Unmarshal(data, &cpt); err != nil {
			return nil, err
		}
		out = append(out, cpt)
	}

	return out, nil
}
