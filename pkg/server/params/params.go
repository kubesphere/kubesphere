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
)

func ParsePaging(req *restful.Request) (limit, offset int) {
	paging := req.QueryParameter(PagingParam)
	limit = 10
	offset = 0
	if groups := regexp.MustCompile(`^limit=(-?\d+),page=(\d+)$`).FindStringSubmatch(paging); len(groups) == 3 {
		limit, _ = strconv.Atoi(groups[1])
		page, _ := strconv.Atoi(groups[2])
		offset = (page - 1) * limit
	}
	return
}

var (
	invalidKeyRegex = regexp.MustCompile(`[\s(){}\[\]]`)
)

// Ref: stdlib url.ParseQuery
func parseConditions(conditionsStr string) (*Conditions, error) {
	// string likes: key1=value1,key2~value2,key3=
	// exact query: key=value, if value is empty means label value must be ""
	// fuzzy query: key~value, if value is empty means label value is "" or label key not exist
	var conditions = &Conditions{Match: make(map[string]string, 0), Fuzzy: make(map[string]string, 0)}

	for conditionsStr != "" {
		key := conditionsStr
		if i := strings.Index(key, ","); i >= 0 {
			key, conditionsStr = key[:i], key[i+1:]
		} else {
			conditionsStr = ""
		}
		if key == "" {
			continue
		}
		value := ""
		var isFuzzy = false
		if i := strings.IndexAny(key, "~="); i >= 0 {
			if key[i] == '~' {
				isFuzzy = true
			}
			key, value = key[:i], key[i+1:]
		}
		if invalidKeyRegex.MatchString(key) {
			return nil, fmt.Errorf("invalid conditions")
		}
		if isFuzzy {
			conditions.Fuzzy[key] = value
		} else {
			conditions.Match[key] = value
		}
	}
	return conditions, nil
}

func ParseConditions(req *restful.Request) (*Conditions, error) {
	return parseConditions(req.QueryParameter(ConditionsParam))
}

type Conditions struct {
	Match map[string]string
	Fuzzy map[string]string
}

func GetBoolValueWithDefault(req *restful.Request, name string, dv bool) bool {
	reverse := req.QueryParameter(name)
	if v, err := strconv.ParseBool(reverse); err == nil {
		return v
	}
	return dv
}

func GetStringValueWithDefault(req *restful.Request, name string, dv string) string {
	v := req.QueryParameter(name)
	if v == "" {
		v = dv
	}
	return v
}
