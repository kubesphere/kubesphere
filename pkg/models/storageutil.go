package models

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type jsonRawMessage []byte

func (m jsonRawMessage) Find(key string) jsonRawMessage {
	var objmap map[string]json.RawMessage
	err := json.Unmarshal(m, &objmap)
	if err != nil {
		glog.Errorf("Resolve JSON Key failed, find key =%s, err=%s",
			key, err)
		return nil
	}
	return jsonRawMessage(objmap[key])
}

func (m jsonRawMessage) ToList() []jsonRawMessage {
	var lists []json.RawMessage
	err := json.Unmarshal(m, &lists)
	if err != nil {
		glog.Errorf("Resolve JSON List failed, err=%s",
			err)
		return nil
	}
	var res []jsonRawMessage
	for _, v := range lists {
		res = append(res, jsonRawMessage(v))
	}
	return res
}

func (m jsonRawMessage) ToString() string {
	res := strings.Replace(string(m[:]), "\"", "", -1)
	return res
}

func GetApiserver(uri string) string {
	url := "http://139.198.6.45:8001" + uri
	response, err := http.Get(url)
	defer response.Body.Close()
	if err != nil {
		fmt.Errorf("%s", err)
		os.Exit(1)
	}
	if response.StatusCode != http.StatusOK {
		glog.Infof("URL=%s, Status=%d", url, response.StatusCode)
		return ""
	}
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}
	return string(contents)
}
