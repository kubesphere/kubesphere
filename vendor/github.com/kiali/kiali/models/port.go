package models

import "k8s.io/api/core/v1"

type Ports []Port
type Port struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Port     int32  `json:"port"`
}

func (ports *Ports) Parse(ps []v1.ServicePort) {
	for _, servicePort := range ps {
		port := Port{}
		port.Parse(servicePort)
		*ports = append(*ports, port)
	}
}

func (port *Port) Parse(p v1.ServicePort) {
	port.Name = p.Name
	port.Protocol = string(p.Protocol)
	port.Port = p.Port
}

func (ports *Ports) ParseEndpointPorts(ps []v1.EndpointPort) {
	for _, endpointPort := range ps {
		port := Port{}
		port.ParseEndpointPort(endpointPort)
		*ports = append(*ports, port)
	}
}

func (port *Port) ParseEndpointPort(p v1.EndpointPort) {
	port.Name = p.Name
	port.Protocol = string(p.Protocol)
	port.Port = p.Port
}
