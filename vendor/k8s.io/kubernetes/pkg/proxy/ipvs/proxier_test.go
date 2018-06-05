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

package ipvs

import (
	"bytes"
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	api "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/proxy"
	"k8s.io/utils/exec"
	fakeexec "k8s.io/utils/exec/testing"

	"k8s.io/apimachinery/pkg/util/sets"
	netlinktest "k8s.io/kubernetes/pkg/proxy/ipvs/testing"
	utilproxy "k8s.io/kubernetes/pkg/proxy/util"
	proxyutiltest "k8s.io/kubernetes/pkg/proxy/util/testing"
	utilipset "k8s.io/kubernetes/pkg/util/ipset"
	ipsettest "k8s.io/kubernetes/pkg/util/ipset/testing"
	utiliptables "k8s.io/kubernetes/pkg/util/iptables"
	iptablestest "k8s.io/kubernetes/pkg/util/iptables/testing"
	utilipvs "k8s.io/kubernetes/pkg/util/ipvs"
	ipvstest "k8s.io/kubernetes/pkg/util/ipvs/testing"
)

const testHostname = "test-hostname"

type fakeIPGetter struct {
	nodeIPs []net.IP
}

func (f *fakeIPGetter) NodeIPs() ([]net.IP, error) {
	return f.nodeIPs, nil
}

type fakeHealthChecker struct {
	services  map[types.NamespacedName]uint16
	Endpoints map[types.NamespacedName]int
}

func newFakeHealthChecker() *fakeHealthChecker {
	return &fakeHealthChecker{
		services:  map[types.NamespacedName]uint16{},
		Endpoints: map[types.NamespacedName]int{},
	}
}

// fakePortOpener implements portOpener.
type fakePortOpener struct {
	openPorts []*utilproxy.LocalPort
}

// OpenLocalPort fakes out the listen() and bind() used by syncProxyRules
// to lock a local port.
func (f *fakePortOpener) OpenLocalPort(lp *utilproxy.LocalPort) (utilproxy.Closeable, error) {
	f.openPorts = append(f.openPorts, lp)
	return nil, nil
}

func (fake *fakeHealthChecker) SyncServices(newServices map[types.NamespacedName]uint16) error {
	fake.services = newServices
	return nil
}

func (fake *fakeHealthChecker) SyncEndpoints(newEndpoints map[types.NamespacedName]int) error {
	fake.Endpoints = newEndpoints
	return nil
}

// fakeKernelHandler implements KernelHandler.
type fakeKernelHandler struct {
	modules []string
}

func (fake *fakeKernelHandler) GetModules() ([]string, error) {
	return fake.modules, nil
}

// fakeKernelHandler implements KernelHandler.
type fakeIPSetVersioner struct {
	version string
	err     error
}

func (fake *fakeIPSetVersioner) GetVersion() (string, error) {
	return fake.version, fake.err
}

func NewFakeProxier(ipt utiliptables.Interface, ipvs utilipvs.Interface, ipset utilipset.Interface, nodeIPs []net.IP) *Proxier {
	fcmd := fakeexec.FakeCmd{
		CombinedOutputScript: []fakeexec.FakeCombinedOutputAction{
			func() ([]byte, error) { return []byte("dummy device have been created"), nil },
		},
	}
	fexec := &fakeexec.FakeExec{
		CommandScript: []fakeexec.FakeCommandAction{
			func(cmd string, args ...string) exec.Cmd { return fakeexec.InitFakeCmd(&fcmd, cmd, args...) },
		},
		LookPathFunc: func(cmd string) (string, error) { return cmd, nil },
	}
	return &Proxier{
		exec:                fexec,
		serviceMap:          make(proxy.ServiceMap),
		serviceChanges:      proxy.NewServiceChangeTracker(newServiceInfo, nil, nil),
		endpointsMap:        make(proxy.EndpointsMap),
		endpointsChanges:    proxy.NewEndpointChangeTracker(testHostname, nil, nil, nil),
		excludeCIDRs:        make([]string, 0),
		iptables:            ipt,
		ipvs:                ipvs,
		ipset:               ipset,
		clusterCIDR:         "10.0.0.0/24",
		hostname:            testHostname,
		portsMap:            make(map[utilproxy.LocalPort]utilproxy.Closeable),
		portMapper:          &fakePortOpener{[]*utilproxy.LocalPort{}},
		healthChecker:       newFakeHealthChecker(),
		ipvsScheduler:       DefaultScheduler,
		ipGetter:            &fakeIPGetter{nodeIPs: nodeIPs},
		iptablesData:        bytes.NewBuffer(nil),
		natChains:           bytes.NewBuffer(nil),
		natRules:            bytes.NewBuffer(nil),
		filterChains:        bytes.NewBuffer(nil),
		filterRules:         bytes.NewBuffer(nil),
		netlinkHandle:       netlinktest.NewFakeNetlinkHandle(),
		loopbackSet:         NewIPSet(ipset, KubeLoopBackIPSet, utilipset.HashIPPortIP, false),
		clusterIPSet:        NewIPSet(ipset, KubeClusterIPSet, utilipset.HashIPPort, false),
		externalIPSet:       NewIPSet(ipset, KubeExternalIPSet, utilipset.HashIPPort, false),
		lbSet:               NewIPSet(ipset, KubeLoadBalancerSet, utilipset.HashIPPort, false),
		lbFWSet:             NewIPSet(ipset, KubeLoadbalancerFWSet, utilipset.HashIPPort, false),
		lbLocalSet:          NewIPSet(ipset, KubeLoadBalancerLocalSet, utilipset.HashIPPort, false),
		lbWhiteListIPSet:    NewIPSet(ipset, KubeLoadBalancerSourceIPSet, utilipset.HashIPPortIP, false),
		lbWhiteListCIDRSet:  NewIPSet(ipset, KubeLoadBalancerSourceCIDRSet, utilipset.HashIPPortNet, false),
		nodePortSetTCP:      NewIPSet(ipset, KubeNodePortSetTCP, utilipset.BitmapPort, false),
		nodePortLocalSetTCP: NewIPSet(ipset, KubeNodePortLocalSetTCP, utilipset.BitmapPort, false),
		nodePortLocalSetUDP: NewIPSet(ipset, KubeNodePortLocalSetUDP, utilipset.BitmapPort, false),
		nodePortSetUDP:      NewIPSet(ipset, KubeNodePortSetUDP, utilipset.BitmapPort, false),
		nodePortAddresses:   make([]string, 0),
		networkInterfacer:   proxyutiltest.NewFakeNetwork(),
	}
}

func makeNSN(namespace, name string) types.NamespacedName {
	return types.NamespacedName{Namespace: namespace, Name: name}
}

func makeServiceMap(proxier *Proxier, allServices ...*api.Service) {
	for i := range allServices {
		proxier.OnServiceAdd(allServices[i])
	}

	proxier.mu.Lock()
	defer proxier.mu.Unlock()
	proxier.servicesSynced = true
}

func makeTestService(namespace, name string, svcFunc func(*api.Service)) *api.Service {
	svc := &api.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: map[string]string{},
		},
		Spec:   api.ServiceSpec{},
		Status: api.ServiceStatus{},
	}
	svcFunc(svc)
	return svc
}

func makeEndpointsMap(proxier *Proxier, allEndpoints ...*api.Endpoints) {
	for i := range allEndpoints {
		proxier.OnEndpointsAdd(allEndpoints[i])
	}

	proxier.mu.Lock()
	defer proxier.mu.Unlock()
	proxier.endpointsSynced = true
}

func makeTestEndpoints(namespace, name string, eptFunc func(*api.Endpoints)) *api.Endpoints {
	ept := &api.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	eptFunc(ept)
	return ept
}

func TestCanUseIPVSProxier(t *testing.T) {
	testCases := []struct {
		mods         []string
		kernelErr    error
		ipsetVersion string
		ipsetErr     error
		ok           bool
	}{
		// case 0, kernel error
		{
			mods:         []string{"foo", "bar", "baz"},
			kernelErr:    fmt.Errorf("oops"),
			ipsetVersion: "0.0",
			ok:           false,
		},
		// case 1, ipset error
		{
			mods:         []string{"foo", "bar", "baz"},
			ipsetVersion: MinIPSetCheckVersion,
			ipsetErr:     fmt.Errorf("oops"),
			ok:           false,
		},
		// case 2, missing required kernel modules and ipset version too low
		{
			mods:         []string{"foo", "bar", "baz"},
			ipsetVersion: "1.1",
			ok:           false,
		},
		// case 3, missing required ip_vs_* kernel modules
		{
			mods:         []string{"ip_vs", "a", "bc", "def"},
			ipsetVersion: MinIPSetCheckVersion,
			ok:           false,
		},
		// case 4, ipset version too low
		{
			mods:         []string{"ip_vs", "ip_vs_rr", "ip_vs_wrr", "ip_vs_sh", "nf_conntrack_ipv4"},
			ipsetVersion: "4.3.0",
			ok:           false,
		},
		// case 5
		{
			mods:         []string{"ip_vs", "ip_vs_rr", "ip_vs_wrr", "ip_vs_sh", "nf_conntrack_ipv4"},
			ipsetVersion: MinIPSetCheckVersion,
			ok:           true,
		},
		// case 6
		{
			mods:         []string{"ip_vs", "ip_vs_rr", "ip_vs_wrr", "ip_vs_sh", "nf_conntrack_ipv4", "foo", "bar"},
			ipsetVersion: "6.19",
			ok:           true,
		},
	}

	for i := range testCases {
		handle := &fakeKernelHandler{modules: testCases[i].mods}
		versioner := &fakeIPSetVersioner{version: testCases[i].ipsetVersion, err: testCases[i].ipsetErr}
		ok, _ := CanUseIPVSProxier(handle, versioner)
		if ok != testCases[i].ok {
			t.Errorf("Case [%d], expect %v, got %v", i, testCases[i].ok, ok)
		}
	}
}

func TestGetNodeIPs(t *testing.T) {
	testCases := []struct {
		devAddresses map[string][]string
		expectIPs    []string
	}{
		// case 0
		{
			devAddresses: map[string][]string{"eth0": {"1.2.3.4"}, "lo": {"127.0.0.1"}},
			expectIPs:    []string{"1.2.3.4", "127.0.0.1"},
		},
		// case 1
		{
			devAddresses: map[string][]string{"lo": {"127.0.0.1"}},
			expectIPs:    []string{"127.0.0.1"},
		},
		// case 2
		{
			devAddresses: map[string][]string{},
			expectIPs:    []string{},
		},
		// case 3
		{
			devAddresses: map[string][]string{"encap0": {"10.20.30.40"}, "lo": {"127.0.0.1"}, "docker0": {"172.17.0.1"}},
			expectIPs:    []string{"10.20.30.40", "127.0.0.1", "172.17.0.1"},
		},
		// case 4
		{
			devAddresses: map[string][]string{"encaps9": {"10.20.30.40"}, "lo": {"127.0.0.1"}, "encap7": {"10.20.30.31"}},
			expectIPs:    []string{"10.20.30.40", "127.0.0.1", "10.20.30.31"},
		},
		// case 5
		{
			devAddresses: map[string][]string{"kube-ipvs0": {"1.2.3.4"}, "lo": {"127.0.0.1"}, "encap7": {"10.20.30.31"}},
			expectIPs:    []string{"127.0.0.1", "10.20.30.31"},
		},
		// case 6
		{
			devAddresses: map[string][]string{"kube-ipvs0": {"1.2.3.4", "2.3.4.5"}, "lo": {"127.0.0.1"}},
			expectIPs:    []string{"127.0.0.1"},
		},
		// case 7
		{
			devAddresses: map[string][]string{"kube-ipvs0": {"1.2.3.4", "2.3.4.5"}},
			expectIPs:    []string{},
		},
		// case 8
		{
			devAddresses: map[string][]string{"kube-ipvs0": {"1.2.3.4", "2.3.4.5"}, "eth5": {"3.4.5.6"}, "lo": {"127.0.0.1"}},
			expectIPs:    []string{"127.0.0.1", "3.4.5.6"},
		},
		// case 9
		{
			devAddresses: map[string][]string{"ipvs0": {"1.2.3.4"}, "lo": {"127.0.0.1"}, "encap7": {"10.20.30.31"}},
			expectIPs:    []string{"127.0.0.1", "10.20.30.31", "1.2.3.4"},
		},
	}

	for i := range testCases {
		fake := netlinktest.NewFakeNetlinkHandle()
		for dev, addresses := range testCases[i].devAddresses {
			fake.SetLocalAddresses(dev, addresses...)
		}
		r := realIPGetter{nl: fake}
		ips, err := r.NodeIPs()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		ipStrs := sets.NewString()
		for _, ip := range ips {
			ipStrs.Insert(ip.String())
		}
		if !ipStrs.Equal(sets.NewString(testCases[i].expectIPs...)) {
			t.Errorf("case[%d], unexpected mismatch, expected: %v, got: %v", i, testCases[i].expectIPs, ips)
		}
	}
}

