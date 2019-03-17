package business

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	errors2 "k8s.io/apimachinery/pkg/api/errors"

	"github.com/kiali/kiali/config"
	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/log"
	"github.com/kiali/kiali/models"
	"github.com/kiali/kiali/prometheus/internalmetrics"
)

type IstioConfigService struct {
	k8s kubernetes.IstioClientInterface
}

type IstioConfigCriteria struct {
	Namespace                  string
	IncludeGateways            bool
	IncludeVirtualServices     bool
	IncludeDestinationRules    bool
	IncludeServiceEntries      bool
	IncludeRules               bool
	IncludeAdapters            bool
	IncludeTemplates           bool
	IncludeQuotaSpecs          bool
	IncludeQuotaSpecBindings   bool
	IncludePolicies            bool
	IncludeMeshPolicies        bool
	IncludeClusterRbacConfigs  bool
	IncludeServiceRoles        bool
	IncludeServiceRoleBindings bool
}

const (
	VirtualServices     = "virtualservices"
	DestinationRules    = "destinationrules"
	ServiceEntries      = "serviceentries"
	Gateways            = "gateways"
	Rules               = "rules"
	Adapters            = "adapters"
	Templates           = "templates"
	QuotaSpecs          = "quotaspecs"
	QuotaSpecBindings   = "quotaspecbindings"
	Policies            = "policies"
	MeshPolicies        = "meshpolicies"
	ClusterRbacConfigs  = "clusterrbacconfigs"
	ServiceRoles        = "serviceroles"
	ServiceRoleBindings = "servicerolebindings"
)

var resourceTypesToAPI = map[string]string{
	DestinationRules:    kubernetes.NetworkingGroupVersion.Group,
	VirtualServices:     kubernetes.NetworkingGroupVersion.Group,
	ServiceEntries:      kubernetes.NetworkingGroupVersion.Group,
	Gateways:            kubernetes.NetworkingGroupVersion.Group,
	Adapters:            kubernetes.ConfigGroupVersion.Group,
	Templates:           kubernetes.ConfigGroupVersion.Group,
	Rules:               kubernetes.ConfigGroupVersion.Group,
	QuotaSpecs:          kubernetes.ConfigGroupVersion.Group,
	QuotaSpecBindings:   kubernetes.ConfigGroupVersion.Group,
	Policies:            kubernetes.AuthenticationGroupVersion.Group,
	MeshPolicies:        kubernetes.AuthenticationGroupVersion.Group,
	ClusterRbacConfigs:  kubernetes.RbacGroupVersion.Group,
	ServiceRoles:        kubernetes.RbacGroupVersion.Group,
	ServiceRoleBindings: kubernetes.RbacGroupVersion.Group,
}

var apiToVersion = map[string]string{
	kubernetes.NetworkingGroupVersion.Group: kubernetes.ApiNetworkingVersion,
	kubernetes.ConfigGroupVersion.Group:     kubernetes.ApiConfigVersion,
	kubernetes.ApiAuthenticationVersion:     kubernetes.ApiAuthenticationVersion,
	kubernetes.RbacGroupVersion.Group:       kubernetes.ApiRbacVersion,
}

const (
	MeshmTLSEnabled          = "MESH_MTLS_ENABLED"
	MeshmTLSPartiallyEnabled = "MESH_MTLS_PARTIALLY_ENABLED"
	MeshmTLSNotEnabled       = "MESH_MTLS_NOT_ENABLED"
)

