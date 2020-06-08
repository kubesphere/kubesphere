package auditing

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"k8s.io/klog"
	"os"
	"regexp"
)

const (
	ConfigFile = "/etc/kubesphere/auditing/config"
)

type Config struct {
	// URI is the request URI as sent by the client to a server.
	URI string `yaml:"uri,omitempty"`
	// Http request method.
	Method string `yaml:"method,omitempty"`
	// Verb is the operator of this request, such as login.
	Verb string `yaml:"verb,omitempty"`
	// Resource of this request.
	// For non-resource requests, this is nil.
	Resource string `yaml:"resource,omitempty"`
	// Subresource of this request.
	// For non-resource requests, this is nil.
	Subresource string `yaml:"subresource,omitempty"`
	// Name of the resource, it will be Splicing with commas.
	// If the path is from uri, the format like this: parameter.devops .
	// If the path is from body, the format like this: body.metadata.name
	NamPath []string `yaml:"name,omitempty"`
	// URI regular expression
	regular *regexp.Regexp
}

func loadAuditingConfig() []Config {
	var configs []Config

	f, err := os.Open(ConfigFile)
	if err != nil {
		klog.Errorf("load auditing config file error %s", err)
		return configs
	}

	bs, err := ioutil.ReadAll(f)
	if err != nil {
		klog.Errorf("read auditing config file error %s", err)
		return configs
	}

	err = yaml.Unmarshal(bs, &configs)
	if err != nil {
		klog.Errorf("auditing config file error %s", err)
		return configs
	}

	for index := range configs {
		uri := configs[index].URI

		compile, err := regexp.Compile("{(.*?)}")
		if err != nil {
			klog.Error(err)
			continue
		}
		s := compile.ReplaceAllString(uri, "(.*)")
		compile, err = regexp.Compile(s)
		if err != nil {
			klog.Error(err)
			continue
		}

		configs[index].regular = compile
	}

	return configs
}
