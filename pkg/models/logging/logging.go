/*
Copyright 2020 KubeSphere Authors

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

package logging

import (
	"io"
	"kubesphere.io/kubesphere/pkg/api/logging/v1alpha2"
	"kubesphere.io/kubesphere/pkg/simple/client/logging"
)

type LoggingOperator interface {
	GetCurrentStats(sf logging.SearchFilter) (v1alpha2.APIResponse, error)
	CountLogsByInterval(sf logging.SearchFilter, interval string) (v1alpha2.APIResponse, error)
	ExportLogs(sf logging.SearchFilter, w io.Writer) error
	SearchLogs(sf logging.SearchFilter, from, size int64, order string) (v1alpha2.APIResponse, error)
}

type loggingOperator struct {
	c logging.Interface
}

func NewLoggingOperator(client logging.Interface) LoggingOperator {
	return &loggingOperator{client}
}

func (l loggingOperator) GetCurrentStats(sf logging.SearchFilter) (v1alpha2.APIResponse, error) {
	res, err := l.c.GetCurrentStats(sf)
	return v1alpha2.APIResponse{Statistics: &res}, err
}

func (l loggingOperator) CountLogsByInterval(sf logging.SearchFilter, interval string) (v1alpha2.APIResponse, error) {
	res, err := l.c.CountLogsByInterval(sf, interval)
	return v1alpha2.APIResponse{Histogram: &res}, err
}

func (l loggingOperator) ExportLogs(sf logging.SearchFilter, w io.Writer) error {
	return l.c.ExportLogs(sf, w)
}

func (l loggingOperator) SearchLogs(sf logging.SearchFilter, from, size int64, order string) (v1alpha2.APIResponse, error) {
	res, err := l.c.SearchLogs(sf, from, size, order)
	return v1alpha2.APIResponse{Logs: &res}, err
}
