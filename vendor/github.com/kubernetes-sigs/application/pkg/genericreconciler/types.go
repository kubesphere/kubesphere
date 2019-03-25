/*
Copyright 2018 The Kubernetes Authors
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

package genericreconciler

import (
	cr "github.com/kubernetes-sigs/application/pkg/customresource"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ reconcile.Reconciler = &Reconciler{}

// Reconciler defines fields needed for all airflow controllers
// +k8s:deepcopy-gen=false
type Reconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Handle  cr.Handle
	Manager manager.Manager
}

// ReconcilerConfig config defines reconciler parameters
// +k8s:deepcopy-gen=false
type ReconcilerConfig struct {
}

// KVmap is a map[string]string
type KVmap map[string]string
