package models

import "k8s.io/api/core/v1"

type Addresses []Address
type Address struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
	IP   string `json:"ip"`
}

func (addresses *Addresses) Parse(as []v1.EndpointAddress) {
	for _, address := range as {
		castedAddress := Address{}
		castedAddress.Parse(address)
		*addresses = append(*addresses, castedAddress)
	}
}

func (address *Address) Parse(a v1.EndpointAddress) {
	address.IP = a.IP

	if a.TargetRef != nil {
		address.Kind = a.TargetRef.Kind
		address.Name = a.TargetRef.Name
	}
}
