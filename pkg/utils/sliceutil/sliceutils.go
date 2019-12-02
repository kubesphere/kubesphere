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
package sliceutil

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

// StringDiff calculate the difference between two string slices.
// `more` returns elements that are present in the second slice but not present in the first slice.
// `less` returns elements that are not present in the second slice but are present in the first slice.
func StringDiff(slice1 []string, slice2 []string) (more []string, less []string) {
	more = make([]string, 0)
	less = make([]string, 0)
	for i := 0; i < 2; i++ {
		m := make(map[string]bool)
		for _, item := range slice1 {
			m[item] = true
		}

		for _, item := range slice2 {
			if _, ok := m[item]; !ok {
				if i == 0 {
					more = append(more, item)
				} else {
					less = append(less, item)
				}
			}
		}
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}
	return
}
