package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/docker/distribution/uuid"
	"github.com/docker/docker/api"
	apiserver "github.com/docker/docker/api/server"
	buildbackend "github.com/docker/docker/api/server/backend/build"
	"github.com/docker/docker/api/server/middleware"
	"github.com/docker/docker/api/server/router"
	"github.com/docker/docker/api/server/router/build"
	checkpointrouter "github.com/docker/docker/api/server/router/checkpoint"
	"github.com/docker/docker/api/server/router/container"
	distributionrouter "github.com/docker/docker/api/server/router/distribution"
	"github.com/docker/docker/api/server/router/image"
	"github.com/docker/docker/api/server/router/network"
	pluginrouter "github.com/docker/docker/api/server/router/plugin"
	sessionrouter "github.com/docker/docker/api/server/router/session"
	swarmrouter "github.com/docker/docker/api/server/router/swarm"
	systemrouter "github.com/docker/docker/api/server/router/system"
	"github.com/docker/docker/api/server/router/volume"
	"github.com/docker/docker/builder/dockerfile"
	"github.com/docker/docker/builder/fscache"
	"github.com/docker/docker/cli/debug"
	"github.com/docker/docker/daemon"
	"github.com/docker/docker/daemon/cluster"
	"github.com/docker/docker/daemon/config"
	"github.com/docker/docker/daemon/listeners"
	"github.com/docker/docker/dockerversion"
	"github.com/docker/docker/libcontainerd"
	dopts "github.com/docker/docker/opts"
	"github.com/docker/docker/pkg/authorization"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/pidfile"
	"github.com/docker/docker/pkg/plugingetter"
	"github.com/docker/docker/pkg/signal"
	"github.com/docker/docker/pkg/system"
	"github.com/docker/docker/plugin"
	"github.com/docker/docker/registry"
	"github.com/docker/docker/runconfig"
	"github.com/docker/go-connections/tlsconfig"
	swarmapi "github.com/docker/swarmkit/api"
	"github.com/moby/buildkit/session"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

// DaemonCli represents the daemon CLI.
type DaemonCli struct {
	*config.Config
	configFile *string
	flags      *pflag.FlagSet

	api             *apiserver.Server
	d               *daemon.Daemon
	authzMiddleware *authorization.Middleware // authzMiddleware enables to dynamically reload the authorization plugins
}

// NewDaemonCli returns a daemon CLI
func NewDaemonCli() *DaemonCli {
	return &DaemonCli{}
}

