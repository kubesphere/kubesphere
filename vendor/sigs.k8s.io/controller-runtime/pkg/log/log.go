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

// Package log contains utilities for fetching a new logger
// when one is not already available.
//
// The Log Handle
//
// This package contains a root logr.Logger Log.  It may be used to
// get a handle to whatever the root logging implementation is.  By
// default, no implementation exists, and the handle returns "promises"
// to loggers.  When the implementation is set using SetLogger, these
// "promises" will be converted over to real loggers.
//
// Logr
//
// All logging in controller-runtime is structured, using a set of interfaces
// defined by a package called logr
// (https://godoc.org/github.com/go-logr/logr).  The sub-package zap provides
// helpers for setting up logr backed by Zap (go.uber.org/zap).
package log

import (
	"context"

	"github.com/go-logr/logr"
)

var (
	contextKey = &struct{}{}
)

// SetLogger sets a concrete logging implementation for all deferred Loggers.
func SetLogger(l logr.Logger) {
	Log.Fulfill(l)
}

// Log is the base logger used by kubebuilder.  It delegates
// to another logr.Logger.  You *must* call SetLogger to
// get any actual logging.
var Log = NewDelegatingLogger(NullLogger{})

// FromContext returns a logger with predefined values from a context.Context.
func FromContext(ctx context.Context, keysAndValues ...interface{}) logr.Logger {
	var log logr.Logger
	if ctx == nil {
		log = Log
	} else {
		lv := ctx.Value(contextKey)
		log = lv.(logr.Logger)
	}
	log.WithValues(keysAndValues...)
	return log
}

// IntoContext takes a context and sets the logger as one of its keys.
// Use FromContext function to retrieve the logger.
func IntoContext(ctx context.Context, log logr.Logger) context.Context {
	return context.WithValue(ctx, contextKey, log)
}
