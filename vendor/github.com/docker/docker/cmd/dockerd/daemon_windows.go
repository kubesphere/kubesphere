package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/docker/docker/libcontainerd"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
)

var defaultDaemonConfigFile = ""

// setDefaultUmask doesn't do anything on windows
func setDefaultUmask() error {
	return nil
}

func getDaemonConfDir(root string) string {
	return filepath.Join(root, `\config`)
}

// preNotifySystem sends a message to the host when the API is active, but before the daemon is
func preNotifySystem() {
	// start the service now to prevent timeouts waiting for daemon to start
	// but still (eventually) complete all requests that are sent after this
	if service != nil {
		err := service.started()
		if err != nil {
			logrus.Fatal(err)
		}
	}
}

// notifySystem sends a message to the host when the server is ready to be used
func notifySystem() {
}

// notifyShutdown is called after the daemon shuts down but before the process exits.
func notifyShutdown(err error) {
	if service != nil {
		if err != nil {
			logrus.Fatal(err)
		}
		service.stopped(err)
	}
}

func (cli *DaemonCli) getPlatformRemoteOptions() ([]libcontainerd.RemoteOption, error) {
	return nil, nil
}

// setupConfigReloadTrap configures a Win32 event to reload the configuration.
func (cli *DaemonCli) setupConfigReloadTrap() {
	go func() {
		sa := windows.SecurityAttributes{
			Length: 0,
		}
		event := "Global\\docker-daemon-config-" + fmt.Sprint(os.Getpid())
		ev, _ := windows.UTF16PtrFromString(event)
		if h, _ := windows.CreateEvent(&sa, 0, 0, ev); h != 0 {
			logrus.Debugf("Config reload - waiting signal at %s", event)
			for {
				windows.WaitForSingleObject(h, windows.INFINITE)
				cli.reloadConfig()
			}
		}
	}()
}

// getSwarmRunRoot gets the root directory for swarm to store runtime state
// For example, the control socket
func (cli *DaemonCli) getSwarmRunRoot() string {
	return ""
}

func allocateDaemonPort(addr string) error {
	return nil
}

func wrapListeners(proto string, ls []net.Listener) []net.Listener {
	return ls
}