// GetIstioConfigList returns a list of Istio routing objects, Mixer Rules, (etc.)
// per a given Namespace.
func (in *IstioConfigService) GetIstioConfigList(criteria IstioConfigCriteria) (models.IstioConfigList, error) {
	var err error
	promtimer := internalmetrics.GetGoFunctionMetric("business", "IstioConfigService", "GetIstioConfigList")
	defer promtimer.ObserveNow(&err)

	if criteria.Namespace == "" {
		return models.IstioConfigList{}, errors.New("GetIstioConfigList needs a non empty Namespace")
	}
	istioConfigList := models.IstioConfigList{
		Namespace:           models.Namespace{Name: criteria.Namespace},
		Gateways:            models.Gateways{},
		VirtualServices:     models.VirtualServices{Items: []models.VirtualService{}},
		DestinationRules:    models.DestinationRules{Items: []models.DestinationRule{}},
		ServiceEntries:      models.ServiceEntries{},
		Rules:               models.IstioRules{},
		Adapters:            models.IstioAdapters{},
		Templates:           models.IstioTemplates{},
		QuotaSpecs:          models.QuotaSpecs{},
		QuotaSpecBindings:   models.QuotaSpecBindings{},
		Policies:            models.Policies{},
		MeshPolicies:        models.MeshPolicies{},
		ClusterRbacConfigs:  models.ClusterRbacConfigs{},
		ServiceRoles:        models.ServiceRoles{},
		ServiceRoleBindings: models.ServiceRoleBindings{},
	}
	var gg, vs, dr, se, qs, qb, aa, tt, mr, pc, mp, rc, sr, srb []kubernetes.IstioObject
	var ggErr, vsErr, drErr, seErr, mrErr, qsErr, qbErr, aaErr, ttErr, pcErr, mpErr, rcErr, srErr, srbErr error
	var wg sync.WaitGroup
	wg.Add(14)

	go func() {
		defer wg.Done()
		if criteria.IncludeGateways {
			if gg, ggErr = in.k8s.GetGateways(criteria.Namespace); ggErr == nil {
				(&istioConfigList.Gateways).Parse(gg)
			}
		}
	}()

	go func() {
		defer wg.Done()
		if criteria.IncludeVirtualServices {
			if vs, vsErr = in.k8s.GetVirtualServices(criteria.Namespace, ""); vsErr == nil {
				(&istioConfigList.VirtualServices).Parse(vs)
			}
		}
	}()

	go func() {
		defer wg.Done()
		if criteria.IncludeDestinationRules {
			if dr, drErr = in.k8s.GetDestinationRules(criteria.Namespace, ""); drErr == nil {
				(&istioConfigList.DestinationRules).Parse(dr)
			}
		}
	}()

	go func() {
		defer wg.Done()
		if criteria.IncludeServiceEntries {
			if se, seErr = in.k8s.GetServiceEntries(criteria.Namespace); seErr == nil {
				(&istioConfigList.ServiceEntries).Parse(se)
			}
		}
	}()

	go func() {
		defer wg.Done()
		if criteria.IncludeRules {
			if mr, mrErr = in.k8s.GetIstioRules(criteria.Namespace); mrErr == nil {
				istioConfigList.Rules = models.CastIstioRulesCollection(mr)
			}
		}
	}()

	go func() {
		defer wg.Done()
		if criteria.IncludeAdapters {
			if aa, aaErr = in.k8s.GetAdapters(criteria.Namespace); aaErr == nil {
				istioConfigList.Adapters = models.CastIstioAdaptersCollection(aa)
			}
		}
	}()

	go func() {
		defer wg.Done()
		if criteria.IncludeTemplates {
			if tt, ttErr = in.k8s.GetTemplates(criteria.Namespace); ttErr == nil {
				istioConfigList.Templates = models.CastIstioTemplatesCollection(tt)
			}
		}
	}()

	go func() {
		defer wg.Done()
		if criteria.IncludeQuotaSpecs {
			if qs, qsErr = in.k8s.GetQuotaSpecs(criteria.Namespace); qsErr == nil {
				(&istioConfigList.QuotaSpecs).Parse(qs)
			}
		}
	}()

	go func() {
		defer wg.Done()
		if criteria.IncludeQuotaSpecBindings {
			if qb, qbErr = in.k8s.GetQuotaSpecBindings(criteria.Namespace); qbErr == nil {
				(&istioConfigList.QuotaSpecBindings).Parse(qb)
			}
		}
	}()

	go func() {
		defer wg.Done()
		if criteria.IncludePolicies {
			if pc, pcErr = in.k8s.GetPolicies(criteria.Namespace); pcErr == nil {
				(&istioConfigList.Policies).Parse(pc)
			}
		}
	}()

	go func() {
		defer wg.Done()
		// MeshPolicies are not namespaced. They will be only listed for the namespace
		// where istio is deployed.
		if criteria.IncludeMeshPolicies && criteria.Namespace == config.Get().IstioNamespace {
			if mp, mpErr = in.k8s.GetMeshPolicies(criteria.Namespace); mpErr == nil {
				(&istioConfigList.MeshPolicies).Parse(mp)
			}
		}
	}()

	go func() {
		defer wg.Done()
		if criteria.IncludeClusterRbacConfigs && criteria.Namespace == config.Get().IstioNamespace {
			if rc, rcErr = in.k8s.GetClusterRbacConfigs(criteria.Namespace); rcErr == nil {
				(&istioConfigList.ClusterRbacConfigs).Parse(rc)
			}
		}
	}()

	go func() {
		defer wg.Done()
		if criteria.IncludeServiceRoles {
			if sr, srErr = in.k8s.GetServiceRoles(criteria.Namespace); srErr == nil {
				(&istioConfigList.ServiceRoles).Parse(sr)
			}
		}
	}()

	go func() {
		defer wg.Done()
		if criteria.IncludeServiceRoleBindings {
			if srb, srbErr = in.k8s.GetServiceRoleBindings(criteria.Namespace); srbErr == nil {
				(&istioConfigList.ServiceRoleBindings).Parse(srb)
			}
		}
	}()

	wg.Wait()

	for _, genErr := range []error{ggErr, vsErr, drErr, seErr, mrErr, qsErr, qbErr, aaErr, ttErr, mpErr, pcErr, rcErr, srErr, srbErr} {
		if genErr != nil {
			err = genErr
			return models.IstioConfigList{}, err
		}
	}

	return istioConfigList, nil
}

