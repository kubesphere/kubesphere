package json

import v1 "github.com/open-policy-agent/opa/v1/ast/json"

// Options defines the options for JSON operations,
// currently only marshaling can be configured
type Options = v1.Options

// MarshalOptions defines the options for JSON marshaling,
// currently only toggling the marshaling of location information is supported
type MarshalOptions = v1.MarshalOptions

// NodeToggle is a generic struct to allow the toggling of
// settings for different ast node types
type NodeToggle = v1.NodeToggle
