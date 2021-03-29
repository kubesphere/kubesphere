package jenkins

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseJenkinsQuery(t *testing.T) {
	table := []testData{
		{
			param: "start=0&limit=10&branch=master",
			expected: url.Values{
				"start":  []string{"0"},
				"limit":  []string{"10"},
				"branch": []string{"master"},
			}, err: false,
		},
		{
			param: "branch=master", expected: url.Values{
				"branch": []string{"master"},
			}, err: false,
		},
		{
			param: "&branch=master", expected: url.Values{
				"branch": []string{"master"},
			}, err: false,
		},
		{
			param: "branch=master&", expected: url.Values{
				"branch": []string{"master"},
			}, err: false,
		},
		{
			param: "branch=%gg", expected: url.Values{}, err: true,
		},
		{
			param: "%gg=fake", expected: url.Values{}, err: true,
		},
	}

	for index, item := range table {
		result, err := ParseJenkinsQuery(item.param)
		if item.err {
			assert.NotNil(t, err, "index: [%d], unexpected error happen %v", index, err)
		} else {
			assert.Nil(t, err, "index: [%d], unexpected error happen %v", index, err)
		}
		assert.Equal(t, item.expected, result, "index: [%d], result do not match with the expect value", index)
	}
}

type testData struct {
	param    string
	expected interface{}
	err      bool
}
