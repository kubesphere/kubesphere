package status

import (
	"github.com/kiali/kiali/business"
)

func (si *StatusInfo) getmTLSStatus() {
	// Get business layer
	business, err := business.Get()
	if err != nil {
		Put(ClusterMTLS, "error")
		return
	}

	namespaces, err := business.Namespace.GetNamespaces()
	if err != nil {
		Put(ClusterMTLS, "error")
		return
	}

	nsNames := make([]string, 0, len(namespaces))
	for _, ns := range namespaces {
		nsNames = append(nsNames, ns.Name)
	}

	globalmTLSStatus, err := business.IstioConfig.MeshWidemTLSStatus(nsNames)
	if err != nil {
		Put(ClusterMTLS, "error")
		return
	}

	Put(ClusterMTLS, globalmTLSStatus)
}
