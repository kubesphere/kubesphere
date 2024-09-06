/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v2

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog/v2"

	"helm.sh/helm/v3/pkg/chart"

	"helm.sh/helm/v3/pkg/chart/loader"
	"k8s.io/client-go/dynamic"
	appv2 "kubesphere.io/api/application/v2"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/simple/client/application"
)

const (
	Status = "status"
)

func parseRequest(createRequest application.AppRequest) (appRequest application.AppRequest, err error) {
	if createRequest.AppType == appv2.AppTypeHelm {
		req, err := parseHelmRequest(createRequest)
		return req, err
	}
	_, err = application.ReadYaml(createRequest.Package)

	return createRequest, err
}

func parseHelmRequest(createRequest application.AppRequest) (req application.AppRequest, err error) {
	if createRequest.Package == nil || len(createRequest.Package) == 0 {
		return req, errors.New("package is empty")
	}
	chartPack, err := loader.LoadArchive(bytes.NewReader(createRequest.Package))
	if err != nil {
		return createRequest, err
	}

	crdFiles := chartPack.CRDObjects()
	for _, i := range crdFiles {
		dataList, err := readYaml(i.File.Data)
		if err != nil {
			klog.Errorf("failed to read %s yaml: %v", i.Filename, err)
			return createRequest, err
		}
		for _, d := range dataList {
			decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(d), 1024)
			crd := v1.CustomResourceDefinition{}
			err = decoder.Decode(&crd)
			if err != nil {
				klog.Error(err, "Failed to decode crd file")
				return application.AppRequest{}, err
			}
			var servedVersion *v1.CustomResourceDefinitionVersion
			for _, v := range crd.Spec.Versions {
				if v.Served && v.Storage {
					servedVersion = &v
				}
			}
			if servedVersion == nil {
				klog.Warningf("no served and storage version found in crd %s", crd.Name)
				continue
			}
			ins := appv2.GroupVersionResource{
				Group:    crd.Spec.Group,
				Version:  servedVersion.Name,
				Resource: crd.Spec.Names.Plural,
			}
			createRequest.Resources = append(createRequest.Resources, ins)
		}
	}

	shortName := application.GenerateShortNameMD5Hash(chartPack.Metadata.Name)
	fillEmptyFields(&createRequest, chartPack, shortName)

	return createRequest, nil
}

func readYaml(data []byte) (yamlList [][]byte, err error) {
	// Read yaml file which has multi line yaml and split it into a list of yaml documents.
	r := yaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(data)))
	for {
		d, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			klog.Errorf("failed to read yaml: %v", err)
			return nil, err
		}
		//skip empty yaml or empty ---
		if len(strings.TrimSpace(string(d))) > 3 {
			yamlList = append(yamlList, d)
		}
	}
	return yamlList, nil
}

func fillEmptyFields(createRequest *application.AppRequest, chartPack *chart.Chart, shortName string) {
	if createRequest.AppName == "" {
		createRequest.AppName = application.GetUuid36(shortName + "-")
	}
	if createRequest.OriginalName == "" {
		createRequest.OriginalName = chartPack.Metadata.Name
	}
	if createRequest.AppHome == "" {
		createRequest.AppHome = chartPack.Metadata.Home
	}
	if createRequest.Icon == "" {
		createRequest.Icon = chartPack.Metadata.Icon
	}
	if createRequest.VersionName == "" {
		createRequest.VersionName = chartPack.Metadata.Version
	}

	if createRequest.Maintainers == nil || len(createRequest.Maintainers) == 0 {
		createRequest.Maintainers = application.GetMaintainers(chartPack.Metadata.Maintainers)
	}
	if createRequest.RepoName == "" {
		createRequest.RepoName = appv2.UploadRepoKey
	}
	if createRequest.Description == "" {
		createRequest.Description = chartPack.Metadata.Description
	}
	if createRequest.Abstraction == "" {
		createRequest.Abstraction = chartPack.Metadata.Description
	}
	if createRequest.AliasName == "" {
		createRequest.AliasName = chartPack.Metadata.Name
	}
}

func (h *appHandler) getCluster(clusterName string) (runtimeclient.Client, *dynamic.DynamicClient, *clusterv1alpha1.Cluster, error) {
	klog.Infof("get cluster %s", clusterName)
	runtimeClient, err := h.clusterClient.GetRuntimeClient(clusterName)
	if err != nil {
		return nil, nil, nil, err
	}
	clusterClient, err := h.clusterClient.GetClusterClient(clusterName)
	if err != nil {
		return nil, nil, nil, err
	}
	dynamicClient, err := dynamic.NewForConfig(clusterClient.RestConfig)
	if err != nil {
		return nil, nil, nil, err
	}
	cluster, err := h.clusterClient.Get(clusterName)
	if err != nil {
		return nil, nil, nil, err
	}
	return runtimeClient, dynamicClient, cluster, nil
}