// GetIstioConfigDetails returns a specific Istio configuration object.
// It uses following parameters:
// - "namespace": 		namespace where configuration is stored
// - "objectType":		type of the configuration
// - "objectSubtype":	subtype of the configuration, used when objectType == "adapters" or "templates", empty/not used otherwise
// - "object":			name of the configuration
func (in *IstioConfigService) GetIstioConfigDetails(namespace, objectType, objectSubtype, object string) (models.IstioConfigDetails, error) {
	var err error
	promtimer := internalmetrics.GetGoFunctionMetric("business", "IstioConfigService", "GetIstioConfigDetails")
	defer promtimer.ObserveNow(&err)

	istioConfigDetail := models.IstioConfigDetails{}
	istioConfigDetail.Namespace = models.Namespace{Name: namespace}
	istioConfigDetail.ObjectType = objectType
	var gw, vs, dr, se, qs, qb, r, a, t, pc, mp, rc, sr, srb kubernetes.IstioObject
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		canCreate, canUpdate, canDelete := getPermissions(in.k8s, namespace, objectType, objectSubtype)
		istioConfigDetail.Permissions = models.ResourcePermissions{
			Create: canCreate,
			Update: canUpdate,
			Delete: canDelete,
		}
	}()

	switch objectType {
	case Gateways:
		if gw, err = in.k8s.GetGateway(namespace, object); err == nil {
			istioConfigDetail.Gateway = &models.Gateway{}
			istioConfigDetail.Gateway.Parse(gw)
		}
	case VirtualServices:
		if vs, err = in.k8s.GetVirtualService(namespace, object); err == nil {
			istioConfigDetail.VirtualService = &models.VirtualService{}
			istioConfigDetail.VirtualService.Parse(vs)
		}
	case DestinationRules:
		if dr, err = in.k8s.GetDestinationRule(namespace, object); err == nil {
			istioConfigDetail.DestinationRule = &models.DestinationRule{}
			istioConfigDetail.DestinationRule.Parse(dr)
		}
	case ServiceEntries:
		if se, err = in.k8s.GetServiceEntry(namespace, object); err == nil {
			istioConfigDetail.ServiceEntry = &models.ServiceEntry{}
			istioConfigDetail.ServiceEntry.Parse(se)
		}
	case Rules:
		if r, err = in.k8s.GetIstioRule(namespace, object); err == nil {
			istioRule := models.CastIstioRule(r)
			istioConfigDetail.Rule = &istioRule
		}
	case Adapters:
		if a, err = in.k8s.GetAdapter(namespace, objectSubtype, object); err == nil {
			adapter := models.CastIstioAdapter(a)
			istioConfigDetail.Adapter = &adapter
		}
	case Templates:
		if t, err = in.k8s.GetTemplate(namespace, objectSubtype, object); err == nil {
			template := models.CastIstioTemplate(t)
			istioConfigDetail.Template = &template
		}
	case QuotaSpecs:
		if qs, err = in.k8s.GetQuotaSpec(namespace, object); err == nil {
			istioConfigDetail.QuotaSpec = &models.QuotaSpec{}
			istioConfigDetail.QuotaSpec.Parse(qs)
		}
	case QuotaSpecBindings:
		if qb, err = in.k8s.GetQuotaSpecBinding(namespace, object); err == nil {
			istioConfigDetail.QuotaSpecBinding = &models.QuotaSpecBinding{}
			istioConfigDetail.QuotaSpecBinding.Parse(qb)
		}
	case Policies:
		if pc, err = in.k8s.GetPolicy(namespace, object); err == nil {
			istioConfigDetail.Policy = &models.Policy{}
			istioConfigDetail.Policy.Parse(pc)
		}
	case MeshPolicies:
		if mp, err = in.k8s.GetMeshPolicy(namespace, object); err == nil {
			istioConfigDetail.MeshPolicy = &models.MeshPolicy{}
			istioConfigDetail.MeshPolicy.Parse(mp)
		}
	case ClusterRbacConfigs:
		if rc, err = in.k8s.GetClusterRbacConfig(namespace, object); err == nil {
			istioConfigDetail.ClusterRbacConfig = &models.ClusterRbacConfig{}
			istioConfigDetail.ClusterRbacConfig.Parse(rc)
		}
	case ServiceRoles:
		if sr, err = in.k8s.GetServiceRole(namespace, object); err == nil {
			istioConfigDetail.ServiceRole = &models.ServiceRole{}
			istioConfigDetail.ServiceRole.Parse(sr)
		}
	case ServiceRoleBindings:
		if srb, err = in.k8s.GetServiceRoleBinding(namespace, object); err == nil {
			istioConfigDetail.ServiceRoleBinding = &models.ServiceRoleBinding{}
			istioConfigDetail.ServiceRoleBinding.Parse(srb)
		}
	default:
		err = fmt.Errorf("Object type not found: %v", objectType)
	}

	wg.Wait()

	return istioConfigDetail, err
}

