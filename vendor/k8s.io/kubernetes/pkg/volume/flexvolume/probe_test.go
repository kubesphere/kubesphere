/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package flexvolume

import (
	"fmt"
	"path"
	"testing"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	utilfs "k8s.io/kubernetes/pkg/util/filesystem"
	"k8s.io/kubernetes/pkg/volume"
)

const (
	pluginDir  = "/flexvolume"
	driverName = "fake-driver"
)

// Probes a driver installed before prober initialization.
func TestProberExistingDriverBeforeInit(t *testing.T) {
	// Arrange
	driverPath, _, watcher, prober := initTestEnvironment(t)

	// Act
	events, err := prober.Probe()

	// Assert
	// Probe occurs, 1 plugin should be returned, and 2 watches (pluginDir and all its
	// current subdirectories) registered.
	assert.Equal(t, 1, len(events))
	assert.Equal(t, volume.ProbeAddOrUpdate, events[0].Op)
	assert.Equal(t, pluginDir, watcher.watches[0])
	assert.Equal(t, driverPath, watcher.watches[1])
	assert.NoError(t, err)

	// Should no longer probe.

	// Act
	events, err = prober.Probe()
	// Assert
	assert.Equal(t, 0, len(events))
	assert.NoError(t, err)
}

// Probes newly added drivers after prober is running.
func TestProberAddRemoveDriver(t *testing.T) {
	// Arrange
	_, fs, watcher, prober := initTestEnvironment(t)
	prober.Probe()
	events, err := prober.Probe()
	assert.Equal(t, 0, len(events))

	// Call probe after a file is added. Should return 1 event.

	// add driver
	const driverName2 = "fake-driver2"
	driverPath := path.Join(pluginDir, driverName2)
	executablePath := path.Join(driverPath, driverName2)
	installDriver(driverName2, fs)
	watcher.TriggerEvent(fsnotify.Create, driverPath)
	watcher.TriggerEvent(fsnotify.Create, executablePath)

	// Act
	events, err = prober.Probe()

	// Assert
	assert.Equal(t, 1, len(events))
	assert.Equal(t, volume.ProbeAddOrUpdate, events[0].Op)               // 1 newly added
	assert.Equal(t, driverPath, watcher.watches[len(watcher.watches)-1]) // Checks most recent watch
	assert.NoError(t, err)

	// Call probe again, should return 0 event.

	// Act
	events, err = prober.Probe()
	// Assert
	assert.Equal(t, 0, len(events))
	assert.NoError(t, err)

	// Call probe after a non-driver file is added in a subdirectory. should return 1 event.
	fp := path.Join(driverPath, "dummyfile")
	fs.Create(fp)
	watcher.TriggerEvent(fsnotify.Create, fp)

	// Act
	events, err = prober.Probe()

	// Assert
	assert.Equal(t, 1, len(events))
	assert.Equal(t, volume.ProbeAddOrUpdate, events[0].Op)
	assert.NoError(t, err)

	// Call probe again, should return 0 event.
	// Act
	events, err = prober.Probe()
	// Assert
	assert.Equal(t, 0, len(events))
	assert.NoError(t, err)

	// Call probe after a subdirectory is added in a driver directory. should return 1 event.
	subdirPath := path.Join(driverPath, "subdir")
	fs.Create(subdirPath)
	watcher.TriggerEvent(fsnotify.Create, subdirPath)

	// Act
	events, err = prober.Probe()

	// Assert
	assert.Equal(t, 1, len(events))
	assert.Equal(t, volume.ProbeAddOrUpdate, events[0].Op)
	assert.NoError(t, err)

	// Call probe again, should return 0 event.
	// Act
	events, err = prober.Probe()
	// Assert
	assert.Equal(t, 0, len(events))
	assert.NoError(t, err)

	// Call probe after a subdirectory is removed in a driver directory. should return 1 event.
	fs.Remove(subdirPath)
	watcher.TriggerEvent(fsnotify.Remove, subdirPath)

	// Act
	events, err = prober.Probe()

	// Assert
	assert.Equal(t, 1, len(events))
	assert.Equal(t, volume.ProbeAddOrUpdate, events[0].Op)
	assert.NoError(t, err)

	// Call probe again, should return 0 event.
	// Act
	events, err = prober.Probe()
	// Assert
	assert.Equal(t, 0, len(events))
	assert.NoError(t, err)

	// Call probe after a driver executable and driver directory is remove. should return 1 event.
	fs.Remove(executablePath)
	fs.Remove(driverPath)
	watcher.TriggerEvent(fsnotify.Remove, executablePath)
	watcher.TriggerEvent(fsnotify.Remove, driverPath)
	// Act and Assert: 1 ProbeRemove event
	events, err = prober.Probe()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, volume.ProbeRemove, events[0].Op)
	assert.NoError(t, err)

	// Act and Assert: 0 event
	events, err = prober.Probe()
	assert.Equal(t, 0, len(events))
	assert.NoError(t, err)
}

// Tests the behavior when no drivers exist in the plugin directory.
func TestEmptyPluginDir(t *testing.T) {
	// Arrange
	fs := utilfs.NewFakeFs()
	watcher := NewFakeWatcher()
	prober := &flexVolumeProber{
		pluginDir: pluginDir,
		watcher:   watcher,
		fs:        fs,
		factory:   fakePluginFactory{error: false},
	}
	prober.Init()

	// Act
	events, err := prober.Probe()

	// Assert
	assert.Equal(t, 0, len(events))
	assert.NoError(t, err)
}

