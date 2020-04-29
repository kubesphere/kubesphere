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

package util

// The functions in this file are exposed as variables to allow them
// to be overridden for testing purposes. Simulated scale testing
// requires being able to change the namespace of target resources
// (NamespaceForCluster and QualifiedNameForCluster) and ensure that
// the namespace of a federated resource will always be the kubefed
// system namespace (NamespaceForResource).

func namespaceForCluster(clusterName, namespace string) string {
	return namespace
}

// NamespaceForCluster returns the namespace to use for the given cluster.
var NamespaceForCluster = namespaceForCluster

func namespaceForResource(resourceNamespace, fedNamespace string) string {
	return resourceNamespace
}

// NamespaceForResource returns either the kubefed namespace or
// resource namespace.
var NamespaceForResource = namespaceForResource

func qualifiedNameForCluster(clusterName string, qualifiedName QualifiedName) QualifiedName {
	return qualifiedName
}

// QualifiedNameForCluster returns the qualified name to use for the
// given cluster.
var QualifiedNameForCluster = qualifiedNameForCluster
