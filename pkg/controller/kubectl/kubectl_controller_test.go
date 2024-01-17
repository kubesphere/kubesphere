package kubectl

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/informers"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/util/workqueue"
)

func Test_NewController(t *testing.T) {
	cs := fakeclientset.NewSimpleClientset(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fake-node1",
			Labels: map[string]string{
				"fake.label.kubernetes.io": "fake",
			},
		},
	})
	informerFactory := informers.NewSharedInformerFactory(cs, time.Minute*10)
	c := NewController(cs, informerFactory.Core().V1().Pods())
	if c.podSynced == nil || c.getPod == nil || c.updatePod == nil || c.deletePod == nil || c.workqueue == nil || c.sync == nil {
		t.Error("NewController failed")
	}
}

func Test_Controller_Start(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*1)
	defer cancel()
	uid := uuid.New().String()
	c := &Controller{
		podSynced: func() bool { return true },
		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "kubectl"),
		sync: func(key string) error {
			if key != uid {
				t.Errorf("Test Controller Start failed: %s!=%s", key, uid)
			}
			cancel()
			return nil
		},
	}
	c.workqueue.Add(uid)
	if err := c.Start(ctx); err != nil {
		t.Errorf("Test Controller Start failed: %s", err)
	}
}

func Test_Controller_enqueuePod(t *testing.T) {
	c := Controller{workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "kubectl")}
	// Test invalid object
	c.enqueuePod("test")
	if c.workqueue.Len() != 0 {
		t.Errorf("enqueuePod failed")
	}
	c.enqueuePod(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-name",
			Namespace: "test-ns",
		},
	})
	if c.workqueue.Len() != 1 {
		t.Errorf("enqueuePod failed:len != %d", c.workqueue.Len())
	} else if obj, shutdown := c.workqueue.Get(); shutdown || !reflect.DeepEqual(obj, "test-ns/test-name") {
		t.Errorf("enqueuePod failed:%v,%v", shutdown, obj)
	}
}

func Test_Controller_processNextWorkItem(t *testing.T) {
	c := &Controller{
		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "kubectl"),
	}

	// 1. Test work queue is closed.
	c.workqueue.ShutDown()
	if c.processNextWorkItem() {
		t.Errorf("processNextWorkItem failed")
	}

	// 2. Test invalid key
	c.workqueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "kubectl")
	c.sync = func(key string) error {
		t.Errorf("Test_Controller_processNextWorkItem: %s", key)
		return nil
	}
	c.workqueue.Add(2)
	if !c.processNextWorkItem() {
		t.Errorf("processNextWorkItem failed")
	}

	// 3. Test sync failed.
	c.sync = func(key string) error {
		return fmt.Errorf("test error")
	}
	uid := uuid.New().String()
	c.workqueue.Add(uid)
	if !c.processNextWorkItem() {
		t.Errorf("processNextWorkItem failed")
	} else if key, shutdown := c.workqueue.Get(); shutdown || !reflect.DeepEqual(key, uid) {
		t.Errorf("processNextWorkItem failed: %v, %v", key, shutdown)
	}

	// 3. Test sync succeeded.
	c.sync = func(key string) error {
		return nil
	}
	uid = uuid.New().String()
	c.workqueue.Add(uid)
	if !c.processNextWorkItem() {
		t.Errorf("processNextWorkItem failed")
	}
	go func() {
		time.Sleep(time.Second)
		c.workqueue.ShutDown()
	}()
	if key, shutdown := c.workqueue.Get(); !shutdown || reflect.DeepEqual(key, uid) {
		t.Errorf("processNextWorkItem failed: %v, %v", key, shutdown)
	}
}

