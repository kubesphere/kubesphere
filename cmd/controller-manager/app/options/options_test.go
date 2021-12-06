package options

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// ref: https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/controller-manager/app/helper_test.go
func TestIsControllerEnabled(t *testing.T) {
	testcases := []struct {
		name string
		controllerName string
		controllerFlags []string
		expected bool
	}{
		{
			name:                         "on by name",
			controllerName:               "bravo",
			controllerFlags:              []string{"alpha", "bravo", "-charlie"},
			expected:                     true,
		},
		{
			name:                         "off by name",
			controllerName:               "charlie",
			controllerFlags:              []string{"alpha", "bravo", "-charlie"},
			expected:                     false,
		},
		{
			name:                         "on by default",
			controllerName:               "alpha",
			controllerFlags:              []string{"*"},
			expected:                     true,
		},
		{
			name:                         "on by star, not off by name",
			controllerName:               "alpha",
			controllerFlags:              []string{"*", "-charlie"},
			expected:                     true,
		},
		{
			name:                         "off by name with star",
			controllerName:               "charlie",
			controllerFlags:              []string{"*", "-charlie"},
			expected:                     false,
		},
		{
			name:                         "off then on",
			controllerName:               "alpha",
			controllerFlags:              []string{"-alpha", "alpha"},
			expected:                     false,
		},
		{
			name:                         "on then off",
			controllerName:               "alpha",
			controllerFlags:              []string{"alpha", "-alpha"},
			expected:                     true,
		},
	}

	for _, tc := range testcases {
		option := NewKubeSphereControllerManagerOptions()
		option.ControllerSelectors = tc.controllerFlags
		actual := option.IsControllerEnabled(tc.controllerName)
		assert.Equal(t, tc.expected, actual, "%v: expected %v, got %v", tc.name, tc.expected, actual)
	}
}