func (cli *DaemonCli) start(opts *daemonOptions) (err error) {
	stopc := make(chan bool)
	defer close(stopc)

	// warn from uuid package when running the daemon
	uuid.Loggerf = logrus.Warnf

	opts.SetDefaultOptions(opts.flags)

	if cli.Config, err = loadDaemonCliConfig(opts); err != nil {
		return err
	}
	cli.configFile = &opts.configFile
	cli.flags = opts.flags

	if cli.Config.Debug {
		debug.Enable()
	}

	if cli.Config.Experimental {
		logrus.Warn("Running experimental build")
	}

	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: jsonmessage.RFC3339NanoFixed,
		DisableColors:   cli.Config.RawLogs,
		FullTimestamp:   true,
	})

	system.InitLCOW(cli.Config.Experimental)

	if err := setDefaultUmask(); err != nil {
		return fmt.Errorf("Failed to set umask: %v", err)
	}

	// Create the daemon root before we create ANY other files (PID, or migrate keys)
	// to ensure the appropriate ACL is set (particularly relevant on Windows)
	if err := daemon.CreateDaemonRoot(cli.Config); err != nil {
		return err
	}

	if cli.Pidfile != "" {
		pf, err := pidfile.New(cli.Pidfile)
		if err != nil {
			return fmt.Errorf("Error starting daemon: %v", err)
		}
		defer func() {
			if err := pf.Remove(); err != nil {
				logrus.Error(err)
			}
		}()
	}

	serverConfig, err := newAPIServerConfig(cli)
	if err != nil {
		return fmt.Errorf("Failed to create API server: %v", err)
	}
	cli.api = apiserver.New(serverConfig)

	hosts, err := loadListeners(cli, serverConfig)
	if err != nil {
		return fmt.Errorf("Failed to load listeners: %v", err)
	}

	registryService, err := registry.NewService(cli.Config.ServiceOptions)
	if err != nil {
		return err
	}

	rOpts, err := cli.getRemoteOptions()
	if err != nil {
		return fmt.Errorf("Failed to generate containerd options: %v", err)
	}
	containerdRemote, err := libcontainerd.New(filepath.Join(cli.Config.Root, "containerd"), filepath.Join(cli.Config.ExecRoot, "containerd"), rOpts...)
	if err != nil {
		return err
	}
	signal.Trap(func() {
		cli.stop()
		<-stopc // wait for daemonCli.start() to return
	}, logrus.StandardLogger())

	// Notify that the API is active, but before daemon is set up.
	preNotifySystem()

	pluginStore := plugin.NewStore()

	if err := cli.initMiddlewares(cli.api, serverConfig, pluginStore); err != nil {
		logrus.Fatalf("Error creating middlewares: %v", err)
	}

	d, err := daemon.NewDaemon(cli.Config, registryService, containerdRemote, pluginStore)
	if err != nil {
		return fmt.Errorf("Error starting daemon: %v", err)
	}

	d.StoreHosts(hosts)

	// validate after NewDaemon has restored enabled plugins. Dont change order.
	if err := validateAuthzPlugins(cli.Config.AuthorizationPlugins, pluginStore); err != nil {
		return fmt.Errorf("Error validating authorization plugin: %v", err)
	}

	// TODO: move into startMetricsServer()
	if cli.Config.MetricsAddress != "" {
		if !d.HasExperimental() {
			return fmt.Errorf("metrics-addr is only supported when experimental is enabled")
		}
		if err := startMetricsServer(cli.Config.MetricsAddress); err != nil {
			return err
		}
	}

	c, err := createAndStartCluster(cli, d)
	if err != nil {
		logrus.Fatalf("Error starting cluster component: %v", err)
	}

	// Restart all autostart containers which has a swarm endpoint
	// and is not yet running now that we have successfully
	// initialized the cluster.
	d.RestartSwarmContainers()

	logrus.Info("Daemon has completed initialization")

	cli.d = d

	routerOptions, err := newRouterOptions(cli.Config, d)
	if err != nil {
		return err
	}
	routerOptions.api = cli.api
	routerOptions.cluster = c

	initRouter(routerOptions)

	// process cluster change notifications
	watchCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go d.ProcessClusterNotifications(watchCtx, c.GetWatchStream())

	cli.setupConfigReloadTrap()

	// The serve API routine never exits unless an error occurs
	// We need to start it as a goroutine and wait on it so
	// daemon doesn't exit
	serveAPIWait := make(chan error)
	go cli.api.Wait(serveAPIWait)

	// after the daemon is done setting up we can notify systemd api
	notifySystem()

	// Daemon is fully initialized and handling API traffic
	// Wait for serve API to complete
	errAPI := <-serveAPIWait
	c.Cleanup()
	shutdownDaemon(d)
	containerdRemote.Cleanup()
	if errAPI != nil {
		return fmt.Errorf("Shutting down due to ServeAPI error: %v", errAPI)
	}

	return nil
}

type routerOptions struct {
	sessionManager *session.Manager
	buildBackend   *buildbackend.Backend
	buildCache     *fscache.FSCache
	daemon         *daemon.Daemon
	api            *apiserver.Server
	cluster        *cluster.Cluster
}

func newRouterOptions(config *config.Config, daemon *daemon.Daemon) (routerOptions, error) {
	opts := routerOptions{}
	sm, err := session.NewManager()
	if err != nil {
		return opts, errors.Wrap(err, "failed to create sessionmanager")
	}

	builderStateDir := filepath.Join(config.Root, "builder")

	buildCache, err := fscache.NewFSCache(fscache.Opt{
		Backend: fscache.NewNaiveCacheBackend(builderStateDir),
		Root:    builderStateDir,
		GCPolicy: fscache.GCPolicy{ // TODO: expose this in config
			MaxSize:         1024 * 1024 * 512,  // 512MB
			MaxKeepDuration: 7 * 24 * time.Hour, // 1 week
		},
	})
	if err != nil {
		return opts, errors.Wrap(err, "failed to create fscache")
	}

	manager, err := dockerfile.NewBuildManager(daemon.BuilderBackend(), sm, buildCache, daemon.IDMappings())
	if err != nil {
		return opts, err
	}

	bb, err := buildbackend.NewBackend(daemon.ImageService(), manager, buildCache)
	if err != nil {
		return opts, errors.Wrap(err, "failed to create buildmanager")
	}

	return routerOptions{
		sessionManager: sm,
		buildBackend:   bb,
		buildCache:     buildCache,
		daemon:         daemon,
	}, nil
}