func TestNodePort(t *testing.T) {
	ipt := iptablestest.NewFake()
	ipvs := ipvstest.NewFake()
	ipset := ipsettest.NewFake(testIPSetVersion)
	nodeIPv4 := net.ParseIP("100.101.102.103")
	nodeIPv6 := net.ParseIP("2001:db8::1:1")
	nodeIPs := sets.NewString(nodeIPv4.String(), nodeIPv6.String())
	fp := NewFakeProxier(ipt, ipvs, ipset, []net.IP{nodeIPv4, nodeIPv6})
	svcIP := "10.20.30.41"
	svcPort := 80
	svcNodePort := 3001
	svcPortName := proxy.ServicePortName{
		NamespacedName: makeNSN("ns1", "svc1"),
		Port:           "p80",
	}

	makeServiceMap(fp,
		makeTestService(svcPortName.Namespace, svcPortName.Name, func(svc *api.Service) {
			svc.Spec.Type = "NodePort"
			svc.Spec.ClusterIP = svcIP
			svc.Spec.Ports = []api.ServicePort{{
				Name:     svcPortName.Port,
				Port:     int32(svcPort),
				Protocol: api.ProtocolTCP,
				NodePort: int32(svcNodePort),
			}}
		}),
	)
	epIPv4 := "10.180.0.1"
	epIPv6 := "1002:ab8::2:10"
	epIPs := sets.NewString(epIPv4, epIPv6)
	makeEndpointsMap(fp,
		makeTestEndpoints(svcPortName.Namespace, svcPortName.Name, func(ept *api.Endpoints) {
			ept.Subsets = []api.EndpointSubset{{
				Addresses: []api.EndpointAddress{{
					IP: epIPv4,
				}, {
					IP: epIPv6,
				}},
				Ports: []api.EndpointPort{{
					Name: svcPortName.Port,
					Port: int32(svcPort),
				}},
			}}
		}),
	)

	fp.nodePortAddresses = []string{"0.0.0.0/0"}

	fp.syncProxyRules()

	// Check ipvs service and destinations
	services, err := ipvs.GetVirtualServers()
	if err != nil {
		t.Errorf("Failed to get ipvs services, err: %v", err)
	}
	if len(services) != 3 {
		t.Errorf("Expect 3 ipvs services, got %d", len(services))
	}
	found := false
	for _, svc := range services {
		if nodeIPs.Has(svc.Address.String()) && svc.Port == uint16(svcNodePort) && svc.Protocol == string(api.ProtocolTCP) {
			found = true
			destinations, err := ipvs.GetRealServers(svc)
			if err != nil {
				t.Errorf("Failed to get ipvs destinations, err: %v", err)
			}
			for _, dest := range destinations {
				if !epIPs.Has(dest.Address.String()) || dest.Port != uint16(svcPort) {
					t.Errorf("service Endpoint mismatch ipvs service destination")
				}
			}
			break
		}
	}
	if !found {
		t.Errorf("Expect node port type service, got none")
	}
}

func TestNodePortNoEndpoint(t *testing.T) {
	ipt := iptablestest.NewFake()
	ipvs := ipvstest.NewFake()
	ipset := ipsettest.NewFake(testIPSetVersion)
	nodeIP := net.ParseIP("100.101.102.103")
	fp := NewFakeProxier(ipt, ipvs, ipset, []net.IP{nodeIP})
	svcIP := "10.20.30.41"
	svcPort := 80
	svcNodePort := 3001
	svcPortName := proxy.ServicePortName{
		NamespacedName: makeNSN("ns1", "svc1"),
		Port:           "p80",
	}

	makeServiceMap(fp,
		makeTestService(svcPortName.Namespace, svcPortName.Name, func(svc *api.Service) {
			svc.Spec.Type = "NodePort"
			svc.Spec.ClusterIP = svcIP
			svc.Spec.Ports = []api.ServicePort{{
				Name:     svcPortName.Port,
				Port:     int32(svcPort),
				Protocol: api.ProtocolTCP,
				NodePort: int32(svcNodePort),
			}}
		}),
	)
	makeEndpointsMap(fp)

	fp.nodePortAddresses = []string{"0.0.0.0/0"}

	fp.syncProxyRules()

	// Check ipvs service and destinations
	services, err := ipvs.GetVirtualServers()
	if err != nil {
		t.Errorf("Failed to get ipvs services, err: %v", err)
	}
	if len(services) != 2 {
		t.Errorf("Expect 2 ipvs services, got %d", len(services))
	}
	found := false
	for _, svc := range services {
		if svc.Address.Equal(nodeIP) && svc.Port == uint16(svcNodePort) && svc.Protocol == string(api.ProtocolTCP) {
			found = true
			destinations, _ := ipvs.GetRealServers(svc)
			if len(destinations) != 0 {
				t.Errorf("Unexpected %d destinations, expect 0 destinations", len(destinations))
			}
			break
		}
	}
	if !found {
		t.Errorf("Expect node port type service, got none")
	}
}

func TestClusterIPNoEndpoint(t *testing.T) {
	ipt := iptablestest.NewFake()
	ipvs := ipvstest.NewFake()
	ipset := ipsettest.NewFake(testIPSetVersion)
	fp := NewFakeProxier(ipt, ipvs, ipset, nil)
	svcIP := "10.20.30.41"
	svcPort := 80
	svcPortName := proxy.ServicePortName{
		NamespacedName: makeNSN("ns1", "svc1"),
		Port:           "p80",
	}

	makeServiceMap(fp,
		makeTestService(svcPortName.Namespace, svcPortName.Namespace, func(svc *api.Service) {
			svc.Spec.ClusterIP = svcIP
			svc.Spec.Ports = []api.ServicePort{{
				Name:     svcPortName.Port,
				Port:     int32(svcPort),
				Protocol: api.ProtocolTCP,
			}}
		}),
	)
	makeEndpointsMap(fp)
	fp.syncProxyRules()

	// check ipvs service and destinations
	services, err := ipvs.GetVirtualServers()
	if err != nil {
		t.Errorf("Failed to get ipvs services, err: %v", err)
	}
	if len(services) != 1 {
		t.Errorf("Expect 1 ipvs services, got %d", len(services))
	} else {
		if services[0].Address.String() != svcIP || services[0].Port != uint16(svcPort) || services[0].Protocol != string(api.ProtocolTCP) {
			t.Errorf("Unexpected mismatch service")
		} else {
			destinations, _ := ipvs.GetRealServers(services[0])
			if len(destinations) != 0 {
				t.Errorf("Unexpected %d destinations, expect 0 destinations", len(destinations))
			}
		}
	}
}

func TestClusterIP(t *testing.T) {
	ipt := iptablestest.NewFake()
	ipvs := ipvstest.NewFake()
	ipset := ipsettest.NewFake(testIPSetVersion)
	fp := NewFakeProxier(ipt, ipvs, ipset, nil)

	svcIPv4 := "10.20.30.41"
	svcPortV4 := 80
	svcPortNameV4 := proxy.ServicePortName{
		NamespacedName: makeNSN("ns1", "svc1"),
		Port:           "p80",
	}
	svcIPv6 := "1002:ab8::2:1"
	svcPortV6 := 8080
	svcPortNameV6 := proxy.ServicePortName{
		NamespacedName: makeNSN("ns2", "svc2"),
		Port:           "p8080",
	}
	makeServiceMap(fp,
		makeTestService(svcPortNameV4.Namespace, svcPortNameV4.Name, func(svc *api.Service) {
			svc.Spec.ClusterIP = svcIPv4
			svc.Spec.Ports = []api.ServicePort{{
				Name:     svcPortNameV4.Port,
				Port:     int32(svcPortV4),
				Protocol: api.ProtocolTCP,
			}}
		}),
		makeTestService(svcPortNameV6.Namespace, svcPortNameV6.Name, func(svc *api.Service) {
			svc.Spec.ClusterIP = svcIPv6
			svc.Spec.Ports = []api.ServicePort{{
				Name:     svcPortNameV6.Port,
				Port:     int32(svcPortV6),
				Protocol: api.ProtocolTCP,
			}}
		}),
	)

	epIPv4 := "10.180.0.1"
	epIPv6 := "1009:ab8::5:6"
	makeEndpointsMap(fp,
		makeTestEndpoints(svcPortNameV4.Namespace, svcPortNameV4.Name, func(ept *api.Endpoints) {
			ept.Subsets = []api.EndpointSubset{{
				Addresses: []api.EndpointAddress{{
					IP: epIPv4,
				}},
				Ports: []api.EndpointPort{{
					Name: svcPortNameV4.Port,
					Port: int32(svcPortV4),
				}},
			}}
		}),
		makeTestEndpoints(svcPortNameV6.Namespace, svcPortNameV6.Name, func(ept *api.Endpoints) {
			ept.Subsets = []api.EndpointSubset{{
				Addresses: []api.EndpointAddress{{
					IP: epIPv6,
				}},
				Ports: []api.EndpointPort{{
					Name: svcPortNameV6.Port,
					Port: int32(svcPortV6),
				}},
			}}
		}),
	)

	fp.syncProxyRules()

	// check ipvs service and destinations
	services, err := ipvs.GetVirtualServers()
	if err != nil {
		t.Errorf("Failed to get ipvs services, err: %v", err)
	}
	if len(services) != 2 {
		t.Errorf("Expect 2 ipvs services, got %d", len(services))
	}
	for i := range services {
		// Check services
		if services[i].Address.String() == svcIPv4 {
			if services[i].Port != uint16(svcPortV4) || services[i].Protocol != string(api.ProtocolTCP) {
				t.Errorf("Unexpected mismatch service")
			}
			// Check destinations
			destinations, _ := ipvs.GetRealServers(services[i])
			if len(destinations) != 1 {
				t.Errorf("Expected 1 destinations, got %d destinations", len(destinations))
				continue
			}
			if destinations[0].Address.String() != epIPv4 || destinations[0].Port != uint16(svcPortV4) {
				t.Errorf("Unexpected mismatch destinations")
			}
		}
		if services[i].Address.String() == svcIPv6 {
			if services[i].Port != uint16(svcPortV6) || services[i].Protocol != string(api.ProtocolTCP) {
				t.Errorf("Unexpected mismatch service")
			}
			// Check destinations
			destinations, _ := ipvs.GetRealServers(services[i])
			if len(destinations) != 1 {
				t.Errorf("Expected 1 destinations, got %d destinations", len(destinations))
				continue
			}
			if destinations[0].Address.String() != epIPv6 || destinations[0].Port != uint16(svcPortV6) {
				t.Errorf("Unexpected mismatch destinations")
			}
		}
	}
}

func TestExternalIPsNoEndpoint(t *testing.T) {
	ipt := iptablestest.NewFake()
	ipvs := ipvstest.NewFake()
	ipset := ipsettest.NewFake(testIPSetVersion)
	fp := NewFakeProxier(ipt, ipvs, ipset, nil)
	svcIP := "10.20.30.41"
	svcPort := 80
	svcExternalIPs := "50.60.70.81"
	svcPortName := proxy.ServicePortName{
		NamespacedName: makeNSN("ns1", "svc1"),
		Port:           "p80",
	}

	makeServiceMap(fp,
		makeTestService(svcPortName.Namespace, svcPortName.Name, func(svc *api.Service) {
			svc.Spec.Type = "ClusterIP"
			svc.Spec.ClusterIP = svcIP
			svc.Spec.ExternalIPs = []string{svcExternalIPs}
			svc.Spec.Ports = []api.ServicePort{{
				Name:       svcPortName.Port,
				Port:       int32(svcPort),
				Protocol:   api.ProtocolTCP,
				TargetPort: intstr.FromInt(svcPort),
			}}
		}),
	)

	makeEndpointsMap(fp)

	fp.syncProxyRules()

	// check ipvs service and destinations
	services, err := ipvs.GetVirtualServers()
	if err != nil {
		t.Errorf("Failed to get ipvs services, err: %v", err)
	}
	if len(services) != 2 {
		t.Errorf("Expect 2 ipvs services, got %d", len(services))
	}
	found := false
	for _, svc := range services {
		if svc.Address.String() == svcExternalIPs && svc.Port == uint16(svcPort) && svc.Protocol == string(api.ProtocolTCP) {
			found = true
			destinations, _ := ipvs.GetRealServers(svc)
			if len(destinations) != 0 {
				t.Errorf("Unexpected %d destinations, expect 0 destinations", len(destinations))
			}
			break
		}
	}
	if !found {
		t.Errorf("Expect external ip type service, got none")
	}
}

