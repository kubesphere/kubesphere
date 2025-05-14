// Copyright 2023 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// This package provides options for JSON marshalling of AST nodes, and location
// data in particular. Since location data occupies a significant portion of the
// AST when included, it is excluded by default. The options provided here allow
// changing that behavior — either for all nodes or for specific types. Since
// JSONMarshaller implementations have access only to the node being marshaled,
// our options are to either attach these settings to *all* nodes in the AST, or
// to provide them via global state. The former is perhaps a little more elegant,
// and is what we went with initially. The cost of attaching these settings to
// every node however turned out to be non-negligible, and given that the number
// of users who have an interest in AST serialization are likely to be few, we
// have since switched to using global state, as provided here. Note that this
// is mostly to provide an equivalent feature to what we had before, should
// anyone depend on that. Users who need fine-grained control over AST
// serialization are recommended to use external libraries for that purpose,
// such as `github.com/json-iterator/go`.
package json

import "sync"

// Options defines the options for JSON operations,
// currently only marshaling can be configured
type Options struct {
	MarshalOptions MarshalOptions
}

// MarshalOptions defines the options for JSON marshaling,
// currently only toggling the marshaling of location information is supported
type MarshalOptions struct {
	// IncludeLocation toggles the marshaling of location information
	IncludeLocation NodeToggle
	// IncludeLocationText additionally/optionally includes the text of the location
	IncludeLocationText bool
	// ExcludeLocationFile additionally/optionally excludes the file of the location
	// Note that this is inverted (i.e. not "include" as the default needs to remain false)
	ExcludeLocationFile bool
}

// NodeToggle is a generic struct to allow the toggling of
// settings for different ast node types
type NodeToggle struct {
	Term           bool
	Package        bool
	Comment        bool
	Import         bool
	Rule           bool
	Head           bool
	Expr           bool
	SomeDecl       bool
	Every          bool
	With           bool
	Annotations    bool
	AnnotationsRef bool
}

// configuredJSONOptions synchronizes access to the global JSON options
type configuredJSONOptions struct {
	options Options
	lock    sync.RWMutex
}

var options = &configuredJSONOptions{
	options: Defaults(),
}

// SetOptions sets the global options for marshalling AST nodes to JSON
func SetOptions(opts Options) {
	options.lock.Lock()
	defer options.lock.Unlock()
	options.options = opts
}

// GetOptions returns (a copy of) the global options for marshalling AST nodes to JSON
func GetOptions() Options {
	options.lock.RLock()
	defer options.lock.RUnlock()
	return options.options
}

// Defaults returns the default JSON options, which is to exclude location
// information in serialized JSON AST.
func Defaults() Options {
	return Options{
		MarshalOptions: MarshalOptions{
			IncludeLocation: NodeToggle{
				Term:           false,
				Package:        false,
				Comment:        false,
				Import:         false,
				Rule:           false,
				Head:           false,
				Expr:           false,
				SomeDecl:       false,
				Every:          false,
				With:           false,
				Annotations:    false,
				AnnotationsRef: false,
			},
			IncludeLocationText: false,
			ExcludeLocationFile: false,
		},
	}
}
