// +build windows

package client

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Microsoft/hcsshim"
	"github.com/sirupsen/logrus"
)

// Mode is the operational mode, both requested, and actual after verification
type Mode uint

const (
	// Constants for the actual mode after validation

	// ModeActualError means an error has occurred during validation
	ModeActualError = iota
	// ModeActualVhdx means that we are going to use VHDX boot after validation
	ModeActualVhdx
	// ModeActualKernelInitrd means that we are going to use kernel+initrd for boot after validation
	ModeActualKernelInitrd

	// Constants for the requested mode

	// ModeRequestAuto means auto-select the boot mode for a utility VM
	ModeRequestAuto = iota // VHDX will be priority over kernel+initrd
	// ModeRequestVhdx means request VHDX boot if possible
	ModeRequestVhdx
	// ModeRequestKernelInitrd means request Kernel+initrd boot if possible
	ModeRequestKernelInitrd

	// defaultUvmTimeoutSeconds is the default time to wait for utility VM operations
	defaultUvmTimeoutSeconds = 5 * 60

	// DefaultVhdxSizeGB is the size of the default sandbox & scratch in GB
	DefaultVhdxSizeGB = 20

	// defaultVhdxBlockSizeMB is the block-size for the sandbox/scratch VHDx's this package can create.
	defaultVhdxBlockSizeMB = 1
)

// Config is the structure used to configuring a utility VM. There are two ways
// of starting. Either supply a VHD, or a Kernel+Initrd. For the latter, both
// must be supplied, and both must be in the same directory.
//
// VHD is the priority.
type Config struct {
	Options                                        // Configuration options
	Name               string                      // Name of the utility VM
	RequestedMode      Mode                        // What mode is preferred when validating
	ActualMode         Mode                        // What mode was obtained during validation
	UvmTimeoutSeconds  int                         // How long to wait for the utility VM to respond in seconds
	Uvm                hcsshim.Container           // The actual container
	MappedVirtualDisks []hcsshim.MappedVirtualDisk // Data-disks to be attached
}

// Options is the structure used by a client to define configurable options for a utility VM.
type Options struct {
	KirdPath       string // Path to where kernel/initrd are found (defaults to %PROGRAMFILES%\Linux Containers)
	KernelFile     string // Kernel for Utility VM (embedded in a UEFI bootloader) - does NOT include full path, just filename
	InitrdFile     string // Initrd image for Utility VM - does NOT include full path, just filename
	Vhdx           string // VHD for booting the utility VM - is a full path
	TimeoutSeconds int    // Requested time for the utility VM to respond in seconds (may be over-ridden by environment)
	BootParameters string // Additional boot parameters for initrd booting (not VHDx)
}

// ParseOptions parses a set of K-V pairs into options used by opengcs. Note
// for consistency with the LCOW graphdriver in docker, we keep the same
// convention of an `lcow.` prefix.
func ParseOptions(options []string) (Options, error) {
	rOpts := Options{TimeoutSeconds: 0}
	for _, v := range options {
		opt := strings.SplitN(v, "=", 2)
		if len(opt) == 2 {
			switch strings.ToLower(opt[0]) {
			case "lcow.kirdpath":
				rOpts.KirdPath = opt[1]
			case "lcow.kernel":
				rOpts.KernelFile = opt[1]
			case "lcow.initrd":
				rOpts.InitrdFile = opt[1]
			case "lcow.vhdx":
				rOpts.Vhdx = opt[1]
			case "lcow.bootparameters":
				rOpts.BootParameters = opt[1]
			case "lcow.timeout":
				var err error
				if rOpts.TimeoutSeconds, err = strconv.Atoi(opt[1]); err != nil {
					return rOpts, fmt.Errorf("lcow.timeout option could not be interpreted as an integer")
				}
				if rOpts.TimeoutSeconds < 0 {
					return rOpts, fmt.Errorf("lcow.timeout option cannot be negative")
				}
			}
		}
	}

	// Set default values if not supplied
	if rOpts.KirdPath == "" {
		rOpts.KirdPath = filepath.Join(os.Getenv("ProgramFiles"), "Linux Containers")
	}
	if rOpts.Vhdx == "" {
		rOpts.Vhdx = filepath.Join(rOpts.KirdPath, `uvm.vhdx`)
	}
	if rOpts.KernelFile == "" {
		rOpts.KernelFile = `bootx64.efi`
	}
	if rOpts.InitrdFile == "" {
		rOpts.InitrdFile = `initrd.img`
	}

	return rOpts, nil
}