// Issue an event to remove plugindir. New directory should still be watched.
func TestRemovePluginDir(t *testing.T) {
	// Arrange
	driverPath, fs, watcher, _ := initTestEnvironment(t)
	fs.RemoveAll(pluginDir)
	watcher.TriggerEvent(fsnotify.Remove, path.Join(driverPath, driverName))
	watcher.TriggerEvent(fsnotify.Remove, driverPath)
	watcher.TriggerEvent(fsnotify.Remove, pluginDir)

	// Act: The handler triggered by the above events should have already handled the event appropriately.

	// Assert
	assert.Equal(t, 3, len(watcher.watches)) // 2 from initial setup, 1 from new watch.
	assert.Equal(t, pluginDir, watcher.watches[len(watcher.watches)-1])
}

// Issue an event to remove plugindir. New directory should still be watched.
func TestNestedDriverDir(t *testing.T) {
	// Arrange
	_, fs, watcher, _ := initTestEnvironment(t)
	// Assert
	assert.Equal(t, 2, len(watcher.watches)) // 2 from initial setup

	// test add testDriverName
	testDriverName := "testDriverName"
	testDriverPath := path.Join(pluginDir, testDriverName)
	fs.MkdirAll(testDriverPath, 0666)
	watcher.TriggerEvent(fsnotify.Create, testDriverPath)
	// Assert
	assert.Equal(t, 3, len(watcher.watches)) // 2 from initial setup, 1 from new watch.
	assert.Equal(t, testDriverPath, watcher.watches[len(watcher.watches)-1])

	// test add nested subdir inside testDriverName
	basePath := testDriverPath
	for i := 0; i < 10; i++ {
		subdirName := "subdirName"
		subdirPath := path.Join(basePath, subdirName)
		fs.MkdirAll(subdirPath, 0666)
		watcher.TriggerEvent(fsnotify.Create, subdirPath)
		// Assert
		assert.Equal(t, 4+i, len(watcher.watches)) // 3 + newly added
		assert.Equal(t, subdirPath, watcher.watches[len(watcher.watches)-1])
		basePath = subdirPath
	}
}

// Issue multiple events and probe multiple times.
func TestProberMultipleEvents(t *testing.T) {
	const iterations = 5

	// Arrange
	_, fs, watcher, prober := initTestEnvironment(t)
	for i := 0; i < iterations; i++ {
		newDriver := fmt.Sprintf("multi-event-driver%d", 1)
		installDriver(newDriver, fs)
		driverPath := path.Join(pluginDir, newDriver)
		watcher.TriggerEvent(fsnotify.Create, driverPath)
		watcher.TriggerEvent(fsnotify.Create, path.Join(driverPath, newDriver))
	}

	// Act
	events, err := prober.Probe()

	// Assert
	assert.Equal(t, 2, len(events))
	assert.Equal(t, volume.ProbeAddOrUpdate, events[0].Op)
	assert.Equal(t, volume.ProbeAddOrUpdate, events[1].Op)
	assert.NoError(t, err)
	for i := 0; i < iterations-1; i++ {
		events, err = prober.Probe()
		assert.Equal(t, 0, len(events))
		assert.NoError(t, err)
	}
}

func TestProberError(t *testing.T) {
	fs := utilfs.NewFakeFs()
	watcher := NewFakeWatcher()
	prober := &flexVolumeProber{
		pluginDir: pluginDir,
		watcher:   watcher,
		fs:        fs,
		factory:   fakePluginFactory{error: true},
	}
	installDriver(driverName, fs)
	prober.Init()

	_, err := prober.Probe()
	assert.Error(t, err)
}

// Installs a mock driver (an empty file) in the mock fs.
func installDriver(driverName string, fs utilfs.Filesystem) {
	driverPath := path.Join(pluginDir, driverName)
	fs.MkdirAll(driverPath, 0666)
	fs.Create(path.Join(driverPath, driverName))
}

// Initializes mocks, installs a single driver in the mock fs, then initializes prober.
func initTestEnvironment(t *testing.T) (
	driverPath string,
	fs utilfs.Filesystem,
	watcher *fakeWatcher,
	prober volume.DynamicPluginProber) {
	fs = utilfs.NewFakeFs()
	watcher = NewFakeWatcher()
	prober = &flexVolumeProber{
		pluginDir: pluginDir,
		watcher:   watcher,
		fs:        fs,
		factory:   fakePluginFactory{error: false},
	}
	driverPath = path.Join(pluginDir, driverName)
	installDriver(driverName, fs)
	prober.Init()

	assert.NotNilf(t, watcher.eventHandler,
		"Expect watch event handler to be registered after prober init, but is not.")
	return
}

// Fake Flexvolume plugin
type fakePluginFactory struct {
	error bool // Indicates whether an error should be returned.
}

var _ PluginFactory = fakePluginFactory{}

func (m fakePluginFactory) NewFlexVolumePlugin(_, driverName string) (volume.VolumePlugin, error) {
	if m.error {
		return nil, fmt.Errorf("Flexvolume plugin error")
	}
	// Dummy Flexvolume plugin. Prober never interacts with the plugin.
	return &flexVolumePlugin{driverName: driverName}, nil
}