// GetIstioAPI provides the Kubernetes API that manages this Istio resource type
// or empty string if it's not managed
func GetIstioAPI(resourceType string) string {
	return resourceTypesToAPI[resourceType]
}

// ParseJsonForCreate checks if a json is well formed according resourceType/subresourceType.
// It returns a json validated to be used in the Create operation, or an error to report in the handler layer.
func (in *IstioConfigService) ParseJsonForCreate(resourceType, subresourceType string, body []byte) (string, error) {
	var err error
	istioConfigDetail := models.IstioConfigDetails{}
	apiVersion := apiToVersion[resourceTypesToAPI[resourceType]]
	var kind string
	var marshalled string
	if resourceType == Adapters || resourceType == Templates {
		kind = kubernetes.PluralType[subresourceType]
	} else {
		kind = kubernetes.PluralType[resourceType]
	}
	switch resourceType {
	case Gateways:
		istioConfigDetail.Gateway = &models.Gateway{}
		err = json.Unmarshal(body, istioConfigDetail.Gateway)
	case VirtualServices:
		istioConfigDetail.VirtualService = &models.VirtualService{}
		err = json.Unmarshal(body, istioConfigDetail.VirtualService)
	case DestinationRules:
		istioConfigDetail.DestinationRule = &models.DestinationRule{}
		err = json.Unmarshal(body, istioConfigDetail.DestinationRule)
	case ServiceEntries:
		istioConfigDetail.ServiceEntry = &models.ServiceEntry{}
		err = json.Unmarshal(body, istioConfigDetail.ServiceEntry)
	case Rules:
		istioConfigDetail.Rule = &models.IstioRule{}
		err = json.Unmarshal(body, istioConfigDetail.Rule)
	case Adapters:
		istioConfigDetail.Adapter = &models.IstioAdapter{}
		err = json.Unmarshal(body, istioConfigDetail.Adapter)
	case Templates:
		istioConfigDetail.Template = &models.IstioTemplate{}
		err = json.Unmarshal(body, istioConfigDetail.Template)
	case QuotaSpecs:
		istioConfigDetail.QuotaSpec = &models.QuotaSpec{}
		err = json.Unmarshal(body, istioConfigDetail.QuotaSpec)
	case QuotaSpecBindings:
		istioConfigDetail.QuotaSpecBinding = &models.QuotaSpecBinding{}
		err = json.Unmarshal(body, istioConfigDetail.QuotaSpecBinding)
	case Policies:
		istioConfigDetail.Policy = &models.Policy{}
		err = json.Unmarshal(body, istioConfigDetail.Policy)
	case MeshPolicies:
		istioConfigDetail.MeshPolicy = &models.MeshPolicy{}
		err = json.Unmarshal(body, istioConfigDetail.MeshPolicy)
	default:
		err = fmt.Errorf("Object type not found: %v", resourceType)
	}
	if err != nil {
		return "", err
	}
	// Append apiVersion and kind
	marshalled = string(body)
	marshalled = strings.TrimSpace(marshalled)
	marshalled = "" +
		"{\n" +
		"\"kind\": \"" + kind + "\",\n" +
		"\"apiVersion\": \"" + apiVersion + "\"," +
		marshalled[1:]

	return marshalled, nil
}

