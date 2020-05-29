/*
Copyright 2018 The KubeSphere Authors.

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

package test

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
	"time"

	"k8s.io/klog"

	"github.com/prometheus/common/log"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	extscheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type TestCtx struct {
	Client     client.Client
	ID         string
	t          *testing.T
	cleanupFns []cleanupFn
	Namespace  string
}

type CleanupOptions struct {
	TestContext   *TestCtx
	Timeout       time.Duration
	RetryInterval time.Duration
}
type cleanupFn func(option *CleanupOptions) error

type AddToSchemeFunc = func(*runtime.Scheme) error

func NewTestCtx(t *testing.T, namespace string) *TestCtx {
	var prefix string
	if t != nil {
		// TestCtx is used among others for namespace names where '/' is forbidden
		prefix = strings.TrimPrefix(
			strings.Replace(
				strings.ToLower(t.Name()),
				"/",
				"-",
				-1,
			),
			"test",
		)
	} else {
		prefix = "main"
	}
	id := prefix + "-" + strconv.FormatInt(time.Now().Unix(), 10)
	return &TestCtx{
		ID:        id,
		t:         t,
		Namespace: namespace,
	}
}

func (t *TestCtx) Setup(yamlPath string, crdPath string, schemes ...AddToSchemeFunc) error {
	cfg, err := config.GetConfig()
	if err != nil {
		klog.Error("Failed to get kubeconfig")
		return err
	}
	for _, f := range schemes {
		err = f(scheme.Scheme)
		if err != nil {
			klog.Errorln("Failed to add scheme")
			return err
		}
	}
	extscheme.AddToScheme(scheme.Scheme)
	dynClient, err := client.New(cfg, client.Options{})
	if err != nil {
		return err
	}
	t.Client = dynClient
	err = EnsureNamespace(t.Client, t.Namespace)
	if err != nil {
		return err
	}
	bytes, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		klog.Errorln("Failed to read yaml file")
		return err
	}
	err = t.CreateFromYAML(bytes, true)
	if err != nil {
		klog.Error("Failed to install controller")
		return err
	}
	return nil
}

func WaitForController(c client.Client, namespace, name string, replica int32, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		controller := &appsv1.Deployment{}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		err = c.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, controller)
		if apierrors.IsNotFound(err) {
			klog.Infof("Cannot find controller %s", name)
			return false, nil
		}
		if err != nil {
			klog.Errorf("Get error %s when waiting for controller up", err.Error())
			return false, err
		}
		if controller.Status.ReadyReplicas == replica {
			return true, nil
		}
		return false, nil
	})
	return err
}

func WaitForDeletion(dynclient client.Client, obj runtime.Object, retryInterval, timeout time.Duration) error {
	key, err := client.ObjectKeyFromObject(obj)
	if err != nil {
		return err
	}
	kind := obj.GetObjectKind().GroupVersionKind().Kind
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err = wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		err = dynclient.Get(ctx, key, obj)
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			klog.Errorf("Get error %s when waiting for controller down", err.Error())
			return false, err
		}
		klog.Infof("Waiting for %s %s to be deleted\n", kind, key)
		return false, nil
	})
	if err != nil {
		return err
	}
	return nil
}

func GetLogOfPod(rest *rest.RESTClient, namespace, name string, logOptions *corev1.PodLogOptions, out io.Writer) error {
	req := rest.Get().Namespace(namespace).Name(name).SubResource("log").Param("follow", strconv.FormatBool(logOptions.Follow)).
		Param("container", logOptions.Container).
		Param("previous", strconv.FormatBool(logOptions.Previous)).
		Param("timestamps", strconv.FormatBool(logOptions.Timestamps))
	if logOptions.SinceSeconds != nil {
		req.Param("sinceSeconds", strconv.FormatInt(*logOptions.SinceSeconds, 10))
	}
	if logOptions.SinceTime != nil {
		req.Param("sinceTime", logOptions.SinceTime.Format(time.RFC3339))
	}
	if logOptions.LimitBytes != nil {
		req.Param("limitBytes", strconv.FormatInt(*logOptions.LimitBytes, 10))
	}
	if logOptions.TailLines != nil {
		req.Param("tailLines", strconv.FormatInt(*logOptions.TailLines, 10))
	}
	readCloser, err := req.Stream()
	if err != nil {
		return err
	}
	defer readCloser.Close()
	_, err = io.Copy(out, readCloser)
	return err
}

func (ctx *TestCtx) CreateFromYAML(yamlFile []byte, skipIfExists bool) error {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	scanner := NewYAMLScanner(yamlFile)
	for scanner.Scan() {
		yamlSpec := scanner.Bytes()
		obj, groupVersionKind, err := decode(yamlSpec, nil, nil)
		if err != nil {
			klog.Errorf("Error while decoding YAML object. Err was: %s", err)
			return err
		}
		klog.Infof("Successfully decode object %v", groupVersionKind)
		err = ctx.Client.Create(context.TODO(), obj)
		if skipIfExists && apierrors.IsAlreadyExists(err) {
			continue
		}
		if err != nil {
			klog.Errorf("Failed to create %v to k8s", obj)
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan manifest: (%v)", err)
	}
	return nil
}

func (ctx *TestCtx) GetID() string {
	return ctx.ID
}

func (ctx *TestCtx) Cleanup(option *CleanupOptions) {
	failed := false
	for i := len(ctx.cleanupFns) - 1; i >= 0; i-- {
		err := ctx.cleanupFns[i](option)
		if err != nil {
			failed = true
			if ctx.t != nil {
				ctx.t.Errorf("A cleanup function failed with error: (%v)\n", err)
			} else {
				log.Errorf("A cleanup function failed with error: (%v)", err)
			}
		}
	}
	if ctx.t == nil && failed {
		log.Fatal("A cleanup function failed")
	}
}

func (ctx *TestCtx) AddCleanupFn(fn cleanupFn) {
	ctx.cleanupFns = append(ctx.cleanupFns, fn)
}

func WaitForJobSucceed(c client.Client, namespace, name string, retryInterval, timeout time.Duration) error {
	return waitForJobStatus(c, namespace, name, batchv1.JobComplete, retryInterval, timeout)
}

func WaitForJobFail(c client.Client, namespace, name string, retryInterval, timeout time.Duration) error {
	return waitForJobStatus(c, namespace, name, batchv1.JobFailed, retryInterval, timeout)
}

func waitForJobStatus(c client.Client, namespace, name string, jobstatus batchv1.JobConditionType, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		job := &batchv1.Job{}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		err = c.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, job)
		if apierrors.IsNotFound(err) {
			klog.Infof("Cannot find job %s", name)
			return false, nil
		}
		if err != nil {
			klog.Errorf("Get error %s when waiting for job up", err.Error())
			return false, err
		}
		if len(job.Status.Conditions) > 0 && job.Status.Conditions[0].Type == jobstatus {
			return true, nil
		}
		return false, nil
	})
	return err
}

func EnsureNamespace(c client.Client, namespace string) error {
	ns := &corev1.Namespace{}
	ns.Name = namespace
	err := c.Create(context.TODO(), ns)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			klog.Infof("Namespace %s is existed", namespace)
			return nil
		}
	}
	return err
}

func DeleteNamespace(c client.Client, namespace string) error {
	ns := &corev1.Namespace{}
	ns.Name = namespace
	return c.Delete(context.TODO(), ns)
}
