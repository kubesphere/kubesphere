package devops

import (
	"gotest.tools/assert"
	"testing"
)

func TestGetSubmitters(t *testing.T) {
	input := &Input{}
	assert.Equal(t, len(input.GetSubmitters()), 0,
		"errors happen when try to get submitters without any submitters")

	input.Submitter = "a , b, c,d"
	submitters := input.GetSubmitters()
	assert.Equal(t, len(submitters), 4, "get incorrect number of submitters")
	assert.DeepEqual(t, submitters, []string{"a", "b", "c", "d"})
}

func TestApprovable(t *testing.T) {
	input := &Input{}

	assert.Equal(t, input.Approvable(""), true, "should allow anyone to approve it if there's no submitter given")
	assert.Equal(t, input.Approvable("fake"), true, "should allow anyone to approve it if there's no submitter given")

	input.Submitter = "fake"
	assert.Equal(t, input.Approvable(""), false, "should not approve by nobody if there's a particular submitter")
	assert.Equal(t, input.Approvable("rick"), false, "should not approve by who is not the specific one")
	assert.Equal(t, input.Approvable("fake"), true, "should be approvable")

	input.Submitter = "fake, good ,bad"
	assert.Equal(t, input.Approvable("fake"), true, "should be approvable")
	assert.Equal(t, input.Approvable("good"), true, "should be approvable")
	assert.Equal(t, input.Approvable("bad"), true, "should be approvable")
}