func Test_Controller_reconcile(t *testing.T) {
	c := &Controller{}

	// 1. test invalid key
	err := c.reconcile("a/b/c")
	if err == nil {
		t.Errorf("reconcile() failed")
	}

	// 2. test not found.
	c.getPod = func(namespace, name string) (*v1.Pod, error) {
		return nil, errors.NewNotFound(schema.GroupResource{Group: "v1", Resource: "pod"}, "test not found")
	}
	err = c.reconcile("test")
	if err != nil {
		t.Errorf("reconcile() failed: %v", err)
	}

	// 3. test get pod failed.
	werr := fmt.Errorf(uuid.New().String())
	c.getPod = func(namespace, name string) (*v1.Pod, error) {
		return nil, werr
	}
	err = c.reconcile("test")
	if err != werr {
		t.Errorf("reconcile() failed: %v", err)
	}

	// 4. test add finalizer failed.
	name := uuid.New().String()
	c.getPod = func(namespace, name string) (*v1.Pod, error) {
		return &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}, nil
	}
	c.updatePod = func(pod *v1.Pod) (*v1.Pod, error) {
		if pod.Name != name {
			return nil, fmt.Errorf("invalid pod %s!=%s", name, pod.Name)
		}
		if !sets.NewString(pod.ObjectMeta.Finalizers...).Has(Finalizer) {
			return nil, fmt.Errorf("without finalizer %v", pod.ObjectMeta.Finalizers)
		}

		return nil, fmt.Errorf("test add finalizer failed")
	}
	err = c.reconcile(name)
	if err == nil || err.Error() != "test add finalizer failed" {
		t.Errorf("reconcile() failed: %v", err)
	}

	// 5. test add finalizer succeeded.
	c.updatePod = func(pod *v1.Pod) (*v1.Pod, error) {
		return pod.DeepCopy(), nil
	}
	err = c.reconcile(name)
	if err != nil {
		t.Errorf("reconcile() failed: %v", err)
	}

	// 6. test remove finalizer failed.
	name = uuid.New().String()
	c.getPod = func(namespace, name string) (*v1.Pod, error) {
		return &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:       name,
				Namespace:  namespace,
				Finalizers: []string{Finalizer},
			},
			Status: v1.PodStatus{Phase: v1.PodSucceeded},
		}, nil
	}
	c.updatePod = func(pod *v1.Pod) (*v1.Pod, error) {
		if pod.Name != name {
			return nil, fmt.Errorf("invalid pod %s!=%s", name, pod.Name)
		}
		if sets.NewString(pod.ObjectMeta.Finalizers...).Has(Finalizer) {
			return nil, fmt.Errorf("with finalizer %v", pod.ObjectMeta.Finalizers)
		}
		return nil, fmt.Errorf("test remove finalizer failed")
	}
	err = c.reconcile(name)
	if err == nil || err.Error() != "test remove finalizer failed" {
		t.Errorf("reconcile() failed: %v", err)
	}

	// 7. test remove finalizer succeeded.
	c.updatePod = func(pod *v1.Pod) (*v1.Pod, error) {
		if pod.Name != name {
			return nil, fmt.Errorf("invalid pod %s!=%s", name, pod.Name)
		}
		if sets.NewString(pod.ObjectMeta.Finalizers...).Has(Finalizer) {
			return nil, fmt.Errorf("with finalizer %v", pod.ObjectMeta.Finalizers)
		}
		return pod.DeepCopy(), nil
	}
	err = c.reconcile(name)
	if err != nil {
		t.Errorf("reconcile() failed: %v", err)
	}

	// 8. test delete pod failed.
	deleteTime := metav1.Now()
	c.getPod = func(namespace, name string) (*v1.Pod, error) {
		return &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:              name,
				Namespace:         namespace,
				DeletionTimestamp: &deleteTime,
			},
			Status: v1.PodStatus{Phase: v1.PodSucceeded},
		}, nil
	}
	c.deletePod = func(namespace, name string) error {
		return fmt.Errorf("test delete pod failed")
	}
	err = c.reconcile(name)
	if err == nil || err.Error() != "test delete pod failed" {
		t.Errorf("reconcile() failed: %v", err)
	}

	// 9. test delete pod succeeded.
	c.deletePod = func(namespace, name string) error {
		return nil
	}
	err = c.reconcile(name)
	if err != nil {
		t.Errorf("reconcile() failed: %v", err)
	}
}
