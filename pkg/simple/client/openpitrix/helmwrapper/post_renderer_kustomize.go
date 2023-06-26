// Copyright 2022 The KubeSphere Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package helmwrapper

import (
	"bytes"
	"encoding/json"
	"sync"

	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resmap"
	kustypes "sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type postRendererKustomize struct {
	labels      map[string]string
	annotations map[string]string
}

func newPostRendererKustomize(labels, annotations map[string]string) *postRendererKustomize {
	return &postRendererKustomize{
		labels,
		annotations,
	}
}

func writeToFile(fs filesys.FileSystem, path string, content []byte) error {
	helmOutput, err := fs.Create(path)
	if err != nil {
		return err
	}
	helmOutput.Write(content)
	if err := helmOutput.Close(); err != nil {
		return err
	}
	return nil
}

func writeFile(fs filesys.FileSystem, path string, content *bytes.Buffer) error {
	helmOutput, err := fs.Create(path)
	if err != nil {
		return err
	}
	content.WriteTo(helmOutput)
	if err := helmOutput.Close(); err != nil {
		return err
	}
	return nil
}

func (k *postRendererKustomize) Run(renderedManifests *bytes.Buffer) (modifiedManifests *bytes.Buffer, err error) {
	fs := filesys.MakeFsInMemory()
	input := "./.local-helm-output.yaml"
	cfg := kustypes.Kustomization{
		Resources:         []string{input},
		CommonAnnotations: k.annotations,                       // add extra annotations to output
		Labels:            []kustypes.Label{{Pairs: k.labels}}, // Labels to add to all objects but not selectors.
	}
	cfg.APIVersion = kustypes.KustomizationVersion
	cfg.Kind = kustypes.KustomizationKind
	if err := writeFile(fs, input, renderedManifests); err != nil {
		return nil, err
	}

	// Write kustomization config to file.
	kustomization, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	if err := writeToFile(fs, "kustomization.yaml", kustomization); err != nil {
		return nil, err
	}
	resMap, err := buildKustomization(fs, ".")
	if err != nil {
		return nil, err
	}
	yaml, err := resMap.AsYaml()
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(yaml), nil
}

var kustomizeRenderMutex sync.Mutex

func buildKustomization(fs filesys.FileSystem, dirPath string) (resmap.ResMap, error) {
	kustomizeRenderMutex.Lock()
	defer kustomizeRenderMutex.Unlock()

	buildOptions := &krusty.Options{
		DoLegacyResourceSort: true,
		LoadRestrictions:     kustypes.LoadRestrictionsNone,
		AddManagedbyLabel:    false,
		DoPrune:              false,
		PluginConfig:         kustypes.DisabledPluginConfig(),
	}

	k := krusty.MakeKustomizer(buildOptions)
	return k.Run(fs, dirPath)
}