func (cli *DaemonCli) reloadConfig() {
	reload := func(c *config.Config) {

		// Revalidate and reload the authorization plugins
		if err := validateAuthzPlugins(c.AuthorizationPlugins, cli.d.PluginStore); err != nil {
			logrus.Fatalf("Error validating authorization plugin: %v", err)
			return
		}
		cli.authzMiddleware.SetPlugins(c.AuthorizationPlugins)

		// The namespaces com.docker.*, io.docker.*, org.dockerproject.* have been documented
		// to be reserved for Docker's internal use, but this was never enforced.  Allowing
		// configured labels to use these namespaces are deprecated for 18.05.
		//
		// The following will check the usage of such labels, and report a warning for deprecation.
		//
		// TODO: At the next stable release, the validation should be folded into the other
		// configuration validation functions and an error will be returned instead, and this
		// block should be deleted.
		if err := config.ValidateReservedNamespaceLabels(c.Labels); err != nil {
			logrus.Warnf("Configured labels using reserved namespaces is deprecated: %s", err)
		}

		if err := cli.d.Reload(c); err != nil {
			logrus.Errorf("Error reconfiguring the daemon: %v", err)
			return
		}

		if c.IsValueSet("debug") {
			debugEnabled := debug.IsEnabled()
			switch {
			case debugEnabled && !c.Debug: // disable debug
				debug.Disable()
			case c.Debug && !debugEnabled: // enable debug
				debug.Enable()
			}
		}
	}

	if err := config.Reload(*cli.configFile, cli.flags, reload); err != nil {
		logrus.Error(err)
	}
}

func (cli *DaemonCli) stop() {
	cli.api.Close()
}

// shutdownDaemon just wraps daemon.Shutdown() to handle a timeout in case
// d.Shutdown() is waiting too long to kill container or worst it's
// blocked there
func shutdownDaemon(d *daemon.Daemon) {
	shutdownTimeout := d.ShutdownTimeout()
	ch := make(chan struct{})
	go func() {
		d.Shutdown()
		close(ch)
	}()
	if shutdownTimeout < 0 {
		<-ch
		logrus.Debug("Clean shutdown succeeded")
		return
	}
	select {
	case <-ch:
		logrus.Debug("Clean shutdown succeeded")
	case <-time.After(time.Duration(shutdownTimeout) * time.Second):
		logrus.Error("Force shutdown daemon")
	}
}

func loadDaemonCliConfig(opts *daemonOptions) (*config.Config, error) {
	conf := opts.daemonConfig
	flags := opts.flags
	conf.Debug = opts.Debug
	conf.Hosts = opts.Hosts
	conf.LogLevel = opts.LogLevel
	conf.TLS = opts.TLS
	conf.TLSVerify = opts.TLSVerify
	conf.CommonTLSOptions = config.CommonTLSOptions{}

	if opts.TLSOptions != nil {
		conf.CommonTLSOptions.CAFile = opts.TLSOptions.CAFile
		conf.CommonTLSOptions.CertFile = opts.TLSOptions.CertFile
		conf.CommonTLSOptions.KeyFile = opts.TLSOptions.KeyFile
	}

	if conf.TrustKeyPath == "" {
		conf.TrustKeyPath = filepath.Join(
			getDaemonConfDir(conf.Root),
			defaultTrustKeyFile)
	}

	if flags.Changed("graph") && flags.Changed("data-root") {
		return nil, fmt.Errorf(`cannot specify both "--graph" and "--data-root" option`)
	}

	if opts.configFile != "" {
		c, err := config.MergeDaemonConfigurations(conf, flags, opts.configFile)
		if err != nil {
			if flags.Changed("config-file") || !os.IsNotExist(err) {
				return nil, fmt.Errorf("unable to configure the Docker daemon with file %s: %v", opts.configFile, err)
			}
		}
		// the merged configuration can be nil if the config file didn't exist.
		// leave the current configuration as it is if when that happens.
		if c != nil {
			conf = c
		}
	}

	if err := config.Validate(conf); err != nil {
		return nil, err
	}

	if runtime.GOOS != "windows" {
		if flags.Changed("disable-legacy-registry") {
			// TODO: Remove this error after 3 release cycles (18.03)
			return nil, errors.New("ERROR: The '--disable-legacy-registry' flag has been removed. Interacting with legacy (v1) registries is no longer supported")
		}
		if !conf.V2Only {
			// TODO: Remove this error after 3 release cycles (18.03)
			return nil, errors.New("ERROR: The 'disable-legacy-registry' configuration option has been removed. Interacting with legacy (v1) registries is no longer supported")
		}
	}

	if flags.Changed("graph") {
		logrus.Warnf(`The "-g / --graph" flag is deprecated. Please use "--data-root" instead`)
	}

	// Check if duplicate label-keys with different values are found
	newLabels, err := config.GetConflictFreeLabels(conf.Labels)
	if err != nil {
		return nil, err
	}
	// The namespaces com.docker.*, io.docker.*, org.dockerproject.* have been documented
	// to be reserved for Docker's internal use, but this was never enforced.  Allowing
	// configured labels to use these namespaces are deprecated for 18.05.
	//
	// The following will check the usage of such labels, and report a warning for deprecation.
	//
	// TODO: At the next stable release, the validation should be folded into the other
	// configuration validation functions and an error will be returned instead, and this
	// block should be deleted.
	if err := config.ValidateReservedNamespaceLabels(newLabels); err != nil {
		logrus.Warnf("Configured labels using reserved namespaces is deprecated: %s", err)
	}
	conf.Labels = newLabels

	// Regardless of whether the user sets it to true or false, if they
	// specify TLSVerify at all then we need to turn on TLS
	if conf.IsValueSet(FlagTLSVerify) {
		conf.TLS = true
	}

	// ensure that the log level is the one set after merging configurations
	setLogLevel(conf.LogLevel)

	return conf, nil
}

