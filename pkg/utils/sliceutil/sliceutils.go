/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package sliceutil

import "sort"

func RemoveString(slice []string, remove func(item string) bool) []string {
	for i := 0; i < len(slice); i++ {
		if remove(slice[i]) {
			slice = append(slice[:i], slice[i+1:]...)
			i--
		}
	}
	return slice
}

func HasString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func Equal(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	sort.Strings(slice1)
	sort.Strings(slice2)

	for i, s := range slice1 {
		if s != slice2[i] {
			return false
		}
	}
	return true
}
