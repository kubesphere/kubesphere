package s3

import (
	"testing"

	"gotest.tools/assert"
)

func TestCalculateConcurrency(t *testing.T) {
	assert.Equal(t, 5, calculateConcurrency(1*1024*1024))
	assert.Equal(t, 5, calculateConcurrency(5*1024*1024))
	assert.Equal(t, 20, calculateConcurrency(99*1024*1024))
	assert.Equal(t, 128, calculateConcurrency(129*5*1024*1024))
}
