/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
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
