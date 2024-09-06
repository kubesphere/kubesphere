package json

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