// DeleteIstioConfigDetail deletes the given Istio resource
func (in *IstioConfigService) DeleteIstioConfigDetail(api, namespace, resourceType, resourceSubtype, name string) (err error) {
	promtimer := internalmetrics.GetGoFunctionMetric("business", "IstioConfigService", "DeleteIstioConfigDetail")
	defer promtimer.ObserveNow(&err)

	if resourceType == Adapters || resourceType == Templates {
		err = in.k8s.DeleteIstioObject(api, namespace, resourceSubtype, name)
	} else {
		err = in.k8s.DeleteIstioObject(api, namespace, resourceType, name)
	}
	return err
}

func (in *IstioConfigService) UpdateIstioConfigDetail(api, namespace, resourceType, resourceSubtype, name, jsonPatch string) (models.IstioConfigDetails, error) {
	var err error
	promtimer := internalmetrics.GetGoFunctionMetric("business", "IstioConfigService", "UpdateIstioConfigDetail")
	defer promtimer.ObserveNow(&err)

	return in.modifyIstioConfigDetail(api, namespace, resourceType, resourceSubtype, name, jsonPatch, false)
}

func (in *IstioConfigService) modifyIstioConfigDetail(api, namespace, resourceType, resourceSubtype, name, json string, create bool) (models.IstioConfigDetails, error) {
	var err error
	updatedType := resourceType
	if resourceType == Adapters || resourceType == Templates {
		updatedType = resourceSubtype
	}

	var result kubernetes.IstioObject
	istioConfigDetail := models.IstioConfigDetails{}
	istioConfigDetail.Namespace = models.Namespace{Name: namespace}
	istioConfigDetail.ObjectType = resourceType

	if create {
		// Create new object
		result, err = in.k8s.CreateIstioObject(api, namespace, updatedType, json)
	} else {
		// Update/Path existing object
		result, err = in.k8s.UpdateIstioObject(api, namespace, updatedType, name, json)
	}
	if err != nil {
		return istioConfigDetail, err
	}

	switch resourceType {
	case Gateways:
		istioConfigDetail.Gateway = &models.Gateway{}
		istioConfigDetail.Gateway.Parse(result)
	case VirtualServices:
		istioConfigDetail.VirtualService = &models.VirtualService{}
		istioConfigDetail.VirtualService.Parse(result)
	case DestinationRules:
		istioConfigDetail.DestinationRule = &models.DestinationRule{}
		istioConfigDetail.DestinationRule.Parse(result)
	case ServiceEntries:
		istioConfigDetail.ServiceEntry = &models.ServiceEntry{}
		istioConfigDetail.ServiceEntry.Parse(result)
	case Rules:
		istioRule := models.CastIstioRule(result)
		istioConfigDetail.Rule = &istioRule
	case Adapters:
		adapter := models.CastIstioAdapter(result)
		istioConfigDetail.Adapter = &adapter
	case Templates:
		template := models.CastIstioTemplate(result)
		istioConfigDetail.Template = &template
	case QuotaSpecs:
		istioConfigDetail.QuotaSpec = &models.QuotaSpec{}
		istioConfigDetail.QuotaSpec.Parse(result)
	case QuotaSpecBindings:
		istioConfigDetail.QuotaSpecBinding = &models.QuotaSpecBinding{}
		istioConfigDetail.QuotaSpecBinding.Parse(result)
	case Policies:
		istioConfigDetail.Policy = &models.Policy{}
		istioConfigDetail.Policy.Parse(result)
	case MeshPolicies:
		istioConfigDetail.MeshPolicy = &models.MeshPolicy{}
		istioConfigDetail.MeshPolicy.Parse(result)
	case ClusterRbacConfigs:
		istioConfigDetail.ClusterRbacConfig = &models.ClusterRbacConfig{}
		istioConfigDetail.ClusterRbacConfig.Parse(result)
	case ServiceRoles:
		istioConfigDetail.ServiceRole = &models.ServiceRole{}
		istioConfigDetail.ServiceRole.Parse(result)
	case ServiceRoleBindings:
		istioConfigDetail.ServiceRoleBinding = &models.ServiceRoleBinding{}
		istioConfigDetail.ServiceRoleBinding.Parse(result)
	default:
		err = fmt.Errorf("Object type not found: %v", resourceType)
	}
	return istioConfigDetail, err

}

