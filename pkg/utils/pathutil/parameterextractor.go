package pathutil

import (
	"regexp"
	"strings"
)

const (
	clusterResourcePath    = "/kapis/{group}/{version}/{resources}/{name}"
	namespacedResourcePath = "/kapis/{group}/{version}/namespaces/{namespace}/{resources}/{name}"

	clusterRequest    = "cluster request"
	namespacedRequest = "namespaced request"
)

type PathParameterExtractor interface {
	Extract(string) map[string]string
}

func NewKapisPathParameterExtractor() PathParameterExtractor {
	return &kapisPathParameterExtractor{
		clusterResPathParts: getPathParts(clusterResourcePath),
		namespacedResParts:  getPathParts(namespacedResourcePath),
		reg:                 regexp.MustCompile("\\{([a-z]+)\\}"),
	}
}

type kapisPathParameterExtractor struct {
	clusterResPathParts []string
	namespacedResParts  []string
	reg                 *regexp.Regexp
}

func (extractor *kapisPathParameterExtractor) Extract(path string) map[string]string {
	var reqType string
	pathParameter := make(map[string]string, 0)
	reqPathParts := getPathParts(path)
	clusterPartsLen := len(extractor.clusterResPathParts)
	namespacedPartsLen := len(extractor.namespacedResParts)
	reqPartsLen := len(reqPathParts)

	if reqPartsLen <= namespacedPartsLen {
		reqType = namespacedRequest
		if reqPartsLen <= clusterPartsLen {
			reqType = clusterRequest
		}
	} else {
		return pathParameter
	}

	switch reqType {
	case clusterRequest:
		for i, v := range extractor.clusterResPathParts {
			if i < clusterPartsLen-1 {
				if match := extractor.reg.FindStringSubmatch(v); len(match) >= 2 {
					pathParameter[match[1]] = reqPathParts[i]
				}
			}
		}
		if reqPartsLen == clusterPartsLen {
			if match := extractor.reg.FindStringSubmatch(extractor.clusterResPathParts[clusterPartsLen-1]); len(match) >= 2 {
				pathParameter[match[1]] = reqPathParts[clusterPartsLen-1]
			}
		}
	case namespacedRequest:
		for i, v := range extractor.namespacedResParts {
			if i < namespacedPartsLen-1 {
				if match := extractor.reg.FindStringSubmatch(v); len(match) >= 2 {
					pathParameter[match[1]] = reqPathParts[i]
				}
			}

		}
		if reqPartsLen == namespacedPartsLen {
			if match := extractor.reg.FindStringSubmatch(extractor.namespacedResParts[namespacedPartsLen-1]); len(match) >= 2 {
				pathParameter[match[1]] = reqPathParts[namespacedPartsLen-1]
			}
		}
	}

	return pathParameter
}

func getPathParts(path string) []string {
	return strings.Split(strings.Trim(path, "/"), "/")
}
