// +build !windows

/*
Copyright 2014 The Kubernetes Authors.

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

// Package app does all of the work necessary to configure and run a
// Kubernetes app process.
package app

import (
	"errors"
	"fmt"
	"net"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/features"
	"k8s.io/kubernetes/pkg/proxy"
	proxyconfigapi "k8s.io/kubernetes/pkg/proxy/apis/kubeproxyconfig"
	proxyconfig "k8s.io/kubernetes/pkg/proxy/config"
	"k8s.io/kubernetes/pkg/proxy/healthcheck"
	"k8s.io/kubernetes/pkg/proxy/iptables"
	"k8s.io/kubernetes/pkg/proxy/ipvs"
	"k8s.io/kubernetes/pkg/proxy/metrics"
	"k8s.io/kubernetes/pkg/proxy/userspace"
	"k8s.io/kubernetes/pkg/util/configz"
	utildbus "k8s.io/kubernetes/pkg/util/dbus"
	utilipset "k8s.io/kubernetes/pkg/util/ipset"
	utiliptables "k8s.io/kubernetes/pkg/util/iptables"
	utilipvs "k8s.io/kubernetes/pkg/util/ipvs"
	utilnode "k8s.io/kubernetes/pkg/util/node"
	utilsysctl "k8s.io/kubernetes/pkg/util/sysctl"
	"k8s.io/utils/exec"

	"github.com/golang/glog"
)

// NewProxyServer returns a new ProxyServer.
func NewProxyServer(o *Options) (*ProxyServer, error) {
	return newProxyServer(o.config, o.CleanupAndExit, o.CleanupIPVS, o.scheme, o.master)
}

func newProxyServer(
	config *proxyconfigapi.KubeProxyConfiguration,
	cleanupAndExit bool,
	cleanupIPVS bool,
	scheme *runtime.Scheme,
	master string) (*ProxyServer, error) {

	if config == nil {
		return nil, errors.New("config is required")
	}

	if c, err := configz.New(proxyconfigapi.GroupName); err == nil {
		c.Set(config)
	} else {
		return nil, fmt.Errorf("unable to register configz: %s", err)
	}

	protocol := utiliptables.ProtocolIpv4
	if net.ParseIP(config.BindAddress).To4() == nil {
		glog.V(0).Infof("IPv6 bind address (%s), assume IPv6 operation", config.BindAddress)
		protocol = utiliptables.ProtocolIpv6
	}

	var iptInterface utiliptables.Interface
	var ipvsInterface utilipvs.Interface
	var kernelHandler ipvs.KernelHandler
	var ipsetInterface utilipset.Interface
	var dbus utildbus.Interface

	// Create a iptables utils.
	execer := exec.New()

	dbus = utildbus.New()
	iptInterface = utiliptables.New(execer, dbus, protocol)
	ipvsInterface = utilipvs.New(execer)
	kernelHandler = ipvs.NewLinuxKernelHandler()
	ipsetInterface = utilipset.New(execer)

	// We omit creation of pretty much everything if we run in cleanup mode
	if cleanupAndExit {
		return &ProxyServer{
			execer:         execer,
			IptInterface:   iptInterface,
			IpvsInterface:  ipvsInterface,
			IpsetInterface: ipsetInterface,
			CleanupAndExit: cleanupAndExit,
		}, nil
	}

	client, eventClient, err := createClients(config.ClientConnection, master)
	if err != nil {
		return nil, err
	}

	// Create event recorder
	hostname := utilnode.GetHostname(config.HostnameOverride)
	eventBroadcaster := record.NewBroadcaster()
	recorder := eventBroadcaster.NewRecorder(scheme, v1.EventSource{Component: "kube-proxy", Host: hostname})

	nodeRef := &v1.ObjectReference{
		Kind:      "Node",
		Name:      hostname,
		UID:       types.UID(hostname),
		Namespace: "",
	}

	var healthzServer *healthcheck.HealthzServer
	var healthzUpdater healthcheck.HealthzUpdater
	if len(config.HealthzBindAddress) > 0 {
		healthzServer = healthcheck.NewDefaultHealthzServer(config.HealthzBindAddress, 2*config.IPTables.SyncPeriod.Duration, recorder, nodeRef)
		healthzUpdater = healthzServer
	}

	var proxier proxy.ProxyProvider
	var serviceEventHandler proxyconfig.ServiceHandler
	var endpointsEventHandler proxyconfig.EndpointsHandler

	proxyMode := getProxyMode(string(config.Mode), iptInterface, kernelHandler, ipsetInterface, iptables.LinuxKernelCompatTester{})
	if proxyMode == proxyModeIPTables {
		glog.V(0).Info("Using iptables Proxier.")
		nodeIP := net.ParseIP(config.BindAddress)
		if nodeIP.Equal(net.IPv4zero) || nodeIP.Equal(net.IPv6zero) {
			nodeIP = getNodeIP(client, hostname)
		}
		if config.IPTables.MasqueradeBit == nil {
			// MasqueradeBit must be specified or defaulted.
			return nil, fmt.Errorf("unable to read IPTables MasqueradeBit from config")
		}

		// TODO this has side effects that should only happen when Run() is invoked.
		proxierIPTables, err := iptables.NewProxier(
			iptInterface,
			utilsysctl.New(),
			execer,
			config.IPTables.SyncPeriod.Duration,
			config.IPTables.MinSyncPeriod.Duration,
			config.IPTables.MasqueradeAll,
			int(*config.IPTables.MasqueradeBit),
			config.ClusterCIDR,
			hostname,
			nodeIP,
			recorder,
			healthzUpdater,
			config.NodePortAddresses,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to create proxier: %v", err)
		}
		metrics.RegisterMetrics()
		proxier = proxierIPTables
		serviceEventHandler = proxierIPTables
		endpointsEventHandler = proxierIPTables
		// No turning back. Remove artifacts that might still exist from the userspace Proxier.
		glog.V(0).Info("Tearing down inactive rules.")
		// TODO this has side effects that should only happen when Run() is invoked.
		userspace.CleanupLeftovers(iptInterface)
		// IPVS Proxier will generate some iptables rules, need to clean them before switching to other proxy mode.
		// Besides, ipvs proxier will create some ipvs rules as well.  Because there is no way to tell if a given
		// ipvs rule is created by IPVS proxier or not.  Users should explicitly specify `--clean-ipvs=true` to flush
		// all ipvs rules when kube-proxy start up.  Users do this operation should be with caution.
		ipvs.CleanupLeftovers(ipvsInterface, iptInterface, ipsetInterface, cleanupIPVS)
	} else if proxyMode == proxyModeIPVS {
		glog.V(0).Info("Using ipvs Proxier.")
		proxierIPVS, err := ipvs.NewProxier(
			iptInterface,
			ipvsInterface,
			ipsetInterface,
			utilsysctl.New(),
			execer,
			config.IPVS.SyncPeriod.Duration,
			config.IPVS.MinSyncPeriod.Duration,
			config.IPVS.ExcludeCIDRs,
			config.IPTables.MasqueradeAll,
			int(*config.IPTables.MasqueradeBit),
			config.ClusterCIDR,
			hostname,
			getNodeIP(client, hostname),
			recorder,
			healthzServer,
			config.IPVS.Scheduler,
			config.NodePortAddresses,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to create proxier: %v", err)
		}
		metrics.RegisterMetrics()
		proxier = proxierIPVS
		serviceEventHandler = proxierIPVS
		endpointsEventHandler = proxierIPVS
		glog.V(0).Info("Tearing down inactive rules.")
		// TODO this has side effects that should only happen when Run() is invoked.
		userspace.CleanupLeftovers(iptInterface)
		iptables.CleanupLeftovers(iptInterface)
	} else {
		glog.V(0).Info("Using userspace Proxier.")
		// This is a proxy.LoadBalancer which NewProxier needs but has methods we don't need for
		// our config.EndpointsConfigHandler.
		loadBalancer := userspace.NewLoadBalancerRR()
		// set EndpointsConfigHandler to our loadBalancer
		endpointsEventHandler = loadBalancer

		// TODO this has side effects that should only happen when Run() is invoked.
		proxierUserspace, err := userspace.NewProxier(
			loadBalancer,
			net.ParseIP(config.BindAddress),
			iptInterface,
			execer,
			*utilnet.ParsePortRangeOrDie(config.PortRange),
			config.IPTables.SyncPeriod.Duration,
			config.IPTables.MinSyncPeriod.Duration,
			config.UDPIdleTimeout.Duration,
			config.NodePortAddresses,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to create proxier: %v", err)
		}
		serviceEventHandler = proxierUserspace
		proxier = proxierUserspace

		// Remove artifacts from the iptables and ipvs Proxier, if not on Windows.
		glog.V(0).Info("Tearing down inactive rules.")
		// TODO this has side effects that should only happen when Run() is invoked.
		iptables.CleanupLeftovers(iptInterface)
		// IPVS Proxier will generate some iptables rules, need to clean them before switching to other proxy mode.
		// Besides, ipvs proxier will create some ipvs rules as well.  Because there is no way to tell if a given
		// ipvs rule is created by IPVS proxier or not.  Users should explicitly specify `--clean-ipvs=true` to flush
		// all ipvs rules when kube-proxy start up.  Users do this operation should be with caution.
		ipvs.CleanupLeftovers(ipvsInterface, iptInterface, ipsetInterface, cleanupIPVS)
	}

	iptInterface.AddReloadFunc(proxier.Sync)

	return &ProxyServer{
		Client:                 client,
		EventClient:            eventClient,
		IptInterface:           iptInterface,
		IpvsInterface:          ipvsInterface,
		IpsetInterface:         ipsetInterface,
		execer:                 execer,
		Proxier:                proxier,
		Broadcaster:            eventBroadcaster,
		Recorder:               recorder,
		ConntrackConfiguration: config.Conntrack,
		Conntracker:            &realConntracker{},
		ProxyMode:              proxyMode,
		NodeRef:                nodeRef,
		MetricsBindAddress:     config.MetricsBindAddress,
		EnableProfiling:        config.EnableProfiling,
		OOMScoreAdj:            config.OOMScoreAdj,
		ResourceContainer:      config.ResourceContainer,
		ConfigSyncPeriod:       config.ConfigSyncPeriod.Duration,
		ServiceEventHandler:    serviceEventHandler,
		EndpointsEventHandler:  endpointsEventHandler,
		HealthzServer:          healthzServer,
	}, nil
}

func getProxyMode(proxyMode string, iptver iptables.IPTablesVersioner, khandle ipvs.KernelHandler, ipsetver ipvs.IPSetVersioner, kcompat iptables.KernelCompatTester) string {
	if proxyMode == proxyModeUserspace {
		return proxyModeUserspace
	}

	if len(proxyMode) > 0 && proxyMode == proxyModeIPTables {
		return tryIPTablesProxy(iptver, kcompat)
	}

	if utilfeature.DefaultFeatureGate.Enabled(features.SupportIPVSProxyMode) {
		if proxyMode == proxyModeIPVS {
			return tryIPVSProxy(iptver, khandle, ipsetver, kcompat)
		} else {
			glog.Warningf("Can't use ipvs proxier, trying iptables proxier")
			return tryIPTablesProxy(iptver, kcompat)
		}
	}
	glog.Warningf("Flag proxy-mode=%q unknown, assuming iptables proxy", proxyMode)
	return tryIPTablesProxy(iptver, kcompat)
}

func tryIPVSProxy(iptver iptables.IPTablesVersioner, khandle ipvs.KernelHandler, ipsetver ipvs.IPSetVersioner, kcompat iptables.KernelCompatTester) string {
	// guaranteed false on error, error only necessary for debugging
	// IPVS Proxier relies on ip_vs_* kernel modules and ipset
	useIPVSProxy, err := ipvs.CanUseIPVSProxier(khandle, ipsetver)
	if err != nil {
		// Try to fallback to iptables before falling back to userspace
		utilruntime.HandleError(fmt.Errorf("can't determine whether to use ipvs proxy, error: %v", err))
	}
	if useIPVSProxy {
		return proxyModeIPVS
	}

	// Try to fallback to iptables before falling back to userspace
	glog.V(1).Infof("Can't use ipvs proxier, trying iptables proxier")
	return tryIPTablesProxy(iptver, kcompat)
}

func tryIPTablesProxy(iptver iptables.IPTablesVersioner, kcompat iptables.KernelCompatTester) string {
	// guaranteed false on error, error only necessary for debugging
	useIPTablesProxy, err := iptables.CanUseIPTablesProxier(iptver, kcompat)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("can't determine whether to use iptables proxy, using userspace proxier: %v", err))
		return proxyModeUserspace
	}
	if useIPTablesProxy {
		return proxyModeIPTables
	}
	// Fallback.
	glog.V(1).Infof("Can't use iptables proxy, using userspace proxier")
	return proxyModeUserspace
}
