/*
Copyright 2019 The Kubernetes Authors.

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

// Package signals contains libraries for handling signals to gracefully
// shutdown the manager in combination with Kubernetes pod graceful termination
// policy.
//
// Deprecated: use pkg/manager/signals instead.
package signals

import (
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

var (
	// SetupSignalHandler registers for SIGTERM and SIGINT. A stop channel is returned
	// which is closed on one of these signals. If a second signal is caught, the program
	// is terminated with exit code 1.
	SetupSignalHandler = signals.SetupSignalHandler
)
