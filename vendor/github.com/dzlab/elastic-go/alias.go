package elastic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

const (
	ALIASES = "_aliases"
	ACTIONS = "actions"
)

type Alias struct {
	url  string
	dict Dict
}

/*
 * Return a JSOn representation of the body of this Alias
 */
func (alias *Alias) String() string {
	result, err := json.Marshal(alias.dict)
	if err != nil {
		log.Println(err)
	}
	return string(result)
}

/*
 * Create an alias
 */
func newAlias() *Alias {
	return &Alias{url: "", dict: make(Dict)}
}

func (client *Elasticsearch) Alias() *Alias {
	url := fmt.Sprintf("http://%s/%s", client.Addr, ALIASES)
	return &Alias{url: url, dict: make(Dict)}
}

/*
 * Add an Alias operation (e.g. remove index's alias)
 */
func (alias *Alias) AddAction(operation, index, name string) *Alias {
	if alias.dict[ACTIONS] == nil {
		alias.dict[ACTIONS] = []Dict{}
	}
	action := make(Dict)
	action[operation] = Dict{"index": index, "alias": name}
	alias.dict[ACTIONS] = append(alias.dict[ACTIONS].([]Dict), action)
	return alias
}

/*
 * Submit an Aliases POST operation
 * POST /:index
 */
func (alias *Alias) Post() {
	log.Println("POST", alias.url)
	body := String(alias.dict)
	reader, err := exec("POST", alias.url, bytes.NewReader([]byte(body)))
	if err != nil {
		log.Println(err)
		return
	}
	if data, err := ioutil.ReadAll(reader); err == nil {
		fmt.Println(string(data))
	}

}
