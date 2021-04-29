package v1alpha2

import (
	"strings"
	"testing"

	"kubesphere.io/api/devops/v1alpha3"

	"kubesphere.io/kubesphere/pkg/models/devops"
)

func TestParseNameFilterFromQuery(t *testing.T) {
	table := []struct {
		query           string
		pipeline        *v1alpha3.Pipeline
		expectFilter    devops.PipelineFilter
		expectNamespace string
		message         string
	}{{
		query:           "type:pipeline;organization:jenkins;pipeline:serverjkq4c/*",
		pipeline:        &v1alpha3.Pipeline{},
		expectFilter:    nil,
		expectNamespace: "serverjkq4c",
		message:         "query all pipelines with filter *",
	}, {
		query:    "type:pipeline;organization:jenkins;pipeline:cccc/*abc*",
		pipeline: &v1alpha3.Pipeline{},
		expectFilter: func(pipeline *v1alpha3.Pipeline) bool {
			return strings.Contains(pipeline.Name, "abc")
		},
		expectNamespace: "cccc",
		message:         "query all pipelines with filter abc",
	}, {
		query:           "type:pipeline;organization:jenkins;pipeline:pai-serverjkq4c/*",
		pipeline:        &v1alpha3.Pipeline{},
		expectFilter:    nil,
		expectNamespace: "pai-serverjkq4c",
		message:         "query all pipelines with filter *",
	}, {
		query:           "type:pipeline;organization:jenkins;pipeline:defdef",
		pipeline:        &v1alpha3.Pipeline{},
		expectFilter:    nil,
		expectNamespace: "defdef",
		message:         "query all pipelines with filter *",
	}}

	for i, item := range table {
		filter, ns := parseNameFilterFromQuery(item.query)
		if item.expectFilter == nil {
			if filter != nil {
				t.Fatalf("invalid filter, index: %d, message: %s", i, item.message)
			}
		} else {
			if filter == nil || filter(item.pipeline) != item.expectFilter(item.pipeline) {
				t.Fatalf("invalid filter, index: %d, message: %s", i, item.message)
			}
		}
		if ns != item.expectNamespace {
			t.Fatalf("invalid namespace, index: %d, message: %s", i, item.message)
		}
	}
}
