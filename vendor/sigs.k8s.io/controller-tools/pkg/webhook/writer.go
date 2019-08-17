package webhook

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/spf13/afero"

	"k8s.io/apimachinery/pkg/runtime"
)

// WriterOptions specifies the input and output.
type WriterOptions struct {
	InputDir       string
	OutputDir      string
	PatchOutputDir string

	// inFs is filesystem to be used for reading input
	inFs afero.Fs
	// outFs is filesystem to be used for writing out the result
	outFs afero.Fs
}

// SetDefaults sets up the default options for RBAC Manifest generator.
func (o *WriterOptions) SetDefaults() {
	if o.inFs == nil {
		o.inFs = afero.NewOsFs()
	}
	if o.outFs == nil {
		o.outFs = afero.NewOsFs()
	}

	if len(o.InputDir) == 0 {
		o.InputDir = filepath.Join(".", "pkg", "webhook")
	}
	if len(o.OutputDir) == 0 {
		o.OutputDir = filepath.Join(".", "config", "webhook")
	}
	if len(o.PatchOutputDir) == 0 {
		o.PatchOutputDir = filepath.Join(".", "config", "default")
	}
}

// Validate validates the input options.
func (o *WriterOptions) Validate() error {
	if _, err := o.inFs.Stat(o.InputDir); err != nil {
		return fmt.Errorf("invalid input directory '%s' %v", o.InputDir, err)
	}
	return nil
}

// WriteObjectsToDisk writes object to the location specified in WriterOptions.
func (o *WriterOptions) WriteObjectsToDisk(objects ...runtime.Object) error {
	exists, err := afero.DirExists(o.outFs, o.OutputDir)
	if err != nil {
		return err
	}
	if !exists {
		err = o.outFs.MkdirAll(o.OutputDir, 0766)
		if err != nil {
			return err
		}
	}

	var buf bytes.Buffer
	isFirstObject := true
	for _, obj := range objects {
		if !isFirstObject {
			_, err = buf.WriteString("---\n")
			if err != nil {
				return err
			}
		}
		marshalled, err := yaml.Marshal(obj)
		if err != nil {
			return err
		}
		_, err = buf.Write(marshalled)
		if err != nil {
			return err
		}
		isFirstObject = false
	}
	err = afero.WriteFile(o.outFs, path.Join(o.OutputDir, "webhookmanifests.yaml"), buf.Bytes(), 0644)
	if err != nil {
		return err
	}
	return nil
}
