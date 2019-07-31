package testing

import (
	"strings"

	"k8s.io/apimachinery/pkg/util/yaml"
)

func StringToObject(data string, obj interface{}) error {
	reader := strings.NewReader(data)
	return yaml.NewYAMLOrJSONDecoder(reader, 10).Decode(obj)
}
