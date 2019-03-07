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
	"regexp"
	"strconv"

	"golang.org/x/tools/container/intsets"
)

const (
	Paging     = "paging"
	OrderBy    = "orderBy"
	Conditions = "conditions"
	Reserve    = "reserve"
)

func ParsePaging(paging string) (limit, offset int) {
	limit = intsets.MaxInt
	offset = 0
	if groups := regexp.MustCompile(`^limit=(\d+),page=(\d+)$`).FindStringSubmatch(paging); len(groups) == 3 {
		limit, _ = strconv.Atoi(groups[1])
		page, _ := strconv.Atoi(groups[2])
		if page < 0 {
			page = 1
		}
		offset = (page - 1) * limit
	}
	return
}

func ParseReserve(reserve string) bool {
	b, err := strconv.ParseBool(reserve)
	if err != nil {
		return false
	}
	return b
}
