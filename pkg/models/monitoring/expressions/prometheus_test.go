package expressions

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestLabelReplace(t *testing.T) {
	tests := []struct {
		expr        string
		expected    string
		expectedErr bool
	}{
		{
			expr:        "up",
			expected:    `up{namespace="default"}`,
			expectedErr: false,
		},
		{
			expr:        `up{namespace="random"}`,
			expected:    `up{namespace="default"}`,
			expectedErr: false,
		},
		{
			expr:        `up{namespace="random"} + up{job="test"}`,
			expected:    `up{namespace="default"} + up{job="test",namespace="default"}`,
			expectedErr: false,
		},
		{
			expr:        `@@@@`,
			expectedErr: true,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			result, err := labelReplace(tt.expr, "default")
			if err != nil {
				if !tt.expectedErr {
					t.Fatal(err)
				}
				return
			}

			if diff := cmp.Diff(result, tt.expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", tt.expected, diff)
			}
		})
	}
}
