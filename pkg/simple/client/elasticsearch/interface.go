package esclient

import (
	"context"
	"encoding/json"
	"fmt"
	v5 "kubesphere.io/kubesphere/pkg/simple/client/elasticsearch/versions/v5"
	v6 "kubesphere.io/kubesphere/pkg/simple/client/elasticsearch/versions/v6"
	v7 "kubesphere.io/kubesphere/pkg/simple/client/elasticsearch/versions/v7"
	"strings"
)

const (
	ElasticV5 = "5"
	ElasticV6 = "6"
	ElasticV7 = "7"
)

type Client interface {
	// Perform Search API
	Search(body []byte) ([]byte, error)
	GetTotalHitCount(v interface{}) int64
}

func NewForConfig(cfg *Config) Client {
	address := fmt.Sprintf("http://%s:%s", cfg.Host, cfg.Port)
	index := cfg.Index
	switch cfg.VersionMajor {
	case ElasticV5:
		return v5.New(address, index)
	case ElasticV6:
		return v6.New(address, index)
	case ElasticV7:
		return v7.New(address, index)
	default:
		return nil
	}
}

func detectVersionMajor(cfg *Config) error {

	// Info APIs are backward compatible with versions of v5.x, v6.x and v7.x
	address := fmt.Sprintf("http://%s:%s", cfg.Host, cfg.Port)
	es := v6.New(address, "")
	res, err := es.Client.Info(
		es.Client.Info.WithContext(context.Background()),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var b map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&b); err != nil {
		return err
	}
	if res.IsError() {
		// Print the response status and error information.
		e, _ := b["error"].(map[string]interface{})
		return fmt.Errorf("[%s] %s: %s", res.Status(), e["type"], e["reason"])
	}

	// get the major version
	version, _ := b["version"].(map[string]interface{})
	number, _ := version["number"].(string)
	if number == "" {
		return fmt.Errorf("failed to detect elastic version number")
	}

	cfg.VersionMajor = strings.Split(number, ".")[0]
	return nil
}
