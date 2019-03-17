package models

import "k8s.io/api/core/v1"

type Endpoints []Endpoint
type Endpoint struct {
	Addresses Addresses `json:"addresses"`
	Ports     Ports     `json:"ports"`
}

func (endpoints *Endpoints) Parse(es *v1.Endpoints) {
	if es == nil {
		return
	}

	for _, subset := range es.Subsets {
		endpoint := Endpoint{}
		endpoint.Parse(subset)
		*endpoints = append(*endpoints, endpoint)
	}
}

func (endpoint *Endpoint) Parse(s v1.EndpointSubset) {
	(&endpoint.Ports).ParseEndpointPorts(s.Ports)
	(&endpoint.Addresses).Parse(s.Addresses)
}