func TestExternalIPs(t *testing.T) {
	ipt := iptablestest.NewFake()
	ipvs := ipvstest.NewFake()
	ipset := ipsettest.NewFake(testIPSetVersion)
	fp := NewFakeProxier(ipt, ipvs, ipset, nil)
	svcIP := "10.20.30.41"
	svcPort := 80
	svcExternalIPs := sets.NewString("50.60.70.81", "2012::51")
	svcPortName := proxy.ServicePortName{
		NamespacedName: makeNSN("ns1", "svc1"),
		Port:           "p80",
	}

	makeServiceMap(fp,
		makeTestService(svcPortName.Namespace, svcPortName.Name, func(svc *api.Service) {
			svc.Spec.Type = "ClusterIP"
			svc.Spec.ClusterIP = svcIP
			svc.Spec.ExternalIPs = svcExternalIPs.UnsortedList()
			svc.Spec.Ports = []api.ServicePort{{
				Name:       svcPortName.Port,
				Port:       int32(svcPort),
				Protocol:   api.ProtocolTCP,
				TargetPort: intstr.FromInt(svcPort),
			}}
		}),
	)

	epIP := "10.180.0.1"
	makeEndpointsMap(fp,
		makeTestEndpoints(svcPortName.Namespace, svcPortName.Name, func(ept *api.Endpoints) {
			ept.Subsets = []api.EndpointSubset{{
				Addresses: []api.EndpointAddress{{
					IP: epIP,
				}},
				Ports: []api.EndpointPort{{
					Name: svcPortName.Port,
					Port: int32(svcPort),
				}},
			}}
		}),
	)

	fp.syncProxyRules()

	// check ipvs service and destinations
	services, err := ipvs.GetVirtualServers()
	if err != nil {
		t.Errorf("Failed to get ipvs services, err: %v", err)
	}
	if len(services) != 3 {
		t.Errorf("Expect 3 ipvs services, got %d", len(services))
	}
	found := false
	for _, svc := range services {
		if svcExternalIPs.Has(svc.Address.String()) && svc.Port == uint16(svcPort) && svc.Protocol == string(api.ProtocolTCP) {
			found = true
			destinations, _ := ipvs.GetRealServers(svc)
			for _, dest := range destinations {
				if dest.Address.String() != epIP || dest.Port != uint16(svcPort) {
					t.Errorf("service Endpoint mismatch ipvs service destination")
				}
			}
			break
		}
	}
	if !found {
		t.Errorf("Expect external ip type service, got none")
	}
}

func TestLoadBalancer(t *testing.T) {
	ipt := iptablestest.NewFake()
	ipvs := ipvstest.NewFake()
	ipset := ipsettest.NewFake(testIPSetVersion)
	fp := NewFakeProxier(ipt, ipvs, ipset, nil)
	svcIP := "10.20.30.41"
	svcPort := 80
	svcNodePort := 3001
	svcLBIP := "1.2.3.4"
	svcPortName := proxy.ServicePortName{
		NamespacedName: makeNSN("ns1", "svc1"),
		Port:           "p80",
	}

	makeServiceMap(fp,
		makeTestService(svcPortName.Namespace, svcPortName.Name, func(svc *api.Service) {
			svc.Spec.Type = "LoadBalancer"
			svc.Spec.ClusterIP = svcIP
			svc.Spec.Ports = []api.ServicePort{{
				Name:     svcPortName.Port,
				Port:     int32(svcPort),
				Protocol: api.ProtocolTCP,
				NodePort: int32(svcNodePort),
			}}
			svc.Status.LoadBalancer.Ingress = []api.LoadBalancerIngress{{
				IP: svcLBIP,
			}}
		}),
	)

	epIP := "10.180.0.1"
	makeEndpointsMap(fp,
		makeTestEndpoints(svcPortName.Namespace, svcPortName.Name, func(ept *api.Endpoints) {
			ept.Subsets = []api.EndpointSubset{{
				Addresses: []api.EndpointAddress{{
					IP: epIP,
				}},
				Ports: []api.EndpointPort{{
					Name: svcPortName.Port,
					Port: int32(svcPort),
				}},
			}}
		}),
	)

	fp.syncProxyRules()
}

func TestOnlyLocalNodePorts(t *testing.T) {
	nodeIP := net.ParseIP("100.101.102.103")
	ipt, fp := buildFakeProxier([]net.IP{nodeIP})

	svcIP := "10.20.30.41"
	svcPort := 80
	svcNodePort := 3001
	svcPortName := proxy.ServicePortName{
		NamespacedName: makeNSN("ns1", "svc1"),
		Port:           "p80",
	}

	makeServiceMap(fp,
		makeTestService(svcPortName.Namespace, svcPortName.Name, func(svc *api.Service) {
			svc.Spec.Type = "NodePort"
			svc.Spec.ClusterIP = svcIP
			svc.Spec.Ports = []api.ServicePort{{
				Name:     svcPortName.Port,
				Port:     int32(svcPort),
				Protocol: api.ProtocolTCP,
				NodePort: int32(svcNodePort),
			}}
			svc.Spec.ExternalTrafficPolicy = api.ServiceExternalTrafficPolicyTypeLocal
		}),
	)

	epIP := "10.180.0.1"
	makeEndpointsMap(fp,
		makeTestEndpoints(svcPortName.Namespace, svcPortName.Name, func(ept *api.Endpoints) {
			ept.Subsets = []api.EndpointSubset{{
				Addresses: []api.EndpointAddress{{
					IP:       epIP,
					NodeName: nil,
				}},
				Ports: []api.EndpointPort{{
					Name: svcPortName.Port,
					Port: int32(svcPort),
				}},
			}}
		}),
	)

	itf := net.Interface{Index: 0, MTU: 0, Name: "eth0", HardwareAddr: nil, Flags: 0}
	addrs := []net.Addr{proxyutiltest.AddrStruct{Val: "100.101.102.103/24"}}
	itf1 := net.Interface{Index: 1, MTU: 0, Name: "eth1", HardwareAddr: nil, Flags: 0}
	addrs1 := []net.Addr{proxyutiltest.AddrStruct{Val: "2001:db8::0/64"}}
	fp.networkInterfacer.(*proxyutiltest.FakeNetwork).AddInterfaceAddr(&itf, addrs)
	fp.networkInterfacer.(*proxyutiltest.FakeNetwork).AddInterfaceAddr(&itf1, addrs1)
	fp.nodePortAddresses = []string{"100.101.102.0/24", "2001:db8::0/64"}

	fp.syncProxyRules()

	// Expect 3 services and 1 destination
	epVS := &netlinktest.ExpectedVirtualServer{
		VSNum: 3, IP: nodeIP.String(), Port: uint16(svcNodePort), Protocol: string(api.ProtocolTCP),
		RS: []netlinktest.ExpectedRealServer{{
			IP: epIP, Port: uint16(svcPort),
		}}}
	checkIPVS(t, fp, epVS)

	// check ipSet rules
	epEntry := &utilipset.Entry{
		Port:     svcNodePort,
		Protocol: strings.ToLower(string(api.ProtocolTCP)),
		SetType:  utilipset.BitmapPort,
	}
	epIPSet := netlinktest.ExpectedIPSet{
		KubeNodePortSetTCP:      {epEntry},
		KubeNodePortLocalSetTCP: {epEntry},
	}
	checkIPSet(t, fp, epIPSet)

	// Check iptables chain and rules
	epIpt := netlinktest.ExpectedIptablesChain{
		string(kubeServicesChain): {{
			JumpChain: string(KubeNodePortChain), MatchSet: KubeNodePortSetTCP,
		}},
		string(KubeNodePortChain): {{
			JumpChain: "ACCEPT", MatchSet: KubeNodePortLocalSetTCP,
		}, {
			JumpChain: string(KubeMarkMasqChain), MatchSet: "",
		}},
	}
	checkIptables(t, ipt, epIpt)
}
func TestLoadBalanceSourceRanges(t *testing.T) {
	nodeIP := net.ParseIP("100.101.102.103")
	ipt, fp := buildFakeProxier([]net.IP{nodeIP})

	svcIP := "10.20.30.41"
	svcPort := 80
	svcLBIP := "1.2.3.4"
	svcLBSource := "10.0.0.0/8"
	svcPortName := proxy.ServicePortName{
		NamespacedName: makeNSN("ns1", "svc1"),
		Port:           "p80",
	}
	epIP := "10.180.0.1"

	makeServiceMap(fp,
		makeTestService(svcPortName.Namespace, svcPortName.Name, func(svc *api.Service) {
			svc.Spec.Type = "LoadBalancer"
			svc.Spec.ClusterIP = svcIP
			svc.Spec.Ports = []api.ServicePort{{
				Name:     svcPortName.Port,
				Port:     int32(svcPort),
				Protocol: api.ProtocolTCP,
			}}
			svc.Status.LoadBalancer.Ingress = []api.LoadBalancerIngress{{
				IP: svcLBIP,
			}}
			svc.Spec.LoadBalancerSourceRanges = []string{
				svcLBSource,
			}
		}),
	)
	makeEndpointsMap(fp,
		makeTestEndpoints(svcPortName.Namespace, svcPortName.Name, func(ept *api.Endpoints) {
			ept.Subsets = []api.EndpointSubset{{
				Addresses: []api.EndpointAddress{{
					IP:       epIP,
					NodeName: nil,
				}},
				Ports: []api.EndpointPort{{
					Name: svcPortName.Port,
					Port: int32(svcPort),
				}},
			}}
		}),
	)

	fp.syncProxyRules()

	// Check ipvs service and destinations
	epVS := &netlinktest.ExpectedVirtualServer{
		VSNum: 2, IP: svcLBIP, Port: uint16(svcPort), Protocol: string(api.ProtocolTCP),
		RS: []netlinktest.ExpectedRealServer{{
			IP: epIP, Port: uint16(svcPort),
		}}}
	checkIPVS(t, fp, epVS)

	// Check ipset entry
	epIPSet := netlinktest.ExpectedIPSet{
		KubeLoadBalancerSet: {{
			IP:       svcLBIP,
			Port:     svcPort,
			Protocol: strings.ToLower(string(api.ProtocolTCP)),
			SetType:  utilipset.HashIPPort,
		}},
		KubeLoadbalancerFWSet: {{
			IP:       svcLBIP,
			Port:     svcPort,
			Protocol: strings.ToLower(string(api.ProtocolTCP)),
			SetType:  utilipset.HashIPPort,
		}},
		KubeLoadBalancerSourceCIDRSet: {{
			IP:       svcLBIP,
			Port:     svcPort,
			Protocol: strings.ToLower(string(api.ProtocolTCP)),
			Net:      svcLBSource,
			SetType:  utilipset.HashIPPortNet,
		}},
	}
	checkIPSet(t, fp, epIPSet)

	// Check iptables chain and rules
	epIpt := netlinktest.ExpectedIptablesChain{
		string(kubeServicesChain): {{
			JumpChain: string(KubeLoadBalancerChain), MatchSet: KubeLoadBalancerSet,
		}},
		string(KubeLoadBalancerChain): {{
			JumpChain: string(KubeFireWallChain), MatchSet: KubeLoadbalancerFWSet,
		}, {
			JumpChain: string(KubeMarkMasqChain), MatchSet: "",
		}},
		string(KubeFireWallChain): {{
			JumpChain: "RETURN", MatchSet: KubeLoadBalancerSourceCIDRSet,
		}, {
			JumpChain: string(KubeMarkDropChain), MatchSet: "",
		}},
	}
	checkIptables(t, ipt, epIpt)
}

