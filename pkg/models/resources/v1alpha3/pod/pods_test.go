/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package pod

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

func TestCompareByRestartCount(t *testing.T) {
	now := time.Now()
	podLow := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "pod-low",
			Namespace:         "default",
			CreationTimestamp: metav1.NewTime(now.Add(-2 * time.Hour)),
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "c1", RestartCount: 1},
				{Name: "c2", RestartCount: 0},
			},
		},
	}
	podHigh := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "pod-high",
			Namespace:         "default",
			CreationTimestamp: metav1.NewTime(now.Add(-1 * time.Hour)),
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "c1", RestartCount: 2},
				{Name: "c2", RestartCount: 1},
			},
		},
	}

	g := &podsGetter{}
	if !g.compare(podHigh, podLow, query.FieldRestartCount) {
		t.Fatalf("expected podHigh to be greater than podLow by restartCount")
	}
	if g.compare(podLow, podHigh, query.FieldRestartCount) {
		t.Fatalf("did not expect podLow to be greater than podHigh by restartCount")
	}
}

func TestListSortByRestartCount(t *testing.T) {
	now := time.Now()
	podA := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "pod-a",
			Namespace:         "ns",
			CreationTimestamp: metav1.NewTime(now.Add(-3 * time.Hour)),
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "c1", RestartCount: 3},
			},
		},
	}
	podB := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "pod-b",
			Namespace:         "ns",
			CreationTimestamp: metav1.NewTime(now.Add(-2 * time.Hour)),
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "c1", RestartCount: 1},
			},
		},
	}
	podC := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "pod-c",
			Namespace:         "ns",
			CreationTimestamp: metav1.NewTime(now.Add(-1 * time.Hour)),
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "c1", RestartCount: 3},
			},
		},
	}

	objects := []runtime.Object{podA, podB, podC}

	q := query.New()
	q.SortBy = query.FieldRestartCount
	q.Ascending = false

	g := &podsGetter{}
	result := v1alpha3.DefaultList(objects, q, g.compare, g.filter)
	if result.TotalItems != 3 {
		t.Fatalf("expected 3 items, got %d", result.TotalItems)
	}

	got := result.Items
	// Expect highest restart first: podA and podC both 3; tie-breaker by creationTimestamp (newer first)
	// DefaultObjectMetaCompare uses creationTimestamp desc when equal sort field not specified.
	// In our compare, on tie we fallback to CreationTimeStamp descending.
	if got[0].(*corev1.Pod).Name != "pod-c" {
		t.Fatalf("expected first item to be pod-c, got %s", got[0].(*corev1.Pod).Name)
	}
	if got[1].(*corev1.Pod).Name != "pod-a" {
		t.Fatalf("expected second item to be pod-a, got %s", got[1].(*corev1.Pod).Name)
	}
	if got[2].(*corev1.Pod).Name != "pod-b" {
		t.Fatalf("expected third item to be pod-b, got %s", got[2].(*corev1.Pod).Name)
	}
}


