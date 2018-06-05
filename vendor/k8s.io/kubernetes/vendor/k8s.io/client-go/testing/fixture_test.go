/*
Copyright 2015 The Kubernetes Authors.

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

package testing

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/watch"
)

func getArbitraryResource(s schema.GroupVersionResource, name, namespace string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       s.Resource,
			"apiVersion": s.Version,
			"metadata": map[string]interface{}{
				"name":            name,
				"namespace":       namespace,
				"generateName":    "test_generateName",
				"uid":             "test_uid",
				"resourceVersion": "test_resourceVersion",
				"selfLink":        "test_selfLink",
			},
			"data": strconv.Itoa(rand.Int()),
		},
	}
}

func TestWatchCallNonNamespace(t *testing.T) {
	testResource := schema.GroupVersionResource{Group: "", Version: "test_version", Resource: "test_kind"}
	testObj := getArbitraryResource(testResource, "test_name", "test_namespace")
	accessor, err := meta.Accessor(testObj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ns := accessor.GetNamespace()
	scheme := runtime.NewScheme()
	codecs := serializer.NewCodecFactory(scheme)
	o := NewObjectTracker(scheme, codecs.UniversalDecoder())
	watch, err := o.Watch(testResource, ns)
	go func() {
		err := o.Create(testResource, testObj, ns)
		if err != nil {
			t.Errorf("test resource creation failed: %v", err)
		}
	}()
	out := <-watch.ResultChan()
	assert.Equal(t, testObj, out.Object, "watched object mismatch")
}

func TestWatchCallAllNamespace(t *testing.T) {
	testResource := schema.GroupVersionResource{Group: "", Version: "test_version", Resource: "test_kind"}
	testObj := getArbitraryResource(testResource, "test_name", "test_namespace")
	accessor, err := meta.Accessor(testObj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ns := accessor.GetNamespace()
	scheme := runtime.NewScheme()
	codecs := serializer.NewCodecFactory(scheme)
	o := NewObjectTracker(scheme, codecs.UniversalDecoder())
	w, err := o.Watch(testResource, "test_namespace")
	wAll, err := o.Watch(testResource, "")
	go func() {
		err := o.Create(testResource, testObj, ns)
		assert.NoError(t, err, "test resource creation failed")
	}()
	out := <-w.ResultChan()
	outAll := <-wAll.ResultChan()
	assert.Equal(t, watch.Added, out.Type, "watch event mismatch")
	assert.Equal(t, watch.Added, outAll.Type, "watch event mismatch")
	assert.Equal(t, testObj, out.Object, "watched created object mismatch")
	assert.Equal(t, testObj, outAll.Object, "watched created object mismatch")
	go func() {
		err := o.Update(testResource, testObj, ns)
		assert.NoError(t, err, "test resource updating failed")
	}()
	out = <-w.ResultChan()
	outAll = <-wAll.ResultChan()
	assert.Equal(t, watch.Modified, out.Type, "watch event mismatch")
	assert.Equal(t, watch.Modified, outAll.Type, "watch event mismatch")
	assert.Equal(t, testObj, out.Object, "watched updated object mismatch")
	assert.Equal(t, testObj, outAll.Object, "watched updated object mismatch")
	go func() {
		err := o.Delete(testResource, "test_namespace", "test_name")
		assert.NoError(t, err, "test resource deletion failed")
	}()
	out = <-w.ResultChan()
	outAll = <-wAll.ResultChan()
	assert.Equal(t, watch.Deleted, out.Type, "watch event mismatch")
	assert.Equal(t, watch.Deleted, outAll.Type, "watch event mismatch")
	assert.Equal(t, testObj, out.Object, "watched deleted object mismatch")
	assert.Equal(t, testObj, outAll.Object, "watched deleted object mismatch")
}

func TestWatchCallMultipleInvocation(t *testing.T) {
	cases := []struct {
		name string
		op   watch.EventType
	}{
		{
			"foo",
			watch.Added,
		},
		{
			"bar",
			watch.Added,
		},
		{
			"bar",
			watch.Modified,
		},
		{
			"foo",
			watch.Deleted,
		},
		{
			"bar",
			watch.Deleted,
		},
	}

	scheme := runtime.NewScheme()
	codecs := serializer.NewCodecFactory(scheme)
	testResource := schema.GroupVersionResource{Group: "", Version: "test_version", Resource: "test_kind"}

	o := NewObjectTracker(scheme, codecs.UniversalDecoder())
	watchNamespaces := []string{
		"",
		"",
		"test_namespace",
		"test_namespace",
	}
	var wg sync.WaitGroup
	wg.Add(len(watchNamespaces))
	for idx, watchNamespace := range watchNamespaces {
		i := idx
		w, err := o.Watch(testResource, watchNamespace)
		go func() {
			assert.NoError(t, err, "watch invocation failed")
			for _, c := range cases {
				fmt.Printf("%#v %#v\n", c, i)
				event := <-w.ResultChan()
				accessor, err := meta.Accessor(event.Object)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				assert.Equal(t, c.op, event.Type, "watch event mismatched")
				assert.Equal(t, c.name, accessor.GetName(), "watched object mismatch")
			}
			wg.Done()
		}()
	}
	for _, c := range cases {
		switch c.op {
		case watch.Added:
			obj := getArbitraryResource(testResource, c.name, "test_namespace")
			o.Create(testResource, obj, "test_namespace")
		case watch.Modified:
			obj := getArbitraryResource(testResource, c.name, "test_namespace")
			o.Update(testResource, obj, "test_namespace")
		case watch.Deleted:
			o.Delete(testResource, "test_namespace", c.name)
		}
	}
	wg.Wait()
}

func TestWatchAddAfterStop(t *testing.T) {
	testResource := schema.GroupVersionResource{Group: "", Version: "test_version", Resource: "test_kind"}
	testObj := getArbitraryResource(testResource, "test_name", "test_namespace")
	accessor, err := meta.Accessor(testObj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ns := accessor.GetNamespace()
	scheme := runtime.NewScheme()
	codecs := serializer.NewCodecFactory(scheme)
	o := NewObjectTracker(scheme, codecs.UniversalDecoder())
	watch, err := o.Watch(testResource, ns)
	if err != nil {
		t.Errorf("watch creation failed: %v", err)
	}

	// When the watch is stopped it should ignore later events without panicking.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Watch panicked when it should have ignored create after stop: %v", r)
		}
	}()

	watch.Stop()
	err = o.Create(testResource, testObj, ns)
	if err != nil {
		t.Errorf("test resource creation failed: %v", err)
	}
}