func TestAcceptIPVSTraffic(t *testing.T) {
	ipt, fp := buildFakeProxier(nil)

	ingressIP := "1.2.3.4"
	externalIP := []string{"5.6.7.8"}
	svcInfos := []struct {
		svcType api.ServiceType
		svcIP   string
		svcName string
		epIP    string
	}{
		{api.ServiceTypeClusterIP, "10.20.30.40", "svc1", "10.180.0.1"},
		{api.ServiceTypeLoadBalancer, "10.20.30.41", "svc2", "10.180.0.2"},
		{api.ServiceTypeNodePort, "10.20.30.42", "svc3", "10.180.0.3"},
	}

	for _, svcInfo := range svcInfos {
		makeServiceMap(fp,
			makeTestService("ns1", svcInfo.svcName, func(svc *api.Service) {
				svc.Spec.Type = svcInfo.svcType
				svc.Spec.ClusterIP = svcInfo.svcIP
				svc.Spec.Ports = []api.ServicePort{{
					Name:     "p80",
					Port:     80,
					Protocol: api.ProtocolTCP,
					NodePort: 80,
				}}
				if svcInfo.svcType == api.ServiceTypeLoadBalancer {
					svc.Status.LoadBalancer.Ingress = []api.LoadBalancerIngress{{
						IP: ingressIP,
					}}
				}
				if svcInfo.svcType == api.ServiceTypeClusterIP {
					svc.Spec.ExternalIPs = externalIP
				}
			}),
		)

		makeEndpointsMap(fp,
			makeTestEndpoints("ns1", "p80", func(ept *api.Endpoints) {
				ept.Subsets = []api.EndpointSubset{{
					Addresses: []api.EndpointAddress{{
						IP: svcInfo.epIP,
					}},
					Ports: []api.EndpointPort{{
						Name: "p80",
						Port: 80,
					}},
				}}
			}),
		)
	}
	fp.syncProxyRules()

	// Check iptables chain and rules
	epIpt := netlinktest.ExpectedIptablesChain{
		string(kubeServicesChain): {
			{JumpChain: "ACCEPT", MatchSet: KubeClusterIPSet},
			{JumpChain: "ACCEPT", MatchSet: KubeLoadBalancerSet},
			{JumpChain: "ACCEPT", MatchSet: KubeExternalIPSet},
		},
	}
	checkIptables(t, ipt, epIpt)
}

func TestOnlyLocalLoadBalancing(t *testing.T) {
	ipt, fp := buildFakeProxier(nil)

	svcIP := "10.20.30.41"
	svcPort := 80
	svcNodePort := 3001
	svcLBIP := "1.2.3.4"
	svcPortName := proxy.ServicePortName{
		NamespacedName: makeNSN("ns1", "svc1"),
		Port:           "p80",
	}

	makeServiceMap(fp,
		makeTestService(svcPortName.Namespace, svcPortName.Name, func(svc *api.Service) {
			svc.Spec.Type = "LoadBalancer"
			svc.Spec.ClusterIP = svcIP
			svc.Spec.Ports = []api.ServicePort{{
				Name:     svcPortName.Port,
				Port:     int32(svcPort),
				Protocol: api.ProtocolTCP,
				NodePort: int32(svcNodePort),
			}}
			svc.Status.LoadBalancer.Ingress = []api.LoadBalancerIngress{{
				IP: svcLBIP,
			}}
			svc.Spec.ExternalTrafficPolicy = api.ServiceExternalTrafficPolicyTypeLocal
		}),
	)

	epIP := "10.180.0.1"
	makeEndpointsMap(fp,
		makeTestEndpoints(svcPortName.Namespace, svcPortName.Name, func(ept *api.Endpoints) {
			ept.Subsets = []api.EndpointSubset{{
				Addresses: []api.EndpointAddress{{
					IP:       epIP,
					NodeName: nil,
				}},
				Ports: []api.EndpointPort{{
					Name: svcPortName.Port,
					Port: int32(svcPort),
				}},
			}}
		}),
	)

	fp.syncProxyRules()

	// Expect 2 services and 1 destination
	epVS := &netlinktest.ExpectedVirtualServer{
		VSNum: 2, IP: svcLBIP, Port: uint16(svcPort), Protocol: string(api.ProtocolTCP),
		RS: []netlinktest.ExpectedRealServer{{
			IP: epIP, Port: uint16(svcPort),
		}}}
	checkIPVS(t, fp, epVS)

	// check ipSet rules
	epIPSet := netlinktest.ExpectedIPSet{
		KubeLoadBalancerSet: {{
			IP:       svcLBIP,
			Port:     svcPort,
			Protocol: strings.ToLower(string(api.ProtocolTCP)),
			SetType:  utilipset.HashIPPort,
		}},
		KubeLoadBalancerLocalSet: {{
			IP:       svcLBIP,
			Port:     svcPort,
			Protocol: strings.ToLower(string(api.ProtocolTCP)),
			SetType:  utilipset.HashIPPort,
		}},
	}
	checkIPSet(t, fp, epIPSet)

	// Check iptables chain and rules
	epIpt := netlinktest.ExpectedIptablesChain{
		string(kubeServicesChain): {{
			JumpChain: string(KubeLoadBalancerChain), MatchSet: KubeLoadBalancerSet,
		}},
		string(KubeLoadBalancerChain): {{
			JumpChain: "RETURN", MatchSet: KubeLoadBalancerLocalSet,
		}, {
			JumpChain: string(KubeMarkMasqChain), MatchSet: "",
		}},
	}
	checkIptables(t, ipt, epIpt)
}

func addTestPort(array []api.ServicePort, name string, protocol api.Protocol, port, nodeport int32, targetPort int) []api.ServicePort {
	svcPort := api.ServicePort{
		Name:       name,
		Protocol:   protocol,
		Port:       port,
		NodePort:   nodeport,
		TargetPort: intstr.FromInt(targetPort),
	}
	return append(array, svcPort)
}

func TestBuildServiceMapAddRemove(t *testing.T) {
	ipt := iptablestest.NewFake()
	ipvs := ipvstest.NewFake()
	ipset := ipsettest.NewFake(testIPSetVersion)
	fp := NewFakeProxier(ipt, ipvs, ipset, nil)

	services := []*api.Service{
		makeTestService("somewhere-else", "cluster-ip", func(svc *api.Service) {
			svc.Spec.Type = api.ServiceTypeClusterIP
			svc.Spec.ClusterIP = "172.16.55.4"
			svc.Spec.Ports = addTestPort(svc.Spec.Ports, "something", "UDP", 1234, 4321, 0)
			svc.Spec.Ports = addTestPort(svc.Spec.Ports, "somethingelse", "UDP", 1235, 5321, 0)
		}),
		makeTestService("somewhere-else", "node-port", func(svc *api.Service) {
			svc.Spec.Type = api.ServiceTypeNodePort
			svc.Spec.ClusterIP = "172.16.55.10"
			svc.Spec.Ports = addTestPort(svc.Spec.Ports, "blahblah", "UDP", 345, 678, 0)
			svc.Spec.Ports = addTestPort(svc.Spec.Ports, "moreblahblah", "TCP", 344, 677, 0)
		}),
		makeTestService("somewhere", "load-balancer", func(svc *api.Service) {
			svc.Spec.Type = api.ServiceTypeLoadBalancer
			svc.Spec.ClusterIP = "172.16.55.11"
			svc.Spec.LoadBalancerIP = "5.6.7.8"
			svc.Spec.Ports = addTestPort(svc.Spec.Ports, "foobar", "UDP", 8675, 30061, 7000)
			svc.Spec.Ports = addTestPort(svc.Spec.Ports, "baz", "UDP", 8676, 30062, 7001)
			svc.Status.LoadBalancer = api.LoadBalancerStatus{
				Ingress: []api.LoadBalancerIngress{
					{IP: "10.1.2.4"},
				},
			}
		}),
		makeTestService("somewhere", "only-local-load-balancer", func(svc *api.Service) {
			svc.Spec.Type = api.ServiceTypeLoadBalancer
			svc.Spec.ClusterIP = "172.16.55.12"
			svc.Spec.LoadBalancerIP = "5.6.7.8"
			svc.Spec.Ports = addTestPort(svc.Spec.Ports, "foobar2", "UDP", 8677, 30063, 7002)
			svc.Spec.Ports = addTestPort(svc.Spec.Ports, "baz", "UDP", 8678, 30064, 7003)
			svc.Status.LoadBalancer = api.LoadBalancerStatus{
				Ingress: []api.LoadBalancerIngress{
					{IP: "10.1.2.3"},
				},
			}
			svc.Spec.ExternalTrafficPolicy = api.ServiceExternalTrafficPolicyTypeLocal
			svc.Spec.HealthCheckNodePort = 345
		}),
	}

	for i := range services {
		fp.OnServiceAdd(services[i])
	}
	result := proxy.UpdateServiceMap(fp.serviceMap, fp.serviceChanges)
	if len(fp.serviceMap) != 8 {
		t.Errorf("expected service map length 8, got %v", fp.serviceMap)
	}

	// The only-local-loadbalancer ones get added
	if len(result.HCServiceNodePorts) != 1 {
		t.Errorf("expected 1 healthcheck port, got %v", result.HCServiceNodePorts)
	} else {
		nsn := makeNSN("somewhere", "only-local-load-balancer")
		if port, found := result.HCServiceNodePorts[nsn]; !found || port != 345 {
			t.Errorf("expected healthcheck port [%q]=345: got %v", nsn, result.HCServiceNodePorts)
		}
	}

	if len(result.UDPStaleClusterIP) != 0 {
		// Services only added, so nothing stale yet
		t.Errorf("expected stale UDP services length 0, got %d", len(result.UDPStaleClusterIP))
	}

	// Remove some stuff
	// oneService is a modification of services[0] with removed first port.
	oneService := makeTestService("somewhere-else", "cluster-ip", func(svc *api.Service) {
		svc.Spec.Type = api.ServiceTypeClusterIP
		svc.Spec.ClusterIP = "172.16.55.4"
		svc.Spec.Ports = addTestPort(svc.Spec.Ports, "somethingelse", "UDP", 1235, 5321, 0)
	})

	fp.OnServiceUpdate(services[0], oneService)
	fp.OnServiceDelete(services[1])
	fp.OnServiceDelete(services[2])
	fp.OnServiceDelete(services[3])

	result = proxy.UpdateServiceMap(fp.serviceMap, fp.serviceChanges)
	if len(fp.serviceMap) != 1 {
		t.Errorf("expected service map length 1, got %v", fp.serviceMap)
	}

	if len(result.HCServiceNodePorts) != 0 {
		t.Errorf("expected 0 healthcheck ports, got %v", result.HCServiceNodePorts)
	}

	// All services but one were deleted. While you'd expect only the ClusterIPs
	// from the three deleted services here, we still have the ClusterIP for
	// the not-deleted service, because one of it's ServicePorts was deleted.
	expectedStaleUDPServices := []string{"172.16.55.10", "172.16.55.4", "172.16.55.11", "172.16.55.12"}
	if len(result.UDPStaleClusterIP) != len(expectedStaleUDPServices) {
		t.Errorf("expected stale UDP services length %d, got %v", len(expectedStaleUDPServices), result.UDPStaleClusterIP.List())
	}
	for _, ip := range expectedStaleUDPServices {
		if !result.UDPStaleClusterIP.Has(ip) {
			t.Errorf("expected stale UDP service service %s", ip)
		}
	}
}

func TestBuildServiceMapServiceHeadless(t *testing.T) {
	ipt := iptablestest.NewFake()
	ipvs := ipvstest.NewFake()
	ipset := ipsettest.NewFake(testIPSetVersion)
	fp := NewFakeProxier(ipt, ipvs, ipset, nil)

	makeServiceMap(fp,
		makeTestService("somewhere-else", "headless", func(svc *api.Service) {
			svc.Spec.Type = api.ServiceTypeClusterIP
			svc.Spec.ClusterIP = api.ClusterIPNone
			svc.Spec.Ports = addTestPort(svc.Spec.Ports, "rpc", "UDP", 1234, 0, 0)
		}),
		makeTestService("somewhere-else", "headless-without-port", func(svc *api.Service) {
			svc.Spec.Type = api.ServiceTypeClusterIP
			svc.Spec.ClusterIP = api.ClusterIPNone
		}),
	)

	// Headless service should be ignored
	result := proxy.UpdateServiceMap(fp.serviceMap, fp.serviceChanges)
	if len(fp.serviceMap) != 0 {
		t.Errorf("expected service map length 0, got %d", len(fp.serviceMap))
	}

	// No proxied services, so no healthchecks
	if len(result.HCServiceNodePorts) != 0 {
		t.Errorf("expected healthcheck ports length 0, got %d", len(result.HCServiceNodePorts))
	}

	if len(result.UDPStaleClusterIP) != 0 {
		t.Errorf("expected stale UDP services length 0, got %d", len(result.UDPStaleClusterIP))
	}
}