func initRouter(opts routerOptions) {
	decoder := runconfig.ContainerDecoder{}

	routers := []router.Router{
		// we need to add the checkpoint router before the container router or the DELETE gets masked
		checkpointrouter.NewRouter(opts.daemon, decoder),
		container.NewRouter(opts.daemon, decoder),
		image.NewRouter(opts.daemon.ImageService()),
		systemrouter.NewRouter(opts.daemon, opts.cluster, opts.buildCache),
		volume.NewRouter(opts.daemon),
		build.NewRouter(opts.buildBackend, opts.daemon),
		sessionrouter.NewRouter(opts.sessionManager),
		swarmrouter.NewRouter(opts.cluster),
		pluginrouter.NewRouter(opts.daemon.PluginManager()),
		distributionrouter.NewRouter(opts.daemon.ImageService()),
	}

	if opts.daemon.NetworkControllerEnabled() {
		routers = append(routers, network.NewRouter(opts.daemon, opts.cluster))
	}

	if opts.daemon.HasExperimental() {
		for _, r := range routers {
			for _, route := range r.Routes() {
				if experimental, ok := route.(router.ExperimentalRoute); ok {
					experimental.Enable()
				}
			}
		}
	}

	opts.api.InitRouter(routers...)
}

// TODO: remove this from cli and return the authzMiddleware
func (cli *DaemonCli) initMiddlewares(s *apiserver.Server, cfg *apiserver.Config, pluginStore plugingetter.PluginGetter) error {
	v := cfg.Version

	exp := middleware.NewExperimentalMiddleware(cli.Config.Experimental)
	s.UseMiddleware(exp)

	vm := middleware.NewVersionMiddleware(v, api.DefaultVersion, api.MinVersion)
	s.UseMiddleware(vm)

	if cfg.CorsHeaders != "" {
		c := middleware.NewCORSMiddleware(cfg.CorsHeaders)
		s.UseMiddleware(c)
	}

	cli.authzMiddleware = authorization.NewMiddleware(cli.Config.AuthorizationPlugins, pluginStore)
	cli.Config.AuthzMiddleware = cli.authzMiddleware
	s.UseMiddleware(cli.authzMiddleware)
	return nil
}

func (cli *DaemonCli) getRemoteOptions() ([]libcontainerd.RemoteOption, error) {
	opts := []libcontainerd.RemoteOption{}

	pOpts, err := cli.getPlatformRemoteOptions()
	if err != nil {
		return nil, err
	}
	opts = append(opts, pOpts...)
	return opts, nil
}

