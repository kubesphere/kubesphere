/*
Copyright 2019 The KubeSphere Authors.

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

package healthz

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"k8s.io/klog/v2"
	"net/http"

	"kubesphere.io/kubesphere/pkg/server/errors"
)

func AddToContainer(container *restful.Container, path string, checks ...HealthChecker) error {
	container.Handle(path, http.HandlerFunc(adaptCheckToHandler(checks...)))

	for _, check := range checks {
		container.Handle(fmt.Sprintf("%s/%v", path, check.Name()), http.HandlerFunc(adaptCheckToHandler([]HealthChecker{check}...)))
	}

	return nil
}

func InstallHandler(container *restful.Container, checks ...HealthChecker) error {
	if len(checks) == 0 {
		klog.V(4).Info("No default health checks specified. Installing the ping handler.")
		checks = []HealthChecker{PingHealthz}
	}
	return AddToContainer(container, "/healthz", checks...)
}

func InstallLivezHandler(container *restful.Container, checks ...HealthChecker) error {
	if len(checks) == 0 {
		klog.V(4).Info("No default health checks specified. Installing the ping handler.")
		checks = []HealthChecker{PingHealthz}
	}
	return AddToContainer(container, "/livez", checks...)
}

// adaptCheckToHandler returns an http.HandlerFunc that serves the provided checks.
func adaptCheckToHandler(checks ...HealthChecker) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("X-Content-Type-Options", "nosniff")

		for _, check := range checks {
			err := check.Check(request)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(fmt.Sprintf(`{"message": %s}`, errors.Wrap(err))))
				return
			}
		}
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte(`{"message": "success"}`))
	}
}

// HealthChecker is a named healthz checker.
type HealthChecker interface {
	Name() string
	Check(req *http.Request) error
}

// PingHealthz returns true automatically when checked
var PingHealthz HealthChecker = ping{}

// ping implements the simplest possible healthz checker.
type ping struct{}

func (ping) Name() string {
	return "ping"
}

// PingHealthz is a health check that returns true.
func (ping) Check(_ *http.Request) error {
	return nil
}
