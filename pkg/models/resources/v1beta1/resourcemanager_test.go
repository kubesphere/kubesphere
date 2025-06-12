package v1beta1

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	"kubesphere.io/kubesphere/pkg/apiserver/query"
)

func TestContains_MatchesFieldSelector(t *testing.T) {
	object := &runtime.Unknown{
		Raw: []byte(`{"metadata":{"name":"test-object","namespace":"default"}}`),
	}
	queryValue := query.Value("metadata.name=test-object")

	if !contains(object, queryValue) {
		t.Errorf("Expected object to match field selector, but it did not")
	}
}

func TestContains_DoesNotMatchFieldSelector(t *testing.T) {
	object := &runtime.Unknown{
		Raw: []byte(`{"metadata":{"name":"test-object","namespace":"default"}}`),
	}
	queryValue := query.Value("metadata.name=nonexistent-object")

	if contains(object, queryValue) {
		t.Errorf("Expected object not to match field selector, but it did")
	}
}

func TestContains_InvalidFieldSelector(t *testing.T) {
	object := &runtime.Unknown{
		Raw: []byte(`{"metadata":{"name":"test-object","namespace":"default"}}`),
	}
	queryValue := query.Value("invalid-selector")

	if contains(object, queryValue) {
		t.Errorf("Expected object not to match invalid field selector, but it did")
	}
}

func TestContains_CaseInsensitiveMatch(t *testing.T) {
	object := &runtime.Unknown{
		Raw: []byte(`{"metadata":{"name":"Test-Object","namespace":"default"}}`),
	}
	queryValue := query.Value("metadata.name=~test-object")

	if !contains(object, queryValue) {
		t.Errorf("Expected object to match field selector case-insensitively, but it did not")
	}
}

func TestContains_EmptyObject(t *testing.T) {
	object := &runtime.Unknown{
		Raw: []byte(`{}`),
	}
	queryValue := query.Value("metadata.name=test-object")

	if contains(object, queryValue) {
		t.Errorf("Expected empty object not to match field selector, but it did")
	}
}

func TestContains_NestedDataWithSpecialCharacters(t *testing.T) {
	object := &runtime.Unknown{
		Raw: []byte(`{"metadata":{"name":"test-object","annotations":{"example.com/special-key":"special-value"}}}`),
	}
	queryValue := query.Value(`metadata.annotations.example\.com/special-key=special-value`)

	if !contains(object, queryValue) {
		t.Errorf("Expected object to match field selector with nested data and special characters, but it did not")
	}
}
