/*
Copyright 2018 The Kubernetes Authors.

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

// Package log contains (deprecated) utilities for fetching a new logger when
// one is not already available.
//
// Deprecated: use pkg/log
package log

import (
	"github.com/go-logr/logr"

	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	// ZapLogger is a Logger implementation.
	// If development is true, a Zap development config will be used
	// (stacktraces on warnings, no sampling), otherwise a Zap production
	// config will be used (stacktraces on errors, sampling).
	ZapLogger = zap.Logger

	// ZapLoggerTo returns a new Logger implementation using Zap which logs
	// to the given destination, instead of stderr.  It otherise behaves like
	// ZapLogger.
	ZapLoggerTo = zap.Logger

	// SetLogger sets a concrete logging implementation for all deferred Loggers.
	SetLogger = log.SetLogger

	// Log is the base logger used by kubebuilder.  It delegates
	// to another logr.Logger.  You *must* call SetLogger to
	// get any actual logging.
	Log = log.Log

	// KBLog is a base parent logger for use inside controller-runtime.
	// Deprecated: don't use this outside controller-runtime
	// (inside CR, use pkg/internal/log.RuntimeLog)
	KBLog logr.Logger
)

func init() {
	KBLog = log.Log.WithName("controller-runtime")
}