func TestBuildServiceMapServiceTypeExternalName(t *testing.T) {
	ipt := iptablestest.NewFake()
	ipvs := ipvstest.NewFake()
	ipset := ipsettest.NewFake(testIPSetVersion)
	fp := NewFakeProxier(ipt, ipvs, ipset, nil)

	makeServiceMap(fp,
		makeTestService("somewhere-else", "external-name", func(svc *api.Service) {
			svc.Spec.Type = api.ServiceTypeExternalName
			svc.Spec.ClusterIP = "172.16.55.4" // Should be ignored
			svc.Spec.ExternalName = "foo2.bar.com"
			svc.Spec.Ports = addTestPort(svc.Spec.Ports, "blah", "UDP", 1235, 5321, 0)
		}),
	)

	result := proxy.UpdateServiceMap(fp.serviceMap, fp.serviceChanges)
	if len(fp.serviceMap) != 0 {
		t.Errorf("expected service map length 0, got %v", fp.serviceMap)
	}
	// No proxied services, so no healthchecks
	if len(result.HCServiceNodePorts) != 0 {
		t.Errorf("expected healthcheck ports length 0, got %v", result.HCServiceNodePorts)
	}
	if len(result.UDPStaleClusterIP) != 0 {
		t.Errorf("expected stale UDP services length 0, got %v", result.UDPStaleClusterIP)
	}
}

func TestBuildServiceMapServiceUpdate(t *testing.T) {
	ipt := iptablestest.NewFake()
	ipvs := ipvstest.NewFake()
	ipset := ipsettest.NewFake(testIPSetVersion)
	fp := NewFakeProxier(ipt, ipvs, ipset, nil)

	servicev1 := makeTestService("somewhere", "some-service", func(svc *api.Service) {
		svc.Spec.Type = api.ServiceTypeClusterIP
		svc.Spec.ClusterIP = "172.16.55.4"
		svc.Spec.Ports = addTestPort(svc.Spec.Ports, "something", "UDP", 1234, 4321, 0)
		svc.Spec.Ports = addTestPort(svc.Spec.Ports, "somethingelse", "TCP", 1235, 5321, 0)
	})
	servicev2 := makeTestService("somewhere", "some-service", func(svc *api.Service) {
		svc.Spec.Type = api.ServiceTypeLoadBalancer
		svc.Spec.ClusterIP = "172.16.55.4"
		svc.Spec.LoadBalancerIP = "5.6.7.8"
		svc.Spec.Ports = addTestPort(svc.Spec.Ports, "something", "UDP", 1234, 4321, 7002)
		svc.Spec.Ports = addTestPort(svc.Spec.Ports, "somethingelse", "TCP", 1235, 5321, 7003)
		svc.Status.LoadBalancer = api.LoadBalancerStatus{
			Ingress: []api.LoadBalancerIngress{
				{IP: "10.1.2.3"},
			},
		}
		svc.Spec.ExternalTrafficPolicy = api.ServiceExternalTrafficPolicyTypeLocal
		svc.Spec.HealthCheckNodePort = 345
	})

	fp.OnServiceAdd(servicev1)

	result := proxy.UpdateServiceMap(fp.serviceMap, fp.serviceChanges)
	if len(fp.serviceMap) != 2 {
		t.Errorf("expected service map length 2, got %v", fp.serviceMap)
	}
	if len(result.HCServiceNodePorts) != 0 {
		t.Errorf("expected healthcheck ports length 0, got %v", result.HCServiceNodePorts)
	}
	if len(result.UDPStaleClusterIP) != 0 {
		// Services only added, so nothing stale yet
		t.Errorf("expected stale UDP services length 0, got %d", len(result.UDPStaleClusterIP))
	}

	// Change service to load-balancer
	fp.OnServiceUpdate(servicev1, servicev2)
	result = proxy.UpdateServiceMap(fp.serviceMap, fp.serviceChanges)
	if len(fp.serviceMap) != 2 {
		t.Errorf("expected service map length 2, got %v", fp.serviceMap)
	}
	if len(result.HCServiceNodePorts) != 1 {
		t.Errorf("expected healthcheck ports length 1, got %v", result.HCServiceNodePorts)
	}
	if len(result.UDPStaleClusterIP) != 0 {
		t.Errorf("expected stale UDP services length 0, got %v", result.UDPStaleClusterIP.List())
	}

	// No change; make sure the service map stays the same and there are
	// no health-check changes
	fp.OnServiceUpdate(servicev2, servicev2)
	result = proxy.UpdateServiceMap(fp.serviceMap, fp.serviceChanges)
	if len(fp.serviceMap) != 2 {
		t.Errorf("expected service map length 2, got %v", fp.serviceMap)
	}
	if len(result.HCServiceNodePorts) != 1 {
		t.Errorf("expected healthcheck ports length 1, got %v", result.HCServiceNodePorts)
	}
	if len(result.UDPStaleClusterIP) != 0 {
		t.Errorf("expected stale UDP services length 0, got %v", result.UDPStaleClusterIP.List())
	}

	// And back to ClusterIP
	fp.OnServiceUpdate(servicev2, servicev1)
	result = proxy.UpdateServiceMap(fp.serviceMap, fp.serviceChanges)
	if len(fp.serviceMap) != 2 {
		t.Errorf("expected service map length 2, got %v", fp.serviceMap)
	}
	if len(result.HCServiceNodePorts) != 0 {
		t.Errorf("expected healthcheck ports length 0, got %v", result.HCServiceNodePorts)
	}
	if len(result.UDPStaleClusterIP) != 0 {
		// Services only added, so nothing stale yet
		t.Errorf("expected stale UDP services length 0, got %d", len(result.UDPStaleClusterIP))
	}
}

func TestSessionAffinity(t *testing.T) {
	ipt := iptablestest.NewFake()
	ipvs := ipvstest.NewFake()
	ipset := ipsettest.NewFake(testIPSetVersion)
	nodeIP := net.ParseIP("100.101.102.103")
	fp := NewFakeProxier(ipt, ipvs, ipset, []net.IP{nodeIP})
	svcIP := "10.20.30.41"
	svcPort := 80
	svcNodePort := 3001
	svcExternalIPs := "50.60.70.81"
	svcPortName := proxy.ServicePortName{
		NamespacedName: makeNSN("ns1", "svc1"),
		Port:           "p80",
	}
	timeoutSeconds := api.DefaultClientIPServiceAffinitySeconds

	makeServiceMap(fp,
		makeTestService(svcPortName.Namespace, svcPortName.Name, func(svc *api.Service) {
			svc.Spec.Type = "NodePort"
			svc.Spec.ClusterIP = svcIP
			svc.Spec.ExternalIPs = []string{svcExternalIPs}
			svc.Spec.SessionAffinity = api.ServiceAffinityClientIP
			svc.Spec.SessionAffinityConfig = &api.SessionAffinityConfig{
				ClientIP: &api.ClientIPConfig{
					TimeoutSeconds: &timeoutSeconds,
				},
			}
			svc.Spec.Ports = []api.ServicePort{{
				Name:     svcPortName.Port,
				Port:     int32(svcPort),
				Protocol: api.ProtocolTCP,
				NodePort: int32(svcNodePort),
			}}
		}),
	)
	makeEndpointsMap(fp)

	fp.syncProxyRules()

	// check ipvs service and destinations
	services, err := ipvs.GetVirtualServers()
	if err != nil {
		t.Errorf("Failed to get ipvs services, err: %v", err)
	}
	for _, svc := range services {
		if svc.Timeout != uint32(api.DefaultClientIPServiceAffinitySeconds) {
			t.Errorf("Unexpected mismatch ipvs service session affinity timeout: %d, expected: %d", svc.Timeout, api.DefaultClientIPServiceAffinitySeconds)
		}
	}
}

func makeServicePortName(ns, name, port string) proxy.ServicePortName {
	return proxy.ServicePortName{
		NamespacedName: makeNSN(ns, name),
		Port:           port,
	}
}

