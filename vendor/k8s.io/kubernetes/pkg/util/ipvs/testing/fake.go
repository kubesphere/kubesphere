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

package testing

import (
	"fmt"
	"net"
	"strconv"

	utilipvs "k8s.io/kubernetes/pkg/util/ipvs"
)

//FakeIPVS no-op implementation of ipvs Interface
type FakeIPVS struct {
	Scheduler    string
	Services     map[serviceKey]*utilipvs.VirtualServer
	Destinations map[serviceKey][]*utilipvs.RealServer
}

type serviceKey struct {
	IP       string
	Port     uint16
	Protocol string
}

func (s *serviceKey) String() string {
	return fmt.Sprintf("%s:%d/%s", s.IP, s.Port, s.Protocol)
}

type realServerKey struct {
	Address net.IP
	Port    uint16
}

func (r *realServerKey) String() string {
	return net.JoinHostPort(r.Address.String(), strconv.Itoa(int(r.Port)))
}

//NewFake creates a fake ipvs implementation - a cache store.
func NewFake() *FakeIPVS {
	return &FakeIPVS{
		Services:     make(map[serviceKey]*utilipvs.VirtualServer),
		Destinations: make(map[serviceKey][]*utilipvs.RealServer),
	}
}

func toServiceKey(serv *utilipvs.VirtualServer) serviceKey {
	return serviceKey{
		IP:       serv.Address.String(),
		Port:     serv.Port,
		Protocol: serv.Protocol,
	}
}

func toRealServerKey(rs *utilipvs.RealServer) *realServerKey {
	return &realServerKey{
		Address: rs.Address,
		Port:    rs.Port,
	}
}

//AddVirtualServer is a fake implementation, it simply adds the VirtualServer into the cache store.
func (f *FakeIPVS) AddVirtualServer(serv *utilipvs.VirtualServer) error {
	if serv == nil {
		return fmt.Errorf("Failed to add virtual server, error: virtual server can't be nil")
	}
	key := toServiceKey(serv)
	f.Services[key] = serv
	// make sure no destination present when creating new service
	f.Destinations[key] = make([]*utilipvs.RealServer, 0)
	return nil
}

//UpdateVirtualServer is a fake implementation, it updates the VirtualServer in the cache store.
func (f *FakeIPVS) UpdateVirtualServer(serv *utilipvs.VirtualServer) error {
	if serv == nil {
		return fmt.Errorf("Failed to update service, service can't be nil")
	}
	key := toServiceKey(serv)
	f.Services[key] = serv
	return nil
}

//DeleteVirtualServer is a fake implementation, it simply deletes the VirtualServer from the cache store.
func (f *FakeIPVS) DeleteVirtualServer(serv *utilipvs.VirtualServer) error {
	if serv == nil {
		return fmt.Errorf("Failed to delete service: service can't be nil")
	}
	key := toServiceKey(serv)
	delete(f.Services, key)
	// clear specific destinations as well
	f.Destinations[key] = nil
	return nil
}

//GetVirtualServer is a fake implementation, it tries to find a specific VirtualServer from the cache store.
func (f *FakeIPVS) GetVirtualServer(serv *utilipvs.VirtualServer) (*utilipvs.VirtualServer, error) {
	if serv == nil {
		return nil, fmt.Errorf("Failed to get service: service can't be nil")
	}
	key := toServiceKey(serv)
	svc, found := f.Services[key]
	if found {
		return svc, nil
	}
	return nil, fmt.Errorf("Not found serv: %v", key.String())
}

//GetVirtualServers is a fake implementation, it simply returns all VirtualServers in the cache store.
func (f *FakeIPVS) GetVirtualServers() ([]*utilipvs.VirtualServer, error) {
	res := make([]*utilipvs.VirtualServer, 0)
	for _, svc := range f.Services {
		res = append(res, svc)
	}
	return res, nil
}

//Flush is a fake implementation, it simply clears the cache store.
func (f *FakeIPVS) Flush() error {
	// directly drop old data
	f.Services = nil
	f.Destinations = nil
	return nil
}

//AddRealServer is a fake implementation, it simply creates a RealServer for a VirtualServer in the cache store.
func (f *FakeIPVS) AddRealServer(serv *utilipvs.VirtualServer, dest *utilipvs.RealServer) error {
	if serv == nil || dest == nil {
		return fmt.Errorf("Failed to add destination for service, neither service nor destination shouldn't be nil")
	}
	key := toServiceKey(serv)
	if _, ok := f.Services[key]; !ok {
		return fmt.Errorf("Failed to add destination for service %v, service not found", key.String())
	}
	dests := f.Destinations[key]
	if dests == nil {
		dests = make([]*utilipvs.RealServer, 0)
		f.Destinations[key] = dests
	}
	f.Destinations[key] = append(f.Destinations[key], dest)
	return nil
}

//GetRealServers is a fake implementation, it simply returns all RealServers in the cache store.
func (f *FakeIPVS) GetRealServers(serv *utilipvs.VirtualServer) ([]*utilipvs.RealServer, error) {
	if serv == nil {
		return nil, fmt.Errorf("Failed to get destination for nil service")
	}
	key := toServiceKey(serv)
	if _, ok := f.Services[key]; !ok {
		return nil, fmt.Errorf("Failed to get destinations for service %v, service not found", key.String())
	}
	return f.Destinations[key], nil
}

//DeleteRealServer is a fake implementation, it deletes the real server in the cache store.
func (f *FakeIPVS) DeleteRealServer(serv *utilipvs.VirtualServer, dest *utilipvs.RealServer) error {
	if serv == nil || dest == nil {
		return fmt.Errorf("Failed to delete destination, neither service nor destination can't be nil")
	}
	key := toServiceKey(serv)
	if _, ok := f.Services[key]; !ok {
		return fmt.Errorf("Failed to delete destination for service %v, service not found", key.String())
	}
	dests := f.Destinations[key]
	exist := false
	for i := range dests {
		if toRealServerKey(dests[i]).String() == toRealServerKey(dest).String() {
			// Delete one element
			f.Destinations[key] = append(f.Destinations[key][:i], f.Destinations[key][i+1:]...)
			exist = true
			break
		}
	}
	// Not Found
	if !exist {
		return fmt.Errorf("Failed to delete real server for service %v, real server not found", key.String())
	}
	return nil
}

var _ = utilipvs.Interface(&FakeIPVS{})