// GenerateDefault generates a default config from a set of options
// If baseDir is not supplied, defaults to $env:ProgramFiles\Linux Containers
func (config *Config) GenerateDefault(options []string) error {
	// Parse the options that the user supplied.
	var err error
	config.Options, err = ParseOptions(options)
	if err != nil {
		return err
	}

	// Get the timeout from the environment
	envTimeoutSeconds := 0
	envTimeout := os.Getenv("OPENGCS_UVM_TIMEOUT_SECONDS")
	if len(envTimeout) > 0 {
		var err error
		if envTimeoutSeconds, err = strconv.Atoi(envTimeout); err != nil {
			return fmt.Errorf("OPENGCS_UVM_TIMEOUT_SECONDS could not be interpreted as an integer")
		}
		if envTimeoutSeconds < 0 {
			return fmt.Errorf("OPENGCS_UVM_TIMEOUT_SECONDS cannot be negative")
		}
	}

	// Priority to the requested timeout from the options.
	if config.TimeoutSeconds != 0 {
		config.UvmTimeoutSeconds = config.TimeoutSeconds
		return nil
	}

	// Next priority, the environment
	if envTimeoutSeconds != 0 {
		config.UvmTimeoutSeconds = envTimeoutSeconds
		return nil
	}

	// Last priority is the default timeout
	config.UvmTimeoutSeconds = defaultUvmTimeoutSeconds

	// Set the default requested mode
	config.RequestedMode = ModeRequestAuto

	return nil
}

// Validate validates a Config structure for starting a utility VM.
func (config *Config) Validate() error {
	config.ActualMode = ModeActualError

	if config.RequestedMode == ModeRequestVhdx && config.Vhdx == "" {
		return fmt.Errorf("VHDx mode must supply a VHDx")
	}
	if config.RequestedMode == ModeRequestKernelInitrd && (config.KernelFile == "" || config.InitrdFile == "") {
		return fmt.Errorf("kernel+initrd mode must supply both kernel and initrd")
	}

	// Validate that if VHDX requested or auto, it exists.
	if config.RequestedMode == ModeRequestAuto || config.RequestedMode == ModeRequestVhdx {
		if _, err := os.Stat(config.Vhdx); os.IsNotExist(err) {
			if config.RequestedMode == ModeRequestVhdx {
				return fmt.Errorf("VHDx '%s' not found", config.Vhdx)
			}
		} else {
			config.ActualMode = ModeActualVhdx

			// Can't specify boot parameters with VHDx
			if config.BootParameters != "" {
				return fmt.Errorf("Boot parameters cannot be specified in VHDx mode")
			}
			return nil
		}
	}

	// So must be kernel+initrd, or auto where we fallback as the VHDX doesn't exist
	if config.InitrdFile == "" || config.KernelFile == "" {
		if config.RequestedMode == ModeRequestKernelInitrd {
			return fmt.Errorf("initrd and kernel options must be supplied")
		}
		return fmt.Errorf("opengcs: configuration is invalid")
	}

	if _, err := os.Stat(filepath.Join(config.KirdPath, config.KernelFile)); os.IsNotExist(err) {
		return fmt.Errorf("kernel '%s' not found", filepath.Join(config.KirdPath, config.KernelFile))
	}
	if _, err := os.Stat(filepath.Join(config.KirdPath, config.InitrdFile)); os.IsNotExist(err) {
		return fmt.Errorf("initrd '%s' not found", filepath.Join(config.KirdPath, config.InitrdFile))
	}

	config.ActualMode = ModeActualKernelInitrd

	// Ensure all the MappedVirtualDisks exist on the host
	for _, mvd := range config.MappedVirtualDisks {
		if _, err := os.Stat(mvd.HostPath); err != nil {
			return fmt.Errorf("mapped virtual disk '%s' not found", mvd.HostPath)
		}
		if mvd.ContainerPath == "" {
			return fmt.Errorf("mapped virtual disk '%s' requested without a container path", mvd.HostPath)
		}
	}

	return nil
}

// StartUtilityVM creates and starts a utility VM from a configuration.
func (config *Config) StartUtilityVM() error {
	logrus.Debugf("opengcs: StartUtilityVM: %+v", config)

	if err := config.Validate(); err != nil {
		return err
	}

	configuration := &hcsshim.ContainerConfig{
		HvPartition:                 true,
		Name:                        config.Name,
		SystemType:                  "container",
		ContainerType:               "linux",
		TerminateOnLastHandleClosed: true,
		MappedVirtualDisks:          config.MappedVirtualDisks,
	}

	if config.ActualMode == ModeActualVhdx {
		configuration.HvRuntime = &hcsshim.HvRuntime{
			ImagePath:          config.Vhdx,
			BootSource:         "Vhd",
			WritableBootSource: false,
		}
	} else {
		configuration.HvRuntime = &hcsshim.HvRuntime{
			ImagePath:           config.KirdPath,
			LinuxInitrdFile:     config.InitrdFile,
			LinuxKernelFile:     config.KernelFile,
			LinuxBootParameters: config.BootParameters,
		}
	}

	configurationS, _ := json.Marshal(configuration)
	logrus.Debugf("opengcs: StartUtilityVM: calling HCS with '%s'", string(configurationS))
	uvm, err := hcsshim.CreateContainer(config.Name, configuration)
	if err != nil {
		return err
	}
	logrus.Debugf("opengcs: StartUtilityVM: uvm created, starting...")
	err = uvm.Start()
	if err != nil {
		logrus.Debugf("opengcs: StartUtilityVM: uvm failed to start: %s", err)
		// Make sure we don't leave it laying around as it's been created in HCS
		uvm.Terminate()
		return err
	}

	config.Uvm = uvm
	logrus.Debugf("opengcs StartUtilityVM: uvm %s is running", config.Name)
	return nil
}
