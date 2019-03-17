package models

// IstioConfigList istioConfigList
//
// This type is used for returning a response of IstioConfigList
//
// swagger:model IstioConfigList
type IstioConfigList struct {
	// The namespace of istioConfiglist
	//
	// required: true
	Namespace           Namespace           `json:"namespace"`
	Gateways            Gateways            `json:"gateways"`
	VirtualServices     VirtualServices     `json:"virtualServices"`
	DestinationRules    DestinationRules    `json:"destinationRules"`
	ServiceEntries      ServiceEntries      `json:"serviceEntries"`
	Rules               IstioRules          `json:"rules"`
	Adapters            IstioAdapters       `json:"adapters"`
	Templates           IstioTemplates      `json:"templates"`
	QuotaSpecs          QuotaSpecs          `json:"quotaSpecs"`
	QuotaSpecBindings   QuotaSpecBindings   `json:"quotaSpecBindings"`
	Policies            Policies            `json:"policies"`
	MeshPolicies        MeshPolicies        `json:"meshPolicies"`
	ClusterRbacConfigs  ClusterRbacConfigs  `json:"clusterRbacConfigs"`
	ServiceRoles        ServiceRoles        `json:"serviceRoles"`
	ServiceRoleBindings ServiceRoleBindings `json:"serviceRoleBindings"`
	IstioValidations    IstioValidations    `json:"validations"`
}

type IstioConfigDetails struct {
	Namespace          Namespace           `json:"namespace"`
	ObjectType         string              `json:"objectType"`
	Gateway            *Gateway            `json:"gateway"`
	VirtualService     *VirtualService     `json:"virtualService"`
	DestinationRule    *DestinationRule    `json:"destinationRule"`
	ServiceEntry       *ServiceEntry       `json:"serviceEntry"`
	Rule               *IstioRule          `json:"rule"`
	Adapter            *IstioAdapter       `json:"adapter"`
	Template           *IstioTemplate      `json:"template"`
	QuotaSpec          *QuotaSpec          `json:"quotaSpec"`
	QuotaSpecBinding   *QuotaSpecBinding   `json:"quotaSpecBinding"`
	Policy             *Policy             `json:"policy"`
	MeshPolicy         *MeshPolicy         `json:"meshPolicy"`
	ClusterRbacConfig  *ClusterRbacConfig  `json:"clusterRbacConfig"`
	ServiceRole        *ServiceRole        `json:"serviceRole"`
	ServiceRoleBinding *ServiceRoleBinding `json:"serviceRoleBinding"`
	Permissions        ResourcePermissions `json:"permissions"`
	IstioValidation    *IstioValidation    `json:"validation"`
}

// ResourcePermissions holds permission flags for an object type
// True means allowed.
type ResourcePermissions struct {
	Create bool `json:"create"`
	Update bool `json:"update"`
	Delete bool `json:"delete"`
}
