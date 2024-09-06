/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha3

import "testing"

func TestLabelMatch(t *testing.T) {
	tests := []struct {
		labels       map[string]string
		filter       string
		expectResult bool
	}{
		{
			labels: map[string]string{
				"kubesphere.io/workspace": "kubesphere-system",
			},
			filter:       "kubesphere.io/workspace",
			expectResult: true,
		},
		{
			labels: map[string]string{
				"kubesphere.io/creator": "system",
			},
			filter:       "kubesphere.io/workspace",
			expectResult: false,
		},
		{
			labels: map[string]string{
				"kubesphere.io/workspace": "kubesphere-system",
			},
			filter:       "kubesphere.io/workspace=",
			expectResult: false,
		},
		{
			labels: map[string]string{
				"kubesphere.io/workspace": "kubesphere-system",
			},
			filter:       "kubesphere.io/workspace!=",
			expectResult: true,
		},
		{
			labels: map[string]string{
				"kubesphere.io/workspace": "kubesphere-system",
			},
			filter:       "kubesphere.io/workspace!=kubesphere-system",
			expectResult: false,
		},
		{
			labels: map[string]string{
				"kubesphere.io/workspace": "kubesphere-system",
			},
			filter:       "kubesphere.io/workspace=kubesphere-system",
			expectResult: true,
		},
		{
			labels: map[string]string{
				"kubesphere.io/workspace": "kubesphere-system",
			},
			filter:       "kubesphere.io/workspace=system",
			expectResult: false,
		},
	}
	for i, test := range tests {
		result := labelMatch(test.labels, test.filter)
		if result != test.expectResult {
			t.Errorf("case %d, got %#v, expected %#v", i, result, test.expectResult)
		}
	}
}