func Test_updateEndpointsMap(t *testing.T) {
	var nodeName = testHostname

	emptyEndpoint := func(ept *api.Endpoints) {
		ept.Subsets = []api.EndpointSubset{}
	}
	unnamedPort := func(ept *api.Endpoints) {
		ept.Subsets = []api.EndpointSubset{{
			Addresses: []api.EndpointAddress{{
				IP: "1.1.1.1",
			}},
			Ports: []api.EndpointPort{{
				Port: 11,
			}},
		}}
	}
	unnamedPortLocal := func(ept *api.Endpoints) {
		ept.Subsets = []api.EndpointSubset{{
			Addresses: []api.EndpointAddress{{
				IP:       "1.1.1.1",
				NodeName: &nodeName,
			}},
			Ports: []api.EndpointPort{{
				Port: 11,
			}},
		}}
	}
	namedPortLocal := func(ept *api.Endpoints) {
		ept.Subsets = []api.EndpointSubset{{
			Addresses: []api.EndpointAddress{{
				IP:       "1.1.1.1",
				NodeName: &nodeName,
			}},
			Ports: []api.EndpointPort{{
				Name: "p11",
				Port: 11,
			}},
		}}
	}
	namedPort := func(ept *api.Endpoints) {
		ept.Subsets = []api.EndpointSubset{{
			Addresses: []api.EndpointAddress{{
				IP: "1.1.1.1",
			}},
			Ports: []api.EndpointPort{{
				Name: "p11",
				Port: 11,
			}},
		}}
	}
	namedPortRenamed := func(ept *api.Endpoints) {
		ept.Subsets = []api.EndpointSubset{{
			Addresses: []api.EndpointAddress{{
				IP: "1.1.1.1",
			}},
			Ports: []api.EndpointPort{{
				Name: "p11-2",
				Port: 11,
			}},
		}}
	}
	namedPortRenumbered := func(ept *api.Endpoints) {
		ept.Subsets = []api.EndpointSubset{{
			Addresses: []api.EndpointAddress{{
				IP: "1.1.1.1",
			}},
			Ports: []api.EndpointPort{{
				Name: "p11",
				Port: 22,
			}},
		}}
	}
	namedPortsLocalNoLocal := func(ept *api.Endpoints) {
		ept.Subsets = []api.EndpointSubset{{
			Addresses: []api.EndpointAddress{{
				IP: "1.1.1.1",
			}, {
				IP:       "1.1.1.2",
				NodeName: &nodeName,
			}},
			Ports: []api.EndpointPort{{
				Name: "p11",
				Port: 11,
			}, {
				Name: "p12",
				Port: 12,
			}},
		}}
	}
	multipleSubsets := func(ept *api.Endpoints) {
		ept.Subsets = []api.EndpointSubset{{
			Addresses: []api.EndpointAddress{{
				IP: "1.1.1.1",
			}},
			Ports: []api.EndpointPort{{
				Name: "p11",
				Port: 11,
			}},
		}, {
			Addresses: []api.EndpointAddress{{
				IP: "1.1.1.2",
			}},
			Ports: []api.EndpointPort{{
				Name: "p12",
				Port: 12,
			}},
		}}
	}
	multipleSubsetsWithLocal := func(ept *api.Endpoints) {
		ept.Subsets = []api.EndpointSubset{{
			Addresses: []api.EndpointAddress{{
				IP: "1.1.1.1",
			}},
			Ports: []api.EndpointPort{{
				Name: "p11",
				Port: 11,
			}},
		}, {
			Addresses: []api.EndpointAddress{{
				IP:       "1.1.1.2",
				NodeName: &nodeName,
			}},
			Ports: []api.EndpointPort{{
				Name: "p12",
				Port: 12,
			}},
		}}
	}
	multipleSubsetsMultiplePortsLocal := func(ept *api.Endpoints) {
		ept.Subsets = []api.EndpointSubset{{
			Addresses: []api.EndpointAddress{{
				IP:       "1.1.1.1",
				NodeName: &nodeName,
			}},
			Ports: []api.EndpointPort{{
				Name: "p11",
				Port: 11,
			}, {
				Name: "p12",
				Port: 12,
			}},
		}, {
			Addresses: []api.EndpointAddress{{
				IP: "1.1.1.3",
			}},
			Ports: []api.EndpointPort{{
				Name: "p13",
				Port: 13,
			}},
		}}
	}
	multipleSubsetsIPsPorts1 := func(ept *api.Endpoints) {
		ept.Subsets = []api.EndpointSubset{{
			Addresses: []api.EndpointAddress{{
				IP: "1.1.1.1",
			}, {
				IP:       "1.1.1.2",
				NodeName: &nodeName,
			}},
			Ports: []api.EndpointPort{{
				Name: "p11",
				Port: 11,
			}, {
				Name: "p12",
				Port: 12,
			}},
		}, {
			Addresses: []api.EndpointAddress{{
				IP: "1.1.1.3",
			}, {
				IP:       "1.1.1.4",
				NodeName: &nodeName,
			}},
			Ports: []api.EndpointPort{{
				Name: "p13",
				Port: 13,
			}, {
				Name: "p14",
				Port: 14,
			}},
		}}
	}
	multipleSubsetsIPsPorts2 := func(ept *api.Endpoints) {
		ept.Subsets = []api.EndpointSubset{{
			Addresses: []api.EndpointAddress{{
				IP: "2.2.2.1",
			}, {
				IP:       "2.2.2.2",
				NodeName: &nodeName,
			}},
			Ports: []api.EndpointPort{{
				Name: "p21",
				Port: 21,
			}, {
				Name: "p22",
				Port: 22,
			}},
		}}
	}
	complexBefore1 := func(ept *api.Endpoints) {
		ept.Subsets = []api.EndpointSubset{{
			Addresses: []api.EndpointAddress{{
				IP: "1.1.1.1",
			}},
			Ports: []api.EndpointPort{{
				Name: "p11",
				Port: 11,
			}},
		}}
	}
	complexBefore2 := func(ept *api.Endpoints) {
		ept.Subsets = []api.EndpointSubset{{
			Addresses: []api.EndpointAddress{{
				IP:       "2.2.2.2",
				NodeName: &nodeName,
			}, {
				IP:       "2.2.2.22",
				NodeName: &nodeName,
			}},
			Ports: []api.EndpointPort{{
				Name: "p22",
				Port: 22,
			}},
		}, {
			Addresses: []api.EndpointAddress{{
				IP:       "2.2.2.3",
				NodeName: &nodeName,
			}},
			Ports: []api.EndpointPort{{
				Name: "p23",
				Port: 23,
			}},
		}}
	}
	complexBefore4 := func(ept *api.Endpoints) {
		ept.Subsets = []api.EndpointSubset{{
			Addresses: []api.EndpointAddress{{
				IP:       "4.4.4.4",
				NodeName: &nodeName,
			}, {
				IP:       "4.4.4.5",
				NodeName: &nodeName,
			}},
			Ports: []api.EndpointPort{{
				Name: "p44",
				Port: 44,
			}},
		}, {
			Addresses: []api.EndpointAddress{{
				IP:       "4.4.4.6",
				NodeName: &nodeName,
			}},
			Ports: []api.EndpointPort{{
				Name: "p45",
				Port: 45,
			}},
		}}
	}
	complexAfter1 := func(ept *api.Endpoints) {
		ept.Subsets = []api.EndpointSubset{{
			Addresses: []api.EndpointAddress{{
				IP: "1.1.1.1",
			}, {
				IP: "1.1.1.11",
			}},
			Ports: []api.EndpointPort{{
				Name: "p11",
				Port: 11,
			}},
		}, {
			Addresses: []api.EndpointAddress{{
				IP: "1.1.1.2",
			}},
			Ports: []api.EndpointPort{{
				Name: "p12",
				Port: 12,
			}, {
				Name: "p122",
				Port: 122,
			}},
		}}
	}
	complexAfter3 := func(ept *api.Endpoints) {
		ept.Subsets = []api.EndpointSubset{{
			Addresses: []api.EndpointAddress{{
				IP: "3.3.3.3",
			}},
			Ports: []api.EndpointPort{{
				Name: "p33",
				Port: 33,
			}},
		}}
	}
	complexAfter4 := func(ept *api.Endpoints) {
		ept.Subsets = []api.EndpointSubset{{
			Addresses: []api.EndpointAddress{{
				IP:       "4.4.4.4",
				NodeName: &nodeName,
			}},
			Ports: []api.EndpointPort{{
				Name: "p44",
				Port: 44,
			}},
		}}
	}

	testCases := []struct {
		// previousEndpoints and currentEndpoints are used to call appropriate
		// handlers OnEndpoints* (based on whether corresponding values are nil
		// or non-nil) and must be of equal length.
		previousEndpoints         []*api.Endpoints
		currentEndpoints          []*api.Endpoints
		oldEndpoints              map[proxy.ServicePortName][]*proxy.BaseEndpointInfo
		expectedResult            map[proxy.ServicePortName][]*proxy.BaseEndpointInfo
		expectedStaleEndpoints    []proxy.ServiceEndpoint
		expectedStaleServiceNames map[proxy.ServicePortName]bool
		expectedHealthchecks      map[types.NamespacedName]int
	}{{
		// Case[0]: nothing
		oldEndpoints:              map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{},
		expectedResult:            map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{},
		expectedStaleEndpoints:    []proxy.ServiceEndpoint{},
		expectedStaleServiceNames: map[proxy.ServicePortName]bool{},
		expectedHealthchecks:      map[types.NamespacedName]int{},
	}, {
		// Case[1]: no change, unnamed port
		previousEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", unnamedPort),
		},
		currentEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", unnamedPort),
		},
		oldEndpoints: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", ""): {
				{Endpoint: "1.1.1.1:11", IsLocal: false},
			},
		},
		expectedResult: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", ""): {
				{Endpoint: "1.1.1.1:11", IsLocal: false},
			},
		},
		expectedStaleEndpoints:    []proxy.ServiceEndpoint{},
		expectedStaleServiceNames: map[proxy.ServicePortName]bool{},
		expectedHealthchecks:      map[types.NamespacedName]int{},
	}, {
		// Case[2]: no change, named port, local
		previousEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", namedPortLocal),
		},
		currentEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", namedPortLocal),
		},
		oldEndpoints: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", "p11"): {
				{Endpoint: "1.1.1.1:11", IsLocal: true},
			},
		},
		expectedResult: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", "p11"): {
				{Endpoint: "1.1.1.1:11", IsLocal: true},
			},
		},
		expectedStaleEndpoints:    []proxy.ServiceEndpoint{},
		expectedStaleServiceNames: map[proxy.ServicePortName]bool{},
		expectedHealthchecks: map[types.NamespacedName]int{
			makeNSN("ns1", "ep1"): 1,
		},
	}, {
		// Case[3]: no change, multiple subsets
		previousEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", multipleSubsets),
		},
		currentEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", multipleSubsets),
		},
		oldEndpoints: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", "p11"): {
				{Endpoint: "1.1.1.1:11", IsLocal: false},
			},
			makeServicePortName("ns1", "ep1", "p12"): {
				{Endpoint: "1.1.1.2:12", IsLocal: false},
			},
		},
		expectedResult: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", "p11"): {
				{Endpoint: "1.1.1.1:11", IsLocal: false},
			},
			makeServicePortName("ns1", "ep1", "p12"): {
				{Endpoint: "1.1.1.2:12", IsLocal: false},
			},
		},
		expectedStaleEndpoints:    []proxy.ServiceEndpoint{},
		expectedStaleServiceNames: map[proxy.ServicePortName]bool{},
		expectedHealthchecks:      map[types.NamespacedName]int{},
	}, {
		// Case[4]: no change, multiple subsets, multiple ports, local
		previousEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", multipleSubsetsMultiplePortsLocal),
		},
		currentEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", multipleSubsetsMultiplePortsLocal),
		},
		oldEndpoints: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", "p11"): {
				{Endpoint: "1.1.1.1:11", IsLocal: true},
			},
			makeServicePortName("ns1", "ep1", "p12"): {
				{Endpoint: "1.1.1.1:12", IsLocal: true},
			},
			makeServicePortName("ns1", "ep1", "p13"): {
				{Endpoint: "1.1.1.3:13", IsLocal: false},
			},
		},
		expectedResult: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", "p11"): {
				{Endpoint: "1.1.1.1:11", IsLocal: true},
			},
			makeServicePortName("ns1", "ep1", "p12"): {
				{Endpoint: "1.1.1.1:12", IsLocal: true},
			},
			makeServicePortName("ns1", "ep1", "p13"): {
				{Endpoint: "1.1.1.3:13", IsLocal: false},
			},
		},
		expectedStaleEndpoints:    []proxy.ServiceEndpoint{},
		expectedStaleServiceNames: map[proxy.ServicePortName]bool{},
		expectedHealthchecks: map[types.NamespacedName]int{
			makeNSN("ns1", "ep1"): 1,
		},
	}, {
		// Case[5]: no change, multiple endpoints, subsets, IPs, and ports
		previousEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", multipleSubsetsIPsPorts1),
			makeTestEndpoints("ns2", "ep2", multipleSubsetsIPsPorts2),
		},
		currentEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", multipleSubsetsIPsPorts1),
			makeTestEndpoints("ns2", "ep2", multipleSubsetsIPsPorts2),
		},
		oldEndpoints: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", "p11"): {
				{Endpoint: "1.1.1.1:11", IsLocal: false},
				{Endpoint: "1.1.1.2:11", IsLocal: true},
			},
			makeServicePortName("ns1", "ep1", "p12"): {
				{Endpoint: "1.1.1.1:12", IsLocal: false},
				{Endpoint: "1.1.1.2:12", IsLocal: true},
			},
			makeServicePortName("ns1", "ep1", "p13"): {
				{Endpoint: "1.1.1.3:13", IsLocal: false},
				{Endpoint: "1.1.1.4:13", IsLocal: true},
			},
			makeServicePortName("ns1", "ep1", "p14"): {
				{Endpoint: "1.1.1.3:14", IsLocal: false},
				{Endpoint: "1.1.1.4:14", IsLocal: true},
			},
			makeServicePortName("ns2", "ep2", "p21"): {
				{Endpoint: "2.2.2.1:21", IsLocal: false},
				{Endpoint: "2.2.2.2:21", IsLocal: true},
			},
			makeServicePortName("ns2", "ep2", "p22"): {
				{Endpoint: "2.2.2.1:22", IsLocal: false},
				{Endpoint: "2.2.2.2:22", IsLocal: true},
			},
		},
		expectedResult: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", "p11"): {
				{Endpoint: "1.1.1.1:11", IsLocal: false},
				{Endpoint: "1.1.1.2:11", IsLocal: true},
			},
			makeServicePortName("ns1", "ep1", "p12"): {
				{Endpoint: "1.1.1.1:12", IsLocal: false},
				{Endpoint: "1.1.1.2:12", IsLocal: true},
			},
			makeServicePortName("ns1", "ep1", "p13"): {
				{Endpoint: "1.1.1.3:13", IsLocal: false},
				{Endpoint: "1.1.1.4:13", IsLocal: true},
			},
			makeServicePortName("ns1", "ep1", "p14"): {
				{Endpoint: "1.1.1.3:14", IsLocal: false},
				{Endpoint: "1.1.1.4:14", IsLocal: true},
			},
			makeServicePortName("ns2", "ep2", "p21"): {
				{Endpoint: "2.2.2.1:21", IsLocal: false},
				{Endpoint: "2.2.2.2:21", IsLocal: true},
			},
			makeServicePortName("ns2", "ep2", "p22"): {
				{Endpoint: "2.2.2.1:22", IsLocal: false},
				{Endpoint: "2.2.2.2:22", IsLocal: true},
			},
		},
		expectedStaleEndpoints:    []proxy.ServiceEndpoint{},
		expectedStaleServiceNames: map[proxy.ServicePortName]bool{},
		expectedHealthchecks: map[types.NamespacedName]int{
			makeNSN("ns1", "ep1"): 2,
			makeNSN("ns2", "ep2"): 1,
		},
	}, {
		// Case[6]: add an Endpoints
		previousEndpoints: []*api.Endpoints{
			nil,
		},
		currentEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", unnamedPortLocal),
		},
		oldEndpoints: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{},
		expectedResult: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", ""): {
				{Endpoint: "1.1.1.1:11", IsLocal: true},
			},
		},
		expectedStaleEndpoints: []proxy.ServiceEndpoint{},
		expectedStaleServiceNames: map[proxy.ServicePortName]bool{
			makeServicePortName("ns1", "ep1", ""): true,
		},
		expectedHealthchecks: map[types.NamespacedName]int{
			makeNSN("ns1", "ep1"): 1,
		},
	}, {
		// Case[7]: remove an Endpoints
		previousEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", unnamedPortLocal),
		},
		currentEndpoints: []*api.Endpoints{
			nil,
		},
		oldEndpoints: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", ""): {
				{Endpoint: "1.1.1.1:11", IsLocal: true},
			},
		},
		expectedResult: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{},
		expectedStaleEndpoints: []proxy.ServiceEndpoint{{
			Endpoint:        "1.1.1.1:11",
			ServicePortName: makeServicePortName("ns1", "ep1", ""),
		}},
		expectedStaleServiceNames: map[proxy.ServicePortName]bool{},
		expectedHealthchecks:      map[types.NamespacedName]int{},
	}, {
		// Case[8]: add an IP and port
		previousEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", namedPort),
		},
		currentEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", namedPortsLocalNoLocal),
		},
		oldEndpoints: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", "p11"): {
				{Endpoint: "1.1.1.1:11", IsLocal: false},
			},
		},
		expectedResult: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", "p11"): {
				{Endpoint: "1.1.1.1:11", IsLocal: false},
				{Endpoint: "1.1.1.2:11", IsLocal: true},
			},
			makeServicePortName("ns1", "ep1", "p12"): {
				{Endpoint: "1.1.1.1:12", IsLocal: false},
				{Endpoint: "1.1.1.2:12", IsLocal: true},
			},
		},
		expectedStaleEndpoints: []proxy.ServiceEndpoint{},
		expectedStaleServiceNames: map[proxy.ServicePortName]bool{
			makeServicePortName("ns1", "ep1", "p12"): true,
		},
		expectedHealthchecks: map[types.NamespacedName]int{
			makeNSN("ns1", "ep1"): 1,
		},
	}, {
		// Case[9]: remove an IP and port
		previousEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", namedPortsLocalNoLocal),
		},
		currentEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", namedPort),
		},
		oldEndpoints: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", "p11"): {
				{Endpoint: "1.1.1.1:11", IsLocal: false},
				{Endpoint: "1.1.1.2:11", IsLocal: true},
			},
			makeServicePortName("ns1", "ep1", "p12"): {
				{Endpoint: "1.1.1.1:12", IsLocal: false},
				{Endpoint: "1.1.1.2:12", IsLocal: true},
			},
		},
		expectedResult: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", "p11"): {
				{Endpoint: "1.1.1.1:11", IsLocal: false},
			},
		},
		expectedStaleEndpoints: []proxy.ServiceEndpoint{{
			Endpoint:        "1.1.1.2:11",
			ServicePortName: makeServicePortName("ns1", "ep1", "p11"),
		}, {
			Endpoint:        "1.1.1.1:12",
			ServicePortName: makeServicePortName("ns1", "ep1", "p12"),
		}, {
			Endpoint:        "1.1.1.2:12",
			ServicePortName: makeServicePortName("ns1", "ep1", "p12"),
		}},
		expectedStaleServiceNames: map[proxy.ServicePortName]bool{},
		expectedHealthchecks:      map[types.NamespacedName]int{},
	}, {
		// Case[10]: add a subset
		previousEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", namedPort),
		},
		currentEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", multipleSubsetsWithLocal),
		},
		oldEndpoints: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", "p11"): {
				{Endpoint: "1.1.1.1:11", IsLocal: false},
			},
		},
		expectedResult: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", "p11"): {
				{Endpoint: "1.1.1.1:11", IsLocal: false},
			},
			makeServicePortName("ns1", "ep1", "p12"): {
				{Endpoint: "1.1.1.2:12", IsLocal: true},
			},
		},
		expectedStaleEndpoints: []proxy.ServiceEndpoint{},
		expectedStaleServiceNames: map[proxy.ServicePortName]bool{
			makeServicePortName("ns1", "ep1", "p12"): true,
		},
		expectedHealthchecks: map[types.NamespacedName]int{
			makeNSN("ns1", "ep1"): 1,
		},
	}, {
		// Case[11]: remove a subset
		previousEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", multipleSubsets),
		},
		currentEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", namedPort),
		},
		oldEndpoints: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", "p11"): {
				{Endpoint: "1.1.1.1:11", IsLocal: false},
			},
			makeServicePortName("ns1", "ep1", "p12"): {
				{Endpoint: "1.1.1.2:12", IsLocal: false},
			},
		},
		expectedResult: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", "p11"): {
				{Endpoint: "1.1.1.1:11", IsLocal: false},
			},
		},
		expectedStaleEndpoints: []proxy.ServiceEndpoint{{
			Endpoint:        "1.1.1.2:12",
			ServicePortName: makeServicePortName("ns1", "ep1", "p12"),
		}},
		expectedStaleServiceNames: map[proxy.ServicePortName]bool{},
		expectedHealthchecks:      map[types.NamespacedName]int{},
	}, {
		// Case[12]: rename a port
		previousEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", namedPort),
		},
		currentEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", namedPortRenamed),
		},
		oldEndpoints: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", "p11"): {
				{Endpoint: "1.1.1.1:11", IsLocal: false},
			},
		},
		expectedResult: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", "p11-2"): {
				{Endpoint: "1.1.1.1:11", IsLocal: false},
			},
		},
		expectedStaleEndpoints: []proxy.ServiceEndpoint{{
			Endpoint:        "1.1.1.1:11",
			ServicePortName: makeServicePortName("ns1", "ep1", "p11"),
		}},
		expectedStaleServiceNames: map[proxy.ServicePortName]bool{
			makeServicePortName("ns1", "ep1", "p11-2"): true,
		},
		expectedHealthchecks: map[types.NamespacedName]int{},
	}, {
		// Case[13]: renumber a port
		previousEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", namedPort),
		},
		currentEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", namedPortRenumbered),
		},
		oldEndpoints: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", "p11"): {
				{Endpoint: "1.1.1.1:11", IsLocal: false},
			},
		},
		expectedResult: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", "p11"): {
				{Endpoint: "1.1.1.1:22", IsLocal: false},
			},
		},
		expectedStaleEndpoints: []proxy.ServiceEndpoint{{
			Endpoint:        "1.1.1.1:11",
			ServicePortName: makeServicePortName("ns1", "ep1", "p11"),
		}},
		expectedStaleServiceNames: map[proxy.ServicePortName]bool{},
		expectedHealthchecks:      map[types.NamespacedName]int{},
	}, {
		// Case[14]: complex add and remove
		previousEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", complexBefore1),
			makeTestEndpoints("ns2", "ep2", complexBefore2),
			nil,
			makeTestEndpoints("ns4", "ep4", complexBefore4),
		},
		currentEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", complexAfter1),
			nil,
			makeTestEndpoints("ns3", "ep3", complexAfter3),
			makeTestEndpoints("ns4", "ep4", complexAfter4),
		},
		oldEndpoints: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", "p11"): {
				{Endpoint: "1.1.1.1:11", IsLocal: false},
			},
			makeServicePortName("ns2", "ep2", "p22"): {
				{Endpoint: "2.2.2.2:22", IsLocal: true},
				{Endpoint: "2.2.2.22:22", IsLocal: true},
			},
			makeServicePortName("ns2", "ep2", "p23"): {
				{Endpoint: "2.2.2.3:23", IsLocal: true},
			},
			makeServicePortName("ns4", "ep4", "p44"): {
				{Endpoint: "4.4.4.4:44", IsLocal: true},
				{Endpoint: "4.4.4.5:44", IsLocal: true},
			},
			makeServicePortName("ns4", "ep4", "p45"): {
				{Endpoint: "4.4.4.6:45", IsLocal: true},
			},
		},
		expectedResult: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", "p11"): {
				{Endpoint: "1.1.1.1:11", IsLocal: false},
				{Endpoint: "1.1.1.11:11", IsLocal: false},
			},
			makeServicePortName("ns1", "ep1", "p12"): {
				{Endpoint: "1.1.1.2:12", IsLocal: false},
			},
			makeServicePortName("ns1", "ep1", "p122"): {
				{Endpoint: "1.1.1.2:122", IsLocal: false},
			},
			makeServicePortName("ns3", "ep3", "p33"): {
				{Endpoint: "3.3.3.3:33", IsLocal: false},
			},
			makeServicePortName("ns4", "ep4", "p44"): {
				{Endpoint: "4.4.4.4:44", IsLocal: true},
			},
		},
		expectedStaleEndpoints: []proxy.ServiceEndpoint{{
			Endpoint:        "2.2.2.2:22",
			ServicePortName: makeServicePortName("ns2", "ep2", "p22"),
		}, {
			Endpoint:        "2.2.2.22:22",
			ServicePortName: makeServicePortName("ns2", "ep2", "p22"),
		}, {
			Endpoint:        "2.2.2.3:23",
			ServicePortName: makeServicePortName("ns2", "ep2", "p23"),
		}, {
			Endpoint:        "4.4.4.5:44",
			ServicePortName: makeServicePortName("ns4", "ep4", "p44"),
		}, {
			Endpoint:        "4.4.4.6:45",
			ServicePortName: makeServicePortName("ns4", "ep4", "p45"),
		}},
		expectedStaleServiceNames: map[proxy.ServicePortName]bool{
			makeServicePortName("ns1", "ep1", "p12"):  true,
			makeServicePortName("ns1", "ep1", "p122"): true,
			makeServicePortName("ns3", "ep3", "p33"):  true,
		},
		expectedHealthchecks: map[types.NamespacedName]int{
			makeNSN("ns4", "ep4"): 1,
		},
	}, {
		// Case[15]: change from 0 endpoint address to 1 unnamed port
		previousEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", emptyEndpoint),
		},
		currentEndpoints: []*api.Endpoints{
			makeTestEndpoints("ns1", "ep1", unnamedPort),
		},
		oldEndpoints: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{},
		expectedResult: map[proxy.ServicePortName][]*proxy.BaseEndpointInfo{
			makeServicePortName("ns1", "ep1", ""): {
				{Endpoint: "1.1.1.1:11", IsLocal: false},
			},
		},
		expectedStaleEndpoints: []proxy.ServiceEndpoint{},
		expectedStaleServiceNames: map[proxy.ServicePortName]bool{
			makeServicePortName("ns1", "ep1", ""): true,
		},
		expectedHealthchecks: map[types.NamespacedName]int{},
	},
	}

	for tci, tc := range testCases {
		ipt := iptablestest.NewFake()
		ipvs := ipvstest.NewFake()
		ipset := ipsettest.NewFake(testIPSetVersion)
		fp := NewFakeProxier(ipt, ipvs, ipset, nil)
		fp.hostname = nodeName

		// First check that after adding all previous versions of endpoints,
		// the fp.oldEndpoints is as we expect.
		for i := range tc.previousEndpoints {
			if tc.previousEndpoints[i] != nil {
				fp.OnEndpointsAdd(tc.previousEndpoints[i])
			}
		}
		proxy.UpdateEndpointsMap(fp.endpointsMap, fp.endpointsChanges)
		compareEndpointsMaps(t, tci, fp.endpointsMap, tc.oldEndpoints)

		// Now let's call appropriate handlers to get to state we want to be.
		if len(tc.previousEndpoints) != len(tc.currentEndpoints) {
			t.Fatalf("[%d] different lengths of previous and current endpoints", tci)
			continue
		}

		for i := range tc.previousEndpoints {
			prev, curr := tc.previousEndpoints[i], tc.currentEndpoints[i]
			switch {
			case prev == nil:
				fp.OnEndpointsAdd(curr)
			case curr == nil:
				fp.OnEndpointsDelete(prev)
			default:
				fp.OnEndpointsUpdate(prev, curr)
			}
		}
		result := proxy.UpdateEndpointsMap(fp.endpointsMap, fp.endpointsChanges)
		newMap := fp.endpointsMap
		compareEndpointsMaps(t, tci, newMap, tc.expectedResult)
		if len(result.StaleEndpoints) != len(tc.expectedStaleEndpoints) {
			t.Errorf("[%d] expected %d staleEndpoints, got %d: %v", tci, len(tc.expectedStaleEndpoints), len(result.StaleEndpoints), result.StaleEndpoints)
		}
		for _, x := range tc.expectedStaleEndpoints {
			found := false
			for _, stale := range result.StaleEndpoints {
				if stale == x {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("[%d] expected staleEndpoints[%v], but didn't find it: %v", tci, x, result.StaleEndpoints)
			}
		}
		if len(result.StaleServiceNames) != len(tc.expectedStaleServiceNames) {
			t.Errorf("[%d] expected %d staleServiceNames, got %d: %v", tci, len(tc.expectedStaleServiceNames), len(result.StaleServiceNames), result.StaleServiceNames)
		}
		for svcName := range tc.expectedStaleServiceNames {
			found := false
			for _, stale := range result.StaleServiceNames {
				if stale == svcName {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("[%d] expected staleServiceNames[%v], but didn't find it: %v", tci, svcName, result.StaleServiceNames)
			}
		}
		if !reflect.DeepEqual(result.HCEndpointsLocalIPSize, tc.expectedHealthchecks) {
			t.Errorf("[%d] expected healthchecks %v, got %v", tci, tc.expectedHealthchecks, result.HCEndpointsLocalIPSize)
		}
	}
}

func compareEndpointsMaps(t *testing.T, tci int, newMap proxy.EndpointsMap, expected map[proxy.ServicePortName][]*proxy.BaseEndpointInfo) {
	if len(newMap) != len(expected) {
		t.Errorf("[%d] expected %d results, got %d: %v", tci, len(expected), len(newMap), newMap)
	}
	for x := range expected {
		if len(newMap[x]) != len(expected[x]) {
			t.Errorf("[%d] expected %d endpoints for %v, got %d", tci, len(expected[x]), x, len(newMap[x]))
		} else {
			for i := range expected[x] {
				newEp, ok := newMap[x][i].(*proxy.BaseEndpointInfo)
				if !ok {
					t.Errorf("Failed to cast proxy.BaseEndpointInfo")
					continue
				}
				if *newEp != *(expected[x][i]) {
					t.Errorf("[%d] expected new[%v][%d] to be %v, got %v", tci, x, i, expected[x][i], newEp)
				}
			}
		}
	}
}

func Test_syncService(t *testing.T) {
	testCases := []struct {
		oldVirtualServer *utilipvs.VirtualServer
		svcName          string
		newVirtualServer *utilipvs.VirtualServer
		bindAddr         bool
	}{
		{
			// case 0, old virtual server is same as new virtual server
			oldVirtualServer: &utilipvs.VirtualServer{
				Address:   net.ParseIP("1.2.3.4"),
				Protocol:  string(api.ProtocolTCP),
				Port:      80,
				Scheduler: "rr",
				Flags:     utilipvs.FlagHashed,
			},
			svcName: "foo",
			newVirtualServer: &utilipvs.VirtualServer{
				Address:   net.ParseIP("1.2.3.4"),
				Protocol:  string(api.ProtocolTCP),
				Port:      80,
				Scheduler: "rr",
				Flags:     utilipvs.FlagHashed,
			},
			bindAddr: false,
		},
		{
			// case 1, old virtual server is different from new virtual server
			oldVirtualServer: &utilipvs.VirtualServer{
				Address:   net.ParseIP("1.2.3.4"),
				Protocol:  string(api.ProtocolTCP),
				Port:      8080,
				Scheduler: "rr",
				Flags:     utilipvs.FlagHashed,
			},
			svcName: "bar",
			newVirtualServer: &utilipvs.VirtualServer{
				Address:   net.ParseIP("1.2.3.4"),
				Protocol:  string(api.ProtocolTCP),
				Port:      8080,
				Scheduler: "rr",
				Flags:     utilipvs.FlagPersistent,
			},
			bindAddr: false,
		},
		{
			// case 2, old virtual server is different from new virtual server
			oldVirtualServer: &utilipvs.VirtualServer{
				Address:   net.ParseIP("1.2.3.4"),
				Protocol:  string(api.ProtocolTCP),
				Port:      8080,
				Scheduler: "rr",
				Flags:     utilipvs.FlagHashed,
			},
			svcName: "bar",
			newVirtualServer: &utilipvs.VirtualServer{
				Address:   net.ParseIP("1.2.3.4"),
				Protocol:  string(api.ProtocolTCP),
				Port:      8080,
				Scheduler: "wlc",
				Flags:     utilipvs.FlagHashed,
			},
			bindAddr: false,
		},
		{
			// case 3, old virtual server is nil, and create new virtual server
			oldVirtualServer: nil,
			svcName:          "baz",
			newVirtualServer: &utilipvs.VirtualServer{
				Address:   net.ParseIP("1.2.3.4"),
				Protocol:  string(api.ProtocolUDP),
				Port:      53,
				Scheduler: "rr",
				Flags:     utilipvs.FlagHashed,
			},
			bindAddr: true,
		},
	}

	for i := range testCases {
		ipt := iptablestest.NewFake()
		ipvs := ipvstest.NewFake()
		ipset := ipsettest.NewFake(testIPSetVersion)
		proxier := NewFakeProxier(ipt, ipvs, ipset, nil)

		if testCases[i].oldVirtualServer != nil {
			if err := proxier.ipvs.AddVirtualServer(testCases[i].oldVirtualServer); err != nil {
				t.Errorf("Case [%d], unexpected add IPVS virtual server error: %v", i, err)
			}
		}
		if err := proxier.syncService(testCases[i].svcName, testCases[i].newVirtualServer, testCases[i].bindAddr); err != nil {
			t.Errorf("Case [%d], unexpected sync IPVS virtual server error: %v", i, err)
		}
		// check
		list, err := proxier.ipvs.GetVirtualServers()
		if err != nil {
			t.Errorf("Case [%d], unexpected list IPVS virtual server error: %v", i, err)
		}
		if len(list) != 1 {
			t.Errorf("Case [%d], expect %d virtual servers, got %d", i, 1, len(list))
			continue
		}
		if !list[0].Equal(testCases[i].newVirtualServer) {
			t.Errorf("Case [%d], unexpected mismatch, expect: %#v, got: %#v", i, testCases[i].newVirtualServer, list[0])
		}
	}
}

func Test_cleanLegacyService(t *testing.T) {
	// All ipvs services that were processed in the latest sync loop.
	activeServices := map[string]bool{"ipvs0": true, "ipvs1": true}
	// All ipvs services in the system.
	currentServices := map[string]*utilipvs.VirtualServer{
		// Created by kube-proxy.
		"ipvs0": {
			Address:   net.ParseIP("1.1.1.1"),
			Protocol:  string(api.ProtocolUDP),
			Port:      53,
			Scheduler: "rr",
			Flags:     utilipvs.FlagHashed,
		},
		// Created by kube-proxy.
		"ipvs1": {
			Address:   net.ParseIP("2.2.2.2"),
			Protocol:  string(api.ProtocolUDP),
			Port:      54,
			Scheduler: "rr",
			Flags:     utilipvs.FlagHashed,
		},
		// Created by an external party.
		"ipvs2": {
			Address:   net.ParseIP("3.3.3.3"),
			Protocol:  string(api.ProtocolUDP),
			Port:      55,
			Scheduler: "rr",
			Flags:     utilipvs.FlagHashed,
		},
		// Created by an external party.
		"ipvs3": {
			Address:   net.ParseIP("4.4.4.4"),
			Protocol:  string(api.ProtocolUDP),
			Port:      56,
			Scheduler: "rr",
			Flags:     utilipvs.FlagHashed,
		},
		// Created by an external party.
		"ipvs4": {
			Address:   net.ParseIP("5.5.5.5"),
			Protocol:  string(api.ProtocolUDP),
			Port:      57,
			Scheduler: "rr",
			Flags:     utilipvs.FlagHashed,
		},
		// Created by kube-proxy, but now stale.
		"ipvs5": {
			Address:   net.ParseIP("6.6.6.6"),
			Protocol:  string(api.ProtocolUDP),
			Port:      58,
			Scheduler: "rr",
			Flags:     utilipvs.FlagHashed,
		},
	}

	ipt := iptablestest.NewFake()
	ipvs := ipvstest.NewFake()
	ipset := ipsettest.NewFake(testIPSetVersion)
	proxier := NewFakeProxier(ipt, ipvs, ipset, nil)
	// These CIDRs cover only ipvs2 and ipvs3.
	proxier.excludeCIDRs = []string{"3.3.3.0/24", "4.4.4.0/24"}
	for v := range currentServices {
		proxier.ipvs.AddVirtualServer(currentServices[v])
	}
	proxier.cleanLegacyService(activeServices, currentServices)
	// ipvs4 and ipvs5 should have been cleaned.
	remainingVirtualServers, _ := proxier.ipvs.GetVirtualServers()
	if len(remainingVirtualServers) != 4 {
		t.Errorf("Expected number of remaining IPVS services after cleanup to be %v. Got %v", 4, len(remainingVirtualServers))
	}
	for _, vs := range remainingVirtualServers {
		// Checking that ipvs4 and ipvs5 were removed.
		if vs.Port == 57 {
			t.Errorf("Expected ipvs4 to be removed after cleanup. It still remains")
		}
		if vs.Port == 58 {
			t.Errorf("Expected ipvs5 to be removed after cleanup. It still remains")
		}
	}
}

func buildFakeProxier(nodeIP []net.IP) (*iptablestest.FakeIPTables, *Proxier) {
	ipt := iptablestest.NewFake()
	ipvs := ipvstest.NewFake()
	ipset := ipsettest.NewFake(testIPSetVersion)
	return ipt, NewFakeProxier(ipt, ipvs, ipset, nil)
}

func hasJump(rules []iptablestest.Rule, destChain, ipSet string) bool {
	for _, r := range rules {
		if r[iptablestest.Jump] == destChain {
			if ipSet == "" {
				return true
			}
			if strings.Contains(r[iptablestest.MatchSet], ipSet) {
				return true
			}
		}
	}
	return false
}

// checkIptabless to check expected iptables chain and rules
func checkIptables(t *testing.T, ipt *iptablestest.FakeIPTables, epIpt netlinktest.ExpectedIptablesChain) {
	for epChain, epRules := range epIpt {
		rules := ipt.GetRules(epChain)
		for _, epRule := range epRules {
			if !hasJump(rules, epRule.JumpChain, epRule.MatchSet) {
				t.Errorf("Didn't find jump from chain %v match set %v to %v", epChain, epRule.MatchSet, epRule.JumpChain)
			}
		}
	}
}

// checkIPSet to check expected ipset and entries
func checkIPSet(t *testing.T, fp *Proxier, ipSet netlinktest.ExpectedIPSet) {
	for set, entries := range ipSet {
		ents, err := fp.ipset.ListEntries(set)
		if err != nil || len(ents) != len(entries) {
			t.Errorf("Check ipset entries failed for ipset: %q, expect %d, got %d", set, len(entries), len(ents))
			continue
		}
		if len(entries) == 1 {
			if ents[0] != entries[0].String() {
				t.Errorf("Check ipset entries failed for ipset: %q", set)
			}
		}
	}
}

// checkIPVS to check expected ipvs service and destination
func checkIPVS(t *testing.T, fp *Proxier, vs *netlinktest.ExpectedVirtualServer) {
	services, err := fp.ipvs.GetVirtualServers()
	if err != nil {
		t.Errorf("Failed to get ipvs services, err: %v", err)
	}
	if len(services) != vs.VSNum {
		t.Errorf("Expect %d ipvs services, got %d", vs.VSNum, len(services))
	}
	for _, svc := range services {
		if svc.Address.String() == vs.IP && svc.Port == vs.Port && svc.Protocol == vs.Protocol {
			destinations, _ := fp.ipvs.GetRealServers(svc)
			if len(destinations) != len(vs.RS) {
				t.Errorf("Expected %d destinations, got %d destinations", len(vs.RS), len(destinations))
			}
			if len(vs.RS) == 1 {
				if destinations[0].Address.String() != vs.RS[0].IP || destinations[0].Port != vs.RS[0].Port {
					t.Errorf("Unexpected mismatch destinations")
				}
			}
		}
	}
}
