package regorewriter

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// FilePath represents a path on the filesystem and handles reparenting the file relative to a
// path prefix.
type FilePath struct {
	path string
}

// Path returns the current path value.
func (f *FilePath) Path() string {
	return f.path
}

// Reparent adjusts the parent from a current path prefix to a new path prefix.
func (f *FilePath) Reparent(old, new string) error {
	if filepath.IsAbs(f.path) != filepath.IsAbs(old) ||
		filepath.IsAbs(old) != filepath.IsAbs(new) {
		return errors.Errorf("relative path / absolute path mismatch: %s %s %s", f.path, old, new)
	}

	relPath, err := filepath.Rel(old, f.path)
	if err != nil {
		return err
	}
	if strings.HasPrefix(relPath, "..") {
		return errors.Errorf("old is not a prefix of path")
	}

	f.path = filepath.Join(new, relPath)
	return nil
}