func (in *IstioConfigService) CreateIstioConfigDetail(api, namespace, resourceType, resourceSubtype string, body []byte) (models.IstioConfigDetails, error) {
	var err error
	promtimer := internalmetrics.GetGoFunctionMetric("business", "IstioConfigService", "CreateIstioConfigDetail")
	defer promtimer.ObserveNow(&err)

	json, err := in.ParseJsonForCreate(resourceType, resourceSubtype, body)
	if err != nil {
		return models.IstioConfigDetails{}, errors2.NewBadRequest(err.Error())
	}
	return in.modifyIstioConfigDetail(api, namespace, resourceType, resourceSubtype, "", json, true)
}

func getPermissions(k8s kubernetes.IstioClientInterface, namespace, objectType, objectSubtype string) (bool, bool, bool) {
	var canCreate, canPatch, canUpdate, canDelete bool
	if api, ok := resourceTypesToAPI[objectType]; ok {
		// objectType will always match the api used in adapters/templates
		// but if objectSubtype is present it should be used as resourceType
		resourceType := objectType
		if objectSubtype != "" {
			resourceType = objectSubtype
		}
		ssars, permErr := k8s.GetSelfSubjectAccessReview(namespace, api, resourceType, []string{"create", "patch", "update", "delete"})
		if permErr == nil {
			for _, ssar := range ssars {
				if ssar.Spec.ResourceAttributes != nil {
					switch ssar.Spec.ResourceAttributes.Verb {
					case "create":
						canCreate = ssar.Status.Allowed
					case "patch":
						canPatch = ssar.Status.Allowed
					case "update":
						canUpdate = ssar.Status.Allowed
					case "delete":
						canDelete = ssar.Status.Allowed
					}
				}
			}
		} else {
			log.Errorf("Error getting permissions [namespace: %s, api: %s, resourceType: %s]: %v", namespace, api, resourceType, permErr)
		}
	}
	return canCreate, (canUpdate || canPatch), canDelete
}

