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

package webhook

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"
	"github.com/spf13/afero"

	"sigs.k8s.io/controller-tools/pkg/internal/general"
)

// Options represent options for generating the webhook manifests.
type Options struct {
	// WriterOptions specifies the input and output
	WriterOptions

	generatorOptions
}

// Generate generates RBAC manifests by parsing the RBAC annotations in Go source
// files specified in the input directory.
func Generate(o *Options) error {
	if err := o.WriterOptions.Validate(); err != nil {
		return err
	}

	err := general.ParseDir(o.InputDir, o.parseAnnotation)
	if err != nil {
		return fmt.Errorf("failed to parse the input dir: %v", err)
	}

	if len(o.webhooks) == 0 {
		return nil
	}

	objs, err := o.Generate()
	if err != nil {
		return err
	}

	err = o.WriteObjectsToDisk(objs...)
	if err != nil {
		return err
	}

	return o.controllerManagerPatch()
}

func (o *Options) controllerManagerPatch() error {
	var kustomizeLabelPatch = `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: controller-manager
spec:
  template:
    metadata:
{{- with .Labels }}
      labels:
{{ toYaml . | indent 8 }}
{{- end }}
    spec:
      containers:
      - name: manager
        ports:
        - containerPort: {{ .Port }}
          name: webhook-server
          protocol: TCP
        volumeMounts:
        - mountPath: {{ .CertDir }}
          name: cert
          readOnly: true
      volumes:
      - name: cert
        secret:
          defaultMode: 420
          secretName: {{ .SecretName }}
`

	type KustomizeLabelPatch struct {
		Labels     map[string]string
		SecretName string
		Port       int32
		CertDir    string
	}

	p := KustomizeLabelPatch{
		Labels:     o.service.selectors,
		SecretName: o.secret.Name,
		Port:       o.port,
		CertDir:    o.certDir,
	}
	funcMap := template.FuncMap{
		"toYaml": toYAML,
		"indent": indent,
	}
	temp, err := template.New("kustomizeLabelPatch").Funcs(funcMap).Parse(kustomizeLabelPatch)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(nil)
	if err := temp.Execute(buf, p); err != nil {
		return err
	}
	return afero.WriteFile(o.outFs, path.Join(o.PatchOutputDir, "manager_patch.yaml"), buf.Bytes(), 0644)
}

func toYAML(m map[string]string) (string, error) {
	d, err := yaml.Marshal(m)
	return string(d), err
}

func indent(n int, s string) (string, error) {
	buf := bytes.NewBuffer(nil)
	for _, elem := range strings.Split(s, "\n") {
		for i := 0; i < n; i++ {
			_, err := buf.WriteRune(' ')
			if err != nil {
				return "", err
			}
		}
		_, err := buf.WriteString(elem)
		if err != nil {
			return "", err
		}
		_, err = buf.WriteRune('\n')
		if err != nil {
			return "", err
		}
	}
	return strings.TrimRight(buf.String(), " \n"), nil
}