func newAPIServerConfig(cli *DaemonCli) (*apiserver.Config, error) {
	serverConfig := &apiserver.Config{
		Logging:     true,
		SocketGroup: cli.Config.SocketGroup,
		Version:     dockerversion.Version,
		CorsHeaders: cli.Config.CorsHeaders,
	}

	if cli.Config.TLS {
		tlsOptions := tlsconfig.Options{
			CAFile:             cli.Config.CommonTLSOptions.CAFile,
			CertFile:           cli.Config.CommonTLSOptions.CertFile,
			KeyFile:            cli.Config.CommonTLSOptions.KeyFile,
			ExclusiveRootPools: true,
		}

		if cli.Config.TLSVerify {
			// server requires and verifies client's certificate
			tlsOptions.ClientAuth = tls.RequireAndVerifyClientCert
		}
		tlsConfig, err := tlsconfig.Server(tlsOptions)
		if err != nil {
			return nil, err
		}
		serverConfig.TLSConfig = tlsConfig
	}

	if len(cli.Config.Hosts) == 0 {
		cli.Config.Hosts = make([]string, 1)
	}

	return serverConfig, nil
}

func loadListeners(cli *DaemonCli, serverConfig *apiserver.Config) ([]string, error) {
	var hosts []string
	for i := 0; i < len(cli.Config.Hosts); i++ {
		var err error
		if cli.Config.Hosts[i], err = dopts.ParseHost(cli.Config.TLS, cli.Config.Hosts[i]); err != nil {
			return nil, fmt.Errorf("error parsing -H %s : %v", cli.Config.Hosts[i], err)
		}

		protoAddr := cli.Config.Hosts[i]
		protoAddrParts := strings.SplitN(protoAddr, "://", 2)
		if len(protoAddrParts) != 2 {
			return nil, fmt.Errorf("bad format %s, expected PROTO://ADDR", protoAddr)
		}

		proto := protoAddrParts[0]
		addr := protoAddrParts[1]

		// It's a bad idea to bind to TCP without tlsverify.
		if proto == "tcp" && (serverConfig.TLSConfig == nil || serverConfig.TLSConfig.ClientAuth != tls.RequireAndVerifyClientCert) {
			logrus.Warn("[!] DON'T BIND ON ANY IP ADDRESS WITHOUT setting --tlsverify IF YOU DON'T KNOW WHAT YOU'RE DOING [!]")
		}
		ls, err := listeners.Init(proto, addr, serverConfig.SocketGroup, serverConfig.TLSConfig)
		if err != nil {
			return nil, err
		}
		ls = wrapListeners(proto, ls)
		// If we're binding to a TCP port, make sure that a container doesn't try to use it.
		if proto == "tcp" {
			if err := allocateDaemonPort(addr); err != nil {
				return nil, err
			}
		}
		logrus.Debugf("Listener created for HTTP on %s (%s)", proto, addr)
		hosts = append(hosts, protoAddrParts[1])
		cli.api.Accept(addr, ls...)
	}

	return hosts, nil
}

func createAndStartCluster(cli *DaemonCli, d *daemon.Daemon) (*cluster.Cluster, error) {
	name, _ := os.Hostname()

	// Use a buffered channel to pass changes from store watch API to daemon
	// A buffer allows store watch API and daemon processing to not wait for each other
	watchStream := make(chan *swarmapi.WatchMessage, 32)

	c, err := cluster.New(cluster.Config{
		Root:                   cli.Config.Root,
		Name:                   name,
		Backend:                d,
		ImageBackend:           d.ImageService(),
		PluginBackend:          d.PluginManager(),
		NetworkSubnetsProvider: d,
		DefaultAdvertiseAddr:   cli.Config.SwarmDefaultAdvertiseAddr,
		RaftHeartbeatTick:      cli.Config.SwarmRaftHeartbeatTick,
		RaftElectionTick:       cli.Config.SwarmRaftElectionTick,
		RuntimeRoot:            cli.getSwarmRunRoot(),
		WatchStream:            watchStream,
	})
	if err != nil {
		return nil, err
	}
	d.SetCluster(c)
	err = c.Start()

	return c, err
}

// validates that the plugins requested with the --authorization-plugin flag are valid AuthzDriver
// plugins present on the host and available to the daemon
func validateAuthzPlugins(requestedPlugins []string, pg plugingetter.PluginGetter) error {
	for _, reqPlugin := range requestedPlugins {
		if _, err := pg.Get(reqPlugin, authorization.AuthZApiImplements, plugingetter.Lookup); err != nil {
			return err
		}
	}
	return nil
}
