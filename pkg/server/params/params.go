/*

 Copyright 2019 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.

*/
package params

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"regexp"
	"strconv"
	"strings"
)

const (
	PagingParam     = "paging"
	OrderByParam    = "orderBy"
	ConditionsParam = "conditions"
	ReverseParam    = "reverse"
	NameParam       = "name"
)

func ParsePaging(paging string) (limit, offset int) {

	limit = 10
	offset = 0
	if groups := regexp.MustCompile(`^limit=(-?\d+),page=(\d+)$`).FindStringSubmatch(paging); len(groups) == 3 {
		limit, _ = strconv.Atoi(groups[1])
		page, _ := strconv.Atoi(groups[2])
		offset = (page - 1) * limit
	}
	return
}

func ParseConditions(conditionsStr string) (*Conditions, error) {

	conditions := &Conditions{Match: make(map[string]string, 0), Fuzzy: make(map[string]string, 0)}

	if conditionsStr == "" {
		return conditions, nil
	}

	// ?conditions=key1=value1,key2~value2,key3=
	for _, item := range strings.Split(conditionsStr, ",") {
		// exact query: key=value, if value is empty means label value must be ""
		// fuzzy query: key~value, if value is empty means label value is "" or label key not exist
		if groups := regexp.MustCompile(`(\S+)([=~])(\S+)?`).FindStringSubmatch(item); len(groups) >= 3 {
			value := ""

			if len(groups) > 3 {
				value = groups[3]
			}

			if groups[2] == "=" {
				conditions.Match[groups[1]] = value
			} else {
				conditions.Fuzzy[groups[1]] = value
			}
		} else {
			return nil, fmt.Errorf("invalid conditions")
		}
	}
	return conditions, nil
}

func ParseReverse(req *restful.Request) bool {
	reverse := req.QueryParameter(ReverseParam)
	b, err := strconv.ParseBool(reverse)
	if err != nil {
		return false
	}
	return b
}

type Conditions struct {
	Match map[string]string
	Fuzzy map[string]string
}

func GetStringValueWithDefault(req *restful.Request, name string, dv string) string {
	v := req.QueryParameter(name)
	if v == "" {
		v = dv
	}
	return v
}
