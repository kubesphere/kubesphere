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

package printers

import (
	"strings"

	"github.com/spf13/cobra"

	"k8s.io/kubernetes/pkg/api/legacyscheme"
	kubectlscheme "k8s.io/kubernetes/pkg/kubectl/scheme"
)

// JSONYamlPrintFlags provides default flags necessary for json/yaml printing.
// Given the following flag values, a printer can be requested that knows
// how to handle printing based on these values.
type JSONYamlPrintFlags struct{}

// ToPrinter receives an outputFormat and returns a printer capable of
// handling --output=(yaml|json) printing.
// Returns false if the specified outputFormat does not match a supported format.
// Supported Format types can be found in pkg/printers/printers.go
func (f *JSONYamlPrintFlags) ToPrinter(outputFormat string) (ResourcePrinter, error) {
	var printer ResourcePrinter

	outputFormat = strings.ToLower(outputFormat)
	switch outputFormat {
	case "json":
		printer = &JSONPrinter{}
	case "yaml":
		printer = &YAMLPrinter{}
	default:
		return nil, NoCompatiblePrinterError{Options: f, OutputFormat: &outputFormat}
	}

	// wrap the printer in a versioning printer that understands when to convert and when not to convert
	return NewVersionedPrinter(printer, legacyscheme.Scheme, legacyscheme.Scheme, kubectlscheme.Versions...), nil

}

// AddFlags receives a *cobra.Command reference and binds
// flags related to JSON or Yaml printing to it
func (f *JSONYamlPrintFlags) AddFlags(c *cobra.Command) {}

// NewJSONYamlPrintFlags returns flags associated with
// yaml or json printing, with default values set.
func NewJSONYamlPrintFlags() *JSONYamlPrintFlags {
	return &JSONYamlPrintFlags{}
}
