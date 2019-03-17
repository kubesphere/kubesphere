package gateways

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/models"
)

type MultiMatchChecker struct {
	GatewaysPerNamespace [][]kubernetes.IstioObject
	existingList         []Host
}

const (
	GatewayCheckerType = "gateway"
	wildCardMatch      = "*"
)

type Host struct {
	Port            int
	Hostname        string
	ServerIndex     int
	HostIndex       int
	GatewayRuleName string
}

// Check validates that no two gateways share the same host+port combination
func (m MultiMatchChecker) Check() models.IstioValidations {
	validations := models.IstioValidations{}
	m.existingList = make([]Host, 0)

	for _, nsG := range m.GatewaysPerNamespace {
		for _, g := range nsG {
			gatewayRuleName := g.GetObjectMeta().Name
			if specServers, found := g.GetSpec()["servers"]; found {
				if servers, ok := specServers.([]interface{}); ok {
					for i, def := range servers {
						if serverDef, ok := def.(map[string]interface{}); ok {
							hosts := parsePortAndHostnames(serverDef)
							for hi, host := range hosts {
								host.ServerIndex = i
								host.HostIndex = hi
								host.GatewayRuleName = gatewayRuleName
								duplicate, dhosts := m.findMatch(host)
								if duplicate {
									validations = addError(validations, gatewayRuleName, i, hi)
									for _, dh := range dhosts {
										validations = addError(validations, dh.GatewayRuleName, dh.ServerIndex, dh.HostIndex)
									}
								}
								m.existingList = append(m.existingList, host)
							}
						}
					}
				}
			}
		}
	}
	return validations
}

func addError(validations models.IstioValidations, gatewayRuleName string, serverIndex, hostIndex int) models.IstioValidations {
	key := models.IstioValidationKey{Name: gatewayRuleName, ObjectType: GatewayCheckerType}
	checks := models.Build("gateways.multimatch",
		"spec/servers["+strconv.Itoa(serverIndex)+"]/hosts["+strconv.Itoa(hostIndex)+"]")
	rrValidation := &models.IstioValidation{
		Name:       gatewayRuleName,
		ObjectType: GatewayCheckerType,
		Valid:      true,
		Checks: []*models.IstioCheck{
			&checks,
		},
	}

	if _, exists := validations[key]; !exists {
		validations.MergeValidations(models.IstioValidations{key: rrValidation})
	}
	return validations
}

func parsePortAndHostnames(serverDef map[string]interface{}) []Host {
	var port int
	if portDef, found := serverDef["port"]; found {
		if ports, ok := portDef.(map[string]interface{}); ok {
			if numberDef, found := ports["number"]; found {
				if portNumber, ok := numberDef.(int64); ok {
					port = int(portNumber)
				}
			}
		}
	}

	if hostDef, found := serverDef["hosts"]; found {
		if hostnames, ok := hostDef.([]interface{}); ok {
			hosts := make([]Host, 0, len(hostnames))
			for _, hostinterface := range hostnames {
				if hostname, ok := hostinterface.(string); ok {
					hosts = append(hosts, Host{
						Port:     port,
						Hostname: hostname,
					})
				}
			}
			return hosts
		}
	}
	return nil
}

// findMatch uses a linear search with regexp to check for matching gateway host + port combinations. If this becomes a bottleneck for performance, replace with a graph or trie algorithm.
func (m MultiMatchChecker) findMatch(host Host) (bool, []Host) {
	duplicates := make([]Host, 0)
	for _, h := range m.existingList {
		if h.Port == host.Port {
			// wildcardMatches will always match
			if host.Hostname == wildCardMatch || h.Hostname == wildCardMatch {
				duplicates = append(duplicates, h)
				break
			}

			// Either one could include wildcards, so we need to check both ways and fix "*" -> ".*" for regexp engine
			current := strings.Replace(host.Hostname, "*", ".*", -1)
			previous := strings.Replace(h.Hostname, "*", ".*", -1)

			if regexp.MustCompile(current).MatchString(previous) || regexp.MustCompile(previous).MatchString(current) {
				duplicates = append(duplicates, h)
				break
			}
		}
	}
	return len(duplicates) > 0, duplicates
}
