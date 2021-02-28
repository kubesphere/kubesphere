package constant

const (
	FinalizerName     string = "finalizer.lb.kubesphere.io/v1alpha1"
	IPAMFinalizerName string = "finalizer.ipam.kubesphere.io/v1alpha1"

	// When used for annotation, it means that the service address is assigned by the porter
	// When used as a label, it indicates on which node the porter manager is deployed
	PorterAnnotationKey   string = "lb.kubesphere.io/v1alpha1"
	PorterAnnotationValue string = "porter"

	//Indicates the node to which layer2 traffic is sent
	PorterLayer2Annotation string = "layer2.porter.kubesphere.io/v1alpha1"

	PorterEIPAnnotationKey         string = "eip.porter.kubesphere.io/v1alpha1"
	PorterEIPAnnotationKeyV1Alpha2 string = "eip.porter.kubesphere.io/v1alpha2"

	PorterProtocolAnnotationKey string = "protocol.porter.kubesphere.io/v1alpha1"

	PorterNodeRack string = "porter.kubesphere.io/rack"

	PorterProtocolBGP    string = "bgp"
	PorterProtocolLayer2 string = "layer2"
	EipRangeSeparator    string = "-"

	PorterSpeakerLocker = "porter-speaker"
	PorterNamespace     = "porter-system"

	EnvPorterNamespace = "PORTER_NAMESPACE"
	EnvNodeName        = "NODE_NAME"
)
