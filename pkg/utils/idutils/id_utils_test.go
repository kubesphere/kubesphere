/*
Copyright 2018 The KubeSphere Authors.

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

package idutils

import (
	"fmt"
	"sort"
	"testing"
)

func TestGetUuid(t *testing.T) {
	fmt.Println(GetUuid(""))
}

func TestGetUuid36(t *testing.T) {
	fmt.Println(GetUuid36(""))
}

func TestGetManyUuid(t *testing.T) {
	var strSlice []string
	for i := 0; i < 10000; i++ {
		testId := GetUuid("")
		strSlice = append(strSlice, testId)
	}
	sort.Strings(strSlice)
}