func (in *IstioConfigService) MeshWidemTLSStatus(namespaces []string) (string, error) {
	mpp, mpErr := in.hasMeshPolicyEnabled(namespaces)
	if mpErr != nil {
		return "", mpErr
	}

	drp, drErr := in.hasDestinationRuleEnabled(namespaces)
	if drErr != nil {
		return "", drErr
	}

	if drp && mpp {
		return MeshmTLSEnabled, nil
	} else if drp || mpp {
		return MeshmTLSPartiallyEnabled, nil
	}

	return MeshmTLSNotEnabled, nil
}

func (in *IstioConfigService) hasMeshPolicyEnabled(namespaces []string) (bool, error) {
	if len(namespaces) < 1 {
		return false, fmt.Errorf("Can't find MeshPolicies without a namespace")
	}

	// MeshPolicies are not namespaced. So any namespace user has access to
	// will work to retrieve all the MeshPolicies.
	mps, err := in.k8s.GetMeshPolicies(namespaces[0])
	if err != nil {
		return false, err
	}

	for _, mp := range mps {

		// It is mandatory to have default as a name
		if meshMeta := mp.GetObjectMeta(); meshMeta.Name != "default" {
			continue
		}

		// It is no globally enabled when has targets
		targets, targetPresent := mp.GetSpec()["targets"]
		specificTarget := targetPresent && len(targets.([]interface{})) > 0
		if specificTarget {
			continue
		}

		// It is globally enabled when a peer has mtls enabled
		peers, peersPresent := mp.GetSpec()["peers"]
		if !peersPresent {
			continue
		}

		for _, peer := range peers.([]interface{}) {
			peerMap := peer.(map[string]interface{})
			if mtls, present := peerMap["mtls"]; present {
				if mtlsMap, ok := mtls.(map[string]interface{}); ok {
					// mTLS enabled in case there is an empty map or mode is STRICT
					if mode, found := mtlsMap["mode"]; !found || mode == "STRICT" {
						return true, nil
					}
				} else {
					// mTLS enabled in case mtls object is empty
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func (in *IstioConfigService) hasDestinationRuleEnabled(namespaces []string) (bool, error) {
	drs, err := in.getAllDestinationRules(namespaces)
	if err != nil {
		return false, err
	}

	mtlsEnabled := false

	for _, dr := range drs {
		// Following the suggested procedure to enable mesh-wide mTLS, host might be '*.local':
		// https://istio.io/docs/tasks/security/authn-policy/#globally-enabling-istio-mutual-tls
		host, hostPresent := dr.GetSpec()["host"]
		if !hostPresent || host != "*.local" {
			continue
		}

		if trafficPolicy, trafficPresent := dr.GetSpec()["trafficPolicy"]; trafficPresent {
			if trafficCasted, ok := trafficPolicy.(map[string]interface{}); ok {
				if tls, found := trafficCasted["tls"]; found {
					if tlsCasted, ok := tls.(map[string]interface{}); ok {
						if mode, found := tlsCasted["mode"]; found {
							if modeCasted, ok := mode.(string); ok {
								if modeCasted == "ISTIO_MUTUAL" {
									mtlsEnabled = true
									break
								}
							}
						}
					}
				}
			}
		}
	}

	return mtlsEnabled, nil
}

func (in *IstioConfigService) getAllDestinationRules(namespaces []string) ([]kubernetes.IstioObject, error) {
	drChan := make(chan []kubernetes.IstioObject, len(namespaces))
	errChan := make(chan error, 1)
	wg := sync.WaitGroup{}

	wg.Add(len(namespaces))

	for _, namespace := range namespaces {
		go func(ns string) {
			defer wg.Done()

			drs, err := in.k8s.GetDestinationRules(ns, "")
			if err != nil {
				errChan <- err
				return
			}

			drChan <- drs
		}(namespace)
	}

	wg.Wait()
	close(errChan)
	close(drChan)

	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	allDestinationRules := make([]kubernetes.IstioObject, 0)
	for drs := range drChan {
		allDestinationRules = append(allDestinationRules, drs...)
	}

	return allDestinationRules, nil
}
