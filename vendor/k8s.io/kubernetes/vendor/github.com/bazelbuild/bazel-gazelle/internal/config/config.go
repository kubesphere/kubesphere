/* Copyright 2017 The Bazel Authors. All rights reserved.

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

package config

import (
	"fmt"
	"go/build"
	"strings"
)

// Config holds information about how Gazelle should run. This is mostly
// based on command-line arguments.
type Config struct {
	// Dirs is a list of absolute paths to directories where Gazelle should run.
	Dirs []string

	// RepoRoot is the absolute, canonical path to the root directory of the
	// repository with all symlinks resolved.
	RepoRoot string

	// ValidBuildFileNames is a list of base names that are considered valid
	// build files. Some repositories may have files named "BUILD" that are not
	// used by Bazel and should be ignored. Must contain at least one string.
	ValidBuildFileNames []string

	// GenericTags is a set of build constraints that are true on all platforms.
	// It should not be nil.
	GenericTags BuildTags

	// GoPrefix is the portion of the import path for the root of this repository.
	// This is used to map imports to labels within the repository.
	GoPrefix string

	// GoPrefixRel is the slash-separated path to the directory where GoPrefix
	// was set, relative to the repository root. "" for the repository root.
	GoPrefixRel string

	// ShouldFix determines whether Gazelle attempts to remove and replace
	// usage of deprecated rules.
	ShouldFix bool

	// DepMode determines how imports outside of GoPrefix are resolved.
	DepMode DependencyMode

	// ProtoMode determines how rules are generated for protos.
	ProtoMode ProtoMode

	// ProtoModeExplicit indicates whether the proto mode was set explicitly.
	ProtoModeExplicit bool
}

var DefaultValidBuildFileNames = []string{"BUILD.bazel", "BUILD"}

func (c *Config) IsValidBuildFileName(name string) bool {
	for _, n := range c.ValidBuildFileNames {
		if name == n {
			return true
		}
	}
	return false
}

func (c *Config) DefaultBuildFileName() string {
	return c.ValidBuildFileNames[0]
}

// BuildTags is a set of build constraints.
type BuildTags map[string]bool

// SetBuildTags sets GenericTags by parsing as a comma separated list. An
// error will be returned for tags that wouldn't be recognized by "go build".
// PreprocessTags should be called after this.
func (c *Config) SetBuildTags(tags string) error {
	c.GenericTags = make(BuildTags)
	if tags == "" {
		return nil
	}
	for _, t := range strings.Split(tags, ",") {
		if strings.HasPrefix(t, "!") {
			return fmt.Errorf("build tags can't be negated: %s", t)
		}
		c.GenericTags[t] = true
	}
	return nil
}

// PreprocessTags adds some tags which are on by default before they are
// used to match files.
func (c *Config) PreprocessTags() {
	if c.GenericTags == nil {
		c.GenericTags = make(BuildTags)
	}
	c.GenericTags["gc"] = true
}

// CheckPrefix checks that a string may be used as a prefix. We forbid local
// (relative) imports and those beginning with "/". We allow the empty string,
// but generated rules must not have an empty importpath.
func CheckPrefix(prefix string) error {
	if strings.HasPrefix(prefix, "/") || build.IsLocalImport(prefix) {
		return fmt.Errorf("invalid prefix: %q", prefix)
	}
	return nil
}

// DependencyMode determines how imports of packages outside of the prefix
// are resolved.
type DependencyMode int

const (
	// ExternalMode indicates imports should be resolved to external dependencies
	// (declared in WORKSPACE).
	ExternalMode DependencyMode = iota

	// VendorMode indicates imports should be resolved to libraries in the
	// vendor directory.
	VendorMode
)

// DependencyModeFromString converts a string from the command line
// to a DependencyMode. Valid strings are "external", "vendor". An error will
// be returned for an invalid string.
func DependencyModeFromString(s string) (DependencyMode, error) {
	switch s {
	case "external":
		return ExternalMode, nil
	case "vendored":
		return VendorMode, nil
	default:
		return 0, fmt.Errorf("unrecognized dependency mode: %q", s)
	}
}

// ProtoMode determines how proto rules are generated.
type ProtoMode int

const (
	// DefaultProtoMode generates proto_library and new grpc_proto_library rules.
	// .pb.go files are excluded when there is a .proto file with a similar name.
	DefaultProtoMode ProtoMode = iota

	// DisableProtoMode ignores .proto files. .pb.go files are treated
	// as normal sources.
	DisableProtoMode

	// LegacyProtoMode generates filegroups for .proto files if .pb.go files
	// are present in the same directory.
	LegacyProtoMode
)

func ProtoModeFromString(s string) (ProtoMode, error) {
	switch s {
	case "default":
		return DefaultProtoMode, nil
	case "disable":
		return DisableProtoMode, nil
	case "legacy":
		return LegacyProtoMode, nil
	default:
		return 0, fmt.Errorf("unrecognized proto mode: %q", s)
	}
}
